package gitstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// GitStore manages local repository access and synchronization.
type GitStore struct {
	mu       sync.RWMutex
	rootPath string
}

// NewGitStore constructs a GitStore instance for the target root directory.
func NewGitStore(rootPath string) *GitStore {
	return &GitStore{
		rootPath: filepath.Clean(rootPath),
	}
}

// RootPath returns the clean absolute/relative root path of the Git repo.
func (s *GitStore) RootPath() string {
	return s.rootPath
}

// Init ensures the target directory exists, initializes Git if missing, and sets local config.
func (s *GitStore) Init(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.rootPath, 0755); err != nil {
		return fmt.Errorf("failed creating data directory: %w", err)
	}

	gitDir := filepath.Join(s.rootPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := s.runGit(ctx, "init"); err != nil {
			return fmt.Errorf("failed running git init: %w", err)
		}
		_ = s.runGit(ctx, "config", "user.name", defaultAuthorName)
		_ = s.runGit(ctx, "config", "user.email", defaultAuthorEmail)
	}

	return nil
}
