package apperror

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"sync"
)

// Error — прикладная ошибка: код, операция, человекочитаемое сообщение,
// причина и набор контекстных атрибутов.
type Error struct {
	Code    Code
	Op      string
	Message string
	Cause   error

	attrs map[string]any
}

// New — ошибка с кодом и сообщением.
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Newf — ошибка с форматированным сообщением.
func Newf(code Code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// WithOp — устанавливает имя операции ("scanner.scan", "db.query").
func (e *Error) WithOp(op string) *Error {
	if e == nil {
		return nil
	}
	e.Op = op
	return e
}

// Wrap — оборачивает ошибку, сохраняя её в Cause и унаследуя код,
// если сама ошибка не имеет кода.
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{Code: Detect(err), Message: message, Cause: err}
}

// WrapCode — оборачивает ошибку с явным кодом.
func WrapCode(code Code, err error, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{Code: code, Message: message, Cause: err}
}

// From — нормализует любую ошибку в *Error (nil остаётся nil).
//
// Уже нормализованные ошибки возвращаются как есть, чтобы не плодить слои.
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return &Error{Code: Detect(err), Message: err.Error(), Cause: err}
}

// Error — реализует error. Формат: "op: message: cause".
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	parts := make([]string, 0, 3)
	if e.Op != "" {
		parts = append(parts, e.Op)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.Cause != nil {
		parts = append(parts, e.Cause.Error())
	}
	if len(parts) == 0 {
		if e.Code != "" {
			return string(e.Code)
		}
		return string(CodeUnknown)
	}
	return joinParts(parts)
}

// Unwrap — причина ошибки (для errors.Is/As).
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// With — добавляет контекстный атрибут для логов и диагностики.
func (e *Error) With(key string, value any) *Error {
	if e == nil {
		return nil
	}
	if e.attrs == nil {
		e.attrs = map[string]any{}
	}
	e.attrs[key] = value
	return e
}

// Attr — значение контекстного атрибута.
func (e *Error) Attr(key string) (any, bool) {
	if e == nil || e.attrs == nil {
		return nil, false
	}
	value, ok := e.attrs[key]
	return value, ok
}

// Attrs — копия контекстных атрибутов.
func (e *Error) Attrs() map[string]any {
	if e == nil || len(e.attrs) == 0 {
		return nil
	}
	out := make(map[string]any, len(e.attrs))
	for k, v := range e.attrs {
		out[k] = v
	}
	return out
}

// Of — код ошибки из цепочки обёрток (CodeUnknown если не найден).
func Of(err error) Code {
	if err == nil {
		return ""
	}
	var appErr *Error
	if errors.As(err, &appErr) && appErr.Code != "" {
		return appErr.Code
	}
	return Detect(err)
}

// Is — проверка кода в цепочке (аналог errors.Is для кодов).
func Is(err error, code Code) bool {
	return Of(err) == code
}

// Detector — пользовательский маппинг ошибки в код.
type Detector func(err error) (Code, bool)

var detectorRegistry = struct {
	mu        sync.RWMutex
	detectors []Detector
}{}

// RegisterDetector — добавляет маппинг для доменных ошибок (БД, драйверы).
func RegisterDetector(detector Detector) {
	if detector == nil {
		return
	}
	detectorRegistry.mu.Lock()
	defer detectorRegistry.mu.Unlock()
	detectorRegistry.detectors = append(detectorRegistry.detectors, detector)
}

// ResetDetectors — убирает пользовательские детекторы (тесты).
func ResetDetectors() {
	detectorRegistry.mu.Lock()
	defer detectorRegistry.mu.Unlock()
	detectorRegistry.detectors = nil
}

// Detect — определяет код ошибки по стандартным признакам Go.
func Detect(err error) Code {
	if err == nil {
		return CodeUnknown
	}

	var appErr *Error
	if errors.As(err, &appErr) && appErr.Code != "" {
		return appErr.Code
	}

	switch {
	case errors.Is(err, context.Canceled):
		return CodeCanceled
	case errors.Is(err, context.DeadlineExceeded):
		return CodeTimeout
	case errors.Is(err, os.ErrNotExist), errors.Is(err, fs.ErrNotExist):
		return CodeNotFound
	case errors.Is(err, os.ErrExist), errors.Is(err, fs.ErrExist):
		return CodeAlreadyExists
	case errors.Is(err, os.ErrPermission), errors.Is(err, fs.ErrPermission):
		return CodePermissionDenied
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return CodeTimeout
		}
		return CodeUnavailable
	}

	detectorRegistry.mu.RLock()
	detectors := append([]Detector{}, detectorRegistry.detectors...)
	detectorRegistry.mu.RUnlock()

	for _, detect := range detectors {
		if code, ok := detect(err); ok {
			return code
		}
	}

	return CodeUnknown
}

func joinParts(parts []string) string {
	out := parts[0]
	for _, part := range parts[1:] {
		out += ": " + part
	}
	return out
}
