package identity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chorus/internal/domain"
	"chorus/internal/idgen"
)

// Repository defines storage operations required by the Identity service.
type Repository interface {
	Save(ctx context.Context, id *Identity) error
	FindByID(ctx context.Context, id string) (*Identity, error)
	FindByEmail(ctx context.Context, email string) (*Identity, error)
}

// Service handles identity business logic and validation.
type Service struct {
	repo      Repository
	idGen     idgen.IDGenerator
	nowClock  func() time.Time
}

// NewService constructs a concrete identity Service instance.
func NewService(repo Repository, idGen idgen.IDGenerator, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		repo:     repo,
		idGen:    idGen,
		nowClock: clock,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Identity, error) {
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

	idStr, err := s.idGen.GenerateID("usr_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	ident := &Identity{
		ID:        idStr,
		Email:     input.Email,
		Name:      input.Name,
		CreatedAt: s.nowClock().UTC(),
	}

	if err := s.repo.Save(ctx, ident); err != nil {
		return nil, err
	}

	return ident, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Identity, error) {
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
