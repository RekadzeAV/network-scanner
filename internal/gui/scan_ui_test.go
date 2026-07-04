package gui

import (
	"testing"
	"time"
)

// --- Test initScanUI ---

func TestInitScanUI(t *testing.T) {
	// Создаём App без GUI (mock)
	app := &App{}

	// Вызываем initScanUI
	app.initScanUI()

	// Проверяем, что все поля инициализированы
	if app.networkEntry == nil {
		t.Error("networkEntry should not be nil")
	}

	if app.portRangeEntry == nil {
		t.Error("portRangeEntry should not be nil")
	}

	if app.scanTCPPortsCheck == nil {
		t.Error("scanTCPPortsCheck should not be nil")
	}

	if app.timeoutEntry == nil {
		t.Error("timeoutEntry should not be nil")
	}

	if app.threadsEntry == nil {
		t.Error("threadsEntry should not be nil")
	}

	if app.scanUDPCheck == nil {
		t.Error("scanUDPCheck should not be nil")
	}

	if app.scanBannersCheck == nil {
		t.Error("scanBannersCheck should not be nil")
	}

	if app.scanOSActiveCheck == nil {
		t.Error("scanOSActiveCheck should not be nil")
	}

	if app.scanButton == nil {
		t.Error("scanButton should not be nil")
	}

	if app.stopButton == nil {
		t.Error("stopButton should not be nil")
	}

	if app.saveButton == nil {
		t.Error("saveButton should not be nil")
	}

	if app.statusLabel == nil {
		t.Error("statusLabel should not be nil")
	}

	if app.progressBar == nil {
		t.Error("progressBar should not be nil")
	}
}

// --- Test buildScanControlsContainer ---

func TestBuildScanControlsContainer(t *testing.T) {
	app := &App{}
	app.initScanUI()

	container := app.buildScanControlsContainer()

	if container == nil {
		t.Error("buildScanControlsContainer() should not return nil")
	}
}

// --- Test buildResultsContainer ---

func TestBuildResultsContainer(t *testing.T) {
	app := &App{}
	app.initScanUI()

	container := app.buildResultsContainer()

	if container == nil {
		t.Error("buildResultsContainer() should not return nil")
	}
}

// --- Test buildScanTabContent ---

func TestBuildScanTabContent(t *testing.T) {
	app := &App{}
	app.initScanUI()

	content := app.buildScanTabContent()

	if content == nil {
		t.Error("buildScanTabContent() should not return nil")
	}
}

// --- Test results state constants ---

func TestResultsStateConstants(t *testing.T) {
	if resultsStateIdle != "idle" {
		t.Errorf("resultsStateIdle = %v, want idle", resultsStateIdle)
	}

	if resultsStateScanning != "scanning" {
		t.Errorf("resultsStateScanning = %v, want scanning", resultsStateScanning)
	}

	if resultsStateDone != "done" {
		t.Errorf("resultsStateDone = %v, want done", resultsStateDone)
	}

	if resultsStateStopped != "stopped" {
		t.Errorf("resultsStateStopped = %v, want stopped", resultsStateStopped)
	}

	if resultsStateTimeout != "timeout" {
		t.Errorf("resultsStateTimeout = %v, want timeout", resultsStateTimeout)
	}
}

// --- Test constants ---

func TestMaxScanThreadsGUI(t *testing.T) {
	if maxScanThreadsGUI != 512 {
		t.Errorf("maxScanThreadsGUI = %v, want 512", maxScanThreadsGUI)
	}
}

func TestLargeSubnetWarnHostGUI(t *testing.T) {
	if largeSubnetWarnHostGUI != 512 {
		t.Errorf("largeSubnetWarnHostGUI = %v, want 512", largeSubnetWarnHostGUI)
	}
}

func TestAutoProfileHostWarn(t *testing.T) {
	if autoProfileHostWarn != 256 {
		t.Errorf("autoProfileHostWarn = %v, want 256", autoProfileHostWarn)
	}
}

func TestAutoProfileHostLarge(t *testing.T) {
	if autoProfileHostLarge != 512 {
		t.Errorf("autoProfileHostLarge = %v, want 512", autoProfileHostLarge)
	}
}

func TestAutoProfileHostXLarge(t *testing.T) {
	if autoProfileHostXLarge != 1024 {
		t.Errorf("autoProfileHostXLarge = %v, want 1024", autoProfileHostXLarge)
	}
}

func TestAutoProfileHostXXLarge(t *testing.T) {
	if autoProfileHostXXLarge != 2048 {
		t.Errorf("autoProfileHostXXLarge = %v, want 2048", autoProfileHostXXLarge)
	}
}

func TestMinWindowWidth(t *testing.T) {
	if minWindowWidth != 1024 {
		t.Errorf("minWindowWidth = %v, want 1024", minWindowWidth)
	}
}

// --- Test resultsRenderDebounce ---

func TestResultsRenderDebounce(t *testing.T) {
	app := &App{}
	app.initScanUI()

	// Проверяем, что resultsRenderDebounce установлен по умолчанию
	if app.resultsRenderDebounce != 180*time.Millisecond {
		t.Errorf("resultsRenderDebounce = %v, want 180ms", app.resultsRenderDebounce)
	}
}

// --- Test resultsMode defaults ---

func TestResultsModeDefaults(t *testing.T) {
	app := &App{}
	app.initScanUI()
	app.buildResultsContainer()

	if app.resultsMode != "Таблица" {
		t.Errorf("resultsMode = %v, want Таблица", app.resultsMode)
	}

	if app.resultsSubMode != "Devices" {
		t.Errorf("resultsSubMode = %v, want Devices", app.resultsSubMode)
	}

	if app.resultsSort != "IP" {
		t.Errorf("resultsSort = %v, want IP", app.resultsSort)
	}

	if app.maxPortChips != 24 {
		t.Errorf("maxPortChips = %v, want 24", app.maxPortChips)
	}

	if app.resultsState != resultsStateIdle {
		t.Errorf("resultsState = %v, want idle", app.resultsState)
	}
}

// --- Test chip limit ---

func TestChipLimitSelOptions(t *testing.T) {
	app := &App{}
	app.initScanUI()
	app.buildResultsContainer()

	// Проверяем, что chipLimitSel инициализирован
	if app.chipLimitSel == nil {
		t.Error("chipLimitSel should not be nil")
	}

	// Проверяем, что выбрано значение по умолчанию
	if app.chipLimitSel.Selected != "24" {
		t.Errorf("chipLimitSel.Selected = %v, want 24", app.chipLimitSel.Selected)
	}
}

// --- Test port state mode ---

func TestPortStateModeDefaults(t *testing.T) {
	app := &App{}
	app.initScanUI()
	app.buildResultsContainer()

	if app.resultsPortStateMode != "all" {
		t.Errorf("resultsPortStateMode = %v, want all", app.resultsPortStateMode)
	}
}

// --- Test showRawBanners ---

func TestShowRawBannersDefault(t *testing.T) {
	app := &App{}
	app.initScanUI()
	app.buildResultsContainer()

	if app.showRawBanners {
		t.Error("showRawBanners should be false by default")
	}
}

// --- Test onlyWithOpenPorts ---

func TestOnlyWithOpenPortsDefault(t *testing.T) {
	app := &App{}
	app.initScanUI()
	app.buildResultsContainer()

	if app.onlyWithOpenPorts {
		t.Error("onlyWithOpenPorts should be false by default")
	}
}

// --- Test quickTypeChecks ---

func TestQuickTypeChecksInit(t *testing.T) {
	app := &App{}
	app.initScanUI()
	app.buildResultsContainer()

	if app.quickTypeChecks == nil {
		t.Error("quickTypeChecks should not be nil")
	}

	// Проверяем, что все типы инициализированы
	expectedTypes := []string{"Network Device", "Computer", "Server", "Unknown"}
	for _, typ := range expectedTypes {
		if app.quickTypeChecks[typ] == nil {
			t.Errorf("quickTypeChecks[%s] should not be nil", typ)
		}
	}
}

// --- Test inventory settings ---

func TestInventorySettingsDefaults(t *testing.T) {
	app := &App{}
	app.initScanUI()
	app.buildResultsContainer()

	if app.inventoryDBEntry == nil {
		t.Error("inventoryDBEntry should not be nil")
	}

	if app.inventoryAutoSaveCheck == nil {
		t.Error("inventoryAutoSaveCheck should not be nil")
	}

	// Проверяем, что автосохранение включено по умолчанию
	if !app.inventoryAutoSaveCheck.Checked {
		t.Error("inventoryAutoSaveCheck should be checked by default")
	}
}

// --- Test autoProfileCheck ---

func TestAutoProfileCheckDefault(t *testing.T) {
	app := &App{}
	app.initScanUI()

	// Проверяем, что autoProfileCheck инициализирован
	if app.autoProfileCheck == nil {
		t.Error("autoProfileCheck should not be nil")
	}

	// Проверяем, что autoProfileCheck включён по умолчанию
	if !app.autoProfileCheck.Checked {
		t.Error("autoProfileCheck should be checked by default")
	}
}

// --- Test scanTCPPortsCheck ---

func TestScanTCPPortsCheckDefault(t *testing.T) {
	app := &App{}
	app.initScanUI()

	// Проверяем, что scanTCPPortsCheck инициализирован
	if app.scanTCPPortsCheck == nil {
		t.Error("scanTCPPortsCheck should not be nil")
	}

	// Проверяем, что scanTCPPortsCheck включён по умолчанию
	if !app.scanTCPPortsCheck.Checked {
		t.Error("scanTCPPortsCheck should be checked by default")
	}
}

// --- Test port range entry ---

func TestPortRangeEntryDefault(t *testing.T) {
	app := &App{}
	app.initScanUI()

	// Проверяем, что portRangeEntry инициализирован
	if app.portRangeEntry == nil {
		t.Error("portRangeEntry should not be nil")
	}

	// Проверяем, что установлено значение по умолчанию
	if app.portRangeEntry.Text != "1-65535" {
		t.Errorf("portRangeEntry.Text = %v, want 1-65535", app.portRangeEntry.Text)
	}
}

// --- Test timeout entry ---

func TestTimeoutEntryDefault(t *testing.T) {
	app := &App{}
	app.initScanUI()

	// Проверяем, что timeoutEntry инициализирован
	if app.timeoutEntry == nil {
		t.Error("timeoutEntry should not be nil")
	}

	// Проверяем, что установлено значение по умолчанию
	if app.timeoutEntry.Text != "2" {
		t.Errorf("timeoutEntry.Text = %v, want 2", app.timeoutEntry.Text)
	}
}

// --- Test threads entry ---

func TestThreadsEntryDefault(t *testing.T) {
	app := &App{}
	app.initScanUI()

	// Проверяем, что threadsEntry инициализирован
	if app.threadsEntry == nil {
		t.Error("threadsEntry should not be nil")
	}

	// Проверяем, что установлено значение по умолчанию
	if app.threadsEntry.Text != "50" {
		t.Errorf("threadsEntry.Text = %v, want 50", app.threadsEntry.Text)
	}
}

// --- Test snmp settings ---

func TestSNMPSettingsDefaults(t *testing.T) {
	app := &App{}
	app.initScanUI()

	// Проверяем, что snmpCommEntry инициализирован
	if app.snmpCommEntry == nil {
		t.Error("snmpCommEntry should not be nil")
	}

	// Проверяем, что установлено значение по умолчанию
	if app.snmpCommEntry.Text != "public" {
		t.Errorf("snmpCommEntry.Text = %v, want public", app.snmpCommEntry.Text)
	}

	// Проверяем, что snmpTimeoutEnt инициализирован
	if app.snmpTimeoutEnt == nil {
		t.Error("snmpTimeoutEnt should not be nil")
	}

	// Проверяем, что установлено значение по умолчанию
	if app.snmpTimeoutEnt.Text != "2" {
		t.Errorf("snmpTimeoutEnt.Text = %v, want 2", app.snmpTimeoutEnt.Text)
	}
}
