package controller

import (
	"os"
	"testing"
	"time"

	scand "network-scanner/internal/scanner/daemon"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// mockScanActions implements ScanActions for testing
type mockScanActions struct{}

func (m *mockScanActions) ApplyScanRunStart(autoProfileNote string) {}
func (m *mockScanActions) ObserveScanRunner(runner *scand.Runner, startTime time.Time, timeout time.Duration) {
}
func (m *mockScanActions) RenderScanResultsView()             {}
func (m *mockScanActions) ConfirmLargeScanBypass() bool       { return false }
func (m *mockScanActions) SetConfirmLargeScanBypass(val bool) {}

func TestScanController_ApplyPreset(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	a := app.New()

	ui := &ScanUI{
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("", nil),
		ScanBannersCheck:  widget.NewCheck("", nil),
		ScanOSActiveCheck: widget.NewCheck("", nil),
		StatusLabel:       widget.NewLabel(""),
	}
	ctrl := NewScanController(a, ui, &mockScanActions{})

	// Test "quick" preset
	ctrl.ApplyPreset("quick")
	if ui.PortRangeEntry.Text != "22,80,443,445,3389" {
		t.Fatalf("quick preset port range: got %q, want %q", ui.PortRangeEntry.Text, "22,80,443,445,3389")
	}
	if ui.TimeoutEntry.Text != "1" {
		t.Fatalf("quick preset timeout: got %q, want %q", ui.TimeoutEntry.Text, "1")
	}
	if ui.ThreadsEntry.Text != "120" {
		t.Fatalf("quick preset threads: got %q, want %q", ui.ThreadsEntry.Text, "120")
	}
	if ui.ScanUDPCheck.Checked {
		t.Fatal("quick preset: UDP should be unchecked")
	}
	if ui.ScanBannersCheck.Checked {
		t.Fatal("quick preset: banners should be unchecked")
	}
	if ui.ScanOSActiveCheck.Checked {
		t.Fatal("quick preset: OS active should be unchecked")
	}

	// Test "deep" preset
	ctrl.ApplyPreset("deep")
	if ui.PortRangeEntry.Text != "1-2000" {
		t.Fatalf("deep preset port range: got %q, want %q", ui.PortRangeEntry.Text, "1-2000")
	}
	if ui.TimeoutEntry.Text != "3" {
		t.Fatalf("deep preset timeout: got %q, want %q", ui.TimeoutEntry.Text, "3")
	}
	if ui.ThreadsEntry.Text != "40" {
		t.Fatalf("deep preset threads: got %q, want %q", ui.ThreadsEntry.Text, "40")
	}
	if !ui.ScanUDPCheck.Checked {
		t.Fatal("deep preset: UDP should be checked")
	}
	if !ui.ScanBannersCheck.Checked {
		t.Fatal("deep preset: banners should be checked")
	}
	if !ui.ScanOSActiveCheck.Checked {
		t.Fatal("deep preset: OS active should be checked")
	}

	// Test "balanced" preset (default)
	ctrl.ApplyPreset("balanced")
	if ui.PortRangeEntry.Text != "1-1000" {
		t.Fatalf("balanced preset port range: got %q, want %q", ui.PortRangeEntry.Text, "1-1000")
	}
	if ui.TimeoutEntry.Text != "2" {
		t.Fatalf("balanced preset timeout: got %q, want %q", ui.TimeoutEntry.Text, "2")
	}
	if ui.ThreadsEntry.Text != "50" {
		t.Fatalf("balanced preset threads: got %q, want %q", ui.ThreadsEntry.Text, "50")
	}

	// Test unknown preset (should fall back to default)
	ctrl.ApplyPreset("unknown")
	if ui.PortRangeEntry.Text != "1-1000" {
		t.Fatalf("unknown preset should use default: got %q", ui.PortRangeEntry.Text)
	}
}

func TestScanController_ApplyRecommendedProfile(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	a := app.New()

	ui := &ScanUI{
		NetworkEntry:      widget.NewEntry(),
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("", nil),
		ScanBannersCheck:  widget.NewCheck("", nil),
		ScanOSActiveCheck: widget.NewCheck("", nil),
		AutoProfileCheck:  widget.NewCheck("", nil),
		StatusLabel:       widget.NewLabel(""),
		RecommendedBadge:  nil, // не тестируем badge
	}
	ctrl := NewScanController(a, ui, &mockScanActions{})

	// Test with empty network (default profile)
	ctrl.ApplyRecommendedProfile("")
	if ui.StatusLabel.Text == "" {
		t.Fatalf("recommended profile status should not be empty: got %q", ui.StatusLabel.Text)
	}
	if ui.ScanUDPCheck.Checked {
		t.Fatal("recommended profile: UDP should be unchecked")
	}
	if ui.ScanBannersCheck.Checked {
		t.Fatal("recommended profile: banners should be unchecked")
	}
	if ui.ScanOSActiveCheck.Checked {
		t.Fatal("recommended profile: OS active should be unchecked")
	}

	// Test with small subnet (192.168.1.0/24 = 256 hosts)
	ui2 := &ScanUI{
		NetworkEntry:      widget.NewEntry(),
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("", nil),
		ScanBannersCheck:  widget.NewCheck("", nil),
		ScanOSActiveCheck: widget.NewCheck("", nil),
		AutoProfileCheck:  widget.NewCheck("", nil),
		StatusLabel:       widget.NewLabel(""),
	}
	ctrl2 := NewScanController(a, ui2, &mockScanActions{})
	ui2.NetworkEntry.SetText("192.168.1.0/24")
	ctrl2.ApplyRecommendedProfile("192.168.1.0/24")
	if ui2.StatusLabel.Text == "" {
		t.Fatalf("small subnet profile status should not be empty: got %q", ui2.StatusLabel.Text)
	}
}

func TestToolsController_WithHost(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	a := app.New()

	// Test with valid host — проверяем положительный сценарий
	ui := &ToolsUI{
		HostEntry: widget.NewEntry(),
	}
	ctrl := NewToolsController(a, ui)

	ui.HostEntry.SetText("192.168.1.1")
	host, ok := ctrl.withHost()
	if !ok {
		t.Fatal("withHost should return true for valid host")
	}
	if host != "192.168.1.1" {
		t.Fatalf("withHost should return entered host: got %q", host)
	}
}

func TestToolsController_WithHost_Empty(t *testing.T) {
	// Тест пустого хоста — может паниковать при попытке показать dialog
	// в headless-режиме, поэтому оборачиваем в recover
	os.Setenv("FYNE_SCALE", "1")
	a := app.New()

	ui := &ToolsUI{
		HostEntry: widget.NewEntry(),
	}
	ctrl := NewToolsController(a, ui)

	defer func() {
		if r := recover(); r != nil {
			// В headless-режиме dialog.ShowInformation паникует — это ожидаемо
			t.Skipf("Dialog not available in headless mode: %v", r)
		}
	}()

	host, ok := ctrl.withHost()
	if ok {
		t.Fatal("withHost should return false for empty host")
	}
	if host != "" {
		t.Fatalf("withHost should return empty string for empty host: got %q", host)
	}
}

func TestToolsController_ParseIntOrDefault(t *testing.T) {
	if parseIntOrDefault("", 42) != 42 {
		t.Fatal("parseIntOrDefault should return default for empty string")
	}
	if parseIntOrDefault("0", 42) != 42 {
		t.Fatal("parseIntOrDefault should return default for 0")
	}
	if parseIntOrDefault("-5", 42) != 42 {
		t.Fatal("parseIntOrDefault should return default for negative")
	}
	if parseIntOrDefault("100", 42) != 100 {
		t.Fatal("parseIntOrDefault should return parsed value")
	}
}
