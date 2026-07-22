package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	chttp "chorus/internal/http"
	"chorus/internal/http/handler"
	"chorus/internal/identity"
	"chorus/internal/repository/memory"
	"chorus/internal/thread"
)

func setupTestServer() http.Handler {
	identityRepo := memory.NewIdentityRepository()
	threadRepo := memory.NewThreadRepository()
	idGen := identity.NewRandomIDGenerator()

	identityService := identity.NewService(identityRepo, idGen)
	threadService := thread.NewService(threadRepo)

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

	// 2. Create Identity
	identBody := []byte(`{"email":"bob@example.com","name":"Bob"}`)
	req = httptest.NewRequest("POST", "/api/v1/identities", bytes.NewBuffer(identBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for identity, got %d: %s", rec.Code, rec.Body.String())
	}

	var createdIdent identity.Identity
	if err := json.NewDecoder(rec.Body).Decode(&createdIdent); err != nil {
		t.Fatalf("failed decoding identity response: %v", err)
	}

	// 3. Create Thread
	threadBody, _ := json.Marshal(map[string]string{
		"title":     "General Chat",
		"author_id": createdIdent.ID,
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

	// 4. Add Message to Thread
	msgBody, _ := json.Marshal(map[string]string{
		"author_id": createdIdent.ID,
		"content":   "Hello world!",
	})
	req = httptest.NewRequest("POST", "/api/v1/threads/"+createdThread.ID+"/messages", bytes.NewBuffer(msgBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for message, got %d: %s", rec.Code, rec.Body.String())
	}

	// 5. Fetch Messages for Thread
	req = httptest.NewRequest("GET", "/api/v1/threads/"+createdThread.ID+"/messages", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing messages, got %d: %s", rec.Code, rec.Body.String())
	}
}
