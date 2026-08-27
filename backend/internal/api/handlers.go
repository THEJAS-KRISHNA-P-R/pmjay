package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pmjay-advocate/backend/internal/document"
	"github.com/pmjay-advocate/backend/internal/extract"
	"github.com/pmjay-advocate/backend/internal/hbp"
	"github.com/pmjay-advocate/backend/internal/response"
	"github.com/pmjay-advocate/backend/internal/retrieval"
	"github.com/pmjay-advocate/backend/internal/store"
	"github.com/pmjay-advocate/backend/internal/tiering"
)

// Server holds every dependency the HTTP handlers need. Constructed once
// in cmd/server/main.go and passed real implementations; tests construct
// it with fakes (extract.FakeClient, store.NewMemStore()) so the entire
// HTTP layer is testable offline.
type Server struct {
	Dataset   *hbp.Dataset
	Extractor extract.Extractor
	Store     store.Store
	Logger    *slog.Logger
}

const (
	minDescriptionLength = 5
	maxDescriptionLength = 4000

	// maxRequestBodyBytes caps the raw HTTP request body this API will
	// read into memory, on every endpoint that accepts one. Without
	// this, a caller can send an arbitrarily large body — gigabytes, if
	// they choose — and json.Decode will read all of it into memory
	// before maxDescriptionLength is ever checked, since that check
	// happens on the decoded field, not the wire size. This is
	// deliberately generous relative to maxDescriptionLength (a ~4000
	// character description is a few KB even with JSON escaping) so it
	// never rejects a legitimate request, while still bounding the
	// worst case to something a single instance can absorb from many
	// concurrent callers at once.
	maxRequestBodyBytes = 64 * 1024

	// Evidence fields are short, specific facts by design (Section 7
	// Step 6: a name, a time, a one-line note) — these caps are
	// generous relative to that intent, not an attempt to fit a real
	// note into the smallest space possible. Without them,
	// maxRequestBodyBytes was the only limit on any one field, meaning
	// a single field could balloon to ~64KB. That matters here
	// specifically because every evidence append triggers a full
	// rewrite of the case store to disk (see store.FileStore.flush) —
	// unlike the intake endpoint, this one has no per-call LLM cost to
	// otherwise bound abuse.
	maxStaffNameLength  = 200
	maxApproxTimeLength = 100
	maxNoteLength       = 2000

	// maxEvidenceEntriesPerCase bounds how many evidence entries a
	// single case can accumulate. A real family capturing evidence
	// during one dispute needs a handful of entries at most; this
	// leaves generous headroom while still bounding the worst case —
	// again because of the full-rewrite-per-append cost above.
	maxEvidenceEntriesPerCase = 20
)

// handleIntake implements POST /api/v1/cases — Section 7's Steps 1
// through 4 (extraction, matching, tier decision, drafted response) in
// one request. This is the one endpoint that costs money to call (one
// LLM request per call), which is exactly why internal/api/middleware.go
// rate-limits it specifically.
func (s *Server) handleIntake(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var req IntakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBodyDecodeError(w, err)
		return
	}

	if len(req.Description) < minDescriptionLength {
		writeError(w, http.StatusBadRequest, "description is too short to be meaningful", "")
		return
	}
	if len(req.Description) > maxDescriptionLength {
		writeError(w, http.StatusBadRequest, "description is too long", "")
		return
	}

	ctx := r.Context()

	packageCandidates := retrieval.Retrieve(s.Dataset, req.Description)
	exclusionCandidates := retrieval.RetrieveExclusions(s.Dataset, req.Description)

	extractPkgInfo := make([]extract.CandidatePackageInfo, 0, len(packageCandidates))
	for _, c := range packageCandidates {
		extractPkgInfo = append(extractPkgInfo, extract.CandidatePackageInfo{
			PackageCode: c.Package.PackageCode,
			PackageName: c.Package.PackageName,
			Specialty:   c.Package.Specialty,
			Keywords:    c.Package.CommonDescriptionKeywords,
		})
	}
	extractExclInfo := make([]extract.CandidateExclusionInfo, 0, len(exclusionCandidates))
	for _, c := range exclusionCandidates {
		extractExclInfo = append(extractExclInfo, extract.CandidateExclusionInfo{
			Category:    c.Exclusion.Category,
			DisplayName: c.Exclusion.DisplayName,
			Description: c.Exclusion.Description,
			Keywords:    c.Exclusion.Keywords,
		})
	}

	result, err := s.Extractor.Extract(ctx, req.Description, extractPkgInfo, extractExclInfo)
	if err != nil {
		// Section 10's care-first rule is written as an absolute
		// ("Always") — an infrastructure failure is not an exception to
		// it. A family gets a clear "the system had a problem" message,
		// PLUS the one piece of guidance that's always safe to give
		// regardless of what went wrong: get treatment first, and here
		// is the human fallback (the helpline) that doesn't depend on
		// this system working at all.
		s.Logger.Error("extraction failed", "error", err)
		writeError(w, http.StatusBadGateway,
			"the system could not process this request right now",
			response.CareFirstText+" In the meantime, call the PMJAY helpline directly at 14555 — they can help without needing this tool to be working.",
		)
		return
	}

	decision := tiering.Decide(s.Dataset, req.Description, result)
	built := response.Build(decision)

	now := time.Now()
	record := store.CaseRecord{
		ID:                   store.NewCaseID(),
		CreatedAt:            now,
		UpdatedAt:            now,
		FamilyDescriptionRaw: req.Description,
		Outcome:              string(built.Outcome()),
		Citation:             built.Citation(),
		CareFirstMessage:     built.CareFirstMessage(),
		Disclaimer:           built.Disclaimer(),
		TierMessage:          built.TierMessage(),
		ActionSteps:          built.ActionSteps(),
		ComplaintText:        built.ComplaintText(),
		HospitalScript:       built.HospitalScript(),
		HandoffSummary:       built.HandoffSummary(),
		EvidencePrompt:       built.EvidencePrompt(),
	}

	if err := s.Store.Create(ctx, record); err != nil {
		// The response was already computed correctly; a storage failure
		// shouldn't mean the family gets nothing. Log it, serve the
		// answer anyway, but flag that follow-up (evidence capture,
		// re-fetching this case later) may not work.
		s.Logger.Error("failed to persist case", "error", err, "case_id", record.ID)
	}

	s.Logger.Info("case processed", "case_id", record.ID, "outcome", record.Outcome)
	writeJSON(w, http.StatusCreated, caseRecordToResponse(record))
}

// handleGetCase implements GET /api/v1/cases/{id}.
func (s *Server) handleGetCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	record, found, err := s.Store.Get(r.Context(), id)
	if err != nil {
		s.Logger.Error("store get failed", "error", err, "case_id", id)
		writeError(w, http.StatusInternalServerError, "could not retrieve case", "")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "case not found", "")
		return
	}
	writeJSON(w, http.StatusOK, caseRecordToResponse(record))
}

// handleGetCaseDocument implements GET /api/v1/cases/{id}/document — the
// same case data handleGetCase returns as JSON, rendered instead as a
// downloadable/printable PDF a family can hand to hospital billing
// staff, attach to a CGRMS complaint, or bring to a NALSA Para Legal
// Volunteer. See internal/document's README.md for the design
// reasoning and docs/OPEN_QUESTIONS.md for how this relates to (and
// does not close) the "automated CGRMS complaint submission" gap.
//
// Content-Disposition is deliberately "inline" rather than
// "attachment": on both desktop and mobile browsers this opens the PDF
// in-place using the browser's native viewer, which itself offers
// print/save/share controls — strictly more options than forcing an
// immediate download, and the more forgiving choice on a phone with
// limited storage (see docs/OPEN_QUESTIONS.md's note on the lack of a
// no-install distribution channel — a forced download is one more
// small thing that can go wrong on exactly the kind of device this
// project is most worried about failing).
func (s *Server) handleGetCaseDocument(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	record, found, err := s.Store.Get(r.Context(), id)
	if err != nil {
		s.Logger.Error("store get failed", "error", err, "case_id", id)
		writeError(w, http.StatusInternalServerError, "could not retrieve case", "")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "case not found", "")
		return
	}

	pdfBytes, err := document.BuildCase(record)
	if err != nil {
		// Not currently reachable (see BuildCase's doc comment on why
		// its error return can't presently fail), but a document
		// generation bug that does start failing must degrade to a
		// clear 500, never a silently truncated or corrupt PDF sent to
		// a family mid-crisis.
		s.Logger.Error("document generation failed", "error", err, "case_id", id)
		writeError(w, http.StatusInternalServerError, "could not generate document", "")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="pmjay-case-%s.pdf"`, id))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// handleAddEvidence implements POST /api/v1/cases/{id}/evidence — Section
// 7 Step 6, capturing evidence while it's still available.
func (s *Server) handleAddEvidence(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var req AddEvidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBodyDecodeError(w, err)
		return
	}

	req.StaffName = strings.TrimSpace(req.StaffName)
	req.ApproxTime = strings.TrimSpace(req.ApproxTime)
	req.Note = strings.TrimSpace(req.Note)

	if req.StaffName == "" && req.ApproxTime == "" && req.Note == "" {
		writeError(w, http.StatusBadRequest, "at least one of staff_name, approx_time, or note is required", "")
		return
	}
	if len(req.StaffName) > maxStaffNameLength {
		writeError(w, http.StatusBadRequest, "staff_name is too long", "")
		return
	}
	if len(req.ApproxTime) > maxApproxTimeLength {
		writeError(w, http.StatusBadRequest, "approx_time is too long", "")
		return
	}
	if len(req.Note) > maxNoteLength {
		writeError(w, http.StatusBadRequest, "note is too long", "")
		return
	}

	existing, found, err := s.Store.Get(r.Context(), id)
	if err != nil {
		s.Logger.Error("store get failed", "error", err, "case_id", id)
		writeError(w, http.StatusInternalServerError, "could not save evidence", "")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "case not found", "")
		return
	}
	if len(existing.Evidence) >= maxEvidenceEntriesPerCase {
		writeError(w, http.StatusBadRequest, "this case already has the maximum number of evidence entries", "")
		return
	}

	entry := store.EvidenceEntry{
		CapturedAt: time.Now(),
		StaffName:  req.StaffName,
		ApproxTime: req.ApproxTime,
		Note:       req.Note,
	}

	updated, err := s.Store.AppendEvidence(r.Context(), id, entry)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "case not found", "")
			return
		}
		s.Logger.Error("failed to append evidence", "error", err, "case_id", id)
		writeError(w, http.StatusInternalServerError, "could not save evidence", "")
		return
	}

	writeJSON(w, http.StatusOK, caseRecordToResponse(updated))
}

// handleHealth implements GET /api/v1/health — a cheap, dependency-free
// liveness check (no LLM call, no store access) suitable for a load
// balancer or uptime monitor to poll frequently without it costing
// anything.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":            "ok",
		"packages_loaded":   strconv.Itoa(len(s.Dataset.Packages)),
		"exclusions_loaded": strconv.Itoa(len(s.Dataset.Exclusions)),
	})
}

// writeBodyDecodeError inspects a json.Decode failure and returns the
// right status: 413 specifically when the cause was http.MaxBytesReader
// rejecting an oversized body, 400 for every other decode failure
// (malformed JSON, wrong types, etc.). Go 1.19+'s http.MaxBytesError is
// how MaxBytesReader signals which case this is — see handleIntake and
// handleAddEvidence, both wrapped with it.
func writeBodyDecodeError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		writeError(w, http.StatusRequestEntityTooLarge, "request body is too large", "")
		return
	}
	writeError(w, http.StatusBadRequest, "invalid JSON body", "")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message, fallbackGuidance string) {
	writeJSON(w, status, ErrorResponse{Error: message, FallbackGuidance: fallbackGuidance})
}
