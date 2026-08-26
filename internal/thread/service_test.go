package thread_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"chorus/internal/conversationname"
	"chorus/internal/domain"
	"chorus/internal/idgen"
	"chorus/internal/repository/memory"
	"chorus/internal/thread"
)

func TestThreadService_Operations(t *testing.T) {
	repo := memory.NewThreadRepository()
	gen := idgen.NewRandomIDGenerator()
	nameGen := conversationname.NewDefaultGenerator([]string{"Ash", "River", "Echo", "Stone"})
	fixedTime := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	mockClock := func() time.Time { return fixedTime }

	svc := thread.NewService(repo, gen, nameGen, mockClock)
	ctx := context.Background()

	t.Run("create thread and list with conversation name", func(t *testing.T) {
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
		if th.ConversationName == "" {
			t.Errorf("expected non-empty ConversationName")
		}
		if th.Country == nil || *th.Country != "TR" {
			t.Errorf("expected country TR, got %v", th.Country)
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
		if detail.Messages[0].ConversationName != th.ConversationName {
			t.Errorf("expected initial message conversation name %s, got %s", th.ConversationName, detail.Messages[0].ConversationName)
		}
	})

	t.Run("add message generates unique conversation name in thread", func(t *testing.T) {
		th, err := svc.CreateThread(ctx, thread.CreateThreadInput{
			Title:            "Go Best Practices",
			ConversationName: "Ash",
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed creating thread: %v", err)
		}

		// Reply attempting to use existing "Ash" without token -> Should assign a unique name
		msg1, err := svc.AddMessage(ctx, th.ID, thread.CreateMessageInput{
			Body:             "Use interfaces where they add value.",
			ConversationName: "Ash",
			ShowCountry:      true,
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed adding message: %v", err)
		}
		if msg1.ConversationName == "Ash" {
			t.Errorf("expected new unique conversation name without token, got Ash")
		}

		msgs, err := svc.ListMessages(ctx, th.ID)
		if err != nil {
			t.Fatalf("failed listing messages: %v", err)
		}
		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}
	})

	t.Run("thread-scoped participant continuity with token", func(t *testing.T) {
		// Creator creates thread
		th, err := svc.CreateThread(ctx, thread.CreateThreadInput{
			Title:            "Continuity Thread",
			BoardSlug:        "programming",
			ConversationName: "River",
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed creating thread: %v", err)
		}
		if th.ParticipantToken == "" {
			t.Fatalf("expected non-empty ParticipantToken on created thread")
		}
		if th.BoardSlug != "programming" {
			t.Errorf("expected board_slug programming, got %s", th.BoardSlug)
		}

		// Creator replies in same thread using participant token -> should preserve "River"
		reply1, err := svc.AddMessage(ctx, th.ID, thread.CreateMessageInput{
			Body:             "First follow-up by OP",
			ParticipantToken: th.ParticipantToken,
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed adding reply: %v", err)
		}
		if reply1.ConversationName != "River" {
			t.Errorf("expected conversation name River preserved with token, got %s", reply1.ConversationName)
		}
		if reply1.AuthorID != th.AuthorID {
			t.Errorf("expected author ID %s, got %s", th.AuthorID, reply1.AuthorID)
		}

		// A new participant joins without token -> gets a different name (e.g. Ash)
		p2Reply, err := svc.AddMessage(ctx, th.ID, thread.CreateMessageInput{
			Body: "Hello from participant 2",
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed adding participant 2 message: %v", err)
		}
		if p2Reply.ConversationName == "River" {
			t.Errorf("expected distinct conversation name for new participant, got River")
		}
		if p2Reply.ParticipantToken == "" {
			t.Fatalf("expected non-empty ParticipantToken for new participant")
		}

		// Participant 2 replies again with their own token -> preserves their name
		p2Followup, err := svc.AddMessage(ctx, th.ID, thread.CreateMessageInput{
			Body:             "Second reply by participant 2",
			ParticipantToken: p2Reply.ParticipantToken,
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed adding participant 2 follow-up: %v", err)
		}
		if p2Followup.ConversationName != p2Reply.ConversationName {
			t.Errorf("expected participant 2 name %s, got %s", p2Reply.ConversationName, p2Followup.ConversationName)
		}

		// Security Test 1: Public read paths NEVER contain participant_token
		readTh, err := svc.GetThreadByID(ctx, th.ID)
		if err != nil {
			t.Fatalf("failed reading thread: %v", err)
		}
		if readTh.ParticipantToken != "" {
			t.Errorf("SECURITY LEAK: GetThreadByID returned non-empty ParticipantToken: %s", readTh.ParticipantToken)
		}

		detail, err := svc.GetThreadDetail(ctx, th.ID)
		if err != nil {
			t.Fatalf("failed getting thread detail: %v", err)
		}
		if detail.Thread.ParticipantToken != "" {
			t.Errorf("SECURITY LEAK: GetThreadDetail returned non-empty ParticipantToken on thread: %s", detail.Thread.ParticipantToken)
		}
		for _, m := range detail.Messages {
			if m.ParticipantToken != "" {
				t.Errorf("SECURITY LEAK: GetThreadDetail returned non-empty ParticipantToken on message %s: %s", m.ID, m.ParticipantToken)
			}
		}

		allThreads, err := svc.ListThreads(ctx)
		if err != nil {
			t.Fatalf("failed listing threads: %v", err)
		}
		for _, tItem := range allThreads {
			if tItem.ParticipantToken != "" {
				t.Errorf("SECURITY LEAK: ListThreads returned non-empty ParticipantToken on thread %s: %s", tItem.ID, tItem.ParticipantToken)
			}
		}

		// Security Test 2: Token from Thread A cannot authenticate in Thread B
		thB, err := svc.CreateThread(ctx, thread.CreateThreadInput{
			Title: "Thread B For Isolation Test",
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed creating Thread B: %v", err)
		}

		// Attempt to use Thread A's token in Thread B
		crossReply, err := svc.AddMessage(ctx, thB.ID, thread.CreateMessageInput{
			Body:             "Attempting cross-thread token reuse",
			ParticipantToken: th.ParticipantToken, // Token from Thread A
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed adding message: %v", err)
		}
		// Must NOT get OP's Thread A conversation name ("River")
		if crossReply.ConversationName == "River" {
			t.Errorf("SECURITY FLAW: Cross-thread token allowed identity reuse across different threads")
		}
		if crossReply.AuthorID == th.AuthorID {
			t.Errorf("SECURITY FLAW: Cross-thread token allowed author ID reuse across different threads")
		}
	})

	t.Run("board validation and filtering", func(t *testing.T) {
		// Valid board
		th, err := svc.CreateThread(ctx, thread.CreateThreadInput{
			Title:     "Philosophy of Tech",
			BoardSlug: "philosophy",
		}, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed creating thread: %v", err)
		}
		if th.BoardSlug != "philosophy" || th.Topic != "Philosophy" {
			t.Errorf("expected philosophy board metadata, got %s / %s", th.BoardSlug, th.Topic)
		}

		// Filter by board
		philThreads, err := svc.ListThreads(ctx, "philosophy")
		if err != nil {
			t.Fatalf("failed listing philosophy threads: %v", err)
		}
		if len(philThreads) != 1 {
			t.Fatalf("expected 1 philosophy thread, got %d", len(philThreads))
		}

		progThreads, err := svc.ListThreads(ctx, "programming")
		if err != nil {
			t.Fatalf("failed listing programming threads: %v", err)
		}
		if len(progThreads) != 1 {
			t.Fatalf("expected 1 programming thread, got %d", len(progThreads))
		}

		// Invalid board slug returns ErrValidation
		_, err = svc.CreateThread(ctx, thread.CreateThreadInput{
			Title:     "Invalid Board Thread",
			BoardSlug: "nonexistent_secret_board",
		}, "127.0.0.1")
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for invalid board slug, got %v", err)
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
