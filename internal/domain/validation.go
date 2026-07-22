package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	isoCountryRegex = regexp.MustCompile(`^[A-Z]{2}$`)
)

// ValidateTitle verifies title constraints (required, max 120 chars).
func ValidateTitle(title string) error {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if len(trimmed) > 120 {
		return fmt.Errorf("%w: title must not exceed 120 characters", ErrValidation)
	}
	return nil
}

// ValidateBody verifies body/message content constraints (max 4000 chars).
func ValidateBody(body string, required bool) error {
	trimmed := strings.TrimSpace(body)
	if required && trimmed == "" {
		return fmt.Errorf("%w: message content is required", ErrValidation)
	}
	if len(trimmed) > 4000 {
		return fmt.Errorf("%w: message content must not exceed 4000 characters", ErrValidation)
	}
	return nil
}

// ValidateCountry verifies country code format if present.
func ValidateCountry(country *string) error {
	if country == nil {
		return nil
	}
	code := strings.TrimSpace(strings.ToUpper(*country))
	if code == "" {
		return nil
	}
	if code != "LOCAL" && code != "UN" && !isoCountryRegex.MatchString(code) {
		return fmt.Errorf("%w: invalid ISO 3166-1 alpha-2 country code", ErrValidation)
	}
	return nil
}

// ValidateID verifies required entity ID format.
func ValidateID(id, prefix string) error {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return fmt.Errorf("%w: id is required", ErrValidation)
	}
	if prefix != "" && !strings.HasPrefix(trimmed, prefix) {
		return fmt.Errorf("%w: invalid id prefix format", ErrValidation)
	}
	return nil
}
