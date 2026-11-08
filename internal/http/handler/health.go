package handler

import (
	"net/http"

	"chorus/internal/http/httputil"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	commit string
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// WithCommit sets the source commit reported by Check, so a running instance
// can be matched to the code it was built from.
func (h *HealthHandler) WithCommit(commit string) *HealthHandler {
	h.commit = commit
	return h
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	body := httputil.Envelope{"status": "ok"}
	if h.commit != "" {
		body["commit"] = h.commit
	}
	httputil.WriteJSON(w, http.StatusOK, body)
}
