package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

// === Integration: Results Pipeline ===

func TestIntegrationResultsPipeline_SortFilterAnalyze(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router-main", MAC: "AA:BB:CC:DD:EE:01", DeviceType: "Router", Protocols: []string{"TCP", "UDP"}},
		{IP: "192.168.1.2", Hostname: "server-web", MAC: "AA:BB:CC:DD:EE:02", DeviceType: "Web Server", Protocols: []string{"TCP"}},
		{IP: "192.168.1.3", Hostname: "printer-hp", MAC: "AA:BB:CC:DD:EE:03", DeviceType: "Printer", Protocols: []string{"TCP"}},
		{IP: "192.168.1.10", Hostname: "workstation-1", MAC: "AA:BB:CC:DD:EE:04", DeviceType: "Desktop", Protocols: []string{"TCP", "UDP"}},
		{IP: "192.168.1.100", Hostname: "iot-camera", MAC: "AA:BB:CC:DD:EE:05", DeviceType: "Camera", Protocols: []string{"TCP"}},
	}

	// Step 1: Sort by IP
	sorted := sortedResultsForDisplay(results)
	if sorted[0].IP != "192.168.1.1" {
		t.Errorf("expected first IP '192.168.1.1', got %q", sorted[0].IP)
	}
	if sorted[len(sorted)-1].IP != "192.168.1.100" {
		t.Errorf("expected last IP '192.168.1.100', got %q", sorted[len(sorted)-1].IP)
	}

	// Step 2: Sort by HostName
	sortedByName := sortedResultsForDisplayWithMode(results, "HostName")
	if sortedByName[0].Hostname != "iot-camera" {
		t.Errorf("expected first hostname 'iot-camera', got %q", sortedByName[0].Hostname)
	}
	if sortedByName[len(sortedByName)-1].Hostname != "workstation-1" {
		t.Errorf("expected last hostname 'workstation-1', got %q", sortedByName[len(sortedByName)-1].Hostname)
	}

	// Step 3: Filter by query
	filtered := filterResultsForDisplay(results, "server")
	if len(filtered) != 1 {
		t.Errorf("expected 1 result for 'server', got %d", len(filtered))
	}
	if filtered[0].Hostname != "server-web" {
		t.Errorf("expected 'server-web', got %q", filtered[0].Hostname)
	}

	// Step 4: Advanced filter (open ports only)
	resultsWithOpenPorts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", Ports: []scanner.PortInfo{{Port: 22, State: "open"}, {Port: 80, State: "closed"}}},
		{IP: "192.168.1.2", Hostname: "server", Ports: []scanner.PortInfo{{Port: 80, State: "open"}, {Port: 443, State: "open"}}},
		{IP: "192.168.1.3", Hostname: "printer", Ports: []scanner.PortInfo{{Port: 9100, State: "closed"}}},
	}
	openOnly := filterResultsForDisplayAdvanced(resultsWithOpenPorts, "", []string{}, true)
	if len(openOnly) != 2 {
		t.Errorf("expected 2 open-port results, got %d", len(openOnly))
	}

	// Step 5: Collect analytics
	protocols, deviceTypes := collectAnalytics(results)
	if protocols["TCP"] != 5 {
		t.Errorf("expected 5 TCP results, got %d", protocols["TCP"])
	}
	if protocols["UDP"] != 2 {
		t.Errorf("expected 2 UDP results, got %d", protocols["UDP"])
	}
	// collectAnalytics не нормализует device types, считаем как есть
	if deviceTypes["Web Server"] != 1 {
		t.Errorf("expected 1 Web Server result, got %d", deviceTypes["Web Server"])
	}
}

func TestIntegrationResultsPipeline_EmptyResults(t *testing.T) {
	var results []scanner.Result

	sorted := sortedResultsForDisplay(results)
	if len(sorted) != 0 {
		t.Errorf("expected empty sorted results, got %d", len(sorted))
	}

	filtered := filterResultsForDisplay(results, "test")
	if len(filtered) != 0 {
		t.Errorf("expected empty filtered results, got %d", len(filtered))
	}

	openOnly := filterResultsForDisplayAdvanced(results, "", []string{}, true)
	if len(openOnly) != 0 {
		t.Errorf("expected empty open-only results, got %d", len(openOnly))
	}

	protocols, deviceTypes := collectAnalytics(results)
	if len(protocols) != 0 {
		t.Errorf("expected empty protocols, got %d", len(protocols))
	}
	if len(deviceTypes) != 0 {
		t.Errorf("expected empty device types, got %d", len(deviceTypes))
	}
}

func TestIntegrationResultsPipeline_NilResults(t *testing.T) {
	var results []scanner.Result = nil

	// Функции не должны паниковать при nil input
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Error("sortedResultsForDisplay should not panic for nil results")
			}
		}()
		sorted := sortedResultsForDisplay(results)
		_ = sorted
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Error("filterResultsForDisplay should not panic for nil results")
			}
		}()
		filtered := filterResultsForDisplay(results, "test")
		_ = filtered
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Error("filterResultsForDisplayAdvanced should not panic for nil results")
			}
		}()
		openOnly := filterResultsForDisplayAdvanced(results, "", []string{}, true)
		_ = openOnly
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Error("collectAnalytics should not panic for nil results")
			}
		}()
		protocols, deviceTypes := collectAnalytics(results)
		_ = protocols
		_ = deviceTypes
	}()
}

func TestIntegrationResultsPipeline_FilterByType(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router"},
		{IP: "192.168.1.2", Hostname: "server", DeviceType: "Server"},
		{IP: "192.168.1.3", Hostname: "desktop", DeviceType: "Desktop"},
		{IP: "192.168.1.4", Hostname: "printer", DeviceType: "Printer"},
	}

	filtered := filterResultsForDisplayAdvanced(results, "", []string{"Network Device"}, false)
	if len(filtered) != 2 { // Router + Printer
		t.Errorf("expected 2 Network Device results, got %d", len(filtered))
	}

	filtered = filterResultsForDisplayAdvanced(results, "", []string{"Computer"}, false)
	if len(filtered) != 1 { // Desktop
		t.Errorf("expected 1 Computer result, got %d", len(filtered))
	}
}

func TestIntegrationResultsPipeline_FilterByQueryAndType(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "web-server", DeviceType: "Web Server"},
		{IP: "192.168.1.2", Hostname: "db-server", DeviceType: "Database Server"},
		{IP: "192.168.1.3", Hostname: "web-router", DeviceType: "Router"},
	}

	// Filter by query "server"
	filtered := filterResultsForDisplay(results, "server")
	if len(filtered) != 2 {
		t.Errorf("expected 2 results for 'server', got %d", len(filtered))
	}

	// Filter by query "server" + type "Server"
	filteredAdv := filterResultsForDisplayAdvanced(results, "server", []string{"Server"}, false)
	if len(filteredAdv) != 2 {
		t.Errorf("expected 2 results for 'server' + 'Server', got %d", len(filteredAdv))
	}
}

// === Integration: Port Labels ===

func TestIntegrationPortLabels_Integration(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"},
		{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
		{Port: 443, State: "open", Protocol: "TCP", Service: "https"},
		{Port: 25, State: "closed", Protocol: "TCP", Service: "smtp"},
		{Port: 53, State: "open", Protocol: "UDP", Service: "dns"},
	}

	labels := openPortLabels(ports, 10)
	if len(labels) != 4 { // 4 open ports
		t.Errorf("expected 4 labels, got %d", len(labels))
	}

	labels = openPortLabels(ports, 2)
	if len(labels) != 3 { // 2 open + 1 overflow indicator
		t.Errorf("expected 3 labels (2 + overflow), got %d", len(labels))
	}
	if labels[2] != "+2" {
		t.Errorf("expected overflow indicator '+2', got %q", labels[2])
	}
}

func TestIntegrationPortLabels_EmptyPorts(t *testing.T) {
	labels := openPortLabels([]scanner.PortInfo{}, 10)
	if len(labels) != 0 {
		t.Errorf("expected empty labels, got %d", len(labels))
	}

	labels = openPortLabels(nil, 10)
	if labels == nil {
		t.Error("expected non-nil labels")
	}
}

// === Integration: Device Type Normalization ===

func TestIntegrationDeviceTypes_Normalization(t *testing.T) {
	deviceTypes := map[string]int{
		"Router":         2,
		"Web Server":     1,
		"Desktop":        3,
		"Printer":        1,
		"Unknown Device": 1,
	}

	normalized := normalizeDeviceTypes(deviceTypes)
	if len(normalized) != 4 {
		t.Errorf("expected 4 normalized types, got %d", len(normalized))
	}
	if normalized["Network Device"] != 3 { // Router(2) + Printer(1)
		t.Errorf("expected 3 Network Device, got %d", normalized["Network Device"])
	}
	if normalized["Server"] != 1 {
		t.Errorf("expected 1 Server, got %d", normalized["Server"])
	}
	if normalized["Computer"] != 3 {
		t.Errorf("expected 3 Computer, got %d", normalized["Computer"])
	}
	if normalized["Unknown"] != 1 {
		t.Errorf("expected 1 Unknown, got %d", normalized["Unknown"])
	}
}

// === Integration: Format Device Value ===

func TestIntegrationFormatDeviceValue_Integration(t *testing.T) {
	if formatDeviceValue("") != "-" {
		t.Error("expected '-' for empty value")
	}
	if formatDeviceValue("  ") != "-" {
		t.Error("expected '-' for whitespace value")
	}
	if formatDeviceValue("  Router  ") != "Router" {
		t.Errorf("expected 'Router', got %q", formatDeviceValue("  Router  "))
	}
	if formatDeviceValue("Server-1") != "Server-1" {
		t.Errorf("expected 'Server-1', got %q", formatDeviceValue("Server-1"))
	}
}
