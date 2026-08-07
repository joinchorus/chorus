package middleware

import (
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"chorus/internal/http/httputil"
)

type clientBucket struct {
	tokens     float64
	lastUpdate time.Time
}

// IPRateLimiter manages per-IP token bucket limiters cleanly in memory.
type IPRateLimiter struct {
	mu         sync.Mutex
	clients    map[string]*clientBucket
	rate       float64 // tokens added per second
	burst      float64 // maximum bucket capacity
	lastPurge  time.Time
}

// NewIPRateLimiter constructs a thread-safe token bucket rate limiter for client IPs.
// ratePerSec specifies how many tokens are added per second, and burst is the max capacity.
func NewIPRateLimiter(ratePerSec float64, burst int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients:   make(map[string]*clientBucket),
		rate:      ratePerSec,
		burst:     float64(burst),
		lastPurge: time.Now(),
	}
	return limiter
}

// Allow checks if the client IP is allowed to proceed. Returns (allowed, retryAfterSeconds).
func (l *IPRateLimiter) Allow(ip string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Periodic purge of inactive IP entries every 10 minutes to prevent memory leaks
	if now.Sub(l.lastPurge) > 10*time.Minute {
		for clientIP, b := range l.clients {
			if now.Sub(b.lastUpdate) > 15*time.Minute {
				delete(l.clients, clientIP)
			}
		}
		l.lastPurge = now
	}

	b, exists := l.clients[ip]
	if !exists {
		l.clients[ip] = &clientBucket{
			tokens:     l.burst - 1.0, // consume 1 token for current request
			lastUpdate: now,
		}
		return true, 0
	}

	// Calculate tokens accumulated since last update
	elapsed := now.Sub(b.lastUpdate).Seconds()
	b.tokens = math.Min(l.burst, b.tokens+(elapsed*l.rate))
	b.lastUpdate = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true, 0
	}

	// Calculate retry after time in seconds
	needed := 1.0 - b.tokens
	retryAfter := math.Ceil(needed / l.rate)
	if retryAfter < 1 {
		retryAfter = 1
	}

	return false, int(retryAfter)
}

// RateLimit returns an HTTP middleware enforcing rate limits per client IP.
func RateLimit(limiter *IPRateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := httputil.ExtractClientIP(r)
			allowed, retryAfter := limiter.Allow(ip)

			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				httputil.WriteJSON(w, http.StatusTooManyRequests, httputil.ErrorResponse{
					Error: httputil.APIError{
						Code:    "rate_limited",
						Message: fmt.Sprintf("rate limit exceeded, please try again in %d seconds", retryAfter),
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
