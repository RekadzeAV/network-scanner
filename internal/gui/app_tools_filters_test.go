package gui

import (
	"testing"

	"fyne.io/fyne/v2/widget"
)

// --- app.go: setPortRangeControlsEnabled, withToolHost, setToolsOutputMarkdown,
//     setToolsButtonsEnabled, filterPresetKey, serializeCurrentFilters,
//     saveFilterPreset, applyFilterPreset ---

func TestSetPortRangeControlsEnabled_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.setPortRangeControlsEnabled(true)
}

func TestSetPortRangeControlsEnabled_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.setPortRangeControlsEnabled(true)
	a.setPortRangeControlsEnabled(false)
}

func TestSetPortRangeControlsEnabled_WithControls(t *testing.T) {
	a := &App{}
	a.portRangeEntry = widget.NewEntry()
	a.presetQuickBtn = widget.NewButton("Quick", nil)
	a.presetBalBtn = widget.NewButton("Bal", nil)
	a.presetDeepBtn = widget.NewButton("Deep", nil)
	a.setPortRangeControlsEnabled(false)
	if !a.portRangeEntry.Disabled() {
		t.Error("expected portRangeEntry disabled")
	}
	if !a.presetQuickBtn.Disabled() {
		t.Error("expected presetQuickBtn disabled")
	}
	a.setPortRangeControlsEnabled(true)
	if a.portRangeEntry.Disabled() {
		t.Error("expected portRangeEntry enabled")
	}
	if a.presetQuickBtn.Disabled() {
		t.Error("expected presetQuickBtn enabled")
	}
}

func TestWithToolHost_NilApp(t *testing.T) {
	var a *App
	_, ok := a.withToolHost()
	if ok {
		t.Error("expected ok=false for nil app")
	}
}

func TestWithToolHost_NilEntry(t *testing.T) {
	a := &App{}
	_, ok := a.withToolHost()
	if ok {
		t.Error("expected ok=false for nil entry")
	}
}

func TestWithToolHost_EmptyText(t *testing.T) {
	a := &App{}
	a.toolsHostEntry = widget.NewEntry()
	// dialog.ShowInformation требует активное окно — пропускаем в headless
	t.Skip("требует активное окно Fyne — не применимо в headless")
}

func TestSetToolsOutputMarkdown_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.setToolsOutputMarkdown("test")
}

func TestSetToolsOutputMarkdown_NilOutput(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.setToolsOutputMarkdown("test")
}

func TestSetToolsOutputMarkdown_WithOutput(t *testing.T) {
	a := &App{}
	a.toolsOutput = widget.NewRichText()
	a.setToolsOutputMarkdown("### test")
}

func TestSetToolsButtonsEnabled_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.setToolsButtonsEnabled(false)
	a.setToolsButtonsEnabled(true)
}

func TestSetToolsButtonsEnabled_WithButtons(t *testing.T) {
	a := &App{}
	a.toolsPingBtn = widget.NewButton("Ping", nil)
	a.toolsTraceBtn = widget.NewButton("Trace", nil)
	a.toolsDNSBtn = widget.NewButton("DNS", nil)
	a.setToolsButtonsEnabled(false)
	if !a.toolsPingBtn.Disabled() {
		t.Error("expected PingBtn disabled")
	}
	a.setToolsButtonsEnabled(true)
	if a.toolsPingBtn.Disabled() {
		t.Error("expected PingBtn enabled")
	}
}

func TestFilterPresetKey_Slot1(t *testing.T) {
	a := &App{}
	key := a.filterPresetKey("1")
	if key == "" {
		t.Error("expected non-empty key for slot 1")
	}
}

func TestFilterPresetKey_Slot2(t *testing.T) {
	a := &App{}
	key := a.filterPresetKey("2")
	if key == "" {
		t.Error("expected non-empty key for slot 2")
	}
}

func TestFilterPresetKey_Slot3(t *testing.T) {
	a := &App{}
	key := a.filterPresetKey("3")
	if key == "" {
		t.Error("expected non-empty key for slot 3")
	}
}

func TestFilterPresetKey_InvalidSlot(t *testing.T) {
	a := &App{}
	key := a.filterPresetKey("invalid")
	if key == "" {
		t.Error("expected non-empty key (fallback) for invalid slot")
	}
}

func TestSerializeCurrentFilters_EmptyApp(t *testing.T) {
	a := &App{}
	result := a.serializeCurrentFilters()
	if result == "" {
		t.Error("expected non-empty serialized filters")
	}
}

func TestSerializeCurrentFilters_WithFilters(t *testing.T) {
	a := &App{}
	a.resultsFilterQuery = "router"
	a.resultsPortStateMode = "has_open"
	a.onlyWithOpenPorts = true
	a.resultsCidrFilterEnt = widget.NewEntry()
	a.resultsCidrFilterEnt.SetText("192.168.1.0/24")
	a.quickTypeChecks = map[string]*widget.Check{
		"Router": {Checked: true},
	}
	result := a.serializeCurrentFilters()
	if result == "" {
		t.Error("expected non-empty serialized filters")
	}
}

func TestSaveFilterPreset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.saveFilterPreset("1")
}

func TestSaveFilterPreset_NilMyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.saveFilterPreset("1")
}

func TestApplyFilterPreset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.applyFilterPreset("1")
}

func TestApplyFilterPreset_NilMyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.applyFilterPreset("1")
}

func TestRunToolOperation_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.runToolOperation("test", "started", nil)
}

func TestStartOperationsWatcher_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.startOperationsWatcher()
}

func TestStartOperationsWatcher_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.startOperationsWatcher()
}
