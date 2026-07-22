package handler

import (
	"context"
	"net/http"

	"chorus/internal/http/httputil"
	"chorus/internal/thread"
)

// ThreadService defines thread & message domain operations required by HTTP handlers.
type ThreadService interface {
	CreateThread(ctx context.Context, input thread.CreateThreadInput) (*thread.Thread, error)
	GetThreadByID(ctx context.Context, id string) (*thread.Thread, error)
	ListThreads(ctx context.Context) ([]*thread.Thread, error)

	AddMessage(ctx context.Context, threadID string, input thread.CreateMessageInput) (*thread.Message, error)
	ListMessages(ctx context.Context, threadID string) ([]*thread.Message, error)
}

// ThreadHandler handles thread & message HTTP requests.
type ThreadHandler struct {
	service ThreadService
}

// NewThreadHandler constructs a concrete ThreadHandler instance.
func NewThreadHandler(service ThreadService) *ThreadHandler {
	return &ThreadHandler{service: service}
}

func (h *ThreadHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	var input thread.CreateThreadInput
	if err := httputil.DecodeJSON(w, r, &input); err != nil {
		httputil.WriteError(w, err)
		return
	}

	t, err := h.service.CreateThread(r.Context(), input)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, t)
}

func (h *ThreadHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.service.GetThreadByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, t)
}

func (h *ThreadHandler) ListThreads(w http.ResponseWriter, r *http.Request) {
	threads, err := h.service.ListThreads(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, httputil.Envelope{"threads": threads})
}

func (h *ThreadHandler) AddMessage(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	var input thread.CreateMessageInput
	if err := httputil.DecodeJSON(w, r, &input); err != nil {
		httputil.WriteError(w, err)
		return
	}

	msg, err := h.service.AddMessage(r.Context(), threadID, input)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, msg)
}

func (h *ThreadHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	msgs, err := h.service.ListMessages(r.Context(), threadID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, httputil.Envelope{"messages": msgs})
}
