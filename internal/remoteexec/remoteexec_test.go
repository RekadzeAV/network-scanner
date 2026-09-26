package remoteexec

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestExecute_DryRunHappyPath(t *testing.T) {
	req := Request{
		Transport:     TransportSSH,
		Target:        "10.0.0.10",
		Command:       "uname -a",
		AllowHosts:    []string{"10.0.0.10"},
		AllowCommands: []string{"uname -a"},
		Consent:       ConsentToken,
		DryRun:        true,
	}
	res, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success")
	}
	if !strings.Contains(res.Output, "dry-run") {
		t.Fatalf("unexpected output: %s", res.Output)
	}
}

func TestExecute_RequiresAllowlist(t *testing.T) {
	req := Request{
		Transport:     TransportSSH,
		Target:        "10.0.0.77",
		Command:       "hostname",
		AllowHosts:    []string{"10.0.0.10"},
		AllowCommands: []string{"hostname"},
		Consent:       ConsentToken,
		DryRun:        true,
	}
	_, err := Execute(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "allowlist") {
		t.Fatalf("expected allowlist error, got: %v", err)
	}
}

func TestExecute_RejectsMissingConsent(t *testing.T) {
	req := Request{
		Transport:     TransportSSH,
		Target:        "10.0.0.10",
		Command:       "hostname",
		AllowHosts:    []string{"10.0.0.10"},
		AllowCommands: []string{"hostname"},
		DryRun:        true,
	}
	_, err := Execute(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "consent required") {
		t.Fatalf("expected consent error, got: %v", err)
	}
}

func TestExecute_WinTransportsLimitedToWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-windows assertion")
	}
	req := Request{
		Transport:     TransportWMI,
		Target:        "host1",
		Command:       "ipconfig",
		AllowHosts:    []string{"host1"},
		AllowCommands: []string{"ipconfig"},
		Consent:       ConsentToken,
		DryRun:        true,
	}
	_, err := Execute(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "only on windows") {
		t.Fatalf("expected windows-only error, got: %v", err)
	}
}

func TestExecute_UsesCommandRunner(t *testing.T) {
	orig := execCommandContext
	t.Cleanup(func() { execCommandContext = orig })
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, "go", "version")
		return cmd
	}
	req := Request{
		Transport:       TransportSSH,
		Target:          "10.0.0.10",
		Username:        "user",
		Command:         "hostname",
		AllowHosts:      []string{"10.0.0.10"},
		AllowCommands:   []string{"hostname"},
		Consent:         ConsentToken,
		DryRun:          false,
		ConnectTimeoutS: 3,
	}
	res, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success")
	}
	if !strings.Contains(strings.ToLower(res.Output), "go version") {
		t.Fatalf("unexpected output: %s", res.Output)
	}
}

func TestContainsExactTrim(t *testing.T) {
	if containsExactTrim([]string{"echo test", "hostname"}, "echo test ") != true {
		t.Fatal("expected trimmed exact match")
	}
	if containsExactTrim([]string{"echo test", "hostname"}, "echo") {
		t.Fatal("unexpected partial match")
	}
}

func TestRunTransportUnsupported(t *testing.T) {
	_, err := runTransport(context.Background(), Request{Transport: "bad"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

// ============================================================================
// E7/7.10 — строгий TLS-режим (RequireTLS)
// ============================================================================

// captureExecArgs подменяет execCommandContext, перехватывая имя команды и
// аргументы, и возвращает успешную «пустую» команду.
func captureExecArgs(t *testing.T) (*string, *[]string) {
	t.Helper()
	orig := execCommandContext
	t.Cleanup(func() { execCommandContext = orig })

	name := new(string)
	args := new([]string)
	execCommandContext = func(ctx context.Context, cmdName string, cmdArgs ...string) *exec.Cmd {
		*name = cmdName
		*args = append([]string(nil), cmdArgs...)
		return exec.CommandContext(ctx, "go", "version")
	}
	return name, args
}

func TestRunSSH_RequireTLS_StrictHostKeyChecking(t *testing.T) {
	name, args := captureExecArgs(t)

	if _, err := runSSH(context.Background(), Request{
		Target:          "10.0.0.1",
		Username:        "user",
		Command:         "hostname",
		ConnectTimeoutS: 3,
		RequireTLS:      true,
	}); err != nil {
		t.Fatalf("runSSH() error = %v", err)
	}

	if *name != "ssh" {
		t.Fatalf("command = %q, want ssh", *name)
	}
	joined := strings.Join(*args, " ")
	if !strings.Contains(joined, "StrictHostKeyChecking=yes") {
		t.Fatalf("strict TLS: args = %v, want StrictHostKeyChecking=yes", *args)
	}
	if !strings.HasPrefix(joined, "-o BatchMode=yes") {
		t.Fatalf("args = %v, want BatchMode=yes preserved", *args)
	}
	if !strings.Contains(joined, "user@10.0.0.1") {
		t.Fatalf("args = %v, want user@target", *args)
	}
}

func TestRunSSH_DefaultMode_NoStrictHostKeyChecking(t *testing.T) {
	_, args := captureExecArgs(t)

	if _, err := runSSH(context.Background(), Request{
		Target:          "10.0.0.1",
		Command:         "hostname",
		ConnectTimeoutS: 3,
	}); err != nil {
		t.Fatalf("runSSH() error = %v", err)
	}

	if strings.Contains(strings.Join(*args, " "), "StrictHostKeyChecking") {
		t.Fatalf("без RequireTLS не должно быть StrictHostKeyChecking: %v", *args)
	}
}

func TestRunWinRM_RequireTLS_UsesSSLAndHTTPS(t *testing.T) {
	name, args := captureExecArgs(t)

	if _, err := runWinRM(context.Background(), Request{
		Target:     "host1",
		Command:    "hostname",
		RequireTLS: true,
	}); err != nil {
		t.Fatalf("runWinRM() error = %v", err)
	}

	if *name != "winrs" {
		t.Fatalf("command = %q, want winrs", *name)
	}
	joined := strings.Join(*args, " ")
	if !strings.Contains(joined, "-usessl") {
		t.Fatalf("args = %v, want -usessl", *args)
	}
	if !strings.Contains(joined, "-r:https://host1") {
		t.Fatalf("args = %v, want -r:https://host1", *args)
	}
}

func TestRunWinRM_DefaultMode_NoSSL(t *testing.T) {
	_, args := captureExecArgs(t)

	if _, err := runWinRM(context.Background(), Request{Target: "host1", Command: "hostname"}); err != nil {
		t.Fatalf("runWinRM() error = %v", err)
	}

	joined := strings.Join(*args, " ")
	if strings.Contains(joined, "-usessl") || strings.Contains(joined, "https://") {
		t.Fatalf("без RequireTLS не должно быть TLS-флагов: %v", *args)
	}
	if !strings.Contains(joined, "-r:host1") {
		t.Fatalf("args = %v, want -r:host1 (прежнее поведение)", *args)
	}
}

func TestWinRMTarget(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		requireTLS bool
		want       string
	}{
		{"plain-no-tls", "host1", false, "host1"},
		{"plain-tls", "host1", true, "https://host1"},
		{"http-scheme-tls", "http://host1:5985", true, "http://host1:5985"},
		{"https-scheme-tls", "https://host1:5986", true, "https://host1:5986"},
		{"scheme-case-insensitive", "HTTPS://host1", true, "HTTPS://host1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := winRMTarget(tt.target, tt.requireTLS); got != tt.want {
				t.Errorf("winRMTarget(%q, %v) = %q, want %q", tt.target, tt.requireTLS, got, tt.want)
			}
		})
	}
}

func TestValidateRequest_WMIWithRequireTLS(t *testing.T) {
	err := validateRequest(Request{
		Transport:     TransportWMI,
		Target:        "host1",
		Command:       "ipconfig",
		AllowHosts:    []string{"host1"},
		AllowCommands: []string{"ipconfig"},
		Consent:       ConsentToken,
		RequireTLS:    true,
	})
	if err == nil || !strings.Contains(err.Error(), "does not support a TLS channel") {
		t.Fatalf("expected wmi TLS error, got: %v", err)
	}
}

func TestExecute_DryRunRequireTLS_SSH(t *testing.T) {
	res, err := Execute(context.Background(), Request{
		Transport:     TransportSSH,
		Target:        "10.0.0.10",
		Command:       "hostname",
		AllowHosts:    []string{"10.0.0.10"},
		AllowCommands: []string{"hostname"},
		Consent:       ConsentToken,
		DryRun:        true,
		RequireTLS:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatal("expected success in dry-run with require-tls")
	}
}
