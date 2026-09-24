package devicecontrol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Файл закрывает ветки ошибок audit.go и validateTargetURL, которые
// «счастливые» тесты не достигают.

// TestAppendAudit_MkdirError: родительский путь — обычный файл,
// поэтому os.MkdirAll обязан вернуть ошибку.
func TestAppendAudit_MkdirError(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker.txt")
	if err := os.WriteFile(blocker, []byte("not a dir"), 0o600); err != nil {
		t.Fatalf("prepare blocker file: %v", err)
	}

	err := AppendAudit(filepath.Join(blocker, "sub", "audit.jsonl"), AuditEntry{
		Action: ActionStatus,
	})
	if err == nil {
		t.Fatal("expected error when audit dir cannot be created")
	}
	if !strings.Contains(err.Error(), "create audit dir") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestAppendAudit_OpenFileError: путь указывает на существующий каталог —
// открытие файла на запись невозможно.
func TestAppendAudit_OpenFileError(t *testing.T) {
	err := AppendAudit(t.TempDir(), AuditEntry{Action: ActionReboot})
	if err == nil {
		t.Fatal("expected error when audit path is a directory")
	}
	if !strings.Contains(err.Error(), "open audit file") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestAppendAudit_WriteError использует /dev/full: файл открывается, но
// запись завершается ошибкой. Доступен только на Linux.
func TestAppendAudit_WriteError(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("/dev/full доступен только на Linux")
	}

	err := AppendAudit("/dev/full", AuditEntry{Action: ActionStatus, Message: "x"})
	if err == nil {
		t.Fatal("expected write error on /dev/full")
	}
	if !strings.Contains(err.Error(), "write audit entry") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestAppendAudit_DefaultsToJSONL проверяет, что запись оканчивается
// переводом строки и является валидным JSON (формат JSONL).
func TestAppendAudit_DefaultsToJSONL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")

	for _, action := range []string{ActionStatus, ActionReboot} {
		if err := AppendAudit(path, AuditEntry{
			Action:    action,
			TargetURL: "https://192.168.1.1",
			Vendor:    VendorTPLINKHTTP,
			Success:   action == ActionStatus,
			Message:   "done",
		}); err != nil {
			t.Fatalf("AppendAudit(%s): %v", action, err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2 (%q)", len(lines), string(data))
	}

	for i, line := range lines {
		var entry AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("line %d is not JSON: %v (%s)", i, err, line)
		}
		if entry.Timestamp == "" {
			t.Errorf("line %d: timestamp must be filled automatically", i)
		}
		if entry.Actor == "" {
			t.Errorf("line %d: actor must be filled automatically", i)
		}
	}
}

// TestValidateTargetURL_ParseError закрывает ветку ошибки разбора URL.
func TestValidateTargetURL_ParseError(t *testing.T) {
	for _, raw := range []string{"http://[::1", "http://%zz:80", "https://%%"} {
		if err := validateTargetURL(raw); err == nil {
			t.Errorf("validateTargetURL(%q) = nil, want error", raw)
		} else if !strings.Contains(err.Error(), "invalid target URL") {
			t.Errorf("validateTargetURL(%q) = %v, want parse error", raw, err)
		}
	}
}

// TestExecute_RebootWithoutConsent проверяет, что подтверждение требуется
// до валидации URL и до обращения к устройству.
func TestExecute_RebootWithoutConsent(t *testing.T) {
	_, err := Execute(t.Context(), Request{Action: ActionReboot, TargetURL: "not a url"})
	if err == nil {
		t.Fatal("expected consent error")
	}
	if !strings.Contains(err.Error(), ConsentToken) {
		t.Errorf("error = %v, want mention of %s", err, ConsentToken)
	}
}
