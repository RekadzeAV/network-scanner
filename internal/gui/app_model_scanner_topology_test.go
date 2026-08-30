package gui

import (
	"testing"

	"network-scanner/internal/contracts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// --- model.go: SetStatus tests ---

func TestAppModel_SetStatus_NilBinding(t *testing.T) {
	m := NewAppModel()
	// Не должен паниковать при nil binding
	m.SetStatus("test")
}

// --- scanner_service.go: ValidateConfig tests ---

func TestScannerGUIService_ValidateConfig_EmptyCIDR(t *testing.T) {
	s := &ScannerGUIService{}
	err := s.ValidateConfig(contracts.ScanConfig{})
	if err == nil {
		t.Error("expected error for empty CIDR")
	}
}

func TestScannerGUIService_ValidateConfig_EmptyPortRange(t *testing.T) {
	s := &ScannerGUIService{}
	err := s.ValidateConfig(contracts.ScanConfig{NetworkCIDR: "192.168.1.0/24"})
	if err == nil {
		t.Error("expected error for empty port range")
	}
}

func TestScannerGUIService_ValidateConfig_ZeroTimeout(t *testing.T) {
	s := &ScannerGUIService{}
	err := s.ValidateConfig(contracts.ScanConfig{
		NetworkCIDR: "192.168.1.0/24",
		PortRange:   "1-1000",
		Timeout:     0,
	})
	if err == nil {
		t.Error("expected error for zero timeout")
	}
}

func TestScannerGUIService_ValidateConfig_NegativeTimeout(t *testing.T) {
	s := &ScannerGUIService{}
	err := s.ValidateConfig(contracts.ScanConfig{
		NetworkCIDR: "192.168.1.0/24",
		PortRange:   "1-1000",
		Timeout:     -1,
	})
	if err == nil {
		t.Error("expected error for negative timeout")
	}
}

func TestScannerGUIService_ValidateConfig_Valid(t *testing.T) {
	s := &ScannerGUIService{}
	err := s.ValidateConfig(contracts.ScanConfig{
		NetworkCIDR: "192.168.1.0/24",
		PortRange:   "1-1000",
		Timeout:     2,
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestScannerGUIService_Stop_NilService(t *testing.T) {
	s := &ScannerGUIService{}
	// Не должен паниковать
	s.Stop()
}

// --- results_view.go: adaptivePanelMinHeight tests ---

func TestAdaptivePanelMinHeight_NilApp(t *testing.T) {
	var a *App
	result := a.adaptivePanelMinHeight(100, 1.0, 0.5, 50)
	if result != 100 {
		t.Errorf("expected 100, got %f", result)
	}
}

func TestAdaptivePanelMinHeight_EmptyApp(t *testing.T) {
	a := &App{}
	result := a.adaptivePanelMinHeight(100, 1.0, 0.5, 50)
	if result != 100 {
		t.Errorf("expected 100 (base*1.0 capped by 720*0.5=360), got %f", result)
	}
}

func TestAdaptivePanelMinHeight_WithCanvasSize(t *testing.T) {
	a := &App{}
	a.lastCanvasSize = fyne.NewSize(1920, 1080)
	result := a.adaptivePanelMinHeight(400, 1.0, 0.3, 100)
	if result <= 0 {
		t.Error("expected positive result")
	}
}

func TestAdaptivePanelMinHeight_ClampedToMax(t *testing.T) {
	a := &App{}
	a.lastCanvasSize = fyne.NewSize(1920, 400)
	// base=400, layoutMul=2.0, maxFracWindow=0.5 -> max=200, v=800 -> clamped to 200
	result := a.adaptivePanelMinHeight(400, 2.0, 0.5, 50)
	if result > 200 {
		t.Errorf("expected <= 200 (clamped), got %f", result)
	}
}

func TestAdaptivePanelMinHeight_ClampedToMin(t *testing.T) {
	a := &App{}
	a.lastCanvasSize = fyne.NewSize(1920, 400)
	// base=10, layoutMul=0.1, maxFracWindow=0.5 -> v=1 -> clamped to minAbs=50
	result := a.adaptivePanelMinHeight(10, 0.1, 0.5, 50)
	if result < 50 {
		t.Errorf("expected >= 50 (clamped to min), got %f", result)
	}
}

// --- topology_controller.go (gui): applyTopology* nil-safe tests ---

func TestGuiApplyTopologyRunStart_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.applyTopologyRunStart()
}

func TestGuiApplyTopologyProgress_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.applyTopologyProgress("test", 0.5)
}

func TestGuiApplyTopologyCanceled_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.applyTopologyCanceled()
}

func TestGuiApplyTopologyFailure_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.applyTopologyFailure("snmp")
}

func TestGuiApplyTopologySuccess_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.applyTopologySuccess("ok", "preview", nil, nil, topologyBuildMetrics{})
}

func TestGuiApplyTopologyRunStart_WithUI(t *testing.T) {
	a := &App{}
	a.buildTopoBtn = widget.NewButton("Build", nil)
	a.stopTopoBtn = widget.NewButton("Stop", nil)
	a.saveTopoBtn = widget.NewButton("Save", nil)
	a.copyPerfBtn = widget.NewButton("Copy", nil)
	a.savePerfBtn = widget.NewButton("SavePerf", nil)
	a.refreshPreviewBtn = widget.NewButton("Refresh", nil)
	a.openPreviewBtn = widget.NewButton("Open", nil)
	a.statusLabel = widget.NewLabel("")
	a.topologyStatus = widget.NewLabel("")
	a.snmpStageLabel = widget.NewLabel("")
	a.snmpProgress = widget.NewProgressBar()
	a.applyTopologyRunStart()
	if a.statusLabel.Text != "Сбор SNMP данных и построение топологии..." {
		t.Errorf("expected status text, got %q", a.statusLabel.Text)
	}
}

func TestGuiApplyTopologyProgress_WithUI(t *testing.T) {
	a := &App{}
	a.statusLabel = widget.NewLabel("")
	a.topologyStatus = widget.NewLabel("")
	a.snmpStageLabel = widget.NewLabel("")
	a.snmpProgress = widget.NewProgressBar()
	a.applyTopologyProgress("progress", 0.5)
	if a.statusLabel.Text != "progress" {
		t.Errorf("expected 'progress', got %q", a.statusLabel.Text)
	}
}

func TestGuiApplyTopologyCanceled_WithUI(t *testing.T) {
	a := &App{}
	a.statusLabel = widget.NewLabel("")
	a.topologyStatus = widget.NewLabel("")
	a.snmpStageLabel = widget.NewLabel("")
	a.snmpProgress = widget.NewProgressBar()
	a.buildTopoBtn = widget.NewButton("Build", nil)
	a.stopTopoBtn = widget.NewButton("Stop", nil)
	a.applyTopologyCanceled()
	if a.statusLabel.Text == "" {
		t.Error("expected non-empty status")
	}
}

func TestGuiApplyTopologyFailure_WithUI(t *testing.T) {
	a := &App{}
	a.buildTopoBtn = widget.NewButton("Build", nil)
	a.stopTopoBtn = widget.NewButton("Stop", nil)
	a.topologyStatus = widget.NewLabel("")
	a.snmpStageLabel = widget.NewLabel("")
	a.snmpProgress = widget.NewProgressBar()
	a.applyTopologyFailure("snmp")
	if a.topologyStatus.Text != "Ошибка SNMP опроса" {
		t.Errorf("expected 'Ошибка SNMP опроса', got %q", a.topologyStatus.Text)
	}
}

func TestGuiApplyTopologyFailure_BuildStage(t *testing.T) {
	a := &App{}
	a.buildTopoBtn = widget.NewButton("Build", nil)
	a.stopTopoBtn = widget.NewButton("Stop", nil)
	a.topologyStatus = widget.NewLabel("")
	a.snmpStageLabel = widget.NewLabel("")
	a.snmpProgress = widget.NewProgressBar()
	a.applyTopologyFailure("build")
	if a.topologyStatus.Text != "Ошибка построения топологии" {
		t.Errorf("expected 'Ошибка построения топологии', got %q", a.topologyStatus.Text)
	}
}
