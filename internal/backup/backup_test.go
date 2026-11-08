package backup_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"chorus/internal/backup"
)

// fakeGCS implements the few Cloud Storage JSON API calls the backup uses, with
// generation numbers and ifGenerationMatch preconditions.
type fakeGCS struct {
	t *testing.T

	mu         sync.Mutex
	object     []byte
	generation int64
	uploads    int
	tokenCalls int
	failGet    int // when non-zero, downloads return this status
	noBucket   bool
	failFirst  int // the first N downloads fail with 503, then succeed
	getCalls   int // downloads and generation reads that reached the server
}

func (f *fakeGCS) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.tokenCalls++
		f.mu.Unlock()
		if r.Header.Get("Metadata-Flavor") != "Google" {
			http.Error(w, "missing Metadata-Flavor", http.StatusForbidden)
			return
		}
		fmt.Fprint(w, `{"access_token":"test-token","expires_in":3600}`)
	})
	mux.HandleFunc("/storage/v1/b/bkt/o/", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		f.getCalls++
		if f.failFirst > 0 {
			f.failFirst--
			http.Error(w, "temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		if f.noBucket {
			http.Error(w, `{"error":{"code":404,"message":"The specified bucket does not exist."}}`, http.StatusNotFound)
			return
		}
		if f.failGet != 0 {
			http.Error(w, "forced failure", f.failGet)
			return
		}
		if f.object == nil {
			http.Error(w, `{"error":{"code":404,"message":"No such object: bkt/data.tar.gz"}}`, http.StatusNotFound)
			return
		}
		if r.URL.Query().Get("alt") == "media" {
			w.Header().Set("X-Goog-Generation", strconv.FormatInt(f.generation, 10))
			_, _ = w.Write(f.object)
			return
		}
		fmt.Fprintf(w, `{"generation":"%d"}`, f.generation)
	})
	mux.HandleFunc("/upload/storage/v1/b/bkt/o", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		want, _ := strconv.ParseInt(r.URL.Query().Get("ifGenerationMatch"), 10, 64)
		if want != f.generation {
			http.Error(w, "precondition failed", http.StatusPreconditionFailed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		f.object = body
		f.generation++
		f.uploads++
		fmt.Fprintf(w, `{"generation":"%d"}`, f.generation)
	})
	return mux
}

func newBackup(t *testing.T, fake *fakeGCS, dir string) *backup.Backup {
	t.Helper()
	srv := httptest.NewServer(fake.handler())
	t.Cleanup(srv.Close)
	b := backup.New(backup.Config{
		Bucket:   "bkt",
		Object:   "data.tar.gz",
		DataDir:  dir,
		APIBase:  srv.URL,
		TokenURL: srv.URL + "/token",
		// Retries must not slow the tests down.
		RetryDelay: time.Millisecond,
	})
	if b == nil {
		t.Fatal("New returned nil with a bucket configured")
	}
	return b
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRoundTripRestoresTheDataDirectory(t *testing.T) {
	fake := &fakeGCS{t: t}
	src := filepath.Join(t.TempDir(), "data")
	write(t, filepath.Join(src, "boards", "tech", "threads", "thd_1", "thread.json"), `{"id":"thd_1"}`)
	write(t, filepath.Join(src, "boards", "tech", "threads", "thd_1", "messages.ndjson"), "{\"n\":1}\n{\"n\":2}\n")
	write(t, filepath.Join(src, ".git", "config"), "must not be backed up")
	write(t, filepath.Join(src, "boards", ".tmp-123"), "half written")

	first := newBackup(t, fake, src)
	if restored, err := first.Restore(context.Background()); err != nil || restored {
		t.Fatalf("Restore with no snapshot = %v, %v; want false, nil", restored, err)
	}
	if err := first.Flush(context.Background()); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	dst := filepath.Join(t.TempDir(), "restored")
	second := newBackup(t, fake, dst)
	restored, err := second.Restore(context.Background())
	if err != nil || !restored {
		t.Fatalf("Restore = %v, %v; want true, nil", restored, err)
	}
	if got := read(t, filepath.Join(dst, "boards", "tech", "threads", "thd_1", "messages.ndjson")); got != "{\"n\":1}\n{\"n\":2}\n" {
		t.Errorf("restored messages = %q", got)
	}
	if _, err := os.Stat(filepath.Join(dst, ".git")); !os.IsNotExist(err) {
		t.Errorf(".git must not be restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "boards", ".tmp-123")); !os.IsNotExist(err) {
		t.Errorf("temporary files must not be restored: %v", err)
	}
}

func TestFlushOnlyUploadsWhenDataChanged(t *testing.T) {
	fake := &fakeGCS{t: t}
	dir := filepath.Join(t.TempDir(), "data")
	write(t, filepath.Join(dir, "a.json"), "1")
	b := newBackup(t, fake, dir)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := b.Flush(ctx); err != nil {
			t.Fatalf("Flush %d: %v", i, err)
		}
	}
	if fake.uploads != 1 {
		t.Fatalf("uploads = %d after three flushes of unchanged data, want 1", fake.uploads)
	}

	write(t, filepath.Join(dir, "a.json"), "22")
	if err := b.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if fake.uploads != 2 {
		t.Fatalf("uploads = %d after a change, want 2", fake.uploads)
	}
	if fake.tokenCalls != 1 {
		t.Errorf("token fetched %d times, want it cached (1)", fake.tokenCalls)
	}
}

func TestFlushRefusesToOverwriteANewerSnapshot(t *testing.T) {
	fake := &fakeGCS{t: t}
	dirA := filepath.Join(t.TempDir(), "a")
	dirB := filepath.Join(t.TempDir(), "b")
	ctx := context.Background()

	// Two overlapping instances both start empty and restore "no snapshot yet",
	// then each receives writes.
	a, b := newBackup(t, fake, dirA), newBackup(t, fake, dirB)
	if _, err := a.Restore(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Restore(ctx); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dirA, "x.json"), "from A")
	write(t, filepath.Join(dirB, "x.json"), "from B")
	if err := a.Flush(ctx); err != nil {
		t.Fatalf("first writer: %v", err)
	}
	// b still believes no snapshot exists, so its upload must be rejected.
	if err := b.Flush(ctx); !errors.Is(err, backup.ErrConflict) {
		t.Fatalf("second writer err = %v, want ErrConflict", err)
	}
	if fake.uploads != 1 {
		t.Errorf("uploads = %d, the conflicting upload must not be stored", fake.uploads)
	}
}

func TestRestoreSkipsANonEmptyDirectory(t *testing.T) {
	fake := &fakeGCS{t: t}
	src := filepath.Join(t.TempDir(), "src")
	write(t, filepath.Join(src, "a.json"), "stored")
	if err := newBackup(t, fake, src).Flush(context.Background()); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(t.TempDir(), "dst")
	write(t, filepath.Join(dst, "keep.json"), "local data")
	restored, err := newBackup(t, fake, dst).Restore(context.Background())
	if err != nil || restored {
		t.Fatalf("Restore = %v, %v; want false, nil", restored, err)
	}
	if got := read(t, filepath.Join(dst, "keep.json")); got != "local data" {
		t.Errorf("local data was changed: %q", got)
	}
}

func TestRestoreFailsClosedWhenStorageErrors(t *testing.T) {
	fake := &fakeGCS{t: t, failGet: http.StatusInternalServerError}
	dir := filepath.Join(t.TempDir(), "data")
	if _, err := newBackup(t, fake, dir).Restore(context.Background()); err == nil {
		t.Fatal("Restore must report storage errors so the app does not start empty and overwrite the snapshot")
	}
}

func maliciousArchive(t *testing.T, name string, typeflag byte, linkname string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: name, Typeflag: typeflag, Mode: 0o644, Size: 0, Linkname: linkname}
	if typeflag == tar.TypeReg {
		hdr.Size = 4
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if typeflag == tar.TypeReg {
		_, _ = tw.Write([]byte("evil"))
	}
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func TestRestoreRejectsUnsafeArchiveEntries(t *testing.T) {
	cases := map[string][]byte{
		"parent traversal":  maliciousArchive(t, "../../escape.txt", tar.TypeReg, ""),
		"absolute path":     maliciousArchive(t, "/tmp/escape.txt", tar.TypeReg, ""),
		"nested traversal":  maliciousArchive(t, "boards/../../escape.txt", tar.TypeReg, ""),
		"symlink entry":     maliciousArchive(t, "link", tar.TypeSymlink, "/etc/passwd"),
		"git metadata path": maliciousArchive(t, ".git/hooks/post-commit", tar.TypeReg, ""),
		"not gzip":          []byte("plain text"),
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			fake := &fakeGCS{t: t, object: payload, generation: 7}
			outer := t.TempDir()
			dir := filepath.Join(outer, "data")
			if _, err := newBackup(t, fake, dir).Restore(context.Background()); err == nil {
				t.Fatal("Restore accepted an unsafe snapshot")
			}
			for _, p := range []string{filepath.Join(outer, "escape.txt"), "/tmp/escape.txt"} {
				if _, err := os.Stat(p); err == nil && p != "/tmp/escape.txt" {
					t.Errorf("file written outside the data directory: %s", p)
				}
			}
		})
	}
}

func TestNilBackupIsANoOp(t *testing.T) {
	b := backup.New(backup.Config{})
	if b != nil {
		t.Fatal("New with no bucket must return nil")
	}
	if restored, err := b.Restore(context.Background()); restored || err != nil {
		t.Errorf("nil Restore = %v, %v", restored, err)
	}
	if err := b.Flush(context.Background()); err != nil {
		t.Errorf("nil Flush = %v", err)
	}
}

func TestMissingBucketIsASetupErrorNotAnEmptyStore(t *testing.T) {
	fake := &fakeGCS{t: t, noBucket: true}
	dir := filepath.Join(t.TempDir(), "data")
	b := newBackup(t, fake, dir)
	if _, err := b.Restore(context.Background()); err == nil {
		t.Fatal("Restore must fail when the bucket does not exist")
	}
	write(t, filepath.Join(dir, "a.json"), "1")
	if err := b.Flush(context.Background()); err == nil {
		t.Fatal("Flush must fail when the bucket does not exist")
	}
}

func TestRestoreRetriesTransientErrors(t *testing.T) {
	// A cold instance can fail its first connections to storage; the restore has
	// to ride that out instead of giving up and stopping the app.
	src := filepath.Join(t.TempDir(), "src")
	write(t, filepath.Join(src, "a.json"), "stored")
	fake := &fakeGCS{t: t}
	if err := newBackup(t, fake, src).Flush(context.Background()); err != nil {
		t.Fatal(err)
	}

	fake.failFirst = 3
	fake.getCalls = 0
	dst := filepath.Join(t.TempDir(), "dst")
	restored, err := newBackup(t, fake, dst).Restore(context.Background())
	if err != nil || !restored {
		t.Fatalf("Restore = %v, %v; want true, nil after transient failures", restored, err)
	}
	if fake.getCalls != 4 {
		t.Errorf("download attempts = %d, want 4 (three failures, one success)", fake.getCalls)
	}
	if got := read(t, filepath.Join(dst, "a.json")); got != "stored" {
		t.Errorf("restored content = %q", got)
	}
}

func TestRestoreGivesUpAfterRepeatedFailures(t *testing.T) {
	fake := &fakeGCS{t: t, failGet: http.StatusServiceUnavailable}
	if _, err := newBackup(t, fake, filepath.Join(t.TempDir(), "d")).Restore(context.Background()); err == nil {
		t.Fatal("Restore must fail once the attempts are used up")
	}
	if fake.getCalls < 3 || fake.getCalls > 10 {
		t.Errorf("download attempts = %d, want a few retries, not one and not endless", fake.getCalls)
	}
}

func TestRestoreDoesNotRetryAMissingSnapshot(t *testing.T) {
	fake := &fakeGCS{t: t}
	restored, err := newBackup(t, fake, filepath.Join(t.TempDir(), "d")).Restore(context.Background())
	if err != nil || restored {
		t.Fatalf("Restore = %v, %v; want false, nil", restored, err)
	}
	if fake.getCalls != 1 {
		t.Errorf("download attempts = %d, a missing snapshot is final and must not be retried", fake.getCalls)
	}
}

func TestRestoreHonoursContextCancellationWhileRetrying(t *testing.T) {
	fake := &fakeGCS{t: t, failGet: http.StatusServiceUnavailable}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := newBackup(t, fake, filepath.Join(t.TempDir(), "d")).Restore(ctx); err == nil {
		t.Fatal("Restore must stop when its context is cancelled")
	}
}

func TestFirstFlushWithNothingStoredDoesNotUploadAnEmptyArchive(t *testing.T) {
	fake := &fakeGCS{t: t}
	dir := filepath.Join(t.TempDir(), "data")
	b := newBackup(t, fake, dir)
	ctx := context.Background()
	if _, err := b.Restore(ctx); err != nil {
		t.Fatal(err)
	}
	if err := b.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if fake.uploads != 0 {
		t.Fatalf("uploads = %d for an empty data directory, want 0", fake.uploads)
	}
	write(t, filepath.Join(dir, "a.json"), "1")
	if err := b.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if fake.uploads != 1 {
		t.Fatalf("uploads = %d after real data appeared, want 1", fake.uploads)
	}
}
