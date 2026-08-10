package gui

import (
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

// --- D3.1: GUI-smoke сценарии переключения режимов ---

func TestSortedResultsConsistency(t *testing.T) {
	// Проверка что сортировка детерминирована и стабильна
	results := []scanner.Result{
		{IP: "192.168.1.10", Hostname: "host10", DeviceType: "host"},
		{IP: "192.168.1.1", Hostname: "host1", DeviceType: "router"},
		{IP: "192.168.1.2", Hostname: "host2", DeviceType: "switch"},
		{IP: "10.0.0.1", Hostname: "server", DeviceType: "server"},
	}

	sorted := sortedResultsForDisplay(results)

	// Первая должна быть 10.0.0.1 (наименьший IP)
	if sorted[0].IP != "10.0.0.1" {
		t.Errorf("expected first IP 10.0.0.1, got %s", sorted[0].IP)
	}

	// Проверка стабильности: повторная сортировка даёт тот же результат
	sorted2 := sortedResultsForDisplay(results)
	for i := range sorted {
		if sorted[i].IP != sorted2[i].IP {
			t.Errorf("sorting unstable at index %d: %s vs %s", i, sorted[i].IP, sorted2[i].IP)
		}
	}
}

func TestSortedResultsHostNameMode(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.10", Hostname: "Zebra", DeviceType: "host"},
		{IP: "192.168.1.1", Hostname: "alpha", DeviceType: "router"},
		{IP: "192.168.1.2", Hostname: "beta", DeviceType: "switch"},
	}

	sorted := sortedResultsForDisplayWithMode(results, "HostName")

	// alpha < beta < Zebra (case-insensitive)
	if sorted[0].Hostname != "alpha" {
		t.Errorf("expected first hostname alpha, got %s", sorted[0].Hostname)
	}
	if sorted[1].Hostname != "beta" {
		t.Errorf("expected second hostname beta, got %s", sorted[1].Hostname)
	}
}

func TestFilterResultsEmptyQuery(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "router"},
		{IP: "192.168.1.2", Hostname: "switch", DeviceType: "switch"},
	}

	filtered := filterResultsForDisplay(results, "")

	if len(filtered) != len(results) {
		t.Errorf("expected %d results with empty query, got %d", len(results), len(filtered))
	}
}

func TestFilterResultsByHostname(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router-main", DeviceType: "router"},
		{IP: "192.168.1.2", Hostname: "switch-core", DeviceType: "switch"},
		{IP: "192.168.1.3", Hostname: "router-backup", DeviceType: "router"},
	}

	filtered := filterResultsForDisplay(results, "router")

	if len(filtered) != 2 {
		t.Errorf("expected 2 results for 'router', got %d", len(filtered))
	}
}

func TestFilterResultsByIP(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "router"},
		{IP: "10.0.0.1", Hostname: "server", DeviceType: "server"},
	}

	filtered := filterResultsForDisplay(results, "10.0.0")

	if len(filtered) != 1 {
		t.Errorf("expected 1 result for '10.0.0', got %d", len(filtered))
	}
	if filtered[0].IP != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1, got %s", filtered[0].IP)
	}
}

func TestFilterResultsAdvancedWithTypes(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "router"},
		{IP: "192.168.1.2", Hostname: "switch", DeviceType: "switch"},
		{IP: "192.168.1.3", Hostname: "host", DeviceType: "host"},
	}

	filtered := filterResultsForDisplayAdvanced(results, "", []string{"router", "switch"}, false)

	if len(filtered) != 2 {
		t.Errorf("expected 2 results for router+switch, got %d", len(filtered))
	}
}

func TestFilterResultsAdvancedWithOpenPorts(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.1", Hostname: "router",
			DeviceType: "router",
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Service: "ssh"},
			},
		},
		{
			IP: "192.168.1.2", Hostname: "closed-host",
			DeviceType: "host",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "closed", Service: "http"},
			},
		},
	}

	filtered := filterResultsForDisplayAdvanced(results, "", []string{}, true)

	if len(filtered) != 1 {
		t.Errorf("expected 1 result with open ports only, got %d", len(filtered))
	}
	if filtered[0].Hostname != "router" {
		t.Errorf("expected router, got %s", filtered[0].Hostname)
	}
}

// --- D3.2: Visual baseline tests — port chips formatting ---

func TestOpenPortLabelsEmpty(t *testing.T) {
	labels := openPortLabels([]scanner.PortInfo{}, 10)
	if len(labels) != 0 {
		t.Errorf("expected 0 labels for empty ports, got %d", len(labels))
	}
}

func TestOpenPortLabelsBasic(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
		{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
		{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
	}

	labels := openPortLabels(ports, 10)

	if len(labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(labels))
	}
}

func TestOpenPortLabelsTruncation(t *testing.T) {
	ports := make([]scanner.PortInfo, 0, 30)
	for i := 0; i < 30; i++ {
		ports = append(ports, scanner.PortInfo{
			Port:     1000 + i,
			State:    "open",
			Protocol: "tcp",
			Service:  "service",
		})
	}

	labels := openPortLabels(ports, 10)

	// Должно быть 10 портов + "+20"
	if len(labels) != 11 {
		t.Errorf("expected 11 labels (10 ports + overflow), got %d", len(labels))
	}
	if labels[10] != "+20" {
		t.Errorf("expected '+20' overflow label, got %s", labels[10])
	}
}

func TestOpenPortLabelsNonOpenSkipped(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
		{Port: 80, State: "closed", Protocol: "tcp", Service: "http"},
		{Port: 443, State: "filtered", Protocol: "tcp", Service: "https"},
	}

	labels := openPortLabels(ports, 10)

	if len(labels) != 1 {
		t.Errorf("expected 1 label (only open), got %d", len(labels))
	}
}

func TestOpenPortLabelsDefaultMax(t *testing.T) {
	ports := make([]scanner.PortInfo, 0, 30)
	for i := 0; i < 30; i++ {
		ports = append(ports, scanner.PortInfo{
			Port:     1000 + i,
			State:    "open",
			Protocol: "tcp",
			Service:  "service",
		})
	}

	// maxVisiblePorts <= 0 должен использовать default 24
	labels := openPortLabels(ports, 0)

	if len(labels) != 25 { // 24 ports + "+6" overflow
		t.Errorf("expected 25 labels with default max, got %d", len(labels))
	}
}

func TestFormatPortNumber(t *testing.T) {
	if formatPortNumber(22) != "22" {
		t.Error("formatPortNumber(22) should return '22'")
	}
	if formatPortNumber(8080) != "8080" {
		t.Error("formatPortNumber(8080) should return '8080'")
	}
}

func TestNormalizeServiceName(t *testing.T) {
	if normalizeServiceName("ssh") != "ssh" {
		t.Error("normalizeServiceName('ssh') should return 'ssh'")
	}
	if normalizeServiceName("unknown") != "" {
		t.Error("normalizeServiceName('unknown') should return ''")
	}
	if normalizeServiceName("") != "" {
		t.Error("normalizeServiceName('') should return ''")
	}
	if normalizeServiceName("  HTTP  ") != "HTTP" {
		t.Error("normalizeServiceName should trim whitespace")
	}
}

func TestFormatDeviceValue(t *testing.T) {
	if formatDeviceValue("router") != "router" {
		t.Error("formatDeviceValue('router') should return 'router'")
	}
	if formatDeviceValue("") != "-" {
		t.Error("formatDeviceValue('') should return '-'")
	}
	if formatDeviceValue("  ") != "-" {
		t.Error("formatDeviceValue('  ') should return '-'")
	}
}

// --- D3.3: Cross-check metrics (analytics consistency) ---

func TestCollectAnalyticsEmpty(t *testing.T) {
	protocols, deviceTypes := collectAnalytics([]scanner.Result{})

	if len(protocols) != 0 {
		t.Errorf("expected 0 protocols for empty results, got %d", len(protocols))
	}
	if len(deviceTypes) != 0 {
		t.Errorf("expected 0 deviceTypes for empty results, got %d", len(deviceTypes))
	}
}

func TestCollectAnalyticsBasic(t *testing.T) {
	results := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Protocols:  []string{"tcp", "udp"},
			DeviceType: "router",
		},
		{
			IP:         "192.168.1.2",
			Protocols:  []string{"tcp"},
			DeviceType: "switch",
		},
	}

	protocols, deviceTypes := collectAnalytics(results)

	if protocols["tcp"] != 2 {
		t.Errorf("expected tcp=2, got %d", protocols["tcp"])
	}
	if protocols["udp"] != 1 {
		t.Errorf("expected udp=1, got %d", protocols["udp"])
	}
	if deviceTypes["router"] != 1 {
		t.Errorf("expected router=1, got %d", deviceTypes["router"])
	}
	if deviceTypes["switch"] != 1 {
		t.Errorf("expected switch=1, got %d", deviceTypes["switch"])
	}
}

func TestCollectAnalyticsEmptyFields(t *testing.T) {
	results := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Protocols:  []string{"", "tcp", ""},
			DeviceType: "",
		},
	}

	protocols, deviceTypes := collectAnalytics(results)

	if protocols["tcp"] != 1 {
		t.Errorf("expected tcp=1, got %d", protocols["tcp"])
	}
	if len(deviceTypes) != 0 {
		t.Errorf("expected 0 deviceTypes for empty deviceType, got %d", len(deviceTypes))
	}
}

func TestNormalizeDeviceTypesD3(t *testing.T) {
	raw := map[string]int{
		"Router":                2,
		"Router/Network Device": 1,
		"Windows Computer":      3,
		"Linux Server":          2,
		"Unknown":               1,
	}

	normalized := normalizeDeviceTypes(raw)

	if normalized["Network Device"] != 3 { // Router x2 + Router/Network Device x1
		t.Errorf("expected Network Device=3, got %d", normalized["Network Device"])
	}
	if normalized["Computer"] != 3 {
		t.Errorf("expected Computer=3, got %d", normalized["Computer"])
	}
	if normalized["Server"] != 2 {
		t.Errorf("expected Server=2, got %d", normalized["Server"])
	}
	if normalized["Unknown"] != 1 {
		t.Errorf("expected Unknown=1, got %d", normalized["Unknown"])
	}
}

func TestHasOpenPorts(t *testing.T) {
	if !hasOpenPorts([]scanner.PortInfo{
		{Port: 22, State: "open"},
	}) {
		t.Error("hasOpenPorts should return true for open port")
	}

	if hasOpenPorts([]scanner.PortInfo{
		{Port: 80, State: "closed"},
	}) {
		t.Error("hasOpenPorts should return false for closed port")
	}

	if hasOpenPorts([]scanner.PortInfo{
		{Port: 443, State: "filtered"},
	}) {
		t.Error("hasOpenPorts should return false for filtered port")
	}

	if hasOpenPorts([]scanner.PortInfo{}) {
		t.Error("hasOpenPorts should return false for empty ports")
	}
}

// --- D3.4: Perf-budget smoke test ---

func BenchmarkSortedResults10(b *testing.B) {
	results := make([]scanner.Result, 10)
	for i := range results {
		results[i] = scanner.Result{
			IP:         "192.168.1." + string(rune('0'+i)),
			Hostname:   "host" + string(rune('0'+i)),
			DeviceType: "host",
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedResultsForDisplay(results)
	}
}

func BenchmarkSortedResults100(b *testing.B) {
	results := make([]scanner.Result, 100)
	for i := range results {
		results[i] = scanner.Result{
			IP:         "192.168.1." + string(rune('0'+i%10)),
			Hostname:   "host" + string(rune('0'+i%10)),
			DeviceType: "host",
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedResultsForDisplay(results)
	}
}

func BenchmarkSortedResults1000(b *testing.B) {
	results := make([]scanner.Result, 1000)
	for i := range results {
		results[i] = scanner.Result{
			IP:         "192.168.1." + string(rune('0'+i%10)),
			Hostname:   "host" + string(rune('0'+i%10)),
			DeviceType: "host",
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedResultsForDisplay(results)
	}
}

func BenchmarkCollectAnalytics100(b *testing.B) {
	results := make([]scanner.Result, 100)
	for i := range results {
		results[i] = scanner.Result{
			IP:         "192.168.1.1",
			Protocols:  []string{"tcp", "udp"},
			DeviceType: "router",
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collectAnalytics(results)
	}
}

func TestPerfBudgetSortedResults(t *testing.T) {
	// Perf-budget: сортировка 1000 результатов должна занимать < 1ms
	results := make([]scanner.Result, 1000)
	for i := range results {
		results[i] = scanner.Result{
			IP:         "192.168.1." + string(rune('0'+i%10)),
			Hostname:   "host" + string(rune('0'+i%10)),
			DeviceType: "host",
		}
	}

	start := time.Now()
	for i := 0; i < 100; i++ {
		sortedResultsForDisplay(results)
	}
	elapsed := time.Since(start) / 100 // average per sort

	budget := 1 * time.Millisecond
	if elapsed > budget {
		t.Errorf("perf budget exceeded: sort took %v (budget %v)", elapsed, budget)
	}
}

func TestPerfBudgetCollectAnalytics(t *testing.T) {
	// Perf-budget: сбор аналитики 1000 результатов должен занимать < 1ms
	results := make([]scanner.Result, 1000)
	for i := range results {
		results[i] = scanner.Result{
			IP:         "192.168.1.1",
			Protocols:  []string{"tcp", "udp"},
			DeviceType: "router",
		}
	}

	start := time.Now()
	for i := 0; i < 100; i++ {
		collectAnalytics(results)
	}
	elapsed := time.Since(start) / 100 // average per collect

	budget := 1 * time.Millisecond
	if elapsed > budget {
		t.Errorf("perf budget exceeded: collectAnalytics took %v (budget %v)", elapsed, budget)
	}
}

// --- D3.5: Responsive behavior tests ---

func TestOpenPortLabelsNarrowView(t *testing.T) {
	// Узкий вид: макс 5 портов
	ports := make([]scanner.PortInfo, 0, 20)
	for i := 0; i < 20; i++ {
		ports = append(ports, scanner.PortInfo{
			Port:     1000 + i,
			State:    "open",
			Protocol: "tcp",
			Service:  "service",
		})
	}

	labels := openPortLabels(ports, 5)

	if len(labels) != 6 { // 5 ports + "+15" overflow
		t.Errorf("expected 6 labels for narrow view, got %d", len(labels))
	}
}

func TestOpenPortLabelsWideView(t *testing.T) {
	// Широкий вид: макс 50 портов
	ports := make([]scanner.PortInfo, 0, 30)
	for i := 0; i < 30; i++ {
		ports = append(ports, scanner.PortInfo{
			Port:     1000 + i,
			State:    "open",
			Protocol: "tcp",
			Service:  "service",
		})
	}

	labels := openPortLabels(ports, 50)

	// Все 30 портов должны поместиться
	if len(labels) != 30 {
		t.Errorf("expected 30 labels for wide view, got %d", len(labels))
	}
}

func TestOpenPortLabelsMediumView(t *testing.T) {
	// Средний вид: макс 15 портов
	ports := make([]scanner.PortInfo, 0, 20)
	for i := 0; i < 20; i++ {
		ports = append(ports, scanner.PortInfo{
			Port:     1000 + i,
			State:    "open",
			Protocol: "tcp",
			Service:  "service",
		})
	}

	labels := openPortLabels(ports, 15)

	if len(labels) != 16 { // 15 ports + "+5" overflow
		t.Errorf("expected 16 labels for medium view, got %d", len(labels))
	}
}
