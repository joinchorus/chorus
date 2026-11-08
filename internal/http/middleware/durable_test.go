package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"chorus/internal/http/middleware"
)

type countingFlusher struct {
	calls atomic.Int32
	err   error
}

func (c *countingFlusher) Flush(context.Context) error {
	c.calls.Add(1)
	return c.err
}

func serve(t *testing.T, f middleware.Flusher, method string, status int) *httptest.ResponseRecorder {
	t.Helper()
	h := middleware.DurableWrites(f)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, "/api/v1/threads", nil))
	return rec
}

func TestDurableWritesFlushesAfterSuccessfulWrites(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		f := &countingFlusher{}
		if rec := serve(t, f, method, http.StatusCreated); rec.Code != http.StatusCreated {
			t.Fatalf("%s status = %d", method, rec.Code)
		}
		if got := f.calls.Load(); got != 1 {
			t.Errorf("%s: flushes = %d, want 1", method, got)
		}
	}
}

func TestDurableWritesSkipsReadsAndFailedWrites(t *testing.T) {
	for _, tc := range []struct {
		method string
		status int
	}{
		{http.MethodGet, http.StatusOK},
		{http.MethodHead, http.StatusOK},
		{http.MethodOptions, http.StatusNoContent},
		{http.MethodPost, http.StatusBadRequest},
		{http.MethodPost, http.StatusInternalServerError},
	} {
		f := &countingFlusher{}
		serve(t, f, tc.method, tc.status)
		if got := f.calls.Load(); got != 0 {
			t.Errorf("%s %d: flushes = %d, want 0", tc.method, tc.status, got)
		}
	}
}

func TestDurableWritesKeepsTheResponseWhenFlushFails(t *testing.T) {
	f := &countingFlusher{err: errors.New("storage down")}
	if rec := serve(t, f, http.MethodPost, http.StatusCreated); rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, a failed backup must not change the response", rec.Code)
	}
}

func TestDurableWritesWithoutFlusherPassesThrough(t *testing.T) {
	if rec := serve(t, nil, http.MethodPost, http.StatusCreated); rec.Code != http.StatusCreated {
		t.Fatalf("status = %d", rec.Code)
	}
}
