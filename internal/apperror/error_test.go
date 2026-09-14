package apperror

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"strings"
	"testing"
)

// fakeNetError — заглушка сетевой ошибки для проверки Detect.
type fakeNetError struct {
	timeout bool
	msg     string
}

func (f fakeNetError) Error() string   { return f.msg }
func (f fakeNetError) Timeout() bool   { return f.timeout }
func (f fakeNetError) Temporary() bool { return f.timeout }

func TestErrorFormatting(t *testing.T) {
	cause := errors.New("connection refused")

	tests := []struct {
		name  string
		err   *Error
		want  string
		empty bool
	}{
		{"op+message+cause", Wrap(cause, "connect failed").WithOp("scanner.scan"), "scanner.scan: connect failed: connection refused", false},
		{"message+cause", Wrap(cause, "connect failed"), "connect failed: connection refused", false},
		{"only message", New(CodeNotFound, "хост не найден"), "хост не найден", false},
		{"only code", &Error{Code: CodeConflict}, "conflict", false},
		{"empty", &Error{}, "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNilErrorSafe(t *testing.T) {
	var appErr *Error

	if got := appErr.Error(); got != "<nil>" {
		t.Errorf("nil Error() = %q", got)
	}
	if err := appErr.Unwrap(); err != nil {
		t.Errorf("nil Unwrap() = %v", err)
	}
	if appErr.With("k", "v") != nil {
		t.Error("nil With should stay nil")
	}
	if _, ok := appErr.Attr("k"); ok {
		t.Error("nil Attr should be false")
	}
	if appErr.Attrs() != nil {
		t.Error("nil Attrs should be nil")
	}
	if appErr.WithOp("op") != nil {
		t.Error("nil WithOp should stay nil")
	}
}

func TestWrapNilReturnsNil(t *testing.T) {
	if err := Wrap(nil, "message"); err != nil {
		t.Errorf("Wrap(nil) = %v, want nil", err)
	}
	if err := WrapCode(CodeNotFound, nil, "message"); err != nil {
		t.Errorf("WrapCode(nil) = %v, want nil", err)
	}
	if err := From(nil); err != nil {
		t.Errorf("From(nil) = %v, want nil", err)
	}
}

func TestWrapPreservesChain(t *testing.T) {
	sentinel := errors.New("sentinel")
	err := Wrap(sentinel, "step failed").WithOp("task")

	if !errors.Is(err, sentinel) {
		t.Error("wrapped error must keep errors.Is chain")
	}
	if !errors.Is(UnwrapAll(err), sentinel) {
		t.Error("cause should be reachable")
	}
	if err.Unwrap() != sentinel {
		t.Errorf("Unwrap = %v", err.Unwrap())
	}
}

// UnwrapAll — вспомогательная нормализация (тестирует и From).
func UnwrapAll(err error) error { return From(err) }

func TestFrom(t *testing.T) {
	t.Run("passthrough", func(t *testing.T) {
		original := New(CodeTimeout, "таймаут")
		if got := From(original); got != original {
			t.Error("From should return the same *Error instance")
		}
	})

	t.Run("plain error", func(t *testing.T) {
		plain := fmt.Errorf("wrapped: %w", context.DeadlineExceeded)
		appErr := From(plain)
		if appErr.Code != CodeTimeout {
			t.Errorf("code = %q, want %q", appErr.Code, CodeTimeout)
		}
		if !strings.Contains(appErr.Error(), "wrapped") {
			t.Errorf("message lost: %q", appErr.Error())
		}
	})
}

func TestAttrContext(t *testing.T) {
	err := New(CodeInvalidInput, "неверная маска").With("cidr", "10.0.0/24").With("attempt", 3)

	if value, ok := err.Attr("cidr"); !ok || value != "10.0.0/24" {
		t.Errorf("Attr(cidr) = %v, %v", value, ok)
	}
	if _, ok := err.Attr("missing"); ok {
		t.Error("Attr(missing) should be false")
	}

	attrs := err.Attrs()
	if len(attrs) != 2 {
		t.Fatalf("Attrs = %v", attrs)
	}
	attrs["cidr"] = "mutated"
	if value, _ := err.Attr("cidr"); value == "mutated" {
		t.Error("Attrs must return a copy")
	}

	if err.Attrs() == nil {
		t.Error("Attrs should not be nil")
	}
	empty := New(CodeInternal, "x")
	if empty.Attrs() != nil {
		t.Error("empty Attrs should be nil")
	}
}

func TestOfAndIs(t *testing.T) {
	deep := fmt.Errorf("outer: %w", Wrap(errors.New("inner"), "db query"))

	if got := Of(deep); got != CodeUnknown {
		t.Errorf("Of = %q, want %q", got, CodeUnknown)
	}
	if !Is(deep, CodeUnknown) {
		t.Error("Is(CodeUnknown) should be true")
	}

	notFound := New(CodeNotFound, "нет хоста")
	if Of(notFound) != CodeNotFound {
		t.Errorf("Of = %q", Of(notFound))
	}
	if Of(nil) != "" {
		t.Error("Of(nil) should be empty")
	}
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Code
	}{
		{"nil", nil, CodeUnknown},
		{"canceled", context.Canceled, CodeCanceled},
		{"deadline", context.DeadlineExceeded, CodeTimeout},
		{"wrapped deadline", fmt.Errorf("scan: %w", context.DeadlineExceeded), CodeTimeout},
		{"not exist", fmt.Errorf("open config: %w", os.ErrNotExist), CodeNotFound},
		{"fs not exist", fs.ErrNotExist, CodeNotFound},
		{"already exists", os.ErrExist, CodeAlreadyExists},
		{"fs exist", fs.ErrExist, CodeAlreadyExists},
		{"permission", fmt.Errorf("write: %w", os.ErrPermission), CodePermissionDenied},
		{"fs permission", fs.ErrPermission, CodePermissionDenied},
		{"net timeout", fakeNetError{timeout: true, msg: "i/o timeout"}, CodeTimeout},
		{"net refused", fakeNetError{msg: "connection refused"}, CodeUnavailable},
		{"plain", errors.New("что-то пошло не так"), CodeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Detect(tt.err); got != tt.want {
				t.Errorf("Detect(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

func TestDetectAppErrorInChain(t *testing.T) {
	inner := New(CodePermissionDenied, "запрещено")
	outer := fmt.Errorf("layer: %w", inner)

	if got := Detect(outer); got != CodePermissionDenied {
		t.Errorf("Detect = %q, want %q", got, CodePermissionDenied)
	}
}

var errSentinelDB = errors.New("sql: no rows in result set")

func TestRegisterDetector(t *testing.T) {
	t.Cleanup(ResetDetectors)

	RegisterDetector(nil) // игнорируется
	RegisterDetector(func(err error) (Code, bool) {
		if errors.Is(err, errSentinelDB) {
			return CodeNotFound, true
		}
		return "", false
	})

	if got := Detect(fmt.Errorf("query: %w", errSentinelDB)); got != CodeNotFound {
		t.Errorf("Detect = %q, want %q", got, CodeNotFound)
	}
	if got := Detect(errors.New("other")); got != CodeUnknown {
		t.Errorf("Detect(other) = %q, want %q", got, CodeUnknown)
	}

	ResetDetectors()
	if got := Detect(errSentinelDB); got != CodeUnknown {
		t.Errorf("after reset Detect = %q, want %q", got, CodeUnknown)
	}
}

func TestChain(t *testing.T) {
	if Chain(nil, 0) != nil {
		t.Error("Chain(nil) should be nil")
	}

	inner := errors.New("inner")
	middle := Wrap(inner, "middle")
	outer := Wrap(middle, "outer")

	chain := Chain(outer, 0)
	if len(chain) != 3 {
		t.Fatalf("len(chain) = %d, want 3: %v", len(chain), chain)
	}
	if chain[0] != outer {
		t.Error("chain should start with the outermost error")
	}
	if !errors.Is(chain[len(chain)-1], inner) {
		t.Error("chain should end with the root cause")
	}

	limited := Chain(outer, 2)
	if len(limited) != 2 {
		t.Errorf("limit not respected: %d", len(limited))
	}
}

func TestNetErrorInterface(t *testing.T) {
	// Проверка, что заглушка действительно реализует net.Error.
	var target net.Error = fakeNetError{timeout: true}
	if !target.Timeout() {
		t.Error("fake should report timeout")
	}
}
