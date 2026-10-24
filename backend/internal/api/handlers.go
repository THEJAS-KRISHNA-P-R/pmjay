package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
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
		//
		// Three failure kinds get three different answers, because they
		// mean different things to the person retrying:
		//
		//   503 — this deployment has no provider key. Retrying now is
		//         pointless; the helpline is the real next step.
		//   504 — the provider took too long. Worth retrying later.
		//   502 — the provider failed. Retry, then use the helpline.
		if errors.Is(err, extract.ErrNotConfigured) {
			s.Logger.Error("intake unavailable: no provider key configured", "error", err)
			writeError(w, http.StatusServiceUnavailable,
				"the guided case analysis is not available on this server right now",
				response.CareFirstText+" Call the PMJAY helpline directly at 14555 — they can help without needing this tool to be working.",
			)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			s.Logger.Warn("extraction timed out", "error", err)
			writeError(w, http.StatusGatewayTimeout,
				"the analysis took too long and was stopped",
				response.CareFirstText+" Please try once more, or call the PMJAY helpline directly at 14555 — they can help without needing this tool to be working.",
			)
			return
		}
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

	resp := caseRecordToResponse(record)
	if err := s.Store.Create(ctx, record); err != nil {
		// The response was already computed correctly; a storage failure
		// shouldn't mean the family gets nothing. Log it, serve the
		// answer anyway, but mark it unpersisted so the UI does not
		// promise follow-up (evidence capture, reopening, deletion) that
		// cannot work.
		s.Logger.Error("failed to persist case", "error", err, "case_id", shortCaseID(record.ID))
		resp.Persisted = false
	}

	s.Logger.Info("case processed", "case_id", shortCaseID(record.ID), "outcome", record.Outcome)
	writeJSON(w, http.StatusCreated, resp)
}

// handleGetCase implements GET /api/v1/cases/{id}.
func (s *Server) handleGetCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	record, found, err := s.Store.Get(r.Context(), id)
	if err != nil {
		s.Logger.Error("store get failed", "error", err, "case_id", shortCaseID(id))
		writeError(w, http.StatusInternalServerError, "could not retrieve case", "")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "case not found", "")
		return
	}
	writeJSON(w, http.StatusOK, caseRecordToResponse(record))
}

// handleDeleteCase implements DELETE /api/v1/cases/{id} — the family's
// right to remove their own record. The ID is the only credential the
// system has, so possession of it is treated as authority to delete. That
// is the same trust model as GET, and it is why the ID is generated with
// crypto/rand and kept out of logs (see RedactPath).
//
// Deletion is permanent. A soft delete would leave the description,
// outcome, and evidence on disk, which defeats the point of offering it.
func (s *Server) handleDeleteCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := s.Store.Delete(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "case not found", "")
		return
	}
	if err != nil {
		s.Logger.Error("store delete failed", "error", err, "case_id", shortCaseID(id))
		writeError(w, http.StatusInternalServerError, "could not delete case", "")
		return
	}
	s.Logger.Info("case deleted", "case_id", shortCaseID(id))
	w.WriteHeader(http.StatusNoContent)
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
		s.Logger.Error("store get failed", "error", err, "case_id", shortCaseID(id))
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
		s.Logger.Error("document generation failed", "error", err, "case_id", shortCaseID(id))
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
		s.Logger.Error("store get failed", "error", err, "case_id", shortCaseID(id))
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
		s.Logger.Error("failed to append evidence", "error", err, "case_id", shortCaseID(id))
		writeError(w, http.StatusInternalServerError, "could not save evidence", "")
		return
	}

	writeJSON(w, http.StatusOK, caseRecordToResponse(updated))
}

// handlePackages implements GET /api/v1/packages — a deterministic,
// no-LLM search over the HBP reference data, used by the public package
// search page.
//
// Query parameters:
//
//	q          free text, optional; empty lists everything
//	specialty  exact specialty filter, optional
//	limit      1-100, default 25
//	offset     0-10000, default 0
//
// Rates are only ever returned for records marked Verified. The dataset
// still carries placeholder figures on unverified rows because the
// backend's own arithmetic tests need them; exposing those through a
// public endpoint would turn a placeholder into a citation, which is the
// exact failure this project exists to avoid. Unverified rows are
// returned by name so a person can at least recognise the procedure.
func (s *Server) handlePackages(w http.ResponseWriter, r *http.Request) {
	const (
		defaultLimit = 25
		maxLimit     = 100
		maxOffset    = 10000
	)

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	specialty := strings.TrimSpace(r.URL.Query().Get("specialty"))

	limit := defaultLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 || n > maxLimit {
			writeError(w, http.StatusBadRequest, "limit must be an integer between 1 and 100", "")
			return
		}
		limit = n
	}

	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 || n > maxOffset {
			writeError(w, http.StatusBadRequest, "offset must be an integer between 0 and 10000", "")
			return
		}
		offset = n
	}

	matches := retrieval.Search(s.Dataset, query)
	filtered := matches[:0]
	for _, candidate := range matches {
		if specialty != "" && !strings.EqualFold(candidate.Package.Specialty, specialty) {
			continue
		}
		filtered = append(filtered, candidate)
	}

	total := len(filtered)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	results := make([]PackageSearchResult, 0, end-offset)
	for _, candidate := range filtered[offset:end] {
		results = append(results, packageToSearchResult(candidate.Package))
	}

	writeJSON(w, http.StatusOK, PackageSearchResponse{
		Query:       query,
		Specialty:   specialty,
		Specialties: datasetSpecialties(s.Dataset),
		Total:       total,
		Count:       len(results),
		Offset:      offset,
		Results:     results,
	})
}

// packageToSearchResult maps a package to the public search shape and
// strips every rate and the placeholder code from an unverified record.
// Keep this the only path from hbp.Package to the search API, so the
// stripping cannot be forgotten in a future handler.
func packageToSearchResult(p hbp.Package) PackageSearchResult {
	out := PackageSearchResult{
		PackageName:     p.PackageName,
		Specialty:       p.Specialty,
		RequiresPreauth: p.RequiresPreauth,
		Verified:        p.Verified,
	}
	if !p.Verified {
		return out
	}
	out.PackageCode = p.PackageCode
	out.IndicativeRateINR = p.IndicativeRateINR
	out.RateMaxINR = p.RateMaxINR
	out.RateType = p.RateType
	if p.PerDiemRates != nil {
		out.PerDiemRates = &PerDiemRates{
			RoutineWardINR: p.PerDiemRates.RoutineWardINR,
			HDUINR:         p.PerDiemRates.HDUINR,
			ICUNoVentINR:   p.PerDiemRates.ICUNoVentINR,
			ICUVentINR:     p.PerDiemRates.ICUVentINR,
		}
	}
	return out
}

func datasetSpecialties(ds *hbp.Dataset) []string {
	seen := make(map[string]bool, len(ds.Packages))
	out := make([]string, 0, 16)
	for _, p := range ds.Packages {
		if p.Specialty == "" || seen[p.Specialty] {
			continue
		}
		seen[p.Specialty] = true
		out = append(out, p.Specialty)
	}
	sort.Strings(out)
	return out
}

// handleHealth implements GET /api/v1/health — a cheap, dependency-free
// liveness check (no LLM call, no store access) suitable for a load
// balancer or uptime monitor to poll frequently without it costing
// anything.
//
// extraction_ready is deliberately a separate field from status. "The
// process is up" and "intake can actually work" are different facts, and
// conflating them is how a deployment with a missing API key passes its
// health check while every real request fails. status stays "ok" so a
// liveness probe does not restart a healthy process over a configuration
// problem that a restart cannot fix.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ready := true
	if a, ok := s.Extractor.(extract.Availability); ok {
		ready = a.Available()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":            "ok",
		"packages_loaded":   len(s.Dataset.Packages),
		"exclusions_loaded": len(s.Dataset.Exclusions),
		"extraction_ready":  ready,
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
