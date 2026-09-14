package controller

import (
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func ensureApp() fyne.App {
	os.Setenv("FYNE_SCALE", "1")
	return app.New()
}

// --- scan_controller.go extended tests ---

func TestScanController_LoadSettings_NilApp(t *testing.T) {
	c := &ScanController{}
	// Не должен паниковать
	c.LoadSettings()
}

func TestScanController_LoadSettings_NilUI(t *testing.T) {
	c := &ScanController{app: ensureApp()}
	// Не должен паниковать
	c.LoadSettings()
}

func TestScanController_LoadSettings_WithUI(t *testing.T) {
	ui := &ScanUI{
		NetworkEntry:      widget.NewEntry(),
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("", nil),
		ScanBannersCheck:  widget.NewCheck("", nil),
		ScanOSActiveCheck: widget.NewCheck("", nil),
		StatusLabel:       widget.NewLabel(""),
	}
	c := NewScanController(ensureApp(), ui, nil)
	// Не должен паниковать
	c.LoadSettings()
}

func TestScanController_SaveSettings_NilApp(t *testing.T) {
	c := &ScanController{}
	// Не должен паниковать
	c.SaveSettings()
}

func TestScanController_SaveSettings_NilUI(t *testing.T) {
	c := &ScanController{app: ensureApp()}
	// Не должен паниковать
	c.SaveSettings()
}

func TestScanController_SaveSettings_WithUI(t *testing.T) {
	ui := &ScanUI{
		NetworkEntry:      widget.NewEntry(),
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("", nil),
		ScanBannersCheck:  widget.NewCheck("", nil),
		ScanOSActiveCheck: widget.NewCheck("", nil),
		ScanTCPPortsCheck: widget.NewCheck("", nil),
		AutoProfileCheck:  widget.NewCheck("", nil),
	}
	c := NewScanController(ensureApp(), ui, nil)
	// Не должен паниковать
	c.SaveSettings()
}

func TestScanController_ApplyPreset_Quick(t *testing.T) {
	ui := &ScanUI{
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("", nil),
		ScanBannersCheck:     widget.NewCheck("", nil),
		ScanOSActiveCheck:    widget.NewCheck("", nil),
		ScanVerboseLogsCheck: widget.NewCheck("", nil),
		StatusLabel:          widget.NewLabel(""),
	}
	c := NewScanController(ensureApp(), ui, nil)
	c.ApplyPreset("quick")
	if ui.PortRangeEntry.Text != "22,80,443,445,3389" {
		t.Errorf("expected '22,80,443,445,3389', got %q", ui.PortRangeEntry.Text)
	}
	if ui.TimeoutEntry.Text != "1" {
		t.Errorf("expected '1', got %q", ui.TimeoutEntry.Text)
	}
	if ui.ThreadsEntry.Text != "120" {
		t.Errorf("expected '120', got %q", ui.ThreadsEntry.Text)
	}
	if ui.ScanUDPCheck.Checked {
		t.Error("expected ScanUDPCheck=false for quick preset")
	}
	if ui.StatusLabel.Text != "Пресет: Быстро (обзор)" {
		t.Errorf("expected status text, got %q", ui.StatusLabel.Text)
	}
}

func TestScanController_ApplyPreset_Deep(t *testing.T) {
	ui := &ScanUI{
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("", nil),
		ScanBannersCheck:     widget.NewCheck("", nil),
		ScanOSActiveCheck:    widget.NewCheck("", nil),
		ScanVerboseLogsCheck: widget.NewCheck("", nil),
		StatusLabel:          widget.NewLabel(""),
	}
	c := NewScanController(ensureApp(), ui, nil)
	c.ApplyPreset("deep")
	if ui.PortRangeEntry.Text != "1-2000" {
		t.Errorf("expected '1-2000', got %q", ui.PortRangeEntry.Text)
	}
	if ui.ScanUDPCheck.Checked != true {
		t.Error("expected ScanUDPCheck=true for deep preset")
	}
	if ui.ScanBannersCheck.Checked != true {
		t.Error("expected ScanBannersCheck=true for deep preset")
	}
	if ui.ScanOSActiveCheck.Checked != true {
		t.Error("expected ScanOSActiveCheck=true for deep preset")
	}
	if ui.StatusLabel.Text != "Пресет: Глубоко (детальный анализ)" {
		t.Errorf("expected status text, got %q", ui.StatusLabel.Text)
	}
}

func TestScanController_ApplyPreset_Balanced(t *testing.T) {
	ui := &ScanUI{
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("", nil),
		ScanBannersCheck:     widget.NewCheck("", nil),
		ScanOSActiveCheck:    widget.NewCheck("", nil),
		ScanVerboseLogsCheck: widget.NewCheck("", nil),
		StatusLabel:          widget.NewLabel(""),
	}
	c := NewScanController(ensureApp(), ui, nil)
	c.ApplyPreset("balanced")
	if ui.PortRangeEntry.Text != "1-1000" {
		t.Errorf("expected '1-1000', got %q", ui.PortRangeEntry.Text)
	}
	if ui.ThreadsEntry.Text != "50" {
		t.Errorf("expected '50', got %q", ui.ThreadsEntry.Text)
	}
	if ui.StatusLabel.Text != "Пресет: Баланс" {
		t.Errorf("expected status text, got %q", ui.StatusLabel.Text)
	}
}

func TestScanController_ApplyPreset_Unknown(t *testing.T) {
	ui := &ScanUI{
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("", nil),
		ScanBannersCheck:     widget.NewCheck("", nil),
		ScanOSActiveCheck:    widget.NewCheck("", nil),
		ScanVerboseLogsCheck: widget.NewCheck("", nil),
		StatusLabel:          widget.NewLabel(""),
	}
	c := NewScanController(ensureApp(), ui, nil)
	c.ApplyPreset("unknown")
	if ui.PortRangeEntry.Text != "1-1000" {
		t.Errorf("expected '1-1000' (default), got %q", ui.PortRangeEntry.Text)
	}
}

func TestScanController_RecommendedBadgeClassForHosts(t *testing.T) {
	c := &ScanController{}
	if c.recommendedBadgeClassForHosts(0) != "green" {
		t.Error("expected 'green' for 0 hosts")
	}
	if c.recommendedBadgeClassForHosts(511) != "green" {
		t.Error("expected 'green' for 511 hosts")
	}
	if c.recommendedBadgeClassForHosts(512) != "yellow" {
		t.Error("expected 'yellow' for 512 hosts")
	}
	if c.recommendedBadgeClassForHosts(1023) != "yellow" {
		t.Error("expected 'yellow' for 1023 hosts")
	}
	if c.recommendedBadgeClassForHosts(1024) != "orange" {
		t.Error("expected 'orange' for 1024 hosts")
	}
	if c.recommendedBadgeClassForHosts(2047) != "orange" {
		t.Error("expected 'orange' for 2047 hosts")
	}
	if c.recommendedBadgeClassForHosts(2048) != "red" {
		t.Error("expected 'red' for 2048 hosts")
	}
}

func TestScanController_RecommendedBadgeText(t *testing.T) {
	c := &ScanController{}
	text := c.recommendedBadgeText("test profile", "green")
	expected := "test profile (green)"
	if text != expected {
		t.Errorf("expected %q, got %q", expected, text)
	}
}

func TestScanController_RefreshPresetUI_NilUI(t *testing.T) {
	c := &ScanController{}
	// Не должен паниковать
	c.RefreshPresetUI()
}

func TestScanController_RefreshPresetUI_WithUI(t *testing.T) {
	ui := &ScanUI{
		PortRangeEntry: widget.NewEntry(),
		TimeoutEntry:   widget.NewEntry(),
		ThreadsEntry:   widget.NewEntry(),
		ScanUDPCheck:   widget.NewCheck("", nil),
		StatusLabel:    widget.NewLabel(""),
	}
	c := &ScanController{ui: ui}
	// Не должен паниковать
	c.RefreshPresetUI()
}

func TestScanController_StopScan_NilRunner(t *testing.T) {
	c := &ScanController{}
	// Не должен паниковать при nil runner
	c.StopScan()
}

func TestScanController_FormatDurationMMSS(t *testing.T) {
	if formatDurationMMSS(0) != "00:00" {
		t.Error("expected '00:00' for zero")
	}
	if formatDurationMMSS(-5*time.Second) != "00:00" {
		t.Error("expected '00:00' for negative")
	}
}
