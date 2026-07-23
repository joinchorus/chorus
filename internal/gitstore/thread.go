package gitstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"chorus/internal/domain"
	"chorus/internal/thread"
)

type threadRepository struct {
	mu    sync.RWMutex
	store *GitStore
}

// NewThreadRepository returns a Git-backed thread repository instance.
func NewThreadRepository(store *GitStore) (thread.Repository, error) {
	ctx := context.Background()
	if err := store.Init(ctx); err != nil {
		return nil, err
	}
	if err := store.RecoverAndVerify(ctx); err != nil {
		return nil, err
	}
	return &threadRepository{store: store}, nil
}

func (r *threadRepository) SaveThread(ctx context.Context, t *thread.Thread) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	relThreadDir := filepath.Join("boards", "general", "threads", t.ID)
	fullThreadDir := filepath.Join(r.store.rootPath, relThreadDir)

	if err := os.MkdirAll(fullThreadDir, 0755); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	relThreadFile := filepath.Join(relThreadDir, "thread.json")
	fullThreadFile := filepath.Join(r.store.rootPath, relThreadFile)

	if _, err := os.Stat(fullThreadFile); err == nil {
		return domain.ErrAlreadyExists
	}

	if err := WriteJSONFile(fullThreadFile, t); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	commitMsg := fmt.Sprintf("thread: create %s", t.ID)
	if err := r.store.AddAndCommit(ctx, relThreadFile, commitMsg); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	// Synchronize board index
	_ = r.store.UpdateThreadInIndex(ctx, "general", t, nil)
	return nil
}

func (r *threadRepository) FindThreadByID(ctx context.Context, id string) (*thread.Thread, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	fullThreadFile := filepath.Join(r.store.rootPath, "boards", "general", "threads", id, "thread.json")
	t, err := ReadJSONFile[thread.Thread](fullThreadFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	return t, nil
}

func (r *threadRepository) ListThreads(ctx context.Context) ([]*thread.Thread, error) {
	return r.ListThreadsPaginated(ctx, 0, 100)
}

func (r *threadRepository) ListThreadsPaginated(ctx context.Context, offset, limit int) ([]*thread.Thread, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	idx, err := r.store.ReadBoardIndex("general")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 50
	}

	if offset >= len(idx.Threads) {
		return []*thread.Thread{}, nil
	}

	end := offset + limit
	if end > len(idx.Threads) {
		end = len(idx.Threads)
	}

	slicedItems := idx.Threads[offset:end]
	result := make([]*thread.Thread, 0, len(slicedItems))

	for _, item := range slicedItems {
		fullThreadFile := filepath.Join(r.store.rootPath, "boards", "general", "threads", item.ThreadID, "thread.json")
		t, err := ReadJSONFile[thread.Thread](fullThreadFile)
		if err == nil {
			result = append(result, t)
		}
	}

	return result, nil
}

func (r *threadRepository) SaveMessage(ctx context.Context, m *thread.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	relThreadDir := filepath.Join("boards", "general", "threads", m.ThreadID)
	fullThreadFile := filepath.Join(r.store.rootPath, relThreadDir, "thread.json")
	if _, err := os.Stat(fullThreadFile); os.IsNotExist(err) {
		return domain.ErrNotFound
	}

	relMsgFile := filepath.Join(relThreadDir, "messages.ndjson")
	fullMsgFile := filepath.Join(r.store.rootPath, relMsgFile)

	if err := AppendNDJSONLine(fullMsgFile, m); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	commitMsg := fmt.Sprintf("message: append %s", m.ID)
	if err := r.store.AddAndCommit(ctx, relMsgFile, commitMsg); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	// Update index with new message timestamp & count
	_ = r.store.IncrementMessageInIndex(ctx, "general", m.ThreadID, m.CreatedAt)
	return nil
}

func (r *threadRepository) ListMessagesByThreadID(ctx context.Context, threadID string) ([]*thread.Message, error) {
	return r.ListMessagesPaginated(ctx, threadID, 0, 1000)
}

func (r *threadRepository) ListMessagesPaginated(ctx context.Context, threadID string, offset, limit int) ([]*thread.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	fullThreadFile := filepath.Join(r.store.rootPath, "boards", "general", "threads", threadID, "thread.json")
	if _, err := os.Stat(fullThreadFile); os.IsNotExist(err) {
		return nil, domain.ErrNotFound
	}

	fullMsgFile := filepath.Join(r.store.rootPath, "boards", "general", "threads", threadID, "messages.ndjson")
	return ReadNDJSONLinesPaginated[thread.Message](fullMsgFile, offset, limit)
}
