package reporting

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"chorus/internal/domain"
	"chorus/internal/gitstore"
	"chorus/internal/idgen"
)

// Service manages message report operations and immutable storage.
type Service struct {
	mu       sync.RWMutex
	store    *gitstore.GitStore
	idGen    idgen.IDGenerator
	nowClock func() time.Time
}

// NewService constructs a concrete reporting Service.
func NewService(store *gitstore.GitStore, idGen idgen.IDGenerator, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		store:    store,
		idGen:    idGen,
		nowClock: clock,
	}
}

// SubmitReport validates input, generates unique rpt_<hex> ID, and saves report event to disk and Git log.
func (s *Service) SubmitReport(ctx context.Context, threadID, messageID string, input CreateReportInput) (*Report, error) {
	threadID = strings.TrimSpace(threadID)
	messageID = strings.TrimSpace(messageID)
	details := strings.TrimSpace(input.Details)

	if err := domain.ValidateID(threadID, "thd_"); err != nil {
		return nil, err
	}
	if err := domain.ValidateID(messageID, "msg_"); err != nil {
		return nil, err
	}
	if err := input.Reason.Validate(); err != nil {
		return nil, err
	}
	if len(details) > 1000 {
		return nil, fmt.Errorf("%w: details text must not exceed 1000 characters", domain.ErrValidation)
	}

	reportID, err := s.idGen.GenerateID("rpt_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	now := s.nowClock().UTC()
	report := &Report{
		ID:        reportID,
		ThreadID:  threadID,
		MessageID: messageID,
		Reason:    input.Reason,
		Details:   details,
		CreatedAt: now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	relReportFile := filepath.Join("boards", "general", "threads", threadID, "reports", fmt.Sprintf("%s.json", reportID))
	fullReportFile := filepath.Join(s.store.RootPath(), relReportFile)

	if err := gitstore.WriteJSONFile(fullReportFile, report); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	commitMsg := fmt.Sprintf("report: create %s for %s", reportID, messageID)
	_ = s.store.AddAndCommit(ctx, relReportFile, commitMsg)

	return report, nil
}
