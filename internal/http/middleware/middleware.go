package middleware

import (
	"log"
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

// Logger logs incoming HTTP request method, URI, status, and processing duration.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(srw, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, srw.statusCode, time.Since(start))
	})
}

// Recoverer handles panics gracefully by returning an internal server error response.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERY] %v", err)
				httputil.WriteError(w, http.ErrAbortHandler)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
