package extract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GroqClient implements Extractor against Groq's OpenAI-compatible chat
// completions API (https://api.groq.com/openai/v1/chat/completions).
// Groq hosts open-weight models (Llama, GPT-OSS, Qwen) on their own LPU
// inference hardware rather than serving a proprietary model — see
// ../../../ARCHITECTURE.md for why this system's extraction task (bounded
// classification over a short candidate list, not open-ended generation)
// doesn't need a frontier-tier model to do this well, the same reasoning
// that justifies ClaudeClient's cheap-model default.
//
// Structured output here uses response_format: json_schema with
// strict:false, not strict:true, deliberately — Groq's own structured
// outputs documentation demonstrates strict:false paired with manual
// validation as its recommended, robust pattern (schema-guided
// generation without a hard compile-time compliance guarantee that
// varies by model). validatePayload (shared with ClaudeClient and
// GeminiClient, see claude_client.go) is that manual validation layer,
// not an afterthought bolted onto a supposedly-guaranteed-valid response.
type GroqClient struct {
	apiKey        string
	model         string
	httpClient    *http.Client
	baseURL       string
	maxAttempts   int
	retryDelays   []time.Duration
	maxRetryAfter time.Duration
	sleep         func(time.Duration) <-chan time.Time
}

// GroqClientOption configures a GroqClient at construction.
type GroqClientOption func(*GroqClient)

func WithGroqHTTPClient(c *http.Client) GroqClientOption {
	return func(gc *GroqClient) { gc.httpClient = c }
}

func WithGroqBaseURL(url string) GroqClientOption {
	return func(gc *GroqClient) { gc.baseURL = url }
}

func WithGroqMaxAttempts(n int) GroqClientOption {
	return func(gc *GroqClient) { gc.maxAttempts = n }
}

func WithGroqSleepFunc(sleep func(time.Duration) <-chan time.Time) GroqClientOption {
	return func(gc *GroqClient) { gc.sleep = sleep }
}

// NewGroqClient builds a real Groq-backed extractor. model should be a
// current Groq catalog entry — see .env.example for the current default
// and why it was chosen; Groq's model lineup and pricing change often
// enough that this is worth re-checking against console.groq.com/docs
// before a real deploy, not treated as permanently correct.
func NewGroqClient(apiKey, model string, opts ...GroqClientOption) *GroqClient {
	gc := &GroqClient{
		apiKey:        apiKey,
		model:         model,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		baseURL:       "https://api.groq.com/openai/v1",
		maxAttempts:   3,
		retryDelays:   []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond},
		maxRetryAfter: 3 * time.Second,
		sleep:         func(d time.Duration) <-chan time.Time { return time.After(d) },
	}
	for _, opt := range opts {
		opt(gc)
	}
	return gc
}

// groqChatRequest mirrors the OpenAI-compatible chat completions request
// shape Groq documents at console.groq.com/docs/api-reference.
type groqChatRequest struct {
	Model               string         `json:"model"`
	Messages            []groqMessage  `json:"messages"`
	ResponseFormat      groqRespFormat `json:"response_format"`
	MaxCompletionTokens int            `json:"max_completion_tokens"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqRespFormat struct {
	Type       string             `json:"type"`
	JSONSchema groqJSONSchemaSpec `json:"json_schema"`
}

type groqJSONSchemaSpec struct {
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type groqChatResponse struct {
	Choices []groqChoice   `json:"choices"`
	Error   *groqErrorBody `json:"error"`
}

type groqChoice struct {
	Message      groqMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type groqErrorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// groqExtractSchema reuses extractMatchToolSchema's input_schema
// directly — not a copy, the literal same map — because Groq's
// json_schema.schema field and Anthropic's tool input_schema are both
// standard JSON Schema, so there's no format conversion needed here (see
// geminiResponseSchema in gemini_client.go for the provider that
// genuinely does need one, and why). Sharing the actual object, not a
// lookalike copy, is deliberate: it makes it structurally impossible for
// Claude's and Groq's extraction contracts to silently drift apart —
// every required field, every description, stays identical across both
// by construction, not by remembering to keep two literals in sync.
var groqExtractSchema = extractMatchToolSchema["input_schema"].(map[string]any)

// Extract implements Extractor against Groq. See ClaudeClient.Extract's
// doc comment for the retry policy this mirrors (transient failures
// only, bounded attempts, Retry-After honored under a cap) — the policy
// is the same reasoning applied to a different wire format, not a
// different policy.
func (c *GroqClient) Extract(ctx context.Context, familyDescription string, candidates []CandidatePackageInfo, exclusionCandidates []CandidateExclusionInfo) (Result, error) {
	if c.apiKey == "" {
		return Result{}, fmt.Errorf("extract: no Groq API key configured — set GROQ_API_KEY (see .env.example)")
	}

	userContent, err := buildUserContent(familyDescription, candidates, exclusionCandidates)
	if err != nil {
		return Result{}, err
	}

	reqBody := groqChatRequest{
		Model: c.model,
		Messages: []groqMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		ResponseFormat: groqRespFormat{
			Type: "json_schema",
			JSONSchema: groqJSONSchemaSpec{
				Name:   "extract_match",
				Strict: false,
				Schema: groqExtractSchema,
			},
		},
		MaxCompletionTokens: 1500,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return Result{}, fmt.Errorf("extract: marshalling Groq request: %w", err)
	}

	maxAttempts := c.maxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var last attemptResult
	for attemptNum := 1; attemptNum <= maxAttempts; attemptNum++ {
		last = c.attemptExtract(ctx, bodyBytes)
		if last.err == nil {
			return last.result, nil
		}
		if !last.retryable || attemptNum == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return Result{}, fmt.Errorf("extract: context done while waiting to retry after attempt %d: %w", attemptNum, ctx.Err())
		case <-c.sleep(c.delayFor(attemptNum, last.retryAfter)):
		}
	}
	return Result{}, last.err
}

func (c *GroqClient) attemptExtract(ctx context.Context, bodyBytes []byte) attemptResult {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return attemptResult{err: fmt.Errorf("extract: building Groq request: %w", err)}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return attemptResult{err: fmt.Errorf("extract: calling Groq API: %w", err)}
		}
		return attemptResult{err: fmt.Errorf("extract: calling Groq API: %w", err), retryable: true}
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return attemptResult{err: fmt.Errorf("extract: reading Groq response body: %w", err)}
	}

	if httpResp.StatusCode != http.StatusOK {
		retryable := httpResp.StatusCode == http.StatusTooManyRequests || httpResp.StatusCode >= 500
		var retryAfter time.Duration
		if httpResp.StatusCode == http.StatusTooManyRequests {
			retryAfter = parseRetryAfterSeconds(httpResp.Header.Get("Retry-After"))
		}
		return attemptResult{
			err:        fmt.Errorf("extract: Groq API returned status %d: %s", httpResp.StatusCode, truncate(string(respBytes), 500)),
			retryable:  retryable,
			retryAfter: retryAfter,
		}
	}

	var resp groqChatResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return attemptResult{err: fmt.Errorf("extract: parsing Groq response JSON: %w", err)}
	}
	if resp.Error != nil {
		return attemptResult{err: fmt.Errorf("extract: Groq API error (%s): %s", resp.Error.Type, resp.Error.Message)}
	}
	if len(resp.Choices) == 0 {
		return attemptResult{err: fmt.Errorf("extract: Groq response contained no choices")}
	}

	var payload toolExtractionPayload
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &payload); err != nil {
		return attemptResult{err: fmt.Errorf("extract: parsing Groq structured output as JSON: %w", err)}
	}

	if err := validatePayload(payload); err != nil {
		return attemptResult{err: fmt.Errorf("extract: model output failed validation: %w", err)}
	}

	return attemptResult{result: Result{
		ExtractedSituationSummary:      payload.ExtractedSituationSummary,
		Candidates:                     payload.Candidates,
		ExclusionMatches:               payload.ExclusionMatches,
		Pending:                        payload.PendingSignal,
		MultipleDistinctIssuesDetected: payload.MultipleDistinctIssuesDetected,
		FamilyDistressSignal:           payload.FamilyDistressSignal,
	}}
}

func (c *GroqClient) delayFor(attemptNum int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 && retryAfter <= c.maxRetryAfter {
		return retryAfter
	}
	if len(c.retryDelays) == 0 {
		return 0
	}
	idx := attemptNum - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(c.retryDelays) {
		idx = len(c.retryDelays) - 1
	}
	return c.retryDelays[idx]
}

// buildUserContent is the candidate/exclusion serialization every
// provider client needs — the same regardless of which provider's wire
// format wraps it, so it's written once here rather than in each
// provider's file.
func buildUserContent(familyDescription string, candidates []CandidatePackageInfo, exclusionCandidates []CandidateExclusionInfo) (string, error) {
	candidateJSON, err := json.MarshalIndent(candidates, "", "  ")
	if err != nil {
		return "", fmt.Errorf("extract: marshalling candidates: %w", err)
	}
	exclusionJSON, err := json.MarshalIndent(exclusionCandidates, "", "  ")
	if err != nil {
		return "", fmt.Errorf("extract: marshalling exclusion candidates: %w", err)
	}
	return fmt.Sprintf(
		"Family's description of their situation:\n%q\n\nCandidate HBP packages to evaluate (from keyword pre-filter, may be incomplete or contain irrelevant entries):\n%s\n\nAll confirmed exclusion categories to evaluate (complete list, not pre-filtered):\n%s",
		familyDescription, string(candidateJSON), string(exclusionJSON),
	), nil
}
