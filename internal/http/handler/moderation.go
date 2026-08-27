package handler

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"chorus/internal/domain"
	"chorus/internal/http/httputil"
	"chorus/internal/http/middleware"
	"chorus/internal/moderation"
)

// ModerationHandler handles moderation queue HTTP requests.
type ModerationHandler struct {
	modService     *moderation.Service
	sessionManager *middleware.SessionManager
	adminToken     string
}

// NewModerationHandler constructs a concrete ModerationHandler instance.
func NewModerationHandler(modService *moderation.Service, sessionManager *middleware.SessionManager, adminToken ...string) *ModerationHandler {
	tok := ""
	if len(adminToken) > 0 {
		tok = adminToken[0]
	}
	if sessionManager == nil {
		sessionManager = middleware.NewSessionManager()
	}
	return &ModerationHandler{
		modService:     modService,
		sessionManager: sessionManager,
		adminToken:     tok,
	}
}

type LoginInput struct {
	Token string `json:"token"`
}

// Login authenticates a moderator and sets an HttpOnly session cookie with an opaque random session ID.
func (h *ModerationHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := httputil.DecodeJSON(w, r, &input); err != nil {
		httputil.WriteError(w, err)
		return
	}

	if h.adminToken == "" || subtle.ConstantTimeCompare([]byte(input.Token), []byte(h.adminToken)) != 1 {
		httputil.WriteError(w, domain.ErrForbidden)
		return
	}

	// Generate opaque cryptographically random session identifier (TTL: 7 days)
	sessionTTL := 86400 * 7 * time.Second
	sessionID, err := h.sessionManager.CreateSession(sessionTTL)
	if err != nil {
		httputil.WriteError(w, fmt.Errorf("%w: failed creating session", domain.ErrInternal))
		return
	}

	// Set HttpOnly, SameSite=Lax session cookie carrying ONLY the opaque session ID (NEVER the raw admin token)
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AdminCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})

	httputil.WriteJSON(w, http.StatusOK, httputil.Envelope{"status": "authenticated"})
}

// Logout revokes the session on the server and clears the moderator HttpOnly session cookie.
func (h *ModerationHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(middleware.AdminCookieName); err == nil && cookie.Value != "" {
		if h.sessionManager != nil {
			h.sessionManager.RevokeSession(cookie.Value)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AdminCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	httputil.WriteJSON(w, http.StatusOK, httputil.Envelope{"status": "logged_out"})
}

// GetSession returns session status for authenticated moderators.
func (h *ModerationHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, httputil.Envelope{"authenticated": true})
}

func (h *ModerationHandler) ListQueue(w http.ResponseWriter, r *http.Request) {
	items, err := h.modService.ListQueue(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, httputil.Envelope{"reports": items})
}

func (h *ModerationHandler) GetReportDetail(w http.ResponseWriter, r *http.Request) {
	reportID := r.PathValue("id")
	item, err := h.modService.GetQueueItemByID(r.Context(), reportID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *ModerationHandler) SubmitAction(w http.ResponseWriter, r *http.Request) {
	reportID := r.PathValue("id")
	var input moderation.SubmitActionInput
	if err := httputil.DecodeJSON(w, r, &input); err != nil {
		httputil.WriteError(w, err)
		return
	}

	act, err := h.modService.RecordAction(r.Context(), reportID, input)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, act)
}
