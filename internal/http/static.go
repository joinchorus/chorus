package http

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// fallbackIndexHTML is served when the SPA bundle has not been built yet.
const fallbackIndexHTML = `<!DOCTYPE html><html><head><title>Chorus</title></head><body><div id="root"></div></body></html>`

// hasPathPrefix reports whether p is prefix itself or lies below it, matching on
// whole path segments ("/landing" matches "/landing/x" but not "/landing..").
func hasPathPrefix(p, prefix string) bool {
	return p == prefix || strings.HasPrefix(p, prefix+"/")
}

// resolveDir returns dir, or the same directory three levels up when the process
// runs from a package directory (tests).
func resolveDir(dir string) string {
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	parent := filepath.Join("..", "..", "..", dir)
	if info, err := os.Stat(parent); err == nil && info.IsDir() {
		return parent
	}
	return dir
}

// openUnder opens the regular file that urlPath names inside dir. The lookup goes
// through os.Root, which refuses ".." elements and symbolic links that point
// outside dir, so a request path cannot reach anything else on disk.
func openUnder(dir, urlPath string) (*os.File, fs.FileInfo, error) {
	root, err := os.OpenRoot(resolveDir(dir))
	if err != nil {
		return nil, nil, err
	}
	defer root.Close()

	rel := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
	if rel == "" {
		rel = "."
	}
	f, err := root.Open(rel)
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		_ = f.Close()
		return nil, nil, fs.ErrNotExist
	}
	return f, info, nil
}

// serveFrom writes the file at urlPath below dir and reports whether one was found.
func serveFrom(w http.ResponseWriter, r *http.Request, dir, urlPath string) bool {
	f, info, err := openUnder(dir, urlPath)
	if err != nil {
		return false
	}
	defer f.Close()
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
	return true
}

func writeFallbackIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fallbackIndexHTML))
}
