package http

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenUnderStaysInsideTheDirectory(t *testing.T) {
	outer := t.TempDir()
	dir := filepath.Join(outer, "site")
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite := func(p, content string) {
		t.Helper()
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(filepath.Join(dir, "index.html"), "home")
	mustWrite(filepath.Join(dir, "sub", "a.txt"), "inside")
	mustWrite(filepath.Join(outer, "secret.txt"), "outside")
	if err := os.Symlink(filepath.Join(outer, "secret.txt"), filepath.Join(dir, "link.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	read := func(urlPath string) (string, bool) {
		f, _, err := openUnder(dir, urlPath)
		if err != nil {
			return "", false
		}
		defer f.Close()
		b, _ := io.ReadAll(f)
		return string(b), true
	}

	for urlPath, want := range map[string]string{"/index.html": "home", "/sub/a.txt": "inside", "sub/a.txt": "inside"} {
		if got, ok := read(urlPath); !ok || got != want {
			t.Errorf("openUnder(%q) = %q, %v; want %q", urlPath, got, ok, want)
		}
	}

	// Directories, missing files, a symlink pointing outside, and every spelling of
	// ".." must all come back as not found.
	for _, urlPath := range []string{"/", "/sub", "/missing.txt", "/link.txt", "/../secret.txt", "/sub/../../secret.txt", "../secret.txt", "/..%2Fsecret.txt"} {
		if got, ok := read(urlPath); ok {
			t.Errorf("openUnder(%q) = %q, want not found", urlPath, got)
		}
	}
}
