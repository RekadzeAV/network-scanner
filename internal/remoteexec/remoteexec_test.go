package remoteexec

import (
	"context"
	"crypto/tls"
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

func TestTLSConfig_NonTLS_ReturnsEmpty(t *testing.T) {
	cfg, err := TLSConfig(Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected MinVersion TLS1.2, got %v", cfg.MinVersion)
	}
	if cfg.InsecureSkipVerify {
		t.Fatal("expected SkipVerify=false when TLS not requested")
	}
}

func TestTLSConfig_SkipVerify(t *testing.T) {
	cfg, err := TLSConfig(Request{
		TLS:          true,
		TLSSkipVerify: true,
		Target:       "10.0.0.10",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.InsecureSkipVerify {
		t.Fatal("expected SkipVerify=true")
	}
	if cfg.ServerName != "10.0.0.10" {
		t.Fatalf("expected ServerName=10.0.0.10, got %q", cfg.ServerName)
	}
}

func TestTLSConfig_StrictTLS_ForbidsCustom(t *testing.T) {
	_, err := TLSConfig(Request{
		StrictTLS:   true,
		TLSSkipVerify: true,
	})
	if err == nil || !strings.Contains(err.Error(), "strict TLS mode forbids skipping certificate verification") {
		t.Fatalf("expected strict TLS skipverify error, got: %v", err)
	}

	_, err = TLSConfig(Request{
		StrictTLS: true,
		TLS:       true,
	})
	if err == nil || !strings.Contains(err.Error(), "strict TLS mode does not allow custom TLS config") {
		t.Fatalf("expected strict TLS custom config error, got: %v", err)
	}

	_, err = TLSConfig(Request{
		StrictTLS:   true,
		TLSCertFile: "cert.pem",
	})
	if err == nil || !strings.Contains(err.Error(), "strict TLS mode forbids custom certificate files") {
		t.Fatalf("expected strict TLS custom cert error, got: %v", err)
	}
}

func TestTLSConfig_StrictTLS_Valid(t *testing.T) {
	cfg, err := TLSConfig(Request{
		StrictTLS: true,
		Target:    "10.0.0.10",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.InsecureSkipVerify {
		t.Fatal("expected SkipVerify=false in strict mode")
	}
	if cfg.ServerName != "10.0.0.10" {
		t.Fatalf("expected ServerName=10.0.0.10, got %q", cfg.ServerName)
	}
	if cfg.RootCAs == nil {
		t.Fatal("expected RootCAs to be set in strict mode")
	}
}

func TestValidateRequest_StrictTLS_RejectsCombinations(t *testing.T) {
	base := Request{
		Transport:     TransportSSH,
		Target:        "10.0.0.10",
		Command:       "hostname",
		AllowHosts:    []string{"10.0.0.10"},
		AllowCommands: []string{"hostname"},
		Consent:       ConsentToken,
		DryRun:        true,
		StrictTLS:     true,
	}

	if err := validateRequest(base); err != nil {
		t.Fatalf("strict TLS valid request rejected: %v", err)
	}

	badSkip := base
	badSkip.TLSSkipVerify = true
	if err := validateRequest(badSkip); err == nil || !strings.Contains(err.Error(), "strict TLS mode forbids skipping certificate verification") {
		t.Fatalf("expected strict TLS skipverify rejection, got: %v", err)
	}

	badCustom := base
	badCustom.TLSCertFile = "cert.pem"
	if err := validateRequest(badCustom); err == nil || !strings.Contains(err.Error(), "strict TLS mode forbids custom certificate files") {
		t.Fatalf("expected strict TLS custom cert rejection, got: %v", err)
	}
}
