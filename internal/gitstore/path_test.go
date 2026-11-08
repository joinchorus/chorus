package gitstore_test

import (
	"context"
	"os/exec"
	"testing"

	"chorus/internal/gitstore"
)

func TestAddAndCommitRejectsPathsOutsideTheRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	store := gitstore.NewGitStore(t.TempDir())
	for _, p := range []string{"../escape", "../../etc/passwd", "/etc/passwd", ""} {
		if err := store.AddAndCommit(context.Background(), p, "test"); err == nil {
			t.Errorf("AddAndCommit(%q) succeeded, want an error", p)
		}
	}
}
