// Package benchmark предоставляет бенчмарки для topology и report.
//
// Запуск:
//   go test -bench=. -benchmem ./internal/benchmark/
//
// Сохранение baseline:
//   go test -bench=. -benchmem -run=^$ ./internal/benchmark/ > benchmarks_baseline.txt

package benchmark

import (
	"bytes"
	"strconv"
	"testing"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/report"
	"network-scanner/internal/scanner"
	"network-scanner/internal/topology"
)

// ============================================================================
// Topology Benchmarks
// ============================================================================

// makeTopologyData создаёт тестовые данные для бенчмарков topology.
func makeTopologyData(count int) []scanner.Result {
	results := make([]scanner.Result, count)
	for i := 0; i < count; i++ {
		suffix := strconv.Itoa(i % 1000)
		results[i] = scanner.Result{
			IP:           "192.168.1." + suffix,
			MAC:          "aa:bb:cc:dd:ee:" + suffix,
			Hostname:     "host-" + suffix,
			DeviceType:   "host",
			DeviceVendor: "TestVendor",
			Protocols:    []string{"tcp", "udp"},
			GuessOS:      "Linux",
			IsAlive:      true,
		}
		if i%10 == 0 {
			results[i].Ports = []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", State: "open", Service: "http"},
				{Port: 443, Protocol: "tcp", State: "open", Service: "https"},
			}
		}
	}
	return results
}

// makeContractsResults создаёт contracts.ScanResult для report-бенчмарков.
func makeContractsResults(count int) []contracts.ScanResult {
	results := make([]contracts.ScanResult, count)
	for i := 0; i < count; i++ {
		suffix := strconv.Itoa(i % 1000)
		results[i] = contracts.ScanResult{
			IP:           "192.168.1." + suffix,
			Hostname:     "host-" + suffix,
			MAC:          "aa:bb:cc:dd:ee:" + suffix,
			DeviceType:   "host",
			DeviceVendor: "TestVendor",
			GuessOS:      "Linux",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		}
	}
	return results
}

// BenchmarkBuildTopology_10 — BuildTopology для 10 хостов
func BenchmarkBuildTopology_10(b *testing.B) {
	results := makeTopologyData(10)
	snmpData := map[string]*topology.Device{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = topology.BuildTopology(results, snmpData)
	}
}

// BenchmarkBuildTopology_100 — BuildTopology для 100 хостов
func BenchmarkBuildTopology_100(b *testing.B) {
	results := makeTopologyData(100)
	snmpData := map[string]*topology.Device{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = topology.BuildTopology(results, snmpData)
	}
}

// BenchmarkBuildTopology_1000 — BuildTopology для 1000 хостов
func BenchmarkBuildTopology_1000(b *testing.B) {
	results := makeTopologyData(1000)
	snmpData := map[string]*topology.Device{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = topology.BuildTopology(results, snmpData)
	}
}

// BenchmarkBuildTopology_WithSNMP — BuildTopology с SNMP-данными
func BenchmarkBuildTopology_WithSNMP(b *testing.B) {
	results := makeTopologyData(100)
	snmpData := make(map[string]*topology.Device, 20)
	for i := 0; i < 20; i++ {
		suffix := strconv.Itoa(i)
		snmpData["192.168.1."+suffix] = &topology.Device{
			IP:          "192.168.1." + suffix,
			Type:        topology.DeviceTypeSwitch,
			SNMPEnabled: true,
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = topology.BuildTopology(results, snmpData)
	}
}

// BenchmarkTopology_ToJSON — сериализация топологии в JSON
func BenchmarkTopology_ToJSON(b *testing.B) {
	results := makeTopologyData(50)
	topo, _ := topology.BuildTopology(results, map[string]*topology.Device{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = topo.ToJSONSchema()
	}
}

// BenchmarkTopology_SaveGraphML — сохранение топологии в GraphML
func BenchmarkTopology_SaveGraphML(b *testing.B) {
	results := makeTopologyData(50)
	topo, _ := topology.BuildTopology(results, map[string]*topology.Device{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = topo.SaveGraphMLToBytes()
	}
}

// BenchmarkTopology_ToText — текстовый вывод топологии
func BenchmarkTopology_ToText(b *testing.B) {
	results := makeTopologyData(50)
	topo, _ := topology.BuildTopology(results, map[string]*topology.Device{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		_ = topo.WriteText(buf)
	}
}

// BenchmarkTopology_ToDOT — генерация DOT-графа
func BenchmarkTopology_ToDOT(b *testing.B) {
	results := makeTopologyData(50)
	topo, _ := topology.BuildTopology(results, map[string]*topology.Device{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		_ = topo.ToDOT(buf)
	}
}

// BenchmarkTopology_DedupReport — dedup-отчёт топологии
func BenchmarkTopology_DedupReport(b *testing.B) {
	results := makeTopologyData(50)
	topo, _ := topology.BuildTopology(results, map[string]*topology.Device{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = topo.DedupReport()
	}
}

// ============================================================================
// Report Benchmarks
// ============================================================================

// makeReportData создаёт тестовые данные для бенчмарков report.
func makeReportData(count int) *report.ScanReportData {
	rows := make([]report.ScanResultRow, count)
	for i := 0; i < count; i++ {
		suffix := strconv.Itoa(i % 1000)
		rows[i] = report.ScanResultRow{
			IP:       "192.168.1." + suffix,
			Hostname: "host-" + suffix,
			Ports:    2 + i%5,
			OS:       "Linux",
			Vendor:   "TestVendor",
		}
	}
	return &report.ScanReportData{
		GeneratedAt: "2026-10-25T12:00:00Z",
		ScanID:      "test-scan-001",
		Network:     "192.168.1.0/24",
		HostCount:   count,
		Results:     rows,
		Topology: &report.TopologySummary{
			DeviceCount: count / 2,
			LinkCount:   count,
		},
	}
}

// BenchmarkGenerateScanReportData_10 — GenerateScanReportData для 10 хостов
func BenchmarkGenerateScanReportData_10(b *testing.B) {
	results := makeContractsResults(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = report.GenerateScanReportData("scan-001", "192.168.1.0/24", results, nil, nil)
	}
}

// BenchmarkGenerateScanReportData_100 — GenerateScanReportData для 100 хостов
func BenchmarkGenerateScanReportData_100(b *testing.B) {
	results := makeContractsResults(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = report.GenerateScanReportData("scan-001", "192.168.1.0/24", results, nil, nil)
	}
}

// BenchmarkGenerateScanReportData_1000 — GenerateScanReportData для 1000 хостов
func BenchmarkGenerateScanReportData_1000(b *testing.B) {
	results := makeContractsResults(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = report.GenerateScanReportData("scan-001", "192.168.1.0/24", results, nil, nil)
	}
}

// BenchmarkRenderScanHTML_10 — рендер HTML-отчёта для 10 хостов
func BenchmarkRenderScanHTML_10(b *testing.B) {
	data := makeReportData(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = report.RenderScanHTML(data)
	}
}

// BenchmarkRenderScanHTML_100 — рендер HTML-отчёта для 100 хостов
func BenchmarkRenderScanHTML_100(b *testing.B) {
	data := makeReportData(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = report.RenderScanHTML(data)
	}
}

// BenchmarkRenderScanHTML_1000 — рендер HTML-отчёта для 1000 хостов
func BenchmarkRenderScanHTML_1000(b *testing.B) {
	data := makeReportData(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = report.RenderScanHTML(data)
	}
}

// BenchmarkRenderScanHTML_MemoryAlloc — аллокации памяти при рендере HTML
func BenchmarkRenderScanHTML_MemoryAlloc(b *testing.B) {
	data := makeReportData(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = report.RenderScanHTML(data)
	}
}

// BenchmarkRenderSecurityHTML_10 — рендер Security HTML (10 хостов)
func BenchmarkRenderSecurityHTML_10(b *testing.B) {
	results := makeTopologyData(10)
	b.ResetTimer()
	now := time.Now()
	for i := 0; i < b.N; i++ {
		_, _ = report.RenderSecurityHTML(results, nil, now)
	}
}

// BenchmarkRenderSecurityHTML_100 — рендер Security HTML (100 хостов)
func BenchmarkRenderSecurityHTML_100(b *testing.B) {
	results := makeTopologyData(100)
	b.ResetTimer()
	now := time.Now()
	for i := 0; i < b.N; i++ {
		_, _ = report.RenderSecurityHTML(results, nil, now)
	}
}

// ============================================================================
// Memory Benchmarks (Topology + Report)
// ============================================================================

// BenchmarkTopology_MemoryAlloc — аллокации памяти topology
func BenchmarkTopology_MemoryAlloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		results := makeTopologyData(100)
		topo, _ := topology.BuildTopology(results, map[string]*topology.Device{})
		_, _ = topo.ToJSONSchema()
	}
}

// BenchmarkReport_MemoryAlloc — аллокации памяти report
func BenchmarkReport_MemoryAlloc(b *testing.B) {
	b.ReportAllocs()
	data := makeReportData(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = report.RenderScanHTML(data)
	}
}
