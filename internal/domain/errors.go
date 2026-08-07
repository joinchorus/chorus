package domain

import "errors"

var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when attempting to create a duplicate resource.
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrValidation is returned when input validation fails.
	ErrValidation = errors.New("validation failed")

	// ErrUnauthorized is returned when authentication is missing or invalid.
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrForbidden is returned when the user lacks required permissions.
	ErrForbidden = errors.New("forbidden access")

	// ErrInternal is returned when an unexpected internal error occurs.
	ErrInternal = errors.New("internal server error")
)
