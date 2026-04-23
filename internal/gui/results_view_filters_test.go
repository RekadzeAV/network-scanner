package gui

import (
	"testing"

	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

func TestCurrentDisplayedResultsCombinedFilters(t *testing.T) {
	a := &App{
		resultsSort:         "IP",
		resultsFilterQuery:  "router",
		onlyWithOpenPorts:   true,
		resultsPortStateMode: "has_open",
		resultsCidrFilterEnt: widget.NewEntry(),
		quickTypeChecks: map[string]*widget.Check{
			"Network Device": widget.NewCheck("Network Device", nil),
			"Computer":       widget.NewCheck("Computer", nil),
			"Server":         widget.NewCheck("Server", nil),
			"Unknown":        widget.NewCheck("Unknown", nil),
		},
		scanResults: []scanner.Result{
			{
				Hostname:   "router-main",
				IP:         "192.168.1.1",
				DeviceType: "Router",
				Ports: []scanner.PortInfo{
					{Port: 80, Protocol: "tcp", State: "open"},
					{Port: 22, Protocol: "tcp", State: "closed"},
				},
			},
			{
				Hostname:   "pc-1",
				IP:         "192.168.1.10",
				DeviceType: "Windows Computer",
				Ports: []scanner.PortInfo{
					{Port: 445, Protocol: "tcp", State: "open"},
				},
			},
			{
				Hostname:   "router-lab",
				IP:         "10.0.0.1",
				DeviceType: "Router",
				Ports: []scanner.PortInfo{
					{Port: 80, Protocol: "tcp", State: "open"},
				},
			},
		},
	}

	a.resultsCidrFilterEnt.SetText("192.168.1.0/24")
	a.quickTypeChecks["Network Device"].SetChecked(true)

	got := a.currentDisplayedResults()
	if len(got) != 1 {
		t.Fatalf("expected 1 filtered result, got %d", len(got))
	}
	if got[0].Hostname != "router-main" {
		t.Fatalf("expected router-main, got %s", got[0].Hostname)
	}
}
