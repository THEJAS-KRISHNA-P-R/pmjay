package api

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

const (
	bucketStaleAfter = 10 * time.Minute
	sweepInterval    = time.Minute
)

type rateLimiter struct {
	mu           sync.Mutex
	minuteBuckets map[string]*bucket
	hourBuckets   map[string]*bucket
	perMinute     int
	perHour       int
	now           func() time.Time
	lastSweep     time.Time
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

func newRateLimiter(perMinute, perHour int) *rateLimiter {
	return &rateLimiter{
		minuteBuckets: make(map[string]*bucket),
		hourBuckets:   make(map[string]*bucket),
		perMinute:     perMinute,
		perHour:       perHour,
		now:           time.Now,
	}
}

func (rl *rateLimiter) checkAndReserve(key string) (allowed bool, retryAfterSec int, remainingMinute, remainingHour int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	rl.sweepLocked(now)

	minB, minExists := rl.minuteBuckets[key]
	if !minExists {
		minB = &bucket{tokens: float64(rl.perMinute), lastRefill: now}
		rl.minuteBuckets[key] = minB
	}

	hourB, hourExists := rl.hourBuckets[key]
	if !hourExists {
		hourB = &bucket{tokens: float64(rl.perHour), lastRefill: now}
		rl.hourBuckets[key] = hourB
	}

	rl.refillBucket(minB, now, float64(rl.perMinute)/60.0)
	rl.refillBucket(hourB, now, float64(rl.perHour)/3600.0)

	remainingMinute = int(minB.tokens)
	remainingHour = int(hourB.tokens)

	if minB.tokens < 1 || hourB.tokens < 1 {
		retryAfterSec = int(min(60-minB.lastRefill.Sub(now).Seconds(), 3600-hourB.lastRefill.Sub(now).Seconds()))
		if retryAfterSec < 1 {
			retryAfterSec = 1
		}
		return false, retryAfterSec, remainingMinute, remainingHour
	}

	return true, 0, remainingMinute, remainingHour
}

func (rl *rateLimiter) consume(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	rl.sweepLocked(now)

	if b, ok := rl.minuteBuckets[key]; ok {
		rl.refillBucket(b, now, float64(rl.perMinute)/60.0)
		if b.tokens > 0 {
			b.tokens--
		}
	}
	if b, ok := rl.hourBuckets[key]; ok {
		rl.refillBucket(b, now, float64(rl.perHour)/3600.0)
		if b.tokens > 0 {
			b.tokens--
		}
	}
}

func (rl *rateLimiter) refillBucket(b *bucket, now time.Time, ratePerSec float64) {
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * ratePerSec
	maxTokens := float64(rl.perMinute)
	if ratePerSec < 1 {
		maxTokens = float64(rl.perHour)
	}
	if b.tokens > maxTokens {
		b.tokens = maxTokens
	}
	b.lastRefill = now
}

func (rl *rateLimiter) sweepLocked(now time.Time) {
	if !rl.lastSweep.IsZero() && now.Sub(rl.lastSweep) < sweepInterval {
		return
	}
	rl.lastSweep = now
	for k, b := range rl.minuteBuckets {
		if now.Sub(b.lastRefill) > bucketStaleAfter {
			delete(rl.minuteBuckets, k)
		}
	}
	for k, b := range rl.hourBuckets {
		if now.Sub(b.lastRefill) > bucketStaleAfter {
			delete(rl.hourBuckets, k)
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

type rateLimitResponseWriter struct {
	http.ResponseWriter
	status       int
	key          string
	limiter      *rateLimiter
	retryAfter   int
	remainingMin int
	remainingHour int
}

func (w *rateLimitResponseWriter) WriteHeader(status int) {
	w.status = status
	if status >= 200 && status < 300 {
		w.limiter.consume(w.key)
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *rateLimitResponseWriter) Header() http.Header {
	h := w.ResponseWriter.Header()
	h.Set("X-RateLimit-Limit-Minute", strconv.Itoa(w.limiter.perMinute))
	h.Set("X-RateLimit-Limit-Hour", strconv.Itoa(w.limiter.perHour))
	h.Set("X-RateLimit-Remaining-Minute", strconv.Itoa(max(0, w.remainingMin)))
	h.Set("X-RateLimit-Remaining-Hour", strconv.Itoa(max(0, w.remainingHour)))
	if w.retryAfter > 0 {
		h.Set("Retry-After", strconv.Itoa(w.retryAfter))
		h.Set("X-RateLimit-Reset", strconv.Itoa(int(time.Now().Unix())+w.retryAfter))
	}
	return h
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func rateLimitMiddleware(rl *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		allowed, retryAfter, remMin, remHour := rl.checkAndReserve(key)

		w.Header().Set("X-RateLimit-Limit-Minute", strconv.Itoa(rl.perMinute))
		w.Header().Set("X-RateLimit-Limit-Hour", strconv.Itoa(rl.perHour))
		w.Header().Set("X-RateLimit-Remaining-Minute", strconv.Itoa(max(0, remMin)))
		w.Header().Set("X-RateLimit-Remaining-Hour", strconv.Itoa(max(0, remHour)))

		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(time.Now().Unix())+retryAfter))
			writeError(w, http.StatusTooManyRequests,
				"too many requests, please wait a moment and try again",
				"If this is urgent, call the PMJAY helpline directly at 14555.",
			)
			return
		}

		rw := &rateLimitResponseWriter{
			ResponseWriter: w,
			key:            key,
			limiter:        rl,
			retryAfter:     0,
			remainingMin:   remMin,
			remainingHour:  remHour,
		}
		next.ServeHTTP(rw, r)
	})
}

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

type llmSemaphore struct {
	ch chan struct{}
}

func newLLMSemaphore(maxConcurrent int) *llmSemaphore {
	return &llmSemaphore{ch: make(chan struct{}, maxConcurrent)}
}

func (s *llmSemaphore) Acquire() {
	s.ch <- struct{}{}
}

func (s *llmSemaphore) Release() {
	<-s.ch
}

func llmConcurrencyMiddleware(sem *llmSemaphore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sem.Acquire()
		defer sem.Release()
		next.ServeHTTP(w, r)
	})
}