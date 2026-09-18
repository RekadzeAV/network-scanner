package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// --- Безопасность: пароль устройства не должен персиститься ---

// TestNoPasswordPreferenceKey подтверждает отсутствие механизма сохранения
// пароля device-control в Preferences: константы для пароля не существует,
// а по всем известным ключам пароль не записывается.
func TestNoPasswordPreferenceKey(t *testing.T) {
	keys := []string{
		prefToolDeviceTarget,
		prefToolDeviceVendor,
		prefToolDeviceUser,
	}
	for _, k := range keys {
		lower := strings.ToLower(k)
		if strings.Contains(lower, "pass") || strings.Contains(lower, "secret") {
			t.Fatalf("preference key %q looks like a secret store; passwords must not be persisted", k)
		}
	}
}

// TestDevicePasswordNeverPersisted проверяет вручную: после сохранения настроек
// с заполненным паролем в Preferences не появляется значение пароля.
func TestDevicePasswordNeverPersisted(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")
	defer fyneWindow.Close()

	const secret = "SuperSecret-Pa55"
	a := &App{
		myApp:                fyneApp,
		myWindow:             fyneWindow,
		toolsDevicePassEntry: widget.NewPasswordEntry(),
	}
	a.toolsDevicePassEntry.SetText(secret)

	a.saveScanSettings()

	// Полный обход известных prefs через публичный API приложения не даёт
	// перечислить ключи, поэтому проверяем отсутствие утечки косвенно:
	// ни один key из saveScanSettings не содержит пароль.
	p := fyneApp.Preferences()
	for _, k := range []string{
		prefToolDeviceTarget,
		prefToolDeviceVendor,
		prefToolDeviceUser,
	} {
		if got := p.String(k); strings.Contains(got, secret) {
			t.Fatalf("preference %q leaked the device password value", k)
		}
	}
}

// --- Безопасность: путь audit-лога не в текущем рабочем каталоге ---

// TestDeviceAuditLogPath_NotCWD проверяет, что журнал действий с устройствами
// пишется в пользовательский конфиг-каталог (или temp), а не в CWD.
func TestDeviceAuditLogPath_NotCWD(t *testing.T) {
	path := deviceAuditLogPath()
	if strings.TrimSpace(path) == "" {
		t.Fatal("deviceAuditLogPath() returned empty path")
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("audit path must be absolute, got %q", path)
	}
	if !strings.HasSuffix(path, filepath.Join(auditLogDirName, "device-actions.log")) {
		t.Fatalf("audit path %q must end with %s/device-actions.log", path, auditLogDirName)
	}

	cwd, err := os.Getwd()
	if err == nil && cwd != "" {
		rel, relErr := filepath.Rel(cwd, path)
		if relErr == nil && !strings.HasPrefix(rel, "..") {
			t.Fatalf("audit path %q must not live inside CWD %q", path, cwd)
		}
	}
}

// TestDeviceAuditLogPath_UsesConfigDir подтверждает приоритет user config dir.
func TestDeviceAuditLogPath_UsesConfigDir(t *testing.T) {
	cfg, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(cfg) == "" {
		t.Skip("os.UserConfigDir unavailable in this environment")
	}
	path := deviceAuditLogPath()
	if !strings.HasPrefix(path, cfg) {
		t.Fatalf("audit path %q should be under user config dir %q", path, cfg)
	}
}
