package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"chorus/internal/domain"
	"chorus/internal/thread"
)

// ThreadRepository is a thread-safe in-memory storage implementation for threads and messages.
type ThreadRepository struct {
	mu           sync.RWMutex
	threads      map[string]*thread.Thread
	messages     map[string][]*thread.Message
	participants map[string]map[string]*thread.Participant // threadID -> tokenHash -> Participant
}

// NewThreadRepository constructs a concrete in-memory thread repository.
func NewThreadRepository() *ThreadRepository {
	return &ThreadRepository{
		threads:      make(map[string]*thread.Thread),
		messages:     make(map[string][]*thread.Message),
		participants: make(map[string]map[string]*thread.Participant),
	}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
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
	copied.ParticipantToken = "" // Do not store raw token in thread metadata
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
	copied.ParticipantToken = ""
	return &copied, nil
}

func (r *ThreadRepository) ListThreads(ctx context.Context, boardSlug ...string) ([]*thread.Thread, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	targetBoard := ""
	if len(boardSlug) > 0 && boardSlug[0] != "" && boardSlug[0] != "all" {
		targetBoard = boardSlug[0]
	}

	result := make([]*thread.Thread, 0, len(r.threads))
	for _, t := range r.threads {
		if targetBoard != "" && t.BoardSlug != targetBoard {
			continue
		}
		copied := *t
		copied.ParticipantToken = ""
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
	copied.ParticipantToken = "" // Do not store raw token in message records
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
		copied.ParticipantToken = ""
		result = append(result, &copied)
	}
	return result, nil
}

func (r *ThreadRepository) SaveParticipant(ctx context.Context, threadID string, p *thread.Participant) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.participants[threadID]; !exists {
		r.participants[threadID] = make(map[string]*thread.Participant)
	}

	tokenHash := hashToken(p.Token)
	stored := &thread.Participant{
		TokenHash:        tokenHash,
		ConversationName: p.ConversationName,
		AuthorID:         p.AuthorID,
		CreatedAt:        p.CreatedAt,
	}
	r.participants[threadID][tokenHash] = stored
	return nil
}

func (r *ThreadRepository) FindParticipantByToken(ctx context.Context, threadID, token string) (*thread.Participant, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	partMap, exists := r.participants[threadID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	tokenHash := hashToken(token)
	p, exists := partMap[tokenHash]
	if !exists || p == nil {
		// Backward compatibility fallback
		p, exists = partMap[token]
	}
	if !exists || p == nil {
		return nil, domain.ErrNotFound
	}

	copied := *p
	copied.Token = token
	return &copied, nil
}
