package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"chorus/internal/http/httputil"
)

var validRequestIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]{8,128}$`)

// RequestID attaches a correlation request ID to every incoming HTTP request.
// If the client provides a valid X-Request-ID header, it is preserved; otherwise, a unique ID is generated.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimSpace(r.Header.Get("X-Request-ID"))

		if reqID == "" || !validRequestIDRegex.MatchString(reqID) {
			reqID = generateRequestID()
		}

		// Attach to context and response headers
		ctx := httputil.WithRequestID(r.Context(), reqID)
		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return "req_" + hex.EncodeToString(b)
}
