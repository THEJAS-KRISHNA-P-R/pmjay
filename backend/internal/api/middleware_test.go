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

// testAllow is a test-only convenience wrapping checkAndReserve+consume,
// mirroring the pre-dual-window rate limiter's atomic single-call
// behavior, so most of these tests can stay close to testing one
// decision at a time instead of threading the split API through every
// call site.
func (rl *rateLimiter) testAllow(key string) bool {
	ok, _, _, _ := rl.checkAndReserve(key)
	if ok {
		rl.consume(key)
	}
	return ok
}

func TestRateLimiter_AllowsUpToConfiguredLimit(t *testing.T) {
	rl := newRateLimiter(3, 100000) // generous hour limit — only the minute window should bind here
	for i := 0; i < 3; i++ {
		if !rl.testAllow("client-a") {
			t.Fatalf("request %d should have been allowed within the limit of 3", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := newRateLimiter(3, 100000)
	for i := 0; i < 3; i++ {
		rl.testAllow("client-b")
	}
	if rl.testAllow("client-b") {
		t.Fatal("expected the 4th request in the same window to be blocked")
	}
}

func TestRateLimiter_TracksClientsIndependently(t *testing.T) {
	rl := newRateLimiter(1, 100000)
	if !rl.testAllow("client-c") {
		t.Fatal("client-c's first request should be allowed")
	}
	if !rl.testAllow("client-d") {
		t.Fatal("client-d's first request should be allowed independently of client-c's usage")
	}
	if rl.testAllow("client-c") {
		t.Fatal("client-c's second request should be blocked")
	}
}

func TestRateLimiter_RefillsOverTime(t *testing.T) {
	rl := newRateLimiter(60, 100000) // 1 token/sec on the minute window
	fakeNow := time.Now()
	rl.now = func() time.Time { return fakeNow }

	if !rl.testAllow("client-e") {
		t.Fatal("first request should be allowed")
	}
	// Immediately exhaust remaining tokens isn't needed for 60/min with
	// 1 used — instead, jump time forward by 2 seconds and confirm
	// tokens refilled rather than staying frozen.
	fakeNow = fakeNow.Add(2 * time.Second)
	if !rl.testAllow("client-e") {
		t.Fatal("expected a token to have refilled after 2 seconds at 1 token/sec")
	}
}

func TestRateLimiter_MinuteBucketDoesNotOverfillPastItsOwnLimit(t *testing.T) {
	// Regression test: refillBucket must cap the minute bucket at
	// perMinute tokens and the hour bucket at perHour tokens using the
	// caller-supplied maxTokens, never by inferring which bucket it is
	// from the rate's magnitude. The old inference (ratePerSec < 1 means
	// "this is the hour bucket") is wrong whenever perMinute < 60 —
	// which includes this system's own production default of 1/minute
	// (see config.Load) — because perMinute/60 is then also < 1,
	// misclassifying the MINUTE bucket as the hour one and capping it at
	// perHour instead. With that bug, waiting long enough lets a client
	// burst up to perHour requests at once instead of the intended
	// steady perMinute pace, silently defeating the point of a strict
	// per-minute cost control.
	rl := newRateLimiter(1, 20) // production defaults
	fakeNow := time.Now()
	rl.now = func() time.Time { return fakeNow }

	if !rl.testAllow("client") {
		t.Fatal("first request should be allowed")
	}

	// Long enough for the minute bucket to have refilled far past 1
	// token if the cap were wrong (5 min at 1/60 tokens/sec = 5 tokens'
	// worth of refill if uncapped, correctly clamped to 1 if the cap is
	// right) — but comfortably under bucketStaleAfter (10 min), so the
	// sweep doesn't evict the bucket and mask the bug behind a fresh one.
	fakeNow = fakeNow.Add(5 * time.Minute)

	allowedInBurst := 0
	for i := 0; i < 10; i++ {
		if rl.testAllow("client") {
			allowedInBurst++
		}
	}
	if allowedInBurst > 1 {
		t.Errorf("expected at most 1 request allowed in an immediate burst after refilling under perMinute=1, got %d — the minute bucket refilled past its own limit", allowedInBurst)
	}
}

func TestRateLimiter_HourWindowBindsIndependentlyOfMinuteWindow(t *testing.T) {
	// The other direction: a generous minute allowance must not let a
	// client exceed the (tighter) hour allowance. This is the actual
	// point of having two windows at all — a burst that the minute
	// window alone would wave through is still caught by the hour one.
	rl := newRateLimiter(1000, 2)
	if !rl.testAllow("client") {
		t.Fatal("1st request should be allowed (within both windows)")
	}
	if !rl.testAllow("client") {
		t.Fatal("2nd request should be allowed (within both windows)")
	}
	if rl.testAllow("client") {
		t.Fatal("3rd request should be blocked by the hour window (2/hour) despite the minute window being nowhere near its own limit of 1000")
	}
}

func TestRateLimitMiddleware_Returns429WhenExceeded(t *testing.T) {
	rl := newRateLimiter(1, 100000)
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

func TestLLMSemaphore_LimitsConcurrentAcquires(t *testing.T) {
	sem := newLLMSemaphore(2)
	sem.Acquire()
	sem.Acquire()

	thirdAcquired := make(chan struct{})
	go func() {
		sem.Acquire()
		close(thirdAcquired)
	}()

	select {
	case <-thirdAcquired:
		t.Fatal("expected a 3rd Acquire to block while 2 are already held with a limit of 2")
	case <-time.After(50 * time.Millisecond):
		// expected: still blocked
	}

	sem.Release()
	select {
	case <-thirdAcquired:
		// expected: unblocked after a release freed a slot
	case <-time.After(time.Second):
		t.Fatal("expected the blocked Acquire to proceed after a Release freed a slot")
	}
}

func TestLLMConcurrencyMiddleware_BlocksBeyondLimit(t *testing.T) {
	sem := newLLMSemaphore(1)
	release := make(chan struct{})
	entered := make(chan struct{}, 2)

	handler := llmConcurrencyMiddleware(sem, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-release
		w.WriteHeader(http.StatusOK)
	}))

	firstDone := make(chan struct{})
	go func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
		close(firstDone)
	}()

	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("expected the first request to enter the handler")
	}

	secondDone := make(chan struct{})
	go func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
		close(secondDone)
	}()

	select {
	case <-entered:
		t.Fatal("expected the second request to be blocked by the concurrency limit of 1 while the first is still in-flight")
	case <-time.After(50 * time.Millisecond):
		// expected: still blocked
	}

	close(release) // let both handler invocations proceed past their <-release read
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("expected the second request to enter the handler once the first released its slot")
	}

	<-firstDone
	<-secondDone
}

func TestRateLimiter_RetryAfterReflectsActualDeficitNotHardcoded60(t *testing.T) {
	// Regression test: retryAfterSec must be computed from the bucket's
	// actual token deficit and refill rate, not from lastRefill's
	// distance from now — by the time this calculation runs,
	// refillBucket has already set lastRefill=now moments earlier, so
	// that distance is always ~0 and the old formula always simplified
	// to a hardcoded 60, regardless of how close the bucket actually was
	// to allowing the next request.
	rl := newRateLimiter(60, 100000) // 1 token/sec on the minute window; hour window irrelevant here
	fakeNow := time.Now()
	rl.now = func() time.Time { return fakeNow }

	// Drain the bucket to ~0 tokens.
	for i := 0; i < 60; i++ {
		rl.testAllow("client")
	}

	_, retryAfter, _, _ := rl.checkAndReserve("client")
	if retryAfter > 2 {
		t.Errorf("at 1 token/sec with ~0 tokens remaining, expected retryAfterSec close to 1, got %d (hardcoded-60 bug would show 60)", retryAfter)
	}
}

func TestRateLimiter_RetryAfterUsesTheLongerBindingWindow(t *testing.T) {
	// The two windows are independent gates — a caller needs BOTH to
	// clear, so the correct advice is the LONGER of the two waits, not
	// the shorter. Here only the hour window is exhausted (6 requests
	// drains it; the minute window has plenty left at 60/min), so the
	// wait should reflect the hour window's own refill rate (~600s to
	// go from 0 to 1 token at 6/hour) — a value that doesn't coincide
	// with either the old hardcoded-60 bug or a hypothetical
	// min-instead-of-max mistake, so this actually discriminates
	// between them instead of passing either way.
	rl := newRateLimiter(60, 6)
	fakeNow := time.Now()
	rl.now = func() time.Time { return fakeNow }

	for i := 0; i < 6; i++ {
		rl.testAllow("client")
	}

	_, retryAfter, remMin, remHour := rl.checkAndReserve("client")
	if remHour >= 1 {
		t.Fatalf("test setup broken: expected the hour bucket to be exhausted, got %d remaining", remHour)
	}
	if remMin < 1 {
		t.Fatalf("test setup broken: expected the minute bucket to still have budget, got %d remaining", remMin)
	}
	if retryAfter < 500 || retryAfter > 700 {
		t.Errorf("expected retryAfterSec near 600 (the hour window's own refill time), got %d", retryAfter)
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

func TestCORSMiddleware_SetsVaryOriginRegardlessOfWhetherAllowed(t *testing.T) {
	// A cache sitting in front of this server must never serve one
	// origin's response (with or without CORS headers) to a different
	// origin — Vary: Origin is what tells it not to, and that's true
	// whether the request's origin was on the allowlist or not.
	handler := corsMiddleware([]string{"https://allowed.example"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, origin := range []string{"https://allowed.example", "https://evil.example", ""} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Vary"); got != "Origin" {
			t.Errorf("origin %q: expected Vary: Origin, got %q", origin, got)
		}
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
	rl := newRateLimiter(5, 100000)
	fakeNow := time.Now()
	rl.now = func() time.Time { return fakeNow }

	rl.testAllow("stale-client")
	rl.testAllow("active-client")

	if len(rl.minuteBuckets) != 2 {
		t.Fatalf("expected 2 minute buckets before any sweep, got %d", len(rl.minuteBuckets))
	}
	if len(rl.hourBuckets) != 2 {
		t.Fatalf("expected 2 hour buckets before any sweep, got %d", len(rl.hourBuckets))
	}

	// Jump forward past bucketStaleAfter, touching only active-client —
	// this both triggers a sweep (past sweepInterval) and refreshes
	// active-client's lastRefill so it survives that sweep.
	fakeNow = fakeNow.Add(bucketStaleAfter + time.Minute)
	rl.testAllow("active-client")

	for _, buckets := range []map[string]*bucket{rl.minuteBuckets, rl.hourBuckets} {
		if _, exists := buckets["stale-client"]; exists {
			t.Error("expected stale-client's bucket to be evicted after being untouched past bucketStaleAfter")
		}
		if _, exists := buckets["active-client"]; !exists {
			t.Error("expected active-client's bucket to survive the sweep since it was just touched")
		}
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
	rl := newRateLimiter(5, 100000)
	fakeNow := time.Now()
	rl.now = func() time.Time { return fakeNow }

	countAt := func(n int) int {
		for i := 0; i < n; i++ {
			fakeNow = fakeNow.Add(time.Second)
			rl.testAllow(fmt.Sprintf("varying-key-%d", i))
		}
		return len(rl.minuteBuckets)
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
