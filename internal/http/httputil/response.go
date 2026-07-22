package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"chorus/internal/domain"
)

// Envelope for consistent API JSON responses.
type Envelope map[string]any

// WriteJSON writes a JSON response with status code and payload.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteError inspects the error and writes the appropriate HTTP error response.
func WriteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		WriteJSON(w, http.StatusNotFound, Envelope{"error": err.Error()})
	case errors.Is(err, domain.ErrValidation):
		WriteJSON(w, http.StatusBadRequest, Envelope{"error": err.Error()})
	case errors.Is(err, domain.ErrAlreadyExists):
		WriteJSON(w, http.StatusConflict, Envelope{"error": err.Error()})
	default:
		WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "internal server error"})
	}
}
