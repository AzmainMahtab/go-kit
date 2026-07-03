// Package validator wraps github.com/go-playground/validator/v10.
//
// It provides a small domain-friendly interface so application code can validate
// DTOs without importing the playground package everywhere.
package validator

import (
	"fmt"
	"strings"

	playground "github.com/go-playground/validator/v10"
)

// Validator is the application-layer port.
type Validator interface {
	ValidateStruct(s any) error
}

// GoPlaygroundValidator implements Validator.
type GoPlaygroundValidator struct {
	v *playground.Validate
}

// New returns a preconfigured validator.
func New() *GoPlaygroundValidator {
	v := playground.New()
	return &GoPlaygroundValidator{v: v}
}

// ValidateStruct validates a struct and returns a ValidationError.
func (val *GoPlaygroundValidator) ValidateStruct(s any) error {
	if err := val.v.Struct(s); err != nil {
		return NewValidationError(err)
	}
	return nil
}

// ValidationError is a user-friendly validation error.
type ValidationError struct {
	Errors map[string]string `json:"errors"`
}

func (e *ValidationError) Error() string {
	var parts []string
	for field, msg := range e.Errors {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// NewValidationError converts playground errors into a map.
func NewValidationError(err error) *ValidationError {
	invalid, ok := err.(playground.ValidationErrors)
	if !ok {
		return &ValidationError{Errors: map[string]string{"_general": err.Error()}}
	}

	errors := make(map[string]string, len(invalid))
	for _, fe := range invalid {
		errors[strings.ToLower(fe.Field())] = formatFieldError(fe)
	}
	return &ValidationError{Errors: errors}
}

func formatFieldError(fe playground.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "invalid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "eqfield":
		return fmt.Sprintf("must match %s", fe.Param())
	default:
		return fmt.Sprintf("failed validation on %s", fe.Tag())
	}
}
