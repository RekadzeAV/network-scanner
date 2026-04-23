package presenter

import (
	"fmt"

	"network-scanner/internal/scanner"
)

// XMLPresenter is a placeholder for XML export.
type XMLPresenter struct{}

func (p XMLPresenter) DisplayHeader() {}

func (p XMLPresenter) DisplayHost(host scanner.HostResult) {
	_ = host
}

func (p XMLPresenter) DisplaySummary(totalHosts int, openPortsCount int) {
	_, _ = totalHosts, openPortsCount
}

func (p XMLPresenter) Export(results []scanner.HostResult, format string) error {
	_, _ = results, format
	return fmt.Errorf("xml export is not implemented yet")
}
