package apperror

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// recordingLogger — захват вызовов структурированного логгера.
type recordingLogger struct {
	msgs [][]any
}

func (l *recordingLogger) Error(msg string, keysAndValues ...any) {
	l.msgs = append(l.msgs, append([]any{msg}, keysAndValues...))
}

func kvValue(pairs []any, key string) (any, bool) {
	for i := 0; i < len(pairs)-1; i++ {
		if k, ok := pairs[i].(string); ok && k == key {
			return pairs[i+1], true
		}
	}
	return nil, false
}

func kvValueLast(pairs []any, key string) (any, bool) {
	for i := len(pairs) - 2; i >= 0; i-- {
		if k, ok := pairs[i].(string); ok && k == key {
			return pairs[i+1], true
		}
	}
	return nil, false
}

func TestLogAttrs(t *testing.T) {
	err := Wrap(errors.New("connection refused"), "connect failed").
		WithOp("scanner.scan").
		With("host", "10.0.0.1").
		With("attempt", 2)

	attrs := LogAttrs(err)

	if value, ok := kvValue(attrs, "code"); !ok || value != string(CodeUnknown) {
		t.Errorf("code = %v (ok=%v)", value, ok)
	}
	if value, ok := kvValue(attrs, "op"); !ok || value != "scanner.scan" {
		t.Errorf("op = %v", value)
	}
	if value, ok := kvValue(attrs, "cause"); !ok || value != "connection refused" {
		t.Errorf("cause = %v", value)
	}
	if value, ok := kvValue(attrs, "host"); !ok || value != "10.0.0.1" {
		t.Errorf("host = %v", value)
	}
	if value, ok := kvValue(attrs, "attempt"); !ok || value != 2 {
		t.Errorf("attempt = %v", value)
	}
	if value, ok := kvValue(attrs, "error"); !ok || !strings.Contains(value.(string), "connect failed") {
		t.Errorf("error = %v", value)
	}
}

func TestLogAttrsCodeInheritance(t *testing.T) {
	err := WrapCode(CodeNotFound, errors.New("no rows"), "load host")
	attrs := LogAttrs(err)

	if value, _ := kvValue(attrs, "code"); value != string(CodeNotFound) {
		t.Errorf("code = %v, want %q", value, CodeNotFound)
	}
}

func TestLogAttrsPlainError(t *testing.T) {
	attrs := LogAttrs(fmt.Errorf("plain: %w", errors.New("root")))

	if _, ok := kvValue(attrs, "code"); !ok {
		t.Error("plain errors should still get a code")
	}
	if _, ok := kvValue(attrs, "cause"); !ok {
		t.Error("cause should be present for wrapped plain error")
	}
}

func TestLogAttrsNilError(t *testing.T) {
	attrs := LogAttrs(nil, "trace", "abc")
	if len(attrs) != 2 {
		t.Fatalf("attrs = %v, want only extra pairs", attrs)
	}
	if value, _ := kvValue(attrs, "trace"); value != "abc" {
		t.Errorf("trace = %v", value)
	}
}

func TestLogAttrsExtraComeLast(t *testing.T) {
	err := New(CodeTimeout, "таймаут").With("env", "internal")
	attrs := LogAttrs(err, "env", "public")

	// Контекст из вызова добавляется последним: логгеры трактуют
	// повторяющийся ключ по принципу «последнее значение важнее».
	value, ok := kvValueLast(attrs, "env")
	if !ok || value != "public" {
		t.Errorf("last env = %v, want public", value)
	}
	if first, _ := kvValue(attrs, "env"); first != "internal" {
		t.Errorf("first env = %v, want internal", first)
	}
}

func TestLogAttrsSortedAttrs(t *testing.T) {
	err := New(CodeInternal, "x").With("zeta", 1).With("alpha", 2).With("mid", 3)
	attrs := LogAttrs(err)

	var seen []string
	for i := 0; i < len(attrs)-1; i++ {
		if key, ok := attrs[i].(string); ok {
			switch key {
			case "zeta", "alpha", "mid":
				seen = append(seen, key)
			}
		}
	}
	if strings.Join(seen, ",") != "alpha,mid,zeta" {
		t.Errorf("attrs not sorted: %v", seen)
	}
}

func TestLogError(t *testing.T) {
	logger := &recordingLogger{}
	err := Wrap(errors.New("disk full"), "write report").WithOp("report.write")

	LogError(logger, err, "report", "weekly")

	if len(logger.msgs) != 1 {
		t.Fatalf("logged %d messages", len(logger.msgs))
	}
	record := logger.msgs[0]
	if record[0] != err.Error() {
		t.Errorf("message = %v", record[0])
	}
	if value, _ := kvValue(record, "op"); value != "report.write" {
		t.Errorf("op = %v", value)
	}
	if value, _ := kvValue(record, "report"); value != "weekly" {
		t.Errorf("report = %v", value)
	}
}

func TestLogErrorNilSafe(t *testing.T) {
	logger := &recordingLogger{}

	LogError(logger, nil)
	LogError(nil, errors.New("boom"))

	if len(logger.msgs) != 0 {
		t.Errorf("nothing should be logged: %v", logger.msgs)
	}
}

func TestToClient(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, ""},
		{"not found", New(CodeNotFound, "хост 10.0.0.9 не найден"), "хост 10.0.0.9 не найден"},
		{"invalid", New(CodeInvalidInput, "неверный CIDR"), "неверный CIDR"},
		{"internal hidden", New(CodeInternal, "panic: nil map"), GenericClientMessage},
		{"unknown hidden", errors.New("runtime detail 0xc000"), GenericClientMessage},
		{"data loss hidden", New(CodeDataLoss, "table dropped"), GenericClientMessage},
		{"empty message hidden", &Error{Code: CodeNotFound}, GenericClientMessage},
		{"nested detail", fmt.Errorf("layer: %w", New(CodeTimeout, "операция прервана по таймауту")), "операция прервана по таймауту"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToClient(tt.err); got != tt.want {
				t.Errorf("ToClient = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToClientHidesStack(t *testing.T) {
	err := Recover(errors.New("секретный путь C:\\Users\\admin"), "job")
	client := ToClient(err)

	if client != GenericClientMessage {
		t.Errorf("panic details leaked: %q", client)
	}
	if logger := (&recordingLogger{}); true {
		LogError(logger, err)
		if len(logger.msgs) == 0 {
			t.Error("panic should still be logged internally")
		}
	}
}

func TestDescribe(t *testing.T) {
	if got := Describe(nil); got != "" {
		t.Errorf("Describe(nil) = %q", got)
	}

	got := Describe(New(CodeConflict, "занято"))
	if !strings.HasPrefix(got, "conflict: ") || !strings.Contains(got, "занято") {
		t.Errorf("Describe = %q", got)
	}

	plain := Describe(errors.New("что-то"))
	if !strings.HasPrefix(plain, string(CodeUnknown)+": ") {
		t.Errorf("Describe(plain) = %q", plain)
	}
}
