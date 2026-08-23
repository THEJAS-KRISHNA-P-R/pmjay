package api

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// recoverMiddleware ensures a panic in any handler becomes a clean 500
// response instead of taking down the whole server process. Directly
// motivated by the same class of bug the spec's own background work
// found worth calling out ("goroutine panic recovery" was a real,
// separate fix needed on the actual production system this spec is
// modeled after) — an unrecovered panic in one request should never be
// able to affect any other in-flight request.
func recoverMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered", "panic", rec, "path", r.URL.Path, "method", r.Method)
				writeError(w, http.StatusInternalServerError, "an unexpected error occurred", "")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware emits one structured log line per request: method,
// path, status, and duration. Plain slog, no external logging
// dependency — matches this backend's zero-third-party-dependency
// design (see ../../../ARCHITECTURE.md).
func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusCapturingWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusCapturingWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCapturingWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// corsMiddleware allows only the configured frontend origin(s). No
// wildcard in production — the intake endpoint triggers a paid API call
// per request, so an open CORS policy would be an open invitation for
// someone else's website to spend this system's API budget.
func corsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimiter is a simple per-IP token bucket. Deliberately not a
// distributed rate limiter (no Redis, no external dependency) — a
// single-instance in-memory limiter is the right amount of protection
// for this system's actual scale, and it is a direct, zero-cost defense
// against the one thing that would otherwise let a bug or abusive
// traffic silently inflate the LLM bill (see config.RateLimitPerMinute's
// doc comment).
//
// bucketStaleAfter bounds the memory this limiter can consume: a key
// that hasn't been seen in this long is guaranteed to have refilled to
// full (any perMinute value refills well inside a minute), so dropping
// it is indistinguishable from that client making a fresh first request
// later — same behavior, but the map doesn't grow forever. Without this,
// a caller that varies its rate-limit key on every request (see
// clientIP's doc comment on X-Forwarded-For spoofing) turns this into an
// unbounded-memory-growth vector, not just a rate-limit bypass.
const bucketStaleAfter = 10 * time.Minute

// sweepInterval bounds how often allow() pays the cost of scanning the
// whole map for stale entries — once per interval is enough to keep
// memory bounded without doing real work on every single request.
const sweepInterval = time.Minute

type rateLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	perMinute int
	now       func() time.Time // overridable in tests
	lastSweep time.Time
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

func newRateLimiter(perMinute int) *rateLimiter {
	return &rateLimiter{
		buckets:   make(map[string]*bucket),
		perMinute: perMinute,
		now:       time.Now,
	}
}

// allow reports whether a request from key (typically a client IP) is
// within the rate limit right now, consuming one token if so.
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	rl.sweepLocked(now)

	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{tokens: float64(rl.perMinute) - 1, lastRefill: now}
		rl.buckets[key] = b
		return true
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	refillRate := float64(rl.perMinute) / 60.0
	b.tokens += elapsed * refillRate
	if b.tokens > float64(rl.perMinute) {
		b.tokens = float64(rl.perMinute)
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// sweepLocked removes buckets untouched for longer than bucketStaleAfter.
// Must be called with rl.mu already held. No-ops unless sweepInterval has
// elapsed since the last sweep, so the common case (allow() under normal
// traffic) stays O(1).
func (rl *rateLimiter) sweepLocked(now time.Time) {
	if !rl.lastSweep.IsZero() && now.Sub(rl.lastSweep) < sweepInterval {
		return
	}
	rl.lastSweep = now
	for k, b := range rl.buckets {
		if now.Sub(b.lastRefill) > bucketStaleAfter {
			delete(rl.buckets, k)
		}
	}
}

func rateLimitMiddleware(rl *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		if !rl.allow(key) {
			writeError(w, http.StatusTooManyRequests,
				"too many requests, please wait a moment and try again",
				"If this is urgent, call the PMJAY helpline directly at 14555.",
			)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the caller's IP for rate-limiting purposes, preferring
// a trusted reverse-proxy header if present (this system is designed to
// run behind exactly one reverse proxy hop, Caddy — see
// docs/DEPLOYMENT.md) and falling back to the direct connection's
// address.
//
// This reads the LAST entry in X-Forwarded-For, not the first, and that
// choice is load-bearing, not stylistic. A standards-compliant reverse
// proxy (Caddy included) APPENDS the connecting peer's address to any
// X-Forwarded-For it already received rather than replacing it — so for
// a request with no proxy involvement, the caller can put anything they
// want in that header themselves. If this function trusted the first
// (leftmost) entry, any client could send
// "X-Forwarded-For: 1.2.3.4" and be rate-limited as "1.2.3.4" instead of
// their real address — a trivial, free way to defeat the one control
// standing between this system and a runaway LLM bill, just by varying
// that one header on every request. The LAST entry is the one Caddy
// itself appended, which is not attacker-controlled as long as the
// documented topology holds: the Go process must never be reachable
// except through Caddy (docs/DEPLOYMENT.md's "one operational
// requirement either way"). If that requirement is ever violated —
// the Go port exposed directly — this protection is bypassable again,
// because there is no cryptographic way to distinguish a proxy-set
// header from a client-set one; the deployment topology is what makes
// trusting this header correct, not anything this function can verify
// on its own.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[len(parts)-1])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
