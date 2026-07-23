package reporting_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"chorus/internal/domain"
	"chorus/internal/gitstore"
	"chorus/internal/idgen"
	"chorus/internal/reporting"
)

func TestReportingService_Submit(t *testing.T) {
	tempDir := t.TempDir()
	store := gitstore.NewGitStore(tempDir)
	gen := idgen.NewRandomIDGenerator()
	fixedTime := time.Date(2026, 7, 22, 16, 0, 0, 0, time.UTC)

	svc := reporting.NewService(store, gen, func() time.Time { return fixedTime })
	ctx := context.Background()

	t.Run("successful report creation", func(t *testing.T) {
		rpt, err := svc.SubmitReport(ctx, "thd_12345", "msg_67890", reporting.CreateReportInput{
			Reason:  reporting.ReasonSpam,
			Details: "Unwanted commercial link",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rpt.ID == "" || rpt.ID[:4] != "rpt_" {
			t.Errorf("expected ID prefix rpt_, got %s", rpt.ID)
		}
		if rpt.Reason != reporting.ReasonSpam {
			t.Errorf("expected reason spam, got %s", rpt.Reason)
		}
	})

	t.Run("invalid reason validation failure", func(t *testing.T) {
		_, err := svc.SubmitReport(ctx, "thd_12345", "msg_67890", reporting.CreateReportInput{
			Reason: "invalid_reason",
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for invalid reason, got %v", err)
		}
	})
}
