package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"chorus/internal/http/httputil"
)

// Middleware type definition for HTTP handler wrappers.
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middlewares in execution order.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Logger logs incoming HTTP requests using log/slog.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(srw, r)

		slog.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", srw.statusCode),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

// Recoverer handles panics gracefully using log/slog and returns an error response.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", slog.Any("error", err))
				httputil.WriteError(w, http.ErrAbortHandler)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
