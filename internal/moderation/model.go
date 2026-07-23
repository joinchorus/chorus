package moderation

import (
	"fmt"
	"strings"
	"time"

	"chorus/internal/domain"
	"chorus/internal/reporting"
	"chorus/internal/thread"
)

// ModerationStatus represents current review state for a report or message.
type ModerationStatus string

const (
	StatusPending   ModerationStatus = "pending"
	StatusReviewed  ModerationStatus = "reviewed"
	StatusDismissed ModerationStatus = "dismissed"
	StatusRemoved   ModerationStatus = "removed"
)

// Validate checks if status is a valid enum value.
func (s ModerationStatus) Validate() error {
	val := strings.ToLower(string(s))
	switch ModerationStatus(val) {
	case StatusPending, StatusReviewed, StatusDismissed, StatusRemoved:
		return nil
	default:
		return fmt.Errorf("%w: invalid moderation status %q", domain.ErrValidation, s)
	}
}

// ModerationAction represents an immutable decision event recorded by a moderator.
type ModerationAction struct {
	ID        string           `json:"id"`
	ReportID  string           `json:"report_id"`
	ThreadID  string           `json:"thread_id"`
	MessageID string           `json:"message_id"`
	Status    ModerationStatus `json:"status"`
	Note      string           `json:"note,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// ModerationQueueItem combines a report, reported message detail, and current moderation status.
type ModerationQueueItem struct {
	Report        *reporting.Report   `json:"report"`
	Message       *thread.Message     `json:"message"`
	CurrentStatus ModerationStatus    `json:"current_status"`
	History       []*ModerationAction `json:"history"`
}

// SubmitActionInput holds request parameters for taking a moderation action.
type SubmitActionInput struct {
	Status ModerationStatus `json:"status"`
	Note   string           `json:"note"`
}
