package controller

import (
	"testing"

	"network-scanner/internal/scanner"

	"fyne.io/fyne/v2/widget"
)

func TestNewResultsController_Created(t *testing.T) {
	rc := NewResultsController(nil, nil)
	if rc == nil {
		t.Fatal("expected non-nil ResultsController")
	}
}

func TestNewResultsController_WithApp(t *testing.T) {
	ui := &ResultsUI{}
	rc := NewResultsController(ensureApp(), ui)
	if rc == nil {
		t.Fatal("expected non-nil ResultsController")
	}
}

func TestResultsController_LoadSettings_NilApp(t *testing.T) {
	rc := &ResultsController{}
	rc.LoadSettings()
}

func TestResultsController_LoadSettings_NilUI(t *testing.T) {
	rc := &ResultsController{app: ensureApp()}
	rc.LoadSettings()
}

func TestResultsController_LoadSettings_WithUI(t *testing.T) {
	ui := &ResultsUI{
		ResultsModeSel:       widget.NewRadioGroup([]string{}, nil),
		ResultsSubModeSel:    widget.NewRadioGroup([]string{}, nil),
		ResultsSortSel:       widget.NewSelect([]string{}, nil),
		ChipLimitSel:         widget.NewSelect([]string{}, nil),
		ShowRawBannersCheck:  widget.NewCheck("", nil),
		ResultsFilterEnt:     widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
		QuickTypeChecks:      map[string]*widget.Check{},
		ResultsCidrFilterEnt: widget.NewEntry(),
		ResultsPortStateSel:  widget.NewSelect([]string{}, nil),
	}
	rc := NewResultsController(ensureApp(), ui)
	rc.LoadSettings()
}

func TestResultsController_SaveSettings_NilApp(t *testing.T) {
	rc := &ResultsController{}
	rc.SaveSettings()
}

func TestResultsController_SaveSettings_NilUI(t *testing.T) {
	rc := &ResultsController{app: ensureApp()}
	rc.SaveSettings()
}

func TestResultsController_SaveSettings_WithUI(t *testing.T) {
	ui := &ResultsUI{
		ResultsModeSel:       widget.NewRadioGroup([]string{"Таблица", "Карточки"}, nil),
		ResultsSubModeSel:    widget.NewRadioGroup([]string{"Devices", "Security", "Inventory"}, nil),
		ResultsSortSel:       widget.NewSelect([]string{"IP", "HostName"}, nil),
		ChipLimitSel:         widget.NewSelect([]string{"12", "24"}, nil),
		ShowRawBannersCheck:  widget.NewCheck("", nil),
		ResultsFilterEnt:     widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
		QuickTypeChecks:      map[string]*widget.Check{},
		ResultsCidrFilterEnt: widget.NewEntry(),
		ResultsPortStateSel:  widget.NewSelect([]string{}, nil),
	}
	rc := NewResultsController(ensureApp(), ui)
	rc.SaveSettings()
}

func TestResultsController_FilterResults_WithUI(t *testing.T) {
	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
	}
	ui.ResultsFilterEnt.Text = "router"
	ui.ResultsCidrFilterEnt.Text = ""
	ui.OpenPortsOnlyCheck.Checked = false
	rc := &ResultsController{ui: ui}
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router-main"},
		{IP: "192.168.1.2", Hostname: "server-1"},
		{IP: "10.0.0.1", Hostname: "router-backup"},
	}
	filtered := rc.FilterResults(results)
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered results, got %d", len(filtered))
	}
}

func TestResultsController_FilterResults_OpenPortsOnly(t *testing.T) {
	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
	}
	ui.ResultsFilterEnt.Text = ""
	ui.ResultsCidrFilterEnt.Text = ""
	ui.OpenPortsOnlyCheck.Checked = true
	rc := &ResultsController{ui: ui}
	results := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, State: "open"}}},
		{IP: "192.168.1.2", Ports: []scanner.PortInfo{{Port: 80, State: "closed"}}},
	}
	filtered := rc.FilterResults(results)
	if len(filtered) != 1 {
		t.Errorf("expected 1 open-port result, got %d", len(filtered))
	}
}

func TestResultsController_FilterResults_CIDR(t *testing.T) {
	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
	}
	ui.ResultsFilterEnt.Text = ""
	ui.ResultsCidrFilterEnt.Text = "192.168.1.0/24"
	ui.OpenPortsOnlyCheck.Checked = false
	rc := &ResultsController{ui: ui}
	results := []scanner.Result{
		{IP: "192.168.1.1"},
		{IP: "10.0.0.1"},
	}
	filtered := rc.FilterResults(results)
	if len(filtered) != 1 {
		t.Errorf("expected 1 CIDR-matched result, got %d", len(filtered))
	}
}

func TestResultsController_SortResults_IP(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.10"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.1"},
	}
	rc := &ResultsController{}
	rc.SortResults(results, "IP")
	if results[0].IP != "192.168.1.1" {
		t.Errorf("expected first IP '192.168.1.1', got %q", results[0].IP)
	}
}

func TestResultsController_SortResults_HostName(t *testing.T) {
	results := []scanner.Result{
		{Hostname: "zebra"},
		{Hostname: "alpha"},
		{Hostname: "beta"},
	}
	rc := &ResultsController{}
	rc.SortResults(results, "HostName")
	if results[0].Hostname != "alpha" {
		t.Errorf("expected first hostname 'alpha', got %q", results[0].Hostname)
	}
}

func TestResultsController_SortResults_Empty(t *testing.T) {
	rc := &ResultsController{}
	rc.SortResults(nil, "IP")
	rc.SortResults(nil, "HostName")
}
