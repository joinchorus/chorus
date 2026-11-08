package domain_test

import (
	"errors"
	"strings"
	"testing"

	"chorus/internal/domain"
)

// IDs are used to build file paths in the git store, so ValidateID must reject
// anything that could leave the intended directory, not just a wrong prefix.
func TestValidateIDRejectsPathTricks(t *testing.T) {
	bad := []string{
		"thd_/../../escape",
		"thd_../x",
		"thd_a/b",
		`thd_a\b`,
		"thd_a b",
		"thd_a\x00b",
		"thd_a.b",
		"thd_",
		"thd_" + strings.Repeat("a", 65),
		"msg_1", // wrong prefix for a thread ID
	}
	for _, id := range bad {
		err := domain.ValidateID(strings.TrimSpace(id), "thd_")
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("ValidateID(%q) = %v, want ErrValidation", id, err)
		}
	}
}

func TestValidateIDAcceptsGeneratedIDs(t *testing.T) {
	good := []struct{ id, prefix string }{
		{"thd_12345", "thd_"},
		{"msg_67890", "msg_"},
		{"thd_" + strings.Repeat("0123456789abcdef", 2), "thd_"}, // idgen: prefix + 32 hex
		{"thd_Mixed-Case_9", "thd_"},
	}
	for _, tc := range good {
		if err := domain.ValidateID(tc.id, tc.prefix); err != nil {
			t.Errorf("ValidateID(%q, %q) = %v, want nil", tc.id, tc.prefix, err)
		}
	}
}
