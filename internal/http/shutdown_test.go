package http_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	chttp "chorus/internal/http"
	"chorus/internal/http/handler"
)

func TestGracefulShutdown_Lifecycle(t *testing.T) {
	// Pick an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed listening on free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	router := chttp.NewRouter(chttp.RouterConfig{
		Health: handler.NewHealthHandler(),
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: router,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	// Wait briefly for server startup
	time.Sleep(50 * time.Millisecond)

	// Send test GET /healthz request
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/healthz", port))
	if err != nil {
		t.Fatalf("failed GET /healthz: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK from /healthz, got %d", resp.StatusCode)
	}

	// Trigger graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("graceful shutdown failed: %v", err)
	}

	if err := <-serverErr; err != nil && err != http.ErrServerClosed {
		t.Errorf("expected http.ErrServerClosed, got %v", err)
	}
}
