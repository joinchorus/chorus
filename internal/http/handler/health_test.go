package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chorus/internal/http/handler"
)

func healthBody(t *testing.T, h *handler.HealthHandler) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	h.Check(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return body
}

func TestHealthReportsCommitWhenSet(t *testing.T) {
	body := healthBody(t, handler.NewHealthHandler().WithCommit("abc1234"))
	if body["status"] != "ok" {
		t.Errorf("status = %v, want ok", body["status"])
	}
	if body["commit"] != "abc1234" {
		t.Errorf("commit = %v, want abc1234", body["commit"])
	}
}

func TestHealthOmitsCommitWhenUnset(t *testing.T) {
	body := healthBody(t, handler.NewHealthHandler())
	if _, ok := body["commit"]; ok {
		t.Errorf("commit must be omitted when not set, got %v", body["commit"])
	}
}
