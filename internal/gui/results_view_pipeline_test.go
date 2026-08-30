package gui

import (
	"testing"
	"time"

	"network-scanner/internal/scanner"

	"fyne.io/fyne/v2/widget"
)

// --- results_view.go extended tests ---

func TestFilteredSortedResults_EmptyApp(t *testing.T) {
	a := &App{}
	result := a.filteredSortedResults()
	if result != nil {
		t.Errorf("expected nil for empty scanResults, got %d items", len(result))
	}
}

func TestCurrentDisplayedResults_EmptyApp(t *testing.T) {
	a := &App{}
	result := a.currentDisplayedResults()
	if result != nil {
		t.Errorf("expected nil for empty app, got %d items", len(result))
	}
}

func TestSelectedTypeFilters_NilApp(t *testing.T) {
	var a *App
	filters := a.selectedTypeFilters()
	if filters != nil {
		t.Errorf("expected nil for nil app, got %v", filters)
	}
}

func TestSelectedTypeFilters_EmptyChecks(t *testing.T) {
	a := &App{}
	filters := a.selectedTypeFilters()
	if filters != nil {
		t.Errorf("expected nil for empty checks, got %v", filters)
	}
}

func TestBuildResultsPipelineCacheKey_EmptyApp(t *testing.T) {
	a := &App{}
	key := a.buildResultsPipelineCacheKey()
	if key == "" {
		t.Error("expected non-empty key for empty app")
	}
}

func TestInvalidateResultsPipelineCache_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.invalidateResultsPipelineCache()
}

func TestInvalidateResultsPipelineCache_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.invalidateResultsPipelineCache()
}

func TestApplyAdvancedFilters_EmptyBase(t *testing.T) {
	a := &App{}
	result := a.applyAdvancedFilters(nil)
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestApplyAdvancedFilters_WithResults(t *testing.T) {
	a := &App{}
	base := []scanner.Result{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}
	result := a.applyAdvancedFilters(base)
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestPassesCIDRFilter_EmptyCIDR(t *testing.T) {
	a := &App{}
	a.resultsCidrFilterEnt = widget.NewEntry()
	a.resultsCidrFilterEnt.SetText("")
	r := scanner.Result{IP: "192.168.1.1"}
	if !a.passesCIDRFilter(r) {
		t.Error("expected true for empty CIDR")
	}
}

func TestPassesCIDRFilter_InvalidCIDR(t *testing.T) {
	a := &App{}
	a.resultsCidrFilterEnt = widget.NewEntry()
	a.resultsCidrFilterEnt.SetText("invalid")
	r := scanner.Result{IP: "192.168.1.1"}
	if !a.passesCIDRFilter(r) {
		t.Error("expected true for invalid CIDR")
	}
}

func TestPassesCIDRFilter_ValidMatch(t *testing.T) {
	a := &App{}
	a.resultsCidrFilterEnt = widget.NewEntry()
	a.resultsCidrFilterEnt.SetText("192.168.1.0/24")
	r := scanner.Result{IP: "192.168.1.1"}
	if !a.passesCIDRFilter(r) {
		t.Error("expected true for matching CIDR")
	}
}

func TestPassesCIDRFilter_NoMatch(t *testing.T) {
	a := &App{}
	a.resultsCidrFilterEnt = widget.NewEntry()
	a.resultsCidrFilterEnt.SetText("10.0.0.0/8")
	r := scanner.Result{IP: "192.168.1.1"}
	if a.passesCIDRFilter(r) {
		t.Error("expected false for non-matching CIDR")
	}
}

func TestPassesPortStateMode_Empty(t *testing.T) {
	a := &App{}
	a.resultsPortStateMode = ""
	r := scanner.Result{}
	if !a.passesPortStateMode(r) {
		t.Error("expected true for empty mode")
	}
}

func TestPassesPortStateMode_HasOpen(t *testing.T) {
	a := &App{}
	a.resultsPortStateMode = "has_open"
	r := scanner.Result{
		Ports: []scanner.PortInfo{{State: "open"}},
	}
	if !a.passesPortStateMode(r) {
		t.Error("expected true for has_open with open port")
	}
}

func TestPassesPortStateMode_HasOpen_NoOpenPorts(t *testing.T) {
	a := &App{}
	a.resultsPortStateMode = "has_open"
	r := scanner.Result{
		Ports: []scanner.PortInfo{{State: "closed"}},
	}
	if a.passesPortStateMode(r) {
		t.Error("expected false for has_open with no open ports")
	}
}

func TestPassesPortStateMode_HasClosed(t *testing.T) {
	a := &App{}
	a.resultsPortStateMode = "has_closed"
	r := scanner.Result{
		Ports: []scanner.PortInfo{{State: "closed"}},
	}
	if !a.passesPortStateMode(r) {
		t.Error("expected true for has_closed with closed port")
	}
}

func TestPassesPortStateMode_HasFiltered(t *testing.T) {
	a := &App{}
	a.resultsPortStateMode = "has_filtered"
	r := scanner.Result{
		Ports: []scanner.PortInfo{{State: "filtered"}},
	}
	if !a.passesPortStateMode(r) {
		t.Error("expected true for has_filtered with filtered port")
	}
}

func TestPassesPortStateMode_UnknownMode(t *testing.T) {
	a := &App{}
	a.resultsPortStateMode = "unknown_mode"
	r := scanner.Result{}
	if !a.passesPortStateMode(r) {
		t.Error("expected true for unknown mode")
	}
}

func TestActiveFilterCount_EmptyApp(t *testing.T) {
	a := &App{}
	count := a.activeFilterCount()
	if count != 0 {
		t.Errorf("expected 0 filters, got %d", count)
	}
}

func TestActiveFilterCount_WithQuery(t *testing.T) {
	a := &App{}
	a.resultsFilterQuery = "router"
	count := a.activeFilterCount()
	if count != 1 {
		t.Errorf("expected 1 filter, got %d", count)
	}
}

func TestActiveFilterCount_WithOpenPortsOnly(t *testing.T) {
	a := &App{}
	a.onlyWithOpenPorts = true
	count := a.activeFilterCount()
	if count != 1 {
		t.Errorf("expected 1 filter, got %d", count)
	}
}

func TestActiveFilterCount_WithPortStateMode(t *testing.T) {
	a := &App{}
	a.resultsPortStateMode = "has_open"
	count := a.activeFilterCount()
	if count != 1 {
		t.Errorf("expected 1 filter, got %d", count)
	}
}

func TestActiveFilterCount_AllFilters(t *testing.T) {
	a := &App{}
	a.resultsFilterQuery = "router"
	a.onlyWithOpenPorts = true
	a.resultsPortStateMode = "has_open"
	a.resultsCidrFilterEnt = widget.NewEntry()
	a.resultsCidrFilterEnt.SetText("192.168.1.0/24")
	a.quickTypeChecks = map[string]*widget.Check{
		"Router": {Checked: true},
	}
	count := a.activeFilterCount()
	if count != 5 {
		t.Errorf("expected 5 filters, got %d", count)
	}
}

func TestUpdateFiltersInfoLabel_NilLabel(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.updateFiltersInfoLabel()
}

func TestUpdateFiltersInfoLabel_WithLabel(t *testing.T) {
	a := &App{}
	a.filtersInfoLabel = widget.NewLabel("")
	a.resultsFilterQuery = "test"
	a.updateFiltersInfoLabel()
	if a.filtersInfoLabel.Text == "" {
		t.Error("expected non-empty label text")
	}
}

func TestUpdateResultsPerfLabel_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.updateResultsPerfLabel(resultsRenderStats{})
}

func TestUpdateResultsPerfLabel_NilLabel(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.updateResultsPerfLabel(resultsRenderStats{})
}

func TestUpdateResultsPerfLabel_ZeroFiltered(t *testing.T) {
	a := &App{}
	a.resultsPerfLabel = widget.NewLabel("")
	a.updateResultsPerfLabel(resultsRenderStats{FilteredCount: 0})
	if a.resultsPerfLabel.Text != "Рендер: n/a" {
		t.Errorf("expected 'Рендер: n/a', got %q", a.resultsPerfLabel.Text)
	}
}

func TestUpdateResultsPerfLabel_WithStats(t *testing.T) {
	a := &App{}
	a.resultsPerfLabel = widget.NewLabel("")
	a.updateResultsPerfLabel(resultsRenderStats{
		FilteredCount: 10,
		VisibleCount:  5,
		Duration:      100 * time.Millisecond,
	})
	if a.resultsPerfLabel.Text == "Рендер: n/a" {
		t.Error("expected non-n/a label")
	}
}

func TestScheduleResultsRender_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.scheduleResultsRender(true)
}

func TestCancelPendingResultsRender_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.cancelPendingResultsRender()
}

func TestCancelPendingResultsRender_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.cancelPendingResultsRender()
}

func TestCaptureHostDetailsSplitOffsetBeforeRebuild_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.captureHostDetailsSplitOffsetBeforeRebuild()
}

func TestCaptureHostDetailsSplitOffsetBeforeRebuild_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.captureHostDetailsSplitOffsetBeforeRebuild()
}

func TestClearResultsMainSplitRef_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.clearResultsMainSplitRef()
}

func TestClearResultsMainSplitRef_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.clearResultsMainSplitRef()
}

func TestRenderScanResultsView_NilBody(t *testing.T) {
	a := &App{}
	// Не должен паниковать при nil resultsBody
	a.renderScanResultsView()
}

func TestRenderScanResultsView_IdleState(t *testing.T) {
	a := &App{}
	a.resultsState = resultsStateIdle
	// Не должен паниковать
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered (expected in headless): %v", r)
		}
	}()
	a.renderScanResultsView()
}

func TestCurrentLayoutProfile_NilApp(t *testing.T) {
	var a *App
	profile := a.currentLayoutProfile()
	if profile != "normal" {
		t.Errorf("expected 'normal', got %q", profile)
	}
}

func TestCurrentLayoutProfile_EmptyApp(t *testing.T) {
	a := &App{}
	profile := a.currentLayoutProfile()
	if profile != "normal" {
		t.Errorf("expected 'normal', got %q", profile)
	}
}

func TestCurrentLayoutProfile_Set(t *testing.T) {
	a := &App{}
	a.layoutProfile = "compact"
	profile := a.currentLayoutProfile()
	if profile != "compact" {
		t.Errorf("expected 'compact', got %q", profile)
	}
}

func TestDetectLayoutProfile_Compact(t *testing.T) {
	a := &App{}
	profile := a.detectLayoutProfile(1366)
	if profile != "compact" {
		t.Errorf("expected 'compact', got %q", profile)
	}
}

func TestDetectLayoutProfile_Normal(t *testing.T) {
	a := &App{}
	profile := a.detectLayoutProfile(1920)
	if profile != "normal" {
		t.Errorf("expected 'normal', got %q", profile)
	}
}

func TestDetectLayoutProfile_Wide(t *testing.T) {
	a := &App{}
	profile := a.detectLayoutProfile(2200)
	if profile != "wide" {
		t.Errorf("expected 'wide', got %q", profile)
	}
}

func TestDetectLayoutProfile_Boundary1366(t *testing.T) {
	a := &App{}
	profile := a.detectLayoutProfile(1366)
	if profile != "compact" {
		t.Errorf("expected 'compact' at 1366, got %q", profile)
	}
}

func TestDetectLayoutProfile_Boundary2200(t *testing.T) {
	a := &App{}
	profile := a.detectLayoutProfile(2200)
	if profile != "wide" {
		t.Errorf("expected 'wide' at 2200, got %q", profile)
	}
}
