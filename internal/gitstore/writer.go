package gitstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteJSONFile writes data as indented JSON to the target path.
func WriteJSONFile(fullPath string, v any) error {
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed creating directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed marshaling json: %w", err)
	}

	if err := writeFileAtomic(fullPath, data); err != nil {
		return fmt.Errorf("failed writing json file %s: %w", fullPath, err)
	}

	return nil
}

// writeFileAtomic writes to a temporary file in the same directory and renames it
// into place, so readers (and backup snapshots) see the old or the new content,
// never a partial file. Temporary names start with ".tmp-", which snapshots skip.
func writeFileAtomic(fullPath string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(fullPath), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, fullPath); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

// AppendNDJSONLine appends a single line JSON payload followed by a newline.
func AppendNDJSONLine(fullPath string, v any) error {
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed creating directory %s: %w", dir, err)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed marshaling ndjson line: %w", err)
	}

	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed opening ndjson file %s: %w", fullPath, err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed appending line to ndjson file %s: %w", fullPath, err)
	}

	return file.Close()
}
