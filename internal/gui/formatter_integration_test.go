package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

// === Integration: Format Results ===

func TestIntegrationFormatResults_Empty(t *testing.T) {
	output := FormatResultsForDisplay(nil)
	if output == "" {
		t.Error("expected non-empty output")
	}
	prefix := "## Ре"
	if len(output) < len(prefix) || output[:len(prefix)] != prefix {
		t.Errorf("expected output to start with %q, got %q", prefix, output[:min(len(prefix), len(output))])
	}
}

func TestIntegrationFormatResults_SingleResult(t *testing.T) {
	results := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Hostname:   "router-main",
			MAC:        "AA:BB:CC:DD:EE:01",
			DeviceType: "Router",
			Protocols:  []string{"TCP", "UDP"},
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"},
				{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
			},
		},
	}

	output := FormatResultsForDisplay(results)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if !contains(output, "router-main") {
		t.Error("expected output to contain 'router-main'")
	}
	if !contains(output, "192.168.1.1") {
		t.Error("expected output to contain '192.168.1.1'")
	}
	if !contains(output, "AA:BB:CC:DD:EE:01") {
		t.Error("expected output to contain MAC address")
	}
	if !contains(output, "## Сетевая аналитика") {
		t.Error("expected output to contain analytics section")
	}
}

func TestIntegrationFormatResults_MultipleResults(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router", Protocols: []string{"TCP"}},
		{IP: "192.168.1.2", Hostname: "server", DeviceType: "Server", Protocols: []string{"TCP", "UDP"}},
		{IP: "192.168.1.3", Hostname: "desktop", DeviceType: "Desktop", Protocols: []string{"TCP"}},
	}

	output := FormatResultsForDisplay(results)
	if !contains(output, "router") {
		t.Error("expected output to contain 'router'")
	}
	if !contains(output, "server") {
		t.Error("expected output to contain 'server'")
	}
	if !contains(output, "desktop") {
		t.Error("expected output to contain 'desktop'")
	}
}

func TestIntegrationFormatResults_EmptyFields(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "", MAC: "", DeviceType: "", Protocols: []string{"TCP"}},
	}

	output := FormatResultsForDisplay(results)
	if !contains(output, "| - | 192.168.1.1 |  | - |") {
		t.Errorf("expected output to contain table row with empty fields, got:\n%s", output)
	}
}

func TestIntegrationFormatResults_LongFieldsTruncated(t *testing.T) {
	longHostname := "this-is-a-very-long-hostname-that-should-be-truncated-to-fit-in-the-table"
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: longHostname, Protocols: []string{"TCP"}},
	}

	output := FormatResultsForDisplay(results)
	// Check that hostname is truncated (max 30 chars)
	if len(longHostname) > 30 && contains(output, longHostname) {
		t.Error("expected long hostname to be truncated")
	}
}

func TestIntegrationFormatResults_MarkdownEscaping(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "test|pipe", MAC: "AA:BB|CC", DeviceType: "Test", Protocols: []string{"TCP"}},
	}

	output := FormatResultsForDisplay(results)
	// Pipe characters should be escaped
	if !contains(output, "\\|") {
		t.Error("expected pipe characters to be escaped")
	}
}

// === Integration: Format Ports ===

func TestIntegrationFormatPorts_Empty(t *testing.T) {
	output := formatPorts([]scanner.PortInfo{})
	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}

	output = formatPorts(nil)
	if output != "" {
		t.Errorf("expected empty output for nil, got %q", output)
	}
}

func TestIntegrationFormatPorts_OpenPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"},
		{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
	}

	output := formatPorts(ports)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if !contains(output, "22") {
		t.Error("expected output to contain port 22")
	}
	if !contains(output, "80") {
		t.Error("expected output to contain port 80")
	}
}

func TestIntegrationFormatPorts_ClosedPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 25, State: "closed", Protocol: "TCP", Service: "smtp"},
	}

	output := formatPorts(ports)
	// formatPorts возвращает только open порты, поэтому для closed ports вернётся ""
	if output != "" {
		t.Errorf("expected empty output for closed ports, got %q", output)
	}
}

func TestIntegrationFormatPorts_MixedPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"},
		{Port: 25, State: "closed", Protocol: "TCP", Service: "smtp"},
		{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
	}

	output := formatPorts(ports)
	if output == "" {
		t.Error("expected non-empty output")
	}
	// formatPorts возвращает только open порты
	if !contains(output, "22") || !contains(output, "80") {
		t.Error("expected output to contain open ports 22 and 80")
	}
	// Closed port 25 не должен быть в output
	if contains(output, "25") {
		t.Error("expected output NOT to contain closed port 25")
	}
}

func TestIntegrationFormatPorts_MaxLimit(t *testing.T) {
	ports := make([]scanner.PortInfo, 0, 30)
	for i := 1; i <= 30; i++ {
		ports = append(ports, scanner.PortInfo{
			Port: i, State: "open", Protocol: "TCP", Service: "test",
		})
	}

	output := formatPorts(ports)
	if output == "" {
		t.Error("expected non-empty output")
	}
	// Should be truncated
	if len(output) > 500 {
		t.Errorf("expected output to be truncated (max 500), got %d chars", len(output))
	}
}

// === Integration: Escape Markdown ===

func TestIntegrationEscapeMarkdown_Empty(t *testing.T) {
	// escapeMarkdown("", false) возвращает "-", а не ""
	output := escapeMarkdown("")
	if output != "-" {
		t.Errorf("expected '-', got %q", output)
	}

	// escapeMarkdown("", true) возвращает ""
	output = escapeMarkdown("", true)
	if output != "" {
		t.Errorf("expected empty string with allowEmpty=true, got %q", output)
	}
}

func TestIntegrationEscapeMarkdown_Pipes(t *testing.T) {
	output := escapeMarkdown("test|pipe")
	if output != "test\\|pipe" {
		t.Errorf("expected 'test\\|pipe', got %q", output)
	}
}

func TestIntegrationEscapeMarkdown_MultiplePipes(t *testing.T) {
	output := escapeMarkdown("a|b|c")
	if output != "a\\|b\\|c" {
		t.Errorf("expected 'a\\|b\\|c', got %q", output)
	}
}

func TestIntegrationEscapeMarkdown_NoPipes(t *testing.T) {
	output := escapeMarkdown("normal text")
	if output != "normal text" {
		t.Errorf("expected 'normal text', got %q", output)
	}
}

// === Integration: Truncate String ===

func TestIntegrationTruncateString_Empty(t *testing.T) {
	output := truncateString("", 10)
	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}
}

func TestIntegrationTruncateString_Shorter(t *testing.T) {
	output := truncateString("short", 10)
	if output != "short" {
		t.Errorf("expected 'short', got %q", output)
	}
}

func TestIntegrationTruncateString_Longer(t *testing.T) {
	output := truncateString("this is a long string", 10)
	if len(output) > 10 {
		t.Errorf("expected output <= 10 chars, got %d", len(output))
	}
}

func TestIntegrationTruncateString_ExactLength(t *testing.T) {
	output := truncateString("exactly10", 10)
	if output != "exactly10" {
		t.Errorf("expected 'exactly10', got %q", output)
	}
}

// === Integration: Full Display Pipeline ===

func TestIntegrationFullDisplayPipeline(t *testing.T) {
	results := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Hostname:   "router-main",
			MAC:        "AA:BB:CC:DD:EE:01",
			DeviceType: "Router",
			Protocols:  []string{"TCP", "UDP"},
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"},
				{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
			},
		},
		{
			IP:         "192.168.1.2",
			Hostname:   "web-server",
			MAC:        "AA:BB:CC:DD:EE:02",
			DeviceType: "Web Server",
			Protocols:  []string{"TCP"},
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
				{Port: 443, State: "open", Protocol: "TCP", Service: "https"},
			},
		},
	}

	// Step 1: Format results
	output := FormatResultsForDisplay(results)
	if output == "" {
		t.Fatal("expected non-empty output")
	}

	// Step 2: Verify output contains all expected data
	expectedStrings := []string{
		"router-main",
		"web-server",
		"192.168.1.1",
		"192.168.1.2",
		"AA:BB:CC:DD:EE:01",
		"AA:BB:CC:DD:EE:02",
		"## Сетевая аналитика",
		"22",
		"80",
		"443",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}

	// Step 3: Collect analytics from same results
	protocols, deviceTypes := collectAnalytics(results)
	if protocols["TCP"] != 2 {
		t.Errorf("expected 2 TCP results, got %d", protocols["TCP"])
	}
	if protocols["UDP"] != 1 {
		t.Errorf("expected 1 UDP result, got %d", protocols["UDP"])
	}
	// collectAnalytics не нормализует device types, считаем как есть
	if deviceTypes["Web Server"] != 1 {
		t.Errorf("expected 1 Web Server result, got %d", deviceTypes["Web Server"])
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
