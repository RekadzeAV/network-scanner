package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"network-scanner/internal/builder"
)

// ============================================================================
// W2: Тесты CLI-команд device-control и remote-exec.
//
// Проверяют реальную привязку cobra-команд к сервисным функциям, валидацию
// аргументов и защиту необратимых действий (reboot/remote-exec).
// ============================================================================

func testCfg(t *testing.T) builder.Config {
	t.Helper()
	return builder.Config{
		LogLevel: "info",
		DBPath:   filepath.Join(t.TempDir(), "inv.db"),
	}
}

// --- device-control ---------------------------------------------------------

func TestRunDeviceControl_RequiresActionAndTarget(t *testing.T) {
	err := RunDeviceControl(testCfg(t))
	if err == nil {
		t.Fatal("ожидалась ошибка отсутствия --action/--target")
	}
	if !strings.Contains(err.Error(), "требуется --action и --target") {
		t.Errorf("неожиданная ошибка: %v", err)
	}
}

func TestRunDeviceControl_RejectsUnsupportedAction(t *testing.T) {
	err := RunDeviceControl(testCfg(t), "--action", "selfdestruct", "--target", "http://127.0.0.1")
	if err == nil {
		t.Fatal("ожидалась ошибка неподдерживаемого действия")
	}
	if !strings.Contains(err.Error(), "неподдерживаемое действие") {
		t.Errorf("неожиданная ошибка: %v", err)
	}
}

func TestRunDeviceControl_RebootNeedsConfirmation(t *testing.T) {
	err := RunDeviceControl(testCfg(t), "--action", "reboot", "--target", "http://127.0.0.1")
	if err == nil {
		t.Fatal("ожидалась ошибка отсутствия подтверждения reboot")
	}
	if !strings.Contains(err.Error(), "I_UNDERSTAND") {
		t.Errorf("неожиданная ошибка: %v", err)
	}
}

// TestRunDeviceControl_RejectsInsecureTarget — валидация URL делегируется
// security-hardened devicecontrol.validateTargetURL (запрет не-HTTP схем).
func TestRunDeviceControl_RejectsInsecureTarget(t *testing.T) {
	err := RunDeviceControl(testCfg(t), "--action", "status", "--target", "file:///etc/passwd")
	if err == nil {
		t.Fatal("ожидалась ошибка небезопасного target URL")
	}
	if !strings.Contains(err.Error(), "device control") {
		t.Errorf("неожиданная ошибка: %v", err)
	}
}

// TestRunDeviceControl_ConnectionRefused_WritesFailedAudit — успешная валидация,
// но недоступный хост: ошибка Execute, audit фиксирует success=false.
func TestRunDeviceControl_ConnectionRefused_WritesFailedAudit(t *testing.T) {
	audit := filepath.Join(t.TempDir(), "audit.jsonl")
	err := RunDeviceControl(testCfg(t),
		"--action", "status",
		"--target", "http://127.0.0.1:1", // порт 1 недоступен
		"--timeout", "1",
		"--audit-log", audit,
	)
	if err == nil {
		t.Fatal("ожидалась ошибка соединения")
	}

	data, readErr := os.ReadFile(audit)
	if readErr != nil {
		t.Fatalf("audit-файл не создан: %v", readErr)
	}
	if !strings.Contains(string(data), `"success":false`) {
		t.Errorf("audit не зафиксировал неуспех: %s", string(data))
	}
	if !strings.Contains(string(data), `"action":"status"`) {
		t.Errorf("audit не содержит action: %s", string(data))
	}
}

func TestDeviceControlCmd_HasExpectedFlags(t *testing.T) {
	for _, name := range []string{"action", "target", "vendor", "user", "pass", "confirm", "timeout", "audit-log"} {
		if deviceControlCmd.Flags().Lookup(name) == nil {
			t.Errorf("device-control не имеет флага --%s", name)
		}
	}
}

// --- remote-exec ------------------------------------------------------------

func TestRunRemoteExecCLI_RequiresTransportTargetCommand(t *testing.T) {
	err := RunRemoteExecCLI(testCfg(t))
	if err == nil {
		t.Fatal("ожидалась ошибка отсутствия обязательных параметров")
	}
	if !strings.Contains(err.Error(), "требуется --transport, --target и --command") {
		t.Errorf("неожиданная ошибка: %v", err)
	}
}

func TestRemoteExecCmd_HasExpectedFlags(t *testing.T) {
	for _, name := range []string{"transport", "target", "user", "pass", "command", "allow-hosts", "allow-commands", "policy-file", "policy-strict", "consent", "dry-run", "timeout", "audit-log"} {
		if remoteExecCmd.Flags().Lookup(name) == nil {
			t.Errorf("remote-exec не имеет флага --%s", name)
		}
	}
}

// TestFlagsToArgs_OnlyChangedFlags — конвертер передаёт только установленные
// флаги; bool без значения, string/int со значением.
func TestFlagsToArgs_OnlyChangedFlags(t *testing.T) {
	c := deviceControlCmd
	_ = c.Flags().Set("action", "status")
	_ = c.Flags().Set("target", "http://127.0.0.1")
	_ = c.Flags().Set("timeout", "7")
	defer func() {
		_ = c.Flags().Set("action", "")
		_ = c.Flags().Set("target", "")
		_ = c.Flags().Set("timeout", "10")
	}()

	args := flagsToArgs(c)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--action status") {
		t.Errorf("flagsToArgs не содержит --action status: %v", args)
	}
	if !strings.Contains(joined, "--target http://127.0.0.1") {
		t.Errorf("flagsToArgs не содержит --target: %v", args)
	}
	if !strings.Contains(joined, "--timeout 7") {
		t.Errorf("flagsToArgs не содержит --timeout 7: %v", args)
	}
}
