package api

import "net/http"

// NewRouter builds the full HTTP handler chain: routes wrapped in
// middleware, in the order panic-recovery -> logging -> CORS ->
// (rate-limit only on the paid endpoint). Uses Go 1.22's standard
// library http.ServeMux, which gained method-and-pattern matching
// ("POST /api/v1/cases") specifically obviating the need for a
// third-party router for an API surface this size — see
// ../../../ARCHITECTURE.md for the fuller reasoning (this backend has zero
// external Go dependencies).
func NewRouter(s *Server, allowedOrigins []string, rateLimitPerMinute int) http.Handler {
	mux := http.NewServeMux()

	rl := newRateLimiter(rateLimitPerMinute)

	mux.Handle("POST /api/v1/cases", rateLimitMiddleware(rl, http.HandlerFunc(s.handleIntake)))
	mux.HandleFunc("GET /api/v1/cases/{id}", s.handleGetCase)
	mux.HandleFunc("GET /api/v1/cases/{id}/document", s.handleGetCaseDocument)
	mux.HandleFunc("POST /api/v1/cases/{id}/evidence", s.handleAddEvidence)
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	var handler http.Handler = mux
	handler = corsMiddleware(allowedOrigins, handler)
	handler = loggingMiddleware(s.Logger, handler)
	handler = recoverMiddleware(s.Logger, handler)

	return handler
}
