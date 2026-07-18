package errors

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// ErrorHandler обрабатывает ошибки с логированием и пользовательскими сообщениями.
type ErrorHandler struct {
	prefix string
}

// NewErrorHandler создает новый обработчик ошибок.
func NewErrorHandler(prefix string) *ErrorHandler {
	if prefix == "" {
		prefix = "GUI"
	}
	return &ErrorHandler{prefix: prefix}
}

// Handle обрабатывает ошибку: логирует и возвращает GUI-сообщение.
func (h *ErrorHandler) Handle(err error) string {
	if err == nil {
		return ""
	}

	// Извлекаем стек вызовов
	stack := getStacktrace(2)

	// Логируем техническую деталь
	var techMsg string
	if guiErr, ok := err.(*GUIError); ok {
		techMsg = guiErr.Technical
		fmt.Fprintf(os.Stderr, "[%s] [%s] %s | %s\n", h.prefix, guiErr.Code, guiErr.Message, techMsg)
	} else {
		techMsg = err.Error()
		fmt.Fprintf(os.Stderr, "[%s] [unknown] %s\n", h.prefix, techMsg)
	}

	// Логируем стек
	if stack != "" {
		fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", stack)
	}

	// Возвращаем пользовательское сообщение
	if guiErr, ok := err.(*GUIError); ok {
		return guiErr.UserMessage()
	}
	return fmt.Sprintf("Произошла ошибка: %s", err.Error())
}

// HandleWithUI обрабатывает ошибку и возвращает форматированное сообщение для UI.
func (h *ErrorHandler) HandleWithUI(err error) string {
	if err == nil {
		return ""
	}

	h.Handle(err)

	if guiErr, ok := err.(*GUIError); ok {
		return guiErr.FormatForUI()
	}
	return fmt.Sprintf("⚠️ **Ошибка**: %s", err.Error())
}

// HandlePanic обрабатывает panic и возвращает GUIError.
func (h *ErrorHandler) HandlePanic(err *error) {
	if r := recover(); r != nil {
		*err = &GUIError{
			Code:       ErrInternal,
			Message:    "Произошла непредвиденная ошибка",
			Technical:  fmt.Sprintf("panic: %v", r),
			Retryable:  false,
			Suggestion: "Сохраните данные и перезапустите приложение",
		}
		fmt.Fprintf(os.Stderr, "[%s] Panic recovered: %v\n", h.prefix, r)
		fmt.Fprintf(os.Stderr, "Stack: %s\n", getStacktrace(3))
	}
}

// getStacktrace возвращает стек вызовов.
func getStacktrace(skip int) string {
	var sb strings.Builder
	pc := make([]uintptr, 20)
	n := runtime.Callers(skip, pc)
	if n == 0 {
		return ""
	}
	pc = pc[:n]
	frames := runtime.CallersFrames(pc)
	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("  %s\n    %s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return sb.String()
}
