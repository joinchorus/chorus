package memory

import (
	"context"
	"sync"

	"chorus/internal/domain"
	"chorus/internal/thread"
)

// ThreadRepository is a thread-safe in-memory storage implementation for threads and messages.
type ThreadRepository struct {
	mu       sync.RWMutex
	threads  map[string]*thread.Thread
	messages map[string][]*thread.Message
}

// NewThreadRepository constructs a concrete in-memory thread repository.
func NewThreadRepository() *ThreadRepository {
	return &ThreadRepository{
		threads:  make(map[string]*thread.Thread),
		messages: make(map[string][]*thread.Message),
	}
}

func (r *ThreadRepository) SaveThread(ctx context.Context, t *thread.Thread) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.threads[t.ID]; exists {
		return domain.ErrAlreadyExists
	}

	copied := *t
	r.threads[t.ID] = &copied
	return nil
}

func (r *ThreadRepository) FindThreadByID(ctx context.Context, id string) (*thread.Thread, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.threads[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	copied := *t
	return &copied, nil
}

func (r *ThreadRepository) ListThreads(ctx context.Context) ([]*thread.Thread, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*thread.Thread, 0, len(r.threads))
	for _, t := range r.threads {
		copied := *t
		result = append(result, &copied)
	}
	return result, nil
}

func (r *ThreadRepository) SaveMessage(ctx context.Context, m *thread.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.threads[m.ThreadID]; !exists {
		return domain.ErrNotFound
	}

	copied := *m
	r.messages[m.ThreadID] = append(r.messages[m.ThreadID], &copied)
	return nil
}

func (r *ThreadRepository) ListMessagesByThreadID(ctx context.Context, threadID string) ([]*thread.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, exists := r.threads[threadID]; !exists {
		return nil, domain.ErrNotFound
	}

	msgs := r.messages[threadID]
	result := make([]*thread.Message, 0, len(msgs))
	for _, m := range msgs {
		copied := *m
		result = append(result, &copied)
	}
	return result, nil
}
