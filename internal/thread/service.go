package thread

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chorus/internal/conversationname"
	"chorus/internal/domain"
	"chorus/internal/http/httputil"
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
	nameGen  conversationname.Generator
	nowClock func() time.Time
}

// NewService constructs a concrete thread Service instance.
func NewService(repo Repository, idGen idgen.IDGenerator, nameGen conversationname.Generator, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	if nameGen == nil {
		nameGen = conversationname.NewDefaultGenerator(nil)
	}
	return &Service{
		repo:     repo,
		idGen:    idGen,
		nameGen:  nameGen,
		nowClock: clock,
	}
}

func (s *Service) CreateThread(ctx context.Context, input CreateThreadInput, clientIP string) (*Thread, error) {
	title := strings.TrimSpace(input.Title)
	body := strings.TrimSpace(input.Body)

	if err := domain.ValidateTitle(title); err != nil {
		return nil, err
	}
	if err := domain.ValidateBody(body, false); err != nil {
		return nil, err
	}

	authorID, err := s.idGen.GenerateID("usr_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	convName := strings.TrimSpace(input.ConversationName)
	if convName == "" {
		convName = s.nameGen.Generate(nil)
	}

	var countryPtr *string
	if input.ShowCountry {
		countryStr := httputil.ResolveCountryFromIP(clientIP)
		countryPtr = &countryStr
	}
	if err := domain.ValidateCountry(countryPtr); err != nil {
		return nil, err
	}

	threadID, err := s.idGen.GenerateID("thd_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	now := s.nowClock().UTC()
	t := &Thread{
		ID:               threadID,
		Title:            title,
		AuthorID:         authorID,
		ConversationName: convName,
		Country:          countryPtr,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.SaveThread(ctx, t); err != nil {
		return nil, err
	}

	if body != "" {
		msgID, err := s.idGen.GenerateID("msg_")
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
		}

		msg := &Message{
			ID:               msgID,
			ThreadID:         threadID,
			AuthorID:         authorID,
			ConversationName: convName,
			Country:          countryPtr,
			Content:          body,
			CreatedAt:        now,
		}

		if err := s.repo.SaveMessage(ctx, msg); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (s *Service) GetThreadByID(ctx context.Context, id string) (*Thread, error) {
	if err := domain.ValidateID(id, "thd_"); err != nil {
		return nil, err
	}
	return s.repo.FindThreadByID(ctx, id)
}

func (s *Service) GetThreadDetail(ctx context.Context, id string) (*ThreadDetail, error) {
	t, err := s.GetThreadByID(ctx, id)
	if err != nil {
		return nil, err
	}

	msgs, err := s.repo.ListMessagesByThreadID(ctx, t.ID)
	if err != nil {
		return nil, err
	}

	return &ThreadDetail{
		Thread:   t,
		Messages: msgs,
	}, nil
}

func (s *Service) ListThreads(ctx context.Context) ([]*Thread, error) {
	return s.repo.ListThreads(ctx)
}

func (s *Service) AddMessage(ctx context.Context, threadID string, input CreateMessageInput, clientIP string) (*Message, error) {
	if err := domain.ValidateID(threadID, "thd_"); err != nil {
		return nil, err
	}

	body := strings.TrimSpace(input.GetBody())
	if err := domain.ValidateBody(body, true); err != nil {
		return nil, err
	}

	t, err := s.repo.FindThreadByID(ctx, threadID)
	if err != nil {
		return nil, err
	}

	authorID, err := s.idGen.GenerateID("usr_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	// Collect used conversation names in this thread to enforce thread-uniqueness
	var usedNames []string
	if t.ConversationName != "" {
		usedNames = append(usedNames, t.ConversationName)
	}
	existingMsgs, _ := s.repo.ListMessagesByThreadID(ctx, threadID)
	for _, m := range existingMsgs {
		if m.ConversationName != "" {
			usedNames = append(usedNames, m.ConversationName)
		}
	}

	convName := strings.TrimSpace(input.ConversationName)
	isUsed := false
	if convName != "" {
		for _, u := range usedNames {
			if strings.EqualFold(u, convName) {
				isUsed = true
				break
			}
		}
	}

	if convName == "" || isUsed {
		convName = s.nameGen.Generate(usedNames)
	}

	var countryPtr *string
	if input.ShowCountry {
		countryStr := httputil.ResolveCountryFromIP(clientIP)
		countryPtr = &countryStr
	}
	if err := domain.ValidateCountry(countryPtr); err != nil {
		return nil, err
	}

	msgID, err := s.idGen.GenerateID("msg_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	msg := &Message{
		ID:               msgID,
		ThreadID:         threadID,
		AuthorID:         authorID,
		ConversationName: convName,
		Country:          countryPtr,
		Content:          body,
		CreatedAt:        s.nowClock().UTC(),
	}

	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *Service) ListMessages(ctx context.Context, threadID string) ([]*Message, error) {
	if err := domain.ValidateID(threadID, "thd_"); err != nil {
		return nil, err
	}
	return s.repo.ListMessagesByThreadID(ctx, threadID)
}
