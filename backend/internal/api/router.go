package api

import (
	"net/http"
	"time"
)

// Options is the HTTP layer's configuration. It is a struct rather than a
// long argument list because the list had already reached seven
// positional ints and strings, where a transposed pair compiles happily
// and misbehaves at runtime.
type Options struct {
	// AllowedOrigins is the CORS allowlist.
	AllowedOrigins []string

	// IntakeRatePerMinute and IntakeRatePerHour bound the LLM-costing
	// endpoint. See config.RateLimitPerMinute's doc comment.
	IntakeRatePerMinute int
	IntakeRatePerHour   int

	// MaxConcurrentLLM bounds simultaneous in-flight LLM calls.
	MaxConcurrentLLM int

	// GeneralRatePerMinute and GeneralRatePerHour bound every other
	// store-touching endpoint.
	GeneralRatePerMinute int
	GeneralRatePerHour   int

	// TrustProxyHeaders enables X-Forwarded-For / X-Real-IP for client
	// identification. False unless a reverse proxy actually overwrites
	// them; see config.TrustProxyHeaders.
	TrustProxyHeaders bool

	// IntakeTimeout bounds one POST /api/v1/cases request, including the
	// provider call. Zero disables the bound.
	IntakeTimeout time.Duration
}

// NewRouter builds the full HTTP handler chain: routes wrapped in
// middleware, in the order panic-recovery -> logging -> security headers
// -> no-store -> CORS -> rate limiting (-> timeout and concurrency
// limiting on the one LLM-calling route).
//
// Uses Go 1.22's standard library http.ServeMux, which gained
// method-and-pattern matching ("POST /api/v1/cases") specifically
// obviating the need for a third-party router for an API surface this
// size — see ../../../ARCHITECTURE.md for the fuller reasoning (this
// backend has zero external Go dependencies).
//
// Two separate limiters, not one shared instance, because they guard
// against different things: the intake limiter (Options.IntakeRate*) is
// the LLM-cost control; the general limiter (Options.GeneralRate*) is a
// flat, deliberately generous allowance shared by every other endpoint
// that touches the store. The health check is deliberately left outside
// every limiter, so a monitor polling it frequently never trips one meant
// for abuse.
func NewRouter(s *Server, opts Options) http.Handler {
	mux := http.NewServeMux()

	intakeRL := newRateLimiter(opts.IntakeRatePerMinute, opts.IntakeRatePerHour)
	generalRL := newRateLimiter(opts.GeneralRatePerMinute, opts.GeneralRatePerHour)
	llmSem := newLLMSemaphore(opts.MaxConcurrentLLM)

	intakeChain := rateLimitMiddleware(intakeRL, opts.TrustProxyHeaders,
		timeoutMiddleware(opts.IntakeTimeout,
			llmConcurrencyMiddleware(llmSem, http.HandlerFunc(s.handleIntake))))

	general := func(h http.HandlerFunc) http.Handler {
		return rateLimitMiddleware(generalRL, opts.TrustProxyHeaders, h)
	}

	mux.Handle("POST /api/v1/cases", intakeChain)
	mux.Handle("GET /api/v1/cases/{id}", general(s.handleGetCase))
	mux.Handle("DELETE /api/v1/cases/{id}", general(s.handleDeleteCase))
	mux.Handle("GET /api/v1/cases/{id}/document", general(s.handleGetCaseDocument))
	mux.Handle("POST /api/v1/cases/{id}/evidence", general(s.handleAddEvidence))
	mux.Handle("GET /api/v1/packages", general(s.handlePackages))
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	// Method-not-allowed and not-found fall through to ServeMux's default
	// text responses. Wrapping the mux in a JSON fallback would be nicer
	// for a public API, but it is not worth shadowing the standard
	// library's own correct 405 behaviour; the frontend never relies on
	// it.
	var handler http.Handler = mux
	handler = corsMiddleware(opts.AllowedOrigins, handler)
	handler = noStoreMiddleware(handler)
	handler = securityHeadersMiddleware(handler)
	handler = loggingMiddleware(s.Logger, handler)
	handler = recoverMiddleware(s.Logger, handler)

	return handler
}
