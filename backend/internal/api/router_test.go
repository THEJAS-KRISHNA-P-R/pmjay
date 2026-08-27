package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pmjay-advocate/backend/internal/extract"
	"github.com/pmjay-advocate/backend/internal/hbp"
	"github.com/pmjay-advocate/backend/internal/store"
)

// routerWithLimits builds a server + router with independently
// controllable limits, so tests can drive the general limiter to its
// threshold without needing hundreds of requests.
func routerWithLimits(t *testing.T, intakePerMinute, intakePerHour, maxConcurrentLLM, generalPerMinute, generalPerHour int) (http.Handler, string) {
	t.Helper()
	ds, err := hbp.Load()
	if err != nil {
		t.Fatalf("failed to load dataset: %v", err)
	}
	fake := extract.NewFakeClient()
	desc := "gallbladder case for router rate-limit test, confirmed stones on scan"
	fake.Register(desc, extract.Result{
		Candidates: []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 90, Reasoning: "x"}},
		Pending:    extract.SignalNotApplicable,
	})
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	s := &Server{
		Dataset:   ds,
		Extractor: fake,
		Store:     store.NewMemStore(),
		Logger:    logger,
	}
	router := NewRouter(s, []string{"http://localhost:3000"}, intakePerMinute, intakePerHour, maxConcurrentLLM, generalPerMinute, generalPerHour)

	// Seed one case directly through the store so read/evidence
	// endpoints have something real to hit, without spending the
	// intake limiter's budget on setup.
	rec, resp := doIntakeWithLimiter(t, router, desc)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup intake failed: %d: %s", rec.Code, rec.Body.String())
	}
	return router, resp.ID
}

func doIntakeWithLimiter(t *testing.T, router http.Handler, description string) (*httptest.ResponseRecorder, CaseResponse) {
	t.Helper()
	body, _ := json.Marshal(IntakeRequest{Description: description})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	var resp CaseResponse
	if rec.Code == http.StatusCreated {
		json.Unmarshal(rec.Body.Bytes(), &resp)
	}
	return rec, resp
}

func TestNewRouter_GeneralRateLimitAppliesToCaseRead(t *testing.T) {
	router, id := routerWithLimits(t, 1000, 1000, 1000, 2, 1000)

	var lastCode int
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/"+id, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		lastCode = rec.Code
	}
	if lastCode != http.StatusTooManyRequests {
		t.Errorf("expected the case-read endpoint to eventually rate-limit a single IP, last status was %d", lastCode)
	}
}

func TestNewRouter_GeneralRateLimitAppliesToDocument(t *testing.T) {
	router, id := routerWithLimits(t, 1000, 1000, 1000, 2, 1000)

	var lastCode int
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/"+id+"/document", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		lastCode = rec.Code
	}
	if lastCode != http.StatusTooManyRequests {
		t.Errorf("expected the document endpoint to eventually rate-limit a single IP, last status was %d", lastCode)
	}
}

func TestNewRouter_GeneralRateLimitAppliesToEvidence(t *testing.T) {
	router, id := routerWithLimits(t, 1000, 1000, 1000, 2, 1000)

	var lastCode int
	for i := 0; i < 5; i++ {
		body, _ := json.Marshal(AddEvidenceRequest{Note: "x"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cases/"+id+"/evidence", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		lastCode = rec.Code
	}
	if lastCode != http.StatusTooManyRequests {
		t.Errorf("expected the evidence endpoint to eventually rate-limit a single IP, last status was %d", lastCode)
	}
}

func TestNewRouter_GeneralRateLimitEndpointsShareOneLimiter(t *testing.T) {
	// The three endpoints are documented as sharing ONE generalRL
	// instance, not three independent ones — spending budget on reads
	// must count against the document and evidence endpoints too.
	router, id := routerWithLimits(t, 1000, 1000, 1000, 3, 1000)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/"+id, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("read %d: expected 200 while under budget, got %d", i, rec.Code)
		}
	}

	// Budget of 3 is now spent entirely on reads — the document
	// endpoint should be blocked immediately, not get its own fresh 3.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/"+id+"/document", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected the document endpoint to be blocked by budget already spent on reads (shared limiter), got %d", rec.Code)
	}
}

func TestNewRouter_HealthCheckNeverRateLimited(t *testing.T) {
	// The general limiter must not reach the health check — a monitor
	// polling it frequently should never see a 429.
	router, _ := routerWithLimits(t, 1000, 1000, 1000, 1, 1)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("health check request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

func TestNewRouter_IntakeLimitIndependentOfGeneralLimit(t *testing.T) {
	// A generous general limit must not loosen intake's own, stricter,
	// cost-control limit.
	router, _ := routerWithLimits(t, 1, 1000, 1000, 1000, 1000)
	desc := "second distinct description so the fake client sees a fresh call, gallbladder stones"

	body, _ := json.Marshal(IntakeRequest{Description: desc})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected intake's own limit (1/min, already spent by setup) to block this request, got %d", rec.Code)
	}
}
