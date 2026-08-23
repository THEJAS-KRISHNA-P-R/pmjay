package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/pmjay-advocate/backend/internal/extract"
	"github.com/pmjay-advocate/backend/internal/hbp"
	"github.com/pmjay-advocate/backend/internal/store"
)

func testServer(t *testing.T, fake *extract.FakeClient) (*Server, http.Handler) {
	t.Helper()
	ds, err := hbp.Load()
	if err != nil {
		t.Fatalf("failed to load dataset: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	s := &Server{
		Dataset:   ds,
		Extractor: fake,
		Store:     store.NewMemStore(),
		Logger:    logger,
	}
	router := NewRouter(s, []string{"http://localhost:3000"}, 1000) // high limit; rate limiting tested separately
	return s, router
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func doIntake(t *testing.T, router http.Handler, description string) (*httptest.ResponseRecorder, CaseResponse) {
	t.Helper()
	body, _ := json.Marshal(IntakeRequest{Description: description})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp CaseResponse
	if rec.Code == http.StatusCreated {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response body: %v (body: %s)", err, rec.Body.String())
		}
	}
	return rec, resp
}

// ---------------------------------------------------------------------
// Full pipeline, one scenario per tier — the spec's own worked examples,
// run end to end through real HTTP, real retrieval, real tiering, real
// response building. Only the LLM call itself is faked.
// ---------------------------------------------------------------------

func TestHandleIntake_GreenCase_FullPipeline(t *testing.T) {
	desc := "My mother needs her gallbladder removed, doctor confirmed stones on scan, hospital billing desk just told us PMJAY won't cover it and we need to pay before they'll schedule the surgery."
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		ExtractedSituationSummary: "Gallbladder removal for confirmed stones; hospital demanding upfront payment.",
		Candidates:                []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 93, Reasoning: "clear match"}},
		ExclusionMatches:          []extract.ExclusionMatch{},
		Pending:                   extract.SignalNotApplicable,
	})
	_, router := testServer(t, fake)

	rec, resp := doIntake(t, router, desc)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if resp.Outcome != "green" {
		t.Errorf("expected green outcome, got %q", resp.Outcome)
	}
	if resp.ID == "" {
		t.Error("expected a non-empty case ID")
	}
	if !strings.Contains(resp.CareFirstMessage, "treatment first") {
		t.Errorf("expected care-first message present, got %q", resp.CareFirstMessage)
	}
	if resp.ComplaintText == "" {
		t.Error("expected complaint text for a green case")
	}
	if resp.Citation == "" {
		t.Error("expected a citation for a green case")
	}
}

func TestHandleIntake_AmberCase_PendingPreauth_FullPipeline(t *testing.T) {
	desc := "The hospital did the operation already, but now they're saying the insurance side hasn't cleared it and we have to pay."
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		ExtractedSituationSummary: "Procedure already performed; hospital citing an uncleared preauth status.",
		Candidates:                []extract.CandidateMatch{{PackageCode: "SEED-CARD-001", ConfidenceP: 80, Reasoning: "plausible"}},
		Pending:                   extract.SignalPendingLikely,
	})
	_, router := testServer(t, fake)

	rec, resp := doIntake(t, router, desc)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if resp.Outcome != "amber" {
		t.Errorf("expected amber outcome, got %q", resp.Outcome)
	}
	if resp.ComplaintText != "" {
		t.Error("amber must never include complaint text")
	}
	if !strings.Contains(resp.CareFirstMessage, "treatment first") {
		t.Error("expected care-first message present even on amber")
	}
}

func TestHandleIntake_RedCase_FullPipeline(t *testing.T) {
	desc := "We want a cosmetic nose job for my daughter before her wedding, hospital says PMJAY won't pay."
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		ExtractedSituationSummary: "Cosmetic rhinoplasty request, hospital says not covered.",
		Candidates:                []extract.CandidateMatch{{PackageCode: "UNSPECIFIED", ConfidenceP: 10, Reasoning: "no procedure package fits"}},
		ExclusionMatches:          []extract.ExclusionMatch{{Category: "cosmetic", ConfidenceP: 95, Reasoning: "explicitly cosmetic"}},
		Pending:                   extract.SignalNotApplicable,
	})
	_, router := testServer(t, fake)

	rec, resp := doIntake(t, router, desc)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if resp.Outcome != "red" {
		t.Errorf("expected red outcome, got %q", resp.Outcome)
	}
	if resp.ComplaintText != "" {
		t.Error("red must never include complaint text")
	}
	if !strings.Contains(resp.CareFirstMessage, "treatment first") {
		t.Error("expected care-first message present even on red")
	}
}

func TestHandleIntake_MixedCase_FullPipeline(t *testing.T) {
	desc := "Doctor wants to do a knee replacement but also mentioned a cosmetic scar-reduction procedure at the same time, hospital is billing both as PMJAY-covered."
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		ExtractedSituationSummary: "Knee replacement bundled with a cosmetic add-on in the same bill.",
		Candidates:                []extract.CandidateMatch{{PackageCode: "SEED-ORTHO-001", ConfidenceP: 90, Reasoning: "clear knee replacement"}},
		ExclusionMatches:          []extract.ExclusionMatch{{Category: "cosmetic", ConfidenceP: 82, Reasoning: "scar reduction is cosmetic"}},
		Pending:                   extract.SignalNotApplicable,
	})
	_, router := testServer(t, fake)

	rec, resp := doIntake(t, router, desc)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if resp.Outcome != "mixed" {
		t.Errorf("expected mixed outcome, got %q", resp.Outcome)
	}
	if !strings.Contains(resp.Citation, "Knee") {
		t.Errorf("expected citation to mention the knee package, got %q", resp.Citation)
	}
}

func TestHandleIntake_HandoffCase_FullPipeline(t *testing.T) {
	desc := "They already operated, then said something about a form not being right, then a different person said something about my husband's ID not matching my name on the card, and now they want us to leave and come back tomorrow."
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		ExtractedSituationSummary:      "Multiple separate issues after the procedure was completed.",
		MultipleDistinctIssuesDetected: true,
		FamilyDistressSignal:           true,
		Candidates:                     []extract.CandidateMatch{{PackageCode: "UNSPECIFIED", ConfidenceP: 10, Reasoning: "too tangled"}},
	})
	_, router := testServer(t, fake)

	rec, resp := doIntake(t, router, desc)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if resp.Outcome != "handoff" {
		t.Errorf("expected handoff outcome, got %q", resp.Outcome)
	}
	if resp.HandoffSummary == "" {
		t.Error("expected a non-empty handoff summary")
	}
	if resp.Citation != "" {
		t.Error("handoff should not cite a specific package/exclusion")
	}
}

// ---------------------------------------------------------------------
// Validation and error handling
// ---------------------------------------------------------------------

func TestHandleIntake_RejectsInvalidJSON(t *testing.T) {
	_, router := testServer(t, extract.NewFakeClient())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", strings.NewReader("{not json"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestHandleIntake_RejectsTooShortDescription(t *testing.T) {
	_, router := testServer(t, extract.NewFakeClient())
	rec, _ := doIntake(t, router, "hi")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for a too-short description, got %d", rec.Code)
	}
}

func TestHandleIntake_RejectsTooLongDescription(t *testing.T) {
	_, router := testServer(t, extract.NewFakeClient())
	rec, _ := doIntake(t, router, strings.Repeat("a", maxDescriptionLength+1))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for a too-long description, got %d", rec.Code)
	}
}

func TestHandleIntake_RejectsOversizedBody(t *testing.T) {
	_, router := testServer(t, extract.NewFakeClient())
	// A well-formed request whose JSON-encoded description alone exceeds
	// maxRequestBodyBytes — this must be rejected on wire size before
	// the maxDescriptionLength check (which runs on the decoded field,
	// after the whole body would already have been read) ever gets a
	// chance to run.
	oversized := strings.Repeat("a", maxRequestBodyBytes*2)
	rec, _ := doIntake(t, router, oversized)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 for an oversized body, got %d", rec.Code)
	}
}

func TestHandleIntake_AcceptsBodyRightAtTheSizeLimit(t *testing.T) {
	// Guards against an off-by-one that would reject legitimate
	// maximum-length requests — maxDescriptionLength characters, once
	// JSON-encoded with its field name and quoting, must still fit
	// under maxRequestBodyBytes.
	desc := strings.Repeat("a", maxDescriptionLength)
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		Candidates: []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 90, Reasoning: "x"}},
		Pending:    extract.SignalNotApplicable,
	})
	_, router := testServer(t, fake)
	rec, _ := doIntake(t, router, desc)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected a description at exactly maxDescriptionLength to be accepted, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestHandleAddEvidence_RejectsOversizedBody(t *testing.T) {
	desc := "family reports a denied claim for surgery, needs evidence captured"
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		Candidates: []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 90, Reasoning: "x"}},
		Pending:    extract.SignalNotApplicable,
	})
	_, router := testServer(t, fake)
	createRec, created := doIntake(t, router, desc)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: failed to create case, got %d (body: %s)", createRec.Code, createRec.Body.String())
	}

	oversizedNote := `{"note":"` + strings.Repeat("a", maxRequestBodyBytes*2) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases/"+created.ID+"/evidence", strings.NewReader(oversizedNote))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 for an oversized evidence body, got %d", rec.Code)
	}
}

func TestHandleIntake_ExtractorFailure_ReturnsCareFirstFallback(t *testing.T) {
	// The infrastructure-failure case: Section 10's care-first rule is
	// absolute, so even a downstream API failure must still hand the
	// family something actionable (the helpline), not a bare error.
	desc := "a description the fake will fail on for this specific test"
	fake := extract.NewFakeClient()
	fake.RegisterError(desc, fmt.Errorf("simulated upstream API failure"))
	_, router := testServer(t, fake)

	body, _ := json.Marshal(IntakeRequest{Description: desc})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 for an extractor failure, got %d: %s", rec.Code, rec.Body.String())
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if !strings.Contains(errResp.FallbackGuidance, "treatment first") {
		t.Errorf("expected fallback guidance to include the care-first message, got %q", errResp.FallbackGuidance)
	}
	if !strings.Contains(errResp.FallbackGuidance, "14555") {
		t.Errorf("expected fallback guidance to include the helpline number, got %q", errResp.FallbackGuidance)
	}
}

// ---------------------------------------------------------------------
// Case retrieval and evidence capture
// ---------------------------------------------------------------------

func TestHandleIntake_CaseIsRetrievableAfterCreation(t *testing.T) {
	desc := "clear gallbladder stones case for retrieval test, doctor confirmed on scan"
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		ExtractedSituationSummary: "Gallbladder case.",
		Candidates:                []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 90, Reasoning: "x"}},
		Pending:                   extract.SignalNotApplicable,
	})
	_, router := testServer(t, fake)

	_, created := doIntake(t, router, desc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/"+created.ID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 fetching an existing case, got %d", rec.Code)
	}
	var fetched CaseResponse
	json.Unmarshal(rec.Body.Bytes(), &fetched)
	if fetched.ID != created.ID {
		t.Errorf("fetched case ID mismatch: got %q, want %q", fetched.ID, created.ID)
	}
}

func TestHandleGetCase_NotFound(t *testing.T) {
	_, router := testServer(t, extract.NewFakeClient())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/does-not-exist", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for a nonexistent case, got %d", rec.Code)
	}
}

func TestHandleGetCaseDocument_NotFound(t *testing.T) {
	_, router := testServer(t, extract.NewFakeClient())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/does-not-exist/document", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for a nonexistent case's document, got %d", rec.Code)
	}
}

func TestHandleGetCaseDocument_Success(t *testing.T) {
	desc := "gallbladder case for document test, doctor confirmed stones need removal"
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		Candidates: []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 90, Reasoning: "x"}},
		Pending:    extract.SignalNotApplicable,
	})
	_, router := testServer(t, fake)
	_, created := doIntake(t, router, desc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/"+created.ID+"/document", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %q", ct)
	}
	wantDisposition := fmt.Sprintf(`inline; filename="pmjay-case-%s.pdf"`, created.ID)
	if cd := rec.Header().Get("Content-Disposition"); cd != wantDisposition {
		t.Errorf("expected Content-Disposition %q, got %q", wantDisposition, cd)
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF-1.4")) {
		t.Error("expected response body to be a valid PDF (start with %PDF-1.4 header)")
	}
	if cl := rec.Header().Get("Content-Length"); cl != strconv.Itoa(rec.Body.Len()) {
		t.Errorf("Content-Length header %q does not match actual body length %d", cl, rec.Body.Len())
	}
	// The case ID appears in the PDF itself (header + footer, see
	// internal/document's tests) — a light end-to-end check that this
	// handler is really passing the right record through, not some
	// fixed/stale sample.
	if !bytes.Contains(rec.Body.Bytes(), []byte(created.ID)) {
		t.Error("expected the case ID to appear within the generated PDF content")
	}
}

func TestHandleGetCaseDocument_MethodNotAllowed(t *testing.T) {
	// The document endpoint is read-only by design (see its handler doc
	// comment) — confirm the router actually enforces GET-only rather
	// than that being true only by omission.
	_, router := testServer(t, extract.NewFakeClient())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases/some-id/document", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST to the document endpoint, got %d", rec.Code)
	}
}

func TestHandleAddEvidence_Success(t *testing.T) {
	desc := "gallbladder case for evidence test, doctor confirmed stones"
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		Candidates: []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 90, Reasoning: "x"}},
		Pending:    extract.SignalNotApplicable,
	})
	_, router := testServer(t, fake)
	_, created := doIntake(t, router, desc)

	evReq := AddEvidenceRequest{StaffName: "Reception Desk A", ApproxTime: "3:30 PM", Note: "verbal denial"}
	body, _ := json.Marshal(evReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases/"+created.ID+"/evidence", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var updated CaseResponse
	json.Unmarshal(rec.Body.Bytes(), &updated)
	if len(updated.Evidence) != 1 {
		t.Fatalf("expected 1 evidence entry, got %d", len(updated.Evidence))
	}
	if updated.Evidence[0].StaffName != "Reception Desk A" {
		t.Errorf("evidence content mismatch: %+v", updated.Evidence[0])
	}
}

func TestHandleAddEvidence_NotFound(t *testing.T) {
	_, router := testServer(t, extract.NewFakeClient())
	body, _ := json.Marshal(AddEvidenceRequest{Note: "x"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases/does-not-exist/evidence", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 adding evidence to a nonexistent case, got %d", rec.Code)
	}
}

func TestHandleAddEvidence_RejectsCompletelyEmptyBody(t *testing.T) {
	desc := "gallbladder case for empty evidence test, confirmed stones"
	fake := extract.NewFakeClient()
	fake.Register(desc, extract.Result{
		Candidates: []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 90, Reasoning: "x"}},
		Pending:    extract.SignalNotApplicable,
	})
	_, router := testServer(t, fake)
	_, created := doIntake(t, router, desc)

	body, _ := json.Marshal(AddEvidenceRequest{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases/"+created.ID+"/evidence", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for a completely empty evidence body, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------
// Health check
// ---------------------------------------------------------------------

func TestHandleHealth(t *testing.T) {
	_, router := testServer(t, extract.NewFakeClient())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body)
	}
	if body["packages_loaded"] == "0" {
		t.Error("expected a non-zero package count reported")
	}
}

func TestHandleHealth_DoesNotCallExtractor(t *testing.T) {
	// The health check must be free to poll frequently — it should never
	// trigger a paid LLM call.
	fake := extract.NewFakeClient()
	_, router := testServer(t, fake)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if len(fake.CallLog) != 0 {
		t.Errorf("expected zero extractor calls from a health check, got %d", len(fake.CallLog))
	}
}
