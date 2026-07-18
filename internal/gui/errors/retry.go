package errors

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// RetryConfig конфигурация повторных попыток.
type RetryConfig struct {
	MaxAttempts   int           // Максимальное количество попыток
	BaseDelay     time.Duration // Начальная задержка
	MaxDelay      time.Duration // Максимальная задержка
	BackoffFactor float64       // Коэффициент увеличения задержки
	Jitter        bool          // Добавлять ли случайную составляющую
}

// DefaultRetryConfig стандартная конфигурация retry.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:   3,
	BaseDelay:     100 * time.Millisecond,
	MaxDelay:      2 * time.Second,
	BackoffFactor: 2.0,
	Jitter:        true,
}

// ExecuteWithRetry выполняет функцию с повторными попытками.
func ExecuteWithRetry(ctx context.Context, fn func() error, config RetryConfig) error {
	var lastErr error
	delay := config.BaseDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			// Exponential backoff с jitter
			if config.Jitter {
				jitter := time.Duration(rand.Int63n(int64(delay) / 2))
				delay += jitter
			}
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Если ошибка не retryable, не повторяем
		if guiErr, ok := lastErr.(*GUIError); ok && !guiErr.IsRetryable() {
			break
		}
	}

	return lastErr
}

// ExecuteWithRetryAndCallback выполняет функцию с retry и callback для каждой попытки.
func ExecuteWithRetryAndCallback(ctx context.Context, fn func() error, onRetry func(attempt int, err error), config RetryConfig) error {
	var lastErr error
	delay := config.BaseDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			if config.Jitter {
				jitter := time.Duration(rand.Int63n(int64(delay) / 2))
				delay += jitter
			}
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if onRetry != nil {
			onRetry(attempt, lastErr)
		}

		if guiErr, ok := lastErr.(*GUIError); ok && !guiErr.IsRetryable() {
			break
		}
	}

	return lastErr
}

// SafeExecute выполняет функцию в recover и возвращает ошибку вместо panic.
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &GUIError{
				Code:       ErrInternal,
				Message:    "Произошла непредвиденная ошибка",
				Technical:  fmt.Sprintf("panic: %v", r),
				Retryable:  false,
				Suggestion: "Попробуйте перезапустить приложение",
			}
		}
	}()
	return fn()
}
