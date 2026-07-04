package gui

import (
	"os"
	"strings"
	"testing"

	"network-scanner/internal/scanner"

	"fyne.io/fyne/v2/widget"
)

// --- filteredSortedResults & cache ---

func TestFilteredSortedResults_Empty(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	res := app.filteredSortedResults()
	if res != nil {
		t.Fatalf("expected nil for empty results, got %d", len(res))
	}
}

func TestFilteredSortedResults_CacheHit(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.scanResults = []scanner.Result{
		{IP: "192.168.1.1", Hostname: "host1"},
	}
	app.resultsSort = "IP"
	app.resultsFilterQuery = ""
	app.onlyWithOpenPorts = false

	// First call populates cache
	res1 := app.filteredSortedResults()
	if len(res1) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res1))
	}

	// Second call should use cache
	res2 := app.filteredSortedResults()
	if len(res2) != 1 {
		t.Fatal("expected cache hit to return same length")
	}
}

func TestFilteredSortedResults_FilterByPortState(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.scanResults = []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, State: "open"}}},
		{IP: "192.168.1.2", Ports: []scanner.PortInfo{{Port: 80, State: "closed"}}},
		{IP: "192.168.1.3", Ports: []scanner.PortInfo{{Port: 80, State: "filtered"}}},
	}
	app.resultsPortStateMode = "has_open"
	app.resultsSort = "IP"

	res := app.filteredSortedResults()
	if len(res) != 1 || res[0].IP != "192.168.1.1" {
		t.Fatalf("expected 1 open host, got %d: %v", len(res), res)
	}
}

// --- selectedTypeFilters ---

func TestSelectedTypeFilters_Empty(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	filters := app.selectedTypeFilters()
	if filters != nil && len(filters) != 0 {
		t.Fatalf("expected empty filters, got %v", filters)
	}
}

func TestSelectedTypeFilters_WithChecks(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.quickTypeChecks = map[string]*widget.Check{
		"Server": {Checked: true},
		"Router": {Checked: false},
		"Switch": {Checked: true},
	}
	filters := app.selectedTypeFilters()
	if len(filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(filters))
	}
	if !containsStr(filters, "Server") || !containsStr(filters, "Switch") {
		t.Fatalf("expected Server and Switch in filters, got %v", filters)
	}
}

// --- buildResultsPipelineCacheKey ---

func TestBuildResultsPipelineCacheKey(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.scanResultsVersion = 123
	app.resultsFilterQuery = "test"
	app.resultsSort = "IP"
	app.resultsPortStateMode = "has_open"
	app.onlyWithOpenPorts = true
	app.resultsCidrFilterEnt = &widget.Entry{Text: "192.168.1.0/24"}

	key := app.buildResultsPipelineCacheKey()
	if !strings.Contains(key, "123") || !strings.Contains(key, "test") {
		t.Fatalf("unexpected cache key: %s", key)
	}
}

func TestBuildResultsPipelineCacheKey_NilApp(t *testing.T) {
	var app *App
	key := app.buildResultsPipelineCacheKey()
	if key != "" {
		t.Fatalf("expected empty key for nil app, got %s", key)
	}
}

// --- applyAdvancedFilters ---

func TestApplyAdvancedFilters_CIDR(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.resultsCidrFilterEnt = &widget.Entry{Text: "192.168.1.0/24"}

	base := []scanner.Result{
		{IP: "192.168.1.5"},
		{IP: "10.0.0.5"},
	}

	res := app.applyAdvancedFilters(base)
	if len(res) != 1 || res[0].IP != "192.168.1.5" {
		t.Fatalf("expected 1 result in CIDR, got %d: %v", len(res), res)
	}
}

func TestApplyAdvancedFilters_PortState(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.resultsPortStateMode = "has_closed"

	base := []scanner.Result{
		{IP: "192.168.1.5", Ports: []scanner.PortInfo{{Port: 80, State: "closed"}}},
		{IP: "192.168.1.6", Ports: []scanner.PortInfo{{Port: 80, State: "open"}}},
	}

	res := app.applyAdvancedFilters(base)
	if len(res) != 1 || res[0].IP != "192.168.1.5" {
		t.Fatalf("expected 1 closed host, got %d: %v", len(res), res)
	}
}

// --- passesCIDRFilter ---

func TestPassesCIDRFilter_Valid(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.resultsCidrFilterEnt = &widget.Entry{Text: "192.168.1.0/24"}

	r := scanner.Result{IP: "192.168.1.100"}
	if !app.passesCIDRFilter(r) {
		t.Fatal("expected host to pass CIDR filter")
	}

	r2 := scanner.Result{IP: "10.0.0.5"}
	if app.passesCIDRFilter(r2) {
		t.Fatal("expected host to fail CIDR filter")
	}
}

func TestPassesCIDRFilter_Empty(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.resultsCidrFilterEnt = &widget.Entry{Text: ""}

	r := scanner.Result{IP: "10.0.0.5"}
	if !app.passesCIDRFilter(r) {
		t.Fatal("expected host to pass when CIDR is empty")
	}
}

func TestPassesCIDRFilter_NilEntry(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.resultsCidrFilterEnt = nil

	r := scanner.Result{IP: "10.0.0.5"}
	if !app.passesCIDRFilter(r) {
		t.Fatal("expected host to pass when CIDR entry is nil")
	}
}

// --- passesPortStateMode ---

func TestPassesPortStateMode_Open(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.resultsPortStateMode = "has_open"

	r := scanner.Result{Ports: []scanner.PortInfo{{Port: 80, State: "open"}}}
	if !app.passesPortStateMode(r) {
		t.Fatal("expected host to pass open filter")
	}
}

func TestPassesPortStateMode_Closed(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.resultsPortStateMode = "has_closed"

	r := scanner.Result{Ports: []scanner.PortInfo{{Port: 80, State: "closed"}}}
	if !app.passesPortStateMode(r) {
		t.Fatal("expected host to pass closed filter")
	}
}

func TestPassesPortStateMode_Filtered(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.resultsPortStateMode = "has_filtered"

	r := scanner.Result{Ports: []scanner.PortInfo{{Port: 80, State: "filtered"}}}
	if !app.passesPortStateMode(r) {
		t.Fatal("expected host to pass filtered filter")
	}
}

func TestPassesPortStateMode_All(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()
	app.resultsPortStateMode = "all"

	r := scanner.Result{Ports: []scanner.PortInfo{{Port: 80, State: "closed"}}}
	if !app.passesPortStateMode(r) {
		t.Fatal("expected host to pass 'all' mode")
	}
}

// --- activeFilterCount ---

func TestActiveFilterCount(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()

	if app.activeFilterCount() != 0 {
		t.Fatalf("expected 0 active filters, got %d", app.activeFilterCount())
	}

	app.resultsFilterQuery = "test"
	if app.activeFilterCount() != 1 {
		t.Fatalf("expected 1 active filter, got %d", app.activeFilterCount())
	}

	app.quickTypeChecks = map[string]*widget.Check{"Server": {Checked: true}}
	if app.activeFilterCount() != 2 {
		t.Fatalf("expected 2 active filters, got %d", app.activeFilterCount())
	}

	app.onlyWithOpenPorts = true
	if app.activeFilterCount() != 3 {
		t.Fatalf("expected 3 active filters, got %d", app.activeFilterCount())
	}
}

// --- truncateStr ---

func TestTruncateStr(t *testing.T) {
	if truncateStr("hello", 10) != "hello" {
		t.Fatal("expected full string when n > len")
	}
	if truncateStr("hello world", 5) != "he..." {
		t.Fatalf("expected truncated string, got %s", truncateStr("hello world", 5))
	}
	if truncateStr("hi", 3) != "hi" {
		t.Fatal("expected full string when n > len")
	}
}

// --- clamp functions ---

func TestClampFloat32(t *testing.T) {
	if clampFloat32(10, 0, 5) != 5 {
		t.Fatal("expected clamped to max")
	}
	if clampFloat32(-10, 0, 5) != 0 {
		t.Fatal("expected clamped to min")
	}
	if clampFloat32(3, 0, 5) != 3 {
		t.Fatal("expected unchanged")
	}
}

// --- layout helpers ---

func TestCurrentLayoutProfile(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()

	if app.currentLayoutProfile() != "normal" {
		t.Fatalf("expected 'normal', got %s", app.currentLayoutProfile())
	}

	app.layoutProfile = "compact"
	if app.currentLayoutProfile() != "compact" {
		t.Fatalf("expected 'compact', got %s", app.currentLayoutProfile())
	}
}

func TestLayoutAdaptiveMultiplier(t *testing.T) {
	mul := layoutAdaptiveMultiplier(1920, 1080, 1)
	if mul <= 0 || mul > 2 {
		t.Fatalf("unexpected multiplier: %f", mul)
	}

	mulZero := layoutAdaptiveMultiplier(0, 0, 1)
	if mulZero != 1 {
		t.Fatalf("expected 1 for zero dimensions, got %f", mulZero)
	}
}

// --- resultsTableColumnWidths ---

func TestResultsTableColumnWidths(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()

	widths := app.resultsTableColumnWidths()
	if len(widths) != 8 {
		t.Fatalf("expected 8 columns, got %d", len(widths))
	}

	app.layoutProfile = "wide"
	wideWidths := app.resultsTableColumnWidths()
	if len(wideWidths) != 8 {
		t.Fatal("expected 8 columns in wide mode")
	}

	app.layoutProfile = "compact"
	compactWidths := app.resultsTableColumnWidths()
	if len(compactWidths) != 8 {
		t.Fatal("expected 8 columns in compact mode")
	}
}

// --- resultsTableHeaders ---

func TestResultsTableHeaders(t *testing.T) {
	os.Setenv("FYNE_SCALE", "1")
	app := NewApp()

	headers := app.resultsTableHeaders()
	if len(headers) != 8 {
		t.Fatalf("expected 8 headers, got %d", len(headers))
	}

	app.layoutProfile = "compact"
	compactHeaders := app.resultsTableHeaders()
	if len(compactHeaders) != 8 {
		t.Fatal("expected 8 headers in compact mode")
	}
}

// --- osGuessLine ---

func TestOsGuessLine(t *testing.T) {
	r := scanner.Result{
		GuessOS:           "Linux",
		GuessOSConfidence: "95%",
		GuessOSReason:     "TCP/IP stack",
	}
	line := osGuessLine(r)
	if !strings.Contains(line, "Linux") || !strings.Contains(line, "95%") {
		t.Fatalf("unexpected OS line: %s", line)
	}

	r2 := scanner.Result{}
	line2 := osGuessLine(r2)
	if line2 != "-" {
		t.Fatalf("expected '-', got %s", line2)
	}
}

// --- nullDash ---

func TestNullDash(t *testing.T) {
	if nullDash("") != "-" {
		t.Fatal("expected '-' for empty")
	}
	if nullDash("  ") != "-" {
		t.Fatal("expected '-' for whitespace")
	}
	if nullDash("test") != "test" {
		t.Fatal("expected 'test' for valid string")
	}
}

// --- deviceTypeWithBadge ---

func TestDeviceTypeWithBadge(t *testing.T) {
	if !strings.HasPrefix(deviceTypeWithBadge("Router"), "[NET]") {
		t.Fatal("expected [NET] prefix for Router")
	}
	if !strings.HasPrefix(deviceTypeWithBadge("Server"), "[SRV]") {
		t.Fatal("expected [SRV] prefix for Server")
	}
	if !strings.HasPrefix(deviceTypeWithBadge("Desktop"), "[PC]") {
		t.Fatal("expected [PC] prefix for Desktop")
	}
	if deviceTypeWithBadge("") != "-" {
		t.Fatal("expected '-' for empty type")
	}
}

// --- countOpenPorts ---

func TestCountOpenPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, State: "open"},
		{Port: 443, State: "open"},
		{Port: 22, State: "closed"},
	}
	if countOpenPorts(ports) != 2 {
		t.Fatalf("expected 2 open ports, got %d", countOpenPorts(ports))
	}

	if countOpenPorts(nil) != 0 {
		t.Fatal("expected 0 for nil ports")
	}
}

// --- helper ---

func containsStr(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
