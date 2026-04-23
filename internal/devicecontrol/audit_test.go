package devicecontrol

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendAudit(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "audit.log")
	err := AppendAudit(file, AuditEntry{
		Action:    ActionStatus,
		TargetURL: "http://192.168.1.1",
		Success:   true,
		Message:   "ok",
	})
	if err != nil {
		t.Fatalf("AppendAudit() error = %v", err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), `"action":"status"`) {
		t.Fatalf("audit log does not contain action, got: %s", string(data))
	}
}
