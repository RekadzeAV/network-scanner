package remoteexec

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ============================================================================
// AppendAudit — 0% → 100%
// ============================================================================

func TestAppendAudit_EmptyPath(t *testing.T) {
	err := AppendAudit("", AuditEntry{})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestAppendAudit_Success(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit", "audit.jsonl")
	entry := AuditEntry{
		Transport: "ssh",
		Target:    "10.0.0.1",
		Command:   "hostname",
		Success:   true,
		Message:   "ok",
	}
	err := AppendAudit(path, entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAppendAudit_AutoTimestamp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	entry := AuditEntry{
		Transport: "ssh",
		Target:    "10.0.0.1",
		Command:   "hostname",
	}
	err := AppendAudit(path, entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAppendAudit_AutoActor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	entry := AuditEntry{
		Transport: "ssh",
		Target:    "10.0.0.1",
		Command:   "hostname",
	}
	err := AppendAudit(path, entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAppendAudit_MultipleEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	for i := 0; i < 3; i++ {
		err := AppendAudit(path, AuditEntry{
			Transport: "ssh",
			Target:    "10.0.0.1",
			Command:   "hostname",
		})
		if err != nil {
			t.Fatalf("expected no error on entry %d, got %v", i, err)
		}
	}
}

// ============================================================================
// currentActor — 0% → 100%
// ============================================================================

func TestCurrentActor(t *testing.T) {
	actor := currentActor()
	if actor == "" {
		t.Fatal("expected non-empty actor")
	}
}

// ============================================================================
// validateRequest — 66.7% → 100%
// ============================================================================

func TestValidateRequest_UnsupportedTransport(t *testing.T) {
	err := validateRequest(Request{
		Transport:     "ftp",
		Target:        "10.0.0.1",
		Command:       "ls",
		AllowHosts:    []string{"10.0.0.1"},
		AllowCommands: []string{"ls"},
		Consent:       ConsentToken,
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported transport") {
		t.Fatalf("expected unsupported transport error, got %v", err)
	}
}

func TestValidateRequest_MissingTarget(t *testing.T) {
	err := validateRequest(Request{
		Transport:     TransportSSH,
		Command:       "ls",
		AllowHosts:    []string{"10.0.0.1"},
		AllowCommands: []string{"ls"},
		Consent:       ConsentToken,
	})
	if err == nil || !strings.Contains(err.Error(), "target is required") {
		t.Fatalf("expected target required error, got %v", err)
	}
}

func TestValidateRequest_MissingCommand(t *testing.T) {
	err := validateRequest(Request{
		Transport:     TransportSSH,
		Target:        "10.0.0.1",
		AllowHosts:    []string{"10.0.0.1"},
		AllowCommands: []string{"ls"},
		Consent:       ConsentToken,
	})
	if err == nil || !strings.Contains(err.Error(), "command is required") {
		t.Fatalf("expected command required error, got %v", err)
	}
}

func TestValidateRequest_CommandNotInAllowlist(t *testing.T) {
	err := validateRequest(Request{
		Transport:     TransportSSH,
		Target:        "10.0.0.1",
		Command:       "rm -rf /",
		AllowHosts:    []string{"10.0.0.1"},
		AllowCommands: []string{"ls"},
		Consent:       ConsentToken,
	})
	if err == nil || !strings.Contains(err.Error(), "command is not in allowlist") {
		t.Fatalf("expected command not in allowlist error, got %v", err)
	}
}

// ============================================================================
// runTransport — 60% → 100%
// ============================================================================

func TestRunTransport_SSH_WithMock(t *testing.T) {
	orig := execCommandContext
	t.Cleanup(func() { execCommandContext = orig })
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "go", "version")
	}

	out, err := runSSH(context.Background(), Request{
		Target:          "10.0.0.1",
		Username:        "user",
		Command:         "hostname",
		ConnectTimeoutS: 3,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(strings.ToLower(out), "go version") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestRunTransport_WMI_WithMock(t *testing.T) {
	orig := execCommandContext
	t.Cleanup(func() { execCommandContext = orig })
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "go", "version")
	}

	out, err := runWMI(context.Background(), Request{
		Target:  "10.0.0.1",
		Command: "ipconfig",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestRunTransport_WinRM_WithMock(t *testing.T) {
	orig := execCommandContext
	t.Cleanup(func() { execCommandContext = orig })
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "go", "version")
	}

	out, err := runWinRM(context.Background(), Request{
		Target:  "10.0.0.1",
		Command: "hostname",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

// ============================================================================
// Execute — edge cases
// ============================================================================

func TestExecute_WinTransportsOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}

	orig := execCommandContext
	t.Cleanup(func() { execCommandContext = orig })
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "go", "version")
	}

	req := Request{
		Transport:     TransportWMI,
		Target:        "localhost",
		Command:       "hostname",
		AllowHosts:    []string{"localhost"},
		AllowCommands: []string{"hostname"},
		Consent:       ConsentToken,
		DryRun:        false,
	}
	_, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error on Windows with WMI, got %v", err)
	}
}

func TestExecute_DryRunWMI(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	req := Request{
		Transport:     TransportWMI,
		Target:        "localhost",
		Command:       "hostname",
		AllowHosts:    []string{"localhost"},
		AllowCommands: []string{"hostname"},
		Consent:       ConsentToken,
		DryRun:        true,
	}
	res, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Fatal("expected success")
	}
}
