package identity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chorus/internal/conversationname"
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
	repo     Repository
	idGen    idgen.IDGenerator
	nameGen  conversationname.Generator
	nowClock func() time.Time
}

// NewService constructs a concrete identity Service instance.
func NewService(repo Repository, idGen idgen.IDGenerator, nameGen conversationname.Generator, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	if nameGen == nil {
		nameGen = conversationname.NewDefaultGenerator(nil)
	}
	return &Service{
		repo:     repo,
		idGen:    idGen,
		nameGen:  nameGen,
		nowClock: clock,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Identity, error) {
	convName := strings.TrimSpace(input.ConversationName)
	if convName == "" {
		convName = s.nameGen.Generate(nil)
	}

	idStr, err := s.idGen.GenerateID("usr_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	ident := &Identity{
		ID:               idStr,
		ConversationName: convName,
		CreatedAt:        s.nowClock().UTC(),
	}

	if s.repo != nil {
		_ = s.repo.Save(ctx, ident)
	}

	return ident, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Identity, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrValidation)
	}

	if s.repo == nil {
		return nil, domain.ErrNotFound
	}

	ident, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ident, nil
}
