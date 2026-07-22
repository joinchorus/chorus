package thread

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"chorus/internal/domain"
)

// Service defines business operations for thread and message management.
type Service interface {
	CreateThread(ctx context.Context, input CreateThreadInput) (*Thread, error)
	GetThreadByID(ctx context.Context, id string) (*Thread, error)
	ListThreads(ctx context.Context) ([]*Thread, error)

	AddMessage(ctx context.Context, threadID string, input CreateMessageInput) (*Message, error)
	ListMessages(ctx context.Context, threadID string) ([]*Message, error)
}

type service struct {
	repo Repository
}

// NewService constructs a thread Service with repository dependency.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateThread(ctx context.Context, input CreateThreadInput) (*Thread, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.AuthorID = strings.TrimSpace(input.AuthorID)

	if input.Title == "" {
		return nil, fmt.Errorf("%w: thread title is required", domain.ErrValidation)
	}
	if input.AuthorID == "" {
		return nil, fmt.Errorf("%w: author_id is required", domain.ErrValidation)
	}

	id, err := generatePrefixedID("thd_")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	now := time.Now().UTC()
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

func (s *service) GetThreadByID(ctx context.Context, id string) (*Thread, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: thread id is required", domain.ErrValidation)
	}
	return s.repo.FindThreadByID(ctx, id)
}

func (s *service) ListThreads(ctx context.Context) ([]*Thread, error) {
	return s.repo.ListThreads(ctx)
}

func (s *service) AddMessage(ctx context.Context, threadID string, input CreateMessageInput) (*Message, error) {
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

	// Verify thread exists
	_, err := s.repo.FindThreadByID(ctx, threadID)
	if err != nil {
		return nil, err
	}

	msgID, err := generatePrefixedID("msg_")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	msg := &Message{
		ID:        msgID,
		ThreadID:  threadID,
		AuthorID:  input.AuthorID,
		Content:   input.Content,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *service) ListMessages(ctx context.Context, threadID string) ([]*Message, error) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return nil, fmt.Errorf("%w: thread_id is required", domain.ErrValidation)
	}

	// Verify thread exists first
	if _, err := s.repo.FindThreadByID(ctx, threadID); err != nil {
		return nil, err
	}

	return s.repo.ListMessagesByThreadID(ctx, threadID)
}

func generatePrefixedID(prefix string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}
