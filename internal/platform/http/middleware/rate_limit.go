package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/elite4print/elite4print-go/internal/platform/cache"
	"github.com/elite4print/elite4print-go/internal/platform/http/responses"
)

// RateLimiter uses a token-bucket algorithm backed by Redis.
//
// Why Redis?
// - Rate limits must be shared across multiple API instances.
// - A local in-memory limiter would be bypassed by load balancing.
type RateLimiter struct {
	cache  cache.Cache
	rps    int
	burst  int
	window time.Duration
}

// NewRateLimiter creates a rate limiter.
func NewRateLimiter(c cache.Cache, rps, burst int) *RateLimiter {
	return &RateLimiter{
		cache:  c,
		rps:    rps,
		burst:  burst,
		window: time.Second,
	}
}

// Handler limits requests per IP address.
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := "rate_limit:" + r.RemoteAddr

		// Simple fixed-window counter using Redis INCR + EXPIRE.
		val, err := rl.cache.Get(r.Context(), key)
		if err != nil {
			// On cache failure, allow the request rather than hard-failing.
			next.ServeHTTP(w, r)
			return
		}

		count, _ := strconv.Atoi(val)
		if count >= rl.burst {
			w.Header().Set("Retry-After", "1")
			responses.JSON(w, http.StatusTooManyRequests, responses.Error(http.StatusTooManyRequests, "RATE_LIMITED", "too many requests"))
			return
		}

		// We ignore errors here; if Redis is down we degrade gracefully.
		_ = rl.cache.Set(r.Context(), key, strconv.Itoa(count+1), int(rl.window.Seconds()))

		next.ServeHTTP(w, r)
	})
}
