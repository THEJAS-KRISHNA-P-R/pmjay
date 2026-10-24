package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pmjay-advocate/backend/internal/extract"
	"github.com/pmjay-advocate/backend/internal/hbp"
	"github.com/pmjay-advocate/backend/internal/store"
)

// hardening_test.go covers the behaviours added for public release:
// package search, deletion, honest health reporting, degraded intake,
// request timeouts, and the privacy headers every response carries.

func testServerWith(t *testing.T, ex extract.Extractor, opts Options) (*Server, http.Handler) {
	t.Helper()
	ds, err := hbp.Load()
	if err != nil {
		t.Fatalf("load dataset: %v", err)
	}
	s := &Server{
		Dataset:   ds,
		Extractor: ex,
		Store:     store.NewMemStore(),
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	if opts.IntakeRatePerMinute == 0 {
		opts.IntakeRatePerMinute = 1000
		opts.IntakeRatePerHour = 1000
		opts.MaxConcurrentLLM = 1000
		opts.GeneralRatePerMinute = 1000
		opts.GeneralRatePerHour = 1000
	}
	if len(opts.AllowedOrigins) == 0 {
		opts.AllowedOrigins = []string{"http://localhost:3000"}
	}
	return s, NewRouter(s, opts)
}

func doJSON(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestPackages_SearchFindsVerifiedRate(t *testing.T) {
	_, router := testServerWith(t, extract.NewFakeClient(), Options{})
	rec := doJSON(t, router, http.MethodGet, "/api/v1/packages?q=dialysis&limit=5", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp PackageSearchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total == 0 || len(resp.Results) == 0 {
		t.Fatal("expected dialysis results")
	}
	if len(resp.Specialties) == 0 {
		t.Error("expected the specialty filter list to be populated")
	}

	var withRate bool
	for _, r := range resp.Results {
		if r.Verified && r.IndicativeRateINR > 0 {
			withRate = true
			break
		}
	}
	if !withRate {
		t.Error("expected at least one verified result with a rate")
	}
}

func TestPackages_UnverifiedRecordNeverExposesARate(t *testing.T) {
	_, router := testServerWith(t, extract.NewFakeClient(), Options{})
	rec := doJSON(t, router, http.MethodGet, "/api/v1/packages?q=cholecystectomy", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp PackageSearchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Results) == 0 {
		t.Fatal("expected the gallbladder package to be recognisable by name")
	}
	for _, r := range resp.Results {
		if r.Verified {
			continue
		}
		if r.IndicativeRateINR != 0 || r.RateMaxINR != 0 || r.RateType != "" || r.PerDiemRates != nil {
			t.Errorf("unverified package %s leaked rate data: %+v", r.PackageCode, r)
		}
		// A code is a citation too. The dataset's internal placeholder
		// (SEED-...) must not be published as if it were a government
		// code; see PackageSearchResult's doc comment.
		if r.PackageCode != "" {
			t.Errorf("unverified package %q leaked a package_code %q", r.PackageName, r.PackageCode)
		}
	}
}

func TestPackages_UnverifiedRecordOmitsTheCodeInRawJSON(t *testing.T) {
	// The struct field being empty is not quite the same guarantee as the
	// key being absent, and the key is what a client actually reads.
	_, router := testServerWith(t, extract.NewFakeClient(), Options{})
	rec := doJSON(t, router, http.MethodGet, "/api/v1/packages?q=cholecystectomy", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, `"package_code":""`) {
		t.Error("expected package_code to be omitted, not emitted empty")
	}
	if strings.Contains(body, "SEED-") {
		t.Error("placeholder SEED- identifier appeared in a public API response")
	}
}

func TestPackages_PaginationAndSpecialtyFilter(t *testing.T) {
	_, router := testServerWith(t, extract.NewFakeClient(), Options{})

	all := doJSON(t, router, http.MethodGet, "/api/v1/packages?limit=100", nil)
	var page PackageSearchResponse
	if err := json.Unmarshal(all.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if page.Total <= page.Count {
		t.Fatalf("expected more total records than one page holds, total=%d count=%d", page.Total, page.Count)
	}

	specialty := page.Results[0].Specialty
	filtered := doJSON(t, router, http.MethodGet, "/api/v1/packages?limit=100&specialty="+url.QueryEscape(specialty), nil)
	var filteredPage PackageSearchResponse
	if err := json.Unmarshal(filtered.Body.Bytes(), &filteredPage); err != nil {
		t.Fatalf("decode filtered: %v", err)
	}
	if filteredPage.Total == 0 {
		t.Fatalf("expected results for specialty %q", specialty)
	}
	if filteredPage.Total >= page.Total {
		t.Errorf("specialty filter did not narrow the result set: %d vs %d", filteredPage.Total, page.Total)
	}
	for _, r := range filteredPage.Results {
		if !strings.EqualFold(r.Specialty, specialty) {
			t.Errorf("filter returned %q for specialty %q", r.Specialty, specialty)
		}
	}
}

func TestPackages_RejectsBadLimit(t *testing.T) {
	_, router := testServerWith(t, extract.NewFakeClient(), Options{})
	for _, query := range []string{"limit=0", "limit=101", "limit=nope", "offset=-1"} {
		rec := doJSON(t, router, http.MethodGet, "/api/v1/packages?"+query, nil)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", query, rec.Code)
		}
	}
}

func TestDeleteCase_RemovesThenReturns404(t *testing.T) {
	fake := extract.NewFakeClient()
	desc := "Hospital says my mother's dialysis is not covered under PMJAY."
	fake.Register(desc, extract.Result{
		Candidates: []extract.CandidateMatch{{PackageCode: "MG072A", ConfidenceP: 90}},
		Pending:    extract.SignalNotApplicable,
	})
	_, router := testServerWith(t, fake, Options{})

	created := doJSON(t, router, http.MethodPost, "/api/v1/cases", IntakeRequest{Description: desc})
	if created.Code != http.StatusCreated {
		t.Fatalf("intake: %d %s", created.Code, created.Body.String())
	}
	var resp CaseResponse
	if err := json.Unmarshal(created.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Persisted {
		t.Fatal("expected a successfully stored case to report persisted=true")
	}

	del := doJSON(t, router, http.MethodDelete, "/api/v1/cases/"+resp.ID, nil)
	if del.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", del.Code, del.Body.String())
	}

	get := doJSON(t, router, http.MethodGet, "/api/v1/cases/"+resp.ID, nil)
	if get.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", get.Code)
	}

	again := doJSON(t, router, http.MethodDelete, "/api/v1/cases/"+resp.ID, nil)
	if again.Code != http.StatusNotFound {
		t.Errorf("expected 404 deleting an already-deleted case, got %d", again.Code)
	}
}

func TestHealth_ReportsExtractionReadiness(t *testing.T) {
	_, withExtractor := testServerWith(t, extract.NewFakeClient(), Options{})
	rec := doJSON(t, withExtractor, http.MethodGet, "/api/v1/health", nil)
	var healthy map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &healthy); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if isReady, ok := healthy["extraction_ready"].(bool); !ok || !isReady {
		t.Errorf("expected extraction_ready=true with a working extractor, got %v", healthy["extraction_ready"])
	}

	_, withoutKey := testServerWith(t, extract.Unavailable{Provider: "anthropic", EnvVar: "ANTHROPIC_API_KEY"}, Options{})
	rec = doJSON(t, withoutKey, http.MethodGet, "/api/v1/health", nil)
	healthy = nil
	if err := json.Unmarshal(rec.Body.Bytes(), &healthy); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if isReady, ok := healthy["extraction_ready"].(bool); !ok || isReady {
		t.Errorf("expected extraction_ready=false with no provider key, got %v", healthy["extraction_ready"])
	}
}

func TestIntake_MissingProviderKeyReturns503WithCareFirstGuidance(t *testing.T) {
	_, router := testServerWith(t, extract.Unavailable{Provider: "gemini", EnvVar: "GEMINI_API_KEY"}, Options{})
	rec := doJSON(t, router, http.MethodPost, "/api/v1/cases", IntakeRequest{
		Description: "The hospital will not treat my child under PMJAY without cash.",
	})
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.Contains(errResp.FallbackGuidance, "14555") {
		t.Errorf("expected the care-first fallback to name the helpline, got %q", errResp.FallbackGuidance)
	}
}

// blockingExtractor waits for the request context to expire, which is
// what a genuinely slow provider call looks like from the handler's
// point of view.
type blockingExtractor struct{}

func (blockingExtractor) Extract(ctx context.Context, _ string, _ []extract.CandidatePackageInfo, _ []extract.CandidateExclusionInfo) (extract.Result, error) {
	<-ctx.Done()
	return extract.Result{}, ctx.Err()
}

func TestIntake_TimeoutReturns504NotATruncatedConnection(t *testing.T) {
	_, router := testServerWith(t, blockingExtractor{}, Options{IntakeTimeout: 20 * time.Millisecond})
	start := time.Now()
	rec := doJSON(t, router, http.MethodPost, "/api/v1/cases", IntakeRequest{
		Description: "My mother is waiting for a covered procedure and the desk will not answer.",
	})
	elapsed := time.Since(start)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d: %s", rec.Code, rec.Body.String())
	}
	if elapsed > 2*time.Second {
		t.Errorf("request was not bounded by IntakeTimeout, took %v", elapsed)
	}
	if !strings.Contains(rec.Body.String(), "14555") {
		t.Error("timeout response must still carry the care-first helpline fallback")
	}
}

func TestResponses_CarryNoStoreAndSecurityHeaders(t *testing.T) {
	_, router := testServerWith(t, extract.NewFakeClient(), Options{})
	for _, path := range []string{"/api/v1/health", "/api/v1/packages?q=dialysis"} {
		rec := doJSON(t, router, http.MethodGet, path, nil)
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("%s: expected Cache-Control no-store, got %q", path, got)
		}
		if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s: expected nosniff, got %q", path, got)
		}
		if got := rec.Header().Get("Referrer-Policy"); got != "no-referrer" {
			t.Errorf("%s: expected no-referrer, got %q", path, got)
		}
		if got := rec.Header().Get("Content-Security-Policy"); !strings.Contains(got, "default-src 'none'") {
			t.Errorf("%s: expected restrictive CSP, got %q", path, got)
		}
	}
}

func TestIntakeError_DoesNotLeakInternals(t *testing.T) {
	_, router := testServerWith(t, extract.Unavailable{Provider: "anthropic"}, Options{})
	rec := doJSON(t, router, http.MethodPost, "/api/v1/cases", IntakeRequest{
		Description: "A family is being asked to pay for a covered procedure today.",
	})
	body := rec.Body.String()
	for _, leak := range []string{"ANTHROPIC", "sk-", "goroutine", "panic"} {
		if strings.Contains(body, leak) {
			t.Errorf("response leaked %q: %s", leak, body)
		}
	}
}

func TestRedactPath_RemovesCaseIDs(t *testing.T) {
	cases := map[string]string{
		"/api/v1/cases/0f8fad5b-d9cb-469f-a165-70867728950e":          "/api/v1/cases/{id}",
		"/api/v1/cases/0f8fad5b-d9cb-469f-a165-70867728950e/document": "/api/v1/cases/{id}/document",
		"/api/v1/cases/abc/evidence":                                  "/api/v1/cases/{id}/evidence",
		"/api/v1/packages":                                            "/api/v1/packages",
		"/api/v1/health":                                              "/api/v1/health",
		"/api/v1/cases/":                                              "/api/v1/cases/",
	}
	for in, want := range cases {
		if got := RedactPath(in); got != want {
			t.Errorf("RedactPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestShortCaseID_IsNotUsableAsACapability(t *testing.T) {
	id := "0f8fad5b-d9cb-469f-a165-70867728950e"
	short := shortCaseID(id)
	if short == id {
		t.Fatal("shortCaseID returned the full ID; logs would contain a working capability")
	}
	if len([]rune(short)) > 10 {
		t.Errorf("shortCaseID returned %q, longer than intended", short)
	}
	if !strings.HasPrefix(id, strings.TrimSuffix(short, "…")) {
		t.Errorf("expected the short form to keep a recognisable prefix, got %q", short)
	}
}
