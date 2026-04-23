package nettools

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestRunCmdNotInstalled(t *testing.T) {
	_, err := runCmd(context.Background(), []string{"definitely-missing-command-xyz"}, time.Second)
	if err == nil {
		t.Fatal("expected error for missing command")
	}
	var te *ToolError
	if !errors.As(err, &te) {
		t.Fatalf("expected ToolError, got %T", err)
	}
	if te.Code != ToolErrorNotInstalled {
		t.Fatalf("expected code %q, got %q", ToolErrorNotInstalled, te.Code)
	}
}

func TestRunCmdEmptyCommand(t *testing.T) {
	_, err := runCmd(context.Background(), []string{}, time.Second)
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	var te *ToolError
	if !errors.As(err, &te) {
		t.Fatalf("expected ToolError, got %T", err)
	}
	if te.Code != ToolErrorUnknown {
		t.Fatalf("expected code %q, got %q", ToolErrorUnknown, te.Code)
	}
}

func TestRunCmdTimeoutClassification(t *testing.T) {
	var args []string
	switch runtime.GOOS {
	case "windows":
		if _, err := exec.LookPath("powershell"); err != nil {
			t.Skip("powershell is not available")
		}
		args = []string{"powershell", "-NoProfile", "-Command", "Start-Sleep -Seconds 2"}
	default:
		if _, err := exec.LookPath("sh"); err != nil {
			t.Skip("sh is not available")
		}
		args = []string{"sh", "-c", "sleep 2"}
	}
	_, err := runCmd(context.Background(), args, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	var te *ToolError
	if !errors.As(err, &te) {
		t.Fatalf("expected ToolError, got %T", err)
	}
	if te.Code != ToolErrorTimeout {
		t.Fatalf("expected code %q, got %q", ToolErrorTimeout, te.Code)
	}
}

func TestRunCmdSuccess(t *testing.T) {
	// "go version" доступен в dev/CI окружениях проекта и не зависит от сети.
	out, err := runCmd(context.Background(), []string{"go", "version"}, 3*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHumanizeToolError(t *testing.T) {
	msg := HumanizeToolError(newToolError("ping", ToolErrorTimeout, "превышен таймаут выполнения", context.DeadlineExceeded))
	if msg == "" {
		t.Fatal("expected non-empty message")
	}
	if msg == "превышен таймаут выполнения" {
		t.Fatal("expected hint in humanized message")
	}
}
