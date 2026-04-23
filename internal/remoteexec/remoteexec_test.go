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
