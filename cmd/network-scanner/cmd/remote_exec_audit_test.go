package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"network-scanner/internal/remoteexec"
)

// ============================================================================
// P4: аудит изменяющих операций remote-exec.
//
// Флаг --audit-log обязан не просто печатать путь, а писать JSONL-запись.
// Логируются dry-run, успешные прогоны и отклонённые политикой попытки.
// ============================================================================

// readAuditEntries разбирает JSONL аудит-лог построчно.
func readAuditEntries(t *testing.T, path string) []remoteexec.AuditEntry {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	entries := make([]remoteexec.AuditEntry, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e remoteexec.AuditEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("unmarshal audit line %q: %v", line, err)
		}
		entries = append(entries, e)
	}
	return entries
}

// TestRemoteExecCLI_DryRunWritesAudit — dry-run с разрешённой политикой пишет
// запись с dry_run=true и success=true.
func TestRemoteExecCLI_DryRunWritesAudit(t *testing.T) {
	auditPath := filepath.Join(t.TempDir(), "remote-exec.jsonl")

	err := RunRemoteExecCLI(testCfg(t),
		"--transport", "ssh",
		"--target", "10.0.0.5",
		"--command", "uptime",
		"--allow-hosts", "10.0.0.5",
		"--allow-commands", "uptime",
		"--audit-log", auditPath,
	)
	if err != nil {
		t.Fatalf("ожидался успешный dry-run, получено: %v", err)
	}

	entries := readAuditEntries(t, auditPath)
	if len(entries) != 1 {
		t.Fatalf("записей аудита = %d, want 1", len(entries))
	}
	e := entries[0]
	if !e.DryRun {
		t.Errorf("dry_run = false, want true")
	}
	if !e.Success {
		t.Errorf("success = false, want true")
	}
	if e.Transport != "ssh" || e.Target != "10.0.0.5" || e.Command != "uptime" {
		t.Errorf("поля прогона = %+v, want ssh/10.0.0.5/uptime", e)
	}
	if strings.TrimSpace(e.Message) == "" {
		t.Errorf("message пуст — ожидалось описание результата")
	}
	if strings.TrimSpace(e.Timestamp) == "" {
		t.Errorf("timestamp не заполнен")
	}
}

// TestRemoteExecCLI_RejectedAttemptWritesAudit — отклонённая политикой попытка
// реального выполнения обязана попасть в аудит (success=false, dry_run=false),
// даже если команда вернула ошибку.
func TestRemoteExecCLI_RejectedAttemptWritesAudit(t *testing.T) {
	auditPath := filepath.Join(t.TempDir(), "remote-exec.jsonl")

	err := RunRemoteExecCLI(testCfg(t),
		"--transport", "ssh",
		"--target", "10.9.9.9",
		"--command", "uptime",
		"--allow-hosts", "10.0.0.5",
		"--allow-commands", "uptime",
		"--execute",
		"--consent", "I_UNDERSTAND",
		"--audit-log", auditPath,
	)
	if err == nil {
		t.Fatal("ожидалась ошибка: target вне allowlist")
	}

	entries := readAuditEntries(t, auditPath)
	if len(entries) != 1 {
		t.Fatalf("записей аудита = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.DryRun {
		t.Errorf("dry_run = true, want false (пытались реальное выполнение)")
	}
	if e.Success {
		t.Errorf("success = true, want false (попытка отклонена)")
	}
	if !strings.Contains(e.Message, "allowlist") {
		t.Errorf("message = %q, want содержать причину отказа allowlist", e.Message)
	}
}

// TestRemoteExecCLI_FailedDryRunWritesAudit — dry-run, не прошедший политику,
// пишется с success=false.
func TestRemoteExecCLI_FailedDryRunWritesAudit(t *testing.T) {
	auditPath := filepath.Join(t.TempDir(), "remote-exec.jsonl")

	err := RunRemoteExecCLI(testCfg(t),
		"--transport", "ssh",
		"--target", "10.0.0.5",
		"--command", "rm -rf /",
		"--allow-hosts", "10.0.0.5",
		"--allow-commands", "uptime",
		"--audit-log", auditPath,
	)
	if err == nil {
		t.Fatal("ожидалась ошибка: команда вне allowlist")
	}

	entries := readAuditEntries(t, auditPath)
	if len(entries) != 1 {
		t.Fatalf("записей аудита = %d, want 1", len(entries))
	}
	if entries[0].Success {
		t.Errorf("success = true, want false для отклонённой команды")
	}
	if !entries[0].DryRun {
		t.Errorf("dry_run = false, want true (прогон был в dry-run)")
	}
}

// TestRemoteExecCLI_WithoutAuditFlagWritesNothing — без --audit-log файл не
// создаётся: аудит остаётся явным выбором оператора.
func TestRemoteExecCLI_WithoutAuditFlagWritesNothing(t *testing.T) {
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "remote-exec.jsonl")

	err := RunRemoteExecCLI(testCfg(t),
		"--transport", "ssh",
		"--target", "10.0.0.5",
		"--command", "uptime",
		"--allow-hosts", "10.0.0.5",
		"--allow-commands", "uptime",
	)
	if err != nil {
		t.Fatalf("ожидался успешный dry-run, получено: %v", err)
	}
	if _, err := os.Stat(auditPath); !os.IsNotExist(err) {
		t.Fatalf("аудит-лог создан без --audit-log (stat err = %v)", err)
	}
}

// TestRecordRemoteExecAudit_EmptyPathNoop — пустой путь не приводит к панике
// и не создаёт файлов.
func TestRecordRemoteExecAudit_EmptyPathNoop(t *testing.T) {
	recordRemoteExecAudit(remoteExecOptions{transport: "ssh", target: "h", command: "c", auditPath: "   "}, true, true, "msg")
}
