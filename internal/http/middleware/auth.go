package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"chorus/internal/domain"
	"chorus/internal/http/httputil"
)

const AdminCookieName = "chorus_admin_session"

// RequireAdminAuth returns a middleware that validates HttpOnly session cookie or Authorization header.
func RequireAdminAuth(expectedToken string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedToken == "" {
				httputil.WriteError(w, domain.ErrUnauthorized)
				return
			}

			token, found := extractTokenOrCookie(r)
			if !found || token == "" {
				httputil.WriteError(w, domain.ErrUnauthorized)
				return
			}

			// Constant-time byte comparison to protect against side-channel timing attacks
			if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
				httputil.WriteError(w, domain.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractTokenOrCookie(r *http.Request) (string, bool) {
	// 1. Check HttpOnly Admin Session Cookie
	if cookie, err := r.Cookie(AdminCookieName); err == nil && cookie.Value != "" {
		return strings.TrimSpace(cookie.Value), true
	}

	// 2. Check Authorization header (Bearer <token>)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1]), true
		}
		return strings.TrimSpace(authHeader), true
	}

	// 3. Fallback to X-Admin-License-Key or X-Admin-Token header
	if licenseHeader := r.Header.Get("X-Admin-License-Key"); licenseHeader != "" {
		return strings.TrimSpace(licenseHeader), true
	}
	if customHeader := r.Header.Get("X-Admin-Token"); customHeader != "" {
		return strings.TrimSpace(customHeader), true
	}

	return "", false
}
