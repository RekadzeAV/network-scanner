package errors

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- retry.go tests ---

func TestDefaultRetryConfig_Values(t *testing.T) {
	if DefaultRetryConfig.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", DefaultRetryConfig.MaxAttempts)
	}
	if DefaultRetryConfig.BaseDelay != 100*time.Millisecond {
		t.Errorf("expected BaseDelay=100ms, got %v", DefaultRetryConfig.BaseDelay)
	}
	if DefaultRetryConfig.MaxDelay != 2*time.Second {
		t.Errorf("expected MaxDelay=2s, got %v", DefaultRetryConfig.MaxDelay)
	}
	if DefaultRetryConfig.BackoffFactor != 2.0 {
		t.Errorf("expected BackoffFactor=2.0, got %v", DefaultRetryConfig.BackoffFactor)
	}
	if !DefaultRetryConfig.Jitter {
		t.Error("expected Jitter=true")
	}
}

func TestExecuteWithRetry_Success(t *testing.T) {
	attempts := 0
	ctx := context.Background()
	err := ExecuteWithRetry(ctx, func() error {
		attempts++
		return nil
	}, DefaultRetryConfig)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestExecuteWithRetry_FailsOnceThenSucceeds(t *testing.T) {
	attempts := 0
	ctx := context.Background()
	err := ExecuteWithRetry(ctx, func() error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	}, DefaultRetryConfig)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestExecuteWithRetry_AllFail(t *testing.T) {
	attempts := 0
	ctx := context.Background()
	err := ExecuteWithRetry(ctx, func() error {
		attempts++
		return errors.New("permanent error")
	}, RetryConfig{
		MaxAttempts:  3,
		BaseDelay:    1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		BackoffFactor: 1.0,
		Jitter:       false,
	})
	if err == nil {
		t.Error("expected non-nil error")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestExecuteWithRetry_ContextCancelled(t *testing.T) {
	attempts := 0
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем сразу
	err := ExecuteWithRetry(ctx, func() error {
		attempts++
		return errors.New("error")
	}, RetryConfig{
		MaxAttempts:  5,
		BaseDelay:    1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		BackoffFactor: 1.0,
		Jitter:       false,
	})
	if err == nil {
		t.Error("expected context error")
	}
	if attempts < 1 {
		t.Errorf("expected at least 1 attempt, got %d", attempts)
	}
}

func TestExecuteWithRetry_NonRetryableError(t *testing.T) {
	attempts := 0
	ctx := context.Background()
	err := ExecuteWithRetry(ctx, func() error {
		attempts++
		return NewGUIError(ErrNotFound, "Not found", "404")
	}, DefaultRetryConfig)
	if err == nil {
		t.Error("expected non-nil error")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestExecuteWithRetryAndCallback_Success(t *testing.T) {
	attempts := 0
	retryCalled := 0
	ctx := context.Background()
	err := ExecuteWithRetryAndCallback(ctx, func() error {
		attempts++
		return nil
	}, func(attempt int, err error) {
		retryCalled++
		_ = attempt
		_ = err
	}, DefaultRetryConfig)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if retryCalled != 0 {
		t.Errorf("expected onRetry not called, got %d calls", retryCalled)
	}
}

func TestExecuteWithRetryAndCallback_OnRetryCalled(t *testing.T) {
	attempts := 0
	retryCalled := 0
	ctx := context.Background()
	err := ExecuteWithRetryAndCallback(ctx, func() error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	}, func(attempt int, err error) {
		retryCalled++
		if attempt != 0 {
			t.Errorf("expected attempt 0, got %d", attempt)
		}
		_ = err
	}, RetryConfig{
		MaxAttempts:  3,
		BaseDelay:    1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		BackoffFactor: 1.0,
		Jitter:       false,
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if retryCalled != 1 {
		t.Errorf("expected onRetry called 1 time, got %d", retryCalled)
	}
}

func TestExecuteWithRetryAndCallback_NilCallback(t *testing.T) {
	attempts := 0
	ctx := context.Background()
	err := ExecuteWithRetryAndCallback(ctx, func() error {
		attempts++
		return errors.New("error")
	}, nil, RetryConfig{
		MaxAttempts:  2,
		BaseDelay:    1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		BackoffFactor: 1.0,
		Jitter:       false,
	})
	if err == nil {
		t.Error("expected non-nil error")
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestSafeExecute_Success(t *testing.T) {
	err := SafeExecute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestSafeExecute_ReturnsError(t *testing.T) {
	expectedErr := errors.New("function error")
	err := SafeExecute(func() error {
		return expectedErr
	})
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestSafeExecute_Panic(t *testing.T) {
	err := SafeExecute(func() error {
		panic("test panic")
	})
	if err == nil {
		t.Error("expected non-nil error for panic")
	}
	if guiErr, ok := err.(*GUIError); ok {
		if guiErr.Code != ErrInternal {
			t.Errorf("expected ErrInternal, got %q", guiErr.Code)
		}
		if guiErr.Retryable {
			t.Error("expected Retryable=false for panic")
		}
	} else {
		t.Error("expected error to be *GUIError")
	}
}

func TestSafeExecute_PanicWithMessage(t *testing.T) {
	err := SafeExecute(func() error {
		panic("something went wrong")
	})
	if err == nil {
		t.Error("expected non-nil error")
	}
	if guiErr, ok := err.(*GUIError); ok {
		if guiErr.Suggestion != "Попробуйте перезапустить приложение" {
			t.Errorf("expected suggestion, got %q", guiErr.Suggestion)
		}
	}
}
