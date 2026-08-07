package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"chorus/internal/http/httputil"
	"chorus/internal/http/middleware"
)

func TestRequestIDMiddleware_GeneratesNewID(t *testing.T) {
	var capturedID string
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = httputil.GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resID := rec.Header().Get("X-Request-ID")
	if resID == "" {
		t.Fatalf("expected X-Request-ID header in response")
	}

	if capturedID != resID {
		t.Errorf("expected context Request ID %s to match header %s", capturedID, resID)
	}

	if len(resID) < 10 {
		t.Errorf("expected generated request ID to have reasonable length, got %s", resID)
	}
}

func TestRequestIDMiddleware_PreservesValidClientHeader(t *testing.T) {
	customID := "client_request_uuid_12345"
	var capturedID string
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = httputil.GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/healthz", nil)
	req.Header.Set("X-Request-ID", customID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resID := rec.Header().Get("X-Request-ID")
	if resID != customID {
		t.Errorf("expected preserved header %s, got %s", customID, resID)
	}

	if capturedID != customID {
		t.Errorf("expected context Request ID %s to match client header %s", capturedID, customID)
	}
}
