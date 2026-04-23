package presenter

import (
	"fmt"

	"network-scanner/internal/scanner"
)

// HTMLPresenter is a placeholder for HTML export.
type HTMLPresenter struct{}

func (p HTMLPresenter) DisplayHeader() {}

func (p HTMLPresenter) DisplayHost(host scanner.HostResult) {
	_ = host
}

func (p HTMLPresenter) DisplaySummary(totalHosts int, openPortsCount int) {
	_, _ = totalHosts, openPortsCount
}

func (p HTMLPresenter) Export(results []scanner.HostResult, format string) error {
	_, _ = results, format
	return fmt.Errorf("html export is not implemented yet")
}
