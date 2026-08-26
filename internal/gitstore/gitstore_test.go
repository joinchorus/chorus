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

	// 2. Thread Operations with Board Scopes
	th := &thread.Thread{
		ID:        "thd_test456",
		Title:     "Git Repository Architecture",
		BoardSlug: "technology",
		AuthorID:  "usr_test123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := threadRepo.SaveThread(ctx, th); err != nil {
		t.Fatalf("failed saving thread to git: %v", err)
	}

	// Verify thread is in technology board index
	techThreads, err := threadRepo.ListThreads(ctx, "technology")
	if err != nil {
		t.Fatalf("failed listing technology threads: %v", err)
	}
	if len(techThreads) != 1 {
		t.Fatalf("expected 1 technology thread, got %d", len(techThreads))
	}
	if techThreads[0].ID != "thd_test456" {
		t.Errorf("expected thread ID thd_test456, got %s", techThreads[0].ID)
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
		Content:   "Second append-only message (abusive content to be removed)",
		CreatedAt: now.Add(time.Second),
	}

	if err := threadRepo.SaveMessage(ctx, msg1); err != nil {
		t.Fatalf("failed appending msg1: %v", err)
	}
	if err := threadRepo.SaveMessage(ctx, msg2); err != nil {
		t.Fatalf("failed appending msg2: %v", err)
	}

	// 4. Verify Message Listing Before Moderation
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

	// 5. Test Moderation Redaction
	modDir := filepath.Join(tempDir, "boards", "technology", "threads", "thd_test456", "moderation")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatalf("failed creating moderation dir: %v", err)
	}
	actionJSON := `{"id":"mod_123","report_id":"rpt_1","thread_id":"thd_test456","message_id":"msg_222","status":"removed","note":"Violated guidelines"}`
	if err := os.WriteFile(filepath.Join(modDir, "mod_123.json"), []byte(actionJSON), 0644); err != nil {
		t.Fatalf("failed writing moderation action: %v", err)
	}

	// Public listing should now redact msg_222
	moderatedMsgs, err := threadRepo.ListMessagesByThreadID(ctx, "thd_test456")
	if err != nil {
		t.Fatalf("failed listing messages after moderation: %v", err)
	}
	if len(moderatedMsgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(moderatedMsgs))
	}
	if moderatedMsgs[1].Content != "[This message was removed by moderation]" {
		t.Errorf("expected redacted content for removed message, got %s", moderatedMsgs[1].Content)
	}
	if !moderatedMsgs[1].IsRemoved {
		t.Errorf("expected is_removed = true on removed message")
	}
	if moderatedMsgs[0].IsRemoved {
		t.Errorf("expected is_removed = false on regular message")
	}

	// 6. Test Participant Continuity and Credential Hashing
	p := &thread.Participant{
		Token:            "ptk_secret123",
		ConversationName: "Falcon",
		AuthorID:         "usr_test123",
		CreatedAt:        now,
	}
	if err := threadRepo.SaveParticipant(ctx, "thd_test456", p); err != nil {
		t.Fatalf("failed saving participant: %v", err)
	}
	fetchedPart, err := threadRepo.FindParticipantByToken(ctx, "thd_test456", "ptk_secret123")
	if err != nil {
		t.Fatalf("failed finding participant by token: %v", err)
	}
	if fetchedPart.ConversationName != "Falcon" {
		t.Errorf("expected Falcon, got %s", fetchedPart.ConversationName)
	}

	// Verify raw bearer token is NOT stored in participants.json on disk
	partBytes, err := os.ReadFile(filepath.Join(tempDir, "boards", "technology", "threads", "thd_test456", "participants.json"))
	if err != nil {
		t.Fatalf("failed reading participants.json: %v", err)
	}
	if strings.Contains(string(partBytes), "ptk_secret123") {
		t.Errorf("SECURITY FLAW: Raw bearer token found inside participants.json: %s", string(partBytes))
	}
	if !strings.Contains(string(partBytes), "token_hash") {
		t.Errorf("expected token_hash inside participants.json: %s", string(partBytes))
	}

	// Verify thread.json and messages.ndjson do NOT contain participant_token
	threadBytes, err := os.ReadFile(filepath.Join(tempDir, "boards", "technology", "threads", "thd_test456", "thread.json"))
	if err != nil {
		t.Fatalf("failed reading thread.json: %v", err)
	}
	if strings.Contains(string(threadBytes), "participant_token") {
		t.Errorf("SECURITY FLAW: participant_token found inside thread.json: %s", string(threadBytes))
	}

	msgBytes, err := os.ReadFile(filepath.Join(tempDir, "boards", "technology", "threads", "thd_test456", "messages.ndjson"))
	if err != nil {
		t.Fatalf("failed reading messages.ndjson: %v", err)
	}
	if strings.Contains(string(msgBytes), "participant_token") {
		t.Errorf("SECURITY FLAW: participant_token found inside messages.ndjson: %s", string(msgBytes))
	}

	// 7. Verify Git Commit History
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

	// 8. Verify Filesystem Structure & Index
	indexPath := filepath.Join(tempDir, "boards", "technology", "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("expected index.json at %s", indexPath)
	}

	// 9. Verify Recovery after Index deletion
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
