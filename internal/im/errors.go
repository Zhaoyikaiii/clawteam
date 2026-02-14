// Package im provides IM-specific error definitions
package im

import (
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeStorage   ErrorType = "storage"
	ErrorTypeAuth      ErrorType = "auth"
	ErrorTypeNotFound  ErrorType = "not_found"
)

// Error represents an IM service error
type Error struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Code    int       `json:"code"`
	cause   error     `json:"-"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.cause
}

// Predefined errors
var (
	ErrInvalidMessage  = &Error{Type: ErrorTypeValidation, Message: "invalid message", Code: 400}
	ErrChannelNotFound = &Error{Type: ErrorTypeNotFound, Message: "channel not found", Code: 404}
	ErrUserNotFound    = &Error{Type: ErrorTypeNotFound, Message: "user not found", Code: 404}
	ErrUnauthorized    = &Error{Type: ErrorTypeAuth, Message: "unauthorized", Code: 401}
	ErrForbidden      = &Error{Type: ErrorTypeAuth, Message: "forbidden", Code: 403}
)

// Helper functions for creating errors

// ErrInvalidInput creates a validation error
func ErrInvalidInput(msg string) *Error {
	return &Error{
		Type:    ErrorTypeValidation,
		Message: msg,
		Code:    400,
	}
}

// ErrStorageFailed creates a storage error
func ErrStorageFailed(msg string, args ...interface{}) *Error {
	return &Error{
		Type:    ErrorTypeStorage,
		Message: fmt.Sprintf(msg, args...),
		Code:    500,
	}
}

// ErrNotFound creates a not found error
func ErrNotFound(resource, id string) *Error {
	return &Error{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s '%s' not found", resource, id),
		Code:    404,
	}
}

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeValidation
	}
	return false
}

// IsStorageError checks if error is a storage error
func IsStorageError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeStorage
	}
	return false
}

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeNotFound
	}
	return false
}
