package auditpath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPath_BaseDirAndName(t *testing.T) {
	path := DefaultPath(DeviceActionsFile)
	if strings.TrimSpace(path) == "" {
		t.Fatal("DefaultPath returned empty path")
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("path must be absolute, got %q", path)
	}
	if !strings.HasSuffix(path, filepath.Join(DirName, DeviceActionsFile)) {
		t.Fatalf("path %q must end with %s/%s", path, DirName, DeviceActionsFile)
	}
}

func TestDefaultPath_EmptyNameFallsBack(t *testing.T) {
	if got := DefaultPath("   "); !strings.HasSuffix(got, "actions.log") {
		t.Fatalf("DefaultPath(\"  \") = %q, want actions.log suffix", got)
	}
}

func TestDeviceAndRemoteExecPathsDiffer(t *testing.T) {
	if DeviceActionsPath() == RemoteExecPath() {
		t.Fatal("device-control и remote-exec журналы должны быть разными файлами")
	}
	if !strings.HasSuffix(RemoteExecPath(), filepath.Join(DirName, RemoteExecFile)) {
		t.Fatalf("RemoteExecPath() = %q, want %s/%s", RemoteExecPath(), DirName, RemoteExecFile)
	}
}

// TestPathsNotInCWD — инвариант безопасности: журналы не пишутся в рабочий каталог.
func TestPathsNotInCWD(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil || cwd == "" {
		t.Skipf("os.Getwd unavailable: %v", err)
	}
	for _, path := range []string{DeviceActionsPath(), RemoteExecPath()} {
		rel, relErr := filepath.Rel(cwd, path)
		if relErr == nil && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel) {
			t.Fatalf("audit path %q must not live inside CWD %q", path, cwd)
		}
	}
}
