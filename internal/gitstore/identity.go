package gitstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"chorus/internal/domain"
	"chorus/internal/identity"
)

type identityRepository struct {
	mu    sync.RWMutex
	store *GitStore
}

// NewIdentityRepository returns a Git-backed identity repository.
func NewIdentityRepository(store *GitStore) (identity.Repository, error) {
	if err := store.Init(context.Background()); err != nil {
		return nil, err
	}
	return &identityRepository{store: store}, nil
}

func (r *identityRepository) Save(ctx context.Context, id *identity.Identity) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	fileName := fmt.Sprintf("%s.json", id.ID)
	relPath := filepath.Join("identities", fileName)
	fullPath := filepath.Join(r.store.rootPath, relPath)

	if _, err := os.Stat(fullPath); err == nil {
		return domain.ErrAlreadyExists
	}

	if err := WriteJSONFile(fullPath, id); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	commitMsg := fmt.Sprintf("identity: create %s", id.ID)
	if err := r.store.AddAndCommit(ctx, relPath, commitMsg); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	return nil
}

func (r *identityRepository) FindByID(ctx context.Context, id string) (*identity.Identity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	fullPath := filepath.Join(r.store.rootPath, "identities", fmt.Sprintf("%s.json", id))
	ident, err := ReadJSONFile[identity.Identity](fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	return ident, nil
}

func (r *identityRepository) FindByEmail(ctx context.Context, email string) (*identity.Identity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	identDir := filepath.Join(r.store.rootPath, "identities")
	entries, err := os.ReadDir(identDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		fullPath := filepath.Join(identDir, entry.Name())
		ident, err := ReadJSONFile[identity.Identity](fullPath)
		if err == nil && ident.Email == email {
			return ident, nil
		}
	}

	return nil, domain.ErrNotFound
}
