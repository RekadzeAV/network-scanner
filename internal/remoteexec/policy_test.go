package remoteexec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPolicy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.json")
	content := `{
  "allow_hosts": ["10.0.0.10", " 10.0.0.10 ", "host-a"],
  "allow_commands": ["hostname", " hostname ", "uname -a"]
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	p, err := LoadPolicy(path)
	if err != nil {
		t.Fatalf("load policy: %v", err)
	}
	if len(p.AllowHosts) != 2 {
		t.Fatalf("unexpected hosts length: %d", len(p.AllowHosts))
	}
	if len(p.AllowCommands) != 2 {
		t.Fatalf("unexpected commands length: %d", len(p.AllowCommands))
	}
}

func TestLoadPolicy_RejectsWildcard(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.json")
	content := `{
  "allow_hosts": ["*"],
  "allow_commands": ["hostname"]
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	_, err := LoadPolicy(path)
	if err == nil || !strings.Contains(err.Error(), "wildcard") {
		t.Fatalf("expected wildcard error, got: %v", err)
	}
}
