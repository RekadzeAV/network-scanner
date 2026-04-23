package presenter

import "network-scanner/internal/scanner"

// JSONPresenter exports scan results in JSON format.
type JSONPresenter struct{}

// DisplayHeader is a no-op for JSON presenter.
func (p JSONPresenter) DisplayHeader() {}

// DisplayHost is a no-op for JSON presenter.
func (p JSONPresenter) DisplayHost(host scanner.HostResult) { _ = host }

// DisplaySummary is a no-op for JSON presenter.
func (p JSONPresenter) DisplaySummary(totalHosts int, openPortsCount int) {
	_, _ = totalHosts, openPortsCount
}

// Export saves scan results to a JSON file.
func (p JSONPresenter) Export(results []scanner.HostResult, format string) error {
	_, _ = results, format
	return CLIPresenter{}.Export(results, "json")
}
