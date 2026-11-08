package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chorus/internal/backup"
	"chorus/internal/config"
	"chorus/internal/conversationname"
	"chorus/internal/geoip"
	"chorus/internal/gitstore"
	chttp "chorus/internal/http"
	"chorus/internal/http/handler"
	"chorus/internal/http/middleware"
	"chorus/internal/identity"
	"chorus/internal/idgen"
	"chorus/internal/moderation"
	"chorus/internal/reporting"
	"chorus/internal/thread"
	"chorus/internal/translation"
)

// commit is the source revision, set at build time by scripts/build-backend.sh
// (-ldflags "-X main.commit=...").
var commit = "unknown"

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("refusing to start with insecure configuration", slog.Any("error", err))
		os.Exit(1)
	}

	// 0. Restore the data directory from its Cloud Storage snapshot (ephemeral-disk
	// hosting). Refusing to start on a restore error is deliberate: serving an
	// empty directory would let the next backup replace the stored snapshot.
	dataBackup := backup.New(backup.Config{Bucket: cfg.BackupBucket, Object: cfg.BackupObject, DataDir: cfg.DataDir})
	restoreCtx, cancelRestore := context.WithTimeout(context.Background(), 60*time.Second)
	restored, err := dataBackup.Restore(restoreCtx)
	cancelRestore()
	if err != nil {
		slog.Error("restoring the data backup failed; not starting", slog.Any("error", err))
		os.Exit(1)
	}
	if dataBackup != nil {
		slog.Info("data backup enabled", slog.String("bucket", cfg.BackupBucket), slog.Bool("restored", restored))
	}

	// 1. Repositories (Git-Backed Persistence)
	gitStore := gitstore.NewGitStore(cfg.DataDir)
	identityRepo, err := gitstore.NewIdentityRepository(gitStore)
	if err != nil {
		slog.Error("failed initializing git identity repository", slog.Any("error", err))
		os.Exit(1)
	}

	threadRepo, err := gitstore.NewThreadRepository(gitStore)
	if err != nil {
		slog.Error("failed initializing git thread repository", slog.Any("error", err))
		os.Exit(1)
	}

	// 2. Auxiliary Services (ID generator, Name generator & GeoIP)
	idGen := idgen.NewRandomIDGenerator()
	nameGen := conversationname.NewDefaultGenerator(nil)
	// Private/loopback peers only occur in development; in production they must
	// not be reported as a real country.
	devCountry := "US"
	if cfg.Environment == "production" {
		devCountry = "UN"
	}
	geoService := geoip.NewService(devCountry, cfg.EdgeSecret)

	// 3. Services (Concrete instances with injected dependencies)
	identityService := identity.NewService(identityRepo, idGen, nameGen, time.Now)
	threadService := thread.NewService(threadRepo, idGen, nameGen, time.Now)

	// Translation service
	googleApiKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	googleProvider := translation.NewGoogleProvider(googleApiKey, nil)
	transService := translation.NewService(googleProvider, cfg.DataDir)

	// Reporting & Moderation services & Session Manager
	reportingService := reporting.NewService(gitStore, idGen, time.Now)
	moderationService := moderation.NewService(gitStore, threadService, idGen, time.Now)
	sessionManager := middleware.NewSessionManager()

	// 4. HTTP Handlers
	healthH := handler.NewHealthHandler().WithCommit(commit)
	identityH := handler.NewIdentityHandler(identityService)
	threadH := handler.NewThreadHandler(threadService, geoService)
	transH := handler.NewTranslationHandler(transService, threadService)
	reportH := handler.NewReportHandler(reportingService, threadService)
	modH := handler.NewModerationHandler(moderationService, sessionManager, cfg.AdminToken)

	// 5. Router
	var flusher middleware.Flusher
	if dataBackup != nil {
		flusher = dataBackup
	}
	router := chttp.NewRouter(chttp.RouterConfig{
		Flusher:        flusher,
		Health:         healthH,
		Identity:       identityH,
		Thread:         threadH,
		Translation:    transH,
		Report:         reportH,
		Moderation:     modH,
		SessionManager: sessionManager,
		AdminToken:     cfg.AdminToken,
		StaticDir:      "web/dist",
	})

	// 6. HTTP Server Configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	backupCtx, stopBackup := context.WithCancel(context.Background())
	defer stopBackup()
	go dataBackup.Run(backupCtx)

	// 7. Graceful Shutdown listener
	shutdownErr := make(chan error, 1)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		slog.Info("shutting down server gracefully...", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		stopBackup()
		// Last chance to persist writes that arrived since the previous flush.
		if flushErr := dataBackup.Flush(ctx); flushErr != nil {
			slog.Error("final backup failed", slog.Any("error", flushErr))
		}
		shutdownErr <- err
	}()

	slog.Info("server starting", slog.String("port", cfg.Port), slog.String("env", cfg.Environment), slog.String("data_dir", cfg.DataDir))
	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server forced to shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	if err := <-shutdownErr; err != nil {
		slog.Error("graceful shutdown failed", slog.Any("error", err))
	} else {
		slog.Info("server stopped cleanly")
	}
}
