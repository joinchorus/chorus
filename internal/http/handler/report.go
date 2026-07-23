package handler

import (
	"fmt"
	"net/http"

	"chorus/internal/domain"
	"chorus/internal/http/httputil"
	"chorus/internal/reporting"
)


// ReportHandler handles message reporting HTTP requests.
type ReportHandler struct {
	reportService *reporting.Service
	threadService ThreadService
}

// NewReportHandler constructs a concrete ReportHandler instance.
func NewReportHandler(reportService *reporting.Service, threadService ThreadService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		threadService: threadService,
	}
}

func (h *ReportHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	messageID := r.PathValue("msg_id")

	var input reporting.CreateReportInput
	if err := httputil.DecodeJSON(w, r, &input); err != nil {
		httputil.WriteError(w, err)
		return
	}

	// 1. Verify message exists in thread
	msgs, err := h.threadService.ListMessages(r.Context(), threadID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	found := false
	for _, m := range msgs {
		if m.ID == messageID {
			found = true
			break
		}
	}
	if !found {
		httputil.WriteError(w, fmt.Errorf("%w: message %s not found in thread %s", domain.ErrNotFound, messageID, threadID))
		return
	}

	// 2. Delegate to reporting service
	rpt, err := h.reportService.SubmitReport(r.Context(), threadID, messageID, input)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, rpt)
}
