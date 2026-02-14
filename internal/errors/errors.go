// Package errors provides centralized error types and handling for ClawTeam
package errors

import (
	"errors"
	"fmt"
)

// Standard error variables
var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when a resource already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput is returned for invalid user input
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when authorization fails
	ErrForbidden = errors.New("forbidden")

	// ErrInternal is returned for internal server errors
	ErrInternal = errors.New("internal server error")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timed out")

	// ErrRateLimited is returned when rate limit is exceeded
	ErrRateLimited = errors.New("rate limit exceeded")
)

// AppError represents an application error with additional context
type AppError struct {
	Code    string
	Message string
	Err     error
	Details map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// Common error constructors
func NotFound(resource string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
		Err:     ErrNotFound,
	}
}

func AlreadyExists(resource string) *AppError {
	return &AppError{
		Code:    "ALREADY_EXISTS",
		Message: fmt.Sprintf("%s already exists", resource),
		Err:     ErrAlreadyExists,
	}
}

func InvalidInput(message string) *AppError {
	return &AppError{
		Code:    "INVALID_INPUT",
		Message: message,
		Err:     ErrInvalidInput,
	}
}

func Unauthorized() *AppError {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: "authentication required",
		Err:     ErrUnauthorized,
	}
}

func Forbidden(action string) *AppError {
	return &AppError{
		Code:    "FORBIDDEN",
		Message: fmt.Sprintf("not allowed to %s", action),
		Err:     ErrForbidden,
	}
}

func Internal(err error) *AppError {
	return &AppError{
		Code:    "INTERNAL",
		Message: "internal server error",
		Err:     err,
	}
}

func Timeout(operation string) *AppError {
	return &AppError{
		Code:    "TIMEOUT",
		Message: fmt.Sprintf("%s timed out", operation),
		Err:     ErrTimeout,
	}
}

// Is checks if an error is of a specific type
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As checks if an error can be cast to a specific type
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if an error is an already exists error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInput checks if an error is an invalid input error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if an error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}
