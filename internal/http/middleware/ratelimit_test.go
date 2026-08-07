package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chorus/internal/http/middleware"
)

func TestRateLimiter_BurstAndThrottle(t *testing.T) {
	// Rate limiter: 1 token/sec, burst capacity of 2
	limiter := middleware.NewIPRateLimiter(1.0, 2)
	rateGuard := middleware.RateLimit(limiter)

	handler := rateGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	ip := "192.168.1.100"

	makeRequest := func() int {
		req := httptest.NewRequest("POST", "/api/v1/threads", nil)
		req.RemoteAddr = ip + ":12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	// 1st request -> PASS (200)
	if code := makeRequest(); code != http.StatusOK {
		t.Errorf("request 1 expected 200, got %d", code)
	}

	// 2nd request -> PASS (200) - uses burst token
	if code := makeRequest(); code != http.StatusOK {
		t.Errorf("request 2 expected 200, got %d", code)
	}

	// 3rd request -> BLOCKED (429 Too Many Requests)
	req := httptest.NewRequest("POST", "/api/v1/threads", nil)
	req.RemoteAddr = ip + ":12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("request 3 expected 429 Too Many Requests, got %d", rec.Code)
	}

	if retryAfter := rec.Header().Get("Retry-After"); retryAfter == "" {
		t.Errorf("expected Retry-After header on 429 response")
	}

	// Wait 1.1 seconds for token replenishment
	time.Sleep(1100 * time.Millisecond)

	// 4th request -> PASS (200)
	if code := makeRequest(); code != http.StatusOK {
		t.Errorf("request 4 expected 200 after sleep, got %d", code)
	}
}
