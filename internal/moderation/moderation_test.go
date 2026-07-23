package moderation_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"chorus/internal/domain"
	"chorus/internal/gitstore"
	"chorus/internal/idgen"
	"chorus/internal/moderation"
	"chorus/internal/reporting"
	"chorus/internal/thread"
)

func TestModerationService_Workflow(t *testing.T) {
	tempDir := t.TempDir()
	store := gitstore.NewGitStore(tempDir)
	gen := idgen.NewRandomIDGenerator()
	fixedTime := time.Date(2026, 7, 22, 16, 0, 0, 0, time.UTC)

	threadRepo, _ := gitstore.NewThreadRepository(store)
	threadSvc := thread.NewService(threadRepo, gen, nil, func() time.Time { return fixedTime })
	reportSvc := reporting.NewService(store, gen, func() time.Time { return fixedTime })

	modSvc := moderation.NewService(store, threadSvc, gen, func() time.Time { return fixedTime })
	ctx := context.Background()

	// 1. Create Thread and Message
	th, err := threadSvc.CreateThread(ctx, thread.CreateThreadInput{
		Title: "Moderation Queue Thread",
		Body:  "Offensive message content for report test",
	}, "US")
	if err != nil {
		t.Fatalf("failed creating thread: %v", err)
	}

	detail, _ := threadSvc.GetThreadDetail(ctx, th.ID)
	targetMsgID := detail.Messages[0].ID

	// 2. Submit Report
	rpt, err := reportSvc.SubmitReport(ctx, th.ID, targetMsgID, reporting.CreateReportInput{
		Reason:  reporting.ReasonHarassment,
		Details: "Personal attack in thread",
	})
	if err != nil {
		t.Fatalf("failed submitting report: %v", err)
	}

	// 3. List Queue and verify pending status
	queue, err := modSvc.ListQueue(ctx)
	if err != nil {
		t.Fatalf("failed listing queue: %v", err)
	}
	if len(queue) != 1 {
		t.Fatalf("expected 1 item in queue, got %d", len(queue))
	}
	if queue[0].CurrentStatus != moderation.StatusPending {
		t.Errorf("expected status pending, got %s", queue[0].CurrentStatus)
	}

	// 4. Record Moderation Decision (Dismissed)
	act, err := modSvc.RecordAction(ctx, rpt.ID, moderation.SubmitActionInput{
		Status: moderation.StatusDismissed,
		Note:   "Does not violate community policy",
	})
	if err != nil {
		t.Fatalf("failed recording action: %v", err)
	}
	if act.ID[:4] != "mod_" {
		t.Errorf("expected ID prefix mod_, got %s", act.ID)
	}

	// 5. Verify updated status
	item, err := modSvc.GetQueueItemByID(ctx, rpt.ID)
	if err != nil {
		t.Fatalf("failed getting queue item: %v", err)
	}
	if item.CurrentStatus != moderation.StatusDismissed {
		t.Errorf("expected status dismissed, got %s", item.CurrentStatus)
	}
	if len(item.History) != 1 {
		t.Errorf("expected 1 history record, got %d", len(item.History))
	}

	// 6. Validation error when attempting to set status back to pending
	_, err = modSvc.RecordAction(ctx, rpt.ID, moderation.SubmitActionInput{
		Status: moderation.StatusPending,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("expected ErrValidation for pending transition, got %v", err)
	}
}
