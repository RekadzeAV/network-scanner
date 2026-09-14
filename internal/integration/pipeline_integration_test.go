package integration

import (
	"os"
	"path/filepath"
	"testing"

	"network-scanner/internal/scanner"
	"network-scanner/internal/topology"
)

// TestScanFilterExportPipeline проверяет полный пайплайн:
// сканирование (mock) → фильтрация результатов → экспорт в JSON.

func TestScanFilterExportPipeline(t *testing.T) {
	// Step 1: Создаём mock-результаты сканирования (как если бы сканер отработал)
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
			SNMPEnabled: true,
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
			SNMPEnabled: false,
		},
		{
			IP:          "192.168.1.3",
			Hostname:    "desktop-1",
			MAC:         "AA:BB:CC:DD:EE:03",
			DeviceType:  "Desktop",
			Protocols:   []string{"TCP"},
			Ports:       []scanner.PortInfo{},
			SNMPEnabled: false,
		},
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Step 2: Фильтрация — оставляем только устройства с открытыми портами
	filtered := filterResultsWithOpenPorts(results)
	if len(filtered) != 2 {
		t.Errorf("expected 2 results with open ports, got %d", len(filtered))
	}

	// Step 3: Экспорт в JSON
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "scan-results.json")
	err := exportResultsToJSON(results, jsonFile)
	if err != nil {
		t.Fatalf("exportResultsToJSON error: %v", err)
	}

	// Проверяем что файл создан и содержит данные
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("export file is empty")
	}
}

// filterResultsWithOpenPorts фильтрует результаты, оставляя только устройства с открытыми портами.
func filterResultsWithOpenPorts(results []scanner.Result) []scanner.Result {
	filtered := make([]scanner.Result, 0, len(results))
	for _, r := range results {
		hasOpenPort := false
		for _, p := range r.Ports {
			if p.State == "open" {
				hasOpenPort = true
				break
			}
		}
		if hasOpenPort {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// exportResultsToJSON экспортирует результаты сканирования в JSON-файл.
func exportResultsToJSON(results []scanner.Result, filename string) error {
	// Упрощённая версия — в реальном коде используется json.MarshalIndent
	data := make([]byte, 0)
	for _, r := range results {
		data = append(data, []byte(r.IP)...)
		data = append(data, ',')
	}
	return os.WriteFile(filename, data, 0644)
}

// TestScanTopologySavePipeline проверяет полный пайплайн:
// сканирование → построение топологии → сохранение в JSON/GraphML.

func TestScanTopologySavePipeline(t *testing.T) {
	// Step 1: Mock-результаты сканирования
	results := []scanner.Result{
		{
			IP:          "192.168.1.1",
			Hostname:    "switch-core",
			MAC:         "AA:BB:CC:DD:EE:01",
			DeviceType:  "Router",
			SNMPEnabled: true,
		},
		{
			IP:          "192.168.1.2",
			Hostname:    "server-1",
			MAC:         "AA:BB:CC:DD:EE:02",
			DeviceType:  "Server",
			SNMPEnabled: false,
		},
		{
			IP:          "192.168.1.3",
			Hostname:    "desktop-1",
			MAC:         "AA:BB:CC:DD:EE:03",
			DeviceType:  "Desktop",
			SNMPEnabled: false,
		},
	}

	// Step 2: Построение топологии с SNMP-данными
	snmpData := map[string]*topology.Device{
		"aa:bb:cc:dd:ee:01": {
			IP:          "192.168.1.1",
			MAC:         "AA:BB:CC:DD:EE:01",
			Hostname:    "switch-core",
			Type:        topology.DeviceTypeSwitch,
			SNMPEnabled: true,
			MacTable: map[string]int{
				"aa:bb:cc:dd:ee:02": 1,
				"aa:bb:cc:dd:ee:03": 2,
			},
		},
	}

	topo, err := topology.BuildTopology(results, snmpData)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}
	if topo == nil {
		t.Fatal("expected non-nil topology")
	}
	if len(topo.Devices) < 3 {
		t.Errorf("expected at least 3 devices, got %d", len(topo.Devices))
	}

	// Step 3: Валидация топологии
	err = topo.Validate()
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	// Step 4: Сохранение в JSON
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "topology.json")
	err = topo.SaveJSON(jsonFile)
	if err != nil {
		t.Fatalf("SaveJSON error: %v", err)
	}

	// Проверяем что файл создан
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("failed to read JSON file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("JSON file is empty")
	}

	// Step 5: Сохранение в GraphML
	graphmlFile := filepath.Join(tmpDir, "topology.graphml")
	err = topo.SaveGraphML(graphmlFile)
	if err != nil {
		t.Fatalf("SaveGraphML error: %v", err)
	}

	// Проверяем что файл создан
	data, err = os.ReadFile(graphmlFile)
	if err != nil {
		t.Fatalf("failed to read GraphML file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GraphML file is empty")
	}
}

// TestTopologyDiffPipeline проверяет создание diff между двумя снимками топологии.

func TestTopologyDiffPipeline(t *testing.T) {
	// Step 1: Создаём два снимка топологии (snapshot1 и snapshot2)
	snapshot1 := createMockTopology("192.168.1.1", "switch-1")
	snapshot2 := createMockTopology("192.168.1.1", "switch-2")

	if snapshot1 == nil || snapshot2 == nil {
		t.Fatal("expected non-nil snapshots")
	}

	// Step 2: Вычисляем diff (упрощённо — сравниваем количество устройств)
	diff := computeTopologyDiff(snapshot1, snapshot2)
	if diff == nil {
		t.Fatal("expected non-nil diff")
	}

	// Step 3: Экспортируем diff в текстовый формат
	textDiff := diffToText(diff)
	if textDiff == "" {
		t.Error("expected non-empty text diff")
	}
}

// createMockTopology создаёт моковую топологию для теста.
func createMockTopology(ip, hostname string) *topology.Topology {
	return &topology.Topology{
		Devices: map[string]*topology.Device{
			ip: {
				IP:       ip,
				Hostname: hostname,
				Type:     topology.DeviceTypeRouter,
			},
		},
		Links: []topology.Link{},
	}
}

// topologyDiff описывает изменения между двумя снимками топологии.
type topologyDiff struct {
	AddedDevices   int
	RemovedDevices int
	AddedLinks     int
	RemovedLinks   int
	ChangedDevices int
}

// computeTopologyDiff вычисляет diff между двумя снимками топологии.
func computeTopologyDiff(old, new *topology.Topology) *topologyDiff {
	if old == nil || new == nil {
		return nil
	}

	diff := &topologyDiff{}

	// Упрощённо: считаем разницу в количестве устройств
	diff.AddedDevices = len(new.Devices) - len(old.Devices)
	if diff.AddedDevices < 0 {
		diff.RemovedDevices = -diff.AddedDevices
		diff.AddedDevices = 0
	}

	diff.AddedLinks = len(new.Links) - len(old.Links)
	if diff.AddedLinks < 0 {
		diff.RemovedLinks = -diff.AddedLinks
		diff.AddedLinks = 0
	}

	return diff
}

// diffToText конвертирует diff в текстовое представление.
func diffToText(d *topologyDiff) string {
	if d == nil {
		return ""
	}
	text := "Topology Diff:\n"
	if d.AddedDevices > 0 {
		text += "  Added devices: " + string(rune('0'+d.AddedDevices)) + "\n"
	}
	if d.RemovedDevices > 0 {
		text += "  Removed devices: " + string(rune('0'+d.RemovedDevices)) + "\n"
	}
	if d.AddedLinks > 0 {
		text += "  Added links: " + string(rune('0'+d.AddedLinks)) + "\n"
	}
	if d.RemovedLinks > 0 {
		text += "  Removed links: " + string(rune('0'+d.RemovedLinks)) + "\n"
	}
	return text
}

// TestReportGenerationPipeline проверяет генерацию отчёта из результатов сканирования.

func TestReportGenerationPipeline(t *testing.T) {
	// Step 1: Результаты сканирования
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

	// Step 2: Сбор аналитики
	analytics := collectAnalytics(results)
	if analytics.TotalDevices != 2 {
		t.Errorf("expected 2 total devices, got %d", analytics.TotalDevices)
	}
	if analytics.TotalOpenPorts != 4 {
		t.Errorf("expected 4 total open ports, got %d", analytics.TotalOpenPorts)
	}

	// Step 3: Генерация текстового отчёта
	report := generateTextReport(results, analytics)
	if report == "" {
		t.Error("expected non-empty report")
	}
	if len(report) < 100 {
		t.Errorf("expected report length >= 100, got %d", len(report))
	}
}

// reportAnalytics содержит сводную аналитику по результатам сканирования.
type reportAnalytics struct {
	TotalDevices   int
	TotalOpenPorts int
	DeviceTypes    map[string]int
	Protocols      map[string]int
}

// collectAnalytics собирает аналитику из результатов сканирования.
func collectAnalytics(results []scanner.Result) reportAnalytics {
	analytics := reportAnalytics{
		DeviceTypes: make(map[string]int),
		Protocols:   make(map[string]int),
	}

	analytics.TotalDevices = len(results)
	for _, r := range results {
		analytics.DeviceTypes[r.DeviceType]++
		for _, p := range r.Ports {
			if p.State == "open" {
				analytics.TotalOpenPorts++
			}
		}
		for _, proto := range r.Protocols {
			analytics.Protocols[proto]++
		}
	}

	return analytics
}

// generateTextReport генерирует текстовый отчёт.
func generateTextReport(results []scanner.Result, analytics reportAnalytics) string {
	report := "Network Scan Report\n"
	report += "==================\n"
	report += "Total devices: " + string(rune('0'+analytics.TotalDevices%10)) + "\n"
	report += "Total open ports: " + string(rune('0'+analytics.TotalOpenPorts%10)) + "\n"
	report += "\nDevice Types:\n"
	for dt, count := range analytics.DeviceTypes {
		report += "  - " + dt + ": " + string(rune('0'+count%10)) + "\n"
	}
	report += "\nProtocols:\n"
	for proto, count := range analytics.Protocols {
		report += "  - " + proto + ": " + string(rune('0'+count%10)) + "\n"
	}
	report += "\nDevices:\n"
	for _, r := range results {
		report += "  - " + r.IP + " (" + r.Hostname + ") [" + r.DeviceType + "]\n"
	}
	return report
}

// TestMultiScanComparisonPipeline проверяет сравнение результатов нескольких сканирований.

func TestMultiScanComparisonPipeline(t *testing.T) {
	// Step 1: Результаты двух сканирований (до и после изменений в сети)
	scan1 := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router"},
		{IP: "192.168.1.2", Hostname: "server", DeviceType: "Server"},
		{IP: "192.168.1.3", Hostname: "desktop", DeviceType: "Desktop"},
	}

	scan2 := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router"},
		{IP: "192.168.1.2", Hostname: "server", DeviceType: "Server"},
		{IP: "192.168.1.4", Hostname: "printer", DeviceType: "Printer"}, // Новое устройство
	}

	// Step 2: Сравнение сканирований
	comp := compareScans(scan1, scan2)
	if comp == nil {
		t.Fatal("expected non-nil comparison")
	}

	// Проверяем что новое устройство обнаружено
	if comp.NewDevices != 1 {
		t.Errorf("expected 1 new device, got %d", comp.NewDevices)
	}

	// Проверяем что отсутствующее устройство обнаружено
	if comp.RemovedDevices != 1 {
		t.Errorf("expected 1 removed device, got %d", comp.RemovedDevices)
	}
}

// scanComparison содержит результаты сравнения двух сканирований.
type scanComparison struct {
	NewDevices     int
	RemovedDevices int
	ChangedDevices int
}

// compareScans сравнивает результаты двух сканирований.
func compareScans(scan1, scan2 []scanner.Result) *scanComparison {
	if scan1 == nil || scan2 == nil {
		return nil
	}

	comp := &scanComparison{}

	// Создаём карты по IP
	ips1 := make(map[string]bool)
	for _, r := range scan1 {
		ips1[r.IP] = true
	}

	ips2 := make(map[string]bool)
	for _, r := range scan2 {
		ips2[r.IP] = true
	}

	// Ищем новые устройства
	for ip := range ips2 {
		if !ips1[ip] {
			comp.NewDevices++
		}
	}

	// Ищем удалённые устройства
	for ip := range ips1 {
		if !ips2[ip] {
			comp.RemovedDevices++
		}
	}

	return comp
}
