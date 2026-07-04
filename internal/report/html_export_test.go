package report

import (
	"os"
	"path/filepath"
	"testing"

	"network-scanner/internal/contracts"
)

func TestRenderScanHTML(t *testing.T) {
	data := &ScanReportData{
		GeneratedAt: "2024-01-01T00:00:00Z",
		ScanID:      "test-scan-1",
		Network:     "192.168.1.0/24",
		HostCount:   2,
		Results: []ScanResultRow{
			{IP: "192.168.1.1", Hostname: "router", Ports: 3, OS: "Cisco IOS", Vendor: "Cisco"},
			{IP: "192.168.1.2", Hostname: "switch", Ports: 5, OS: "Linux", Vendor: "TP-Link"},
		},
		Findings: []SecurityFinding{
			{Severity: "HIGH", Title: "Open SSH", Description: "Close SSH or restrict access", HostIP: "192.168.1.1"},
		},
		Topology: &TopologySummary{
			DeviceCount: 2,
			LinkCount:   1,
			Devices: []TopologyDevice{
				{IP: "192.168.1.1", Hostname: "router", Vendor: "router"},
				{IP: "192.168.1.2", Hostname: "switch", Vendor: "switch"},
			},
		},
	}

	html, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("failed to render HTML: %v", err)
	}

	if len(html) == 0 {
		t.Fatal("expected non-empty HTML")
	}

	if !contains(string(html), "Network Scan Report") {
		t.Error("expected HTML to contain title")
	}

	if !contains(string(html), "192.168.1.1") {
		t.Error("expected HTML to contain IP")
	}
}

func TestSaveScanHTML(t *testing.T) {
	data := &ScanReportData{
		GeneratedAt: "2024-01-01T00:00:00Z",
		ScanID:      "test-scan-1",
		Network:     "192.168.1.0/24",
		HostCount:   1,
		Results: []ScanResultRow{
			{IP: "192.168.1.1", Hostname: "router", Ports: 3},
		},
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "report.html")

	err := SaveScanHTML(path, data)
	if err != nil {
		t.Fatalf("failed to save HTML: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("expected non-empty file")
	}
}

func TestGenerateScanReportData(t *testing.T) {
	results := []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "router", GuessOS: "Cisco", DeviceVendor: "Cisco"},
	}
	findings := []contracts.Finding{
		{Host: "192.168.1.1", Title: "Open SSH", Severity: "HIGH", Recommendation: "Close SSH"},
	}
	topology := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "router", Type: "router"},
		},
		Links: []*contracts.Link{},
	}

	data := GenerateScanReportData("scan-1", "192.168.1.0/24", results, findings, topology)

	if data.ScanID != "scan-1" {
		t.Errorf("expected scan ID 'scan-1', got '%s'", data.ScanID)
	}

	if data.HostCount != 1 {
		t.Errorf("expected 1 host, got %d", data.HostCount)
	}

	if len(data.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(data.Results))
	}

	if len(data.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(data.Findings))
	}

	if data.Topology == nil {
		t.Error("expected topology to be set")
	}
}

func TestDefaultHTMLReportOptions(t *testing.T) {
	opts := DefaultHTMLReportOptions()

	if !opts.IncludeScanResults {
		t.Error("expected IncludeScanResults to be true")
	}
	if !opts.IncludeSecurity {
		t.Error("expected IncludeSecurity to be true")
	}
	if !opts.IncludeTopology {
		t.Error("expected IncludeTopology to be true")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
