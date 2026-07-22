package handler

import (
	"context"
	"net/http"

	"chorus/internal/http/httputil"
	"chorus/internal/identity"
)

// IdentityService defines identity domain operations required by HTTP handlers.
type IdentityService interface {
	Create(ctx context.Context, input identity.CreateInput) (*identity.Identity, error)
	GetByID(ctx context.Context, id string) (*identity.Identity, error)
}

// IdentityHandler handles identity-related HTTP requests.
type IdentityHandler struct {
	service IdentityService
}

// NewIdentityHandler constructs a concrete IdentityHandler instance.
func NewIdentityHandler(service IdentityService) *IdentityHandler {
	return &IdentityHandler{service: service}
}

func (h *IdentityHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input identity.CreateInput
	if err := httputil.DecodeJSON(w, r, &input); err != nil {
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
