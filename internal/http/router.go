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
	Health         *handler.HealthHandler
	Identity       *handler.IdentityHandler
	Board          *handler.BoardHandler
	Thread         *handler.ThreadHandler
	Translation    *handler.TranslationHandler
	Report         *handler.ReportHandler
	Moderation     *handler.ModerationHandler
	SessionManager *middleware.SessionManager
	AdminToken     string
	StaticDir      string
}

// NewRouter constructs and configures an http.Handler with all routes, SPA fallback, & global middlewares.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	if cfg.Board == nil {
		cfg.Board = handler.NewBoardHandler()
	}
	if cfg.SessionManager == nil {
		cfg.SessionManager = middleware.NewSessionManager()
	}

	// Health endpoint
	mux.HandleFunc("GET /healthz", cfg.Health.Check)

	// Rate limiters for public write & cost-sensitive endpoints (Token Bucket IP Rate Limiting)
	threadLimiter := middleware.NewIPRateLimiter(0.1, 5)   // ~6 threads/min burst 5
	msgLimiter := middleware.NewIPRateLimiter(0.5, 10)     // ~30 msgs/min burst 10
	reportLimiter := middleware.NewIPRateLimiter(0.2, 5)   // ~12 reports/min burst 5
	identityLimiter := middleware.NewIPRateLimiter(0.2, 5) // ~12 identities/min burst 5 (abuse prevention against disk exhaustion)
	transLimiter := middleware.NewIPRateLimiter(0.2, 5)    // ~12 translations/min burst 5 (cost & quota protection)

	threadRateGuard := middleware.RateLimit(threadLimiter)
	msgRateGuard := middleware.RateLimit(msgLimiter)
	reportRateGuard := middleware.RateLimit(reportLimiter)
	identityRateGuard := middleware.RateLimit(identityLimiter)
	transRateGuard := middleware.RateLimit(transLimiter)

	// Helper to register API routes under a version prefix (/api/v0.1 and /api/v1)
	registerRoutes := func(prefix string) {
		mux.Handle("POST "+prefix+"/identities", identityRateGuard(http.HandlerFunc(cfg.Identity.Create)))
		mux.HandleFunc("GET "+prefix+"/identities/{id}", cfg.Identity.GetByID)

		mux.HandleFunc("GET "+prefix+"/boards", cfg.Board.ListBoards)
		mux.HandleFunc("GET "+prefix+"/boards/{slug}", cfg.Board.GetBoard)

		mux.Handle("POST "+prefix+"/threads", threadRateGuard(http.HandlerFunc(cfg.Thread.CreateThread)))
		mux.HandleFunc("GET "+prefix+"/threads", cfg.Thread.ListThreads)
		mux.HandleFunc("GET "+prefix+"/threads/{id}", cfg.Thread.GetThread)
		mux.Handle("POST "+prefix+"/threads/{id}/messages", msgRateGuard(http.HandlerFunc(cfg.Thread.AddMessage)))
		mux.HandleFunc("GET "+prefix+"/threads/{id}/messages", cfg.Thread.ListMessages)

		if cfg.Translation != nil {
			mux.Handle("POST "+prefix+"/threads/{id}/messages/{msg_id}/translate", transRateGuard(http.HandlerFunc(cfg.Translation.TranslateMessage)))
		}

		if cfg.Report != nil {
			mux.Handle("POST "+prefix+"/threads/{id}/messages/{msg_id}/report", reportRateGuard(http.HandlerFunc(cfg.Report.CreateReport)))
		}

		if cfg.Moderation != nil {
			authGuard := middleware.RequireAdminAuth(cfg.AdminToken, cfg.SessionManager)

			// Auth endpoints for HttpOnly Cookie session management
			mux.HandleFunc("POST "+prefix+"/moderation/login", cfg.Moderation.Login)
			mux.HandleFunc("POST "+prefix+"/moderation/logout", cfg.Moderation.Logout)
			mux.Handle("GET "+prefix+"/moderation/session", authGuard(http.HandlerFunc(cfg.Moderation.GetSession)))

			// Protected Moderation Queue Endpoints
			mux.Handle("GET "+prefix+"/moderation/reports", authGuard(http.HandlerFunc(cfg.Moderation.ListQueue)))
			mux.Handle("GET "+prefix+"/moderation/reports/{id}", authGuard(http.HandlerFunc(cfg.Moderation.GetReportDetail)))
			mux.Handle("POST "+prefix+"/moderation/reports/{id}/action", authGuard(http.HandlerFunc(cfg.Moderation.SubmitAction)))
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
			relSubPath := strings.TrimPrefix(r.URL.Path, "/landing")
			if relSubPath == "" || relSubPath == "/" {
				relSubPath = "index.html"
			}
			landingAsset := findFile(filepath.Join("public", "landing", filepath.Clean(relSubPath)))
			if info, err := os.Stat(landingAsset); err == nil && !info.IsDir() {
				serveFile(landingAsset)
				return
			}
			landingPath := findFile(filepath.Join("public", "landing", "index.html"))
			if _, err := os.Stat(landingPath); err == nil {
				serveFile(landingPath)
				return
			}
		}

		if host == "docs.joinchorus.app" || strings.HasPrefix(r.URL.Path, "/docs") {
			relSubPath := strings.TrimPrefix(r.URL.Path, "/docs")
			if relSubPath == "" || relSubPath == "/" {
				relSubPath = "index.html"
			}
			docsAsset := findFile(filepath.Join("public", "docs", filepath.Clean(relSubPath)))
			if info, err := os.Stat(docsAsset); err == nil && !info.IsDir() {
				serveFile(docsAsset)
				return
			}
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

		// SPA Fallback: serve index.html for page navigation routes (exclude asset extensions like .css, .js, .png)
		ext := filepath.Ext(r.URL.Path)
		if ext != "" && ext != ".html" {
			http.NotFound(w, r)
			return
		}

		indexPath := findFile(filepath.Join(cfg.StaticDir, "index.html"))
		serveFile(indexPath)
	})

	// Wrap in global middleware stack (outermost to innermost: RequestID -> Logger -> Recoverer -> CORS -> SecurityHeaders -> Mux)
	var handler http.Handler = mux
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.CORS(handler)
	handler = middleware.Recoverer(handler)
	handler = middleware.Logger(handler)
	handler = middleware.RequestID(handler)

	return handler
}
