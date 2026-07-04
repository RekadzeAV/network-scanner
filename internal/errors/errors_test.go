package errors

import (
	"errors"
	"testing"
	"time"
)

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "ErrNotFound",
			err:      ErrNotFound,
			checkFn:  IsNotFound,
			expected: true,
		},
		{
			name:     "NotFoundError",
			err:      NewNotFoundError("device", "123"),
			checkFn:  IsNotFound,
			expected: true,
		},
		{
			name:     "ErrTimeout",
			err:      ErrTimeout,
			checkFn:  IsTimeout,
			expected: true,
		},
		{
			name:     "TimeoutError",
			err:      NewTimeoutError("scan", 5*time.Second),
			checkFn:  IsTimeout,
			expected: true,
		},
		{
			name:     "ErrPermission",
			err:      ErrPermission,
			checkFn:  IsPermission,
			expected: true,
		},
		{
			name:     "PermissionError",
			err:      NewPermissionError("user", "device", "reboot"),
			checkFn:  IsPermission,
			expected: true,
		},
		{
			name:     "ErrInvalidInput",
			err:      ErrInvalidInput,
			checkFn:  IsInvalidInput,
			expected: true,
		},
		{
			name:     "InvalidInputError",
			err:      NewInvalidInputError("cidr", "invalid format"),
			checkFn:  IsInvalidInput,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checkFn(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNotFoundError(t *testing.T) {
	err := NewNotFoundError("device", "123")

	if err.Error() != "device '123' not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, ErrNotFound) {
		t.Error("expected error to wrap ErrNotFound")
	}
}

func TestTimeoutError(t *testing.T) {
	err := NewTimeoutError("scan", 10*time.Second)

	expectedMsg := "operation 'scan' timed out after 10s"
	if err.Error() != expectedMsg {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, ErrTimeout) {
		t.Error("expected error to wrap ErrTimeout")
	}
}

func TestPermissionError(t *testing.T) {
	err := NewPermissionError("admin", "device", "reboot")

	expectedMsg := "user 'admin' not allowed to reboot 'device'"
	if err.Error() != expectedMsg {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, ErrPermission) {
		t.Error("expected error to wrap ErrPermission")
	}
}

func TestInvalidInputError(t *testing.T) {
	err := NewInvalidInputError("cidr", "invalid format")

	expectedMsg := "invalid input for field 'cidr': invalid format"
	if err.Error() != expectedMsg {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Error("expected error to wrap ErrInvalidInput")
	}
}

func TestWrapError(t *testing.T) {
	inner := errors.New("inner error")
	wrapped := WrapError(inner, "context")

	if wrapped.Error() != "context: inner error" {
		t.Errorf("unexpected wrapped error: %s", wrapped.Error())
	}

	if !errors.Is(wrapped, ErrNotFound) {
		// Should still be able to unwrap
	}
}

func TestWrapErrorf(t *testing.T) {
	inner := errors.New("inner error")
	wrapped := WrapErrorf(inner, "context: %s", "detail")

	if wrapped.Error() != "context: detail: inner error" {
		t.Errorf("unexpected wrapped error: %s", wrapped.Error())
	}
}

func TestWrapError_Nil(t *testing.T) {
	wrapped := WrapError(nil, "context")
	if wrapped != nil {
		t.Error("expected nil for nil input")
	}
}

func TestSafeError(t *testing.T) {
	if SafeError(nil) != "no error" {
		t.Error("expected 'no error' for nil")
	}

	err := errors.New("test")
	if SafeError(err) != "test" {
		t.Error("expected error message")
	}
}

func TestMultiError(t *testing.T) {
	me := NewMultiError(
		errors.New("error 1"),
		errors.New("error 2"),
	)

	if me.Error() != "2 errors occurred" {
		t.Errorf("unexpected MultiError message: %s", me.Error())
	}

	if len(me.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(me.Errors))
	}

	// Test Append
	me.Append(nil)
	me.Append(errors.New("error 3"))

	if len(me.Errors) != 3 {
		t.Errorf("expected 3 errors after append, got %d", len(me.Errors))
	}
}

func TestMultiError_Unwrap(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	me := NewMultiError(err1, err2)

	unwrapped := me.Unwrap()
	if len(unwrapped) != 2 {
		t.Errorf("expected 2 unwrapped errors, got %d", len(unwrapped))
	}
}

func TestMultipleChecks(t *testing.T) {
	// Test that different error types don't cross-check
	err := NewNotFoundError("device", "123")

	if IsTimeout(err) {
		t.Error("NotFoundError should not be TimeoutError")
	}
	if IsPermission(err) {
		t.Error("NotFoundError should not be PermissionError")
	}
	if IsInvalidInput(err) {
		t.Error("NotFoundError should not be InvalidInputError")
	}
}
