package presenter

import (
	"fmt"

	"network-scanner/internal/display"
	"network-scanner/internal/scanner"
)

// CLIPresenter prints scan results to stdout and supports export.
type CLIPresenter struct{}

// DisplayHeader prints a basic section header.
func (p CLIPresenter) DisplayHeader() {
	fmt.Println("=== Network Scanner Results ===")
}

// DisplayHost prints a single host in table-compatible way.
func (p CLIPresenter) DisplayHost(host scanner.HostResult) {
	display.DisplayResults([]scanner.Result{host})
}

// DisplaySummary prints brief scan totals.
func (p CLIPresenter) DisplaySummary(totalHosts int, openPortsCount int) {
	fmt.Printf("Hosts: %d, Open ports: %d\n", totalHosts, openPortsCount)
}

// Export writes results in text/json/csv format.
func (p CLIPresenter) Export(results []scanner.HostResult, format string) error {
	switch format {
	case "json":
		return display.SaveResultsToJSON(results, "scan-results.json")
	case "csv":
		return display.SaveResultsToCSV(results, "scan-results.csv")
	default:
		return display.SaveResultsToFile(results, "scan-results.txt")
	}
}
