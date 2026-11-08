package gitstore

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	defaultAuthorName  = "Chorus"
	defaultAuthorEmail = "system@chorus.local"
)

// AddAndCommit stages a relative path and executes git commit with standardized metadata.
func (s *GitStore) AddAndCommit(ctx context.Context, relativePath string, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addAndCommitLocked(ctx, relativePath, message)
}

func (s *GitStore) addAndCommitLocked(ctx context.Context, relativePath string, message string) error {
	if !s.hasGit {
		return nil
	}

	if err := validateRelativePath(relativePath); err != nil {
		return err
	}

	// "--" ends option parsing, so a path can never be read as a git flag.
	if err := s.runGit(ctx, "add", "--", relativePath); err != nil {
		return fmt.Errorf("git add failed for %s: %w", relativePath, err)
	}

	authorFlag := fmt.Sprintf("%s <%s>", defaultAuthorName, defaultAuthorEmail)
	if err := s.runGit(ctx, "commit", "-m", message, "--author", authorFlag); err != nil {
		return fmt.Errorf("git commit failed for %s: %w", relativePath, err)
	}

	return nil
}

// validateRelativePath rejects paths that are absolute or climb out of the repository.
func validateRelativePath(p string) error {
	clean := filepath.Clean(p)
	if p == "" || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("invalid repository path %q", p)
	}
	return nil
}

func (s *GitStore) runGit(ctx context.Context, args ...string) error {
	if !s.hasGit {
		return nil
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = s.rootPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v failed: %w (output: %s)", args, err, string(output))
	}
	return nil
}
