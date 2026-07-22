package memory

import (
	"context"
	"sync"

	"chorus/internal/domain"
	"chorus/internal/identity"
)

type identityRepository struct {
	mu        sync.RWMutex
	byID      map[string]*identity.Identity
	byEmail   map[string]*identity.Identity
}

// NewIdentityRepository returns a thread-safe in-memory identity repository.
func NewIdentityRepository() identity.Repository {
	return &identityRepository{
		byID:    make(map[string]*identity.Identity),
		byEmail: make(map[string]*identity.Identity),
	}
}

func (r *identityRepository) Save(_ context.Context, id *identity.Identity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byID[id.ID]; exists {
		return domain.ErrAlreadyExists
	}
	if _, exists := r.byEmail[id.Email]; exists {
		return domain.ErrAlreadyExists
	}

	// Copy to prevent external mutation
	copied := *id
	r.byID[id.ID] = &copied
	r.byEmail[id.Email] = &copied
	return nil
}

func (r *identityRepository) FindByID(_ context.Context, id string) (*identity.Identity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ident, exists := r.byID[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	copied := *ident
	return &copied, nil
}

func (r *identityRepository) FindByEmail(_ context.Context, email string) (*identity.Identity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ident, exists := r.byEmail[email]
	if !exists {
		return nil, domain.ErrNotFound
	}
	copied := *ident
	return &copied, nil
}
