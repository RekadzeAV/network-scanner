package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"network-scanner/internal/builder"
	"network-scanner/internal/contracts"
)

// remoteExecConsentToken — токен явного согласия на реальное удалённое
// выполнение. Совпадает с remoteexec.ConsentToken.
const remoteExecConsentToken = "I_UNDERSTAND"

// defaultRemoteExecTimeout — таймаут выполнения по умолчанию, секунды.
const defaultRemoteExecTimeout = 15

// remoteExecOptions — разобранные параметры подкоманды remote-exec.
//
// dryRun по умолчанию включён (P0-2): без явного --execute команда не
// запускает транспорт, а только прогоняет проверку политики.
type remoteExecOptions struct {
	transport     string
	target        string
	user          string
	pass          string
	command       string
	allowHosts    []string
	allowCommands []string
	policyFile    string
	policyStrict  bool
	consent       string
	dryRun        bool
	timeout       int
	auditPath     string
	requireTLS    bool
}

// parseRemoteExecArgs разбирает и валидирует argv подкоманды remote-exec.
//
// Безопасность (P0-2 «dry-run по умолчанию»):
//   - реальное выполнение включается только явным --execute (синоним
//     --dry-run=false сохранён для обратной совместимости);
//   - реальное выполнение требует --consent I_UNDERSTAND.
func parseRemoteExecArgs(args []string) (remoteExecOptions, error) {
	opts := remoteExecOptions{dryRun: true, timeout: defaultRemoteExecTimeout}

	for i := 0; i < len(args); i++ {
		name, inline, hasInline := splitFlag(args[i])

		// next читает значение: сначала из "--flag=value", затем из следующего
		// аргумента.
		next := func() (string, bool) {
			if hasInline {
				return inline, true
			}
			if i+1 < len(args) {
				i++
				return args[i], true
			}
			return "", false
		}

		switch name {
		case "--transport", "-t":
			if v, ok := next(); ok {
				opts.transport = v
			}
		case "--target", "-T":
			if v, ok := next(); ok {
				opts.target = v
			}
		case "--user", "-u":
			if v, ok := next(); ok {
				opts.user = v
			}
		case "--pass", "-p":
			if v, ok := next(); ok {
				opts.pass = v
			}
		case "--command", "-c":
			if v, ok := next(); ok {
				opts.command = v
			}
		case "--allow-hosts":
			if v, ok := next(); ok {
				opts.allowHosts = parseCSV(v)
			}
		case "--allow-commands":
			if v, ok := next(); ok {
				opts.allowCommands = parseCSV(v)
			}
		case "--policy-file":
			if v, ok := next(); ok {
				opts.policyFile = v
			}
		case "--policy-strict":
			opts.policyStrict = parseBoolFlag(inline, hasInline, true)
		case "--consent":
			if v, ok := next(); ok {
				opts.consent = v
			}
		case "--dry-run":
			// Обратная совместимость: явный --dry-run=false означает «выполнить».
			opts.dryRun = parseBoolFlag(inline, hasInline, true)
		case "--execute":
			opts.dryRun = !parseBoolFlag(inline, hasInline, true)
		case "--timeout":
			if v, ok := next(); ok {
				if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
					opts.timeout = n
				}
			}
		case "--audit-log":
			if v, ok := next(); ok {
				opts.auditPath = v
			}
		case "--require-tls", "--strict-tls":
			// Строгий режим канала (E7/7.10): ssh — StrictHostKeyChecking=yes,
			// winrm — -usessl по https. Для wmi TLS не поддерживается (ошибка).
			opts.requireTLS = parseBoolFlag(inline, hasInline, true)
		}
	}

	if opts.transport == "" || opts.target == "" || opts.command == "" {
		return opts, fmt.Errorf("требуется --transport, --target и --command")
	}
	if opts.timeout <= 0 {
		opts.timeout = defaultRemoteExecTimeout
	}
	if !opts.dryRun && strings.TrimSpace(opts.consent) != remoteExecConsentToken {
		return opts, fmt.Errorf("для реального выполнения требуется --consent %s", remoteExecConsentToken)
	}
	if opts.dryRun && strings.TrimSpace(opts.consent) == "" {
		// Dry-run не обращается к транспорту — только проверяет политику.
		// Токен подставляется, чтобы сквозная валидация remoteexec не требовала
		// согласия на неизменяющую операцию.
		opts.consent = remoteExecConsentToken
	}

	return opts, nil
}

// splitFlag разбирает "--name=value" в ("--name", "value", true). Короткие
// флаги ("-t ssh") значения в форме "=" не поддерживают.
func splitFlag(token string) (string, string, bool) {
	if !strings.HasPrefix(token, "--") {
		return token, "", false
	}
	if idx := strings.Index(token, "="); idx > 2 {
		return token[:idx], token[idx+1:], true
	}
	return token, "", false
}

// parseBoolFlag читает значение булева флага: без значения — def, иначе
// strconv.ParseBool (ошибка разбора → def).
func parseBoolFlag(inline string, hasInline bool, def bool) bool {
	if !hasInline {
		return def
	}
	v, err := strconv.ParseBool(strings.TrimSpace(inline))
	if err != nil {
		return def
	}
	return v
}

// RunRemoteExecCLI запускает удалённое выполнение.
//
// По умолчанию работает в dry-run (P0-2): проверяет транспорт, цель и команду
// против политики, но не подключается к хосту. Реальное выполнение — только с
// --execute и --consent I_UNDERSTAND.
func RunRemoteExecCLI(cfg builder.Config, args ...string) error {
	opts, err := parseRemoteExecArgs(args)
	if err != nil {
		return err
	}

	req := contracts.RemoteExecRequest{
		Transport: opts.transport,
		Target:    opts.target,
		User:      opts.user,
		Password:  opts.pass,
		Command:   opts.command,
		Policy: contracts.PolicyConfig{
			FilePath:      opts.policyFile,
			Strict:        opts.policyStrict,
			AllowHosts:    opts.allowHosts,
			AllowCommands: opts.allowCommands,
		},
		Consent:    opts.consent,
		DryRun:     opts.dryRun,
		Timeout:    time.Duration(opts.timeout) * time.Second,
		RequireTLS: opts.requireTLS,
	}

	container := builder.NewContainer(cfg)
	remoteExecService := container.GetRemoteExec()

	if opts.dryRun {
		fmt.Println("=== Dry Run (реальное выполнение выключено; добавьте --execute) ===")
		fmt.Printf("Transport: %s\nTarget: %s\nCommand: %s\n", opts.transport, opts.target, opts.command)
		if opts.requireTLS {
			fmt.Println("TLS: strict (require-tls)")
		}
		if err := remoteExecService.DryRun(context.TODO(), req); err != nil {
			return fmt.Errorf("dry run failed: %w", err)
		}
		fmt.Println("Policy check passed")
		if opts.auditPath != "" {
			fmt.Printf("Audit log: %s\n", opts.auditPath)
		}
		return nil
	}

	fmt.Println("=== Remote Exec ===")
	fmt.Printf("Transport: %s\nTarget: %s\nCommand: %s\n", opts.transport, opts.target, opts.command)
	if opts.requireTLS {
		fmt.Println("TLS: strict (require-tls)")
	}

	res, err := remoteExecService.Execute(context.TODO(), req)
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
	if opts.auditPath != "" {
		fmt.Printf("Audit log: %s\n", opts.auditPath)
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
