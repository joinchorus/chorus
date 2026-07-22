package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chttp "chorus/internal/http"
	"chorus/internal/http/handler"
	"chorus/internal/idgen"
	"chorus/internal/identity"
	"chorus/internal/repository/memory"
	"chorus/internal/thread"
)

func setupTestServer() http.Handler {
	identityRepo := memory.NewIdentityRepository()
	threadRepo := memory.NewThreadRepository()
	idGen := idgen.NewRandomIDGenerator()

	identityService := identity.NewService(identityRepo, idGen, time.Now)
	threadService := thread.NewService(threadRepo, idGen, time.Now)

	return chttp.NewRouter(chttp.RouterConfig{
		Health:   handler.NewHealthHandler(),
		Identity: handler.NewIdentityHandler(identityService),
		Thread:   handler.NewThreadHandler(threadService),
	})
}

func TestHTTP_EndToEndFlow(t *testing.T) {
	router := setupTestServer()

	// 1. Check Health
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for healthz, got %d", rec.Code)
	}

	// 2. Create Thread directly with title & body & show_country
	threadBody, _ := json.Marshal(map[string]any{
		"title":        "General Chat",
		"body":         "First post content",
		"show_country": true,
	})
	req = httptest.NewRequest("POST", "/api/v1/threads", bytes.NewBuffer(threadBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for thread, got %d: %s", rec.Code, rec.Body.String())
	}

	var createdThread thread.Thread
	if err := json.NewDecoder(rec.Body).Decode(&createdThread); err != nil {
		t.Fatalf("failed decoding thread response: %v", err)
	}

	if createdThread.AuthorID == "" {
		t.Errorf("expected backend-generated author_id, got empty")
	}

	// 3. Add Message to Thread
	msgBody, _ := json.Marshal(map[string]any{
		"body":         "Hello world reply!",
		"show_country": true,
	})
	req = httptest.NewRequest("POST", "/api/v1/threads/"+createdThread.ID+"/messages", bytes.NewBuffer(msgBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for message, got %d: %s", rec.Code, rec.Body.String())
	}

	// 4. Fetch Thread Detail
	req = httptest.NewRequest("GET", "/api/v1/threads/"+createdThread.ID, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK getting thread detail, got %d: %s", rec.Code, rec.Body.String())
	}

	var detail thread.ThreadDetail
	if err := json.NewDecoder(rec.Body).Decode(&detail); err != nil {
		t.Fatalf("failed decoding thread detail response: %v", err)
	}

	if len(detail.Messages) != 2 { // 1 initial message + 1 reply
		t.Fatalf("expected 2 messages in detail, got %d", len(detail.Messages))
	}
}
