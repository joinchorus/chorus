package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// skipEntry reports whether a path inside the data directory is left out of
// snapshots: git metadata (hosting without persistent disk has none, and a restored
// copy is re-initialised) and the temporary files atomic writes leave behind.
func skipEntry(rel string) bool {
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if part == ".git" || strings.HasPrefix(part, ".tmp-") {
			return true
		}
	}
	return false
}

// listEntries returns the relative paths of the directories and regular files to
// snapshot, sorted so archives and fingerprints are deterministic.
func listEntries(dir string) ([]string, error) {
	var rels []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(dir, p)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		if skipEntry(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || d.Type().IsRegular() {
			rels = append(rels, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(rels)
	return rels, nil
}

// fingerprint summarises names, sizes and modification times, so unchanged data
// is not uploaded again.
func fingerprint(dir string) (string, error) {
	rels, err := listEntries(dir)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	for _, rel := range rels {
		info, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s|%d|%d\n", rel, info.Size(), info.ModTime().UnixNano())
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// writeArchive writes dir as a gzip-compressed tar stream.
func writeArchive(w io.Writer, dir string) error {
	rels, err := listEntries(dir)
	if err != nil {
		return err
	}
	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)
	for _, rel := range rels {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		info, err := os.Lstat(full)
		if err != nil {
			if os.IsNotExist(err) {
				continue // removed since the walk
			}
			return err
		}
		hdr := &tar.Header{Name: rel, ModTime: info.ModTime()}
		if info.IsDir() {
			hdr.Typeflag, hdr.Name, hdr.Mode = tar.TypeDir, rel+"/", 0o755
		} else {
			hdr.Typeflag, hdr.Size, hdr.Mode = tar.TypeReg, info.Size(), 0o644
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.IsDir() {
			continue
		}
		if err := copyFile(tw, full, info.Size()); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// copyFile copies exactly size bytes, so a file that grows or shrinks while it is
// read cannot corrupt the tar stream.
func copyFile(w io.Writer, full string, size int64) error {
	f, err := os.Open(full)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := io.Copy(w, io.LimitReader(f, size))
	if err != nil {
		return err
	}
	if n < size {
		_, err = w.Write(make([]byte, size-n))
	}
	return err
}

// extractArchive unpacks a snapshot into dir through os.Root, which refuses ".."
// elements and symlinks leaving dir, and stops after limit bytes.
func extractArchive(r io.Reader, dir string, limit int64) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()

	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("snapshot is not gzip data: %w", err)
	}
	defer gz.Close()

	var total int64
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading snapshot: %w", err)
		}

		name := path.Clean(strings.TrimSuffix(hdr.Name, "/"))
		if name == "." || path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") || skipEntry(name) {
			return fmt.Errorf("snapshot contains an unsafe path %q", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := root.MkdirAll(name, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			total += hdr.Size
			if total > limit {
				return fmt.Errorf("snapshot is larger than the %d byte restore limit", limit)
			}
			if err := root.MkdirAll(path.Dir(name), 0o755); err != nil {
				return err
			}
			f, err := root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(f, io.LimitReader(tr, hdr.Size))
			if closeErr := f.Close(); copyErr == nil {
				copyErr = closeErr
			}
			if copyErr != nil {
				return copyErr
			}
			_ = root.Chtimes(name, hdr.ModTime, hdr.ModTime)
		default:
			return fmt.Errorf("snapshot contains an unsupported entry %q", hdr.Name)
		}
	}
}
