package memory

import (
	"context"
	"sync"

	"chorus/internal/domain"
	"chorus/internal/identity"
)

// IdentityRepository is a thread-safe in-memory storage implementation for identities.
type IdentityRepository struct {
	mu      sync.RWMutex
	byID    map[string]*identity.Identity
	byEmail map[string]*identity.Identity
}

// NewIdentityRepository constructs a concrete in-memory identity repository.
func NewIdentityRepository() *IdentityRepository {
	return &IdentityRepository{
		byID:    make(map[string]*identity.Identity),
		byEmail: make(map[string]*identity.Identity),
	}
}

func (r *IdentityRepository) Save(ctx context.Context, id *identity.Identity) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byID[id.ID]; exists {
		return domain.ErrAlreadyExists
	}
	if _, exists := r.byEmail[id.Email]; exists {
		return domain.ErrAlreadyExists
	}

	copied := *id
	r.byID[id.ID] = &copied
	r.byEmail[id.Email] = &copied
	return nil
}

func (r *IdentityRepository) FindByID(ctx context.Context, id string) (*identity.Identity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	ident, exists := r.byID[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	copied := *ident
	return &copied, nil
}

func (r *IdentityRepository) FindByEmail(ctx context.Context, email string) (*identity.Identity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	ident, exists := r.byEmail[email]
	if !exists {
		return nil, domain.ErrNotFound
	}
	copied := *ident
	return &copied, nil
}
