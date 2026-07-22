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

	"chorus/internal/config"
	"chorus/internal/gitstore"
	chttp "chorus/internal/http"
	"chorus/internal/http/handler"
	"chorus/internal/idgen"
	"chorus/internal/identity"
	"chorus/internal/thread"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

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

	// 2. Auxiliary ID generator
	idGen := idgen.NewRandomIDGenerator()

	// 3. Services (Concrete instances with injected dependencies)
	identityService := identity.NewService(identityRepo, idGen, time.Now)
	threadService := thread.NewService(threadRepo, idGen, time.Now)

	// 4. HTTP Handlers
	healthH := handler.NewHealthHandler()
	identityH := handler.NewIdentityHandler(identityService)
	threadH := handler.NewThreadHandler(threadService)

	// 5. Router
	router := chttp.NewRouter(chttp.RouterConfig{
		Health:   healthH,
		Identity: identityH,
		Thread:   threadH,
	})

	// 6. HTTP Server Configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// 7. Graceful Shutdown listener
	shutdownErr := make(chan error, 1)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		slog.Info("shutting down server gracefully...", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownErr <- srv.Shutdown(ctx)
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
