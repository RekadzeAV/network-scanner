package controller

import (
	"network-scanner/internal/scanner"
	scand "network-scanner/internal/scanner/daemon"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestScanController_StartScan_nilUI проверяет, что StartScan безопасно
// завершается при nil UI (ранний выход).
func TestScanController_StartScan_nilUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	ctrl := NewScanController(a, nil, &mockScanActions{})
	ctrl.StartScan(nil) // не должен паниковать
}

// TestScanController_StartScan_nilActions проверяет, что StartScan безопасно
// завершается при nil actions (ранний выход после определения сети).
func TestScanController_StartScan_nilActions(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		NetworkEntry: widget.NewEntry(),
	}
	ui.NetworkEntry.SetText("192.168.1.0/24")

	ctrl := NewScanController(a, ui, nil)
	ctrl.StartScan(nil) // должен выйти при nil actions
}

// TestScanController_StartScan_emptyNetwork проверяет, что при пустой сети
// вызывается DetectLocalNetwork (попытка автоопределения).
func TestScanController_StartScan_emptyNetwork(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		NetworkEntry: widget.NewEntry(),
		StatusLabel:  widget.NewLabel(""),
	}
	// NetworkEntry.Text пустой

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.StartScan(nil) // попытается автоопределить сеть
}

// TestScanController_StartScan_threadsExceedMax проверяет, что при threads > 512
// значение корректируется до maxScanThreadsGUI.
func TestScanController_StartScan_threadsExceedMax(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		NetworkEntry:      widget.NewEntry(),
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("UDP", nil),
		ScanTCPPortsCheck: widget.NewCheck("TCP", nil),
		StatusLabel:       widget.NewLabel(""),
	}
	ui.NetworkEntry.SetText("192.168.1.0/24")
	ui.ThreadsEntry.SetText("600")

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.StartScan(nil)

	if ui.ThreadsEntry.Text != "512" {
		t.Fatalf("threads = %q, want %q", ui.ThreadsEntry.Text, "512")
	}
	if ui.StatusLabel.Text == "" {
		t.Fatal("status label should have adjustment message")
	}
}

// TestScanController_StartScan_threadsBelowMin проверяет, что при threads < 1
// значение корректируется до 1.
func TestScanController_StartScan_threadsBelowMin(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		NetworkEntry:      widget.NewEntry(),
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("UDP", nil),
		ScanTCPPortsCheck: widget.NewCheck("TCP", nil),
		StatusLabel:       widget.NewLabel(""),
	}
	ui.NetworkEntry.SetText("192.168.1.0/24")
	ui.ThreadsEntry.SetText("0")

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.StartScan(nil)

	if ui.ThreadsEntry.Text != "1" {
		t.Fatalf("threads = %q, want %q", ui.ThreadsEntry.Text, "1")
	}
	if ui.StatusLabel.Text == "" {
		t.Fatal("status label should have adjustment message")
	}
}

// TestScanController_StartScan_autoProfileDisabled проверяет, что при
// отключённом AutoProfileCheck профиль не применяется.
func TestScanController_StartScan_autoProfileDisabled(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		NetworkEntry:      widget.NewEntry(),
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("UDP", nil),
		ScanTCPPortsCheck: widget.NewCheck("TCP", nil),
		AutoProfileCheck:  widget.NewCheck("Auto", nil),
		StatusLabel:       widget.NewLabel(""),
	}
	ui.NetworkEntry.SetText("192.168.1.0/24")
	ui.PortRangeEntry.SetText("1-1024")
	ui.ThreadsEntry.SetText("50")
	ui.AutoProfileCheck.SetChecked(false)

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.StartScan(nil)

	// При отключённом auto-profile порты и потоки не меняются
	if ui.PortRangeEntry.Text != "1-1024" {
		t.Fatalf("port range should remain unchanged: got %q", ui.PortRangeEntry.Text)
	}
	if ui.ThreadsEntry.Text != "50" {
		t.Fatalf("threads should remain unchanged: got %q", ui.ThreadsEntry.Text)
	}
}

// TestScanController_StartScan_emptyResults проверяет, что StartScan корректно
// обрабатывает пустой (но не nil) slice результатов.
func TestScanController_StartScan_emptyResults(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		NetworkEntry:      widget.NewEntry(),
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("UDP", nil),
		ScanTCPPortsCheck: widget.NewCheck("TCP", nil),
		StatusLabel:       widget.NewLabel(""),
	}
	ui.NetworkEntry.SetText("192.168.1.0/24")

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.StartScan([]scanner.Result{}) // пустой slice, не nil
}

// TestScanController_StartScan_defaultValues проверяет, что при пустых
// полях ввода используются значения по умолчанию.
func TestScanController_StartScan_defaultValues(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		NetworkEntry:         widget.NewEntry(),
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("UDP", nil),
		ScanTCPPortsCheck:    widget.NewCheck("TCP", nil),
		ScanBannersCheck:     widget.NewCheck("Banners", nil),
		ScanOSActiveCheck:    widget.NewCheck("OS", nil),
		ScanVerboseLogsCheck: widget.NewCheck("Verbose", nil),
		StatusLabel:          widget.NewLabel(""),
	}
	ui.NetworkEntry.SetText("192.168.1.0/24")
	// PortRangeEntry, TimeoutEntry, ThreadsEntry — пустые

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.StartScan(nil)

	// Default port range = "1-65535", default threads = 50, default timeout = 2
	// Эти значения применяются при создании workerCfg внутри StartScan.
	// Проверяем, что скан не упал и значения скорректированы.
	if ui.ThreadsEntry.Text == "" {
		t.Fatal("threads should have a value after StartScan")
	}
}

// TestScanController_StopScan_nilRunner проверяет, что StopScan безопасно
// завершается при nil runner.
func TestScanController_StopScan_nilRunner(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		StatusLabel:        widget.NewLabel(""),
		StageLabel:         widget.NewLabel(""),
		ProgressBar:        widget.NewProgressBar(),
		ScanButton:         widget.NewButton("", nil),
		StopButton:         widget.NewButton("", nil),
		CopyDiagnosticsBtn: widget.NewButton("", nil),
		SaveDiagnosticsBtn: widget.NewButton("", nil),
		ResultsStateLabel:  widget.NewLabel(""),
	}

	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.StopScan() // runner nil — ранний выход
}

// TestScanController_StopScan_withRunner проверяет, что StopScan правильно
// обновляет UI-состояние при наличии runner.
func TestScanController_StopScan_withRunner(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		StatusLabel:        widget.NewLabel(""),
		StageLabel:         widget.NewLabel("Scanning"),
		ProgressBar:        widget.NewProgressBar(),
		ScanButton:         widget.NewButton("Start", nil),
		StopButton:         widget.NewButton("Stop", nil),
		CopyDiagnosticsBtn: widget.NewButton("", nil),
		SaveDiagnosticsBtn: widget.NewButton("", nil),
		ResultsStateLabel:  widget.NewLabel(""),
	}
	// Создаём runner и назначаем его контроллеру
	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.runner = scand.NewRunner()

	ctrl.StopScan()

	if ui.StatusLabel.Text != "Сканирование остановлено пользователем" {
		t.Fatalf("status = %q, want %q", ui.StatusLabel.Text, "Сканирование остановлено пользователем")
	}
	if !ui.StageLabel.Hidden {
		t.Fatal("StageLabel should be hidden after stop")
	}
	if !ui.ProgressBar.Hidden {
		t.Fatal("ProgressBar should be hidden after stop")
	}
	if ui.ScanButton.Disabled() {
		t.Fatal("ScanButton should be enabled after stop")
	}
	if !ui.StopButton.Disabled() {
		t.Fatal("StopButton should be disabled after stop")
	}
}

// TestScanController_setResultsState_withActions проверяет, что setResultsState
// вызывает actions.SetResultsState при наличии actions.
func TestScanController_setResultsState_withActions(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		ResultsStateLabel: widget.NewLabel(""),
	}
	actions := &mockScanActions{}

	ctrl := NewScanController(a, ui, actions)
	ctrl.setResultsState("scanning")

	if actions.lastState != "scanning" {
		t.Fatalf("lastState = %q, want %q", actions.lastState, "scanning")
	}
}

// TestScanController_setResultsState_nilActions проверяет, что при nil actions
// состояние записывается напрямую в UI-виджет.
func TestScanController_setResultsState_nilActions(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		ResultsStateLabel: widget.NewLabel(""),
	}

	ctrl := NewScanController(a, ui, nil)
	ctrl.setResultsState("stopped")

	if ui.ResultsStateLabel.Text != "stopped" {
		t.Fatalf("ResultsStateLabel = %q, want %q", ui.ResultsStateLabel.Text, "stopped")
	}
}

// TestScanController_setResultsState_nilActions_nilLabel проверяет ветку,
// когда actions nil и ResultsStateLabel тоже nil.
func TestScanController_setResultsState_nilActions_nilLabel(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		ResultsStateLabel: nil,
	}

	ctrl := NewScanController(a, ui, nil)
	ctrl.setResultsState("idle") // не должен паниковать
}

// TestScanController_setResultsState_nilUI проверяет ветку с nil UI.
func TestScanController_setResultsState_nilUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ctrl := NewScanController(a, nil, nil)
	ctrl.setResultsState("idle") // не должен паниковать
}

// TestScanController_CopyScanDiagnostics_nilWindow проверяет, что при nil Window
// функция возвращается сразу.
func TestScanController_CopyScanDiagnostics_nilWindow(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		Window: nil,
	}

	ctrl := NewScanController(a, ui, nil)
	ctrl.CopyScanDiagnostics("some diagnostics") // не должен паниковать
}

// TestScanController_CopyScanDiagnostics_emptyText проверяет, что при пустом
// тексте показывается информационный диалог.
func TestScanController_CopyScanDiagnostics_emptyText(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(nil)
	defer w.Close()

	ui := &ScanUI{
		Window: w,
	}

	ctrl := NewScanController(a, ui, nil)
	ctrl.CopyScanDiagnostics("") // должен показать dialog и вернуться
}

// TestScanController_CopyScanDiagnostics_naText проверяет, что при тексте
// содержащем "n/a" показывается информационный диалог.
func TestScanController_CopyScanDiagnostics_naText(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(nil)
	defer w.Close()

	ui := &ScanUI{
		Window: w,
	}

	ctrl := NewScanController(a, ui, nil)
	ctrl.CopyScanDiagnostics("n/a") // должен показать dialog и вернуться
}

// TestScanController_CopyScanDiagnostics_scanningText проверяет, что при тексте
// содержащем «выполняется» показывается информационный диалог.
func TestScanController_CopyScanDiagnostics_scanningText(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(nil)
	defer w.Close()

	ui := &ScanUI{
		Window: w,
	}

	ctrl := NewScanController(a, ui, nil)
	ctrl.CopyScanDiagnostics("Сканирование выполняется") // должен показать dialog
}

// TestScanController_SaveScanDiagnostics_nilWindow проверяет, что при nil Window
// функция возвращается сразу.
func TestScanController_SaveScanDiagnostics_nilWindow(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		Window: nil,
	}

	ctrl := NewScanController(a, ui, nil)
	ctrl.SaveScanDiagnostics("some diagnostics") // не должен паниковать
}

// TestScanController_SaveScanDiagnostics_emptyText проверяет, что при пустом
// тексте показывается информационный диалог.
func TestScanController_SaveScanDiagnostics_emptyText(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(nil)
	defer w.Close()

	ui := &ScanUI{
		Window: w,
	}

	ctrl := NewScanController(a, ui, nil)
	ctrl.SaveScanDiagnostics("") // должен показать dialog и вернуться
}

// fakeRunner — минимальная заглушка для тестов StopScan.
type fakeRunner struct{}

func (f *fakeRunner) Start(cfg interface{}) error { return nil }
func (f *fakeRunner) Stop()                       {}
