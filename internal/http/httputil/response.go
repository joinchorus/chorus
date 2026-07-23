package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"chorus/internal/domain"
)

// Envelope for consistent API JSON responses.
type Envelope map[string]any

// APIError represent structured error response body.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse envelope containing structured API error.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

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
	var status int
	var code string
	message := err.Error()

	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, domain.ErrValidation):
		status = http.StatusBadRequest
		code = "validation_error"
	case errors.Is(err, domain.ErrAlreadyExists):
		status = http.StatusConflict
		code = "already_exists"
	default:
		status = http.StatusInternalServerError
		code = "internal_error"
		message = "internal server error"
	}

	WriteJSON(w, status, ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	})
}
