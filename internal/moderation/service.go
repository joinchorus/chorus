package moderation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"chorus/internal/domain"
	"chorus/internal/gitstore"
	"chorus/internal/idgen"
	"chorus/internal/reporting"
	"chorus/internal/thread"
)

// ThreadService defines thread operations needed for message inspection.
type ThreadService interface {
	GetThreadByID(ctx context.Context, id string) (*thread.Thread, error)
	ListMessages(ctx context.Context, threadID string) ([]*thread.Message, error)
}

// Service manages human moderation queue workflows and immutable action events.
type Service struct {
	mu            sync.RWMutex
	store         *gitstore.GitStore
	threadService ThreadService
	idGen         idgen.IDGenerator
	nowClock      func() time.Time
}

// NewService constructs a concrete moderation Service instance.
func NewService(store *gitstore.GitStore, threadService ThreadService, idGen idgen.IDGenerator, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		store:         store,
		threadService: threadService,
		idGen:         idGen,
		nowClock:      clock,
	}
}

// ListQueue scans reports, correlates with thread messages, and returns pending/reviewed moderation queue items.
func (s *Service) ListQueue(ctx context.Context) ([]*ModerationQueueItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	threadsDir := filepath.Join(s.store.RootPath(), "boards", "general", "threads")
	threadEntries, err := os.ReadDir(threadsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*ModerationQueueItem{}, nil
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	var items []*ModerationQueueItem

	for _, tEntry := range threadEntries {
		if !tEntry.IsDir() {
			continue
		}
		threadID := tEntry.Name()
		reportsDir := filepath.Join(threadsDir, threadID, "reports")
		rEntries, err := os.ReadDir(reportsDir)
		if err != nil {
			continue
		}

		msgs, _ := s.threadService.ListMessages(ctx, threadID)
		msgMap := make(map[string]*thread.Message)
		for _, m := range msgs {
			msgMap[m.ID] = m
		}

		for _, rEntry := range rEntries {
			if rEntry.IsDir() || filepath.Ext(rEntry.Name()) != ".json" {
				continue
			}

			fullReportPath := filepath.Join(reportsDir, rEntry.Name())
			rpt, err := gitstore.ReadJSONFile[reporting.Report](fullReportPath)
			if err != nil {
				continue
			}

			history, currentStatus := s.getReportHistoryLocked(threadID, rpt.ID)
			msg := msgMap[rpt.MessageID]

			items = append(items, &ModerationQueueItem{
				Report:        rpt,
				Message:       msg,
				CurrentStatus: currentStatus,
				History:       history,
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Report.CreatedAt.After(items[j].Report.CreatedAt)
	})

	return items, nil
}

// GetQueueItemByID finds a specific report by ID and returns its details and history.
func (s *Service) GetQueueItemByID(ctx context.Context, reportID string) (*ModerationQueueItem, error) {
	queue, err := s.ListQueue(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range queue {
		if item.Report.ID == reportID {
			return item, nil
		}
	}

	return nil, domain.ErrNotFound
}

// RecordAction registers an immutable moderation event decision for a report.
func (s *Service) RecordAction(ctx context.Context, reportID string, input SubmitActionInput) (*ModerationAction, error) {
	reportID = strings.TrimSpace(reportID)
	note := strings.TrimSpace(input.Note)

	if err := input.Status.Validate(); err != nil {
		return nil, err
	}
	if input.Status == StatusPending {
		return nil, fmt.Errorf("%w: status cannot be set back to pending", domain.ErrValidation)
	}

	item, err := s.GetQueueItemByID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	actionID, err := s.idGen.GenerateID("mod_")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	action := &ModerationAction{
		ID:        actionID,
		ReportID:  reportID,
		ThreadID:  item.Report.ThreadID,
		MessageID: item.Report.MessageID,
		Status:    input.Status,
		Note:      note,
		CreatedAt: s.nowClock().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	relActionFile := filepath.Join("boards", "general", "threads", item.Report.ThreadID, "moderation", fmt.Sprintf("%s.json", actionID))
	fullActionFile := filepath.Join(s.store.RootPath(), relActionFile)

	if err := gitstore.WriteJSONFile(fullActionFile, action); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	commitMsg := fmt.Sprintf("moderation: action %s status=%s for %s", actionID, input.Status, reportID)
	_ = s.store.AddAndCommit(ctx, relActionFile, commitMsg)

	return action, nil
}

func (s *Service) getReportHistoryLocked(threadID, reportID string) ([]*ModerationAction, ModerationStatus) {
	modDir := filepath.Join(s.store.RootPath(), "boards", "general", "threads", threadID, "moderation")
	entries, err := os.ReadDir(modDir)
	if err != nil {
		return []*ModerationAction{}, StatusPending
	}

	var history []*ModerationAction
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		fullPath := filepath.Join(modDir, entry.Name())
		act, err := gitstore.ReadJSONFile[ModerationAction](fullPath)
		if err == nil && act.ReportID == reportID {
			history = append(history, act)
		}
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].CreatedAt.Before(history[j].CreatedAt)
	})

	currentStatus := StatusPending
	if len(history) > 0 {
		currentStatus = history[len(history)-1].Status
	}

	return history, currentStatus
}
