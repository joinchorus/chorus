package gitstore_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"chorus/internal/gitstore"
)

func TestWriteJSONFileCreatesDirectoriesAndReplacesContent(t *testing.T) {
	target := filepath.Join(t.TempDir(), "boards", "x", "thread.json")

	for _, v := range []map[string]int{{"n": 1}, {"n": 2}} {
		if err := gitstore.WriteJSONFile(target, v); err != nil {
			t.Fatalf("WriteJSONFile: %v", err)
		}
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"n": 2`) || strings.Contains(string(got), `"n": 1`) {
		t.Errorf("file = %q, want only the second write", got)
	}
}
