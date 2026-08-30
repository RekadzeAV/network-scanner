package gui

import (
	"strings"
	"testing"

	"network-scanner/internal/scanner"
)

// --- results_analytics_view.go tests ---

func TestBuildAnalyticsMarkdown_Empty(t *testing.T) {
	result := buildAnalyticsMarkdown(nil, nil)
	if !strings.Contains(result, "Нет данных") {
		t.Errorf("expected 'Нет данных' for empty analytics, got %q", result)
	}
}

func TestBuildAnalyticsMarkdown_ProtocolsOnly(t *testing.T) {
	result := buildAnalyticsMarkdown(map[string]int{"TCP": 5}, nil)
	if !strings.Contains(result, "TCP") {
		t.Errorf("expected TCP in markdown, got %q", result)
	}
	if !strings.Contains(result, "5") {
		t.Errorf("expected 5 in markdown, got %q", result)
	}
}

func TestBuildAnalyticsMarkdown_TypesOnly(t *testing.T) {
	result := buildAnalyticsMarkdown(nil, map[string]int{"Router": 3})
	if !strings.Contains(result, "Router") {
		t.Errorf("expected Router in markdown, got %q", result)
	}
}

func TestBuildAnalyticsMarkdown_Both(t *testing.T) {
	result := buildAnalyticsMarkdown(
		map[string]int{"TCP": 2, "UDP": 1},
		map[string]int{"Router": 1, "Server": 1},
	)
	if !strings.Contains(result, "TCP") || !strings.Contains(result, "UDP") {
		t.Errorf("expected protocols in markdown, got %q", result)
	}
	if !strings.Contains(result, "Router") || !strings.Contains(result, "Server") {
		t.Errorf("expected types in markdown, got %q", result)
	}
}

func TestBuildAnalyticsMarkdown_SortedKeys(t *testing.T) {
	result := buildAnalyticsMarkdown(
		map[string]int{"Zebra": 1, "Alpha": 1},
		nil,
	)
	alphaIdx := strings.Index(result, "Alpha")
	zebraIdx := strings.Index(result, "Zebra")
	if alphaIdx == -1 || zebraIdx == -1 {
		t.Fatalf("expected both keys, got %q", result)
	}
	if alphaIdx > zebraIdx {
		t.Errorf("expected alphabetical order (Alpha before Zebra), got %q", result)
	}
}

func TestBuildResultsAnalyticsView_CacheHit(t *testing.T) {
	a := &App{}
	a.analyticsCacheView = nil
	data := []scanner.Result{
		{IP: "192.168.1.1", DeviceType: "Router", Protocols: []string{"TCP"}},
	}
	first := a.buildResultsAnalyticsView(data)
	if first == nil {
		t.Fatal("expected non-nil view")
	}
	// Второй вызов должен вернуть закешированный view
	second := a.buildResultsAnalyticsView(data)
	if second == nil {
		t.Fatal("expected non-nil view on cache hit")
	}
}

func TestBuildResultsAnalyticsView_EmptyDataCardsMode(t *testing.T) {
	a := &App{}
	a.resultsMode = "Карточки"
	result := a.buildResultsAnalyticsView(nil)
	if result == nil {
		t.Fatal("expected non-nil result for empty data")
	}
}

// --- results_model.go extended tests ---

func TestOpenPortLabels_DefaultLimit(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, Protocol: "tcp", State: "open", Service: "http"},
	}
	labels := openPortLabels(ports, 0)
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0] != "80/TCP http" {
		t.Errorf("expected '80/TCP http', got %q", labels[0])
	}
}

func TestOpenPortLabels_NoOpenPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, Protocol: "tcp", State: "closed"},
	}
	labels := openPortLabels(ports, 10)
	if len(labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(labels))
	}
}

func TestOpenPortLabels_OverLimit(t *testing.T) {
	ports := make([]scanner.PortInfo, 0, 5)
	for i := 0; i < 5; i++ {
		ports = append(ports, scanner.PortInfo{Port: 1000 + i, Protocol: "tcp", State: "open"})
	}
	labels := openPortLabels(ports, 2)
	if len(labels) != 3 {
		t.Errorf("expected 3 labels (2 + overflow marker), got %d", len(labels))
	}
	if labels[2] != "+3" {
		t.Errorf("expected '+3' overflow marker, got %q", labels[2])
	}
}

func TestOpenPortLabels_UnknownService(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 22, Protocol: "tcp", State: "open", Service: "Unknown"},
	}
	labels := openPortLabels(ports, 10)
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0] != "22/TCP" {
		t.Errorf("expected '22/TCP', got %q", labels[0])
	}
}

func TestOpenPortLabels_EmptyProtocol(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 443, State: "open", Service: "https"},
	}
	labels := openPortLabels(ports, 10)
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0] != "443/TCP https" {
		t.Errorf("expected '443/TCP https', got %q", labels[0])
	}
}

func TestOpenPortLabels_MultiWordService(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, Protocol: "tcp", State: "open", Service: "Apache http server"},
	}
	labels := openPortLabels(ports, 10)
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0] != "80/TCP Apache http server" {
		t.Errorf("unexpected label %q", labels[0])
	}
}

func TestFormatPortNumberExtended(t *testing.T) {
	if formatPortNumber(22) != "22" {
		t.Errorf("expected '22', got %q", formatPortNumber(22))
	}
	if formatPortNumber(0) != "0" {
		t.Errorf("expected '0', got %q", formatPortNumber(0))
	}
}

func TestNormalizeServiceName_EmptyExtended(t *testing.T) {
	if normalizeServiceName("") != "" {
		t.Error("expected empty for empty service")
	}
}

func TestNormalizeServiceName_UnknownExtended(t *testing.T) {
	if normalizeServiceName("unknown") != "" {
		t.Error("expected empty for 'unknown'")
	}
	if normalizeServiceName("UNKNOWN") != "" {
		t.Error("expected empty for 'UNKNOWN'")
	}
}

func TestNormalizeServiceName_ValidExtended(t *testing.T) {
	if normalizeServiceName("ssh") != "ssh" {
		t.Errorf("expected 'ssh', got %q", normalizeServiceName("ssh"))
	}
}

func TestCollectAnalytics_WithDataExtended(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", DeviceType: "Router", Protocols: []string{"TCP", "UDP"}},
		{IP: "192.168.1.2", DeviceType: "Router", Protocols: []string{"TCP"}},
		{IP: "192.168.1.3", Protocols: []string{""}},
	}
	protocols, deviceTypes := collectAnalytics(results)
	if protocols["TCP"] != 2 {
		t.Errorf("expected 2 TCP, got %d", protocols["TCP"])
	}
	if protocols["UDP"] != 1 {
		t.Errorf("expected 1 UDP, got %d", protocols["UDP"])
	}
	if deviceTypes["Router"] != 2 {
		t.Errorf("expected 2 Routers, got %d", deviceTypes["Router"])
	}
}

func TestNormalizeDeviceTypes_EmptyExtended(t *testing.T) {
	result := normalizeDeviceTypes(nil)
	if len(result) != 0 {
		t.Errorf("expected 0 types, got %d", len(result))
	}
}

func TestNormalizeDeviceTypes_Mapping(t *testing.T) {
	input := map[string]int{
		"Router":           2,
		"Windows Computer": 3,
		"Web Server":       1,
		"Unknown Device":   4,
	}
	result := normalizeDeviceTypes(input)
	if result["Network Device"] != 2 {
		t.Errorf("expected 2 Network Device, got %d", result["Network Device"])
	}
	if result["Computer"] != 3 {
		t.Errorf("expected 3 Computer, got %d", result["Computer"])
	}
	if result["Server"] != 1 {
		t.Errorf("expected 1 Server, got %d", result["Server"])
	}
	if result["Unknown"] != 4 {
		t.Errorf("expected 4 Unknown, got %d", result["Unknown"])
	}
}

func TestNormalizeDeviceTypes_UnknownKey(t *testing.T) {
	input := map[string]int{"Custom Type": 7}
	result := normalizeDeviceTypes(input)
	if result["Custom Type"] != 7 {
		t.Errorf("expected 7 Custom Type, got %d", result["Custom Type"])
	}
}

func TestNormalizeDeviceTypes_Aggregation(t *testing.T) {
	input := map[string]int{
		"Router":     1,
		"IoT Device": 2,
	}
	result := normalizeDeviceTypes(input)
	if result["Network Device"] != 3 {
		t.Errorf("expected 3 Network Device (aggregated), got %d", result["Network Device"])
	}
}
