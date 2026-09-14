package errors

import (
	"testing"
)

// --- types.go tests ---

func TestErrorCode_Constants(t *testing.T) {
	if ErrNetwork != "network_error" {
		t.Errorf("expected 'network_error', got %q", ErrNetwork)
	}
	if ErrTimeout != "timeout" {
		t.Errorf("expected 'timeout', got %q", ErrTimeout)
	}
	if ErrPermission != "permission_denied" {
		t.Errorf("expected 'permission_denied', got %q", ErrPermission)
	}
	if ErrNotFound != "not_found" {
		t.Errorf("expected 'not_found', got %q", ErrNotFound)
	}
	if ErrInvalidInput != "invalid_input" {
		t.Errorf("expected 'invalid_input', got %q", ErrInvalidInput)
	}
	if ErrInternal != "internal_error" {
		t.Errorf("expected 'internal_error', got %q", ErrInternal)
	}
	if ErrUnknown != "unknown_error" {
		t.Errorf("expected 'unknown_error', got %q", ErrUnknown)
	}
}

func TestNewGUIError_Basic(t *testing.T) {
	err := NewGUIError(ErrNetwork, "Network down", "connection refused")
	if err == nil {
		t.Fatal("expected non-nil GUIError")
	}
	if err.Code != ErrNetwork {
		t.Errorf("expected code 'network_error', got %q", err.Code)
	}
	if err.Message != "Network down" {
		t.Errorf("expected message 'Network down', got %q", err.Message)
	}
	if err.Technical != "connection refused" {
		t.Errorf("expected technical 'connection refused', got %q", err.Technical)
	}
	if !err.Retryable {
		t.Error("expected Retryable=true for ErrNetwork")
	}
}

func TestNewGUIError_NonRetryable(t *testing.T) {
	err := NewGUIError(ErrNotFound, "Not found", "404")
	if err.Retryable {
		t.Error("expected Retryable=false for ErrNotFound")
	}
}

func TestNewGUIError_Unknown(t *testing.T) {
	err := NewGUIError(ErrUnknown, "Unknown", "???")
	if err.Retryable {
		t.Error("expected Retryable=false for ErrUnknown")
	}
}

func TestGUIError_Error(t *testing.T) {
	err := NewGUIError(ErrNetwork, "Network down", "connection refused")
	if err.Error() != "Network down" {
		t.Errorf("expected 'Network down', got %q", err.Error())
	}
}

func TestGUIError_WithSuggestion(t *testing.T) {
	err := NewGUIError(ErrNetwork, "Network down", "connection refused")
	err = err.WithSuggestion("Check your connection")
	if err.Suggestion != "Check your connection" {
		t.Errorf("expected suggestion 'Check your connection', got %q", err.Suggestion)
	}
}

func TestGUIError_Is(t *testing.T) {
	err1 := NewGUIError(ErrNetwork, "Network down", "connection refused")
	err2 := NewGUIError(ErrTimeout, "Timeout", "deadline")
	if err1.Is(err2) {
		t.Error("expected err1 not to match err2")
	}
	if !err1.Is(err1) {
		t.Error("expected err1 to match itself")
	}
}

func TestGUIError_GetCode(t *testing.T) {
	err := NewGUIError(ErrPermission, "Permission denied", "access denied")
	if err.GetCode() != ErrPermission {
		t.Errorf("expected 'permission_denied', got %q", err.GetCode())
	}
}

func TestGUIError_IsRetryable(t *testing.T) {
	err := NewGUIError(ErrNetwork, "Network down", "connection refused")
	if !err.IsRetryable() {
		t.Error("expected IsRetryable=true for ErrNetwork")
	}
}

func TestGUIError_IsNotRetryable(t *testing.T) {
	err := NewGUIError(ErrNotFound, "Not found", "404")
	if err.IsRetryable() {
		t.Error("expected IsRetryable=false for ErrNotFound")
	}
}

func TestGUIError_UserMessage_WithoutSuggestion(t *testing.T) {
	err := NewGUIError(ErrNotFound, "Not found", "404")
	msg := err.UserMessage()
	if msg != "Not found" {
		t.Errorf("expected 'Not found', got %q", msg)
	}
}

func TestGUIError_UserMessage_WithSuggestion(t *testing.T) {
	err := NewGUIError(ErrNetwork, "Network down", "connection refused")
	err = err.WithSuggestion("Check your connection")
	msg := err.UserMessage()
	expected := "Network down\n\n💡 Check your connection"
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}

func TestGUIError_FormatForUI_Network(t *testing.T) {
	err := NewGUIError(ErrNetwork, "Network down", "connection refused")
	formatted := err.FormatForUI()
	if formatted == "" {
		t.Error("expected non-empty formatted message")
	}
}

func TestGUIError_FormatForUI_Timeout(t *testing.T) {
	err := NewGUIError(ErrTimeout, "Timeout", "deadline exceeded")
	formatted := err.FormatForUI()
	if formatted == "" {
		t.Error("expected non-empty formatted message")
	}
}

func TestGUIError_FormatForUI_Permission(t *testing.T) {
	err := NewGUIError(ErrPermission, "Permission denied", "access denied")
	formatted := err.FormatForUI()
	if formatted == "" {
		t.Error("expected non-empty formatted message")
	}
}

func TestGUIError_FormatForUI_NotFound(t *testing.T) {
	err := NewGUIError(ErrNotFound, "Not found", "404")
	formatted := err.FormatForUI()
	if formatted == "" {
		t.Error("expected non-empty formatted message")
	}
}

func TestGUIError_FormatForUI_InvalidInput(t *testing.T) {
	err := NewGUIError(ErrInvalidInput, "Invalid input", "bad format")
	formatted := err.FormatForUI()
	if formatted == "" {
		t.Error("expected non-empty formatted message")
	}
}

func TestGUIError_FormatForUI_Internal(t *testing.T) {
	err := NewGUIError(ErrInternal, "Internal error", "crash")
	formatted := err.FormatForUI()
	if formatted == "" {
		t.Error("expected non-empty formatted message")
	}
}

func TestWrap_NonGUIError(t *testing.T) {
	err := Wrap(testError("some error"), ErrNetwork, "Network down")
	if err == nil {
		t.Fatal("expected non-nil GUIError")
	}
	if err.Code != ErrNetwork {
		t.Errorf("expected code 'network_error', got %q", err.Code)
	}
}

func TestWrap_GUIError(t *testing.T) {
	original := NewGUIError(ErrNetwork, "Network down", "connection refused")
	wrapped := Wrap(original, ErrTimeout, "different")
	// Wrap должен вернуть оригинальный GUIError
	if wrapped != original {
		t.Error("expected Wrap to return original GUIError")
	}
}

func TestWrapWithSuggestion(t *testing.T) {
	wrapped := WrapWithSuggestion(testError("some error"), ErrNetwork, "Network down", "Check connection")
	if wrapped == nil {
		t.Fatal("expected non-nil GUIError")
	}
	if wrapped.Suggestion != "Check connection" {
		t.Errorf("expected suggestion 'Check connection', got %q", wrapped.Suggestion)
	}
}

func TestCommonErrorMessages_AllKeys(t *testing.T) {
	if CommonErrorMessages[ErrNetwork] != "Ошибка подключения к сети" {
		t.Error("expected network error message")
	}
	if CommonErrorMessages[ErrTimeout] != "Превышено время ожидания" {
		t.Error("expected timeout error message")
	}
	if CommonErrorMessages[ErrPermission] != "Недостаточно прав для выполнения операции" {
		t.Error("expected permission error message")
	}
	if CommonErrorMessages[ErrNotFound] != "Запрашиваемый ресурс не найден" {
		t.Error("expected not found error message")
	}
	if CommonErrorMessages[ErrInvalidInput] != "Неверный формат ввода" {
		t.Error("expected invalid input error message")
	}
	if CommonErrorMessages[ErrInternal] != "Внутренняя ошибка приложения" {
		t.Error("expected internal error message")
	}
}
