package reporting

import (
	"fmt"
	"strings"
	"time"

	"chorus/internal/domain"
)

// ReportReason specifies allowed message report category types.
type ReportReason string

const (
	ReasonSpam       ReportReason = "spam"
	ReasonHarassment ReportReason = "harassment"
	ReasonIllegal    ReportReason = "illegal"
	ReasonViolence   ReportReason = "violence"
	ReasonCopyright  ReportReason = "copyright"
	ReasonOther      ReportReason = "other"
)

// Validate checks if reason is valid enum value.
func (r ReportReason) Validate() error {
	s := strings.ToLower(string(r))
	switch ReportReason(s) {
	case ReasonSpam, ReasonHarassment, ReasonIllegal, ReasonViolence, ReasonCopyright, ReasonOther:
		return nil
	default:
		return fmt.Errorf("%w: invalid report reason %q", domain.ErrValidation, r)
	}
}

// Report represents an immutable user flag for a message.
type Report struct {
	ID        string       `json:"id"`
	ThreadID  string       `json:"thread_id"`
	MessageID string       `json:"message_id"`
	Reason    ReportReason `json:"reason"`
	Details   string       `json:"details,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// CreateReportInput holds HTTP request body parameters.
type CreateReportInput struct {
	Reason  ReportReason `json:"reason"`
	Details string       `json:"details"`
}
