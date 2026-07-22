package http

import (
	"net/http"

	"chorus/internal/http/handler"
	"chorus/internal/http/middleware"
)

// RouterConfig contains all handlers needed to register API routes.
type RouterConfig struct {
	Health   *handler.HealthHandler
	Identity *handler.IdentityHandler
	Thread   *handler.ThreadHandler
}

// NewRouter constructs and configures an http.Handler with all routes & global middlewares.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("GET /healthz", cfg.Health.Check)

	// Identity endpoints
	mux.HandleFunc("POST /api/v1/identities", cfg.Identity.Create)
	mux.HandleFunc("GET /api/v1/identities/{id}", cfg.Identity.GetByID)

	// Thread endpoints
	mux.HandleFunc("POST /api/v1/threads", cfg.Thread.CreateThread)
	mux.HandleFunc("GET /api/v1/threads", cfg.Thread.ListThreads)
	mux.HandleFunc("GET /api/v1/threads/{id}", cfg.Thread.GetThread)
	mux.HandleFunc("POST /api/v1/threads/{id}/messages", cfg.Thread.AddMessage)
	mux.HandleFunc("GET /api/v1/threads/{id}/messages", cfg.Thread.ListMessages)

	// Wrap mux with global middleware stack
	return middleware.Chain(
		mux,
		middleware.Recoverer,
		middleware.Logger,
	)
}
