package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

func TestFormatResultsForDisplayEmpty(t *testing.T) {
	result := FormatResultsForDisplay(nil)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestFormatResultsForDisplayWithResults(t *testing.T) {
	results := []scanner.Result{
		{
			IP:       "192.168.1.1",
			Hostname: "router",
			MAC:      "aa:bb:cc:dd:ee:ff",
			Ports: []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", Service: "http", State: "open"},
			},
		},
	}

	result := FormatResultsForDisplay(results)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestFormatPortsEmpty(t *testing.T) {
	result := formatPorts(nil)
	if result != "" {
		t.Errorf("expected empty, got: %s", result)
	}
}

func TestFormatPortsOpen(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, Protocol: "tcp", Service: "http", State: "open"},
	}
	result := formatPorts(ports)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestEscapeMarkdownBasic(t *testing.T) {
	result := escapeMarkdown("hello|world")
	if result != "hello\\|world" {
		t.Errorf("expected 'hello\\|world', got: %s", result)
	}
}

func TestTruncateString(t *testing.T) {
	result := truncateString("hello world", 8)
	if result != "hello..." {
		t.Errorf("expected 'hello...', got: %s", result)
	}
}
