// Package backup keeps the data directory alive across restarts on hosting whose
// disk is ephemeral (Firebase App Hosting / Cloud Run). The directory is stored as
// one gzip-compressed tar object in a Cloud Storage bucket: restored at start-up,
// written again after data changes. It uses only the standard library and the
// instance's service-account token from the metadata server.
//
// Run a single instance (maxInstances: 1). Uploads carry an ifGenerationMatch
// precondition, so a second writer cannot silently overwrite a newer snapshot; it
// gets ErrConflict instead.
package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

// ErrConflict means the stored snapshot changed since this instance last read or
// wrote it, so the upload was refused rather than overwriting someone else's data.
var ErrConflict = errors.New("backup: the stored snapshot was changed by another writer")

const (
	// DefaultObject is the object name used when none is configured.
	DefaultObject = "chorus-data.tar.gz"

	defaultAPIBase  = "https://storage.googleapis.com"
	defaultTokenURL = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"
	defaultInterval = 60 * time.Second
	maxRestoreBytes = 1 << 30
	// restoreAttempts connections are tried with a doubling delay (1+2+4+8+16 s by
	// default), which fits inside the 60 s the server allows for the restore.
	restoreAttempts = 6
	defaultRetry    = time.Second
	maxSnapshotWarn = 100 << 20
)

// Config selects where the snapshot lives. An empty Bucket disables backups.
type Config struct {
	Bucket   string
	Object   string
	DataDir  string
	Interval time.Duration
	// RetryDelay is the first wait between restore attempts; it doubles each time.
	RetryDelay time.Duration

	// Endpoints and client are overridable for tests.
	APIBase    string
	TokenURL   string
	HTTPClient *http.Client
}

// Backup stores and restores one snapshot. A nil *Backup is valid and does nothing.
type Backup struct {
	cfg Config

	mu          sync.Mutex
	generation  int64 // object generation we last read or wrote; -1 = unknown
	fingerprint string
	token       string
	tokenExpiry time.Time
}

// New returns nil (backups off) when no bucket is configured.
func New(cfg Config) *Backup {
	if cfg.Bucket == "" {
		return nil
	}
	if cfg.Object == "" {
		cfg.Object = DefaultObject
	}
	if cfg.Interval <= 0 {
		cfg.Interval = defaultInterval
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = defaultRetry
	}
	if cfg.APIBase == "" {
		cfg.APIBase = defaultAPIBase
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = defaultTokenURL
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}
	return &Backup{cfg: cfg, generation: -1}
}

// Restore downloads the snapshot into an empty data directory. It returns false
// when there is nothing to restore (no snapshot yet, or the directory already has
// data). Any other failure is an error: the caller should not start serving, because
// a later upload of an empty directory would replace the stored snapshot.
func (b *Backup) Restore(ctx context.Context) (bool, error) {
	if b == nil {
		return false, nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	empty, err := isEmptyDir(b.cfg.DataDir)
	if err != nil {
		return false, err
	}
	if !empty {
		return false, nil
	}

	resp, err := b.fetch(ctx, b.objectURL("?alt=media"))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		if missingBucket(resp) {
			return false, statusError("download snapshot", resp)
		}
		// Nothing stored yet. Remember what the directory looks like now, so the
		// first flush does not upload an empty archive.
		b.generation = 0
		b.fingerprint, _ = fingerprint(b.cfg.DataDir)
		return false, nil
	case http.StatusOK:
	default:
		return false, statusError("download snapshot", resp)
	}

	generation, err := strconv.ParseInt(resp.Header.Get("X-Goog-Generation"), 10, 64)
	if err != nil {
		return false, fmt.Errorf("backup: download had no usable generation header: %w", err)
	}
	if err := extractArchive(resp.Body, b.cfg.DataDir, maxRestoreBytes); err != nil {
		return false, fmt.Errorf("backup: restore failed: %w", err)
	}
	fp, err := fingerprint(b.cfg.DataDir)
	if err != nil {
		return false, err
	}
	b.generation, b.fingerprint = generation, fp
	return true, nil
}

// Flush uploads the data directory when it changed since the last snapshot.
func (b *Backup) Flush(ctx context.Context) error {
	if b == nil {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	fp, err := fingerprint(b.cfg.DataDir)
	if err != nil {
		return err
	}
	if fp == b.fingerprint {
		return nil
	}

	generation := b.generation
	if generation < 0 {
		if generation, err = b.remoteGeneration(ctx); err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	if err := writeArchive(&buf, b.cfg.DataDir); err != nil {
		return fmt.Errorf("backup: building snapshot: %w", err)
	}
	if buf.Len() > maxSnapshotWarn {
		slog.Warn("backup snapshot is large; consider a persistent disk", slog.Int("bytes", buf.Len()))
	}

	query := fmt.Sprintf("uploadType=media&name=%s&ifGenerationMatch=%d", url.QueryEscape(b.cfg.Object), generation)
	resp, err := b.do(ctx, http.MethodPost, b.cfg.APIBase+"/upload/storage/v1/b/"+url.PathEscape(b.cfg.Bucket)+"/o?"+query, &buf, "application/gzip")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPreconditionFailed {
		return ErrConflict
	}
	if resp.StatusCode != http.StatusOK {
		return statusError("upload snapshot", resp)
	}
	var meta struct {
		Generation string `json:"generation"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return fmt.Errorf("backup: unreadable upload response: %w", err)
	}
	newGeneration, err := strconv.ParseInt(meta.Generation, 10, 64)
	if err != nil {
		return fmt.Errorf("backup: upload response had no usable generation: %w", err)
	}
	b.generation, b.fingerprint = newGeneration, fp
	return nil
}

// Run flushes on an interval until ctx ends. Writes are also flushed on request
// (see middleware.DurableWrites); this loop is the safety net.
func (b *Backup) Run(ctx context.Context) {
	if b == nil {
		return
	}
	ticker := time.NewTicker(b.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := b.Flush(ctx); err != nil {
				slog.Warn("periodic backup failed", slog.Any("error", err))
			}
		}
	}
}

// fetch performs a GET and retries connection failures and 429/5xx answers with a
// doubling delay. A cold instance can fail its first connections to storage
// (TLS handshake timeouts were seen), and giving up there would stop the app.
// Other statuses, including 404, are final and returned to the caller.
func (b *Backup) fetch(ctx context.Context, target string) (*http.Response, error) {
	delay := b.cfg.RetryDelay
	for attempt := 1; ; attempt++ {
		resp, err := b.do(ctx, http.MethodGet, target, nil, "")
		transient := err != nil || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError
		if !transient || attempt == restoreAttempts || ctx.Err() != nil {
			return resp, err
		}
		if err != nil {
			slog.Warn("backup download failed, retrying", slog.Int("attempt", attempt), slog.Any("error", err))
		} else {
			slog.Warn("backup download failed, retrying", slog.Int("attempt", attempt), slog.String("status", resp.Status))
			resp.Body.Close()
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
	}
}

func (b *Backup) remoteGeneration(ctx context.Context) (int64, error) {
	resp, err := b.do(ctx, http.MethodGet, b.objectURL("?fields=generation"), nil, "")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound && !missingBucket(resp) {
		return 0, nil
	}
	if resp.StatusCode != http.StatusOK {
		return 0, statusError("read snapshot generation", resp)
	}
	var meta struct {
		Generation string `json:"generation"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return 0, err
	}
	return strconv.ParseInt(meta.Generation, 10, 64)
}

func (b *Backup) objectURL(suffix string) string {
	return b.cfg.APIBase + "/storage/v1/b/" + url.PathEscape(b.cfg.Bucket) + "/o/" + url.PathEscape(b.cfg.Object) + suffix
}

func (b *Backup) do(ctx context.Context, method, target string, body io.Reader, contentType string) (*http.Response, error) {
	token, err := b.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return b.cfg.HTTPClient.Do(req)
}

// accessToken returns the service-account token from the metadata server, cached
// until shortly before it expires. Callers hold b.mu.
func (b *Backup) accessToken(ctx context.Context) (string, error) {
	if b.token != "" && time.Now().Before(b.tokenExpiry) {
		return b.token, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.cfg.TokenURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := b.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("backup: fetching access token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", statusError("fetch access token", resp)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil || tok.AccessToken == "" {
		return "", fmt.Errorf("backup: unreadable access token response: %v", err)
	}
	b.token = tok.AccessToken
	b.tokenExpiry = time.Now().Add(time.Duration(tok.ExpiresIn)*time.Second - time.Minute)
	return b.token, nil
}

// missingBucket tells "no snapshot object yet" (normal on first run) apart from "the
// bucket does not exist" (a setup error that must not be mistaken for an empty store).
func missingBucket(resp *http.Response) bool {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return bytes.Contains(bytes.ToLower(body), []byte("bucket"))
}

func isEmptyDir(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func statusError(what string, resp *http.Response) error {
	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
	return fmt.Errorf("backup: %s: %s: %s", what, resp.Status, bytes.TrimSpace(snippet))
}
