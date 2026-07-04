package topology

import (
	"fmt"
	"testing"

	"network-scanner/internal/scanner"
)

// BenchmarkBuildTopology — построение топологии из результатов сканирования
func BenchmarkBuildTopology(b *testing.B) {
	results := generateTestResults(50)
	snmpData := generateTestSNMPData(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildTopology(results, snmpData)
	}
}

// BenchmarkBuildTopologyLarge — построение большой топологии
func BenchmarkBuildTopologyLarge(b *testing.B) {
	results := generateTestResults(200)
	snmpData := generateTestSNMPData(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildTopology(results, snmpData)
	}
}

// BenchmarkBuildTopologyWithOptions — построение с опциями
func BenchmarkBuildTopologyWithOptions(b *testing.B) {
	results := generateTestResults(50)
	snmpData := generateTestSNMPData(10)
	opts := BuildOptions{
		PartialSNMPKeys: map[string]struct{}{
			"ip:192.168.1.1": {},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildTopologyWithOptions(results, snmpData, opts)
	}
}

// BenchmarkTopologyValidate — валидация топологии
func BenchmarkTopologyValidate(b *testing.B) {
	results := generateTestResults(50)
	snmpData := generateTestSNMPData(10)
	topo, _ := BuildTopology(results, snmpData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = topo.Validate()
	}
}

// BenchmarkTopologySaveJSON — сохранение в JSON
func BenchmarkTopologySaveJSON(b *testing.B) {
	results := generateTestResults(50)
	snmpData := generateTestSNMPData(10)
	topo, _ := BuildTopology(results, snmpData)
	filename := "/tmp/benchmark-topology.json"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = topo.SaveJSON(filename)
	}
	// Cleanup
	for i := 0; i < b.N; i++ {
		_ = topo.SaveJSON(filename)
	}
}

// BenchmarkTopologyToDOT — генерация DOT графа
func BenchmarkTopologyToDOT(b *testing.B) {
	results := generateTestResults(50)
	snmpData := generateTestSNMPData(10)
	topo, _ := BuildTopology(results, snmpData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = topo.ToDOT(nil)
	}
}

// BenchmarkNormalizeMAC — нормализация MAC адреса
func BenchmarkNormalizeMAC(b *testing.B) {
	macs := []string{
		"AA:BB:CC:DD:EE:FF",
		"aa-bb-cc-dd-ee-ff",
		"aa:bb:cc:dd:ee:ff",
		"AABB.CCDD.EEFF",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, mac := range macs {
			_ = normalizeMAC(mac)
		}
	}
}

// BenchmarkNormalizedKey — генерация ключа устройства
func BenchmarkNormalizedKey(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalizedKey("aa:bb:cc:dd:ee:ff", "192.168.1.1")
	}
}

// BenchmarkIsBroadcastOrMulticast — проверка broadcast/multicast MAC
func BenchmarkIsBroadcastOrMulticast(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isBroadcastOrMulticast("ff:ff:ff:ff:ff:ff")
		_ = isBroadcastOrMulticast("01:00:5e:00:00:01")
	}
}

// BenchmarkIsZeroMAC — проверка zero MAC
func BenchmarkIsZeroMAC(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isZeroMAC("00:00:00:00:00:00")
	}
}

// BenchmarkNodeID — генерация ID узла
func BenchmarkNodeID(b *testing.B) {
	dev := &Device{
		IP:       "192.168.1.1",
		MAC:      "aa:bb:cc:dd:ee:ff",
		Hostname: "router",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nodeID(dev)
	}
}

// BenchmarkDeviceDisplayName — получение displayName устройства
func BenchmarkDeviceDisplayName(b *testing.B) {
	dev := &Device{
		IP:       "192.168.1.1",
		MAC:      "aa:bb:cc:dd:ee:ff",
		Hostname: "router",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deviceDisplayName(dev)
	}
}

// BenchmarkPortLabel — получение label порта
func BenchmarkPortLabel(b *testing.B) {
	port := &Port{Index: 1, Name: "eth0"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = portLabel(port)
	}
}

// BenchmarkLinkKey — генерация ключа связи
func BenchmarkLinkKey(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = linkKey("mac_aa_bb_cc_dd_ee_ff", "eth0", "ip_192_168_1_2", "eth1")
	}
}

// BenchmarkClassifyFromScannerResult — классификация устройства
func BenchmarkClassifyFromScannerResult(b *testing.B) {
	types := []string{
		"Router/Network Device",
		"Web Server",
		"Windows Computer",
		"Unknown",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, t := range types {
			_ = classifyFromScannerResult(t)
		}
	}
}

// BenchmarkConfidenceRank — получение ранга уверенности
func BenchmarkConfidenceRank(b *testing.B) {
	confidences := []LinkConfidence{
		LinkConfidenceHigh,
		LinkConfidenceMedium,
		LinkConfidenceLow,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range confidences {
			_ = confidenceRank(c)
		}
	}
}

// BenchmarkEnsurePort — добавление порта
func BenchmarkEnsurePort(b *testing.B) {
	dev := &Device{Ports: []Port{}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ensurePort(dev, i, "")
	}
}

// BenchmarkFindNeighbor — поиск соседа
func BenchmarkFindNeighbor(b *testing.B) {
	byMAC := map[string]*Device{
		"aa:bb:cc:dd:ee:01": {MAC: "aa:bb:cc:dd:ee:01"},
		"aa:bb:cc:dd:ee:02": {MAC: "aa:bb:cc:dd:ee:02"},
	}
	byHostname := map[string]*Device{
		"router": {Hostname: "router"},
	}
	neighbor := &LldpNeighbor{
		RemoteChassisID: "aa:bb:cc:dd:ee:01",
		RemoteSysName:   "router",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findNeighbor(byMAC, byHostname, neighbor)
	}
}

// BenchmarkMaybeLowerConfidence — снижение уверенности
func BenchmarkMaybeLowerConfidence(b *testing.B) {
	opts := BuildOptions{
		PartialSNMPKeys: map[string]struct{}{
			"ip:192.168.1.1": {},
		},
	}
	dev1 := &Device{IP: "192.168.1.1"}
	dev2 := &Device{IP: "192.168.1.2"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = maybeLowerConfidence(LinkConfidenceHigh, dev1, dev2, opts)
	}
}

// BenchmarkIsPartialDevice — проверка частичного устройства
func BenchmarkIsPartialDevice(b *testing.B) {
	opts := BuildOptions{
		PartialSNMPKeys: map[string]struct{}{
			"ip:192.168.1.1": {},
		},
	}
	dev := &Device{IP: "192.168.1.1"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isPartialDevice(dev, opts)
	}
}

// BenchmarkDeviceKeys — получение ключей устройства
func BenchmarkDeviceKeys(b *testing.B) {
	dev := &Device{
		IP:       "192.168.1.1",
		MAC:      "aa:bb:cc:dd:ee:ff",
		Hostname: "router",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deviceKeys(dev)
	}
}

// BenchmarkParseMACBytes — парсинг MAC в байты
func BenchmarkParseMACBytes(b *testing.B) {
	mac := "aa:bb:cc:dd:ee:ff"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseMACBytes(mac)
	}
}

// BenchmarkAddLink — добавление связи
func BenchmarkAddLink(b *testing.B) {
	src := &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01"}
	dst := &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02"}
	dedup := make(map[string]int)
	byEndpoint := make(map[string]int)
	t := &Topology{
		Devices: map[string]*Device{
			"aa:bb:cc:dd:ee:01": src,
			"aa:bb:cc:dd:ee:02": dst,
		},
		Links: []Link{},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addLink(
			dedup, byEndpoint, t,
			src, 1, "eth0",
			dst, 2, "eth1",
			LinkSourceLLDP,
			LinkConfidenceHigh,
			"test_evidence",
		)
	}
}

// BenchmarkBuildTopologyEmpty — построение пустой топологии
func BenchmarkBuildTopologyEmpty(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildTopology(nil, nil)
	}
}

// BenchmarkBuildTopologyNoSNMP — построение без SNMP данных
func BenchmarkBuildTopologyNoSNMP(b *testing.B) {
	results := generateTestResults(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildTopology(results, nil)
	}
}

// BenchmarkBuildTopologyNoResults — построение без результатов сканирования
func BenchmarkBuildTopologyNoResults(b *testing.B) {
	snmpData := generateTestSNMPData(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildTopology(nil, snmpData)
	}
}

// BenchmarkTopologySaveGraphML — сохранение в GraphML
func BenchmarkTopologySaveGraphML(b *testing.B) {
	results := generateTestResults(50)
	snmpData := generateTestSNMPData(10)
	topo, _ := BuildTopology(results, snmpData)
	filename := "/tmp/benchmark-topology.graphml"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = topo.SaveGraphML(filename)
	}
}

// Helper functions
func generateTestResults(count int) []scanner.Result {
	results := make([]scanner.Result, 0, count)
	for i := 0; i < count; i++ {
		results = append(results, scanner.Result{
			IP:          fmt.Sprintf("192.168.1.%d", i+1),
			MAC:         fmt.Sprintf("aa:bb:cc:dd:ee:%02x", i),
			Hostname:    fmt.Sprintf("host-%d", i+1),
			DeviceType:  "Computer",
			SNMPEnabled: i%5 == 0,
		})
	}
	return results
}

func generateTestSNMPData(count int) map[string]*Device {
	data := make(map[string]*Device, count)
	for i := 0; i < count; i++ {
		data[fmt.Sprintf("aa:bb:cc:dd:ee:%02x", i)] = &Device{
			IP:          fmt.Sprintf("192.168.1.%d", i+1),
			MAC:         fmt.Sprintf("aa:bb:cc:dd:ee:%02x", i),
			Hostname:    fmt.Sprintf("host-%d", i+1),
			Type:        DeviceTypeHost,
			SNMPEnabled: true,
		}
	}
	return data
}
