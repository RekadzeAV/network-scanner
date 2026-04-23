package remoteexec

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const ConsentToken = "I_UNDERSTAND"

const (
	TransportSSH   = "ssh"
	TransportWMI   = "wmi"
	TransportWinRM = "winrm"
)

// Request describes one remote execution operation.
type Request struct {
	Transport       string
	Target          string
	Username        string
	Password        string
	Command         string
	AllowHosts      []string
	AllowCommands   []string
	Consent         string
	DryRun          bool
	Timeout         time.Duration
	ConnectTimeoutS int
}

// Response describes the result of remote execution.
type Response struct {
	Transport string
	Target    string
	Command   string
	Success   bool
	Output    string
	Duration  time.Duration
}

var execCommandContext = exec.CommandContext

// Execute validates policy and runs the command via selected transport.
func Execute(ctx context.Context, req Request) (Response, error) {
	req.Transport = strings.ToLower(strings.TrimSpace(req.Transport))
	req.Target = strings.TrimSpace(req.Target)
	req.Username = strings.TrimSpace(req.Username)
	req.Command = strings.TrimSpace(req.Command)
	req.Consent = strings.TrimSpace(req.Consent)
	if req.Timeout <= 0 {
		req.Timeout = 15 * time.Second
	}
	if req.ConnectTimeoutS <= 0 {
		req.ConnectTimeoutS = 8
	}
	res := Response{
		Transport: req.Transport,
		Target:    req.Target,
		Command:   req.Command,
	}
	if err := validateRequest(req); err != nil {
		return res, err
	}
	start := time.Now()
	if req.DryRun {
		res.Success = true
		res.Output = "dry-run: command validated but not executed"
		res.Duration = time.Since(start)
		return res, nil
	}
	runCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	out, err := runTransport(runCtx, req)
	res.Duration = time.Since(start)
	res.Output = strings.TrimSpace(out)
	res.Success = err == nil
	if err != nil {
		return res, err
	}
	return res, nil
}

func validateRequest(req Request) error {
	if req.Transport != TransportSSH && req.Transport != TransportWMI && req.Transport != TransportWinRM {
		return fmt.Errorf("unsupported transport: %s", req.Transport)
	}
	if req.Target == "" {
		return errors.New("target is required")
	}
	if req.Command == "" {
		return errors.New("command is required")
	}
	if req.Consent != ConsentToken {
		return fmt.Errorf("consent required: --remote-exec-consent %s", ConsentToken)
	}
	if !containsFold(req.AllowHosts, req.Target) {
		return errors.New("target is not in allowlist")
	}
	if !containsExactTrim(req.AllowCommands, req.Command) {
		return errors.New("command is not in allowlist")
	}
	if (req.Transport == TransportWMI || req.Transport == TransportWinRM) && runtime.GOOS != "windows" {
		return fmt.Errorf("transport %s is supported only on windows", req.Transport)
	}
	return nil
}

func runTransport(ctx context.Context, req Request) (string, error) {
	switch req.Transport {
	case TransportSSH:
		return runSSH(ctx, req)
	case TransportWMI:
		return runWMI(ctx, req)
	case TransportWinRM:
		return runWinRM(ctx, req)
	default:
		return "", fmt.Errorf("unsupported transport: %s", req.Transport)
	}
}

func runSSH(ctx context.Context, req Request) (string, error) {
	target := req.Target
	if req.Username != "" {
		target = req.Username + "@" + req.Target
	}
	args := []string{
		"-o", "BatchMode=yes",
		"-o", fmt.Sprintf("ConnectTimeout=%d", req.ConnectTimeoutS),
		target,
		req.Command,
	}
	cmd := execCommandContext(ctx, "ssh", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runWMI(ctx context.Context, req Request) (string, error) {
	args := []string{"/node:" + req.Target, "process", "call", "create", req.Command}
	cmd := execCommandContext(ctx, "wmic", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runWinRM(ctx context.Context, req Request) (string, error) {
	args := []string{"-r:" + req.Target, req.Command}
	cmd := execCommandContext(ctx, "winrs", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func containsFold(items []string, value string) bool {
	value = strings.TrimSpace(value)
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), value) {
			return true
		}
	}
	return false
}

func containsExactTrim(items []string, value string) bool {
	value = strings.TrimSpace(value)
	for _, item := range items {
		if strings.TrimSpace(item) == value {
			return true
		}
	}
	return false
}
