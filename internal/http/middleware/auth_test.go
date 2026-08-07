package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"chorus/internal/http/middleware"
)

func TestRequireAdminAuth(t *testing.T) {
	const secretToken = "my-secret-admin-token"

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	authMiddleware := middleware.RequireAdminAuth(secretToken)
	protected := authMiddleware(nextHandler)

	tests := []struct {
		name           string
		headerKey      string
		headerVal      string
		expectedStatus int
	}{
		{
			name:           "missing auth headers returns 401",
			headerKey:      "",
			headerVal:      "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid bearer token returns 403",
			headerKey:      "Authorization",
			headerVal:      "Bearer wrong-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid custom header returns 403",
			headerKey:      "X-Admin-Token",
			headerVal:      "wrong-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "valid bearer token returns 200",
			headerKey:      "Authorization",
			headerVal:      "Bearer " + secretToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid X-Admin-Token returns 200",
			headerKey:      "X-Admin-Token",
			headerVal:      secretToken,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerVal)
			}
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
