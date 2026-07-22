package gitstore

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ReadJSONFile reads and unmarshals a JSON file into target destination.
func ReadJSONFile[T any](fullPath string) (*T, error) {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("failed reading file %s: %w", fullPath, err)
	}

	var dst T
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil, fmt.Errorf("failed unmarshaling json from %s: %w", fullPath, err)
	}

	return &dst, nil
}

// ReadNDJSONLinesPaginated streams lines from an NDJSON file with pagination (offset and limit).
func ReadNDJSONLinesPaginated[T any](fullPath string, offset, limit int) ([]*T, error) {
	file, err := os.Open(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*T{}, nil
		}
		return nil, fmt.Errorf("failed opening ndjson file %s: %w", fullPath, err)
	}
	defer file.Close()

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100 // default sensible pagination cap
	}

	var results []*T
	scanner := bufio.NewScanner(file)
	currentIndex := 0

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		if currentIndex < offset {
			currentIndex++
			continue
		}

		if len(results) >= limit {
			break
		}

		var item T
		if err := json.Unmarshal(line, &item); err == nil {
			results = append(results, &item)
		}
		currentIndex++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning ndjson file %s: %w", fullPath, err)
	}

	return results, nil
}
