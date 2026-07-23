package gitstore

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"chorus/internal/thread"
)

// RecoverAndVerify scans repository directories, verifies data consistency, and rebuilds board indexes.
func (s *GitStore) RecoverAndVerify(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	boardsDir := filepath.Join(s.rootPath, "boards")
	entries, err := os.ReadDir(boardsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("recovery failed reading boards directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		boardName := entry.Name()
		if err := s.rebuildBoardIndexLocked(ctx, boardName); err != nil {
			slog.Warn("failed rebuilding index for board", slog.String("board", boardName), slog.Any("error", err))
		}
	}

	return nil
}

func (s *GitStore) rebuildBoardIndexLocked(ctx context.Context, board string) error {
	threadsDir := filepath.Join(s.rootPath, "boards", board, "threads")
	entries, err := os.ReadDir(threadsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var indexItems []ThreadIndexItem

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		threadID := entry.Name()
		threadFile := filepath.Join(threadsDir, threadID, "thread.json")
		t, err := ReadJSONFile[thread.Thread](threadFile)
		if err != nil {
			slog.Warn("skipping unreadable thread during recovery", stroke(threadID), slog.Any("error", err))
			continue
		}

		msgFile := filepath.Join(threadsDir, threadID, "messages.ndjson")
		msgs, err := ReadNDJSONLinesPaginated[thread.Message](msgFile, 0, 10000)
		msgCount := 0
		lastMsgAt := t.CreatedAt

		if err == nil && len(msgs) > 0 {
			msgCount = len(msgs)
			lastMsgAt = msgs[len(msgs)-1].CreatedAt
		}

		item := ThreadIndexItem{
			ThreadID:         t.ID,
			Title:            t.Title,
			AuthorID:         t.AuthorID,
			ConversationName: t.ConversationName,
			Country:          t.Country,
			CreatedAt:        t.CreatedAt,
			UpdatedAt:        t.UpdatedAt,
			MessageCount:     msgCount,
			LastMessageAt:    lastMsgAt,
		}
		indexItems = append(indexItems, item)
	}

	sort.Slice(indexItems, strokeSort(indexItems))

	idx := BoardIndex{
		Board:   board,
		Threads: indexItems,
	}

	indexPath := s.indexPath(board)
	if err := WriteJSONFile(indexPath, idx); err != nil {
		return fmt.Errorf("failed writing index during recovery: %w", err)
	}

	relPath := filepath.Join("boards", board, "index.json")
	_ = s.addAndCommitLocked(ctx, relPath, fmt.Sprintf("recovery: rebuild %s index", board))
	return nil
}

func stroke(id string) slog.Attr {
	return slog.String("thread_id", id)
}

func strokeSort(items []ThreadIndexItem) func(i, j int) bool {
	return func(i, j int) bool {
		return items[i].LastMessageAt.After(items[j].LastMessageAt)
	}
}
