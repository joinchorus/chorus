package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chorus/internal/http/middleware"
)

func TestSecurityHeaders(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// 1. Verify modern headers
	if nosniff := rec.Header().Get("X-Content-Type-Options"); nosniff != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %s", nosniff)
	}

	if frameOpt := rec.Header().Get("X-Frame-Options"); frameOpt != "DENY" {
		t.Errorf("expected X-Frame-Options: DENY, got %s", frameOpt)
	}

	if coop := rec.Header().Get("Cross-Origin-Opener-Policy"); coop != "same-origin" {
		t.Errorf("expected COOP: same-origin, got %s", coop)
	}

	if corp := rec.Header().Get("Cross-Origin-Resource-Policy"); corp != "same-origin" {
		t.Errorf("expected CORP: same-origin, got %s", corp)
	}

	if perm := rec.Header().Get("Permissions-Policy"); !strings.Contains(perm, "camera=()") {
		t.Errorf("expected Permissions-Policy containing camera=(), got %s", perm)
	}

	// 2. Verify obsolete X-XSS-Protection is NOT set
	if xss := rec.Header().Get("X-XSS-Protection"); xss != "" {
		t.Errorf("expected obsolete X-XSS-Protection header to be absent, got %s", xss)
	}

	// 3. Verify CSP contains essential directives
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "frame-ancestors 'none'") {
		t.Errorf("expected CSP containing frame-ancestors 'none', got %s", csp)
	}
	if !strings.Contains(csp, "worker-src 'self' blob:") {
		t.Errorf("expected CSP containing worker-src for Monaco/WebWorker support, got %s", csp)
	}

	// 4. Verify HSTS header on HTTPS request
	httpsReq := httptest.NewRequest("GET", "/", nil)
	httpsReq.Header.Set("X-Forwarded-Proto", "https")
	httpsRec := httptest.NewRecorder()
	handler.ServeHTTP(httpsRec, httpsReq)

	if hsts := httpsRec.Header().Get("Strict-Transport-Security"); !strings.Contains(hsts, "max-age=31536000") {
		t.Errorf("expected HSTS header on HTTPS request, got %s", hsts)
	}
}
