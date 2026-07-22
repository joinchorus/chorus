package handler

import (
	"net/http"

	"chorus/internal/http/httputil"
	"chorus/internal/identity"
)

// IdentityHandler handles identity-related HTTP requests.
type IdentityHandler struct {
	service identity.Service
}

func NewIdentityHandler(service identity.Service) *IdentityHandler {
	return &IdentityHandler{service: service}
}

func (h *IdentityHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input identity.CreateInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, err)
		return
	}

	ident, err := h.service.Create(r.Context(), input)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, ident)
}

func (h *IdentityHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ident, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, ident)
}
