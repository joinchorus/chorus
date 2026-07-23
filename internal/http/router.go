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
	Board       *handler.BoardHandler
	Thread      *handler.ThreadHandler
	Translation *handler.TranslationHandler
	Report      *handler.ReportHandler
	Moderation  *handler.ModerationHandler
	StaticDir   string
}

// NewRouter constructs and configures an http.Handler with all routes, SPA fallback, & global middlewares.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	if cfg.Board == nil {
		cfg.Board = handler.NewBoardHandler()
	}

	// Health endpoint
	mux.HandleFunc("GET /healthz", cfg.Health.Check)

	// Helper to register API routes under a version prefix (/api/v0.1 and /api/v1)
	registerRoutes := func(prefix string) {
		mux.HandleFunc("POST "+prefix+"/identities", cfg.Identity.Create)
		mux.HandleFunc("GET "+prefix+"/identities/{id}", cfg.Identity.GetByID)

		mux.HandleFunc("GET "+prefix+"/boards", cfg.Board.ListBoards)
		mux.HandleFunc("GET "+prefix+"/boards/{slug}", cfg.Board.GetBoard)

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

		serveFile := func(targetPath string) {
			f, err := os.Open(targetPath)
			if err != nil {
				if strings.HasSuffix(targetPath, "index.html") {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("<!DOCTYPE html><html><head><title>Chorus</title></head><body><div id=\"root\"></div></body></html>"))
					return
				}
				http.NotFound(w, r)
				return
			}
			defer f.Close()
			info, err := f.Stat()
			if err != nil || info.IsDir() {
				if strings.HasSuffix(targetPath, "index.html") {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("<!DOCTYPE html><html><head><title>Chorus</title></head><body><div id=\"root\"></div></body></html>"))
					return
				}
				http.NotFound(w, r)
				return
			}
			http.ServeContent(w, r, info.Name(), info.ModTime(), f)
		}

		if host == "joinchorus.app" || host == "www.joinchorus.app" || strings.HasPrefix(r.URL.Path, "/landing") {
			landingPath := findFile(filepath.Join("public", "landing", "index.html"))
			if _, err := os.Stat(landingPath); err == nil {
				serveFile(landingPath)
				return
			}
		}

		if host == "docs.joinchorus.app" || strings.HasPrefix(r.URL.Path, "/docs") {
			docsPath := findFile(filepath.Join("public", "docs", "index.html"))
			if _, err := os.Stat(docsPath); err == nil {
				serveFile(docsPath)
				return
			}
		}

		// Main Web App (chat.joinchorus.app or localhost)
		staticPath := findFile(filepath.Join(cfg.StaticDir, filepath.Clean(r.URL.Path)))
		if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
			serveFile(staticPath)
			return
		}

		// SPA Fallback: serve index.html for unknown routes
		indexPath := findFile(filepath.Join(cfg.StaticDir, "index.html"))
		serveFile(indexPath)
	})

	// Wrap in global CORS & Logger Middleware
	handler := middleware.CORS(mux)
	handler = middleware.Logger(handler)

	return handler
}
