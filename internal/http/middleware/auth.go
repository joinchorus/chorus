package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"chorus/internal/domain"
	"chorus/internal/http/httputil"
)

const AdminCookieName = "chorus_admin_session"

// SessionManager manages ephemeral server-side moderator sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]time.Time // sessionID -> expiresAt
}

// NewSessionManager constructs a concrete in-memory SessionManager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]time.Time),
	}
}

// CreateSession generates a cryptographically random session identifier with the given TTL.
func (sm *SessionManager) CreateSession(ttl time.Duration) (string, error) {
	bytes := make([]byte, 32) // 256-bit cryptographic entropy
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random session bytes: %w", err)
	}
	sessionID := "adm_sess_" + hex.EncodeToString(bytes)

	sm.mu.Lock()
	sm.sessions[sessionID] = time.Now().Add(ttl)
	sm.mu.Unlock()

	return sessionID, nil
}

// ValidateSession verifies if a session ID is valid and not expired.
func (sm *SessionManager) ValidateSession(sessionID string) bool {
	if sessionID == "" {
		return false
	}

	sm.mu.RLock()
	expiresAt, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiresAt) {
		sm.mu.Lock()
		delete(sm.sessions, sessionID)
		sm.mu.Unlock()
		return false
	}

	return true
}

// RevokeSession explicitly destroys an active session on logout.
func (sm *SessionManager) RevokeSession(sessionID string) {
	if sessionID == "" {
		return
	}

	sm.mu.Lock()
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()
}

// RequireAdminAuth returns a middleware that validates HttpOnly session cookie or Authorization header.
func RequireAdminAuth(expectedToken string, sessionManager ...*SessionManager) Middleware {
	var sm *SessionManager
	if len(sessionManager) > 0 && sessionManager[0] != nil {
		sm = sessionManager[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedToken == "" {
				httputil.WriteError(w, domain.ErrUnauthorized)
				return
			}

			// 1. Check HttpOnly Admin Session Cookie (Browser flow)
			if cookie, err := r.Cookie(AdminCookieName); err == nil && cookie.Value != "" {
				sessionID := strings.TrimSpace(cookie.Value)
				if sm != nil && sm.ValidateSession(sessionID) {
					// CSRF Protection: For cookie-authenticated state-changing requests, validate Origin / Referer
					if isStateChangingMethod(r.Method) {
						if !isValidOriginOrReferer(r) {
							httputil.WriteError(w, domain.ErrForbidden)
							return
						}
					}
					next.ServeHTTP(w, r)
					return
				}
				// Cookie was provided but session is invalid or expired
				httputil.WriteError(w, domain.ErrUnauthorized)
				return
			}

			// 2. Check Authorization header (Bearer <token>) or X-Admin-* headers (API client flow)
			token, found := extractBearerOrCustomToken(r)
			if !found || token == "" {
				httputil.WriteError(w, domain.ErrUnauthorized)
				return
			}

			// Constant-time byte comparison for direct bearer API access
			if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
				httputil.WriteError(w, domain.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractBearerOrCustomToken(r *http.Request) (string, bool) {
	// 1. Check Authorization header (Bearer <token>)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1]), true
		}
		return strings.TrimSpace(authHeader), true
	}

	// 2. Fallback to X-Admin-License-Key or X-Admin-Token header
	if licenseHeader := r.Header.Get("X-Admin-License-Key"); licenseHeader != "" {
		return strings.TrimSpace(licenseHeader), true
	}
	if customHeader := r.Header.Get("X-Admin-Token"); customHeader != "" {
		return strings.TrimSpace(customHeader), true
	}

	return "", false
}

func isStateChangingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return true
	default:
		return false
	}
}

func isValidOriginOrReferer(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin != "" {
		return isAllowedOrigin(origin, r.Host)
	}

	referer := r.Header.Get("Referer")
	if referer != "" {
		if u, err := url.Parse(referer); err == nil {
			return isAllowedOrigin(u.Scheme+"://"+u.Host, r.Host)
		}
	}

	// If neither Origin nor Referer header is present, allow same-host local calls
	return true
}

func isAllowedOrigin(originStr, host string) bool {
	u, err := url.Parse(originStr)
	if err != nil {
		return false
	}

	// If origin host matches incoming Host header
	if strings.EqualFold(u.Host, host) {
		return true
	}

	// Allow standard Chorus origins
	allowedHosts := []string{"chat.joinchorus.app", "joinchorus.app", "localhost", "127.0.0.1"}
	originHostname := u.Hostname()
	for _, allowed := range allowedHosts {
		if strings.EqualFold(originHostname, allowed) {
			return true
		}
	}

	return false
}
