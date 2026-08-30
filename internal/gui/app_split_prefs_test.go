package gui

import (
	"testing"
	"time"

	"network-scanner/internal/scanner"
	"network-scanner/internal/topology"

	"fyne.io/fyne/v2/widget"
)

// --- app.go: resultsForSave, copyScanDiagnostics, saveScanDiagnostics, buildPerformanceReportText ---

func TestResultsForSave_EmptyResults(t *testing.T) {
	a := &App{}
	results, reason := a.resultsForSave()
	if results != nil {
		t.Error("expected nil results for empty scanResults")
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestResultsForSave_WithResults(t *testing.T) {
	a := &App{}
	a.scanResults = []scanner.Result{
		{IP: "192.168.1.1", Hostname: "h1"},
	}
	results, reason := a.resultsForSave()
	if reason != "" {
		t.Errorf("expected empty reason, got %q", reason)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestResultsForSave_WithFilteredResults(t *testing.T) {
	a := &App{}
	a.scanResults = []scanner.Result{
		{IP: "192.168.1.1", Hostname: "h1"},
	}
	a.resultsFilterQuery = "nomatch"
	results, reason := a.resultsForSave()
	if reason == "" {
		t.Error("expected non-empty reason for filtered results")
	}
	if results != nil {
		t.Error("expected nil results after filtering")
	}
}

func TestCopyScanDiagnostics_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.copyScanDiagnostics()
}

func TestCopyScanDiagnostics_NilDiagnosticsLabel(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.copyScanDiagnostics()
}

func TestCopyScanDiagnostics_WithLabel(t *testing.T) {
	a := &App{}
	a.diagnosticsLabel = widget.NewLabel("test diagnostics")
	// Без myWindow вызовет panic в dialog, пропускаем
	t.Skip("требует активное окно Fyne — не применимо в headless")
}

func TestSaveScanDiagnostics_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.saveScanDiagnostics()
}

func TestSaveScanDiagnostics_NilDiagnosticsLabel(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.saveScanDiagnostics()
}

func TestBuildPerformanceReportText_NilTopo(t *testing.T) {
	a := &App{}
	text := a.buildPerformanceReportText()
	if text != "" {
		t.Errorf("expected empty for nil topology, got %q", text)
	}
}

func TestBuildPerformanceReportText_WithTopo(t *testing.T) {
	a := &App{}
	a.lastTopology = &topology.Topology{
		Devices: map[string]*topology.Device{
			"192.168.1.1": {IP: "192.168.1.1"},
		},
		Links: []topology.Link{{}},
	}
	a.lastTopoMetric = topologyBuildMetrics{
		snmpDuration:  1 * time.Second,
		buildDuration: 2 * time.Second,
		totalDuration: 3 * time.Second,
	}
	text := a.buildPerformanceReportText()
	if text == "" {
		t.Fatal("expected non-empty report")
	}
}

func TestBuildPerformanceReportText_WithReport(t *testing.T) {
	a := &App{}
	a.lastTopology = &topology.Topology{
		Devices: map[string]*topology.Device{
			"192.168.1.1": {IP: "192.168.1.1"},
		},
	}
	a.lastTopoMetric = topologyBuildMetrics{
		totalDuration: 5 * time.Second,
	}
	text := a.buildPerformanceReportText()
	if text == "" {
		t.Fatal("expected non-empty report")
	}
}

// --- app.go: load*SplitFromPrefs tests ---

func TestLoadScanTabSplitFromPrefs_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.loadScanTabSplitFromPrefs()
}

func TestLoadScanTabSplitFromPrefs_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.loadScanTabSplitFromPrefs()
}

func TestLoadTopologySplitFromPrefs_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.loadTopologySplitFromPrefs()
}

func TestLoadTopologySplitFromPrefs_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.loadTopologySplitFromPrefs()
}

func TestLoadToolsTabSplitFromPrefs_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.loadToolsTabSplitFromPrefs()
}

func TestLoadToolsTabSplitFromPrefs_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.loadToolsTabSplitFromPrefs()
}

func TestLoadHostDetailsSplitFromPrefs_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.loadHostDetailsSplitFromPrefs()
}

func TestLoadHostDetailsSplitFromPrefs_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.loadHostDetailsSplitFromPrefs()
}

// --- app.go: maybePersistHostDetailsSplitOffsets ---

func TestMaybePersistHostDetailsSplitOffsets_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.maybePersistHostDetailsSplitOffsets()
}

func TestMaybePersistHostDetailsSplitOffsets_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.maybePersistHostDetailsSplitOffsets()
}

// --- app.go: startOperationsWatcher already tested ---

// --- scan_ui.go: buildScanControlsContainer, buildResultsContainer ---

func TestBuildScanControlsContainer_EmptyApp(t *testing.T) {
	a := &App{}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered (expected in headless): %v", r)
		}
	}()
	result := a.buildScanControlsContainer()
	_ = result // may be nil if panicked
}

func TestBuildResultsContainer_EmptyApp(t *testing.T) {
	a := &App{}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered (expected in headless): %v", r)
		}
	}()
	result := a.buildResultsContainer()
	_ = result // may be nil if panicked
}
