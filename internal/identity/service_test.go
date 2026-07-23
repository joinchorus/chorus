package identity_test

import (
	"context"
	"testing"
	"time"

	"chorus/internal/conversationname"
	"chorus/internal/idgen"
	"chorus/internal/identity"
	"chorus/internal/repository/memory"
)

func TestIdentityService_CreateAndGet(t *testing.T) {
	repo := memory.NewIdentityRepository()
	gen := idgen.NewRandomIDGenerator()
	nameGen := conversationname.NewDefaultGenerator([]string{"Ash", "River"})
	fixedTime := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	mockClock := func() time.Time { return fixedTime }

	svc := identity.NewService(repo, gen, nameGen, mockClock)
	ctx := context.Background()

	t.Run("successful creation and lookup with conversation name", func(t *testing.T) {
		input := identity.CreateInput{}
		created, err := svc.Create(ctx, input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if created.ID == "" {
			t.Errorf("expected non-empty ID")
		}
		if created.ConversationName == "" {
			t.Errorf("expected non-empty ConversationName")
		}
		if !created.CreatedAt.Equal(fixedTime) {
			t.Errorf("expected time %v, got %v", fixedTime, created.CreatedAt)
		}

		fetched, err := svc.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("expected no error looking up identity, got %v", err)
		}
		if fetched.ConversationName != created.ConversationName {
			t.Errorf("expected conversation name %s, got %s", created.ConversationName, fetched.ConversationName)
		}
	})

	t.Run("custom conversation name input", func(t *testing.T) {
		input := identity.CreateInput{
			ConversationName: "Echo",
		}
		created, err := svc.Create(ctx, input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if created.ConversationName != "Echo" {
			t.Errorf("expected conversation name Echo, got %s", created.ConversationName)
		}
	})
}
