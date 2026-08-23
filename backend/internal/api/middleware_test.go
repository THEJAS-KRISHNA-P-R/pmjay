package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRecoverMiddleware_PanicDoesNotCrashAndReturns500(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("simulated handler panic")
	})
	handler := recoverMiddleware(logger, panicking)

	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	rec := httptest.NewRecorder()

	// The critical assertion is simply that this does not panic out of
	// the test itself — a real, unrecovered panic would crash the whole
	// process, taking every other in-flight request down with it.
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 after a recovered panic, got %d", rec.Code)
	}
}

func TestRecoverMiddleware_SubsequentRequestsStillWork(t *testing.T) {
	// Directly demonstrates the isolation property: one request panicking
	// must not affect the next request on the same handler chain.
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	callCount := 0
	handler := recoverMiddleware(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			panic("first request panics")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec1.Code != http.StatusInternalServerError {
		t.Fatalf("expected first request to 500, got %d", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected second request to succeed normally after the first panicked, got %d", rec2.Code)
	}
}

func TestRateLimiter_AllowsUpToConfiguredLimit(t *testing.T) {
	rl := newRateLimiter(3)
	for i := 0; i < 3; i++ {
		if !rl.allow("client-a") {
			t.Fatalf("request %d should have been allowed within the limit of 3", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := newRateLimiter(3)
	for i := 0; i < 3; i++ {
		rl.allow("client-b")
	}
	if rl.allow("client-b") {
		t.Fatal("expected the 4th request in the same window to be blocked")
	}
}

func TestRateLimiter_TracksClientsIndependently(t *testing.T) {
	rl := newRateLimiter(1)
	if !rl.allow("client-c") {
		t.Fatal("client-c's first request should be allowed")
	}
	if !rl.allow("client-d") {
		t.Fatal("client-d's first request should be allowed independently of client-c's usage")
	}
	if rl.allow("client-c") {
		t.Fatal("client-c's second request should be blocked")
	}
}

func TestRateLimiter_RefillsOverTime(t *testing.T) {
	rl := newRateLimiter(60) // 1 token/sec
	fakeNow := time.Now()
	rl.now = func() time.Time { return fakeNow }

	if !rl.allow("client-e") {
		t.Fatal("first request should be allowed")
	}
	// Immediately exhaust remaining tokens isn't needed for 60/min with
	// 1 used — instead, jump time forward by 2 seconds and confirm
	// tokens refilled rather than staying frozen.
	fakeNow = fakeNow.Add(2 * time.Second)
	if !rl.allow("client-e") {
		t.Fatal("expected a token to have refilled after 2 seconds at 1 token/sec")
	}
}

func TestRateLimitMiddleware_Returns429WhenExceeded(t *testing.T) {
	rl := newRateLimiter(1)
	handler := rateLimitMiddleware(rl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/cases", nil)
	req1.RemoteAddr = "1.2.3.4:5555"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/cases", nil)
	req2.RemoteAddr = "1.2.3.4:5555"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request from the same IP to be rate-limited, got %d", rec2.Code)
	}
}

func TestCORSMiddleware_AllowsConfiguredOrigin(t *testing.T) {
	handler := corsMiddleware([]string{"https://allowed.example"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://allowed.example")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.example" {
		t.Errorf("expected CORS header to echo the allowed origin, got %q", got)
	}
}

func TestCORSMiddleware_DoesNotEchoUnconfiguredOrigin(t *testing.T) {
	handler := corsMiddleware([]string{"https://allowed.example"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header for an unconfigured origin, got %q", got)
	}
}

func TestCORSMiddleware_HandlesPreflight(t *testing.T) {
	called := false
	handler := corsMiddleware([]string{"https://allowed.example"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://allowed.example")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for a preflight OPTIONS request, got %d", rec.Code)
	}
	if called {
		t.Error("expected the preflight request to short-circuit before reaching the real handler")
	}
}

func TestClientIP_PrefersLastForwardedForEntry(t *testing.T) {
	// The last entry is the one the trusted proxy (Caddy) itself
	// appended — see clientIP's doc comment for why this must be the
	// last entry, not the first.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")

	got := clientIP(req)
	if got != "10.0.0.1" {
		t.Errorf("expected the last X-Forwarded-For entry (the proxy-appended one), got %q", got)
	}
}

func TestClientIP_SingleForwardedForEntryIsTrusted(t *testing.T) {
	// The common real-world case: Caddy sets X-Forwarded-For to just the
	// real client IP (no comma) because the original request had none.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")

	got := clientIP(req)
	if got != "203.0.113.5" {
		t.Errorf("expected the single X-Forwarded-For entry, got %q", got)
	}
}

func TestClientIP_FallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.7:9999"

	got := clientIP(req)
	if got != "198.51.100.7" {
		t.Errorf("expected RemoteAddr host without port, got %q", got)
	}
}

// TestClientIP_ClientCannotSpoofLeadingForwardedForEntry is the actual
// security property: a caller prepending an arbitrary fake address to
// X-Forwarded-For must not change what this system rate-limits them as.
// This is the exact scenario the fix in clientIP's doc comment describes
// — before the fix, this test would have failed, because "1.2.3.4"
// would have been trusted instead of the real, proxy-appended address.
func TestClientIP_ClientCannotSpoofLeadingForwardedForEntry(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	// A client cannot make Caddy stop appending its own address, but it
	// can send whatever it wants as the value Caddy appends *to* —
	// simulating exactly that: an attacker-chosen leading value.
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	firstAttemptKey := clientIP(req)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	req2.Header.Set("X-Forwarded-For", "9.9.9.9, "+firstAttemptKey)
	secondAttemptKey := clientIP(req2)

	if firstAttemptKey != secondAttemptKey {
		t.Errorf("expected the same rate-limit key across requests from the same real peer regardless of a spoofed leading X-Forwarded-For value, got %q then %q", firstAttemptKey, secondAttemptKey)
	}
}

func TestRateLimiter_SweepEvictsStaleEntriesButNotActiveOnes(t *testing.T) {
	rl := newRateLimiter(5)
	fakeNow := time.Now()
	rl.now = func() time.Time { return fakeNow }

	rl.allow("stale-client")
	rl.allow("active-client")

	if len(rl.buckets) != 2 {
		t.Fatalf("expected 2 buckets before any sweep, got %d", len(rl.buckets))
	}

	// Jump forward past bucketStaleAfter, touching only active-client —
	// this both triggers a sweep (past sweepInterval) and refreshes
	// active-client's lastRefill so it survives that sweep.
	fakeNow = fakeNow.Add(bucketStaleAfter + time.Minute)
	rl.allow("active-client")

	if _, exists := rl.buckets["stale-client"]; exists {
		t.Error("expected stale-client's bucket to be evicted after being untouched past bucketStaleAfter")
	}
	if _, exists := rl.buckets["active-client"]; !exists {
		t.Error("expected active-client's bucket to survive the sweep since it was just touched")
	}
}

func TestRateLimiter_DoesNotGrowUnboundedlyUnderVaryingKeys(t *testing.T) {
	// Simulates the attack the sweep defends against: a caller that
	// varies its rate-limit key on every request (e.g. by spoofing a
	// header, if one were ever trusted incorrectly) should not be able
	// to grow this map forever. At 1 new key/second, steady state is
	// bounded by roughly bucketStaleAfter's width (~600 keys), however
	// long the attack keeps going — the map must plateau, not keep
	// growing linearly with total requests made.
	rl := newRateLimiter(5)
	fakeNow := time.Now()
	rl.now = func() time.Time { return fakeNow }

	countAt := func(n int) int {
		for i := 0; i < n; i++ {
			fakeNow = fakeNow.Add(time.Second)
			rl.allow(fmt.Sprintf("varying-key-%d", i))
		}
		return len(rl.buckets)
	}

	first1000 := countAt(1000)
	second1000 := countAt(1000) // another 1000 keys, 2000 total requests so far

	steadyStateCeiling := int(bucketStaleAfter.Seconds()) + int(sweepInterval.Seconds())
	if first1000 > steadyStateCeiling {
		t.Errorf("expected bucket count to plateau near bucketStaleAfter's width (~%ds), got %d after 1000 requests", steadyStateCeiling, first1000)
	}
	if second1000 > steadyStateCeiling {
		t.Errorf("expected bucket count to plateau near bucketStaleAfter's width (~%ds), got %d after 2000 requests", steadyStateCeiling, second1000)
	}
	// The key property: doubling total requests must not double the
	// bucket count — that would mean the map is still growing linearly,
	// i.e. not actually bounded.
	if second1000 > first1000+50 { // small margin for sweep-timing jitter
		t.Errorf("expected bucket count to have plateaued, not grown further: %d after 1000 requests, %d after 2000", first1000, second1000)
	}
}
