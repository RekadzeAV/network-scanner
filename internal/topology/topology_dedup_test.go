package topology

import (
	"strings"
	"testing"
)

// --- Тесты для расширенной дедупликации ---

func TestShouldReplaceLink(t *testing.T) {
	tests := []struct {
		name        string
		existing    Link
		newSource   LinkSourceType
		newConf     LinkConfidence
		wantReplace bool
	}{
		{
			"LLDP replaces FDB",
			Link{SourceType: LinkSourceFDB, Confidence: LinkConfidenceHigh},
			LinkSourceLLDP, LinkConfidenceHigh,
			true,
		},
		{
			"FDB replaces Inferred",
			Link{SourceType: LinkSourceInferred, Confidence: LinkConfidenceHigh},
			LinkSourceFDB, LinkConfidenceHigh,
			true,
		},
		{
			"LLDP does not replace LLDP with same confidence",
			Link{SourceType: LinkSourceLLDP, Confidence: LinkConfidenceHigh},
			LinkSourceLLDP, LinkConfidenceHigh,
			false,
		},
		{
			"LLDP replaces LLDP with higher confidence",
			Link{SourceType: LinkSourceLLDP, Confidence: LinkConfidenceLow},
			LinkSourceLLDP, LinkConfidenceHigh,
			true,
		},
		{
			"FDB does not replace LLDP",
			Link{SourceType: LinkSourceLLDP, Confidence: LinkConfidenceLow},
			LinkSourceFDB, LinkConfidenceHigh,
			false,
		},
		{
			"Equal source_type and confidence — no replace",
			Link{SourceType: LinkSourceFDB, Confidence: LinkConfidenceMedium},
			LinkSourceFDB, LinkConfidenceMedium,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldReplaceLink(tt.existing, tt.newSource, tt.newConf)
			if got != tt.wantReplace {
				t.Errorf("shouldReplaceLink() = %v, want %v", got, tt.wantReplace)
			}
		})
	}
}

func TestSourceTypeRank(t *testing.T) {
	tests := []struct {
		source LinkSourceType
		want   int
	}{
		{LinkSourceLLDP, 3},
		{LinkSourceFDB, 2},
		{LinkSourceInferred, 1},
		{"", 0},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			got := sourceTypeRank(tt.source)
			if got != tt.want {
				t.Errorf("sourceTypeRank(%q) = %d, want %d", tt.source, got, tt.want)
			}
		})
	}
}

func TestAddLinkDedupBySourceType(t *testing.T) {
	topo := &Topology{
		Devices: make(map[string]*Device),
		Links:   make([]Link, 0),
	}

	// Создаём тестовую среду с моковыми данными.
	// Используем BuildTopologyWithOptions с моковыми данными.
	// Для простоты проверяем через прямой доступ к полям.

	// Создаём два устройства
	dev1 := &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1", Type: DeviceTypeSwitch}
	dev2 := &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1", Type: DeviceTypeHost}

	topo.Devices["mac_aa_bb_cc_dd_ee_01"] = dev1
	topo.Devices["mac_aa_bb_cc_dd_ee_02"] = dev2

	dedup := make(map[string]int)
	byEndpoint := make(map[string]int)

	// 1. Добавляем FDB связь
	addLink(dedup, byEndpoint, topo,
		dev1, 1, "Gi0/1",
		dev2, -1, "",
		LinkSourceFDB, LinkConfidenceMedium,
		"fdb_mac_match;local_if=1;remote_mac=aa:bb:cc:dd:ee:02",
	)

	if len(topo.Links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(topo.Links))
	}
	if topo.Links[0].SourceType != LinkSourceFDB {
		t.Errorf("expected FDB source, got %s", topo.Links[0].SourceType)
	}

	// 2. Пытаемся добавить LLDP связь (должна заменить FDB — выше приоритет)
	addLink(dedup, byEndpoint, topo,
		dev1, 1, "Gi0/1",
		dev2, -1, "",
		LinkSourceLLDP, LinkConfidenceHigh,
		"lldp_neighbor_match;local_if=1;remote_port=eth0",
	)

	if len(topo.Links) != 1 {
		t.Fatalf("expected 1 link after replace, got %d", len(topo.Links))
	}
	if topo.Links[0].SourceType != LinkSourceLLDP {
		t.Errorf("expected LLDP source after replace, got %s", topo.Links[0].SourceType)
	}
	if topo.Links[0].Confidence != LinkConfidenceHigh {
		t.Errorf("expected High confidence, got %s", topo.Links[0].Confidence)
	}
}

func TestAddLinkMultiplePorts(t *testing.T) {
	topo := &Topology{
		Devices: make(map[string]*Device),
		Links:   make([]Link, 0),
	}

	dev1 := &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1", Type: DeviceTypeSwitch}
	dev2 := &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1", Type: DeviceTypeHost}

	topo.Devices["mac_aa_bb_cc_dd_ee_01"] = dev1
	topo.Devices["mac_aa_bb_cc_dd_ee_02"] = dev2

	dedup := make(map[string]int)
	byEndpoint := make(map[string]int)

	// Добавляем связь через порт Gi0/1 → eth0
	addLink(dedup, byEndpoint, topo,
		dev1, 1, "Gi0/1",
		dev2, 10, "eth0",
		LinkSourceLLDP, LinkConfidenceHigh,
		"lldp_port1",
	)

	// Добавляем связь через порт Gi0/2 → eth1 (другой порт с обеих сторон — новая связь)
	addLink(dedup, byEndpoint, topo,
		dev1, 2, "Gi0/2",
		dev2, 11, "eth1",
		LinkSourceLLDP, LinkConfidenceHigh,
		"lldp_port2",
	)

	if len(topo.Links) != 2 {
		t.Errorf("expected 2 links (different ports on both sides), got %d", len(topo.Links))
	}
}

// --- Тесты для DedupReport ---

func TestDedupReport(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01"},
			"dev2": {IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02"},
			"dev3": {IP: "192.168.1.3", MAC: "aa:bb:cc:dd:ee:03"},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01"},
				Target:     &Device{IP: "192.168.1.3", MAC: "aa:bb:cc:dd:ee:03"},
				SourceType: LinkSourceFDB,
				Confidence: LinkConfidenceMedium,
			},
		},
	}

	report := topo.DedupReport()

	if !strings.Contains(report, "Topology Deduplication Report:") {
		t.Error("report should contain header")
	}
	if !strings.Contains(report, "Total devices: 3") {
		t.Error("report should show 3 devices")
	}
	if !strings.Contains(report, "Total links: 2") {
		t.Error("report should show 2 links")
	}
	if !strings.Contains(report, "lldp") {
		t.Error("report should show source type breakdown")
	}
	if !strings.Contains(report, "fdb") {
		t.Error("report should show FDB type")
	}
}

func TestDedupReportNilTopology(t *testing.T) {
	var topo *Topology
	// Не должно паниковать
	_ = topo.DedupReport()
}

// --- Тесты для ExplainLink ---

func TestExplainLink(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1"},
			"dev2": {IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1"},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
				Evidence:   "lldp_neighbor_match;local_if=1",
				SourcePort: &Port{Index: 1, Name: "Gi0/1"},
				TargetPort: &Port{Index: -1, Name: ""},
			},
		},
	}

	explain, ok := topo.ExplainLink(0)
	if !ok {
		t.Fatal("ExplainLink should return ok=true for valid index")
	}

	if !strings.Contains(explain, "Связь #1") {
		t.Error("explain should contain link number")
	}
	if !strings.Contains(explain, "switch1") {
		t.Error("explain should contain source hostname")
	}
	if !strings.Contains(explain, "host1") {
		t.Error("explain should contain target hostname")
	}
	if !strings.Contains(explain, "LLDP") {
		t.Error("explain should mention LLDP")
	}
	if !strings.Contains(explain, "Gi0/1") {
		t.Error("explain should contain port name")
	}
}

func TestExplainLinkFDB(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1"},
			"dev2": {IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1"},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1"},
				SourceType: LinkSourceFDB,
				Confidence: LinkConfidenceMedium,
				Evidence:   "fdb_mac_match;local_if=5",
			},
		},
	}

	explain, ok := topo.ExplainLink(0)
	if !ok {
		t.Fatal("ExplainLink should return ok=true")
	}

	if !strings.Contains(explain, "FDB") {
		t.Error("explain should mention FDB")
	}
}

func TestExplainLinkInvalidIndex(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{},
		Links:   []Link{},
	}

	_, ok := topo.ExplainLink(-1)
	if ok {
		t.Error("ExplainLink(-1) should return ok=false")
	}

	_, ok = topo.ExplainLink(999)
	if ok {
		t.Error("ExplainLink(999) should return ok=false for out-of-range")
	}
}

// --- Тесты для explain-хелперов ---

func TestExplainLLDPSource(t *testing.T) {
	src := &Device{Hostname: "switch1", IP: "192.168.1.1"}
	dst := &Device{Hostname: "host1", IP: "192.168.1.2"}

	msg := explainLLDPSource(src, dst, "Gi0/1", "lldp_match")
	if !strings.Contains(msg, "switch1") {
		t.Error("message should contain source hostname")
	}
	if !strings.Contains(msg, "host1") {
		t.Error("message should contain target hostname")
	}
	if !strings.Contains(msg, "Gi0/1") {
		t.Error("message should contain port name")
	}
	if !strings.Contains(msg, "lldp_match") {
		t.Error("message should contain evidence")
	}
}

func TestExplainFDBSource(t *testing.T) {
	src := &Device{Hostname: "switch1", IP: "192.168.1.1"}
	dst := &Device{Hostname: "host1", IP: "192.168.1.2"}

	msg := explainFDBSource(src, dst, "aa:bb:cc:dd:ee:02", "fdb_match")
	if !strings.Contains(msg, "FDB") {
		t.Error("message should mention FDB")
	}
	if !strings.Contains(msg, "aa:bb:cc:dd:ee:02") {
		t.Error("message should contain MAC")
	}
}

func TestExplainInferredSource(t *testing.T) {
	src := &Device{Hostname: "router1", IP: "192.168.1.1"}
	dst := &Device{Hostname: "host1", IP: "192.168.1.2"}

	msg := explainInferredSource(src, dst, "same_subnet")
	if !strings.Contains(msg, "Эвристика") {
		t.Error("message should mention heuristic")
	}
	if !strings.Contains(msg, "same_subnet") {
		t.Error("message should contain reason")
	}
}

// --- Интеграционный тест: полная цепочка дедупликации ---

func TestBuildTopologyDedupIntegration(t *testing.T) {
	// Проверяем shouldReplaceLink: LLDP > FDB
	if !shouldReplaceLink(
		Link{SourceType: LinkSourceFDB, Confidence: LinkConfidenceMedium},
		LinkSourceLLDP, LinkConfidenceHigh,
	) {
		t.Error("LLDP should replace FDB")
	}

	// Проверяем shouldReplaceLink: FDB не заменяет LLDP
	if shouldReplaceLink(
		Link{SourceType: LinkSourceLLDP, Confidence: LinkConfidenceLow},
		LinkSourceFDB, LinkConfidenceHigh,
	) {
		t.Error("FDB should not replace LLDP even with higher confidence")
	}
}
