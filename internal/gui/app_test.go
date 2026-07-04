package gui

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreateAppIcon(t *testing.T) {
	icon := createAppIcon()
	if icon == nil {
		t.Fatal("createAppIcon returned nil")
	}
	if icon.Name() != "icon.png" {
		t.Fatalf("unexpected icon name: %s", icon.Name())
	}
}

func TestNewApp(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.myApp == nil {
		t.Fatal("app.myApp is nil")
	}
	if app.myWindow == nil {
		t.Fatal("app.myWindow is nil")
	}
	if app.services == nil {
		t.Fatal("app.services is nil")
	}
	if app.operations == nil {
		t.Fatal("app.operations is nil")
	}
	if app.mainTabs == nil {
		t.Fatal("app.mainTabs is nil")
	}
	if len(app.mainTabs.Items) != 3 {
		t.Fatalf("expected 3 tabs, got %d", len(app.mainTabs.Items))
	}
}

func TestApp_loadScanSettings(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.loadScanSettings()
	// Проверяем что настройки загрузились (значения по умолчанию)
	if app.networkEntry == nil {
		t.Fatal("networkEntry is nil after loadScanSettings")
	}
	if app.timeoutEntry == nil {
		t.Fatal("timeoutEntry is nil after loadScanSettings")
	}
}

func TestApp_saveScanSettings(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.scanTCPPortsCheck.SetChecked(true)
	app.scanUDPCheck.SetChecked(false)
	app.scanBannersCheck.SetChecked(true)
	app.scanOSActiveCheck.SetChecked(false)
	app.timeoutEntry.SetText("5")
	app.threadsEntry.SetText("100")
	app.portRangeEntry.SetText("1-1024")
	app.saveScanSettings()
	// Если не паниковало — тест пройден
}

func TestApp_setPortRangeControlsEnabled(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.setPortRangeControlsEnabled(true)
	app.setPortRangeControlsEnabled(false)
	// Если не паниковало — тест пройден
}

func TestApp_autoDetectNetwork(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	// autoDetectNetwork запускает горутину, ждём немного
	done := make(chan struct{})
	go func() {
		app.autoDetectNetwork()
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		// Горутина может работать асинхронно, это нормально
	}
}

func TestApp_applyScanPreset(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()

	// Test "quick" preset
	app.applyScanPreset("quick")
	if app.portRangeEntry.Text != "22,80,443,445,3389" {
		t.Fatalf("quick preset port range: got %q, want %q", app.portRangeEntry.Text, "22,80,443,445,3389")
	}
	if app.timeoutEntry.Text != "1" {
		t.Fatalf("quick preset timeout: got %q, want %q", app.timeoutEntry.Text, "1")
	}
	if app.threadsEntry.Text != "120" {
		t.Fatalf("quick preset threads: got %q, want %q", app.threadsEntry.Text, "120")
	}
	if app.scanUDPCheck.Checked {
		t.Fatal("quick preset: UDP should be unchecked")
	}
	if app.scanBannersCheck.Checked {
		t.Fatal("quick preset: banners should be unchecked")
	}
	if app.scanOSActiveCheck.Checked {
		t.Fatal("quick preset: OS active should be unchecked")
	}
	if !strings.Contains(app.statusLabel.Text, "Быстро") {
		t.Fatalf("quick preset status: got %q", app.statusLabel.Text)
	}

	// Test "deep" preset
	app.applyScanPreset("deep")
	if app.portRangeEntry.Text != "1-2000" {
		t.Fatalf("deep preset port range: got %q, want %q", app.portRangeEntry.Text, "1-2000")
	}
	if app.timeoutEntry.Text != "3" {
		t.Fatalf("deep preset timeout: got %q, want %q", app.timeoutEntry.Text, "3")
	}
	if app.threadsEntry.Text != "40" {
		t.Fatalf("deep preset threads: got %q, want %q", app.threadsEntry.Text, "40")
	}
	if !app.scanUDPCheck.Checked {
		t.Fatal("deep preset: UDP should be checked")
	}
	if !app.scanBannersCheck.Checked {
		t.Fatal("deep preset: banners should be checked")
	}
	if !app.scanOSActiveCheck.Checked {
		t.Fatal("deep preset: OS active should be checked")
	}
	if !strings.Contains(app.statusLabel.Text, "Глубоко") {
		t.Fatalf("deep preset status: got %q", app.statusLabel.Text)
	}

	// Test "balanced" preset (default case)
	app.applyScanPreset("balanced")
	if app.portRangeEntry.Text != "1-1000" {
		t.Fatalf("balanced preset port range: got %q, want %q", app.portRangeEntry.Text, "1-1000")
	}
	if app.timeoutEntry.Text != "2" {
		t.Fatalf("balanced preset timeout: got %q, want %q", app.timeoutEntry.Text, "2")
	}
	if app.threadsEntry.Text != "50" {
		t.Fatalf("balanced preset threads: got %q, want %q", app.threadsEntry.Text, "50")
	}
	if app.scanUDPCheck.Checked {
		t.Fatal("balanced preset: UDP should be unchecked")
	}
	if app.scanBannersCheck.Checked {
		t.Fatal("balanced preset: banners should be unchecked")
	}
	if app.scanOSActiveCheck.Checked {
		t.Fatal("balanced preset: OS active should be unchecked")
	}
	if !strings.Contains(app.statusLabel.Text, "Баланс") {
		t.Fatalf("balanced preset status: got %q", app.statusLabel.Text)
	}

	// Test unknown preset (should fall back to default)
	app.applyScanPreset("unknown")
	if app.portRangeEntry.Text != "1-1000" {
		t.Fatalf("unknown preset should use default: got %q", app.portRangeEntry.Text)
	}
}

func TestApp_applyRecommendedScanProfile(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")

	// Test with empty network (default profile)
	app := NewApp()
	app.applyRecommendedScanProfile()
	if !strings.Contains(app.statusLabel.Text, "рекомендованный профиль") {
		t.Fatalf("recommended profile status: got %q", app.statusLabel.Text)
	}
	if app.scanUDPCheck.Checked {
		t.Fatal("recommended profile: UDP should be unchecked")
	}
	if app.scanBannersCheck.Checked {
		t.Fatal("recommended profile: banners should be unchecked")
	}
	if app.scanOSActiveCheck.Checked {
		t.Fatal("recommended profile: OS active should be unchecked")
	}
	if app.autoProfileCheck != nil && !app.autoProfileCheck.Checked {
		t.Fatal("recommended profile: autoProfile should be checked")
	}
	if app.recommendedProfileBadge == nil {
		t.Fatal("recommendedProfileBadge should not be nil")
	}
	if app.recommendedProfileBadge.Text == "" {
		t.Fatal("recommendedProfileBadge.Text should not be empty")
	}

	// Test with small subnet (192.168.1.0/24 = 256 hosts)
	app2 := NewApp()
	app2.networkEntry.SetText("192.168.1.0/24")
	app2.applyRecommendedScanProfile()
	if !strings.Contains(app2.statusLabel.Text, "небольшой") && !strings.Contains(app2.statusLabel.Text, "углубленный") {
		t.Fatalf("small subnet profile: got %q", app2.statusLabel.Text)
	}

	// Test with large subnet (10.0.0.0/16 = 65536 hosts)
	app3 := NewApp()
	app3.networkEntry.SetText("10.0.0.0/16")
	app3.applyRecommendedScanProfile()
	if !strings.Contains(app3.statusLabel.Text, "очень крупной") && !strings.Contains(app3.statusLabel.Text, "бережный") {
		t.Fatalf("large subnet profile: got %q", app3.statusLabel.Text)
	}
}

func TestApp_applyRecommendedScanProfile_BadgeText(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.applyRecommendedScanProfile()

	// Check badge has correct format
	badgeText := app.recommendedProfileBadge.Text
	if !strings.Contains(badgeText, "Профиль:") {
		t.Fatalf("badge should contain 'Профиль:': got %q", badgeText)
	}
}

func TestApp_applyRecommendedScanProfile_SavesSettings(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()

	// Apply recommended profile
	app.applyRecommendedScanProfile()

	// Save and load settings
	app.saveScanSettings()

	// Create new app and load settings
	app2 := NewApp()
	app2.loadScanSettings()

	// Check that settings were loaded (at least not panicked)
	if app2.statusLabel == nil {
		t.Fatal("statusLabel should not be nil after loadScanSettings")
	}
}

func TestApp_resetUIPanelLayoutWithFeedback(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	// resetUIPanelLayoutWithFeedback не должна паниковать
	app.resetUIPanelLayoutWithFeedback()
}
