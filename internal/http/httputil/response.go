package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"chorus/internal/domain"
)

// Envelope for consistent API JSON responses.
type Envelope map[string]any

// APIError represents structured error response body.
type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
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
// Accepts optional *http.Request to populate request correlation ID.
func WriteError(w http.ResponseWriter, err error, r ...*http.Request) {
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
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		code = "unauthorized"
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		code = "forbidden"
	case errors.Is(err, domain.ErrAlreadyExists):
		status = http.StatusConflict
		code = "already_exists"
	default:
		status = http.StatusInternalServerError
		code = "internal_error"
		message = "internal server error"
	}

	var reqID string
	if len(r) > 0 && r[0] != nil {
		reqID = GetRequestID(r[0].Context())
	}

	WriteJSON(w, status, ErrorResponse{
		Error: APIError{
			Code:      code,
			Message:   message,
			RequestID: reqID,
		},
	})
}
