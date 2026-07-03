// Package apperrors provides a small, domain-friendly error taxonomy.
//
// In Go, errors are values. We use sentinel errors + a typed AppError so HTTP
// handlers can map domain errors to status codes without importing every
// package's internal error type.
package apperrors

import (
	"errors"
	"fmt"
)

// Common sentinel errors used across modules.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource already exists")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInternal     = errors.New("internal error")
)

// AppError wraps a domain error with an HTTP-friendly code and status.
type AppError struct {
	Err        error  // underlying error
	Code       string // machine-readable code, e.g. "USER_ALREADY_EXISTS"
	StatusCode int    // HTTP status code
	Message    string // human-readable message
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *AppError) Unwrap() error { return e.Err }

// New wraps an error with an HTTP status and code.
func New(err error, code string, statusCode int) *AppError {
	return &AppError{Err: err, Code: code, StatusCode: statusCode, Message: err.Error()}
}

// Newf is a convenience constructor with formatting.
func Newf(statusCode int, code string, format string, args ...any) *AppError {
	msg := fmt.Sprintf(format, args...)
	return &AppError{Err: errors.New(msg), Code: code, StatusCode: statusCode, Message: msg}
}

// IsNotFound reports whether err is ErrNotFound.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsConflict reports whether err is ErrConflict.
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }

// IsInvalidInput reports whether err is ErrInvalidInput.
func IsInvalidInput(err error) bool { return errors.Is(err, ErrInvalidInput) }

// IsUnauthorized reports whether err is ErrUnauthorized.
func IsUnauthorized(err error) bool { return errors.Is(err, ErrUnauthorized) }
