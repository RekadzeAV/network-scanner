package gui

import (
	"testing"
)

// --- app.go: ApplyScanRunStart, ConfirmLargeScanBypass, SetConfirmLargeScanBypass, RenderScanResultsView ---

func TestApplyScanRunStart_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать (пустая реализация)
	a.ApplyScanRunStart("test")
}

func TestConfirmLargeScanBypass_Default(t *testing.T) {
	a := &App{}
	if a.ConfirmLargeScanBypass() {
		t.Error("expected false by default")
	}
}

func TestSetConfirmLargeScanBypass_True(t *testing.T) {
	a := &App{}
	a.SetConfirmLargeScanBypass(true)
	if !a.ConfirmLargeScanBypass() {
		t.Error("expected true after SetConfirmLargeScanBypass(true)")
	}
}

func TestSetConfirmLargeScanBypass_False(t *testing.T) {
	a := &App{}
	a.SetConfirmLargeScanBypass(true)
	a.SetConfirmLargeScanBypass(false)
	if a.ConfirmLargeScanBypass() {
		t.Error("expected false after SetConfirmLargeScanBypass(false)")
	}
}

func TestRenderScanResultsView_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать при nil resultsBody
	a.RenderScanResultsView()
}

// --- app.go: clamp*SplitOffset tests ---

func TestClampScanTabMainSplitOffset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.clampScanTabMainSplitOffset()
}

func TestClampScanTabMainSplitOffset_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.clampScanTabMainSplitOffset()
}

func TestClampTopologyMainSplitOffset_NilApp(t *testing.T) {
	var a *App
	a.clampTopologyMainSplitOffset()
}

func TestClampTopologyMainSplitOffset_EmptyApp(t *testing.T) {
	a := &App{}
	a.clampTopologyMainSplitOffset()
}

func TestClampToolsTabMainSplitOffset_NilApp(t *testing.T) {
	var a *App
	a.clampToolsTabMainSplitOffset()
}

func TestClampToolsTabMainSplitOffset_EmptyApp(t *testing.T) {
	a := &App{}
	a.clampToolsTabMainSplitOffset()
}

// --- app.go: maybePersist*SplitOffset tests ---

func TestMaybePersistScanTabSplitOffset_NilApp(t *testing.T) {
	var a *App
	a.maybePersistScanTabSplitOffset()
}

func TestMaybePersistScanTabSplitOffset_EmptyApp(t *testing.T) {
	a := &App{}
	a.maybePersistScanTabSplitOffset()
}

func TestMaybePersistTopologySplitOffset_NilApp(t *testing.T) {
	var a *App
	a.maybePersistTopologySplitOffset()
}

func TestMaybePersistTopologySplitOffset_EmptyApp(t *testing.T) {
	a := &App{}
	a.maybePersistTopologySplitOffset()
}

func TestMaybePersistToolsTabSplitOffset_NilApp(t *testing.T) {
	var a *App
	a.maybePersistToolsTabSplitOffset()
}

func TestMaybePersistToolsTabSplitOffset_EmptyApp(t *testing.T) {
	a := &App{}
	a.maybePersistToolsTabSplitOffset()
}

// --- app.go: resetUIPanelLayout ---

func TestResetUIPanelLayout_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.resetUIPanelLayout()
}

func TestResetUIPanelLayout_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.resetUIPanelLayout()
}

// --- results_view.go: applyAdaptiveToolsScrollMinSizes ---

func TestApplyAdaptiveToolsScrollMinSizes_NilApp(t *testing.T) {
	var a *App
	a.applyAdaptiveToolsScrollMinSizes("normal", 1.0)
}

func TestApplyAdaptiveToolsScrollMinSizes_EmptyApp(t *testing.T) {
	a := &App{}
	a.applyAdaptiveToolsScrollMinSizes("normal", 1.0)
}

func TestApplyAdaptiveToolsScrollMinSizes_Compact(t *testing.T) {
	a := &App{}
	a.applyAdaptiveToolsScrollMinSizes("compact", 0.8)
}

func TestApplyAdaptiveToolsScrollMinSizes_Wide(t *testing.T) {
	a := &App{}
	a.applyAdaptiveToolsScrollMinSizes("wide", 1.2)
}

// --- results_view.go: applyDefaultSplitOffsetsForProfile ---

func TestApplyDefaultSplitOffsetsForProfile_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.applyDefaultSplitOffsetsForProfile("normal")
}

func TestApplyDefaultSplitOffsetsForProfile_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.applyDefaultSplitOffsetsForProfile("normal")
}

func TestApplyDefaultSplitOffsetsForProfile_Compact(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.applyDefaultSplitOffsetsForProfile("compact")
}

// --- results_view.go: startResultsLayoutWatcher ---

func TestStartResultsLayoutWatcher_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.startResultsLayoutWatcher()
}

func TestStartResultsLayoutWatcher_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.startResultsLayoutWatcher()
}
