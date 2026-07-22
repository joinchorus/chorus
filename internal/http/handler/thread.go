package handler

import (
	"net/http"

	"chorus/internal/http/httputil"
	"chorus/internal/thread"
)

// ThreadHandler handles thread & message HTTP requests.
type ThreadHandler struct {
	service thread.Service
}

func NewThreadHandler(service thread.Service) *ThreadHandler {
	return &ThreadHandler{service: service}
}

func (h *ThreadHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	var input thread.CreateThreadInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
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
	if err := httputil.DecodeJSON(r, &input); err != nil {
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
