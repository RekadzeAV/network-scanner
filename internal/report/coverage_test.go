package report

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/cve"
	"network-scanner/internal/scanner"
)

// ============================================================================
// PDFReport — 0% → ~100%
// ============================================================================

func TestNewPDFReport(t *testing.T) {
	r := NewPDFReport("Test Report")
	if r == nil {
		t.Fatal("expected non-nil PDFReport")
	}
	if r.title != "Test Report" {
		t.Fatalf("expected title 'Test Report', got %q", r.title)
	}
}

func TestPDFReport_AddMetadata(t *testing.T) {
	r := NewPDFReport("Test")
	r.AddMetadata("key1", "value1")
	r.AddMetadata("key2", "value2")
	if r.metadata["key1"] != "value1" {
		t.Fatalf("expected metadata key1=value1, got %q", r.metadata["key1"])
	}
	if r.metadata["key2"] != "value2" {
		t.Fatalf("expected metadata key2=value2, got %q", r.metadata["key2"])
	}
}

func TestPDFReport_AddScanResults(t *testing.T) {
	r := NewPDFReport("Test")
	results := []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "router", GuessOS: "linux", Ports: []contracts.PortInfo{{Port: 80}}},
		{IP: "192.168.1.2", Hostname: "", GuessOS: "", Ports: nil},
	}
	r.AddScanResults(results)
	// Should not panic
}

func TestPDFReport_AddScanResults_Empty(t *testing.T) {
	r := NewPDFReport("Test")
	r.AddScanResults([]contracts.ScanResult{})
	// Should not panic
}

func TestPDFReport_AddSecurityFindings(t *testing.T) {
	r := NewPDFReport("Test")
	findings := []contracts.Finding{
		{Severity: "high", Title: "Open Telnet", Recommendation: "Disable Telnet"},
		{Severity: "medium", Title: "Open FTP", Recommendation: "Use SFTP"},
	}
	r.AddSecurityFindings(findings)
	// Should not panic
}

func TestPDFReport_AddSecurityFindings_Empty(t *testing.T) {
	r := NewPDFReport("Test")
	r.AddSecurityFindings([]contracts.Finding{})
	// Should not panic
}

func TestPDFReport_AddTopology(t *testing.T) {
	r := NewPDFReport("Test")
	topo := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "10.0.0.1", Hostname: "router", Type: "router"},
			{IP: "10.0.0.2", Hostname: "", Type: ""},
		},
		Links: []*contracts.Link{
			{Source: &contracts.Device{IP: "10.0.0.1"}, Target: &contracts.Device{IP: "10.0.0.2"}},
		},
	}
	r.AddTopology(topo)
	// Should not panic
}

func TestPDFReport_AddTopology_Nil(t *testing.T) {
	r := NewPDFReport("Test")
	r.AddTopology(&contracts.Topology{})
	// Should not panic
}

func TestPDFReport_Save(t *testing.T) {
	r := NewPDFReport("Test Report")
	r.AddMetadata("author", "test")
	r.AddScanResults([]contracts.ScanResult{
		{IP: "10.0.0.1", Hostname: "host1", GuessOS: "linux", Ports: []contracts.PortInfo{{Port: 80}}},
	})

	path := filepath.Join(t.TempDir(), "report.pdf")
	err := r.Save(path)
	if err != nil {
		t.Fatalf("expected no error saving PDF, got %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected PDF file to exist")
	}
}

func TestPDFReport_Bytes(t *testing.T) {
	r := NewPDFReport("Test")
	r.AddScanResults([]contracts.ScanResult{{IP: "10.0.0.1"}})

	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("expected no error getting bytes, got %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}

func TestPDFReport_Bytes_Empty(t *testing.T) {
	r := NewPDFReport("Test")
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF bytes even for empty report")
	}
}

func TestByteWriter_Write(t *testing.T) {
	w := &byteWriter{buf: &[]byte{}}
	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if n != 5 {
		t.Fatalf("expected 5 bytes written, got %d", n)
	}
	if string(*w.buf) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(*w.buf))
	}
}

// ============================================================================
// SaveSecurityHTML / SaveSecurityHTMLWithRisk / SaveSecurityHTMLWithRiskOptions — 0% → 100%
// ============================================================================

func TestSaveSecurityHTML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "security.html")
	results := []scanner.Result{
		{IP: "10.0.0.1", Hostname: "host1"},
	}
	findings := []cve.Match{
		{HostIP: "10.0.0.1", Port: 80, Service: "http"},
	}
	err := SaveSecurityHTML(path, results, findings, time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected HTML file to exist")
	}
}

func TestSaveSecurityHTMLWithRisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "security_risk.html")
	results := []scanner.Result{
		{IP: "10.0.0.1", Hostname: "host1"},
	}
	findings := []cve.Match{
		{HostIP: "10.0.0.1", Port: 80, Service: "http"},
	}
	err := SaveSecurityHTMLWithRisk(path, results, findings, nil, time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected HTML file to exist")
	}
}

func TestSaveSecurityHTMLWithRiskOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "security_opts.html")
	results := []scanner.Result{
		{IP: "10.0.0.1", Hostname: "host1"},
	}
	findings := []cve.Match{
		{HostIP: "10.0.0.1", Port: 80, Service: "http"},
	}
	opts := Options{
		RedactSensitive: false,
		ReportID:        "test-report-001",
		GenerationMode:  "auto",
		PolicyVersion:   "v2",
		UnsafeConsent:   true,
	}
	err := SaveSecurityHTMLWithRiskOptions(path, results, findings, nil, time.Now(), opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected HTML file to exist")
	}
}

func TestSaveSecurityHTMLWithRiskOptions_InvalidPath(t *testing.T) {
	err := SaveSecurityHTMLWithRiskOptions(
		"/nonexistent/path/that/does/not/exist/report.html",
		nil, nil, nil, time.Now(),
		Options{},
	)
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

// ============================================================================
// RenderScanHTML / SaveScanHTML — edge cases
// ============================================================================

func TestSaveScanHTML_Success(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scan.html")
	data := &ScanReportData{
		GeneratedAt: time.Now().Format(time.RFC3339),
		ScanID:      "test-scan",
		Network:     "192.168.1.0/24",
		HostCount:   1,
		Results: []ScanResultRow{
			{IP: "192.168.1.1", Hostname: "router", Ports: 3, OS: "linux", Vendor: "cisco"},
		},
	}
	err := SaveScanHTML(path, data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected HTML file to exist")
	}
}

func TestSaveScanHTML_InvalidPath(t *testing.T) {
	data := &ScanReportData{}
	// Use a path with reserved device name that will fail on Windows
	err := SaveScanHTML("NUL\report.html", data)
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestRenderScanHTML_WithFindings(t *testing.T) {
	data := &ScanReportData{
		ScanID:    "test",
		Network:   "10.0.0.0/24",
		HostCount: 2,
		Results: []ScanResultRow{
			{IP: "10.0.0.1", Hostname: "host1", Ports: 2},
		},
		Findings: []SecurityFinding{
			{Severity: "high", Title: "Open Telnet", HostIP: "10.0.0.1", Port: 23},
		},
		Topology: &TopologySummary{
			DeviceCount: 2,
			LinkCount:   1,
			Devices: []TopologyDevice{
				{IP: "10.0.0.1", Hostname: "router"},
			},
		},
	}
	b, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty HTML")
	}
}

func TestRenderScanHTML_NilTopology(t *testing.T) {
	data := &ScanReportData{
		ScanID:    "test",
		HostCount: 0,
	}
	b, err := RenderScanHTML(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty HTML")
	}
}

// ============================================================================
// DefaultHTMLReportOptions
// ============================================================================

func TestDefaultHTMLReportOptions_Coverage(t *testing.T) {
	opts := DefaultHTMLReportOptions()
	if !opts.IncludeScanResults {
		t.Fatal("expected IncludeScanResults=true")
	}
	if !opts.IncludeSecurity {
		t.Fatal("expected IncludeSecurity=true")
	}
	if !opts.IncludeTopology {
		t.Fatal("expected IncludeTopology=true")
	}
	if opts.RedactSensitive {
		t.Fatal("expected RedactSensitive=false")
	}
}
