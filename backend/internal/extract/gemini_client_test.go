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

func mockGeminiResponse(t *testing.T, payload toolExtractionPayload) geminiResponse {
	t.Helper()
	contentBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal test payload: %v", err)
	}
	return geminiResponse{
		Candidates: []geminiRespCandidate{{
			Content:      geminiContent{Role: "model", Parts: []geminiPart{{Text: string(contentBytes)}}},
			FinishReason: "STOP",
		}},
	}
}

func TestGeminiClient_Extract_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/models/gemini-2.5-flash-lite:generateContent") {
			t.Errorf("expected the model to be embedded in the URL path, got %s", r.URL.Path)
		}
		if got := r.Header.Get("x-goog-api-key"); got != "test-key" {
			t.Errorf("expected x-goog-api-key header, got %q", got)
		}
		var req geminiRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.GenerationConfig.ResponseMimeType != "application/json" {
			t.Errorf("expected responseMimeType application/json, got %s", req.GenerationConfig.ResponseMimeType)
		}
		if len(req.SystemInstruction.Parts) == 0 || req.SystemInstruction.Parts[0].Text == "" {
			t.Error("expected a non-empty system_instruction")
		}
		json.NewEncoder(w).Encode(mockGeminiResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiBaseURL(server.URL))
	result, err := client.Extract(context.Background(), "test description", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].PackageCode != "SEED-GS-001" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestGeminiClient_Extract_NoAPIKey(t *testing.T) {
	client := NewGeminiClient("", "gemini-2.5-flash-lite")
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil || !strings.Contains(err.Error(), "GEMINI_API_KEY") {
		t.Fatalf("expected an error naming GEMINI_API_KEY, got: %v", err)
	}
}

func TestGeminiClient_Extract_DoesNotRetryClientErrors(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":400,"message":"bad request","status":"INVALID_ARGUMENT"}}`))
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiBaseURL(server.URL), WithGeminiSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error for a 400 response")
	}
	if calls.Load() != 1 {
		t.Errorf("expected exactly 1 call — a 400 must not be retried, got %d", calls.Load())
	}
}

func TestGeminiClient_Extract_RetriesOnRateLimitThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":429,"message":"rate limited","status":"RESOURCE_EXHAUSTED"}}`))
			return
		}
		json.NewEncoder(w).Encode(mockGeminiResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiBaseURL(server.URL), WithGeminiSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected the retry to succeed, got: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("expected exactly 2 calls, got %d", calls.Load())
	}
}

func TestGeminiClient_Extract_GivesUpAfterMaxAttempts(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiBaseURL(server.URL), WithGeminiSleepFunc(instantSleep), WithGeminiMaxAttempts(3))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error after exhausting retries")
	}
	if calls.Load() != 3 {
		t.Errorf("expected exactly 3 calls, got %d", calls.Load())
	}
}

func TestGeminiClient_Extract_NoCandidatesReportsBlockReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(geminiResponse{
			Candidates:     nil,
			PromptFeedback: &geminiPromptFeedback{BlockReason: "SAFETY"},
		})
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiBaseURL(server.URL), WithGeminiSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error when the prompt was blocked")
	}
	if !strings.Contains(err.Error(), "SAFETY") {
		t.Errorf("expected the block reason (SAFETY) in the error, got: %v", err)
	}
}

func TestGeminiClient_Extract_EmptyContentPartsIsAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(geminiResponse{
			Candidates: []geminiRespCandidate{{FinishReason: "MAX_TOKENS"}},
		})
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiBaseURL(server.URL), WithGeminiSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error for a candidate with no content parts")
	}
	if !strings.Contains(err.Error(), "MAX_TOKENS") {
		t.Errorf("expected the finish reason (MAX_TOKENS) in the error for debuggability, got: %v", err)
	}
}

func TestGeminiClient_Extract_ValidationFailurePassesThrough(t *testing.T) {
	invalid := validTestPayload()
	invalid.Candidates[0].ConfidenceP = 150 // out of range
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockGeminiResponse(t, invalid))
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiBaseURL(server.URL), WithGeminiSleepFunc(instantSleep))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected validatePayload's rejection (out-of-range confidence) to surface as an error")
	}
}

func TestGeminiClient_Extract_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiBaseURL(server.URL), WithGeminiSleepFunc(instantSleep))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Extract(ctx, "test", testCandidates(), testExclusionCandidates())
	if err == nil {
		t.Fatal("expected an error for an already-cancelled context")
	}
}

func TestGeminiClient_Extract_RetryAfterHeaderIsHonoredUnderCap(t *testing.T) {
	var calls atomic.Int32
	var sleptFor time.Duration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(mockGeminiResponse(t, validTestPayload()))
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiBaseURL(server.URL),
		WithGeminiSleepFunc(func(d time.Duration) <-chan time.Time {
			sleptFor = d
			return instantSleep(d)
		}))
	_, err := client.Extract(context.Background(), "test", testCandidates(), testExclusionCandidates())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if sleptFor != 1*time.Second {
		t.Errorf("expected the 1s Retry-After to be honored, got sleep of %v", sleptFor)
	}
}

func TestGeminiClient_DelayFor_FallsBackPastRetryDelaysLength(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.5-flash-lite", WithGeminiMaxAttempts(5))
	got := client.delayFor(10, 0)
	want := client.retryDelays[len(client.retryDelays)-1]
	if got != want {
		t.Errorf("expected clamping to the last configured delay (%v), got %v", want, got)
	}
}

// --- convertToGeminiSchema: tested directly, not just incidentally
// through the HappyPath test above, since a schema bug that still
// produces syntactically valid JSON could pass a mocked HTTP round-trip
// while genuinely breaking against the real, stricter Gemini API.

func TestConvertToGeminiSchema_UppercasesTypeAtEveryLevel(t *testing.T) {
	input := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"tags": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "integer"},
			},
			"active": map[string]any{"type": "boolean"},
		},
	}

	got := convertToGeminiSchema(input)

	if got["type"] != "OBJECT" {
		t.Errorf("expected top-level type OBJECT, got %v", got["type"])
	}
	props := got["properties"].(map[string]any)
	if props["name"].(map[string]any)["type"] != "STRING" {
		t.Error("expected nested property type STRING")
	}
	tags := props["tags"].(map[string]any)
	if tags["type"] != "ARRAY" {
		t.Error("expected array property type ARRAY")
	}
	if tags["items"].(map[string]any)["type"] != "INTEGER" {
		t.Error("expected array items type INTEGER")
	}
	if props["active"].(map[string]any)["type"] != "BOOLEAN" {
		t.Error("expected boolean property type BOOLEAN")
	}
}

func TestConvertToGeminiSchema_PreservesNonTypeFields(t *testing.T) {
	input := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pending_signal": map[string]any{
				"type": "string",
				"enum": []string{"a", "b", "c"},
			},
			"score": map[string]any{
				"type":    "integer",
				"minimum": 0,
				"maximum": 100,
			},
		},
		"required": []string{"pending_signal", "score"},
	}

	got := convertToGeminiSchema(input)

	if req, ok := got["required"].([]string); !ok || len(req) != 2 {
		t.Errorf("expected required to pass through unchanged, got %v", got["required"])
	}
	props := got["properties"].(map[string]any)
	enumField := props["pending_signal"].(map[string]any)
	enumVals, ok := enumField["enum"].([]string)
	if !ok || len(enumVals) != 3 || enumVals[0] != "a" {
		t.Errorf("expected enum values to pass through unchanged (lowercase, since these are data not types), got %v", enumField["enum"])
	}
	scoreField := props["score"].(map[string]any)
	if scoreField["minimum"] != 0 || scoreField["maximum"] != 100 {
		t.Errorf("expected minimum/maximum to pass through unchanged, got min=%v max=%v", scoreField["minimum"], scoreField["maximum"])
	}
}

func TestConvertToGeminiSchema_RealExtractSchemaConvertsWithoutPanicking(t *testing.T) {
	// The actual schema this system sends, not a simplified stand-in —
	// this is the test that would have caught a real, structurally
	// unhandled shape in the production schema (e.g. a third level of
	// array-of-array nesting) that the simpler tests above wouldn't.
	got := convertToGeminiSchema(extractMatchToolSchema["input_schema"].(map[string]any))
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("expected the converted schema to be JSON-marshalable, got: %v", err)
	}
	s := string(encoded)
	if strings.Contains(s, `"type":"object"`) || strings.Contains(s, `"type":"string"`) {
		t.Error("found a lowercase type value that survived conversion — the real schema has a shape convertToGeminiSchema's simpler tests didn't cover")
	}
	if !strings.Contains(s, `"type":"OBJECT"`) || !strings.Contains(s, `"type":"STRING"`) {
		t.Error("expected at least one OBJECT and one STRING type in the converted real schema")
	}
}
