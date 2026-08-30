package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

// --- results_model.go tests ---

func TestSortedResultsForDisplay_ByIP(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.3"},
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}
	sorted := sortedResultsForDisplay(results)
	if sorted[0].IP != "192.168.1.1" {
		t.Errorf("expected first by IP, got %s", sorted[0].IP)
	}
	if sorted[2].IP != "192.168.1.3" {
		t.Errorf("expected last by IP, got %s", sorted[2].IP)
	}
}

func TestSortedResultsForDisplay_Empty(t *testing.T) {
	sorted := sortedResultsForDisplay(nil)
	if sorted == nil {
		// sortedResultsForDisplay может вернуть nil для nil input
		t.Log("nil result for nil input — допустимо")
	}
}

func TestSortedResultsForDisplayWithMode_ByHostName(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "beta"},
		{IP: "192.168.1.2", Hostname: "alpha"},
		{IP: "192.168.1.3", Hostname: "gamma"},
	}
	sorted := sortedResultsForDisplayWithMode(results, "HostName")
	if sorted[0].Hostname != "alpha" {
		t.Errorf("expected 'alpha' first, got %s", sorted[0].Hostname)
	}
	if sorted[2].Hostname != "gamma" {
		t.Errorf("expected 'gamma' last, got %s", sorted[2].Hostname)
	}
}

func TestSortedResultsForDisplayWithMode_IP(t *testing.T) {
	results := []scanner.Result{
		{IP: "10.0.0.3"},
		{IP: "10.0.0.1"},
	}
	sorted := sortedResultsForDisplayWithMode(results, "IP")
	if sorted[0].IP != "10.0.0.1" {
		t.Errorf("expected IP sort, got %s", sorted[0].IP)
	}
}

func TestSortedResultsForDisplayWithMode_Cards(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.2"},
		{IP: "192.168.1.1"},
	}
	sorted := sortedResultsForDisplayWithMode(results, "Карточки")
	// Карточки — без сортировки
	if len(sorted) != 2 {
		t.Errorf("expected 2 results, got %d", len(sorted))
	}
}

func TestSortedResultsForDisplayWithMode_UnknownMode(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1"},
	}
	sorted := sortedResultsForDisplayWithMode(results, "unknown_mode")
	if len(sorted) != 1 {
		t.Errorf("expected 1 result for unknown mode, got %d", len(sorted))
	}
}

func TestFilterResultsForDisplay_EmptyQuery(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}
	filtered := filterResultsForDisplay(results, "")
	if len(filtered) != 2 {
		t.Errorf("expected 2 results for empty query, got %d", len(filtered))
	}
}

func TestFilterResultsForDisplay_ByHostname(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
		{IP: "192.168.1.3", Hostname: "router2"},
	}
	filtered := filterResultsForDisplay(results, "router")
	if len(filtered) != 2 {
		t.Errorf("expected 2 results for 'router', got %d", len(filtered))
	}
}

func TestFilterResultsForDisplay_ByIP(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "10.0.0.1"},
	}
	filtered := filterResultsForDisplay(results, "10.0.0")
	if len(filtered) != 1 {
		t.Errorf("expected 1 result for '10.0.0', got %d", len(filtered))
	}
	if filtered[0].IP != "10.0.0.1" {
		t.Errorf("expected '10.0.0.1', got %s", filtered[0].IP)
	}
}

func TestFilterResultsForDisplay_NoMatch(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}
	filtered := filterResultsForDisplay(results, "xyz")
	if len(filtered) != 0 {
		t.Errorf("expected 0 results for 'xyz', got %d", len(filtered))
	}
}

func TestFilterResultsForDisplay_EmptyResults(t *testing.T) {
	filtered := filterResultsForDisplay(nil, "test")
	if filtered == nil {
		t.Error("expected non-nil result for nil input")
	}
}

func TestFilterResultsForDisplayAdvanced_EmptyQuery(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", DeviceType: "Router"},
	}
	filtered := filterResultsForDisplayAdvanced(results, "", []string{}, false)
	if len(filtered) != 1 {
		t.Errorf("expected 1 result, got %d", len(filtered))
	}
}

func TestFilterResultsForDisplayAdvanced_ByType(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", DeviceType: "Router"},
		{IP: "192.168.1.2", DeviceType: "Switch"},
		{IP: "192.168.1.3", DeviceType: "Router"},
	}
	filtered := filterResultsForDisplayAdvanced(results, "Router", []string{}, false)
	if len(filtered) != 2 {
		t.Errorf("expected 2 results for 'Router' query, got %d", len(filtered))
	}
}

func TestFilterResultsForDisplayAdvanced_OnlyOpenPorts(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, State: "open"}}},
		{IP: "192.168.1.2", Ports: []scanner.PortInfo{{Port: 22, State: "closed"}}},
	}
	filtered := filterResultsForDisplayAdvanced(results, "", []string{}, true)
	if len(filtered) != 1 {
		t.Errorf("expected 1 result with open ports, got %d", len(filtered))
	}
	if filtered[0].IP != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got %s", filtered[0].IP)
	}
}

func TestFilterResultsForDisplayAdvanced_CIDRFilter(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", DeviceType: "Router"},
		{IP: "192.168.1.2", DeviceType: "Switch"},
	}
	// CIDR фильтр "192.168.1" должен оставить оба
	filtered := filterResultsForDisplayAdvanced(results, "192.168.1", []string{}, false)
	if len(filtered) != 2 {
		t.Errorf("expected 2 results for CIDR '192.168.1', got %d", len(filtered))
	}
}

func TestHasOpenPorts_NoPorts(t *testing.T) {
	if hasOpenPorts(nil) {
		t.Error("expected false for nil ports")
	}
}

func TestHasOpenPorts_EmptySlice(t *testing.T) {
	if hasOpenPorts([]scanner.PortInfo{}) {
		t.Error("expected false for empty slice")
	}
}

func TestHasOpenPorts_OpenPort(t *testing.T) {
	if !hasOpenPorts([]scanner.PortInfo{{Port: 80, State: "open"}}) {
		t.Error("expected true for open port")
	}
}

func TestHasOpenPorts_ClosedPorts(t *testing.T) {
	if hasOpenPorts([]scanner.PortInfo{{Port: 22, State: "closed"}}) {
		t.Error("expected false for closed port")
	}
}

func TestHasOpenPorts_MixedPorts(t *testing.T) {
	if !hasOpenPorts([]scanner.PortInfo{
		{Port: 22, State: "closed"},
		{Port: 80, State: "open"},
	}) {
		t.Error("expected true for mixed ports with one open")
	}
}

func TestOpenPortLabels_Empty(t *testing.T) {
	labels := openPortLabels(nil, 5)
	if len(labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(labels))
	}
}

func TestOpenPortLabels_Basic(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"},
		{Port: 443, State: "open", Protocol: "tcp", Service: "HTTPS"},
	}
	labels := openPortLabels(ports, 5)
	if len(labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(labels))
	}
	// openPortLabels форматирует как "80/TCP HTTP" (верхний регистр протокола)
	if labels[0] != "80/TCP HTTP" {
		t.Errorf("expected '80/TCP HTTP', got %s", labels[0])
	}
}

func TestOpenPortLabels_Limit(t *testing.T) {
	ports := make([]scanner.PortInfo, 10)
	for i := 0; i < 10; i++ {
		ports[i] = scanner.PortInfo{Port: i + 1, State: "open", Protocol: "tcp", Service: "test"}
	}
	labels := openPortLabels(ports, 3)
	// 3 метки + 1 метка "+7" (открыто 10, видно 3, остаток 7)
	if len(labels) != 4 {
		t.Errorf("expected 4 labels (3 + '+7'), got %d", len(labels))
	}
	if labels[3] != "+7" {
		t.Errorf("expected '+7' at index 3, got %s", labels[3])
	}
}

func TestOpenPortLabels_NonOpenSkipped(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 22, State: "closed"},
		{Port: 80, State: "open"},
		{Port: 443, State: "filtered"},
	}
	labels := openPortLabels(ports, 5)
	if len(labels) != 1 {
		t.Errorf("expected 1 label (only open), got %d", len(labels))
	}
}

func TestFormatPortNumber_Basic(t *testing.T) {
	result := formatPortNumber(80)
	if result != "80" {
		t.Errorf("expected '80', got %s", result)
	}
}

func TestFormatPortNumber_Zero(t *testing.T) {
	result := formatPortNumber(0)
	if result != "0" {
		t.Errorf("expected '0', got %s", result)
	}
}

func TestFormatPortNumber_Large(t *testing.T) {
	result := formatPortNumber(65535)
	if result != "65535" {
		t.Errorf("expected '65535', got %s", result)
	}
}

func TestNormalizeServiceName_Basic(t *testing.T) {
	result := normalizeServiceName("HTTP")
	if result != "HTTP" {
		t.Errorf("expected 'HTTP', got %s", result)
	}
}

func TestNormalizeServiceName_Empty(t *testing.T) {
	result := normalizeServiceName("")
	if result != "" {
		t.Errorf("expected empty, got %s", result)
	}
}

func TestNormalizeServiceName_Whitespace(t *testing.T) {
	result := normalizeServiceName(" http ")
	if result != "http" {
		t.Errorf("expected 'http' (trimmed), got %s", result)
	}
}

func TestFormatDeviceValue_Empty(t *testing.T) {
	result := formatDeviceValue("")
	if result != "-" {
		t.Errorf("expected '-', got %s", result)
	}
}

func TestFormatDeviceValue_NonEmpty(t *testing.T) {
	result := formatDeviceValue("Router")
	if result != "Router" {
		t.Errorf("expected 'Router', got %s", result)
	}
}

func TestNormalizeDeviceTypes_Empty(t *testing.T) {
	normalized := normalizeDeviceTypes(nil)
	if normalized == nil {
		t.Error("expected non-nil result for nil input")
	}
}

func TestNormalizeDeviceTypes_WithValues(t *testing.T) {
	input := map[string]int{"Router": 2, "Switch": 3}
	normalized := normalizeDeviceTypes(input)
	// Router → Network Device, Switch не в маппинге → остаётся "Switch"
	if normalized["Network Device"] != 2 {
		t.Errorf("expected Network Device=2 (from Router), got %d", normalized["Network Device"])
	}
	if normalized["Switch"] != 3 {
		t.Errorf("expected Switch=3 (not in mapping), got %d", normalized["Switch"])
	}
}

func TestCollectAnalytics_Empty(t *testing.T) {
	protocols, deviceTypes := collectAnalytics(nil)
	if protocols == nil {
		t.Error("expected non-nil protocols for nil input")
	}
	if deviceTypes == nil {
		t.Error("expected non-nil deviceTypes for nil input")
	}
}

func TestCollectAnalytics_Basic(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Protocols: []string{"TCP", "UDP"}, DeviceType: "Router"},
		{IP: "192.168.1.2", Protocols: []string{"TCP"}, DeviceType: "Switch"},
	}
	protocols, deviceTypes := collectAnalytics(results)
	if protocols["TCP"] != 2 {
		t.Errorf("expected TCP=2, got %d", protocols["TCP"])
	}
	if protocols["UDP"] != 1 {
		t.Errorf("expected UDP=1, got %d", protocols["UDP"])
	}
	if deviceTypes["Router"] != 1 {
		t.Errorf("expected Router=1, got %d", deviceTypes["Router"])
	}
}

func TestCollectAnalytics_EmptyFields(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Protocols: []string{}, DeviceType: ""},
	}
	protocols, _ := collectAnalytics(results)
	if protocols["TCP"] != 0 {
		t.Error("expected TCP not in result")
	}
	// Empty deviceType не должен добавлять в map
}

func TestOpenPortLabels_DefaultMax(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"},
	}
	labels := openPortLabels(ports, 0)
	// maxVisiblePorts=0 → default 24
	if len(labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(labels))
	}
}
