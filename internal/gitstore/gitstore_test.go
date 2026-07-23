package gitstore_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"chorus/internal/gitstore"
	"chorus/internal/identity"
	"chorus/internal/thread"
)

func TestGitStore_FullFlow(t *testing.T) {
	tempDir := t.TempDir()
	store := gitstore.NewGitStore(tempDir)

	identRepo, err := gitstore.NewIdentityRepository(store)
	if err != nil {
		t.Fatalf("failed initializing git identity repo: %v", err)
	}

	threadRepo, err := gitstore.NewThreadRepository(store)
	if err != nil {
		t.Fatalf("failed initializing git thread repo: %v", err)
	}

	ctx := context.Background()
	now := time.Now().UTC()

	// 1. Identity Operations
	ident := &identity.Identity{
		ID:               "usr_test123",
		ConversationName: "Ash",
		CreatedAt:        now,
	}

	if err := identRepo.Save(ctx, ident); err != nil {
		t.Fatalf("failed saving identity to git: %v", err)
	}

	fetchedIdent, err := identRepo.FindByID(ctx, "usr_test123")
	if err != nil {
		t.Fatalf("failed finding identity by ID: %v", err)
	}
	if fetchedIdent.ConversationName != "Ash" {
		t.Errorf("expected Ash, got %s", fetchedIdent.ConversationName)
	}

	// 2. Thread Operations
	th := &thread.Thread{
		ID:        "thd_test456",
		Title:     "Git Repository Architecture",
		AuthorID:  "usr_test123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := threadRepo.SaveThread(ctx, th); err != nil {
		t.Fatalf("failed saving thread to git: %v", err)
	}

	// 3. Message Append Operations
	msg1 := &thread.Message{
		ID:        "msg_111",
		ThreadID:  "thd_test456",
		AuthorID:  "usr_test123",
		Content:   "First append-only message",
		CreatedAt: now,
	}
	msg2 := &thread.Message{
		ID:        "msg_222",
		ThreadID:  "thd_test456",
		AuthorID:  "usr_test123",
		Content:   "Second append-only message",
		CreatedAt: now,
	}

	if err := threadRepo.SaveMessage(ctx, msg1); err != nil {
		t.Fatalf("failed appending msg1: %v", err)
	}
	if err := threadRepo.SaveMessage(ctx, msg2); err != nil {
		t.Fatalf("failed appending msg2: %v", err)
	}

	// 4. Verify Message Listing
	msgs, err := threadRepo.ListMessagesByThreadID(ctx, "thd_test456")
	if err != nil {
		t.Fatalf("failed listing messages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Content != "First append-only message" {
		t.Errorf("expected first message content, got %s", msgs[0].Content)
	}

	// 5. Verify Git Commit History
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = tempDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed running git log: %v", err)
	}

	logStr := string(out)
	t.Logf("Git Commit History:\n%s", logStr)

	if !strings.Contains(logStr, "identity: create usr_test123") {
		t.Errorf("missing identity commit in git log")
	}
	if !strings.Contains(logStr, "thread: create thd_test456") {
		t.Errorf("missing thread commit in git log")
	}
	if !strings.Contains(logStr, "message: append msg_111") {
		t.Errorf("missing message commit in git log")
	}

	// 6. Verify Filesystem Structure & Index
	indexPath := filepath.Join(tempDir, "boards", "general", "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("expected index.json at %s", indexPath)
	}

	// 7. Verify Recovery after Index deletion
	if err := os.Remove(indexPath); err != nil {
		t.Fatalf("failed removing index.json for recovery test: %v", err)
	}

	if err := store.RecoverAndVerify(ctx); err != nil {
		t.Fatalf("failed running RecoverAndVerify: %v", err)
	}

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("expected index.json to be rebuilt during recovery")
	}
}
