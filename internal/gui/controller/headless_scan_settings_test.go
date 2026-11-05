package controller

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestScanController_LoadSettings_nilApp проверяет, что LoadSettings безопасно
// завершается при nil app.
func TestScanController_LoadSettings_nilApp(t *testing.T) {
	ui := &ScanUI{
		NetworkEntry: widget.NewEntry(),
	}
	ctrl := NewScanController(nil, ui, &mockScanActions{})
	ctrl.LoadSettings() // не должен паниковать
}

// TestScanController_LoadSettings_nilUI проверяет, что LoadSettings безопасно
// завершается при nil ui.
func TestScanController_LoadSettings_nilUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	ctrl := NewScanController(a, nil, &mockScanActions{})
	ctrl.LoadSettings() // не должен паниковать
}

// TestScanController_LoadSettings_full загружает все настройки из Preferences
// в UI-виджеты.
func TestScanController_LoadSettings_full(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()
	p.SetString("scan.network", "192.168.1.0/24")
	p.SetString("scan.port_range", "1-1024")
	p.SetString("scan.timeout_sec", "5")
	p.SetString("scan.threads", "100")
	p.SetString("scan.udp", "true")
	p.SetString("scan.grab_banners", "true")
	p.SetString("scan.os_detect_active", "true")
	p.SetString("scan.verbose_port_logs", "true")
	p.SetString("scan.scan_tcp_ports", "true")
	p.SetString("scan.auto_profile", "true")

	ui := &ScanUI{
		NetworkEntry:         widget.NewEntry(),
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("UDP", nil),
		ScanBannersCheck:     widget.NewCheck("Banners", nil),
		ScanOSActiveCheck:    widget.NewCheck("OS", nil),
		ScanVerboseLogsCheck: widget.NewCheck("Verbose", nil),
		ScanTCPPortsCheck:    widget.NewCheck("TCP", nil),
		AutoProfileCheck:     widget.NewCheck("Auto", nil),
	}

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.LoadSettings()

	if ui.NetworkEntry.Text != "192.168.1.0/24" {
		t.Fatalf("network = %q, want %q", ui.NetworkEntry.Text, "192.168.1.0/24")
	}
	if ui.PortRangeEntry.Text != "1-1024" {
		t.Fatalf("port_range = %q, want %q", ui.PortRangeEntry.Text, "1-1024")
	}
	if ui.TimeoutEntry.Text != "5" {
		t.Fatalf("timeout = %q, want %q", ui.TimeoutEntry.Text, "5")
	}
	if ui.ThreadsEntry.Text != "100" {
		t.Fatalf("threads = %q, want %q", ui.ThreadsEntry.Text, "100")
	}
	if !ui.ScanUDPCheck.Checked {
		t.Fatal("UDP check should be checked")
	}
	if !ui.ScanBannersCheck.Checked {
		t.Fatal("Banners check should be checked")
	}
	if !ui.ScanOSActiveCheck.Checked {
		t.Fatal("OS check should be checked")
	}
	if !ui.ScanVerboseLogsCheck.Checked {
		t.Fatal("Verbose check should be checked")
	}
	if !ui.ScanTCPPortsCheck.Checked {
		t.Fatal("TCP check should be checked")
	}
	if !ui.AutoProfileCheck.Checked {
		t.Fatal("Auto profile check should be checked")
	}
}

// TestScanController_LoadSettings_partialUI проверяет ветки с nil-виджетами
// (ScanBannersCheck, ScanOSActiveCheck и т.д. — не инициализированы).
func TestScanController_LoadSettings_partialUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()
	p.SetString("scan.network", "10.0.0.0/8")
	p.SetString("scan.port_range", "22")
	p.SetString("scan.timeout_sec", "3")
	p.SetString("scan.threads", "20")
	p.SetString("scan.udp", "false")

	ui := &ScanUI{
		NetworkEntry:   widget.NewEntry(),
		PortRangeEntry: widget.NewEntry(),
		TimeoutEntry:   widget.NewEntry(),
		ThreadsEntry:   widget.NewEntry(),
		ScanUDPCheck:   widget.NewCheck("UDP", nil),
		// ScanBannersCheck, ScanOSActiveCheck, ScanVerboseLogsCheck,
		// ScanTCPPortsCheck, AutoProfileCheck — nil
	}

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.LoadSettings() // не должен паниковать

	if ui.NetworkEntry.Text != "10.0.0.0/8" {
		t.Fatalf("network = %q, want %q", ui.NetworkEntry.Text, "10.0.0.0/8")
	}
	if ui.PortRangeEntry.Text != "22" {
		t.Fatalf("port_range = %q, want %q", ui.PortRangeEntry.Text, "22")
	}
	if ui.TimeoutEntry.Text != "3" {
		t.Fatalf("timeout = %q, want %q", ui.TimeoutEntry.Text, "3")
	}
	if ui.ThreadsEntry.Text != "20" {
		t.Fatalf("threads = %q, want %q", ui.ThreadsEntry.Text, "20")
	}
	if ui.ScanUDPCheck.Checked {
		t.Fatal("UDP check should be unchecked")
	}
}

// TestScanController_LoadSettings_emptyPrefs проверяет, что пустые Preferences
// не перезаписывают UI.
func TestScanController_LoadSettings_emptyPrefs(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	// Устанавливаем начальные значения в UI
	ui := &ScanUI{
		NetworkEntry:   widget.NewEntry(),
		PortRangeEntry: widget.NewEntry(),
		TimeoutEntry:   widget.NewEntry(),
		ThreadsEntry:   widget.NewEntry(),
		ScanUDPCheck:   widget.NewCheck("UDP", nil),
	}
	ui.NetworkEntry.SetText("172.16.0.0/16")
	ui.PortRangeEntry.SetText("80,443")
	ui.TimeoutEntry.SetText("10")
	ui.ThreadsEntry.SetText("200")
	ui.ScanUDPCheck.SetChecked(true)

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.LoadSettings()

	// При пустых Preferences значения не должны измениться
	if ui.NetworkEntry.Text != "172.16.0.0/16" {
		t.Fatalf("network = %q, want %q", ui.NetworkEntry.Text, "172.16.0.0/16")
	}
	if ui.PortRangeEntry.Text != "80,443" {
		t.Fatalf("port_range = %q, want %q", ui.PortRangeEntry.Text, "80,443")
	}
	if ui.TimeoutEntry.Text != "10" {
		t.Fatalf("timeout = %q, want %q", ui.TimeoutEntry.Text, "10")
	}
	if ui.ThreadsEntry.Text != "200" {
		t.Fatalf("threads = %q, want %q", ui.ThreadsEntry.Text, "200")
	}
	if !ui.ScanUDPCheck.Checked {
		t.Fatal("UDP check should remain checked")
	}
}

// TestScanController_LoadSettings_tcpPortsDefault проверяет ветку, когда
// scan.scan_tcp_ports пустой (по умолчанию true).
func TestScanController_LoadSettings_tcpPortsDefault(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()
	p.SetString("scan.scan_tcp_ports", "") // пустое значение — по умолчанию true

	ui := &ScanUI{
		ScanTCPPortsCheck: widget.NewCheck("TCP", nil),
	}

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.LoadSettings()

	if !ui.ScanTCPPortsCheck.Checked {
		t.Fatal("TCP ports check should be checked when pref is empty")
	}
}

// TestScanController_LoadSettings_autoProfileDefault проверяет ветку, когда
// scan.auto_profile пустой (по умолчанию true).
func TestScanController_LoadSettings_autoProfileDefault(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()
	p.SetString("scan.auto_profile", "") // пустое значение — по умолчанию true

	ui := &ScanUI{
		AutoProfileCheck: widget.NewCheck("Auto", nil),
	}

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.LoadSettings()

	if !ui.AutoProfileCheck.Checked {
		t.Fatal("Auto profile check should be checked when pref is empty")
	}
}

// TestScanController_SaveSettings_nilApp проверяет, что SaveSettings безопасно
// завершается при nil app.
func TestScanController_SaveSettings_nilApp(t *testing.T) {
	ui := &ScanUI{
		NetworkEntry: widget.NewEntry(),
	}
	ctrl := NewScanController(nil, ui, &mockScanActions{})
	ctrl.SaveSettings() // не должен паниковать
}

// TestScanController_SaveSettings_nilUI проверяет, что SaveSettings безопасно
// завершается при nil ui.
func TestScanController_SaveSettings_nilUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	ctrl := NewScanController(a, nil, &mockScanActions{})
	ctrl.SaveSettings() // не должен паниковать
}

// TestScanController_SaveSettings_full сохраняет все значения UI в Preferences.
func TestScanController_SaveSettings_full(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()

	ui := &ScanUI{
		NetworkEntry:         widget.NewEntry(),
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("UDP", nil),
		ScanBannersCheck:     widget.NewCheck("Banners", nil),
		ScanOSActiveCheck:    widget.NewCheck("OS", nil),
		ScanVerboseLogsCheck: widget.NewCheck("Verbose", nil),
		ScanTCPPortsCheck:    widget.NewCheck("TCP", nil),
		AutoProfileCheck:     widget.NewCheck("Auto", nil),
	}

	ui.NetworkEntry.SetText("192.168.100.0/24")
	ui.PortRangeEntry.SetText("8080-8090")
	ui.TimeoutEntry.SetText("10")
	ui.ThreadsEntry.SetText("200")
	ui.ScanUDPCheck.SetChecked(true)
	ui.ScanBannersCheck.SetChecked(true)
	ui.ScanOSActiveCheck.SetChecked(true)
	ui.ScanVerboseLogsCheck.SetChecked(true)
	ui.ScanTCPPortsCheck.SetChecked(true)
	ui.AutoProfileCheck.SetChecked(true)

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.SaveSettings()

	if p.String("scan.network") != "192.168.100.0/24" {
		t.Fatalf("scan.network = %q, want %q", p.String("scan.network"), "192.168.100.0/24")
	}
	if p.String("scan.port_range") != "8080-8090" {
		t.Fatalf("scan.port_range = %q, want %q", p.String("scan.port_range"), "8080-8090")
	}
	if p.String("scan.timeout_sec") != "10" {
		t.Fatalf("scan.timeout_sec = %q, want %q", p.String("scan.timeout_sec"), "10")
	}
	if p.String("scan.threads") != "200" {
		t.Fatalf("scan.threads = %q, want %q", p.String("scan.threads"), "200")
	}
	if p.String("scan.udp") != "true" {
		t.Fatalf("scan.udp = %q, want %q", p.String("scan.udp"), "true")
	}
	if p.String("scan.grab_banners") != "true" {
		t.Fatalf("scan.grab_banners = %q, want %q", p.String("scan.grab_banners"), "true")
	}
	if p.String("scan.os_detect_active") != "true" {
		t.Fatalf("scan.os_detect_active = %q, want %q", p.String("scan.os_detect_active"), "true")
	}
	if p.String("scan.verbose_port_logs") != "true" {
		t.Fatalf("scan.verbose_port_logs = %q, want %q", p.String("scan.verbose_port_logs"), "true")
	}
	if p.String("scan.scan_tcp_ports") != "true" {
		t.Fatalf("scan.scan_tcp_ports = %q, want %q", p.String("scan.scan_tcp_ports"), "true")
	}
	if p.String("scan.auto_profile") != "true" {
		t.Fatalf("scan.auto_profile = %q, want %q", p.String("scan.auto_profile"), "true")
	}
}

// TestScanController_SaveSettings_allFalse сохраняет все чекбоксы в состояние false.
func TestScanController_SaveSettings_allFalse(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()

	ui := &ScanUI{
		NetworkEntry:         widget.NewEntry(),
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("UDP", nil),
		ScanBannersCheck:     widget.NewCheck("Banners", nil),
		ScanOSActiveCheck:    widget.NewCheck("OS", nil),
		ScanVerboseLogsCheck: widget.NewCheck("Verbose", nil),
		ScanTCPPortsCheck:    widget.NewCheck("TCP", nil),
		AutoProfileCheck:     widget.NewCheck("Auto", nil),
	}
	// Все чекбоксы уже unchecked по умолчанию

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.SaveSettings()

	if p.String("scan.udp") != "false" {
		t.Fatalf("scan.udp = %q, want %q", p.String("scan.udp"), "false")
	}
	if p.String("scan.grab_banners") != "false" {
		t.Fatalf("scan.grab_banners = %q, want %q", p.String("scan.grab_banners"), "false")
	}
	if p.String("scan.os_detect_active") != "false" {
		t.Fatalf("scan.os_detect_active = %q, want %q", p.String("scan.os_detect_active"), "false")
	}
	if p.String("scan.verbose_port_logs") != "false" {
		t.Fatalf("scan.verbose_port_logs = %q, want %q", p.String("scan.verbose_port_logs"), "false")
	}
	if p.String("scan.scan_tcp_ports") != "false" {
		t.Fatalf("scan.scan_tcp_ports = %q, want %q", p.String("scan.scan_tcp_ports"), "false")
	}
	if p.String("scan.auto_profile") != "false" {
		t.Fatalf("scan.auto_profile = %q, want %q", p.String("scan.auto_profile"), "false")
	}
}

// TestScanController_SaveSettings_partialUI сохраняет настройки при частичном
// UI (не все чекбоксы инициализированы).
func TestScanController_SaveSettings_partialUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()

	ui := &ScanUI{
		NetworkEntry:   widget.NewEntry(),
		PortRangeEntry: widget.NewEntry(),
		TimeoutEntry:   widget.NewEntry(),
		ThreadsEntry:   widget.NewEntry(),
		ScanUDPCheck:   widget.NewCheck("UDP", nil),
		// ScanBannersCheck и др. — nil
	}
	ui.NetworkEntry.SetText("10.10.0.0/16")
	ui.PortRangeEntry.SetText("22")
	ui.TimeoutEntry.SetText("1")
	ui.ThreadsEntry.SetText("10")
	ui.ScanUDPCheck.SetChecked(true)

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.SaveSettings()

	if p.String("scan.network") != "10.10.0.0/16" {
		t.Fatalf("scan.network = %q, want %q", p.String("scan.network"), "10.10.0.0/16")
	}
	if p.String("scan.udp") != "true" {
		t.Fatalf("scan.udp = %q, want %q", p.String("scan.udp"), "true")
	}
	// Базовые строки должны сохраниться
	if p.String("scan.timeout_sec") != "1" {
		t.Fatalf("scan.timeout_sec = %q, want %q", p.String("scan.timeout_sec"), "1")
	}
	if p.String("scan.threads") != "10" {
		t.Fatalf("scan.threads = %q, want %q", p.String("scan.threads"), "10")
	}
}

// TestScanController_SettingsRoundTrip проверяет round-trip: SaveSettings,
// затем LoadSettings в новый UI — значения должны совпасть.
func TestScanController_SettingsRoundTrip(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	// Исходный UI
	ui1 := &ScanUI{
		NetworkEntry:         widget.NewEntry(),
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("UDP", nil),
		ScanBannersCheck:     widget.NewCheck("Banners", nil),
		ScanOSActiveCheck:    widget.NewCheck("OS", nil),
		ScanVerboseLogsCheck: widget.NewCheck("Verbose", nil),
		ScanTCPPortsCheck:    widget.NewCheck("TCP", nil),
		AutoProfileCheck:     widget.NewCheck("Auto", nil),
	}
	ui1.NetworkEntry.SetText("192.168.50.0/24")
	ui1.PortRangeEntry.SetText("443")
	ui1.TimeoutEntry.SetText("7")
	ui1.ThreadsEntry.SetText("30")
	ui1.ScanUDPCheck.SetChecked(false)
	ui1.ScanBannersCheck.SetChecked(true)
	ui1.ScanOSActiveCheck.SetChecked(false)
	ui1.ScanVerboseLogsCheck.SetChecked(true)
	ui1.ScanTCPPortsCheck.SetChecked(false)
	ui1.AutoProfileCheck.SetChecked(false)

	ctrl := NewScanController(a, ui1, &mockScanActions{})
	ctrl.SaveSettings()

	// Загрузка в новый UI (отдельный контроллер с ui2)
	ui2 := &ScanUI{
		NetworkEntry:         widget.NewEntry(),
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("UDP", nil),
		ScanBannersCheck:     widget.NewCheck("Banners", nil),
		ScanOSActiveCheck:    widget.NewCheck("OS", nil),
		ScanVerboseLogsCheck: widget.NewCheck("Verbose", nil),
		ScanTCPPortsCheck:    widget.NewCheck("TCP", nil),
		AutoProfileCheck:     widget.NewCheck("Auto", nil),
	}
	ctrl2 := NewScanController(a, ui2, &mockScanActions{})
	ctrl2.LoadSettings()

	if ui2.NetworkEntry.Text != "192.168.50.0/24" {
		t.Fatalf("network = %q, want %q", ui2.NetworkEntry.Text, "192.168.50.0/24")
	}
	if ui2.PortRangeEntry.Text != "443" {
		t.Fatalf("port_range = %q, want %q", ui2.PortRangeEntry.Text, "443")
	}
	if ui2.TimeoutEntry.Text != "7" {
		t.Fatalf("timeout = %q, want %q", ui2.TimeoutEntry.Text, "7")
	}
	if ui2.ThreadsEntry.Text != "30" {
		t.Fatalf("threads = %q, want %q", ui2.ThreadsEntry.Text, "30")
	}
	if ui2.ScanUDPCheck.Checked {
		t.Fatal("UDP should be unchecked")
	}
	if !ui2.ScanBannersCheck.Checked {
		t.Fatal("Banners should be checked")
	}
	if ui2.ScanOSActiveCheck.Checked {
		t.Fatal("OS should be unchecked")
	}
	if !ui2.ScanVerboseLogsCheck.Checked {
		t.Fatal("Verbose should be checked")
	}
	if ui2.ScanTCPPortsCheck.Checked {
		t.Fatal("TCP should be unchecked")
	}
	if ui2.AutoProfileCheck.Checked {
		t.Fatal("Auto profile should be unchecked")
	}
}
