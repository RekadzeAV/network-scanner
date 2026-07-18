package presenter

import (
	"os"
	"testing"

	"network-scanner/internal/scanner"
)

// ============================================================================
// CLIPresenter
// ============================================================================

func TestCLIPresenter_DisplayHeader(t *testing.T) {
	p := CLIPresenter{}
	// Should not panic
	p.DisplayHeader()
}

func TestCLIPresenter_DisplayHost(t *testing.T) {
	p := CLIPresenter{}
	host := scanner.HostResult{
		IP:       "192.168.1.1",
		Hostname: "router",
	}
	// Should not panic
	p.DisplayHost(host)
}

func TestCLIPresenter_DisplaySummary(t *testing.T) {
	p := CLIPresenter{}
	// Should not panic
	p.DisplaySummary(5, 10)
}

func TestCLIPresenter_Export_JSON(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(t.TempDir())

	p := CLIPresenter{}
	results := []scanner.HostResult{
		{IP: "10.0.0.1"},
	}
	err := p.Export(results, "json")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCLIPresenter_Export_CSV(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(t.TempDir())

	p := CLIPresenter{}
	results := []scanner.HostResult{
		{IP: "10.0.0.1"},
	}
	err := p.Export(results, "csv")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCLIPresenter_Export_TXT(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(t.TempDir())

	p := CLIPresenter{}
	results := []scanner.HostResult{
		{IP: "10.0.0.1"},
	}
	err := p.Export(results, "txt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// ============================================================================
// JSONPresenter
// ============================================================================

func TestJSONPresenter_DisplayHeader(t *testing.T) {
	p := JSONPresenter{}
	p.DisplayHeader()
}

func TestJSONPresenter_DisplayHost(t *testing.T) {
	p := JSONPresenter{}
	p.DisplayHost(scanner.HostResult{IP: "10.0.0.1"})
}

func TestJSONPresenter_DisplaySummary(t *testing.T) {
	p := JSONPresenter{}
	p.DisplaySummary(3, 5)
}

func TestJSONPresenter_Export(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(t.TempDir())

	p := JSONPresenter{}
	results := []scanner.HostResult{
		{IP: "10.0.0.1"},
	}
	err := p.Export(results, "json")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// ============================================================================
// HTMLPresenter
// ============================================================================

func TestHTMLPresenter_DisplayHeader(t *testing.T) {
	p := HTMLPresenter{}
	p.DisplayHeader()
}

func TestHTMLPresenter_DisplayHost(t *testing.T) {
	p := HTMLPresenter{}
	p.DisplayHost(scanner.HostResult{IP: "10.0.0.1"})
}

func TestHTMLPresenter_DisplaySummary(t *testing.T) {
	p := HTMLPresenter{}
	p.DisplaySummary(3, 5)
}

func TestHTMLPresenter_Export_Success(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(t.TempDir())

	p := HTMLPresenter{}
	results := []scanner.HostResult{
		{
			IP:           "10.0.0.1",
			Hostname:     "router",
			MAC:          "aa:bb:cc:dd:ee:ff",
			DeviceType:   "router",
			DeviceVendor: "cisco",
			GuessOS:      "ios",
			SNMPEnabled:  true,
			Ports: []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", State: "open", Service: "http"},
				{Port: 443, Protocol: "tcp", State: "open", Service: "https"},
			},
		},
	}
	err := p.Export(results, "html")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := os.Stat("scan-report.html"); os.IsNotExist(err) {
		t.Fatal("expected scan-report.html to exist")
	}
}

func TestHTMLPresenter_Export_WrongFormat(t *testing.T) {
	p := HTMLPresenter{}
	err := p.Export(nil, "xml")
	if err == nil {
		t.Fatal("expected error for wrong format")
	}
}

func TestHTMLPresenter_Export_Empty(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(t.TempDir())

	p := HTMLPresenter{}
	err := p.Export([]scanner.HostResult{}, "html")
	if err != nil {
		t.Fatalf("expected no error for empty results, got %v", err)
	}
}

// ============================================================================
// XMLPresenter
// ============================================================================

func TestXMLPresenter_DisplayHeader(t *testing.T) {
	p := XMLPresenter{}
	p.DisplayHeader()
}

func TestXMLPresenter_DisplayHost(t *testing.T) {
	p := XMLPresenter{}
	p.DisplayHost(scanner.HostResult{IP: "10.0.0.1"})
}

func TestXMLPresenter_DisplaySummary(t *testing.T) {
	p := XMLPresenter{}
	p.DisplaySummary(3, 5)
}

func TestXMLPresenter_Export_Success(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(t.TempDir())

	p := XMLPresenter{}
	results := []scanner.HostResult{
		{
			IP:                "10.0.0.1",
			Hostname:          "router",
			MAC:               "aa:bb:cc:dd:ee:ff",
			DeviceType:        "router",
			DeviceVendor:      "cisco",
			GuessOS:           "ios",
			GuessOSConfidence: "95",
			Ports: []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", State: "open", Service: "http", Version: "1.1"},
			},
		},
	}
	err := p.Export(results, "xml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := os.Stat("scan-results.xml"); os.IsNotExist(err) {
		t.Fatal("expected scan-results.xml to exist")
	}
}

func TestXMLPresenter_Export_WrongFormat(t *testing.T) {
	p := XMLPresenter{}
	err := p.Export(nil, "json")
	if err == nil {
		t.Fatal("expected error for wrong format")
	}
}

func TestXMLPresenter_Export_Empty(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(t.TempDir())

	p := XMLPresenter{}
	err := p.Export([]scanner.HostResult{}, "xml")
	if err != nil {
		t.Fatalf("expected no error for empty results, got %v", err)
	}
}

// ============================================================================
// countOpenPorts helper
// ============================================================================

func TestCountOpenPorts_Empty(t *testing.T) {
	if got := countOpenPorts(nil); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestCountOpenPorts_WithPorts(t *testing.T) {
	results := []scanner.HostResult{
		{Ports: []scanner.PortInfo{
			{Port: 80, State: "open"},
			{Port: 443, State: "open"},
			{Port: 22, State: "closed"},
		}},
		{Ports: []scanner.PortInfo{
			{Port: 8080, State: "open"},
		}},
	}
	if got := countOpenPorts(results); got != 3 {
		t.Fatalf("expected 3 open ports, got %d", got)
	}
}

func TestCountOpenPorts_NoOpen(t *testing.T) {
	results := []scanner.HostResult{
		{Ports: []scanner.PortInfo{
			{Port: 80, State: "closed"},
		}},
	}
	if got := countOpenPorts(results); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}
