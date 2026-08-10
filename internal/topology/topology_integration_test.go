package topology

import (
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
