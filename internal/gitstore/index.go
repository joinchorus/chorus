package gitstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"chorus/internal/thread"
)

// ThreadIndexItem represents a cached summary entry in the board index.
type ThreadIndexItem struct {
	ThreadID      string    `json:"thread_id"`
	Title         string    `json:"title"`
	AuthorID      string    `json:"author_id"`
	Country       *string   `json:"country,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	MessageCount  int       `json:"message_count"`
	LastMessageAt time.Time `json:"last_message_at"`
}

// BoardIndex represents the append-friendly index stored at boards/general/index.json.
type BoardIndex struct {
	Board   string            `json:"board"`
	Threads []ThreadIndexItem `json:"threads"`
}

func (s *GitStore) indexPath(board string) string {
	return filepath.Join(s.rootPath, "boards", board, "index.json")
}

// ReadBoardIndex reads the board index from disk.
func (s *GitStore) ReadBoardIndex(board string) (*BoardIndex, error) {
	fullPath := s.indexPath(board)
	idx, err := ReadJSONFile[BoardIndex](fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &BoardIndex{
				Board:   board,
				Threads: []ThreadIndexItem{},
			}, nil
		}
		return nil, err
	}
	return idx, nil
}

// UpdateThreadInIndex adds or updates a thread entry in the index and saves it to disk.
func (s *GitStore) UpdateThreadInIndex(ctx context.Context, board string, t *thread.Thread, initialMsg *thread.Message) error {
	idx, err := s.ReadBoardIndex(board)
	if err != nil {
		return err
	}

	found := false
	for i := range idx.Threads {
		if idx.Threads[i].ThreadID == t.ID {
			idx.Threads[i].Title = t.Title
			idx.Threads[i].UpdatedAt = t.UpdatedAt
			if initialMsg != nil {
				idx.Threads[i].MessageCount++
				idx.Threads[i].LastMessageAt = initialMsg.CreatedAt
			}
			found = true
			break
		}
	}

	if !found {
		msgCount := 0
		lastMsgAt := t.CreatedAt
		if initialMsg != nil {
			msgCount = 1
			lastMsgAt = initialMsg.CreatedAt
		}

		item := ThreadIndexItem{
			ThreadID:      t.ID,
			Title:         t.Title,
			AuthorID:      t.AuthorID,
			Country:       t.Country,
			CreatedAt:     t.CreatedAt,
			UpdatedAt:     t.UpdatedAt,
			MessageCount:  msgCount,
			LastMessageAt: lastMsgAt,
		}
		idx.Threads = append([]ThreadIndexItem{item}, idx.Threads...)
	}

	// Sort threads by LastMessageAt / CreatedAt descending (latest threads first)
	sort.Slice(idx.Threads, func(i, j int) bool {
		return idx.Threads[i].LastMessageAt.After(idx.Threads[j].LastMessageAt)
	})

	fullPath := s.indexPath(board)
	if err := WriteJSONFile(fullPath, idx); err != nil {
		return fmt.Errorf("failed writing board index: %w", err)
	}

	relPath := filepath.Join("boards", board, "index.json")
	_ = s.AddAndCommit(ctx, relPath, fmt.Sprintf("index: update %s", board))
	return nil
}

// IncrementMessageInIndex updates message count and timestamp for a thread in the index.
func (s *GitStore) IncrementMessageInIndex(ctx context.Context, board string, threadID string, msgAt time.Time) error {
	idx, err := s.ReadBoardIndex(board)
	if err != nil {
		return err
	}

	for i := range idx.Threads {
		if idx.Threads[i].ThreadID == threadID {
			idx.Threads[i].MessageCount++
			idx.Threads[i].LastMessageAt = msgAt
			idx.Threads[i].UpdatedAt = msgAt
			break
		}
	}

	sort.Slice(idx.Threads, func(i, j int) bool {
		return idx.Threads[i].LastMessageAt.After(idx.Threads[j].LastMessageAt)
	})

	fullPath := s.indexPath(board)
	if err := WriteJSONFile(fullPath, idx); err != nil {
		return err
	}

	relPath := filepath.Join("boards", board, "index.json")
	_ = s.AddAndCommit(ctx, relPath, fmt.Sprintf("index: touch %s", threadID))
	return nil
}
