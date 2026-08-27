package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"chorus/internal/http/middleware"
)

func TestSessionManager_Lifecycle(t *testing.T) {
	sm := middleware.NewSessionManager()

	// 1. Create Session
	sessionID, err := sm.CreateSession(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("failed creating session: %v", err)
	}

	if !strings.HasPrefix(sessionID, "adm_sess_") {
		t.Errorf("expected sessionID to start with adm_sess_, got %s", sessionID)
	}

	// 2. Validate valid session
	if !sm.ValidateSession(sessionID) {
		t.Errorf("expected session %s to be valid", sessionID)
	}

	// 3. Validate non-existent session
	if sm.ValidateSession("adm_sess_nonexistent") {
		t.Errorf("expected non-existent session to be invalid")
	}

	// 4. Test Revocation
	sm.RevokeSession(sessionID)
	if sm.ValidateSession(sessionID) {
		t.Errorf("expected revoked session to be invalid")
	}

	// 5. Test Expiry
	shortSessionID, err := sm.CreateSession(10 * time.Millisecond)
	if err != nil {
		t.Fatalf("failed creating short-lived session: %v", err)
	}
	time.Sleep(25 * time.Millisecond)
	if sm.ValidateSession(shortSessionID) {
		t.Errorf("expected expired session to be invalid")
	}
}

func TestRequireAdminAuth_BearerAndSession(t *testing.T) {
	const secretToken = "my-secret-admin-token"
	sm := middleware.NewSessionManager()

	validSessionID, _ := sm.CreateSession(1 * time.Hour)
	expiredSessionID, _ := sm.CreateSession(1 * time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	authMiddleware := middleware.RequireAdminAuth(secretToken, sm)
	protected := authMiddleware(nextHandler)

	tests := []struct {
		name           string
		method         string
		headerKey      string
		headerVal      string
		cookieVal      string
		origin         string
		expectedStatus int
	}{
		{
			name:           "missing auth headers returns 401",
			method:         "GET",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid bearer token returns 403",
			method:         "GET",
			headerKey:      "Authorization",
			headerVal:      "Bearer wrong-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "valid bearer token returns 200",
			method:         "GET",
			headerKey:      "Authorization",
			headerVal:      "Bearer " + secretToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid session cookie returns 200 for GET",
			method:         "GET",
			cookieVal:      validSessionID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "expired session cookie returns 401",
			method:         "GET",
			cookieVal:      expiredSessionID,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid session cookie returns 401",
			method:         "GET",
			cookieVal:      "adm_sess_invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "raw secret in cookie is rejected if not registered in session manager",
			method:         "GET",
			cookieVal:      secretToken, // raw token directly in cookie
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "cookie-authenticated POST with matching origin returns 200",
			method:         "POST",
			cookieVal:      validSessionID,
			origin:         "https://chat.joinchorus.app",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "cookie-authenticated POST with cross-origin malicious origin returns 403 CSRF",
			method:         "POST",
			cookieVal:      validSessionID,
			origin:         "https://malicious-attacker.com",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "bearer-authenticated POST with cross-origin origin returns 200 (exempt from CSRF)",
			method:         "POST",
			headerKey:      "Authorization",
			headerVal:      "Bearer " + secretToken,
			origin:         "https://external-api-consumer.com",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Host = "chat.joinchorus.app"

			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerVal)
			}
			if tt.cookieVal != "" {
				req.AddCookie(&http.Cookie{
					Name:  middleware.AdminCookieName,
					Value: tt.cookieVal,
				})
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
