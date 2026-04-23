package gui

import (
	"strings"
	"testing"

	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func TestFilterPresetSaveApplyRoundtrip(t *testing.T) {
	a := &App{
		myApp:               fyneapp.New(),
		statusLabel:         widget.NewLabel(""),
		resultsFilterEnt:    widget.NewEntry(),
		resultsCidrFilterEnt: widget.NewEntry(),
		resultsPortStateSel: widget.NewSelect([]string{"Все", "Есть открытые", "Есть закрытые", "Есть фильтруемые"}, nil),
		openPortsOnlyCheck:  widget.NewCheck("", nil),
		quickTypeChecks: map[string]*widget.Check{
			"Network Device": widget.NewCheck("Network Device", nil),
			"Computer":       widget.NewCheck("Computer", nil),
			"Server":         widget.NewCheck("Server", nil),
			"Unknown":        widget.NewCheck("Unknown", nil),
		},
	}

	// Arrange current filters.
	a.resultsFilterQuery = "router"
	a.resultsFilterEnt.SetText("router")
	a.resultsCidrFilterEnt.SetText("192.168.1.0/24")
	a.resultsPortStateMode = "has_open"
	a.resultsPortStateSel.SetSelected("Есть открытые")
	a.onlyWithOpenPorts = true
	a.openPortsOnlyCheck.SetChecked(true)
	a.quickTypeChecks["Network Device"].SetChecked(true)
	a.quickTypeChecks["Server"].SetChecked(true)

	a.saveFilterPreset("1")

	// Reset values to ensure apply actually restores preset.
	a.resultsFilterQuery = ""
	a.resultsFilterEnt.SetText("")
	a.resultsCidrFilterEnt.SetText("")
	a.resultsPortStateMode = "all"
	a.resultsPortStateSel.SetSelected("Все")
	a.onlyWithOpenPorts = false
	a.openPortsOnlyCheck.SetChecked(false)
	for _, ch := range a.quickTypeChecks {
		ch.SetChecked(false)
	}

	a.applyFilterPreset("1")

	if strings.TrimSpace(a.resultsFilterQuery) != "router" {
		t.Fatalf("expected resultsFilterQuery restored to router, got %q", a.resultsFilterQuery)
	}
	if got := strings.TrimSpace(a.resultsCidrFilterEnt.Text); got != "192.168.1.0/24" {
		t.Fatalf("expected CIDR restored, got %q", got)
	}
	if a.resultsPortStateMode != "has_open" {
		t.Fatalf("expected port state mode has_open, got %q", a.resultsPortStateMode)
	}
	if !a.onlyWithOpenPorts {
		t.Fatalf("expected onlyWithOpenPorts restored to true")
	}
	if !a.quickTypeChecks["Network Device"].Checked || !a.quickTypeChecks["Server"].Checked {
		t.Fatalf("expected selected type checks restored")
	}
	if !strings.Contains(a.statusLabel.Text, "применен") {
		t.Fatalf("expected status to contain 'применен', got %q", a.statusLabel.Text)
	}
}
