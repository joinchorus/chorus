package reporting_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"chorus/internal/domain"
	"chorus/internal/gitstore"
	"chorus/internal/idgen"
	"chorus/internal/reporting"
)

// A thread ID taken from a URL can contain "/" once percent-decoded. It must be
// rejected before any path is built, otherwise a report file is written outside
// the data directory.
func TestSubmitReportRejectsTraversalThreadID(t *testing.T) {
	outer := t.TempDir()
	root := filepath.Join(outer, "data")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	svc := reporting.NewService(gitstore.NewGitStore(root), idgen.NewRandomIDGenerator(),
		func() time.Time { return time.Date(2026, 7, 22, 16, 0, 0, 0, time.UTC) })

	_, err := svc.SubmitReport(context.Background(), "thd_/../../../../../escape", "msg_1",
		reporting.CreateReportInput{Reason: reporting.ReasonSpam})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if _, statErr := os.Stat(filepath.Join(outer, "escape")); !os.IsNotExist(statErr) {
		t.Fatalf("a directory was created outside the data root: %v", statErr)
	}
}

func TestSubmitReportRejectsTraversalMessageID(t *testing.T) {
	root := t.TempDir()
	svc := reporting.NewService(gitstore.NewGitStore(root), idgen.NewRandomIDGenerator(), time.Now)

	_, err := svc.SubmitReport(context.Background(), "thd_12345", "msg_/../x",
		reporting.CreateReportInput{Reason: reporting.ReasonSpam})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}
