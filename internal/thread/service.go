package thread

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chorus/internal/board"
	"chorus/internal/conversationname"
	"chorus/internal/domain"
	"chorus/internal/http/httputil"
	"chorus/internal/idgen"
)

// Repository defines storage operations required by the Thread service.
type Repository interface {
	SaveThread(ctx context.Context, t *Thread) error
	FindThreadByID(ctx context.Context, id string) (*Thread, error)
	ListThreads(ctx context.Context, boardSlug ...string) ([]*Thread, error)

	SaveMessage(ctx context.Context, m *Message) error
	ListMessagesByThreadID(ctx context.Context, threadID string) ([]*Message, error)

	SaveParticipant(ctx context.Context, threadID string, p *Participant) error
	FindParticipantByToken(ctx context.Context, threadID, token string) (*Participant, error)
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

	// 1. Board Resolution & Validation
	boardSlug := strings.TrimSpace(strings.ToLower(input.BoardSlug))
	if boardSlug == "" && input.Topic != "" {
		if b := board.GetBoardBySlug(strings.ToLower(input.Topic)); b != nil {
			boardSlug = b.Slug
		} else {
			for _, sb := range board.SystemBoards {
				if strings.EqualFold(sb.DisplayName, input.Topic) {
					boardSlug = sb.Slug
					break
				}
			}
		}
	}
	if boardSlug == "" {
		boardSlug = "technology" // Default system board
	}

	b := board.GetBoardBySlug(boardSlug)
	if b == nil {
		return nil, fmt.Errorf("%w: invalid board slug %q", domain.ErrValidation, boardSlug)
	}

	authorID, err := s.idGen.GenerateID("usr_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	partToken, err := s.idGen.GenerateID("ptk_")
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
		Topic:            b.DisplayName,
		BoardSlug:        b.Slug,
		BoardDisplayName: b.DisplayName,
		AuthorID:         authorID,
		ConversationName: convName,
		Country:          countryPtr,
		ParticipantToken: partToken,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.SaveThread(ctx, t); err != nil {
		return nil, err
	}

	// Register initial thread participant
	initParticipant := &Participant{
		Token:            partToken,
		ConversationName: convName,
		AuthorID:         authorID,
		CreatedAt:        now,
	}
	_ = s.repo.SaveParticipant(ctx, threadID, initParticipant)

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
			ParticipantToken: partToken,
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
	t, err := s.repo.FindThreadByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t != nil {
		t.ParticipantToken = ""
	}
	return t, nil
}

func (s *Service) GetThreadDetail(ctx context.Context, id string) (*ThreadDetail, error) {
	t, err := s.GetThreadByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t != nil {
		t.ParticipantToken = ""
	}

	msgs, err := s.repo.ListMessagesByThreadID(ctx, t.ID)
	if err != nil {
		return nil, err
	}

	for _, m := range msgs {
		if m != nil {
			m.ParticipantToken = ""
		}
	}

	return &ThreadDetail{
		Thread:   t,
		Messages: msgs,
	}, nil
}

func (s *Service) ListThreads(ctx context.Context, boardSlug ...string) ([]*Thread, error) {
	threads, err := s.repo.ListThreads(ctx, boardSlug...)
	if err != nil {
		return nil, err
	}
	for _, t := range threads {
		if t != nil {
			t.ParticipantToken = ""
		}
	}
	return threads, nil
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

	now := s.nowClock().UTC()
	var authorID string
	var convName string
	var ptk string

	// Check if participant token was provided for thread continuity
	token := strings.TrimSpace(input.ParticipantToken)
	if token != "" {
		if p, err := s.repo.FindParticipantByToken(ctx, threadID, token); err == nil && p != nil {
			authorID = p.AuthorID
			convName = p.ConversationName
			ptk = p.Token
		}
	}

	// If no valid token found, establish a new thread participant
	if authorID == "" {
		var err error
		authorID, err = s.idGen.GenerateID("usr_")
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
		}

		ptk, err = s.idGen.GenerateID("ptk_")
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
		}

		usedNames := s.collectUsedNames(ctx, threadID, t.ConversationName)
		reqName := strings.TrimSpace(input.ConversationName)
		isUsed := false
		if reqName != "" {
			for _, u := range usedNames {
				if strings.EqualFold(u, reqName) {
					isUsed = true
					break
				}
			}
		}

		if reqName == "" || isUsed {
			convName = s.nameGen.Generate(usedNames)
		} else {
			convName = reqName
		}

		newParticipant := &Participant{
			Token:            ptk,
			ConversationName: convName,
			AuthorID:         authorID,
			CreatedAt:        now,
		}
		_ = s.repo.SaveParticipant(ctx, threadID, newParticipant)
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
		ParticipantToken: ptk,
		CreatedAt:        now,
	}

	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *Service) collectUsedNames(ctx context.Context, threadID string, initialAuthorName string) []string {
	var usedNames []string
	if initialAuthorName != "" {
		usedNames = append(usedNames, initialAuthorName)
	}
	existingMsgs, _ := s.repo.ListMessagesByThreadID(ctx, threadID)
	for _, m := range existingMsgs {
		if m.ConversationName != "" {
			usedNames = append(usedNames, m.ConversationName)
		}
	}
	return usedNames
}

func (s *Service) ListMessages(ctx context.Context, threadID string) ([]*Message, error) {
	if err := domain.ValidateID(threadID, "thd_"); err != nil {
		return nil, err
	}
	return s.repo.ListMessagesByThreadID(ctx, threadID)
}
