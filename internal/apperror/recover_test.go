package apperror

import (
	"errors"
	"strings"
	"testing"
)

func TestRecover(t *testing.T) {
	if Recover(nil, "op") != nil {
		t.Error("Recover(nil) should be nil")
	}

	tests := []struct {
		name        string
		value       any
		wantMessage string
		wantCause   bool
	}{
		{"error", errors.New("boom"), "panic: boom", true},
		{"string", "plain panic", "panic: plain panic", true},
		{"int", 42, "panic: 42", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Recover(tt.value, "scanner.scan")
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Code != CodeInternal {
				t.Errorf("code = %q, want %q", err.Code, CodeInternal)
			}
			if err.Op != "scanner.scan" {
				t.Errorf("op = %q", err.Op)
			}
			if err.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", err.Message, tt.wantMessage)
			}
			if (err.Cause != nil) != tt.wantCause {
				t.Errorf("cause = %v, want present %v", err.Cause, tt.wantCause)
			}
			if stack, ok := err.Attr("stack"); !ok || !strings.Contains(stack.(string), "TestRecover") {
				t.Error("stack attribute should capture the panic location")
			}
		})
	}
}

func TestRecoverToCapturesPanic(t *testing.T) {
	fn := func() (err error) {
		defer RecoverTo(&err, "job.run")
		panic(" exploded ")
	}

	err := fn()
	if err == nil {
		t.Fatal("panic should be recorded into err")
	}
	if !strings.Contains(err.Error(), "exploded") {
		t.Errorf("err = %v", err)
	}
	if Of(err) != CodeInternal {
		t.Errorf("code = %q", Of(err))
	}
}

func TestRecoverToKeepsExistingError(t *testing.T) {
	original := errors.New("обычная ошибка")

	fn := func() (err error) {
		defer RecoverTo(&err, "job.run")
		err = original
		return
	}

	if err := fn(); !errors.Is(err, original) {
		t.Errorf("err = %v, want original", err)
	}
}

func TestRecoverToNilPointer(t *testing.T) {
	func() {
		defer RecoverTo(nil, "job.run") // не должен упасть
		panic("boom")
	}()
}

func TestSafe(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		called := false
		err := Safe("op", func() error {
			called = true
			return nil
		})
		if err != nil || !called {
			t.Errorf("err = %v, called = %v", err, called)
		}
	})

	t.Run("business error", func(t *testing.T) {
		business := errors.New("business")
		if err := Safe("op", func() error { return business }); !errors.Is(err, business) {
			t.Errorf("err = %v, want business", err)
		}
	})

	t.Run("panic", func(t *testing.T) {
		err := Safe("worker", func() error { panic("crash") })
		if err == nil {
			t.Fatal("panic should become error")
		}
		if Of(err) != CodeInternal {
			t.Errorf("code = %q", Of(err))
		}
		if !strings.Contains(err.Error(), "worker") {
			t.Errorf("error should mention op: %v", err)
		}
	})

	t.Run("nil func", func(t *testing.T) {
		if err := Safe("op", nil); err != nil {
			t.Errorf("Safe(nil) = %v", err)
		}
	})
}

func TestSafeVoid(t *testing.T) {
	called := false
	if err := SafeVoid("op", func() { called = true }); err != nil || !called {
		t.Errorf("err = %v, called = %v", err, called)
	}

	err := SafeVoid("op", func() { panic("boom") })
	if err == nil || Of(err) != CodeInternal {
		t.Errorf("err = %v", err)
	}

	if err := SafeVoid("op", nil); err != nil {
		t.Errorf("SafeVoid(nil) = %v", err)
	}
}

func TestSafeNestedRecovery(t *testing.T) {
	var innerErr error

	err := Safe("outer", func() error {
		innerErr = Safe("inner", func() error { panic("inner boom") })
		return innerErr
	})

	if err == nil || innerErr == nil {
		t.Fatalf("outer = %v, inner = %v", err, innerErr)
	}
	if !strings.Contains(err.Error(), "inner boom") {
		t.Errorf("outer error lost inner panic: %v", err)
	}
}
