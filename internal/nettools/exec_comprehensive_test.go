package nettools

import (
	"context"
	"fmt"
	"testing"
)

// ============================================================================
// ToolError — Error(), Unwrap(), HumanizeToolError
// ============================================================================

func TestToolError_Error_EmptyTool(t *testing.T) {
	te := &ToolError{Code: ToolErrorUnknown, Tool: "", Message: "test"}
	msg := te.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestToolError_Error_EmptyMessage(t *testing.T) {
	te := &ToolError{Code: ToolErrorUnknown, Tool: "ping", Message: ""}
	msg := te.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestToolError_Error_Full(t *testing.T) {
	te := &ToolError{Code: ToolErrorTimeout, Tool: "ping", Message: "timeout"}
	msg := te.Error()
	expected := "ping (timeout): timeout"
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}

func TestToolError_Error_WhitespaceTool(t *testing.T) {
	te := &ToolError{Code: ToolErrorUnknown, Tool: "  ping  ", Message: "test"}
	msg := te.Error()
	// Tool должен быть trimmed
	if msg != "ping (unknown): test" {
		t.Errorf("expected trimmed tool, got %q", msg)
	}
}

func TestToolError_Error_WhitespaceMessage(t *testing.T) {
	te := &ToolError{Code: ToolErrorUnknown, Tool: "ping", Message: "  test  "}
	msg := te.Error()
	// Message должен быть trimmed
	if msg != "ping (unknown): test" {
		t.Errorf("expected trimmed message, got %q", msg)
	}
}

func TestToolError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("inner error")
	te := &ToolError{Code: ToolErrorUnknown, Tool: "ping", Message: "test", Err: inner}
	if te.Unwrap() != inner {
		t.Error("Unwrap should return inner error")
	}
}

func TestToolError_UnwrapNil(t *testing.T) {
	te := &ToolError{Code: ToolErrorUnknown, Tool: "ping", Message: "test", Err: nil}
	if te.Unwrap() != nil {
		t.Error("Unwrap should return nil when Err is nil")
	}
}

func TestNewToolError(t *testing.T) {
	inner := fmt.Errorf("inner")
	err := newToolError("  ping  ", ToolErrorTimeout, "  test  ", inner)
	te, ok := err.(*ToolError)
	if !ok {
		t.Fatal("expected *ToolError")
	}
	if te.Tool != "ping" {
		t.Errorf("expected trimmed tool, got %q", te.Tool)
	}
	if te.Message != "test" {
		t.Errorf("expected trimmed message, got %q", te.Message)
	}
	if te.Code != ToolErrorTimeout {
		t.Errorf("expected timeout code, got %q", te.Code)
	}
	if te.Err != inner {
		t.Error("expected inner error to be preserved")
	}
}

// ============================================================================
// HumanizeToolError — все ветки
// ============================================================================

func TestHumanizeToolError_Nil(t *testing.T) {
	msg := HumanizeToolError(nil)
	if msg != "" {
		t.Errorf("expected empty for nil, got %q", msg)
	}
}

func TestHumanizeToolError_NotToolError(t *testing.T) {
	err := fmt.Errorf("plain error")
	msg := HumanizeToolError(err)
	if msg != "plain error" {
		t.Errorf("expected plain error message, got %q", msg)
	}
}

func TestHumanizeToolError_NotInstalled(t *testing.T) {
	err := newToolError("nmap", ToolErrorNotInstalled, "not found", nil)
	msg := HumanizeToolError(err)
	if msg == "" {
		t.Error("expected non-empty message")
	}
	if msg == "not found" {
		t.Error("expected hint in humanized message")
	}
}

func TestHumanizeToolError_PermissionDenied(t *testing.T) {
	err := newToolError("arp", ToolErrorPermissionDenied, "access denied", nil)
	msg := HumanizeToolError(err)
	if msg == "" {
		t.Error("expected non-empty message")
	}
	if msg == "access denied" {
		t.Error("expected hint in humanized message")
	}
}

func TestHumanizeToolError_Timeout(t *testing.T) {
	err := newToolError("ping", ToolErrorTimeout, "timeout", context.DeadlineExceeded)
	msg := HumanizeToolError(err)
	if msg == "" {
		t.Error("expected non-empty message")
	}
	if msg == "timeout" {
		t.Error("expected hint in humanized message")
	}
}

func TestHumanizeToolError_Network(t *testing.T) {
	err := newToolError("traceroute", ToolErrorNetwork, "no route", nil)
	msg := HumanizeToolError(err)
	if msg == "" {
		t.Error("expected non-empty message")
	}
	if msg == "no route" {
		t.Error("expected hint in humanized message")
	}
}

func TestHumanizeToolError_Parse(t *testing.T) {
	err := newToolError("whois", ToolErrorParse, "parse failed", nil)
	msg := HumanizeToolError(err)
	if msg == "" {
		t.Error("expected non-empty message")
	}
	if msg == "parse failed" {
		t.Error("expected plain message for parse error")
	}
}

func TestHumanizeToolError_Unknown(t *testing.T) {
	err := newToolError("unknown-tool", ToolErrorUnknown, "unknown error", nil)
	msg := HumanizeToolError(err)
	if msg == "" {
		t.Error("expected non-empty message")
	}
}

func TestHumanizeToolError_EmptyMessage(t *testing.T) {
	err := newToolError("ping", ToolErrorUnknown, "", nil)
	msg := HumanizeToolError(err)
	if msg == "" {
		t.Error("expected non-empty message for empty error")
	}
}

func TestHumanizeToolError_EmptyMessageHint(t *testing.T) {
	err := newToolError("ping", ToolErrorTimeout, "", nil)
	msg := HumanizeToolError(err)
	if msg == "" {
		t.Error("expected non-empty message")
	}
	// Должна быть базовая фраза "не удалось выполнить инструмент"
	if msg == "не удалось выполнить инструмент" {
		t.Error("expected hint appended to base message")
	}
}

// ============================================================================
// ToolErrorCode — проверка констант
// ============================================================================

func TestToolErrorCode_Constants(t *testing.T) {
	expected := []ToolErrorCode{
		ToolErrorNotInstalled,
		ToolErrorPermissionDenied,
		ToolErrorTimeout,
		ToolErrorNetwork,
		ToolErrorParse,
		ToolErrorUnknown,
	}
	for _, code := range expected {
		if code == "" {
			t.Error("expected non-empty code constant")
		}
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkToolError_Error(b *testing.B) {
	te := &ToolError{Code: ToolErrorTimeout, Tool: "ping", Message: "timeout"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = te.Error()
	}
}

func BenchmarkHumanizeToolError(b *testing.B) {
	te := newToolError("ping", ToolErrorTimeout, "timeout", context.DeadlineExceeded)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HumanizeToolError(te)
	}
}
