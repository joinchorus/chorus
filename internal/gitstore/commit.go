package gitstore

import (
	"context"
	"fmt"
	"os/exec"
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
	if err := s.runGit(ctx, "add", relativePath); err != nil {
		return fmt.Errorf("git add failed for %s: %w", relativePath, err)
	}

	authorFlag := fmt.Sprintf("%s <%s>", defaultAuthorName, defaultAuthorEmail)
	if err := s.runGit(ctx, "commit", "-m", message, "--author", authorFlag); err != nil {
		return fmt.Errorf("git commit failed for %s: %w", relativePath, err)
	}

	return nil
}

func (s *GitStore) runGit(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = s.rootPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v failed: %w (output: %s)", args, err, string(output))
	}
	return nil
}
