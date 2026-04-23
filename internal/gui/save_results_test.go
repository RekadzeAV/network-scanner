package gui

import (
	"testing"

	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

func TestResultsForSaveNoScanResults(t *testing.T) {
	a := &App{}
	got, reason := a.resultsForSave()
	if len(got) != 0 {
		t.Fatalf("expected no results, got %d", len(got))
	}
	if reason != "Нет результатов для сохранения" {
		t.Fatalf("unexpected reason: %q", reason)
	}
}

func TestResultsForSaveEmptyAfterFilters(t *testing.T) {
	a := &App{
		scanResults: []scanner.Result{
			{
				Hostname:   "router-main",
				IP:         "192.168.1.1",
				DeviceType: "Router",
				Ports: []scanner.PortInfo{
					{Port: 80, Protocol: "tcp", State: "open"},
				},
			},
		},
		resultsFilterQuery:  "nomatch",
		resultsSort:         "IP",
		resultsPortStateMode: "all",
		resultsCidrFilterEnt: widget.NewEntry(),
		quickTypeChecks:      map[string]*widget.Check{},
	}
	got, reason := a.resultsForSave()
	if len(got) != 0 {
		t.Fatalf("expected no results after filters, got %d", len(got))
	}
	if reason != "После применения фильтров нет данных для сохранения" {
		t.Fatalf("unexpected reason: %q", reason)
	}
}

func TestResultsForSaveSuccess(t *testing.T) {
	a := &App{
		scanResults: []scanner.Result{
			{
				Hostname:   "router-main",
				IP:         "192.168.1.1",
				DeviceType: "Router",
				Ports: []scanner.PortInfo{
					{Port: 80, Protocol: "tcp", State: "open"},
				},
			},
		},
		resultsSort:          "IP",
		resultsPortStateMode: "all",
		resultsCidrFilterEnt: widget.NewEntry(),
		quickTypeChecks:      map[string]*widget.Check{},
	}
	got, reason := a.resultsForSave()
	if reason != "" {
		t.Fatalf("expected empty reason, got %q", reason)
	}
	if len(got) != 1 || got[0].Hostname != "router-main" {
		t.Fatalf("unexpected results payload: %#v", got)
	}
}
