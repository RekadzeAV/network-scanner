package cmd

import (
	"fmt"

	"network-scanner/internal/builder"
)

// RunDeviceControl запускает управление устройством
func RunDeviceControl(cfg builder.Config, args ...string) error {
	action := ""
	target := ""
	vendor := "generic-http"
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
				_ = args[i+1] // user — placeholder
				i++
			}
		case "--pass", "-p":
			if i+1 < len(args) {
				_ = args[i+1] // pass — placeholder
				i++
			}
		case "--confirm":
			if i+1 < len(args) {
				confirm = args[i+1]
				i++
			}
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
	if action == "" || target == "" {
		return fmt.Errorf("требуется --action и --target")
	}
	if action != "status" && action != "reboot" {
		return fmt.Errorf("неподдерживаемое действие: %s (поддерживается status|reboot)", action)
	}
	if action == "reboot" && confirm != "I_UNDERSTAND" {
		return fmt.Errorf("для reboot требуется --confirm I_UNDERSTAND")
	}

	container := builder.NewContainer(cfg)
	inventoryService := container.GetInventory()

	// Получаем снапшот для получения результатов
	_ = inventoryService

	fmt.Printf("Device Control: action=%s target=%s vendor=%s\n", action, target, vendor)
	fmt.Printf("Timeout: %ds\n", timeout)

	// TODO: реальный вызов device-control
	fmt.Println("Status: Implemented (mock)")

	if auditPath != "" {
		fmt.Printf("Audit log: %s\n", auditPath)
	}

	return nil
}
