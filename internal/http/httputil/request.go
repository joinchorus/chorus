package httputil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"chorus/internal/domain"
)

const maxBodyBytes = 1048576 // 1MB limit

// DecodeJSON decodes request body JSON into target struct.
func DecodeJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: invalid JSON payload", domain.ErrValidation)
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: request body must only contain a single JSON object", domain.ErrValidation)
	}

	return nil
}
