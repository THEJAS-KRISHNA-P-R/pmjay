package api

import "net/http"

func NewRouter(s *Server, allowedOrigins []string, rateLimitPerMinute, rateLimitPerHour, maxConcurrentLLM int) http.Handler {
	mux := http.NewServeMux()

	rl := newRateLimiter(rateLimitPerMinute, rateLimitPerHour)
	evidenceRL := newRateLimiter(20, 200)
	llmSem := newLLMSemaphore(maxConcurrentLLM)

	intakeChain := rateLimitMiddleware(rl, llmConcurrencyMiddleware(llmSem, http.HandlerFunc(s.handleIntake)))
	evidenceChain := rateLimitMiddleware(evidenceRL, http.HandlerFunc(s.handleAddEvidence))

	mux.Handle("POST /api/v1/cases", intakeChain)
	mux.HandleFunc("GET /api/v1/cases/{id}", s.handleGetCase)
	mux.HandleFunc("GET /api/v1/cases/{id}/document", s.handleGetCaseDocument)
	mux.Handle("POST /api/v1/cases/{id}/evidence", evidenceChain)
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	var handler http.Handler = mux
	handler = corsMiddleware(allowedOrigins, handler)
	handler = loggingMiddleware(s.Logger, handler)
	handler = recoverMiddleware(s.Logger, handler)

	return handler
}