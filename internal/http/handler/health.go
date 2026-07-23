package handler

import (
	"net/http"

	"chorus/internal/http/httputil"
)

// HealthHandler handles health check requests.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, httputil.Envelope{
		"status": "ok",
	})
}
