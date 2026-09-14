package apperror

import (
	"errors"
	"fmt"
	"sort"
)

// GenericClientMessage — сообщение для клиента при внутренних ошибках.
const GenericClientMessage = "внутренняя ошибка, попробуйте позже"

// Logger — минимальный интерфейс структурированного логгера.
//
// Совместим с *slog.Logger через тонкую обёртку, а также с log/slog, zap
// и собственным логгером приложения.
type Logger interface {
	Error(msg string, keysAndValues ...any)
}

// Chain — цепочка обёрток ошибки (включая саму ошибку), не глубже limit.
func Chain(err error, limit int) []error {
	if err == nil {
		return nil
	}
	if limit <= 0 {
		limit = 16
	}

	out := []error{err}
	current := err
	for len(out) < limit {
		current = errors.Unwrap(current)
		if current == nil {
			break
		}
		out = append(out, current)
	}
	return out
}

// LogAttrs — пары ключ/значение для структурированного логирования ошибки:
// код, операция, причина, глубина цепочки и контекстные атрибуты.
//
// Переданные keysAndValues добавляются в конец, поэтому могут переопределить
// значения по умолчанию.
func LogAttrs(err error, keysAndValues ...any) []any {
	if err == nil {
		return append([]any{}, keysAndValues...)
	}

	attrs := []any{"error", err.Error(), "code", string(Of(err))}

	appErr := From(err)
	if appErr.Op != "" {
		attrs = append(attrs, "op", appErr.Op)
	}
	if appErr.Cause != nil {
		attrs = append(attrs, "cause", appErr.Cause.Error())
	}
	if depth := len(Chain(err, 0)); depth > 1 {
		attrs = append(attrs, "depth", depth)
	}

	for _, key := range sortedAttrKeys(appErr.attrs) {
		attrs = append(attrs, key, appErr.attrs[key])
	}

	return append(attrs, keysAndValues...)
}

// LogError — логирует ошибку с контекстом; nil игнорируется.
func LogError(logger Logger, err error, keysAndValues ...any) {
	if err == nil || logger == nil {
		return
	}
	logger.Error(err.Error(), LogAttrs(err, keysAndValues...)...)
}

// ToClient — безопасное сообщение для внешнего клиента.
//
// Внутренние ошибки (internal/unknown/data_loss) не раскрывают детали,
// их следует искать в логах по коду.
func ToClient(err error) string {
	if err == nil {
		return ""
	}

	code := Of(err)
	switch code {
	case CodeInternal, CodeUnknown, CodeDataLoss, "":
		return GenericClientMessage
	}

	appErr := From(err)
	if appErr.Message != "" {
		return appErr.Message
	}
	return GenericClientMessage
}

// Describe — строка "code: message" для CLI/GUI вывода.
func Describe(err error) string {
	if err == nil {
		return ""
	}
	code := Of(err)
	if code == "" {
		code = CodeUnknown
	}
	return fmt.Sprintf("%s: %s", code, err.Error())
}

func sortedAttrKeys(attrs map[string]any) []string {
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
