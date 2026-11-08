package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chttp "chorus/internal/http"
	"chorus/internal/http/handler"
)

func get(t *testing.T, h http.Handler, target string, headers map[string]string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	body, _ := io.ReadAll(rec.Result().Body)
	return rec.Code, string(body)
}

// "/landing.." starts with the "/landing" prefix, so the remainder "../docs/x"
// used to be joined under public/landing and climbed into public/docs.
func TestStaticRoutesDoNotClimbOutOfTheirDirectory(t *testing.T) {
	router := chttp.NewRouter(chttp.RouterConfig{
		Health:    handler.NewHealthHandler(),
		StaticDir: t.TempDir(),
	})

	_, docs := get(t, router, "/docs/index.html", nil)
	if strings.TrimSpace(docs) == "" {
		t.Skip("public/docs/index.html not found from this test directory")
	}

	for _, target := range []string{
		"/landing../docs/index.html",
		"/docs../landing/index.html",
		"/landing..%2Fdocs%2Findex.html",
	} {
		_, body := get(t, router, target, nil)
		if body == docs && strings.HasPrefix(target, "/landing") {
			t.Errorf("%s served the docs page from outside public/landing", target)
		}
	}

	_, landing := get(t, router, "/landing/index.html", nil)
	for _, target := range []string{"/docs../landing/index.html"} {
		_, body := get(t, router, target, nil)
		if body == landing && landing != docs {
			t.Errorf("%s served the landing page from outside public/docs", target)
		}
	}
}

// The logout cookie must carry Secure whenever the login cookie does.
func TestLogoutCookieIsSecureBehindHTTPS(t *testing.T) {
	h := handler.NewModerationHandler(nil, nil)

	cookies := func(headers map[string]string) []*http.Cookie {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/moderation/logout", nil)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rec := httptest.NewRecorder()
		h.Logout(rec, req)
		return rec.Result().Cookies()
	}

	secure := cookies(map[string]string{"X-Forwarded-Proto": "https"})
	if len(secure) != 1 || !secure[0].Secure {
		t.Errorf("logout cookie over HTTPS must be Secure, got %+v", secure)
	}
	plain := cookies(nil)
	if len(plain) != 1 || plain[0].Secure {
		t.Errorf("logout cookie over plain HTTP must not be Secure (local development), got %+v", plain)
	}
}

// A percent-encoded slash survives ServeMux and reaches the handler as part of
// the {id} value; it must come back as a client error, not touch the disk.
func TestThreadRoutesRejectEncodedPathIDs(t *testing.T) {
	router := setupTestServer()
	for _, target := range []string{
		"/api/v1/threads/thd_%2F..%2F..%2Fx",
		"/api/v1/threads/thd_%2F..%2F..%2Fx/messages",
	} {
		code, _ := get(t, router, target, nil)
		if code != http.StatusBadRequest && code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 400 or 404", target, code)
		}
	}
}
