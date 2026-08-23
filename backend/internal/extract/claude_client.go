package extract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// ClaudeClient is the real, network-backed Extractor. It talks to the
// Anthropic Messages API directly over net/http — deliberately no SDK
// dependency. Two reasons, both worth stating rather than leaving
// implicit: first, this project's build environment could not reliably
// fetch third-party Go module dependencies (see ../../../ARCHITECTURE.md), so
// the whole backend is standard-library-only; second, even without that
// constraint, a single well-understood HTTP call is easier for a solo
// engineer to audit and modify than an SDK dependency pulled in for one
// endpoint.
type ClaudeClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string

	// maxAttempts is the total number of tries (1 initial + retries) for
	// a request that keeps failing with a retryable error. See Extract's
	// retry doc comment for exactly what counts as retryable and why
	// this is deliberately small.
	maxAttempts int
	// retryDelays holds the backoff before attempt 2, 3, ... — index 0
	// is the delay before the second attempt. Fixed, named delays
	// rather than a computed exponential curve, so the actual wait a
	// family experiences is a number a reviewer can read directly here,
	// not derive from a formula.
	retryDelays []time.Duration
	// maxRetryAfter caps how long a Retry-After response header is
	// allowed to make this wait — seen in shared_test.go for why an
	// unbounded Retry-After is not honored as-is.
	maxRetryAfter time.Duration
	// sleep is how a retry wait is performed; overridden in tests so
	// retry tests run at real production retryDelays values without the
	// test suite actually pausing for them.
	sleep func(time.Duration) <-chan time.Time
}

// ClaudeClientOption configures a ClaudeClient at construction.
type ClaudeClientOption func(*ClaudeClient)

// WithHTTPClient overrides the default HTTP client — used by tests to
// point at an httptest server instead of the real API.
func WithHTTPClient(c *http.Client) ClaudeClientOption {
	return func(cc *ClaudeClient) { cc.httpClient = c }
}

// WithBaseURL overrides the API base URL — used by tests.
func WithBaseURL(url string) ClaudeClientOption {
	return func(cc *ClaudeClient) { cc.baseURL = url }
}

// WithMaxAttempts overrides the default retry attempt count — mainly for
// tests that want to assert exactly how many times a persistently
// failing server got called.
func WithMaxAttempts(n int) ClaudeClientOption {
	return func(cc *ClaudeClient) { cc.maxAttempts = n }
}

// WithSleepFunc overrides how retry backoff is performed — tests use
// this to make retries resolve instantly regardless of the configured
// delay, so production delay values don't have to be shortened (and
// therefore made unrealistic) just to keep the test suite fast.
func WithSleepFunc(sleep func(time.Duration) <-chan time.Time) ClaudeClientOption {
	return func(cc *ClaudeClient) { cc.sleep = sleep }
}

// NewClaudeClient builds a real extractor. model should be a current,
// cheap Claude model — this system's per-query cost is dominated by this
// one call (see ../../../ARCHITECTURE.md, "reducing bills"), and the task
// itself (bounded classification over a short candidate list, not
// open-ended generation) does not need a large model to do well.
func NewClaudeClient(apiKey, model string, opts ...ClaudeClientOption) *ClaudeClient {
	cc := &ClaudeClient{
		apiKey:        apiKey,
		model:         model,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		baseURL:       "https://api.anthropic.com",
		maxAttempts:   3,
		retryDelays:   []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond},
		maxRetryAfter: 3 * time.Second,
		sleep:         func(d time.Duration) <-chan time.Time { return time.After(d) },
	}
	for _, opt := range opts {
		opt(cc)
	}
	return cc
}

type messagesRequest struct {
	Model      string         `json:"model"`
	MaxTokens  int            `json:"max_tokens"`
	System     string         `json:"system"`
	Messages   []reqMessage   `json:"messages"`
	Tools      []any          `json:"tools"`
	ToolChoice map[string]any `json:"tool_choice"`
}

type reqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	ID         string          `json:"id"`
	StopReason string          `json:"stop_reason"`
	Content    []contentBlock  `json:"content"`
	Error      *anthropicError `json:"error,omitempty"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// toolExtractionPayload mirrors the tool's input_schema for unmarshalling
// the model's actual structured output.
type toolExtractionPayload struct {
	ExtractedSituationSummary      string           `json:"extracted_situation_summary"`
	Candidates                     []CandidateMatch `json:"candidates"`
	ExclusionMatches               []ExclusionMatch `json:"exclusion_matches"`
	PendingSignal                  PendingSignal    `json:"pending_signal"`
	MultipleDistinctIssuesDetected bool             `json:"multiple_distinct_issues_detected"`
	FamilyDistressSignal           bool             `json:"family_distress_signal"`
}

// Extract implements Extractor against the real API.
//
// Retries: a bounded number of attempts (default 3) are made when a
// failure looks transient — a network-level error, HTTP 429, or any 5xx
// — with a short fixed backoff between them (or the server's own
// Retry-After, capped, when a 429 supplies one). Every other failure
// (missing API key, a malformed request on this system's own side, a
// non-retryable 4xx, a response that parses but fails validation) is
// returned immediately: retrying an identical request against a server
// that is correctly rejecting it cannot produce a different outcome, and
// making a family wait through backoff for a failure that was never
// going to resolve is a worse outcome than failing fast to the
// care-first fallback handleIntake already provides. See
// ../../../ARCHITECTURE.md for why this system does not treat "the LLM call
// failed" as rare enough to leave unhandled.
func (c *ClaudeClient) Extract(ctx context.Context, familyDescription string, candidates []CandidatePackageInfo, exclusionCandidates []CandidateExclusionInfo) (Result, error) {
	if c.apiKey == "" {
		return Result{}, fmt.Errorf("extract: no API key configured — set ANTHROPIC_API_KEY (see .env.example)")
	}

	userContent, err := buildUserContent(familyDescription, candidates, exclusionCandidates)
	if err != nil {
		return Result{}, err
	}

	reqBody := messagesRequest{
		Model:     c.model,
		MaxTokens: 1500,
		System:    systemPrompt,
		Messages: []reqMessage{
			{Role: "user", Content: userContent},
		},
		Tools: []any{extractMatchToolSchema},
		ToolChoice: map[string]any{
			"type": "tool",
			"name": "extract_match",
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return Result{}, fmt.Errorf("extract: marshalling request: %w", err)
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

// attemptResult carries one HTTP attempt's outcome plus enough
// information for Extract's retry loop to decide what to do next.
type attemptResult struct {
	result     Result
	err        error
	retryable  bool
	retryAfter time.Duration // 0 unless a 429 supplied Retry-After
}

// attemptExtract performs exactly one HTTP call and parses/validates the
// response. It never sleeps or retries itself — Extract owns that
// decision, since only Extract knows how many attempts have already
// been made.
func (c *ClaudeClient) attemptExtract(ctx context.Context, bodyBytes []byte) attemptResult {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return attemptResult{err: fmt.Errorf("extract: building request: %w", err)}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			// The caller's own context is what ended this, not a
			// transient server/network condition — retrying would just
			// wait again for the same already-decided outcome.
			return attemptResult{err: fmt.Errorf("extract: calling Anthropic API: %w", err)}
		}
		// DNS failure, connection refused/reset, TLS handshake failure —
		// the class of error a second attempt a moment later often
		// clears on its own.
		return attemptResult{err: fmt.Errorf("extract: calling Anthropic API: %w", err), retryable: true}
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return attemptResult{err: fmt.Errorf("extract: reading response body: %w", err)}
	}

	if httpResp.StatusCode != http.StatusOK {
		retryable := httpResp.StatusCode == http.StatusTooManyRequests || httpResp.StatusCode >= 500
		var retryAfter time.Duration
		if httpResp.StatusCode == http.StatusTooManyRequests {
			retryAfter = parseRetryAfterSeconds(httpResp.Header.Get("Retry-After"))
		}
		return attemptResult{
			err:        fmt.Errorf("extract: API returned status %d: %s", httpResp.StatusCode, truncate(string(respBytes), 500)),
			retryable:  retryable,
			retryAfter: retryAfter,
		}
	}

	var resp messagesResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return attemptResult{err: fmt.Errorf("extract: parsing response JSON: %w", err)}
	}
	if resp.Error != nil {
		return attemptResult{err: fmt.Errorf("extract: API error (%s): %s", resp.Error.Type, resp.Error.Message)}
	}

	var toolBlock *contentBlock
	for i := range resp.Content {
		if resp.Content[i].Type == "tool_use" && resp.Content[i].Name == "extract_match" {
			toolBlock = &resp.Content[i]
			break
		}
	}
	if toolBlock == nil {
		return attemptResult{err: fmt.Errorf("extract: model response contained no extract_match tool call (stop_reason=%s) — this should not happen with tool_choice forced; treat as a hard failure, not a guess", resp.StopReason)}
	}

	var payload toolExtractionPayload
	if err := json.Unmarshal(toolBlock.Input, &payload); err != nil {
		return attemptResult{err: fmt.Errorf("extract: parsing tool_use input: %w", err)}
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

// delayFor returns how long to wait before the attempt after attemptNum.
// A server-specified Retry-After is honored only up to maxRetryAfter —
// past that cap this falls back to the fixed schedule instead of
// honoring an arbitrarily long server-dictated wait, on the reasoning
// that a family waiting at a billing counter is better served by a
// bounded wait followed by the care-first fallback than by an
// open-ended one, even if it costs one wasted retry attempt.
func (c *ClaudeClient) delayFor(attemptNum int, retryAfter time.Duration) time.Duration {
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

// parseRetryAfterSeconds parses Retry-After's integer-seconds form (the
// form Anthropic's API uses for 429 responses). The HTTP-date form
// defined by the spec is deliberately not handled — not seen in
// practice from this API, and not worth the added parsing surface for a
// header this client treats as advisory (see delayFor's cap) rather
// than authoritative. Anything unparseable or negative is treated the
// same as absent: fall back to the fixed backoff schedule.
func parseRetryAfterSeconds(v string) time.Duration {
	if v == "" {
		return 0
	}
	secs, err := strconv.Atoi(v)
	if err != nil || secs < 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

// validatePayload catches a model output that is structurally valid JSON
// but semantically useless (e.g. zero candidates despite minItems in the
// schema — models can still occasionally violate a schema in edge cases).
// Failing loudly here beats silently proceeding with a Result the rest of
// the system wasn't built to handle.
func validatePayload(p toolExtractionPayload) error {
	if len(p.Candidates) == 0 {
		return fmt.Errorf("zero candidates returned")
	}
	for _, c := range p.Candidates {
		if c.PackageCode == "" {
			return fmt.Errorf("candidate with empty package_code")
		}
		if c.ConfidenceP < 0 || c.ConfidenceP > 100 {
			return fmt.Errorf("candidate %q has out-of-range confidence %d", c.PackageCode, c.ConfidenceP)
		}
	}
	switch p.PendingSignal {
	case SignalPendingLikely, SignalDeniedLikely, SignalUnclear, SignalNotApplicable:
		// ok
	default:
		return fmt.Errorf("invalid pending_signal %q", p.PendingSignal)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
