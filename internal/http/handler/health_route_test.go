package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	chttp "chorus/internal/http"
	"chorus/internal/http/handler"
)

// Hosting in front of the app answers /healthz itself, so the build commit has to
// be readable on a path under /api.
func TestVersionedHealthRouteReportsCommit(t *testing.T) {
	router := chttp.NewRouter(chttp.RouterConfig{
		Health: handler.NewHealthHandler().WithCommit("abc1234"),
	})

	for _, path := range []string{"/api/v1/health", "/api/v0.1/health"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200", path, rec.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("GET %s: invalid JSON: %v", path, err)
		}
		if body["status"] != "ok" || body["commit"] != "abc1234" {
			t.Errorf("GET %s = %v, want status ok and commit abc1234", path, body)
		}
	}
}
