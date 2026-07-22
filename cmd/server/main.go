package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chorus/internal/config"
	chttp "chorus/internal/http"
	"chorus/internal/http/handler"
	"chorus/internal/identity"
	"chorus/internal/repository/memory"
	"chorus/internal/thread"
)

func main() {
	cfg := config.Load()

	// 1. Repositories (In-Memory)
	identityRepo := memory.NewIdentityRepository()
	threadRepo := memory.NewThreadRepository()

	// 2. Auxiliary identity generator
	idGen := identity.NewRandomIDGenerator()

	// 3. Services
	identityService := identity.NewService(identityRepo, idGen)
	threadService := thread.NewService(threadRepo)

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

	// 7. Start Server asynchronously
	shutdownErr := make(chan error, 1)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		log.Printf("Received signal %s. Shutting down server gracefully...", s)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownErr <- srv.Shutdown(ctx)
	}()

	log.Printf("Chorus server listening on port %s (%s mode)", cfg.Port, cfg.Environment)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if err := <-shutdownErr; err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	} else {
		log.Println("Server stopped cleanly.")
	}
}
