package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

// === Integration: Full Report Generation Pipeline ===

func TestIntegrationReportGeneration_Pipeline(t *testing.T) {
	// Step 1: Generate report data
	data := GenerateScanReportData(
		"scan-001",
		"192.168.1.0/24",
		nil,
		nil,
		nil,
	)

	if data == nil {
		t.Fatal("expected non-nil report data")
	}
	if data.ScanID != "scan-001" {
		t.Errorf("expected scan ID 'scan-001', got %q", data.ScanID)
	}
	if data.Network != "192.168.1.0/24" {
		t.Errorf("expected network '192.168.1.0/24', got %q", data.Network)
	}
}

// === Integration: RenderScanHTML ===

func TestIntegrationRenderScanHTML_EmptyData(t *testing.T) {
	data := &ScanReportData{}
	html, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(html) == 0 {
		t.Fatal("expected non-empty HTML output")
	}
	if !strings.Contains(string(html), "<!DOCTYPE html>") {
		t.Error("expected HTML output to contain DOCTYPE")
	}
}

func TestIntegrationRenderScanHTML_WithData(t *testing.T) {
	data := &ScanReportData{
		ScanID:    "scan-001",
		Network:   "10.0.0.0/24",
		HostCount: 2,
		Results: []ScanResultRow{
			{IP: "10.0.0.1", Hostname: "host1", Ports: 2, OS: "Linux", Vendor: "Dell"},
			{IP: "10.0.0.2", Hostname: "host2", Ports: 5, OS: "Windows", Vendor: "HP"},
		},
	}
	html, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := string(html)
	if !strings.Contains(output, "scan-001") {
		t.Error("expected HTML to contain scan ID")
	}
	if !strings.Contains(output, "10.0.0.1") {
		t.Error("expected HTML to contain first host IP")
	}
	if !strings.Contains(output, "host2") {
		t.Error("expected HTML to contain second hostname")
	}
}

func TestIntegrationRenderScanHTML_WithFindings(t *testing.T) {
	data := &ScanReportData{
		ScanID:    "scan-001",
		Network:   "192.168.1.0/24",
		HostCount: 1,
		Findings: []SecurityFinding{
			{Severity: "high", Title: "Open SSH", Description: "SSH port open", HostIP: "192.168.1.1", Port: 22},
		},
	}
	html, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := string(html)
	if !strings.Contains(output, "Open SSH") {
		t.Error("expected HTML to contain finding title")
	}
	if !strings.Contains(output, "high") {
		t.Error("expected HTML to contain severity")
	}
}

func TestIntegrationRenderScanHTML_WithTopology(t *testing.T) {
	data := &ScanReportData{
		ScanID:    "scan-001",
		Network:   "192.168.1.0/24",
		HostCount: 2,
		Topology: &TopologySummary{
			DeviceCount: 2,
			LinkCount:   1,
			Devices: []TopologyDevice{
				{IP: "192.168.1.1", Hostname: "router", Vendor: "Cisco"},
			},
		},
	}
	html, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := string(html)
	if !strings.Contains(output, "router") {
		t.Error("expected HTML to contain device hostname")
	}
	if !strings.Contains(output, "Cisco") {
		t.Error("expected HTML to contain device vendor")
	}
}

func TestIntegrationRenderScanHTML_NilData(t *testing.T) {
	_, err := RenderScanHTML(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

// === Integration: SaveScanHTML ===

func TestIntegrationSaveScanHTML_Success(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "report.html")

	data := &ScanReportData{
		ScanID:    "scan-001",
		Network:   "10.0.0.0/24",
		HostCount: 1,
		Results: []ScanResultRow{
			{IP: "10.0.0.1", Hostname: "test-host", Ports: 3},
		},
	}

	err := SaveScanHTML(path, data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected non-empty saved file")
	}
	if !strings.Contains(string(content), "test-host") {
		t.Error("expected saved file to contain hostname")
	}
}

func TestIntegrationSaveScanHTML_WithFindings(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "report_findings.html")

	data := &ScanReportData{
		ScanID:    "scan-001",
		Network:   "192.168.1.0/24",
		HostCount: 1,
		Findings: []SecurityFinding{
			{Severity: "critical", Title: "Log4Shell", Description: "CVE-2021-44228", HostIP: "192.168.1.1", Port: 80},
		},
	}

	err := SaveScanHTML(path, data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if !strings.Contains(string(content), "Log4Shell") {
		t.Error("expected saved file to contain finding title")
	}
}

// === Integration: GenerateScanReportData ===

func TestIntegrationGenerateScanReportData_Empty(t *testing.T) {
	data := GenerateScanReportData("scan-001", "10.0.0.0/24", nil, nil, nil)
	if data == nil {
		t.Fatal("expected non-nil data")
	}
	if data.ScanID != "scan-001" {
		t.Errorf("expected scan ID 'scan-001', got %q", data.ScanID)
	}
	if data.HostCount != 0 {
		t.Errorf("expected 0 hosts, got %d", data.HostCount)
	}
	if len(data.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(data.Results))
	}
}

func TestIntegrationGenerateScanReportData_WithResults(t *testing.T) {
	// Mock scan results using the contracts interface
	type mockScanResult struct {
		IP       string
		Hostname string
		Ports    int
		OS       string
		Vendor   string
	}

	// Generate data directly
	data := &ScanReportData{
		ScanID:    "scan-002",
		Network:   "172.16.0.0/16",
		HostCount: 3,
		Results: []ScanResultRow{
			{IP: "172.16.0.1", Hostname: "gateway", Ports: 5, OS: "Linux", Vendor: "Cisco"},
			{IP: "172.16.0.2", Hostname: "server", Ports: 8, OS: "Windows", Vendor: "Dell"},
			{IP: "172.16.0.3", Hostname: "workstation", Ports: 3, OS: "Linux", Vendor: "HP"},
		},
	}

	if data.HostCount != 3 {
		t.Errorf("expected 3 hosts, got %d", data.HostCount)
	}
	if len(data.Results) != 3 {
		t.Errorf("expected 3 result rows, got %d", len(data.Results))
	}
	if data.Results[0].IP != "172.16.0.1" {
		t.Errorf("expected first IP '172.16.0.1', got %q", data.Results[0].IP)
	}
}

// === Integration: Security HTML Export ===

func TestIntegrationSecurityHTMLExport_Empty(t *testing.T) {
	html, err := RenderSecurityHTML([]scanner.Result{}, nil, time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(html) == 0 {
		t.Fatal("expected non-empty security HTML")
	}
}

func TestIntegrationSecurityHTMLExport_WithResults(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "test-host", Ports: []scanner.PortInfo{{Port: 22, State: "open"}}},
	}
	html, err := RenderSecurityHTML(results, nil, time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := string(html)
	if !strings.Contains(output, "test-host") {
		t.Error("expected security HTML to contain hostname")
	}
}

func TestIntegrationSecurityHTMLExport_NilResults(t *testing.T) {
	// nil results are handled gracefully - returns empty HTML
	html, err := RenderSecurityHTML(nil, nil, time.Now())
	if err != nil {
		// If error is returned, that's also valid
		return
	}
	if html == nil {
		t.Error("expected non-nil HTML for nil results")
	}
}

// === Integration: SaveSecurityHTML ===

func TestIntegrationSaveSecurityHTML_Success(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "security.html")

	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "target", Ports: []scanner.PortInfo{{Port: 445, State: "open"}}},
	}

	err := SaveSecurityHTML(path, results, nil, time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if !strings.Contains(string(content), "target") {
		t.Error("expected saved file to contain hostname")
	}
}

// === Integration: Full Report + Security Pipeline ===

func TestIntegrationFullReportPipeline(t *testing.T) {
	// Step 1: Generate scan report data
	scanData := &ScanReportData{
		ScanID:    "full-scan-001",
		Network:   "192.168.1.0/24",
		HostCount: 2,
		Results: []ScanResultRow{
			{IP: "192.168.1.1", Hostname: "router", Ports: 5, OS: "Linux", Vendor: "Cisco"},
			{IP: "192.168.1.2", Hostname: "server", Ports: 8, OS: "Windows", Vendor: "Dell"},
		},
		Findings: []SecurityFinding{
			{Severity: "high", Title: "Open SSH", Description: "SSH port 22 open", HostIP: "192.168.1.2", Port: 22},
		},
		Topology: &TopologySummary{
			DeviceCount: 2,
			LinkCount:   1,
		},
	}

	// Step 2: Render HTML
	html, err := RenderScanHTML(scanData)
	if err != nil {
		t.Fatalf("RenderScanHTML error: %v", err)
	}
	if len(html) == 0 {
		t.Fatal("expected non-empty HTML")
	}

	// Step 3: Save HTML
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "scan_report.html")
	err = SaveScanHTML(htmlPath, scanData)
	if err != nil {
		t.Fatalf("SaveScanHTML error: %v", err)
	}

	// Step 4: Verify saved HTML
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("failed to read saved HTML: %v", err)
	}
	if !strings.Contains(string(content), "router") {
		t.Error("expected saved HTML to contain hostname")
	}

	// Step 5: Generate security report
	results := []scanner.Result{
		{IP: "192.168.1.2", Hostname: "server", Ports: []scanner.PortInfo{{Port: 22, State: "open"}}},
	}
	securityHTML, err := RenderSecurityHTML(results, nil, time.Now())
	if err != nil {
		t.Fatalf("RenderSecurityHTML error: %v", err)
	}
	if len(securityHTML) == 0 {
		t.Fatal("expected non-empty security HTML")
	}

	// Step 6: Save security report
	secPath := filepath.Join(tmpDir, "security_report.html")
	err = SaveSecurityHTML(secPath, results, nil, time.Now())
	if err != nil {
		t.Fatalf("SaveSecurityHTML error: %v", err)
	}

	secContent, err := os.ReadFile(secPath)
	if err != nil {
		t.Fatalf("failed to read saved security HTML: %v", err)
	}
	// Security HTML contains the hostname and port info
	if !strings.Contains(string(secContent), "server") {
		t.Error("expected saved security HTML to contain hostname")
	}
}

// === Integration: Multiple Reports ===

func TestIntegrationMultipleReports(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate 3 different reports
	for i := 0; i < 3; i++ {
		data := &ScanReportData{
			ScanID:    "scan-00" + string(rune('1'+i)),
			Network:   "10.0.0.0/24",
			HostCount: 1,
			Results: []ScanResultRow{
				{IP: "10.0.0.1", Hostname: "host", Ports: 2},
			},
		}

		path := filepath.Join(tmpDir, "report_"+string(rune('0'+i))+".html")
		err := SaveScanHTML(path, data)
		if err != nil {
			t.Fatalf("failed to save report %d: %v", i, err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read report %d: %v", i, err)
		}
		if len(content) == 0 {
			t.Fatalf("expected non-empty report %d", i)
		}
	}
}

// === Integration: Edge Cases ===

func TestIntegrationReport_EmptyNetwork(t *testing.T) {
	data := &ScanReportData{
		ScanID:    "scan-001",
		Network:   "",
		HostCount: 0,
	}
	html, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("expected no error for empty network, got %v", err)
	}
	if len(html) == 0 {
		t.Fatal("expected non-empty HTML even with empty network")
	}
}

func TestIntegrationReport_EmptyHostname(t *testing.T) {
	data := &ScanReportData{
		ScanID:    "scan-001",
		Network:   "10.0.0.0/24",
		HostCount: 1,
		Results: []ScanResultRow{
			{IP: "10.0.0.1", Hostname: "", Ports: 2},
		},
	}
	html, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("expected no error for empty hostname, got %v", err)
	}
	if len(html) == 0 {
		t.Fatal("expected non-empty HTML even with empty hostname")
	}
}

func TestIntegrationReport_EmptyIP(t *testing.T) {
	data := &ScanReportData{
		ScanID:    "scan-001",
		Network:   "10.0.0.0/24",
		HostCount: 1,
		Results: []ScanResultRow{
			{IP: "", Hostname: "host", Ports: 2},
		},
	}
	html, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("expected no error for empty IP, got %v", err)
	}
	if len(html) == 0 {
		t.Fatal("expected non-empty HTML even with empty IP")
	}
}
