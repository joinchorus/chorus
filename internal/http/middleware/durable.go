package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"chorus/internal/http/httputil"
)

// Flusher persists data written during a request (see internal/backup).
type Flusher interface {
	Flush(ctx context.Context) error
}

// durableWriteTimeout stays below the server's write timeout so a slow upload
// cannot turn into a dropped connection.
const durableWriteTimeout = 5 * time.Second

// DurableWrites makes successful writes durable before the response is sent: after
// a POST, PUT, PATCH or DELETE that answered with a success status, it flushes the
// backup. Hosting that throttles CPU between requests would otherwise pause a
// background upload, losing writes if the instance is stopped.
//
// Concurrent writers share one flush (the backup serialises and skips unchanged
// data). A failed flush is logged, not turned into an error response: the write
// itself succeeded and the periodic backup retries.
func DurableWrites(f Flusher) Middleware {
	return func(next http.Handler) http.Handler {
		if f == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}

			srw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(srw, r)
			if srw.statusCode >= http.StatusBadRequest {
				return
			}

			// A client that disconnects must not cancel the upload of its own write.
			ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), durableWriteTimeout)
			defer cancel()
			if err := f.Flush(ctx); err != nil {
				slog.Warn("backup after write failed",
					slog.String("request_id", httputil.GetRequestID(r.Context())),
					slog.Any("error", err))
			}
		})
	}
}
