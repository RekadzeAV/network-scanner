package errors

import (
	"fmt"
	"strings"
)

// ErrorCode определяет тип ошибки.
type ErrorCode string

const (
	ErrNetwork      ErrorCode = "network_error"
	ErrTimeout      ErrorCode = "timeout"
	ErrPermission   ErrorCode = "permission_denied"
	ErrNotFound     ErrorCode = "not_found"
	ErrInvalidInput ErrorCode = "invalid_input"
	ErrInternal     ErrorCode = "internal_error"
	ErrUnknown      ErrorCode = "unknown_error"
)

// GUIError описывает ошибку с пользовательским сообщением.
type GUIError struct {
	Code       ErrorCode // Код ошибки
	Message    string    // Пользовательское сообщение
	Technical  string    // Техническая деталь (для логов)
	Retryable  bool      // Можно ли попробовать снова
	Suggestion string    // Рекомендация пользователю
}

// Error реализует интерфейс error.
func (e *GUIError) Error() string {
	return e.Message
}

// NewGUIError создает новую GUI-ошибку.
func NewGUIError(code ErrorCode, userMsg, techMsg string) *GUIError {
	return &GUIError{
		Code:      code,
		Message:   userMsg,
		Technical: techMsg,
		Retryable: code == ErrNetwork || code == ErrTimeout,
	}
}

// WithSuggestion добавляет рекомендацию пользователю.
func (e *GUIError) WithSuggestion(suggestion string) *GUIError {
	e.Suggestion = suggestion
	return e
}

// Is возвращает true, если ошибка совпадает по коду.
func (e *GUIError) Is(target error) bool {
	t, ok := target.(*GUIError)
	return ok && e.Code == t.Code
}

// GetCode возвращает код ошибки.
func (e *GUIError) GetCode() ErrorCode {
	return e.Code
}

// IsRetryable возвращает true, если ошибку можно повторить.
func (e *GUIError) IsRetryable() bool {
	return e.Retryable
}

// UserMessage возвращает сообщение для пользователя.
func (e *GUIError) UserMessage() string {
	msg := e.Message
	if e.Suggestion != "" {
		msg += "\n\n💡 " + e.Suggestion
	}
	return msg
}

// FormatForUI форматирует ошибку для отображения в GUI.
func (e *GUIError) FormatForUI() string {
	icon := "⚠️"
	switch e.Code {
	case ErrNetwork:
		icon = "🌐"
	case ErrTimeout:
		icon = "⏱️"
	case ErrPermission:
		icon = "🔒"
	case ErrNotFound:
		icon = "🔍"
	case ErrInvalidInput:
		icon = "❌"
	case ErrInternal:
		icon = "🐛"
	}
	return fmt.Sprintf("%s **%s**: %s", icon, strings.ToUpper(string(e.Code)), e.UserMessage())
}

// Wrap оборачивает ошибку в GUIError.
func Wrap(err error, code ErrorCode, userMsg string) *GUIError {
	if guiErr, ok := err.(*GUIError); ok {
		return guiErr
	}
	techMsg := err.Error()
	return NewGUIError(code, userMsg, techMsg)
}

// WrapWithSuggestion оборачивает ошибку с рекомендацией.
func WrapWithSuggestion(err error, code ErrorCode, userMsg, suggestion string) *GUIError {
	return Wrap(err, code, userMsg).WithSuggestion(suggestion)
}

// CommonErrorMessages возвращает карту стандартных сообщений.
var CommonErrorMessages = map[ErrorCode]string{
	ErrNetwork:      "Ошибка подключения к сети",
	ErrTimeout:      "Превышено время ожидания",
	ErrPermission:   "Недостаточно прав для выполнения операции",
	ErrNotFound:     "Запрашиваемый ресурс не найден",
	ErrInvalidInput: "Неверный формат ввода",
	ErrInternal:     "Внутренняя ошибка приложения",
}
