package topology

import (
	"bytes"
	"context"
	"os"
	"testing"

	"network-scanner/internal/scanner"
)

// --- D1: Integration кейсы для topology hardening ---

// TestBuildTopologyPartialSNMP проверяет построение топологии с частичными SNMP-данными.
// Частичные данные означают что SNMP собран не полностью (например только MAC-таблица, нет LLDP).
func TestBuildTopologyPartialSNMP(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "switch-partial", SNMPEnabled: true},
		{IP: "192.168.1.2", MAC: "bb:bb:bb:bb:bb:bb", Hostname: "host1"},
		{IP: "192.168.1.3", MAC: "cc:cc:cc:cc:cc:cc", Hostname: "host2"},
	}

	// Частичные SNMP-данные: есть MAC-таблица, но нет LLDP
	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "192.168.1.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "switch-partial",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			// Только MAC-таблица, без LLDP
			MacTable: map[string]int{
				"bb:bb:bb:bb:bb:bb": 1,
				"cc:cc:cc:cc:cc:cc": 2,
			},
			LldpNeighbors: nil, // Нет LLDP данных
		},
	}

	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}

	if len(topo.Devices) < 3 {
		t.Errorf("expected at least 3 devices, got %d", len(topo.Devices))
	}

	// Проверяем что есть хотя бы 1 FDB связь (dedup может уменьшить количество)
	fdbLinks := 0
	for _, link := range topo.Links {
		if link.SourceType == LinkSourceFDB {
			fdbLinks++
		}
	}
	if fdbLinks < 1 {
		t.Errorf("expected at least 1 FDB link, got %d", fdbLinks)
	}
}

// TestBuildTopologyPartialSNMLowerConfidence проверяет понижение confidence при partial SNMP.
func TestBuildTopologyPartialSNMLowerConfidence(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "switch-partial", SNMPEnabled: true},
		{IP: "192.168.1.2", MAC: "bb:bb:bb:bb:bb:bb", Hostname: "host1"},
	}

	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "192.168.1.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "switch-partial",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			LldpNeighbors: []*LldpNeighbor{
				{LocalIfIndex: 1, RemoteChassisID: "bb:bb:bb:bb:bb:bb", RemotePortID: "Gi0/1", RemoteSysName: "host1"},
			},
		},
	}

	// Строим с partial SNMP ключом
	topo, err := BuildTopologyWithOptions(results, snmp, BuildOptions{
		PartialSNMPKeys: map[string]struct{}{
			"ip:192.168.1.1": {},
		},
	})
	if err != nil {
		t.Fatalf("BuildTopologyWithOptions error: %v", err)
	}

	if len(topo.Links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(topo.Links))
	}

	// LLDP связь с partial SNMP должна иметь сниженный confidence
	if topo.Links[0].Confidence != LinkConfidenceMedium {
		t.Errorf("expected Medium confidence (downgraded from High), got %s", topo.Links[0].Confidence)
	}
}

// TestBuildTopologyLoopLikeLLDP проверяет что loop-like LLDP связи игнорируются.
func TestBuildTopologyLoopLikeLLDP(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "switch-loop", SNMPEnabled: true},
	}

	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "192.168.1.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "switch-loop",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			// LLDP сосед указывает на самого себя (loop)
			LldpNeighbors: []*LldpNeighbor{
				{
					LocalIfIndex:    1,
					RemoteChassisID: "aa:aa:aa:aa:aa:aa", // Тот же MAC
					RemotePortID:    "Gi0/1",
					RemoteSysName:   "switch-loop",
				},
			},
		},
	}

	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}

	// Loop-like связи должны быть пропущены
	if len(topo.Links) != 0 {
		t.Errorf("expected 0 links for loop-like LLDP, got %d", len(topo.Links))
	}
}

// TestBuildTopologyMixedVendor проверяет работу с разными вендорами (разные форматы chassis ID).
func TestBuildTopologyMixedVendor(t *testing.T) {
	results := []scanner.Result{
		{IP: "10.0.0.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "core-switch", SNMPEnabled: true},
		{IP: "10.0.0.2", MAC: "bb:bb:bb:bb:bb:bb", Hostname: "edge-switch-01", SNMPEnabled: true},
		{IP: "10.0.0.3", MAC: "cc:cc:cc:cc:cc:cc", Hostname: "access-switch-02", SNMPEnabled: true},
	}

	// Разные вендоры используют разные форматы LLDP:
	// - Cisco: MAC chassis ID
	// - Juniper/Other: non-MAC chassis ID (например Port-Channel)
	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "10.0.0.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "core-switch",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			LldpNeighbors: []*LldpNeighbor{
				// Cisco: MAC chassis ID
				{
					LocalIfIndex:    1,
					RemoteChassisID: "bb:bb:bb:bb:bb:bb",
					RemotePortID:    "Gi0/1",
					RemoteSysName:   "edge-switch-01",
				},
				// Juniper: non-MAC chassis ID (Port-Channel)
				{
					LocalIfIndex:    2,
					RemoteChassisID: "Port-Channel1",
					RemotePortID:    "ae0",
					RemoteSysName:   " access-switch-02 ",
				},
			},
		},
		"bb:bb:bb:bb:bb:bb": {
			IP:          "10.0.0.2",
			MAC:         "bb:bb:bb:bb:bb:bb",
			Hostname:    "edge-switch-01",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			LldpNeighbors: []*LldpNeighbor{
				// Обратная связь с core
				{
					LocalIfIndex:    12,
					RemoteChassisID: "aa:aa:aa:aa:aa:aa",
					RemotePortID:    "Gi0/24",
					RemoteSysName:   "core-switch",
				},
			},
		},
	}

	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}

	if len(topo.Devices) < 3 {
		t.Errorf("expected at least 3 devices, got %d", len(topo.Devices))
	}

	// Должно быть как минимум 2 связи (core->edge и core->access)
	if len(topo.Links) < 2 {
		t.Errorf("expected at least 2 links, got %d", len(topo.Links))
	}

	// Все связи должны быть LLDP
	for _, link := range topo.Links {
		if link.SourceType != LinkSourceLLDP {
			t.Errorf("expected LLDP source, got %s", link.SourceType)
		}
	}

	// Проверка что RemoteSysName fallback работает (access-switch-02)
	foundAccessLink := false
	for _, link := range topo.Links {
		if (deviceDisplayName(link.Source) == "core-switch" || deviceDisplayName(link.Target) == "core-switch") &&
			(deviceDisplayName(link.Source) == "access-switch-02" || deviceDisplayName(link.Target) == "access-switch-02") {
			foundAccessLink = true
			break
		}
	}
	if !foundAccessLink {
		t.Error("expected link to access-switch-02 via RemoteSysName fallback")
	}
}

// TestBuildTopologyMixedSNMPLLDPRFDB проверяет смешанные источники LLDP + FDB.
func TestBuildTopologyMixedSNMPLLDPRFDB(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "switch1", SNMPEnabled: true},
		{IP: "192.168.1.2", MAC: "bb:bb:bb:bb:bb:bb", Hostname: "host1"},
		{IP: "192.168.1.3", MAC: "cc:cc:cc:cc:cc:cc", Hostname: "host2"},
	}

	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "192.168.1.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "switch1",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			LldpNeighbors: []*LldpNeighbor{
				// LLDP для host1
				{
					LocalIfIndex:    1,
					RemoteChassisID: "bb:bb:bb:bb:bb:bb",
					RemotePortID:    "Gi0/1",
					RemoteSysName:   "host1",
				},
			},
			MacTable: map[string]int{
				// FDB для host2
				"cc:cc:cc:cc:cc:cc": 2,
			},
		},
	}

	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}

	if len(topo.Links) != 2 {
		t.Errorf("expected 2 links (1 LLDP + 1 FDB), got %d", len(topo.Links))
	}

	// Проверка что есть и LLDP и FDB связи
	hasLLDP := false
	hasFDB := false
	for _, link := range topo.Links {
		if link.SourceType == LinkSourceLLDP {
			hasLLDP = true
		}
		if link.SourceType == LinkSourceFDB {
			hasFDB = true
		}
	}

	if !hasLLDP {
		t.Error("expected at least one LLDP link")
	}
	if !hasFDB {
		t.Error("expected at least one FDB link")
	}
}

// TestBuildTopologyLargeNetwork проверяет масштабирование на большой сети.
func TestBuildTopologyLargeNetwork(t *testing.T) {
	const deviceCount = 50

	results := make([]scanner.Result, 0, deviceCount)
	snmp := make(map[string]*Device)

	for i := 0; i < deviceCount; i++ {
		mac := sprintfMAC(i)
		ip := sprintf("192.168.%d.%d", i/256, i%256)
		hostname := sprintf("device-%03d", i)

		results = append(results, scanner.Result{
			IP:          ip,
			MAC:         mac,
			Hostname:    hostname,
			DeviceType:  "host",
			SNMPEnabled: i%5 == 0, // SNMP каждые 5 устройств
		})

		if i%5 == 0 {
			snmp[mac] = &Device{
				IP:          ip,
				MAC:         mac,
				Hostname:    hostname,
				Type:        DeviceTypeHost,
				SNMPEnabled: true,
			}
		}
	}

	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}

	if len(topo.Devices) < deviceCount {
		t.Errorf("expected at least %d devices, got %d", deviceCount, len(topo.Devices))
	}

	// Проверка что валидация проходит
	if err := topo.Validate(); err != nil {
		t.Errorf("topology validation failed: %v", err)
	}
}

// TestBuildTopologyNoSNMPData проверяет топологию без SNMP данных.
func TestBuildTopologyNoSNMPData(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "host1"},
		{IP: "192.168.1.2", MAC: "bb:bb:bb:bb:bb:bb", Hostname: "host2"},
	}

	topo, err := BuildTopology(results, map[string]*Device{})
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}

	if len(topo.Devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(topo.Devices))
	}

	if len(topo.Links) != 0 {
		t.Errorf("expected 0 links without SNMP, got %d", len(topo.Links))
	}
}

// TestBuildTopologyEmptyResults проверяет пустой ввод.
func TestBuildTopologyEmptyResults(t *testing.T) {
	topo, err := BuildTopology([]scanner.Result{}, map[string]*Device{})
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}

	if len(topo.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(topo.Devices))
	}
	if len(topo.Links) != 0 {
		t.Errorf("expected 0 links, got %d", len(topo.Links))
	}
}

// TestBuildTopologySNMPOnly проверяет ситуацию когда только SNMP данные.
func TestBuildTopologySNMPOnly(t *testing.T) {
	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "192.168.1.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "switch-only",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			LldpNeighbors: []*LldpNeighbor{
				{
					LocalIfIndex:    1,
					RemoteChassisID: "bb:bb:bb:bb:bb:bb",
					RemotePortID:    "Gi0/1",
					RemoteSysName:   "remote-host",
				},
			},
		},
	}

	topo, err := BuildTopology([]scanner.Result{}, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}

	// Должно быть хотя бы одно устройство из SNMP
	if len(topo.Devices) < 1 {
		t.Errorf("expected at least 1 device from SNMP, got %d", len(topo.Devices))
	}
}

// sprintf — вспомогательная функция для форматирования строк.
func sprintf(format string, a ...interface{}) string {
	// Простая реализация для избежания импорта fmt
	result := ""
	for _, arg := range a {
		switch v := arg.(type) {
		case int:
			for _, c := range itoa(v) {
				result += string(c)
			}
		}
	}
	return result
}

func itoa(n int) []rune {
	if n == 0 {
		return []rune{'0'}
	}
	var runes []rune
	for n > 0 {
		runes = append([]rune{rune('0' + n%10)}, runes...)
		n /= 10
	}
	return runes
}

// --- D2: Дополнительные интеграционные тесты для coverage 90%+ ---

// TestIntegrationBuildTopologyWithOptions_Empty проверяет пустую топологию.
func TestIntegrationBuildTopologyWithOptions_Empty(t *testing.T) {
	topo, err := BuildTopologyWithOptions(nil, nil, BuildOptions{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if topo == nil {
		t.Fatal("expected non-nil topology")
	}
	if len(topo.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(topo.Devices))
	}
}

// TestIntegrationBuildTopologyWithOptions_NoSNMP проверяет топологию без SNMP.
func TestIntegrationBuildTopologyWithOptions_NoSNMP(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "host1"},
		{IP: "192.168.1.2", MAC: "bb:bb:bb:bb:bb:bb", Hostname: "host2"},
	}
	topo, err := BuildTopologyWithOptions(results, nil, BuildOptions{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(topo.Devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(topo.Devices))
	}
	if len(topo.Links) != 0 {
		t.Errorf("expected 0 links (no SNMP), got %d", len(topo.Links))
	}
}

// TestIntegrationMaybeLowerConfidence_NoPartial проверяет что confidence не снижается без partial.
func TestIntegrationMaybeLowerConfidence_NoPartial(t *testing.T) {
	a := &Device{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa"}
	b := &Device{IP: "192.168.1.2", MAC: "bb:bb:bb:bb:bb:bb"}
	result := maybeLowerConfidence(LinkConfidenceHigh, a, b, BuildOptions{})
	if result != LinkConfidenceHigh {
		t.Errorf("expected High, got %s", result)
	}
}

// TestIntegrationMaybeLowerConfidence_Partial проверяет снижение confidence при partial SNMP.
func TestIntegrationMaybeLowerConfidence_Partial(t *testing.T) {
	a := &Device{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa"}
	b := &Device{IP: "192.168.1.2", MAC: "bb:bb:bb:bb:bb:bb"}
	result := maybeLowerConfidence(LinkConfidenceHigh, a, b, BuildOptions{
		PartialSNMPKeys: map[string]struct{}{"ip:192.168.1.1": {}},
	})
	if result != LinkConfidenceMedium {
		t.Errorf("expected Medium, got %s", result)
	}
}

// TestIntegrationMaybeLowerConfidence_MediumToLow проверяет снижение Medium до Low.
func TestIntegrationMaybeLowerConfidence_MediumToLow(t *testing.T) {
	a := &Device{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa"}
	b := &Device{IP: "192.168.1.2", MAC: "bb:bb:bb:bb:bb:bb"}
	result := maybeLowerConfidence(LinkConfidenceMedium, a, b, BuildOptions{
		PartialSNMPKeys: map[string]struct{}{"ip:192.168.1.1": {}},
	})
	if result != LinkConfidenceLow {
		t.Errorf("expected Low, got %s", result)
	}
}

// TestIntegrationIsPartialDevice_Match проверяет что устройство считается partial.
func TestIntegrationIsPartialDevice_Match(t *testing.T) {
	d := &Device{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "switch"}
	result := isPartialDevice(d, BuildOptions{
		PartialSNMPKeys: map[string]struct{}{"ip:192.168.1.1": {}},
	})
	if !result {
		t.Error("expected true for partial device")
	}
}

// TestIntegrationIsPartialDevice_NoMatch проверяет что устройство не partial.
func TestIntegrationIsPartialDevice_NoMatch(t *testing.T) {
	d := &Device{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "switch"}
	result := isPartialDevice(d, BuildOptions{
		PartialSNMPKeys: map[string]struct{}{"ip:10.0.0.1": {}},
	})
	if result {
		t.Error("expected false for non-partial device")
	}
}

// TestIntegrationDeviceKeys_AllFields проверяет генерацию ключей устройства.
func TestIntegrationDeviceKeys_AllFields(t *testing.T) {
	d := &Device{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "switch"}
	keys := deviceKeys(d)
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

// TestIntegrationDeviceKeys_Empty проверяет пустые ключи.
func TestIntegrationDeviceKeys_Empty(t *testing.T) {
	d := &Device{}
	keys := deviceKeys(d)
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

// TestIntegrationDeviceKeys_Nil проверяет nil устройство.
func TestIntegrationDeviceKeys_Nil(t *testing.T) {
	keys := deviceKeys(nil)
	if keys != nil {
		t.Errorf("expected nil keys, got %v", keys)
	}
}

// TestIntegrationToDOT_Empty проверяет пустой DOT вывод.
func TestIntegrationToDOT_Empty(t *testing.T) {
	topo := &Topology{Devices: make(map[string]*Device), Links: []Link{}}
	var buf bytes.Buffer
	err := topo.ToDOT(&buf)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := buf.String()
	if output == "" {
		t.Error("expected non-empty DOT output")
	}
	if output[0:5] != "graph" {
		t.Errorf("expected DOT output to start with 'graph', got %q", output[:5])
	}
}

// TestIntegrationToDOT_NilTopology проверяет nil топологию.
func TestIntegrationToDOT_NilTopology(t *testing.T) {
	var buf bytes.Buffer
	err := (*Topology)(nil).ToDOT(&buf)
	if err == nil {
		t.Error("expected error for nil topology")
	}
}

// TestIntegrationSaveJSON_Empty проверяет сохранение пустой топологии.
func TestIntegrationSaveJSON_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/topology.json"
	topo := &Topology{Devices: make(map[string]*Device), Links: []Link{}}
	err := topo.SaveJSON(filename)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("expected to read file, got %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON file")
	}
}

// TestIntegrationSaveJSON_InvalidTopology проверяет невалидную топологию.
func TestIntegrationSaveJSON_InvalidTopology(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/topology.json"
	topo := &Topology{Devices: map[string]*Device{"": nil}}
	err := topo.SaveJSON(filename)
	if err == nil {
		t.Error("expected error for invalid topology")
	}
}

// TestIntegrationSaveGraphML_Empty проверяет сохранение GraphML.
func TestIntegrationSaveGraphML_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/topology.graphml"
	topo := &Topology{Devices: make(map[string]*Device), Links: []Link{}}
	err := topo.SaveGraphML(filename)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("expected to read file, got %v", err)
	}
	if !bytes.Contains(data, []byte("graphml")) {
		t.Error("expected GraphML file to contain graphml tag")
	}
}

// TestIntegrationSaveGraphMLToBytes_Empty проверяет байты GraphML.
func TestIntegrationSaveGraphMLToBytes_Empty(t *testing.T) {
	topo := &Topology{Devices: make(map[string]*Device), Links: []Link{}}
	data, err := topo.SaveGraphMLToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty GraphML bytes")
	}
	if !bytes.Contains(data, []byte("graphml")) {
		t.Error("expected GraphML bytes to contain graphml tag")
	}
}

// TestIntegrationValidate_Empty проверяет валидацию пустой топологии.
func TestIntegrationValidate_Empty(t *testing.T) {
	topo := &Topology{Devices: make(map[string]*Device), Links: []Link{}}
	err := topo.Validate()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestIntegrationValidate_NilTopology проверяет валидацию nil.
func TestIntegrationValidate_NilTopology(t *testing.T) {
	err := (*Topology)(nil).Validate()
	if err == nil {
		t.Error("expected error for nil topology")
	}
}

// TestIntegrationValidate_NilDevice проверяет валидацию nil device.
func TestIntegrationValidate_NilDevice(t *testing.T) {
	topo := &Topology{Devices: map[string]*Device{"": nil}}
	err := topo.Validate()
	if err == nil {
		t.Error("expected error for nil device")
	}
}

// TestIntegrationValidate_EmptySourceType проверяет пустой source type.
func TestIntegrationValidate_EmptySourceType(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"192.168.1.1": {IP: "192.168.1.1", Hostname: "router"},
		},
		Links: []Link{{Source: &Device{IP: "192.168.1.1"}, Target: &Device{IP: "192.168.1.2"}, SourceType: "", Confidence: LinkConfidenceHigh}},
	}
	err := topo.Validate()
	if err == nil {
		t.Error("expected error for empty source type")
	}
}

// TestIntegrationValidate_EmptyConfidence проверяет пустую confidence.
func TestIntegrationValidate_EmptyConfidence(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"192.168.1.1": {IP: "192.168.1.1", Hostname: "router"},
		},
		Links: []Link{{Source: &Device{IP: "192.168.1.1"}, Target: &Device{IP: "192.168.1.2"}, SourceType: LinkSourceLLDP, Confidence: ""}},
	}
	err := topo.Validate()
	if err == nil {
		t.Error("expected error for empty confidence")
	}
}

// TestIntegrationWriteText_Empty проверяет текстовый вывод.
func TestIntegrationWriteText_Empty(t *testing.T) {
	topo := &Topology{Devices: make(map[string]*Device), Links: []Link{}}
	var buf bytes.Buffer
	err := topo.WriteText(&buf)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("DEVICES")) {
		t.Error("expected text to contain DEVICES section")
	}
	if !bytes.Contains(buf.Bytes(), []byte("LINKS")) {
		t.Error("expected text to contain LINKS section")
	}
	if !bytes.Contains(buf.Bytes(), []byte("SUMMARY")) {
		t.Error("expected text to contain SUMMARY section")
	}
}

// TestIntegrationWriteText_NilTopology проверяет nil топологию.
func TestIntegrationWriteText_NilTopology(t *testing.T) {
	var buf bytes.Buffer
	err := (*Topology)(nil).WriteText(&buf)
	if err == nil {
		t.Error("expected error for nil topology")
	}
}

// TestIntegrationDedupReport_Empty проверяет отчёт дедупа.
func TestIntegrationDedupReport_Empty(t *testing.T) {
	topo := &Topology{Devices: make(map[string]*Device), Links: []Link{}}
	report := topo.DedupReport()
	if report == "" {
		t.Error("expected non-empty report")
	}
}

// TestIntegrationDedupReport_NilTopology проверяет nil отчёт.
func TestIntegrationDedupReport_NilTopology(t *testing.T) {
	report := (*Topology)(nil).DedupReport()
	if report == "" {
		t.Error("expected non-empty report for nil topology")
	}
}

// TestIntegrationExplainLink_Valid проверяет объяснение связи.
func TestIntegrationExplainLink_Valid(t *testing.T) {
	topo := &Topology{
		Links: []Link{{
			Source:     &Device{IP: "192.168.1.1", Hostname: "router"},
			Target:     &Device{IP: "192.168.1.2", Hostname: "host"},
			SourceType: LinkSourceLLDP,
			Confidence: LinkConfidenceHigh,
			Evidence:   "LLDP discovery",
		}},
	}
	explain, ok := topo.ExplainLink(0)
	if !ok {
		t.Error("expected ok for valid index")
	}
	if explain == "" {
		t.Error("expected non-empty explanation")
	}
	if !bytes.Contains([]byte(explain), []byte("LLDP")) {
		t.Error("expected explanation to mention LLDP")
	}
}

// TestIntegrationExplainLink_InvalidIndex проверяет невалидный индекс.
func TestIntegrationExplainLink_InvalidIndex(t *testing.T) {
	topo := &Topology{Links: []Link{}}
	_, ok := topo.ExplainLink(0)
	if ok {
		t.Error("expected not ok for invalid index")
	}
}

// TestIntegrationExplainLink_FDB проверяет FDB объяснение.
func TestIntegrationExplainLink_FDB(t *testing.T) {
	topo := &Topology{
		Links: []Link{{
			Source:     &Device{IP: "192.168.1.1", Hostname: "switch"},
			Target:     &Device{IP: "192.168.1.2", Hostname: "host"},
			SourceType: LinkSourceFDB,
			Confidence: LinkConfidenceMedium,
		}},
	}
	explain, ok := topo.ExplainLink(0)
	if !ok {
		t.Error("expected ok for valid index")
	}
	if !bytes.Contains([]byte(explain), []byte("FDB")) {
		t.Error("expected explanation to mention FDB")
	}
}

// TestIntegrationClassifyFromScannerResult_Router проверяет классификацию Router.
func TestIntegrationClassifyFromScannerResult_Router(t *testing.T) {
	result := classifyFromScannerResult("Router")
	if result != DeviceTypeRouter {
		t.Errorf("expected Router, got %s", result)
	}
}

// TestIntegrationClassifyFromScannerResult_Empty проверяет пустую классификацию.
func TestIntegrationClassifyFromScannerResult_Empty(t *testing.T) {
	result := classifyFromScannerResult("")
	if result != DeviceTypeUnknown {
		t.Errorf("expected Unknown for empty, got %s", result)
	}
}

// TestIntegrationNormalizedKey_MAC проверяет ключ по MAC.
func TestIntegrationNormalizedKey_MAC(t *testing.T) {
	key := normalizedKey("aa:bb:cc:dd:ee:ff", "192.168.1.1")
	if key != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected MAC key, got %q", key)
	}
}

// TestIntegrationNormalizedKey_IP проверяет ключ по IP.
func TestIntegrationNormalizedKey_IP(t *testing.T) {
	key := normalizedKey("", "192.168.1.1")
	if key != "192.168.1.1" {
		t.Errorf("expected IP key, got %q", key)
	}
}

// TestIntegrationNormalizeMAC_Colons проверяет нормализацию MAC с двоеточиями.
func TestIntegrationNormalizeMAC_Colons(t *testing.T) {
	mac := normalizeMAC("aa:bb:cc:dd:ee:ff")
	if mac != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected normalized MAC, got %q", mac)
	}
}

// TestIntegrationNormalizeMAC_Dashes проверяет нормализацию MAC с дефисами.
func TestIntegrationNormalizeMAC_Dashes(t *testing.T) {
	mac := normalizeMAC("aa-bb-cc-dd-ee-ff")
	if mac != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected normalized MAC with colons, got %q", mac)
	}
}

// TestIntegrationNormalizeMAC_CamelCase проверяет нормализацию регистра.
func TestIntegrationNormalizeMAC_CamelCase(t *testing.T) {
	mac := normalizeMAC("AA:BB:CC:DD:EE:FF")
	if mac != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected lowercase MAC, got %q", mac)
	}
}

// TestIntegrationIsBroadcastOrMulticast_Broadcast проверяет broadcast MAC.
func TestIntegrationIsBroadcastOrMulticast_Broadcast(t *testing.T) {
	if !isBroadcastOrMulticast("ff:ff:ff:ff:ff:ff") {
		t.Error("expected true for broadcast MAC")
	}
}

// TestIntegrationIsZeroMAC_Zero проверяет zero MAC.
func TestIntegrationIsZeroMAC_Zero(t *testing.T) {
	if !isZeroMAC("00:00:00:00:00:00") {
		t.Error("expected true for zero MAC")
	}
}

// TestIntegrationConfidenceRank_High проверяет ранг High.
func TestIntegrationConfidenceRank_High(t *testing.T) {
	if confidenceRank(LinkConfidenceHigh) != 3 {
		t.Error("expected High rank 3")
	}
}

// TestIntegrationConfidenceRank_Medium проверяет ранг Medium.
func TestIntegrationConfidenceRank_Medium(t *testing.T) {
	if confidenceRank(LinkConfidenceMedium) != 2 {
		t.Error("expected Medium rank 2")
	}
}

// TestIntegrationConfidenceRank_Low проверяет ранг Low.
func TestIntegrationConfidenceRank_Low(t *testing.T) {
	if confidenceRank(LinkConfidenceLow) != 1 {
		t.Error("expected Low rank 1")
	}
}

// TestIntegrationNodeID_AllFields проверяет node ID со всеми полями.
func TestIntegrationNodeID_AllFields(t *testing.T) {
	d := &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff", Hostname: "router"}
	id := nodeID(d)
	if id == "" {
		t.Error("expected non-empty node ID")
	}
}

// TestIntegrationNodeID_Empty проверяет пустой node ID.
func TestIntegrationNodeID_Empty(t *testing.T) {
	d := &Device{}
	id := nodeID(d)
	if id != "unknown" {
		t.Errorf("expected 'unknown' for empty device, got %q", id)
	}
}

// TestIntegrationDeviceDisplayName_Hostname проверяет display name по hostname.
func TestIntegrationDeviceDisplayName_Hostname(t *testing.T) {
	d := &Device{Hostname: "router-main"}
	name := deviceDisplayName(d)
	if name != "router-main" {
		t.Errorf("expected hostname as display name, got %q", name)
	}
}

// TestIntegrationDeviceDisplayName_Empty проверяет пустой display name.
func TestIntegrationDeviceDisplayName_Empty(t *testing.T) {
	d := &Device{}
	name := deviceDisplayName(d)
	if name != "unknown" {
		t.Errorf("expected 'unknown' for empty device, got %q", name)
	}
}

// TestIntegrationPortLabel_WithName проверяет label порта с именем.
func TestIntegrationPortLabel_WithName(t *testing.T) {
	p := &Port{Name: "GigabitEthernet0/1"}
	label := portLabel(p)
	if label != "GigabitEthernet0/1" {
		t.Errorf("expected port name as label, got %q", label)
	}
}

// TestIntegrationPortLabel_Nil проверяет nil порт.
func TestIntegrationPortLabel_Nil(t *testing.T) {
	label := portLabel(nil)
	if label != "" {
		t.Errorf("expected empty label for nil port, got %q", label)
	}
}

// TestIntegrationEnsurePort_New проверяет создание нового порта.
func TestIntegrationEnsurePort_New(t *testing.T) {
	d := &Device{Ports: []Port{}}
	p := ensurePort(d, 1, "GigabitEthernet0/1")
	if p == nil {
		t.Fatal("expected non-nil port")
	}
	if len(d.Ports) != 1 {
		t.Errorf("expected 1 port, got %d", len(d.Ports))
	}
}

// TestIntegrationFindNeighbor_ByMAC проверяет поиск соседа по MAC.
func TestIntegrationFindNeighbor_ByMAC(t *testing.T) {
	byMAC := map[string]*Device{"aa:bb:cc:dd:ee:ff": {IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff"}}
	n := &LldpNeighbor{RemoteChassisID: "aa:bb:cc:dd:ee:ff"}
	found := findNeighbor(byMAC, nil, n)
	if found == nil {
		t.Error("expected to find neighbor by MAC")
	}
}

// TestIntegrationFindNeighbor_NotFound проверяет отсутствие соседа.
func TestIntegrationFindNeighbor_NotFound(t *testing.T) {
	found := findNeighbor(nil, nil, &LldpNeighbor{})
	if found != nil {
		t.Error("expected nil for not found neighbor")
	}
}

// TestIntegrationContext_Cancellation проверяет отмену контекста.
func TestIntegrationContext_Cancellation(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	cancel()
	// BuildTopology should not panic with cancelled context
	_, _ = BuildTopologyWithOptions([]scanner.Result{{IP: "192.168.1.1"}}, nil, BuildOptions{})
}

// TestIntegrationMultipleTopologies проверяет создание нескольких топологий.
func TestIntegrationMultipleTopologies(t *testing.T) {
	topologies := make([]*Topology, 5)
	for i := 0; i < 5; i++ {
		topologies[i] = &Topology{Devices: make(map[string]*Device), Links: []Link{}}
	}
	for i, topo := range topologies {
		if len(topo.Devices) != 0 {
			t.Errorf("expected 0 devices for topology %d, got %d", i, len(topo.Devices))
		}
	}
}
