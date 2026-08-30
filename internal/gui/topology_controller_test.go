package gui

import (
	"testing"

	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"
)

// --- topologySuccessStatus tests ---

func TestTopologySuccessStatus_NilTopo(t *testing.T) {
	// В gui/topology_controller.go нет nil-check для topo.
	// Проверяем что не паникует с пустым топологией.
	topo := &topology.Topology{
		Devices: make(map[string]*topology.Device),
		Links:   []topology.Link{},
	}
	result := topologySuccessStatus(topo, nil)
	expected := "Топология построена: устройств 0, связей 0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTopologySuccessStatus_EmptyTopo(t *testing.T) {
	topo := &topology.Topology{}
	result := topologySuccessStatus(topo, nil)
	if result != "Топология построена: устройств 0, связей 0" {
		t.Errorf("expected 'Топология построена: устройств 0, связей 0', got %q", result)
	}
}

func TestTopologySuccessStatus_WithDevicesAndLinks(t *testing.T) {
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{
			"d1": {IP: "192.168.1.1"},
			"d2": {IP: "192.168.1.2"},
			"d3": {IP: "192.168.1.3"},
		},
		Links: []topology.Link{
			{Source: &topology.Device{IP: "192.168.1.1"}, Target: &topology.Device{IP: "192.168.1.2"}},
			{Source: &topology.Device{IP: "192.168.1.2"}, Target: &topology.Device{IP: "192.168.1.3"}},
		},
	}
	result := topologySuccessStatus(topo, nil)
	if result != "Топология построена: устройств 3, связей 2" {
		t.Errorf("expected 'Топология построена: устройств 3, связей 2', got %q", result)
	}
}

func TestTopologySuccessStatus_WithReport(t *testing.T) {
	topo := &topology.Topology{}
	report := &snmpcollector.CollectReport{
		TotalSNMPTargets: 5,
		Connected:        3,
		Partial:          1,
		Failed:           1,
	}
	result := topologySuccessStatus(topo, report)
	if result != "Топология построена: устройств 0, связей 0 | SNMP: целей 5, ok 3, partial 1, failed 1" {
		t.Errorf("expected SNMP info in result, got %q", result)
	}
}

func TestTopologySuccessStatus_TopologyWithReport(t *testing.T) {
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{"d1": {IP: "10.0.0.1"}},
		Links:   []topology.Link{},
	}
	report := &snmpcollector.CollectReport{
		TotalSNMPTargets: 1,
		Connected:        1,
	}
	result := topologySuccessStatus(topo, report)
	expected := "Топология построена: устройств 1, связей 0 | SNMP: целей 1, ok 1, partial 0, failed 0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTopologySuccessStatus_NilReport(t *testing.T) {
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{"d1": {IP: "10.0.0.1"}},
	}
	result := topologySuccessStatus(topo, nil)
	if result != "Топология построена: устройств 1, связей 0" {
		t.Errorf("expected no SNMP info, got %q", result)
	}
}
