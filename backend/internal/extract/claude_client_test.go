package extract

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// instantSleep replaces real backoff waits in tests: it returns an
// already-closed channel regardless of the requested duration, so retry
// tests exercise production retryDelays/maxRetryAfter values without the
// test suite actually pausing for them.
func instantSleep(time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	return ch
}

func testCandidates() []CandidatePackageInfo {
	return []CandidatePackageInfo{
		{PackageCode: "SEED-GS-001", PackageName: "Laparoscopic Cholecystectomy", Specialty: "General Surgery", Keywords: []string{"gallbladder"}},
	}
}

func testExclusionCandidates() []CandidateExclusionInfo {
	return []CandidateExclusionInfo{
		{Category: "cosmetic", DisplayName: "Cosmetic Surgery", Description: "not covered", Keywords: []string{"cosmetic"}},
	}
}

// mockToolResponse builds a fake Anthropic API response containing a
// tool_use block, shaped exactly like the real API's response format.
func mockToolResponse(t *testing.T, payload toolExtractionPayload) messagesResponse {
	t.Helper()
	inputBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshalling test payload: %v", err)
	}
	return messagesResponse{
		ID:         "msg_test123",
		StopReason: "tool_use",
		Content: []contentBlock{
			{Type: "tool_use", Name: "extract_match", Input: inputBytes},
		},
	}
}

func validTestPayload() toolExtractionPayload {
	return toolExtractionPayload{
		ExtractedSituationSummary: "Family reports gallbladder stones, hospital denying coverage.",
		Candidates: []CandidateMatch{
			{PackageCode: "SEED-GS-001", ConfidenceP: 90, Reasoning: "clear match"},
		},
		ExclusionMatches:               []ExclusionMatch{},
		PendingSignal:                  SignalNotApplicable,
		MultipleDistinctIssuesDetected: false,
		FamilyDistressSignal:           false,
	}
}

func TestClaudeClient_Extract_HappyPath(t *testing.T) {
	var capturedReq messagesRequest
	var capturedAuthHeader, capturedVersionHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuthHeader = r.Header.Get("x-api-key")
		capturedVersionHeader = r.Header.Get("anthropic-version")
		if err := json.NewDecoder(r.Body).Decode(&capturedReq); err != nil {
			t.Errorf("server: failed to decode request body: %v", err)
		}
		resp := mockToolResponse(t, validTestPayload())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClaudeClient("test-api-key", "claude-haiku-4-5-20251001",
		WithBaseURL(server.URL))

	result, err := client.Extract(context.Background(), "mother needs gallbladder surgery", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	// --- assert the request was built correctly ---
	if capturedAuthHeader != "test-api-key" {
		t.Errorf("expected x-api-key header 'test-api-key', got %q", capturedAuthHeader)
	}
	if capturedVersionHeader == "" {
		t.Error("expected anthropic-version header to be set")
	}
	if capturedReq.Model != "claude-haiku-4-5-20251001" {
		t.Errorf("expected model to be passed through, got %q", capturedReq.Model)
	}
	if capturedReq.ToolChoice["type"] != "tool" || capturedReq.ToolChoice["name"] != "extract_match" {
		t.Errorf("expected tool_choice to force extract_match, got %+v", capturedReq.ToolChoice)
	}
	if len(capturedReq.Tools) != 1 {
		t.Errorf("expected exactly one tool definition, got %d", len(capturedReq.Tools))
	}
	if !strings.Contains(capturedReq.System, "extraction-and-matching") {
		t.Error("expected system prompt to be sent")
	}
	if len(capturedReq.Messages) != 1 || !strings.Contains(capturedReq.Messages[0].Content, "gallbladder surgery") {
		t.Error("expected family description to appear in the user message")
	}

	// --- assert the response was parsed correctly ---
	if result.ExtractedSituationSummary == "" {
		t.Error("expected non-empty extracted situation summary")
	}
	if len(result.Candidates) != 1 || result.Candidates[0].PackageCode != "SEED-GS-001" {
		t.Errorf("unexpected candidates: %+v", result.Candidates)
	}
}

func TestClaudeClient_Extract_NoAPIKey(t *testing.T) {
	client := NewClaudeClient("", "claude-haiku-4-5-20251001")
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("expected error to mention API key, got: %v", err)
	}
}

func TestClaudeClient_Extract_APIReturnsErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"type":"error","error":{"type":"rate_limit_error","message":"slow down"}}`))
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL), WithSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestClaudeClient_Extract_RetriesOnRateLimitThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"type":"error","error":{"type":"rate_limit_error","message":"slow down"}}`))
			return
		}
		json.NewEncoder(w).Encode(mockToolResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL), WithSleepFunc(instantSleep))
	result, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected the retry to succeed, got error: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("expected exactly 2 calls (1 failure + 1 retry that succeeded), got %d", calls.Load())
	}
	if len(result.Candidates) != 1 {
		t.Errorf("expected the successful retry's result to be returned, got %+v", result)
	}
}

func TestClaudeClient_Extract_RetriesOnServerErrorThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"type":"error","error":{"type":"overloaded_error","message":"try again"}}`))
			return
		}
		json.NewEncoder(w).Encode(mockToolResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL), WithSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected the 3rd attempt (default maxAttempts) to succeed, got: %v", err)
	}
	if calls.Load() != 3 {
		t.Errorf("expected exactly 3 calls, got %d", calls.Load())
	}
}

func TestClaudeClient_Extract_GivesUpAfterMaxAttempts(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001",
		WithBaseURL(server.URL), WithSleepFunc(instantSleep), WithMaxAttempts(3))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error after exhausting all retries against a persistently failing server")
	}
	if calls.Load() != 3 {
		t.Errorf("expected exactly maxAttempts=3 calls, got %d", calls.Load())
	}
}

func TestClaudeClient_Extract_DoesNotRetryClientErrors(t *testing.T) {
	// A 400 means this system sent a bad request — retrying an identical
	// request cannot produce a different outcome, so this must fail on
	// the first attempt, not burn through the retry budget (and the
	// wait that comes with it) for nothing.
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"type":"error","error":{"type":"invalid_request_error","message":"bad request"}}`))
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL), WithSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error for a 400 response")
	}
	if calls.Load() != 1 {
		t.Errorf("expected exactly 1 call — a 400 must not be retried, got %d calls", calls.Load())
	}
}

func TestClaudeClient_Extract_DoesNotRetryValidationFailures(t *testing.T) {
	// A 200 response whose payload fails this system's own validation
	// (e.g. an out-of-range confidence) is a content problem, not an
	// infrastructure blip — retrying the same request would get the
	// same payload back.
	var calls atomic.Int32
	invalidPayload := validTestPayload()
	invalidPayload.Candidates[0].ConfidenceP = 150 // out of range
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		json.NewEncoder(w).Encode(mockToolResponse(t, invalidPayload))
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL), WithSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected a validation error")
	}
	if calls.Load() != 1 {
		t.Errorf("expected exactly 1 call — a validation failure must not be retried, got %d calls", calls.Load())
	}
}

func TestClaudeClient_Extract_RetriesTransportErrorThenSucceeds(t *testing.T) {
	// The first "server" is closed before use, so the first attempt
	// hits a real connection-refused transport error, not an HTTP
	// status — this is the "DNS/connection/TLS" retry path, distinct
	// from the "got an HTTP response" retry path above.
	realServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockToolResponse(t, validTestPayload()))
	}))
	defer realServer.Close()
	deadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := deadServer.URL
	deadServer.Close() // closed immediately — connections to it now refuse

	var calls atomic.Int32
	// A base URL that fails on the first call and succeeds after: swap
	// it out via a tiny proxy handler rather than changing the client's
	// baseURL mid-test (the client only reads baseURL once per call, so
	// pointing at a handler that redirects its own behavior based on
	// call count is simpler than trying to mutate the client).
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("expected ResponseWriter to support hijacking for this test")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatalf("hijack failed: %v", err)
			}
			conn.Close() // abrupt close — simulates a transport-level failure
			return
		}
		json.NewEncoder(w).Encode(mockToolResponse(t, validTestPayload()))
	}))
	defer proxy.Close()
	_ = realServer
	_ = deadURL

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(proxy.URL), WithSleepFunc(instantSleep))
	result, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected the retry after a transport-level failure to succeed, got: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("expected exactly 2 calls, got %d", calls.Load())
	}
	if len(result.Candidates) != 1 {
		t.Errorf("expected the successful retry's result, got %+v", result)
	}
}

func TestClaudeClient_Extract_RetryAfterHeaderIsHonoredUnderCap(t *testing.T) {
	var calls atomic.Int32
	var sleptFor time.Duration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(mockToolResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL),
		WithSleepFunc(func(d time.Duration) <-chan time.Time {
			sleptFor = d
			return instantSleep(d)
		}))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if sleptFor != 2*time.Second {
		t.Errorf("expected the 2-second Retry-After header to be honored (under the 3s cap), got sleep of %v", sleptFor)
	}
}

func TestClaudeClient_Extract_RetryAfterBeyondCapFallsBackToFixedDelay(t *testing.T) {
	var calls atomic.Int32
	var sleptFor time.Duration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "120") // far beyond maxRetryAfter
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(mockToolResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL),
		WithSleepFunc(func(d time.Duration) <-chan time.Time {
			sleptFor = d
			return instantSleep(d)
		}))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if sleptFor != 500*time.Millisecond {
		t.Errorf("expected a 120s Retry-After to be ignored in favor of the fixed schedule's first delay (500ms), got %v", sleptFor)
	}
}

func TestClaudeClient_Extract_ContextCancelledDuringBackoffStopsRetrying(t *testing.T) {
	var calls atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		cancel() // cancel the caller's context right after the first failure
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL),
		WithSleepFunc(func(d time.Duration) <-chan time.Time {
			// Never fires — the test proves ctx.Done() wins the select
			// in Extract's retry loop instead of this channel.
			return make(chan time.Time)
		}))
	_, err := client.Extract(ctx, "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error when context is cancelled mid-backoff")
	}
	if calls.Load() != 1 {
		t.Errorf("expected exactly 1 call before the context cancellation stopped retrying, got %d", calls.Load())
	}
}

func TestParseRetryAfterSeconds(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want time.Duration
	}{
		{"valid seconds", "5", 5 * time.Second},
		{"zero", "0", 0},
		{"empty", "", 0},
		{"negative", "-1", 0},
		{"non-numeric", "soon", 0},
		{"HTTP-date form, unsupported", "Wed, 21 Oct 2026 07:28:00 GMT", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseRetryAfterSeconds(tc.in)
			if got != tc.want {
				t.Errorf("parseRetryAfterSeconds(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestClaudeClient_DelayFor_FallsBackPastRetryDelaysLength(t *testing.T) {
	// maxAttempts can be configured higher than len(retryDelays); once
	// attempts run past the named schedule, delayFor must keep returning
	// the last configured delay rather than panicking on an out-of-range
	// index.
	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithMaxAttempts(5))
	got := client.delayFor(10, 0)
	want := client.retryDelays[len(client.retryDelays)-1]
	if got != want {
		t.Errorf("expected delayFor to clamp to the last configured delay (%v) for an attempt number beyond the schedule, got %v", want, got)
	}
}

func TestClaudeClient_Extract_MissingToolUseBlock(t *testing.T) {
	// Simulates the model responding with prose instead of the forced
	// tool call — should never happen with tool_choice forced, but the
	// client must fail loudly rather than silently proceed with a zero
	// value that downstream code would misinterpret as a real answer.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := messagesResponse{
			StopReason: "end_turn",
			Content:    []contentBlock{{Type: "text", Text: "I'm not sure how to help."}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected hard failure when no tool_use block is present")
	}
	if !strings.Contains(err.Error(), "extract_match") {
		t.Errorf("expected error to explain the missing tool call, got: %v", err)
	}
}

func TestClaudeClient_Extract_ZeroCandidatesFailsValidation(t *testing.T) {
	payload := validTestPayload()
	payload.Candidates = []CandidateMatch{} // schema says minItems:1, but be defensive anyway

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := mockToolResponse(t, payload)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected validation error for zero candidates despite schema constraint — defense in depth")
	}
}

func TestClaudeClient_Extract_InvalidPendingSignalFailsValidation(t *testing.T) {
	payload := validTestPayload()
	payload.PendingSignal = PendingSignal("something_the_model_made_up")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := mockToolResponse(t, payload)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected validation error for an out-of-enum pending_signal value")
	}
}

func TestClaudeClient_Extract_OutOfRangeConfidenceFailsValidation(t *testing.T) {
	payload := validTestPayload()
	payload.Candidates = []CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 150, Reasoning: "x"}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := mockToolResponse(t, payload)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected validation error for confidence > 100")
	}
}

func TestClaudeClient_Extract_ExclusionMatchesPassThrough(t *testing.T) {
	payload := validTestPayload()
	payload.ExclusionMatches = []ExclusionMatch{
		{Category: "cosmetic", ConfidenceP: 85, Reasoning: "explicitly cosmetic request"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := mockToolResponse(t, payload)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL))
	result, err := client.Extract(context.Background(), "cosmetic nose job", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ExclusionMatches) != 1 || result.ExclusionMatches[0].Category != "cosmetic" {
		t.Errorf("expected exclusion match to pass through, got %+v", result.ExclusionMatches)
	}
}

func TestClaudeClient_Extract_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never respond — the context should cancel first.
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClaudeClient("test-key", "claude-haiku-4-5-20251001", WithBaseURL(server.URL), WithSleepFunc(instantSleep))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.Extract(ctx, "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected error when context is already cancelled")
	}
}
