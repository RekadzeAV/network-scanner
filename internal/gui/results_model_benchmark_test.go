package gui

import (
	"strconv"
	"testing"

	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

func BenchmarkFilteredSortedResultsLarge(b *testing.B) {
	data := makeBenchmarkResults(5000)
	a := &App{
		scanResults:          data,
		resultsSort:          "IP",
		resultsFilterQuery:   "host-1",
		onlyWithOpenPorts:    true,
		resultsPortStateMode: "has_open",
		resultsCidrFilterEnt: widget.NewEntry(),
		quickTypeChecks:      map[string]*widget.Check{},
	}
	a.resultsCidrFilterEnt.SetText("10.0.0.0/8")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.filteredSortedResults()
	}
}

func makeBenchmarkResults(n int) []scanner.Result {
	out := make([]scanner.Result, 0, n)
	for i := 1; i <= n; i++ {
		r := scanner.Result{
			Hostname:   "host-" + strconv.Itoa(i),
			IP:         "10.0." + strconv.Itoa(i/255) + "." + strconv.Itoa(i%255),
			MAC:        "aa:bb:cc:dd:ee:ff",
			DeviceType: "Router",
			Ports: []scanner.PortInfo{
				{Port: 22, Protocol: "tcp", State: "open", Service: "ssh"},
				{Port: 80, Protocol: "tcp", State: "closed", Service: "http"},
			},
		}
		out = append(out, r)
	}
	return out
}

