package handler

import (
	"context"
	"net/http"
	"strconv"

	"chorus/internal/geoip"
	"chorus/internal/http/httputil"
	"chorus/internal/thread"
)

// ThreadService defines thread & message domain operations required by HTTP handlers.
type ThreadService interface {
	CreateThread(ctx context.Context, input thread.CreateThreadInput, resolvedCountry string) (*thread.Thread, error)
	GetThreadByID(ctx context.Context, id string) (*thread.Thread, error)
	GetThreadDetail(ctx context.Context, id string) (*thread.ThreadDetail, error)
	ListThreads(ctx context.Context) ([]*thread.Thread, error)

	AddMessage(ctx context.Context, threadID string, input thread.CreateMessageInput, resolvedCountry string) (*thread.Message, error)
	ListMessages(ctx context.Context, threadID string) ([]*thread.Message, error)
}

// ThreadHandler handles thread & message HTTP requests.
type ThreadHandler struct {
	service ThreadService
	geoSvc  *geoip.Service
}

// NewThreadHandler constructs a concrete ThreadHandler instance.
func NewThreadHandler(service ThreadService, geoSvc *geoip.Service) *ThreadHandler {
	if geoSvc == nil {
		geoSvc = geoip.NewService("TR")
	}
	return &ThreadHandler{
		service: service,
		geoSvc:  geoSvc,
	}
}

func (h *ThreadHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	var input thread.CreateThreadInput
	if err := httputil.DecodeJSON(w, r, &input); err != nil {
		httputil.WriteError(w, err)
		return
	}

	resolvedCountry := h.geoSvc.ResolveCountryFromRequest(r.Context(), r)
	t, err := h.service.CreateThread(r.Context(), input, resolvedCountry)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, t)
}

func (h *ThreadHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detail, err := h.service.GetThreadDetail(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, detail)
}

func (h *ThreadHandler) ListThreads(w http.ResponseWriter, r *http.Request) {
	threads, err := h.service.ListThreads(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	offset := parseQueryInt(r, "offset", 0)
	limit := parseQueryInt(r, "limit", 50)

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 50
	}

	if offset >= len(threads) {
		threads = []*thread.Thread{}
	} else {
		end := offset + limit
		if end > len(threads) {
			end = len(threads)
		}
		threads = threads[offset:end]
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

	resolvedCountry := h.geoSvc.ResolveCountryFromRequest(r.Context(), r)
	msg, err := h.service.AddMessage(r.Context(), threadID, input, resolvedCountry)
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

	offset := parseQueryInt(r, "offset", 0)
	limit := parseQueryInt(r, "limit", 100)

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	}

	if offset >= len(msgs) {
		msgs = []*thread.Message{}
	} else {
		end := offset + limit
		if end > len(msgs) {
			end = len(msgs)
		}
		msgs = msgs[offset:end]
	}

	httputil.WriteJSON(w, http.StatusOK, httputil.Envelope{"messages": msgs})
}

func parseQueryInt(r *http.Request, key string, defaultValue int) int {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}
