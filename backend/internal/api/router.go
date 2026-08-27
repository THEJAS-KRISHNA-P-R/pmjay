package api

import "net/http"

// NewRouter builds the full HTTP handler chain: routes wrapped in
// middleware, in the order panic-recovery -> logging -> CORS -> rate
// limiting (-> concurrency limiting on the one LLM-calling route).
// Uses Go 1.22's standard library http.ServeMux, which gained
// method-and-pattern matching ("POST /api/v1/cases") specifically
// obviating the need for a third-party router for an API surface this
// size — see ../../../ARCHITECTURE.md for the fuller reasoning (this
// backend has zero external Go dependencies).
//
// Two separate limiters, not one shared instance, because they guard
// against different things: rl (rateLimitPerMinute/Hour — see
// config.RateLimitPerMinute/Hour) is intake's LLM-cost control;
// generalRL (generalRateLimitPerMinute/Hour — see
// config.GeneralRateLimitPerMinute/Hour) is a flat, deliberately
// generous allowance shared by every other endpoint that touches the
// store, since none of them carry a per-call LLM cost but none are free
// to hammer either — an evidence submission or a document fetch both
// cost real work (a full store rewrite, or PDF generation) that scales
// with usage. The health check is deliberately left outside every
// limiter, so a monitor polling it frequently never trips one meant for
// abuse.
func NewRouter(s *Server, allowedOrigins []string, rateLimitPerMinute, rateLimitPerHour, maxConcurrentLLM, generalRateLimitPerMinute, generalRateLimitPerHour int) http.Handler {
	mux := http.NewServeMux()

	rl := newRateLimiter(rateLimitPerMinute, rateLimitPerHour)
	generalRL := newRateLimiter(generalRateLimitPerMinute, generalRateLimitPerHour)
	llmSem := newLLMSemaphore(maxConcurrentLLM)

	intakeChain := rateLimitMiddleware(rl, llmConcurrencyMiddleware(llmSem, http.HandlerFunc(s.handleIntake)))

	mux.Handle("POST /api/v1/cases", intakeChain)
	mux.Handle("GET /api/v1/cases/{id}", rateLimitMiddleware(generalRL, http.HandlerFunc(s.handleGetCase)))
	mux.Handle("GET /api/v1/cases/{id}/document", rateLimitMiddleware(generalRL, http.HandlerFunc(s.handleGetCaseDocument)))
	mux.Handle("POST /api/v1/cases/{id}/evidence", rateLimitMiddleware(generalRL, http.HandlerFunc(s.handleAddEvidence)))
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	var handler http.Handler = mux
	handler = corsMiddleware(allowedOrigins, handler)
	handler = loggingMiddleware(s.Logger, handler)
	handler = recoverMiddleware(s.Logger, handler)

	return handler
}
