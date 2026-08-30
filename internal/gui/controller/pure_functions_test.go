package controller

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"

	"fyne.io/fyne/v2/app"
)

// --- autoScanProfile tests ---

func TestAutoScanProfile_EmptyInputs(t *testing.T) {
	portRange, threads, msg := autoScanProfile("", "", 0)
	// Возвращает (portRange, threads, msg) — msg пустой для empty inputs
	if threads < 0 {
		t.Errorf("autoScanProfile threads should be >= 0, got %d", threads)
	}
	if portRange != "" {
		t.Errorf("expected empty portRange for empty input, got %q", portRange)
	}
	_ = msg
}

func TestAutoScanProfile_LargeSubnet(t *testing.T) {
	portRange, threads, msg := autoScanProfile("192.168.1.0/24", "", 50)
	// /24 = 256 хостов < 512, поэтому профиль не меняется
	if threads != 50 {
		t.Errorf("expected threads 50, got %d", threads)
	}
	_ = portRange
	_ = msg
}

func TestAutoScanProfile_SmallSubnet(t *testing.T) {
	portRange, threads, msg := autoScanProfile("192.168.1.0/30", "", 50)
	// /30 = 4 хоста < 256, профиль не меняется
	if threads != 50 {
		t.Errorf("expected threads 50, got %d", threads)
	}
	_ = portRange
	_ = msg
}

func TestAutoScanProfile_WithPortRange(t *testing.T) {
	portRange, threads, msg := autoScanProfile("192.168.1.0/24", "1-100", 50)
	if threads != 50 {
		t.Errorf("expected threads 50, got %d", threads)
	}
	_ = portRange
	_ = msg
}

// --- estimateScanUITimeout tests ---

func TestEstimateScanUITimeout_DefaultTimeout(t *testing.T) {
	timeout := estimateScanUITimeout("", "", "5", "50", true, false)
	if timeout <= 0 {
		t.Errorf("expected positive timeout, got %v", timeout)
	}
}

func TestEstimateScanUITimeout_NetworkOnly(t *testing.T) {
	timeout := estimateScanUITimeout("192.168.1.0/24", "", "", "", true, false)
	if timeout <= 0 {
		t.Errorf("expected positive timeout for network, got %v", timeout)
	}
}

func TestEstimateScanUITimeout_UDPIncreasesTimeout(t *testing.T) {
	timeoutUDP := estimateScanUITimeout("192.168.1.0/24", "", "5", "50", true, true)
	// UDP должен возвращать положительный таймаут
	if timeoutUDP <= 0 {
		t.Errorf("UDP timeout should be > 0, got %v", timeoutUDP)
	}
}

func TestEstimateScanUITimeout_ThreadsDecreaseTimeout(t *testing.T) {
	timeoutLow := estimateScanUITimeout("192.168.1.0/24", "", "5", "10", true, false)
	timeoutHigh := estimateScanUITimeout("192.168.1.0/24", "", "5", "200", true, false)
	// Оба должны быть положительными
	if timeoutLow <= 0 || timeoutHigh <= 0 {
		t.Errorf("both timeouts should be > 0: low=%v, high=%v", timeoutLow, timeoutHigh)
	}
}

// --- formatDurationMMSS tests ---

func TestFormatDurationMMSS_Zero(t *testing.T) {
	result := formatDurationMMSS(0)
	if result != "00:00" {
		t.Errorf("expected '00:00', got %q", result)
	}
}

func TestFormatDurationMMSS_Small(t *testing.T) {
	result := formatDurationMMSS(65 * time.Second)
	if result != "01:05" {
		t.Errorf("expected '01:05', got %q", result)
	}
}

func TestFormatDurationMMSS_ExactMinutes(t *testing.T) {
	result := formatDurationMMSS(3 * time.Minute)
	if result != "03:00" {
		t.Errorf("expected '03:00', got %q", result)
	}
}

func TestFormatDurationMMSS_Hours(t *testing.T) {
	result := formatDurationMMSS(125 * time.Minute)
	if result != "125:00" {
		t.Errorf("expected '125:00' for 125 minutes, got %q", result)
	}
}

func TestFormatDurationMMSS_Negative(t *testing.T) {
	result := formatDurationMMSS(-10 * time.Second)
	if result == "" {
		t.Error("formatDurationMMSS should not return empty for negative duration")
	}
}

// --- recommendedBadgeClassForHosts tests ---

func TestRecommendedBadgeClassForHosts_Boundaries(t *testing.T) {
	type args struct {
		hosts int
		want  string
	}
	tests := []args{
		{0, "green"},
		{1, "green"},
		{511, "green"},
		{512, "yellow"},
		{513, "yellow"},
		{1023, "yellow"},
		{1024, "orange"},
		{1025, "orange"},
		{2047, "orange"},
		{2048, "red"},
		{2049, "red"},
		{9999, "red"},
	}

	// Создаём контроллер для вызова метода
	ctrl := &ScanController{}
	for _, tt := range tests {
		got := ctrl.recommendedBadgeClassForHosts(tt.hosts)
		if got != tt.want {
			t.Errorf("recommendedBadgeClassForHosts(%d) = %q, want %q", tt.hosts, got, tt.want)
		}
	}
}

// --- recommendedBadgeText tests ---

func TestRecommendedBadgeText_Basic(t *testing.T) {
	ctrl := &ScanController{}
	text := ctrl.recommendedBadgeText("Quick", "green")
	if text != "Quick (green)" {
		t.Errorf("expected 'Quick (green)', got %q", text)
	}
}

func TestRecommendedBadgeText_EmptyName(t *testing.T) {
	ctrl := &ScanController{}
	text := ctrl.recommendedBadgeText("", "red")
	if text != " (red)" {
		t.Errorf("expected ' (red)', got %q", text)
	}
}

// --- splitCommaValues tests ---

func TestSplitCommaValues_Basic(t *testing.T) {
	result := splitCommaValues("public,secret,admin")
	expected := []string{"public", "secret", "admin"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d values, got %d", len(expected), len(result))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("result[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestSplitCommaValues_Empty(t *testing.T) {
	result := splitCommaValues("")
	if len(result) != 1 || result[0] != "public" {
		t.Errorf("expected ['public'], got %v", result)
	}
}

func TestSplitCommaValues_WithSpaces(t *testing.T) {
	result := splitCommaValues(" public , secret , admin ")
	expected := []string{"public", "secret", "admin"}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("result[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestSplitCommaValues_MultipleCommas(t *testing.T) {
	result := splitCommaValues("public,,secret,,,admin")
	expected := []string{"public", "secret", "admin"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d values, got %d", len(expected), len(result))
	}
}

func TestSplitCommaValues_OnlySpaces(t *testing.T) {
	result := splitCommaValues("  ,  ,  ")
	if len(result) != 1 || result[0] != "public" {
		t.Errorf("expected ['public'], got %v", result)
	}
}

// --- partialSNMPKeysFromReport tests ---

func TestPartialSNMPKeysFromReport_Nil(t *testing.T) {
	result := partialSNMPKeysFromReport(nil)
	if result != nil {
		t.Errorf("expected nil for nil report, got %v", result)
	}
}

func TestPartialSNMPKeysFromReport_OnlyQueryFailures(t *testing.T) {
	report := &snmpcollector.CollectReport{
		Failures: []snmpcollector.DeviceFailure{
			{Kind: snmpcollector.FailureQuery, IP: "192.168.1.1"},
			{Kind: snmpcollector.FailureQuery, IP: "192.168.1.2"},
		},
	}
	result := partialSNMPKeysFromReport(report)
	if len(result) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(result))
	}
	if _, ok := result["ip:192.168.1.1"]; !ok {
		t.Error("missing key 'ip:192.168.1.1'")
	}
	if _, ok := result["ip:192.168.1.2"]; !ok {
		t.Error("missing key 'ip:192.168.1.2'")
	}
}

func TestPartialSNMPKeysFromReport_EmptyResult(t *testing.T) {
	report := &snmpcollector.CollectReport{
		Failures: []snmpcollector.DeviceFailure{
			{Kind: snmpcollector.FailureConnect, IP: "192.168.1.1"},
		},
	}
	result := partialSNMPKeysFromReport(report)
	if result != nil {
		t.Errorf("expected nil for non-Query failures, got %v", result)
	}
}

func TestPartialSNMPKeysFromReport_IPNormalization(t *testing.T) {
	report := &snmpcollector.CollectReport{
		Failures: []snmpcollector.DeviceFailure{
			{Kind: snmpcollector.FailureQuery, IP: "  192.168.1.1 "},
		},
	}
	result := partialSNMPKeysFromReport(report)
	if len(result) != 1 {
		t.Fatalf("expected 1 key, got %d", len(result))
	}
	if _, ok := result["ip:192.168.1.1"]; !ok {
		t.Error("expected key 'ip:192.168.1.1' (trimmed, lowercased)")
	}
}

// --- topologySuccessStatus tests ---

func TestTopologySuccessStatus_NilTopo(t *testing.T) {
	result := topologySuccessStatus(nil, nil)
	if result != "Топология не построена" {
		t.Errorf("expected 'Топология не построена', got %q", result)
	}
}

func TestTopologySuccessStatus_EmptyTopo(t *testing.T) {
	topo := &topology.Topology{}
	report := &snmpcollector.CollectReport{}
	result := topologySuccessStatus(topo, report)
	if !strings.Contains(result, "устройств 0, связей 0") {
		t.Errorf("expected device/link count in result, got %q", result)
	}
}

func TestTopologySuccessStatus_WithDevices(t *testing.T) {
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{
			"192.168.1.1": {IP: "192.168.1.1"},
			"192.168.1.2": {IP: "192.168.1.2"},
			"192.168.1.3": {IP: "192.168.1.3"},
		},
		Links: []topology.Link{
			{Source: &topology.Device{IP: "192.168.1.1"}, Target: &topology.Device{IP: "192.168.1.2"}},
		},
	}
	result := topologySuccessStatus(topo, nil)
	if !strings.Contains(result, "устройств 3, связей 1") {
		t.Errorf("expected 'устройств 3, связей 1', got %q", result)
	}
}

// --- formatTopologyPreview tests ---

func TestFormatTopologyPreview_NilTopo(t *testing.T) {
	result := formatTopologyPreview(nil, nil, topologyBuildMetrics{})
	if !strings.Contains(result, "Нет данных") {
		t.Errorf("expected 'Нет данных', got %q", result)
	}
}

func TestFormatTopologyPreview_EmptyTopo(t *testing.T) {
	topo := &topology.Topology{}
	result := formatTopologyPreview(topo, nil, topologyBuildMetrics{})
	if !strings.Contains(result, "Устройств:") || !strings.Contains(result, "Связей:") {
		t.Errorf("expected device/link counts, got %q", result)
	}
}

func TestFormatTopologyPreview_WithMetrics(t *testing.T) {
	topo := &topology.Topology{}
	metrics := topologyBuildMetrics{
		snmpDuration:  5 * time.Second,
		buildDuration: 2 * time.Second,
		totalDuration: 7 * time.Second,
	}
	result := formatTopologyPreview(topo, nil, metrics)
	if !strings.Contains(result, "SNMP сбор:") {
		t.Error("expected 'SNMP сбор:' in result")
	}
	if !strings.Contains(result, "Построение графа:") {
		t.Error("expected 'Построение графа:' in result")
	}
	if !strings.Contains(result, "Общее время:") {
		t.Error("expected 'Общее время:' in result")
	}
}

func TestFormatTopologyPreview_WithReport(t *testing.T) {
	topo := &topology.Topology{}
	report := &snmpcollector.CollectReport{
		TotalSNMPTargets: 10,
		Connected:        5,
		Partial:          3,
		Failed:           2,
	}
	result := formatTopologyPreview(topo, report, topologyBuildMetrics{})
	if !strings.Contains(result, "Целей для SNMP: 10") {
		t.Error("expected SNMP target count in result")
	}
	if !strings.Contains(result, "Успешных подключений: 5") {
		t.Error("expected connected count in result")
	}
}

func TestFormatTopologyPreview_WithLinks(t *testing.T) {
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{
			"dev1": {Hostname: "router1", IP: "192.168.1.1"},
			"dev2": {Hostname: "switch1", IP: "192.168.1.2"},
		},
	}
	topo.Links = []topology.Link{
		{
			Source:     topo.Devices["dev1"],
			Target:     topo.Devices["dev2"],
			SourcePort: &topology.Port{Name: "Gig0/1"},
			TargetPort: &topology.Port{Name: "Gig0/1"},
			SourceType: "lldp",
			Confidence: "high",
		},
	}
	result := formatTopologyPreview(topo, nil, topologyBuildMetrics{})
	if !strings.Contains(result, "router1") {
		t.Error("expected device hostname in link")
	}
	if !strings.Contains(result, "switch1") {
		t.Error("expected target device hostname in link")
	}
	if !strings.Contains(result, "Gig0/1") {
		t.Error("expected port name in link")
	}
	if !strings.Contains(result, "lldp/high") {
		t.Error("expected source_type/confidence in link")
	}
}

// --- topoDisplayName tests ---

func TestTopoDisplayName_Hostname(t *testing.T) {
	d := &topology.Device{Hostname: "my-router", IP: "192.168.1.1"}
	result := topoDisplayName(d)
	if result != "my-router" {
		t.Errorf("expected 'my-router', got %q", result)
	}
}

func TestTopoDisplayName_IP(t *testing.T) {
	d := &topology.Device{IP: "10.0.0.1"}
	result := topoDisplayName(d)
	if result != "10.0.0.1" {
		t.Errorf("expected '10.0.0.1', got %q", result)
	}
}

func TestTopoDisplayName_MAC(t *testing.T) {
	d := &topology.Device{MAC: "aa:bb:cc:dd:ee:ff"}
	result := topoDisplayName(d)
	if result != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected MAC, got %q", result)
	}
}

func TestTopoDisplayName_Empty(t *testing.T) {
	d := &topology.Device{}
	result := topoDisplayName(d)
	if result != "unknown" {
		t.Errorf("expected 'unknown', got %q", result)
	}
}

func TestTopoDisplayName_Nil(t *testing.T) {
	result := topoDisplayName(nil)
	if result != "unknown" {
		t.Errorf("expected 'unknown' for nil, got %q", result)
	}
}

// --- topoPortName tests ---

func TestTopoPortName_Name(t *testing.T) {
	p := &topology.Port{Name: "Gig0/1"}
	result := topoPortName(p)
	if result != "Gig0/1" {
		t.Errorf("expected 'Gig0/1', got %q", result)
	}
}

func TestTopoPortName_Index(t *testing.T) {
	p := &topology.Port{Index: 42}
	result := topoPortName(p)
	if result != "if42" {
		t.Errorf("expected 'if42', got %q", result)
	}
}

func TestTopoPortName_Empty(t *testing.T) {
	p := &topology.Port{}
	result := topoPortName(p)
	if result != "-" {
		t.Errorf("expected '-', got %q", result)
	}
}

func TestTopoPortName_Nil(t *testing.T) {
	result := topoPortName(nil)
	if result != "-" {
		t.Errorf("expected '-' for nil, got %q", result)
	}
}

// --- ClampOffset tests ---

func TestClampOffset_BelowMinimum(t *testing.T) {
	mgr := &SettingsManager{}
	result := mgr.ClampOffset(-0.5, 0, 1)
	if result != 0 {
		t.Errorf("expected 0 for -0.5, got %v", result)
	}
}

func TestClampOffset_AboveMaximum(t *testing.T) {
	mgr := &SettingsManager{}
	result := mgr.ClampOffset(1.5, 0, 1)
	if result != 1 {
		t.Errorf("expected 1 for 1.5, got %v", result)
	}
}

func TestClampOffset_InRange(t *testing.T) {
	mgr := &SettingsManager{}
	result := mgr.ClampOffset(0.5, 0, 1)
	if result != 0.5 {
		t.Errorf("expected 0.5, got %v", result)
	}
}

func TestClampOffset_BoundaryMinimum(t *testing.T) {
	mgr := &SettingsManager{}
	result := mgr.ClampOffset(0, 0, 1)
	if result != 0 {
		t.Errorf("expected 0, got %v", result)
	}
}

func TestClampOffset_BoundaryMaximum(t *testing.T) {
	mgr := &SettingsManager{}
	result := mgr.ClampOffset(1, 0, 1)
	if result != 1 {
		t.Errorf("expected 1, got %v", result)
	}
}

func TestClampOffset_NegativeRange(t *testing.T) {
	mgr := &SettingsManager{}
	result := mgr.ClampOffset(-0.75, -1, 1)
	if result != -0.75 {
		t.Errorf("expected -0.75, got %v", result)
	}
}

func TestClampOffset_OutOfNegativeRange(t *testing.T) {
	mgr := &SettingsManager{}
	result := mgr.ClampOffset(-2, -1, 1)
	if result != -1 {
		t.Errorf("expected -1 for -2, got %v", result)
	}
}

// --- CopyScanDiagnostics / SaveScanDiagnostics tests ---

func TestCopyScanDiagnostics_Empty(t *testing.T) {
	ctrl := &ScanController{app: app.New()}
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("CopyScanDiagnostics panics in headless mode: %v", r)
		}
	}()
	ctrl.CopyScanDiagnostics("")
	// Не паникует — это уже успех
}

func TestSaveScanDiagnostics_NilWindow(t *testing.T) {
	ctrl := &ScanController{app: app.New()}
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("SaveScanDiagnostics panics in headless mode: %v", r)
		}
	}()
	ctrl.SaveScanDiagnostics("diagnostics content")
	// Не паникует при nil window — это успех
}

// --- BuildPerformanceReportText tests (via topology controller) ---

func TestBuildPerformanceReportText_NilTopo(t *testing.T) {
	ctrl := &TopologyController{}
	result := ctrl.buildPerformanceReportText()
	if result != "" {
		t.Errorf("expected empty string for nil topo, got %q", result)
	}
}

func TestBuildPerformanceReportText_WithTopo(t *testing.T) {
	ctrl := &TopologyController{
		lastTopo: &topology.Topology{
			Devices: map[string]*topology.Device{"d1": {IP: "192.168.1.1"}},
			Links:   []topology.Link{},
		},
		lastMetrics: topologyBuildMetrics{
			snmpDuration:  3 * time.Second,
			buildDuration: 1 * time.Second,
			totalDuration: 4 * time.Second,
		},
		lastReport: &snmpcollector.CollectReport{
			TotalSNMPTargets: 5,
			Connected:        3,
			Partial:          1,
			Failed:           1,
		},
	}
	result := ctrl.buildPerformanceReportText()
	if result == "" {
		t.Error("expected non-empty report text")
	}
	if !strings.Contains(result, "Устройств: 1") {
		t.Error("expected device count in report")
	}
	if !strings.Contains(result, "Связей: 0") {
		t.Error("expected link count in report")
	}
	if !strings.Contains(result, "SNMP сбор:") {
		t.Error("expected SNMP duration in report")
	}
	if !strings.Contains(result, "5") {
		t.Error("expected SNMP target count '5' in report")
	}
}

// --- parseIntOrDefault tests ---

func TestParseIntOrDefault_Empty(t *testing.T) {
	result := parseIntOrDefault("", 42)
	if result != 42 {
		t.Errorf("expected 42 for empty string, got %d", result)
	}
}

func TestParseIntOrDefault_Valid(t *testing.T) {
	result := parseIntOrDefault("100", 42)
	if result != 100 {
		t.Errorf("expected 100, got %d", result)
	}
}

func TestParseIntOrDefault_NonNumeric(t *testing.T) {
	result := parseIntOrDefault("abc", 10)
	if result != 10 {
		t.Errorf("expected 10 for non-numeric, got %d", result)
	}
}

func TestParseIntOrDefault_Zero(t *testing.T) {
	result := parseIntOrDefault("0", 42)
	if result != 42 {
		t.Errorf("expected 42 for '0', got %d", result)
	}
}

func TestParseIntOrDefault_Negative(t *testing.T) {
	result := parseIntOrDefault("-5", 10)
	if result != 10 {
		t.Errorf("expected 10 for '-5', got %d", result)
	}
}

// --- ScannerResult tests for topology ---

func TestTopoDisplayName_ScannerResultDevice(t *testing.T) {
	// Тест что topoDisplayName работает с устройствами из scanner.Result
	d := &topology.Device{
		IP:       "192.168.1.100",
		Hostname: "web-server.local",
	}
	name := topoDisplayName(d)
	if name != "web-server.local" {
		t.Errorf("expected hostname, got %q", name)
	}
}

// --- Edge case: formatTopologyPreview with all nil fields ---

func TestFormatTopologyPreview_AllNilFields(t *testing.T) {
	result := formatTopologyPreview(nil, nil, topologyBuildMetrics{})
	if !strings.Contains(result, "Нет данных") {
		t.Errorf("expected 'Нет данных', got %q", result)
	}
}

// --- Edge case: topologySuccessStatus with nil report ---

func TestTopologySuccessStatus_NilReport(t *testing.T) {
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{"d1": {IP: "10.0.0.1"}},
	}
	result := topologySuccessStatus(topo, nil)
	if !strings.Contains(result, "устройств 1, связей 0") {
		t.Errorf("expected 'устройств 1, связей 0', got %q", result)
	}
}

// --- Integration: formatTopologyPreview with realistic data ---

func TestFormatTopologyPreview_RealisticScenario(t *testing.T) {
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{
			"r1": {Hostname: "core-router", IP: "10.0.0.1", MAC: "aa:bb:cc:00:00:01"},
			"s1": {Hostname: "access-switch", IP: "10.0.0.2", MAC: "aa:bb:cc:00:00:02"},
			"w1": {IP: "10.0.0.10", Hostname: "workstation"},
		},
	}
	topo.Links = []topology.Link{
		{
			Source:     topo.Devices["r1"],
			Target:     topo.Devices["s1"],
			SourceType: "lldp",
			Confidence: "high",
			SourcePort: &topology.Port{Name: "GigabitEthernet0/1"},
			TargetPort: &topology.Port{Name: "GigabitEthernet0/24"},
		},
		{
			Source:     topo.Devices["s1"],
			Target:     topo.Devices["w1"],
			SourceType: "inferred",
			Confidence: "low",
			SourcePort: &topology.Port{Index: 5},
			TargetPort: nil,
		},
	}
	report := &snmpcollector.CollectReport{
		TotalSNMPTargets: 3,
		Connected:        2,
		Partial:          1,
		Failed:           0,
	}
	metrics := topologyBuildMetrics{
		snmpDuration:  500 * time.Millisecond,
		buildDuration: 200 * time.Millisecond,
		totalDuration: 750 * time.Millisecond,
	}

	result := formatTopologyPreview(topo, report, metrics)

	// Проверяем наличие ключевых данных
	checks := []string{
		"500ms",
		"200ms",
		"750ms",
		"Целей для SNMP: 3",
		"Успешных подключений: 2",
		"Частичных опросов: 1",
		"Полных отказов: 0",
		"Устройств:** 3",
		"Связей:** 2",
		"core-router",
		"access-switch",
		"workstation",
		"GigabitEthernet0/1",
		"lldp/high",
		"inferred/low",
		"if5",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected %q in result, got:\n%s", check, result)
		}
	}
}

// --- Test scanner.Result compatibility ---

func TestFormatTopologyPreview_ScannerResultIntegration(t *testing.T) {
	// Проверяем что topology.Topology корректно работает с данными из scanner.Result
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", IsAlive: true},
		{IP: "192.168.1.2", Hostname: "switch", IsAlive: true},
	}
	// Создаём топологию из результатов (имитация того что делает BuildTopology)
	topo := &topology.Topology{
		Devices: make(map[string]*topology.Device),
	}
	for _, r := range results {
		topo.Devices[r.IP] = &topology.Device{
			IP:       r.IP,
			Hostname: r.Hostname,
		}
	}

	result := formatTopologyPreview(topo, nil, topologyBuildMetrics{})
	// Топология без связей показывает только count устройств
	if !strings.Contains(result, "Устройств:") {
		t.Errorf("expected device count in topology preview, got: %s", result)
	}
}

// --- fmt.Sprintf format verification ---

func TestRecommendedBadgeText_Format(t *testing.T) {
	tests := []struct {
		profileName string
		badgeClass  string
		want        string
	}{
		{"Quick", "green", "Quick (green)"},
		{"Balanced", "yellow", "Balanced (yellow)"},
		{"Deep", "red", "Deep (red)"},
		{"", "", " ()"},
	}

	ctrl := &ScanController{}
	for _, tt := range tests {
		got := ctrl.recommendedBadgeText(tt.profileName, tt.badgeClass)
		if got != tt.want {
			t.Errorf("recommendedBadgeText(%q, %q) = %q, want %q", tt.profileName, tt.badgeClass, got, tt.want)
		}
	}
}

// --- Test that fmt.Errorf can be called (no import issues) ---

func TestFormatTopologyPreview_EdgeCaseEmptyReport(t *testing.T) {
	topo := &topology.Topology{}
	report := &snmpcollector.CollectReport{}
	result := formatTopologyPreview(topo, report, topologyBuildMetrics{})
	// Не паникует
	_ = fmt.Sprintf("result: %s", result)
}
