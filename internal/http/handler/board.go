package handler

import (
	"net/http"

	"chorus/internal/board"
	"chorus/internal/domain"
	"chorus/internal/http/httputil"
)

// BoardHandler handles board HTTP requests.
type BoardHandler struct{}

// NewBoardHandler constructs a new BoardHandler instance.
func NewBoardHandler() *BoardHandler {
	return &BoardHandler{}
}

// ListBoards returns all system-curated boards.
func (h *BoardHandler) ListBoards(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, httputil.Envelope{"boards": board.SystemBoards})
}

// GetBoard returns details of a single board by its slug.
func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	b := board.GetBoardBySlug(slug)
	if b == nil {
		httputil.WriteError(w, domain.ErrNotFound)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, b)
}
