package identity_test

import (
	"context"
	"errors"
	"testing"

	"chorus/internal/domain"
	"chorus/internal/identity"
	"chorus/internal/repository/memory"
)

func TestIdentityService_CreateAndGet(t *testing.T) {
	repo := memory.NewIdentityRepository()
	gen := identity.NewRandomIDGenerator()
	svc := identity.NewService(repo, gen)
	ctx := context.Background()

	t.Run("successful creation and lookup", func(t *testing.T) {
		input := identity.CreateInput{
			Email: "alice@example.com",
			Name:  "Alice",
		}
		created, err := svc.Create(ctx, input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if created.ID == "" {
			t.Errorf("expected non-empty ID")
		}
		if created.Email != "alice@example.com" {
			t.Errorf("expected email alice@example.com, got %s", created.Email)
		}

		fetched, err := svc.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("expected no error looking up identity, got %v", err)
		}
		if fetched.Name != "Alice" {
			t.Errorf("expected name Alice, got %s", fetched.Name)
		}
	})

	t.Run("duplicate email creation fails", func(t *testing.T) {
		input := identity.CreateInput{
			Email: "alice@example.com",
			Name:  "Alice Copy",
		}
		_, err := svc.Create(ctx, input)
		if !errors.Is(err, domain.ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("validation failure on invalid input", func(t *testing.T) {
		invalidInput := identity.CreateInput{
			Email: "not-an-email",
			Name:  "Bob",
		}
		_, err := svc.Create(ctx, invalidInput)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})
}
