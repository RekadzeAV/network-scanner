package gui

import (
	"os"
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
	// Fyne GUI требует дисплея — в headless-режиме (CI, SSH) тест пропускаем
	if os.Getenv("FYNE_HEADLESS") == "1" {
		t.Skip("GUI tests require a display server")
	}
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
	skipHeadless(t)
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
	skipHeadless(t)
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
	skipHeadless(t)
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.setPortRangeControlsEnabled(true)
	app.setPortRangeControlsEnabled(false)
	// Если не паниковало — тест пройден
}

func TestApp_autoDetectNetwork(t *testing.T) {
	skipHeadless(t)
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

// TestApp_applyScanPreset — migrated to controller/controller_test.go
// TestApp_applyRecommendedScanProfile — migrated to controller/controller_test.go
// TestApp_applyRecommendedScanProfile_BadgeText — migrated to controller/controller_test.go
// TestApp_applyRecommendedScanProfile_SavesSettings — migrated to controller/controller_test.go

func TestApp_resetUIPanelLayoutWithFeedback(t *testing.T) {
	skipHeadless(t)
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	// resetUIPanelLayoutWithFeedback не должна паниковать
	app.resetUIPanelLayoutWithFeedback()
}
