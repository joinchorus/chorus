package gitstore

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// GitStore manages local repository access and synchronization.
type GitStore struct {
	mu       sync.RWMutex
	rootPath string
	hasGit   bool
}

// NewGitStore constructs a GitStore instance for the target root directory.
func NewGitStore(rootPath string) *GitStore {
	_, err := exec.LookPath("git")
	return &GitStore{
		rootPath: filepath.Clean(rootPath),
		hasGit:   err == nil,
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

	if !s.hasGit {
		return nil
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

// ThreadLocation holds directory and board context for a discovered thread.
type ThreadLocation struct {
	ThreadID  string
	BoardSlug string
	RelDir    string
}

// ListAllBoardSlugs returns all board slugs currently containing directories under boards/.
func (s *GitStore) ListAllBoardSlugs() ([]string, error) {
	boardsDir := filepath.Join(s.rootPath, "boards")
	entries, err := os.ReadDir(boardsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var slugs []string
	for _, entry := range entries {
		if entry.IsDir() {
			slugs = append(slugs, entry.Name())
		}
	}
	return slugs, nil
}

// FindThreadRelDir scans board directories to find where a specific thread is stored.
func (s *GitStore) FindThreadRelDir(threadID string) (relDir string, boardSlug string, err error) {
	boardsDir := filepath.Join(s.rootPath, "boards")
	entries, err := os.ReadDir(boardsDir)
	if err != nil {
		return "", "", fmt.Errorf("failed reading boards directory: %w", err)
	}

	for _, bEntry := range entries {
		if !bEntry.IsDir() {
			continue
		}
		candidateRel := filepath.Join("boards", bEntry.Name(), "threads", threadID)
		candidateThreadFile := filepath.Join(s.rootPath, candidateRel, "thread.json")
		if _, err := os.Stat(candidateThreadFile); err == nil {
			return candidateRel, bEntry.Name(), nil
		}
	}

	return "", "", os.ErrNotExist
}

// ListAllThreadLocations returns all thread locations discovered across all board directories.
func (s *GitStore) ListAllThreadLocations() ([]ThreadLocation, error) {
	boardsDir := filepath.Join(s.rootPath, "boards")
	bEntries, err := os.ReadDir(boardsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ThreadLocation{}, nil
		}
		return nil, err
	}

	var locations []ThreadLocation
	for _, bEntry := range bEntries {
		if !bEntry.IsDir() {
			continue
		}
		boardSlug := bEntry.Name()
		threadsDir := filepath.Join(boardsDir, boardSlug, "threads")
		tEntries, err := os.ReadDir(threadsDir)
		if err != nil {
			continue
		}

		for _, tEntry := range tEntries {
			if !tEntry.IsDir() {
				continue
			}
			threadID := tEntry.Name()
			threadFile := filepath.Join(threadsDir, threadID, "thread.json")
			if _, err := os.Stat(threadFile); err == nil {
				locations = append(locations, ThreadLocation{
					ThreadID:  threadID,
					BoardSlug: boardSlug,
					RelDir:    filepath.Join("boards", boardSlug, "threads", threadID),
				})
			}
		}
	}

	return locations, nil
}
