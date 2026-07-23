package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chorus/internal/http/handler"
	"chorus/internal/http/middleware"
)

// RouterConfig contains all handlers needed to register API routes and static SPA asset directory.
type RouterConfig struct {
	Health      *handler.HealthHandler
	Identity    *handler.IdentityHandler
	Thread      *handler.ThreadHandler
	Translation *handler.TranslationHandler
	Report      *handler.ReportHandler
	Moderation  *handler.ModerationHandler
	StaticDir   string
}

// NewRouter constructs and configures an http.Handler with all routes, SPA fallback, & global middlewares.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("GET /healthz", cfg.Health.Check)

	// Helper to register API routes under a version prefix (/api/v0.1 and /api/v1)
	registerRoutes := func(prefix string) {
		mux.HandleFunc("POST "+prefix+"/identities", cfg.Identity.Create)
		mux.HandleFunc("GET "+prefix+"/identities/{id}", cfg.Identity.GetByID)

		mux.HandleFunc("POST "+prefix+"/threads", cfg.Thread.CreateThread)
		mux.HandleFunc("GET "+prefix+"/threads", cfg.Thread.ListThreads)
		mux.HandleFunc("GET "+prefix+"/threads/{id}", cfg.Thread.GetThread)
		mux.HandleFunc("POST "+prefix+"/threads/{id}/messages", cfg.Thread.AddMessage)
		mux.HandleFunc("GET "+prefix+"/threads/{id}/messages", cfg.Thread.ListMessages)

		if cfg.Translation != nil {
			mux.HandleFunc("POST "+prefix+"/threads/{id}/messages/{msg_id}/translate", cfg.Translation.TranslateMessage)
		}

		if cfg.Report != nil {
			mux.HandleFunc("POST "+prefix+"/threads/{id}/messages/{msg_id}/report", cfg.Report.CreateReport)
		}

		if cfg.Moderation != nil {
			mux.HandleFunc("GET "+prefix+"/moderation/reports", cfg.Moderation.ListQueue)
			mux.HandleFunc("GET "+prefix+"/moderation/reports/{id}", cfg.Moderation.GetReportDetail)
			mux.HandleFunc("POST "+prefix+"/moderation/reports/{id}/action", cfg.Moderation.SubmitAction)
		}
	}

	// Register current Alpha version /api/v0.1 and backward-compatible /api/v1
	registerRoutes("/api/v0.1")
	registerRoutes("/api/v1")

	// Static SPA Handler with index.html fallback for client-side routing
	if cfg.StaticDir == "" {
		cfg.StaticDir = "web/dist"
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Do not intercept API or health endpoints
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" {
			http.NotFound(w, r)
			return
		}

		// Host-based routing for subdomains (joinchorus.app, docs.joinchorus.app, chat.joinchorus.app)
		host := strings.ToLower(r.Host)
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		findFile := func(relPath string) string {
			if _, err := os.Stat(relPath); err == nil {
				return relPath
			}
			// Fallback for test runners executing in subpackages
			parentPath := filepath.Join("..", "..", "..", relPath)
			if _, err := os.Stat(parentPath); err == nil {
				return parentPath
			}
			return relPath
		}

		if host == "joinchorus.app" || host == "www.joinchorus.app" || strings.HasPrefix(r.URL.Path, "/landing") {
			landingPath := findFile(filepath.Join("public", "landing", "index.html"))
			if _, err := os.Stat(landingPath); err == nil {
				http.ServeFile(w, r, landingPath)
				return
			}
		}

		if host == "docs.joinchorus.app" || strings.HasPrefix(r.URL.Path, "/docs") {
			docsPath := findFile(filepath.Join("public", "docs", "index.html"))
			if _, err := os.Stat(docsPath); err == nil {
				http.ServeFile(w, r, docsPath)
				return
			}
		}

		cleanPath := filepath.Clean(r.URL.Path)
		targetPath := filepath.Join(cfg.StaticDir, cleanPath)

		// Check if exact file exists
		info, err := os.Stat(targetPath)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, targetPath)
			return
		}

		// Fallback to index.html for SPA client-side routes
		indexPath := filepath.Join(cfg.StaticDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}

		// If no static build exists, return 200 with fallback info
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<!DOCTYPE html><html><body><h1>Chorus Application</h1><p>Running chat.joinchorus.app backend API server.</p></body></html>"))
	})

	// Wrap mux with global middleware stack
	return middleware.Chain(
		mux,
		middleware.Recoverer,
		middleware.Logger,
	)
}
