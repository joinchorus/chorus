package thread_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"chorus/internal/domain"
	"chorus/internal/idgen"
	"chorus/internal/repository/memory"
	"chorus/internal/thread"
)

func TestThreadService_Operations(t *testing.T) {
	repo := memory.NewThreadRepository()
	gen := idgen.NewRandomIDGenerator()
	fixedTime := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	mockClock := func() time.Time { return fixedTime }

	svc := thread.NewService(repo, gen, mockClock)
	ctx := context.Background()

	t.Run("create thread and list", func(t *testing.T) {
		th, err := svc.CreateThread(ctx, thread.CreateThreadInput{
			Title:    "Architecture Discussion",
			AuthorID: "usr_123",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if th.ID == "" {
			t.Errorf("expected non-empty thread ID")
		}
		if !th.CreatedAt.Equal(fixedTime) {
			t.Errorf("expected timestamp %v, got %v", fixedTime, th.CreatedAt)
		}

		threads, err := svc.ListThreads(ctx)
		if err != nil {
			t.Fatalf("failed listing threads: %v", err)
		}
		if len(threads) != 1 {
			t.Fatalf("expected 1 thread, got %d", len(threads))
		}
	})

	t.Run("add and list messages", func(t *testing.T) {
		th, err := svc.CreateThread(ctx, thread.CreateThreadInput{
			Title:    "Go Best Practices",
			AuthorID: "usr_456",
		})
		if err != nil {
			t.Fatalf("failed creating thread: %v", err)
		}

		msg, err := svc.AddMessage(ctx, th.ID, thread.CreateMessageInput{
			AuthorID: "usr_456",
			Content:  "Use interfaces where they add value.",
		})
		if err != nil {
			t.Fatalf("failed adding message: %v", err)
		}
		if msg.ID == "" {
			t.Errorf("expected non-empty message ID")
		}

		msgs, err := svc.ListMessages(ctx, th.ID)
		if err != nil {
			t.Fatalf("failed listing messages: %v", err)
		}
		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}
	})

	t.Run("message on non-existent thread fails", func(t *testing.T) {
		_, err := svc.AddMessage(ctx, "thd_nonexistent", thread.CreateMessageInput{
			AuthorID: "usr_123",
			Content:  "Hello?",
		})
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
