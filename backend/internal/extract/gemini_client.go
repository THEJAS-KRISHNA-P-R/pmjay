package extract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GeminiClient implements Extractor against Google's Generative Language
// API (generateContent), the same REST surface Google AI Studio issues
// free-tier keys for. See ../../../ARCHITECTURE.md for why this system's
// extraction task doesn't need a frontier-tier model — model should be a
// current cost-tier entry (see .env.example for the default and why).
//
// Structured output here uses generationConfig.responseSchema with
// responseMimeType "application/json" — Gemini's own structured-output
// mechanism, unrelated to function/tool calling. It has one real
// consequence for this client that Claude's and Groq's don't: Gemini's
// schema format is a subset of OpenAPI 3.0's Schema Object, which spells
// types as uppercase strings ("OBJECT", "STRING") rather than standard
// JSON Schema's lowercase ("object", "string") — see
// convertToGeminiSchema below. This is a genuine wire-format
// incompatibility, not a stylistic choice, which is why this client
// needs a real conversion step where GroqClient could just reuse
// Anthropic's schema object directly.
type GeminiClient struct {
	apiKey        string
	model         string
	httpClient    *http.Client
	baseURL       string
	maxAttempts   int
	retryDelays   []time.Duration
	maxRetryAfter time.Duration
	sleep         func(time.Duration) <-chan time.Time
}

// GeminiClientOption configures a GeminiClient at construction.
type GeminiClientOption func(*GeminiClient)

func WithGeminiHTTPClient(c *http.Client) GeminiClientOption {
	return func(gc *GeminiClient) { gc.httpClient = c }
}

func WithGeminiBaseURL(url string) GeminiClientOption {
	return func(gc *GeminiClient) { gc.baseURL = url }
}

func WithGeminiMaxAttempts(n int) GeminiClientOption {
	return func(gc *GeminiClient) { gc.maxAttempts = n }
}

func WithGeminiSleepFunc(sleep func(time.Duration) <-chan time.Time) GeminiClientOption {
	return func(gc *GeminiClient) { gc.sleep = sleep }
}

// NewGeminiClient builds a real Gemini-backed extractor. model should be
// a current AI Studio catalog entry — Google's Gemini lineup moves fast
// enough (this project's own research found three different generations
// referenced as "current" within the same month) that the default in
// .env.example is worth re-checking against ai.google.dev before a real
// deploy, not treated as permanently correct.
func NewGeminiClient(apiKey, model string, opts ...GeminiClientOption) *GeminiClient {
	gc := &GeminiClient{
		apiKey:        apiKey,
		model:         model,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		baseURL:       "https://generativelanguage.googleapis.com/v1beta",
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

type geminiRequest struct {
	SystemInstruction geminiContent   `json:"system_instruction"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  geminiGenConfig `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	ResponseMimeType string         `json:"responseMimeType"`
	ResponseSchema   map[string]any `json:"responseSchema"`
	MaxOutputTokens  int            `json:"maxOutputTokens"`
}

type geminiResponse struct {
	Candidates     []geminiRespCandidate `json:"candidates"`
	PromptFeedback *geminiPromptFeedback `json:"promptFeedback"`
	Error          *geminiErrorBody      `json:"error"`
}

type geminiRespCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiPromptFeedback struct {
	BlockReason string `json:"blockReason"`
}

type geminiErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// geminiExtractSchema is convertToGeminiSchema applied once, at package
// init, to the same canonical schema Claude and Groq use — computed once
// here rather than on every request, since the input never changes at
// runtime.
var geminiExtractSchema = convertToGeminiSchema(extractMatchToolSchema["input_schema"].(map[string]any))

// convertToGeminiSchema recursively rewrites a standard JSON Schema map
// (lowercase "type" values) into Gemini's OpenAPI-3.0-subset schema
// format (uppercase "type" values), leaving every other key — enum
// values, required, description, minimum, maximum, minItems, all of
// which Gemini's schema format also accepts, per ai.google.dev's schema
// reference — untouched. Recurses through "properties" (a map of nested
// schemas) and "items" (a single nested schema for arrays), the only two
// places a schema can nest another schema in this project's usage.
func convertToGeminiSchema(schema map[string]any) map[string]any {
	out := make(map[string]any, len(schema))
	for k, v := range schema {
		switch k {
		case "type":
			if s, ok := v.(string); ok {
				out[k] = strings.ToUpper(s)
			} else {
				out[k] = v
			}
		case "properties":
			if props, ok := v.(map[string]any); ok {
				newProps := make(map[string]any, len(props))
				for pk, pv := range props {
					if pSchema, ok := pv.(map[string]any); ok {
						newProps[pk] = convertToGeminiSchema(pSchema)
					} else {
						newProps[pk] = pv
					}
				}
				out[k] = newProps
			} else {
				out[k] = v
			}
		case "items":
			if itemSchema, ok := v.(map[string]any); ok {
				out[k] = convertToGeminiSchema(itemSchema)
			} else {
				out[k] = v
			}
		default:
			out[k] = v
		}
	}
	return out
}

// Extract implements Extractor against Gemini. See ClaudeClient.Extract's
// doc comment for the retry policy this mirrors.
func (c *GeminiClient) Extract(ctx context.Context, familyDescription string, candidates []CandidatePackageInfo, exclusionCandidates []CandidateExclusionInfo) (Result, error) {
	if c.apiKey == "" {
		return Result{}, fmt.Errorf("extract: no Gemini API key configured — set GEMINI_API_KEY (see .env.example)")
	}

	userContent, err := buildUserContent(familyDescription, candidates, exclusionCandidates)
	if err != nil {
		return Result{}, err
	}

	reqBody := geminiRequest{
		SystemInstruction: geminiContent{Parts: []geminiPart{{Text: systemPrompt}}},
		Contents:          []geminiContent{{Role: "user", Parts: []geminiPart{{Text: userContent}}}},
		GenerationConfig: geminiGenConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   geminiExtractSchema,
			MaxOutputTokens:  1500,
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return Result{}, fmt.Errorf("extract: marshalling Gemini request: %w", err)
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

func (c *GeminiClient) attemptExtract(ctx context.Context, bodyBytes []byte) attemptResult {
	url := c.baseURL + "/models/" + c.model + ":generateContent"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return attemptResult{err: fmt.Errorf("extract: building Gemini request: %w", err)}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", c.apiKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return attemptResult{err: fmt.Errorf("extract: calling Gemini API: %w", err)}
		}
		return attemptResult{err: fmt.Errorf("extract: calling Gemini API: %w", err), retryable: true}
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return attemptResult{err: fmt.Errorf("extract: reading Gemini response body: %w", err)}
	}

	if httpResp.StatusCode != http.StatusOK {
		retryable := httpResp.StatusCode == http.StatusTooManyRequests || httpResp.StatusCode >= 500
		var retryAfter time.Duration
		if httpResp.StatusCode == http.StatusTooManyRequests {
			retryAfter = parseRetryAfterSeconds(httpResp.Header.Get("Retry-After"))
		}
		return attemptResult{
			err:        fmt.Errorf("extract: Gemini API returned status %d: %s", httpResp.StatusCode, truncate(string(respBytes), 500)),
			retryable:  retryable,
			retryAfter: retryAfter,
		}
	}

	var resp geminiResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return attemptResult{err: fmt.Errorf("extract: parsing Gemini response JSON: %w", err)}
	}
	if resp.Error != nil {
		return attemptResult{err: fmt.Errorf("extract: Gemini API error (%s): %s", resp.Error.Status, resp.Error.Message)}
	}
	if len(resp.Candidates) == 0 {
		if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != "" {
			return attemptResult{err: fmt.Errorf("extract: Gemini blocked the request (reason: %s) — not retryable, the input itself triggered this", resp.PromptFeedback.BlockReason)}
		}
		return attemptResult{err: fmt.Errorf("extract: Gemini response contained no candidates")}
	}
	parts := resp.Candidates[0].Content.Parts
	if len(parts) == 0 {
		return attemptResult{err: fmt.Errorf("extract: Gemini response candidate contained no content parts (finish_reason=%s)", resp.Candidates[0].FinishReason)}
	}

	var payload toolExtractionPayload
	if err := json.Unmarshal([]byte(parts[0].Text), &payload); err != nil {
		return attemptResult{err: fmt.Errorf("extract: parsing Gemini structured output as JSON: %w", err)}
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

func (c *GeminiClient) delayFor(attemptNum int, retryAfter time.Duration) time.Duration {
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
