package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"network-scanner/internal/builder"
	"network-scanner/internal/devicecontrol"
)

// RunDeviceControl запускает управление устройством (status|reboot) через
// реальный devicecontrol-адаптер с обязательным audit-логом.
func RunDeviceControl(cfg builder.Config, args ...string) error {
	action := ""
	target := ""
	vendor := devicecontrol.VendorGenericHTTP
	user := ""
	pass := ""
	confirm := ""
	timeout := 10
	auditPath := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--action", "-a":
			if i+1 < len(args) {
				action = args[i+1]
				i++
			}
		case "--target", "-T":
			if i+1 < len(args) {
				target = args[i+1]
				i++
			}
		case "--vendor":
			if i+1 < len(args) {
				vendor = args[i+1]
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
		case "--confirm":
			if i+1 < len(args) {
				confirm = args[i+1]
				i++
			}
		case "--timeout":
			if i+1 < len(args) {
				_, _ = fmt.Sscanf(args[i+1], "%d", &timeout)
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
	if action == "" || target == "" {
		return fmt.Errorf("требуется --action и --target")
	}
	if action != devicecontrol.ActionStatus && action != devicecontrol.ActionReboot {
		return fmt.Errorf("неподдерживаемое действие: %s (поддерживается status|reboot)", action)
	}
	if action == devicecontrol.ActionReboot && confirm != "I_UNDERSTAND" {
		return fmt.Errorf("для reboot требуется --confirm I_UNDERSTAND")
	}
	if timeout <= 0 {
		timeout = 10
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	req := devicecontrol.Request{
		Action:    action,
		TargetURL: target,
		Vendor:    vendor,
		Username:  user,
		Password:  pass,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	fmt.Printf("Device Control: action=%s target=%s vendor=%s\n", action, target, vendor)
	fmt.Printf("Timeout: %ds\n", timeout)

	resp, err := devicecontrol.Execute(ctx, req)

	// Audit пишется всегда, включая неуспешные попытки (требование для
	// необратимых действий).
	if auditPath != "" {
		entry := devicecontrol.AuditEntry{
			Action:    action,
			TargetURL: target,
			Vendor:    vendor,
			Success:   err == nil && resp.Success,
			Message:   resp.Message,
		}
		if err != nil {
			entry.Message = err.Error()
		}
		if auditErr := devicecontrol.AppendAudit(auditPath, entry); auditErr != nil {
			fmt.Fprintf(os.Stderr, "Audit log error: %v\n", auditErr)
		} else {
			fmt.Printf("Audit log: %s\n", auditPath)
		}
	}

	if err != nil {
		return fmt.Errorf("device control: %w", err)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	if resp.Message != "" {
		fmt.Printf("Message: %s\n", resp.Message)
	}
	return nil
}
