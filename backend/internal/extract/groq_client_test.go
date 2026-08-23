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

func mockGroqResponse(t *testing.T, payload toolExtractionPayload) groqChatResponse {
	t.Helper()
	contentBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal test payload: %v", err)
	}
	return groqChatResponse{
		Choices: []groqChoice{{
			Message:      groqMessage{Role: "assistant", Content: string(contentBytes)},
			FinishReason: "stop",
		}},
	}
}

func TestGroqClient_Extract_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected path /chat/completions, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("expected Bearer auth header, got %q", got)
		}
		var req groqChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.ResponseFormat.Type != "json_schema" {
			t.Errorf("expected response_format type json_schema, got %s", req.ResponseFormat.Type)
		}
		if req.ResponseFormat.JSONSchema.Strict {
			t.Error("expected strict:false — see GroqClient's doc comment for why")
		}
		json.NewEncoder(w).Encode(mockGroqResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "openai/gpt-oss-120b", WithGroqBaseURL(server.URL))
	result, err := client.Extract(context.Background(), "test description", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].PackageCode != "SEED-GS-001" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestGroqClient_Extract_NoAPIKey(t *testing.T) {
	client := NewGroqClient("", "openai/gpt-oss-120b")
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil || !strings.Contains(err.Error(), "GROQ_API_KEY") {
		t.Fatalf("expected an error naming GROQ_API_KEY, got: %v", err)
	}
}

func TestGroqClient_Extract_DoesNotRetryClientErrors(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"bad request","type":"invalid_request_error"}}`))
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "openai/gpt-oss-120b", WithGroqBaseURL(server.URL), WithGroqSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error for a 400 response")
	}
	if calls.Load() != 1 {
		t.Errorf("expected exactly 1 call — a 400 must not be retried, got %d", calls.Load())
	}
}

func TestGroqClient_Extract_RetriesOnRateLimitThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"message":"rate limited","type":"rate_limit_error"}}`))
			return
		}
		json.NewEncoder(w).Encode(mockGroqResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "openai/gpt-oss-120b", WithGroqBaseURL(server.URL), WithGroqSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected the retry to succeed, got: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("expected exactly 2 calls, got %d", calls.Load())
	}
}

func TestGroqClient_Extract_GivesUpAfterMaxAttempts(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "openai/gpt-oss-120b", WithGroqBaseURL(server.URL), WithGroqSleepFunc(instantSleep), WithGroqMaxAttempts(3))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error after exhausting retries")
	}
	if calls.Load() != 3 {
		t.Errorf("expected exactly 3 calls, got %d", calls.Load())
	}
}

func TestGroqClient_Extract_NoChoicesIsAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(groqChatResponse{Choices: nil})
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "openai/gpt-oss-120b", WithGroqBaseURL(server.URL), WithGroqSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error for a response with zero choices")
	}
}

func TestGroqClient_Extract_ValidationFailurePassesThrough(t *testing.T) {
	invalid := validTestPayload()
	invalid.Candidates = nil // zero candidates — validatePayload rejects this
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockGroqResponse(t, invalid))
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "openai/gpt-oss-120b", WithGroqBaseURL(server.URL), WithGroqSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected validatePayload's rejection (zero candidates) to surface as an error")
	}
}

func TestGroqClient_Extract_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "openai/gpt-oss-120b", WithGroqBaseURL(server.URL), WithGroqSleepFunc(instantSleep))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Extract(ctx, "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error for an already-cancelled context")
	}
}

func TestGroqClient_Extract_RetryAfterHeaderIsHonoredUnderCap(t *testing.T) {
	var calls atomic.Int32
	var sleptFor time.Duration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(mockGroqResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "openai/gpt-oss-120b", WithGroqBaseURL(server.URL),
		WithGroqSleepFunc(func(d time.Duration) <-chan time.Time {
			sleptFor = d
			return instantSleep(d)
		}))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if sleptFor != 2*time.Second {
		t.Errorf("expected the 2s Retry-After to be honored, got sleep of %v", sleptFor)
	}
}

func TestGroqClient_DelayFor_FallsBackPastRetryDelaysLength(t *testing.T) {
	client := NewGroqClient("test-key", "openai/gpt-oss-120b", WithGroqMaxAttempts(5))
	got := client.delayFor(10, 0)
	want := client.retryDelays[len(client.retryDelays)-1]
	if got != want {
		t.Errorf("expected clamping to the last configured delay (%v), got %v", want, got)
	}
}
