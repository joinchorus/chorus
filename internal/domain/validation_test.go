package domain_test

import (
	"errors"
	"strings"
	"testing"

	"chorus/internal/domain"
)

func TestValidation(t *testing.T) {
	t.Run("ValidateTitle", func(t *testing.T) {
		if err := domain.ValidateTitle(""); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for empty title, got %v", err)
		}
		if err := domain.ValidateTitle(strings.Repeat("a", 121)); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for >120 chars title, got %v", err)
		}
		if err := domain.ValidateTitle("Valid Title"); err != nil {
			t.Errorf("expected no error for valid title, got %v", err)
		}
	})

	t.Run("ValidateBody", func(t *testing.T) {
		if err := domain.ValidateBody("", true); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for required empty body, got %v", err)
		}
		if err := domain.ValidateBody(strings.Repeat("a", 4001), false); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for >4000 chars body, got %v", err)
		}
		if err := domain.ValidateBody("Valid Message Content", true); err != nil {
			t.Errorf("expected no error for valid body, got %v", err)
		}
	})

	t.Run("ValidateCountry", func(t *testing.T) {
		invalid := "INVALID"
		if err := domain.ValidateCountry(&invalid); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for invalid country, got %v", err)
		}
		valid := "us"
		if err := domain.ValidateCountry(&valid); err != nil {
			t.Errorf("expected no error for us country code, got %v", err)
		}
	})
}
