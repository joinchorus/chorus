package thread

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chorus/internal/domain"
	"chorus/internal/idgen"
)

// Repository defines storage operations required by the Thread service.
type Repository interface {
	SaveThread(ctx context.Context, t *Thread) error
	FindThreadByID(ctx context.Context, id string) (*Thread, error)
	ListThreads(ctx context.Context) ([]*Thread, error)

	SaveMessage(ctx context.Context, m *Message) error
	ListMessagesByThreadID(ctx context.Context, threadID string) ([]*Message, error)
}

// Service handles thread and message business logic and validation.
type Service struct {
	repo     Repository
	idGen    idgen.IDGenerator
	nowClock func() time.Time
}

// NewService constructs a concrete thread Service instance.
func NewService(repo Repository, idGen idgen.IDGenerator, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		repo:     repo,
		idGen:    idGen,
		nowClock: clock,
	}
}

func (s *Service) CreateThread(ctx context.Context, input CreateThreadInput) (*Thread, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.AuthorID = strings.TrimSpace(input.AuthorID)

	if input.Title == "" {
		return nil, fmt.Errorf("%w: thread title is required", domain.ErrValidation)
	}
	if input.AuthorID == "" {
		return nil, fmt.Errorf("%w: author_id is required", domain.ErrValidation)
	}

	id, err := s.idGen.GenerateID("thd_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	now := s.nowClock().UTC()
	t := &Thread{
		ID:        id,
		Title:     input.Title,
		AuthorID:  input.AuthorID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.SaveThread(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) GetThreadByID(ctx context.Context, id string) (*Thread, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: thread id is required", domain.ErrValidation)
	}
	return s.repo.FindThreadByID(ctx, id)
}

func (s *Service) ListThreads(ctx context.Context) ([]*Thread, error) {
	return s.repo.ListThreads(ctx)
}

func (s *Service) AddMessage(ctx context.Context, threadID string, input CreateMessageInput) (*Message, error) {
	threadID = strings.TrimSpace(threadID)
	input.AuthorID = strings.TrimSpace(input.AuthorID)
	input.Content = strings.TrimSpace(input.Content)

	if threadID == "" {
		return nil, fmt.Errorf("%w: thread_id is required", domain.ErrValidation)
	}
	if input.AuthorID == "" {
		return nil, fmt.Errorf("%w: author_id is required", domain.ErrValidation)
	}
	if input.Content == "" {
		return nil, fmt.Errorf("%w: message content cannot be empty", domain.ErrValidation)
	}

	msgID, err := s.idGen.GenerateID("msg_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	msg := &Message{
		ID:        msgID,
		ThreadID:  threadID,
		AuthorID:  input.AuthorID,
		Content:   input.Content,
		CreatedAt: s.nowClock().UTC(),
	}

	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *Service) ListMessages(ctx context.Context, threadID string) ([]*Message, error) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return nil, fmt.Errorf("%w: thread_id is required", domain.ErrValidation)
	}
	return s.repo.ListMessagesByThreadID(ctx, threadID)
}
