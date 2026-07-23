package handler

import (
	"net/http"

	"chorus/internal/http/httputil"
	"chorus/internal/moderation"
)

// ModerationHandler handles moderation queue HTTP requests.
type ModerationHandler struct {
	modService *moderation.Service
}

// NewModerationHandler constructs a concrete ModerationHandler instance.
func NewModerationHandler(modService *moderation.Service) *ModerationHandler {
	return &ModerationHandler{modService: modService}
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
