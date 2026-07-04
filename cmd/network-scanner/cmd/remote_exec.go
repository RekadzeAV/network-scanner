package cmd

import (
	"fmt"
	"strings"
	"time"

	"network-scanner/internal/builder"
	"network-scanner/internal/contracts"
)

// RunRemoteExecCLI запускает удалённое выполнение
func RunRemoteExecCLI(cfg builder.Config, args ...string) error {
	transport := ""
	target := ""
	user := ""
	pass := ""
	command := ""
	allowHosts := ""
	allowCommands := ""
	policyFile := ""
	policyStrict := false
	consent := ""
	dryRun := false
	timeout := 15
	auditPath := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--transport", "-t":
			if i+1 < len(args) {
				transport = args[i+1]
				i++
			}
		case "--target", "-T":
			if i+1 < len(args) {
				target = args[i+1]
				i++
			}
		case "--user", "-u":
			if i+1 < len(args) {
				user = args[i+1]
				i++
			}
		case "--pass", "-p":
			if i+1 < len(args) {
				pass = args[i+1]
				i++
			}
		case "--command", "-c":
			if i+1 < len(args) {
				command = args[i+1]
				i++
			}
		case "--allow-hosts":
			if i+1 < len(args) {
				allowHosts = args[i+1]
				i++
			}
		case "--allow-commands":
			if i+1 < len(args) {
				allowCommands = args[i+1]
				i++
			}
		case "--policy-file":
			if i+1 < len(args) {
				policyFile = args[i+1]
				i++
			}
		case "--policy-strict":
			policyStrict = true
		case "--consent":
			if i+1 < len(args) {
				consent = args[i+1]
				i++
			}
		case "--dry-run":
			dryRun = true
		case "--timeout":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &timeout)
				i++
			}
		case "--audit-log":
			if i+1 < len(args) {
				auditPath = args[i+1]
				i++
			}
		}
	}

	// Валидация
	if transport == "" || target == "" || command == "" {
		return fmt.Errorf("требуется --transport, --target и --command")
	}
	if consent == "" {
		consent = "I_UNDERSTAND"
	}

	// Подготовка policy
	allowHostList := parseCSV(allowHosts)
	allowCmdList := parseCSV(allowCommands)

	policy := contracts.PolicyConfig{
		FilePath:      policyFile,
		Strict:        policyStrict,
		AllowHosts:    allowHostList,
		AllowCommands: allowCmdList,
	}

	container := builder.NewContainer(cfg)
	remoteExecService := container.GetRemoteExec()

	req := contracts.RemoteExecRequest{
		Transport: transport,
		Target:    target,
		User:      user,
		Password:  pass,
		Command:   command,
		Policy:    policy,
		Consent:   consent,
		DryRun:    dryRun,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	if dryRun {
		fmt.Println("=== Dry Run ===")
		fmt.Printf("Transport: %s\nTarget: %s\nCommand: %s\n", transport, target, command)
		if err := remoteExecService.DryRun(nil, req); err != nil {
			return fmt.Errorf("dry run failed: %w", err)
		}
		fmt.Println("Policy check passed")
		return nil
	}

	fmt.Println("=== Remote Exec ===")
	fmt.Printf("Transport: %s\nTarget: %s\nCommand: %s\n", transport, target, command)

	res, err := remoteExecService.Execute(nil, req)
	if err != nil {
		return fmt.Errorf("remote exec failed: %w", err)
	}

	if res.Success {
		fmt.Println("Status: Success")
		if res.Output != "" {
			fmt.Printf("\nOutput:\n%s\n", res.Output)
		}
	} else {
		fmt.Println("Status: Failed")
	}

	// Audit log
	if auditPath != "" {
		fmt.Printf("Audit log: %s\n", auditPath)
	}

	return nil
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
