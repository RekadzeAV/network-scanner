package errors

import (
	"errors"
	"fmt"
	"time"
)

// Common error types
var (
	ErrNotFound       = errors.New("resource not found")
	ErrTimeout        = errors.New("operation timed out")
	ErrPermission     = errors.New("permission denied")
	ErrInvalidInput   = errors.New("invalid input")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrConflict       = errors.New("conflict detected")
	ErrInternal       = errors.New("internal server error")
)

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s '%s' not found", e.Resource, e.ID)
}

func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Operation string
	Duration  time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("operation '%s' timed out after %v", e.Operation, e.Duration)
}

func (e *TimeoutError) Unwrap() error {
	return ErrTimeout
}

// PermissionError represents a permission denied error
type PermissionError struct {
	User    string
	Resource string
	Action  string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("user '%s' not allowed to %s '%s'", e.User, e.Action, e.Resource)
}

func (e *PermissionError) Unwrap() error {
	return ErrPermission
}

// InvalidInputError represents an invalid input error
type InvalidInputError struct {
	Field   string
	Message string
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input for field '%s': %s", e.Field, e.Message)
}

func (e *InvalidInputError) Unwrap() error {
	return ErrInvalidInput
}

// Helper functions

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource, id string) error {
	return &NotFoundError{Resource: resource, ID: id}
}

// IsNotFound checks if error is a NotFoundError
func IsNotFound(err error) bool {
	var notFound *NotFoundError
	return errors.As(err, &notFound) || errors.Is(err, ErrNotFound)
}

// NewTimeoutError creates a new TimeoutError
func NewTimeoutError(operation string, duration time.Duration) error {
	return &TimeoutError{Operation: operation, Duration: duration}
}

// IsTimeout checks if error is a TimeoutError
func IsTimeout(err error) bool {
	var timeout *TimeoutError
	return errors.As(err, &timeout) || errors.Is(err, ErrTimeout)
}

// NewPermissionError creates a new PermissionError
func NewPermissionError(user, resource, action string) error {
	return &PermissionError{User: user, Resource: resource, Action: action}
}

// IsPermission checks if error is a PermissionError
func IsPermission(err error) bool {
	var perm *PermissionError
	return errors.As(err, &perm) || errors.Is(err, ErrPermission)
}

// NewInvalidInputError creates a new InvalidInputError
func NewInvalidInputError(field, message string) error {
	return &InvalidInputError{Field: field, Message: message}
}

// IsInvalidInput checks if error is an InvalidInputError
func IsInvalidInput(err error) bool {
	var input *InvalidInputError
	return errors.As(err, &input) || errors.Is(err, ErrInvalidInput)
}

// WrapError wraps an error with context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapErrorf wraps an error with context and formatting
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", msg, err)
}

// SafeError returns error string or "no error" if nil
func SafeError(err error) string {
	if err == nil {
		return "no error"
	}
	return err.Error()
}

// MultiError represents multiple errors
type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	return fmt.Sprintf("%d errors occurred", len(e.Errors))
}

func (e *MultiError) Unwrap() []error {
	return e.Errors
}

// Append adds an error to MultiError
func (e *MultiError) Append(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// NewMultiError creates a new MultiError
func NewMultiError(errs ...error) *MultiError {
	me := &MultiError{}
	for _, err := range errs {
		me.Append(err)
	}
	return me
}
