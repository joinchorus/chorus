package gitstore_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"chorus/internal/gitstore"
)

func TestWriteJSONFileReplacesAtomicallyAndLeavesNoTemporaryFiles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "boards", "x", "thread.json")

	for _, v := range []map[string]int{{"n": 1}, {"n": 2}} {
		if err := gitstore.WriteJSONFile(target, v); err != nil {
			t.Fatalf("WriteJSONFile: %v", err)
		}
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"n": 2`) {
		t.Errorf("file = %q, want the second write", got)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Errorf("mode = %v, want 0644", info.Mode().Perm())
	}
	entries, err := os.ReadDir(filepath.Dir(target))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("temporary file left behind: %s", e.Name())
		}
	}
}
