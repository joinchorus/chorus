package memory

import (
	"context"
	"sync"

	"chorus/internal/domain"
	"chorus/internal/thread"
)

type threadRepository struct {
	mu       sync.RWMutex
	threads  map[string]*thread.Thread
	messages map[string][]*thread.Message
}

// NewThreadRepository returns a thread-safe in-memory thread repository.
func NewThreadRepository() thread.Repository {
	return &threadRepository{
		threads:  make(map[string]*thread.Thread),
		messages: make(map[string][]*thread.Message),
	}
}

func (r *threadRepository) SaveThread(_ context.Context, t *thread.Thread) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.threads[t.ID]; exists {
		return domain.ErrAlreadyExists
	}

	copied := *t
	r.threads[t.ID] = &copied
	return nil
}

func (r *threadRepository) FindThreadByID(_ context.Context, id string) (*thread.Thread, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.threads[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	copied := *t
	return &copied, nil
}

func (r *threadRepository) ListThreads(_ context.Context) ([]*thread.Thread, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*thread.Thread, 0, len(r.threads))
	for _, t := range r.threads {
		copied := *t
		result = append(result, &copied)
	}
	return result, nil
}

func (r *threadRepository) SaveMessage(_ context.Context, m *thread.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.threads[m.ThreadID]; !exists {
		return domain.ErrNotFound
	}

	copied := *m
	r.messages[m.ThreadID] = append(r.messages[m.ThreadID], &copied)
	return nil
}

func (r *threadRepository) ListMessagesByThreadID(_ context.Context, threadID string) ([]*thread.Message, error) {
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
