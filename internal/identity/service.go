package identity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chorus/internal/domain"
)

// Service defines the business operations for identity management.
type Service interface {
	Create(ctx context.Context, input CreateInput) (*Identity, error)
	GetByID(ctx context.Context, id string) (*Identity, error)
}

type service struct {
	repo      Repository
	generator IDGenerator
}

// NewService constructs an identity Service with required dependencies.
func NewService(repo Repository, generator IDGenerator) Service {
	return &service{
		repo:      repo,
		generator: generator,
	}
}

func (s *service) Create(ctx context.Context, input CreateInput) (*Identity, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Name = strings.TrimSpace(input.Name)

	if input.Email == "" {
		return nil, fmt.Errorf("%w: email is required", domain.ErrValidation)
	}
	if !strings.Contains(input.Email, "@") {
		return nil, fmt.Errorf("%w: invalid email address format", domain.ErrValidation)
	}
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}

	existing, err := s.repo.FindByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("%w: identity with email already exists", domain.ErrAlreadyExists)
	}

	idStr, err := s.generator.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	ident := &Identity{
		ID:        idStr,
		Email:     input.Email,
		Name:      input.Name,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, ident); err != nil {
		return nil, err
	}

	return ident, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Identity, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrValidation)
	}

	ident, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ident, nil
}
