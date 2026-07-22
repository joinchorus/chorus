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
			Title:       "Architecture Discussion",
			Body:        "Initial post body content",
			ShowCountry: true,
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if th.ID == "" {
			t.Errorf("expected non-empty thread ID")
		}
		if th.Country == nil || *th.Country != "US" {
			t.Errorf("expected country US, got %v", th.Country)
		}

		threads, err := svc.ListThreads(ctx)
		if err != nil {
			t.Fatalf("failed listing threads: %v", err)
		}
		if len(threads) != 1 {
			t.Fatalf("expected 1 thread, got %d", len(threads))
		}

		detail, err := svc.GetThreadDetail(ctx, th.ID)
		if err != nil {
			t.Fatalf("failed getting thread detail: %v", err)
		}
		if len(detail.Messages) != 1 {
			t.Fatalf("expected 1 initial message, got %d", len(detail.Messages))
		}
	})

	t.Run("add and list messages", func(t *testing.T) {
		th, err := svc.CreateThread(ctx, thread.CreateThreadInput{
			Title: "Go Best Practices",
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed creating thread: %v", err)
		}

		msg, err := svc.AddMessage(ctx, th.ID, thread.CreateMessageInput{
			Body:        "Use interfaces where they add value.",
			ShowCountry: true,
		}, "127.0.0.1")
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

	t.Run("validation failure on title length", func(t *testing.T) {
		_, err := svc.CreateThread(ctx, thread.CreateThreadInput{
			Title: "",
		}, "127.0.0.1")
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for empty title, got %v", err)
		}
	})
}
