package controller

import (
	"testing"

	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"
)

// Ported from the pre-refactor gui-package tests (v2.2 remote) during the
// 2026-09-15 merge: scenarios not covered by pure_functions_test.go.

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

// Current implementation keeps status line compact; SNMP report details are
// rendered separately by formatTopologyPreview. These tests pin that contract.

func TestTopologySuccessStatus_WithReport_IgnoredInStatus(t *testing.T) {
	topo := &topology.Topology{}
	report := &snmpcollector.CollectReport{
		TotalSNMPTargets: 5,
		Connected:        3,
		Partial:          1,
		Failed:           1,
	}
	result := topologySuccessStatus(topo, report)
	if result != "Топология построена: устройств 0, связей 0" {
		t.Errorf("expected compact status without SNMP, got %q", result)
	}
}

func TestTopologySuccessStatus_TopologyWithReport_IgnoredInStatus(t *testing.T) {
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{"d1": {IP: "10.0.0.1"}},
		Links:   []topology.Link{},
	}
	report := &snmpcollector.CollectReport{
		TotalSNMPTargets: 1,
		Connected:        1,
	}
	result := topologySuccessStatus(topo, report)
	expected := "Топология построена: устройств 1, связей 0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
