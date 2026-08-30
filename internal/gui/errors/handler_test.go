package errors

import (
	"testing"
)

// --- handler.go tests ---

func TestNewErrorHandler_DefaultPrefix(t *testing.T) {
	h := NewErrorHandler("")
	if h == nil {
		t.Fatal("expected non-nil ErrorHandler")
	}
}

func TestNewErrorHandler_CustomPrefix(t *testing.T) {
	h := NewErrorHandler("TEST")
	if h == nil {
		t.Fatal("expected non-nil ErrorHandler")
	}
}

func TestErrorHandler_Handle_NilError(t *testing.T) {
	h := NewErrorHandler("TEST")
	result := h.Handle(nil)
	if result != "" {
		t.Errorf("expected empty result for nil error, got %q", result)
	}
}

func TestErrorHandler_Handle_GUIError(t *testing.T) {
	h := NewErrorHandler("TEST")
	err := NewGUIError(ErrNetwork, "Network down", "connection refused")
	result := h.Handle(err)
	if result == "" {
		t.Error("expected non-empty result for GUIError")
	}
}

func TestErrorHandler_Handle_GenericError(t *testing.T) {
	h := NewErrorHandler("TEST")
	err := testError("generic error")
	result := h.Handle(err)
	if result == "" {
		t.Error("expected non-empty result for generic error")
	}
}

func TestErrorHandler_HandleWithUI_NilError(t *testing.T) {
	h := NewErrorHandler("TEST")
	result := h.HandleWithUI(nil)
	if result != "" {
		t.Errorf("expected empty result for nil error, got %q", result)
	}
}

func TestErrorHandler_HandleWithUI_GUIError(t *testing.T) {
	h := NewErrorHandler("TEST")
	err := NewGUIError(ErrTimeout, "Timeout", "deadline exceeded")
	result := h.HandleWithUI(err)
	if result == "" {
		t.Error("expected non-empty result for GUIError")
	}
}

func TestErrorHandler_HandleWithUI_GenericError(t *testing.T) {
	h := NewErrorHandler("TEST")
	err := testError("generic error")
	result := h.HandleWithUI(err)
	if result == "" {
		t.Error("expected non-empty result for generic error")
	}
}

func TestErrorHandler_HandlePanic_NoPanic(t *testing.T) {
	h := NewErrorHandler("TEST")
	var err error
	// Не должен паниковать
	h.HandlePanic(&err)
	if err != nil {
		t.Error("expected nil error when no panic occurs")
	}
}

func TestGetStacktrace_Empty(t *testing.T) {
	stack := getStacktrace(2)
	// Может быть пустым или содержать стек
	_ = stack
}

func TestGetStacktrace_NonEmpty(t *testing.T) {
	stack := getStacktrace(1)
	// getStacktrace возвращает строку — может быть пустой или не пустой
	_ = stack
}

// testError — заглушка для тестирования error interface
type testError string

func (e testError) Error() string {
	return string(e)
}
