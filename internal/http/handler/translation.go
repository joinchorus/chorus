package handler

import (
	"fmt"
	"net/http"

	"chorus/internal/domain"

	"chorus/internal/http/httputil"
	"chorus/internal/thread"
	"chorus/internal/translation"
)

// TranslationHandler handles message translation HTTP requests.
type TranslationHandler struct {
	transService *translation.Service
	threadService ThreadService
}

// NewTranslationHandler constructs a concrete TranslationHandler.
func NewTranslationHandler(transService *translation.Service, threadService ThreadService) *TranslationHandler {
	return &TranslationHandler{
		transService: transService,
		threadService: threadService,
	}
}

func (h *TranslationHandler) TranslateMessage(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	messageID := r.PathValue("msg_id")

	var input thread.TranslateMessageInput
	_ = httputil.DecodeJSON(w, r, &input)

	if input.TargetLang == "" {
		input.TargetLang = "en"
	}

	// 1. Fetch thread detail to locate target message content
	msgs, err := h.threadService.ListMessages(r.Context(), threadID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	var targetMsg *thread.Message
	for _, m := range msgs {
		if m.ID == messageID {
			targetMsg = m
			break
		}
	}

	if targetMsg == nil {
		httputil.WriteError(w, fmt.Errorf("%w: message %s not found in thread %s", domain.ErrNotFound, messageID, threadID))
		return
	}

	// 2. Delegate to translation service
	rec, err := h.transService.TranslateMessage(r.Context(), threadID, messageID, targetMsg.Content, input.TargetLang)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, rec)
}
