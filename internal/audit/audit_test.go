package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendAudit_CreatesFileAndWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "actions.log")

	err := AppendAudit(path, "admin", "reboot", "10.0.0.5", "success", "scheduled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read audit log: %v", err)
	}
	var e AuditEntry
	if err := json.Unmarshal(content, &e); err != nil {
		t.Fatalf("log line is not valid JSON: %v", err)
	}
	if e.Action != "reboot" || e.Target != "10.0.0.5" || e.Result != "success" || e.Detail != "scheduled" {
		t.Fatalf("unexpected entry: %+v", e)
	}
	if e.Timestamp == "" {
		t.Fatal("expected timestamp to be set")
	}
}

func TestAppendAudit_EmptyPathRejected(t *testing.T) {
	err := AppendAudit("", "admin", "reboot", "host", "success", "")
	if err == nil || !strings.Contains(err.Error(), "audit path is required") {
		t.Fatalf("expected audit path required error, got: %v", err)
	}
}

func TestAppendAudit_CreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "deep", "audit.log")

	err := AppendAudit(path, "actor", "action", "target", "failed", "timeout")
	if err != nil {
		t.Fatalf("unexpected error creating nested dirs: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected audit file to exist: %v", err)
	}
}

func TestAppendAudit_MultipleEntriesAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.log")

	actions := []struct {
		action, target, result string
	}{
		{"reboot", "10.0.0.1", "success"},
		{"shutdown", "10.0.0.2", "failed"},
		{"update-config", "10.0.0.3", "success"},
	}

	for _, a := range actions {
		if err := AppendAudit(path, "admin", a.action, a.target, a.result, ""); err != nil {
			t.Fatalf("append error: %v", err)
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != len(actions) {
		t.Fatalf("expected %d lines, got %d", len(actions), len(lines))
	}

	for i, line := range lines {
		var e AuditEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("line %d is not valid JSON: %v", i, err)
		}
		if e.Action != actions[i].action || e.Target != actions[i].target || e.Result != actions[i].result {
			t.Fatalf("line %d mismatch: %+v", i, e)
		}
		if e.Timestamp == "" {
			t.Fatalf("line %d missing timestamp", i)
		}
		if e.Actor == "" {
			t.Fatalf("line %d missing actor", i)
		}
	}
}

func TestAppendAudit_WithActor_OmitsEmptyActor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "actor.log")

	if err := AppendAudit(path, "", "ping", "10.0.0.1", "success", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read: %v", err)
	}
	var e AuditEntry
	if err := json.Unmarshal(content, &e); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if e.Actor == "" {
		t.Fatal("expected actor to be defaulted when empty")
	}
}

func TestEnsureAuditPath_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ensure.log")

	if err := EnsureAuditPath(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to be created: %v", err)
	}
}

func TestEnsureAuditPath_EmptyPathRejected(t *testing.T) {
	if err := EnsureAuditPath(""); err == nil || !strings.Contains(err.Error(), "audit path is required") {
		t.Fatalf("expected error for empty path, got: %v", err)
	}
}
