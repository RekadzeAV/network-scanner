package gui

import (
	"image/color"
	"testing"

	"network-scanner/internal/topology"
)

// --- topology_interactive_map.go extended tests (без дублей) ---

func TestSelectHostByTopologyDevice_NilApp(t *testing.T) {
	var a *App
	a.selectHostByTopologyDevice(&topology.Device{IP: "192.168.1.1"})
}

func TestSelectHostByTopologyDevice_NilDevice(t *testing.T) {
	a := &App{}
	a.selectHostByTopologyDevice(nil)
}

func TestSelectHostByTopologyDevice_EmptyIP(t *testing.T) {
	a := &App{}
	a.selectHostByTopologyDevice(&topology.Device{IP: ""})
}

func TestSelectHostByTopologyDevice_WithIP(t *testing.T) {
	a := &App{}
	a.selectHostByTopologyDevice(&topology.Device{IP: "192.168.1.1"})
	if a.selectedHostIP != "192.168.1.1" {
		t.Errorf("expected selectedHostIP='192.168.1.1', got %q", a.selectedHostIP)
	}
}

func TestRenderTopologyInteractiveMap_NilApp(t *testing.T) {
	var a *App
	a.renderTopologyInteractiveMap(nil)
}

func TestRenderTopologyInteractiveMap_NilGraphBox(t *testing.T) {
	a := &App{}
	a.renderTopologyInteractiveMap(nil)
}

func TestRenderTopologyInteractiveMap_EmptyTopology(t *testing.T) {
	a := &App{}
	topo := &topology.Topology{}
	a.renderTopologyInteractiveMap(topo)
}

func TestColorByConfidence_Medium(t *testing.T) {
	c := colorByConfidence(topology.LinkConfidenceMedium)
	if c == nil {
		t.Fatal("expected non-nil color")
	}
	rgba, ok := c.(color.RGBA)
	if !ok {
		t.Fatal("expected color.RGBA")
	}
	if rgba.R != 251 || rgba.G != 188 || rgba.B != 4 {
		t.Errorf("expected yellow color, got %v", rgba)
	}
}

func TestColorByConfidence_Low(t *testing.T) {
	c := colorByConfidence(topology.LinkConfidenceLow)
	if c == nil {
		t.Fatal("expected non-nil color")
	}
	rgba, ok := c.(color.RGBA)
	if !ok {
		t.Fatal("expected color.RGBA")
	}
	if rgba.R != 234 || rgba.G != 67 || rgba.B != 53 {
		t.Errorf("expected red color, got %v", rgba)
	}
}

func TestColorByConfidence_Unknown(t *testing.T) {
	c := colorByConfidence(topology.LinkConfidence("unknown"))
	if c == nil {
		t.Fatal("expected non-nil color")
	}
}

func TestColorByDeviceType_AllTypes(t *testing.T) {
	types := []topology.DeviceType{
		topology.DeviceTypeRouter,
		topology.DeviceTypeSwitch,
		topology.DeviceTypeHost,
		topology.DeviceType("unknown"),
	}
	for _, dt := range types {
		c := colorByDeviceType(dt)
		if c == nil {
			t.Errorf("expected non-nil color for type %q", dt)
		}
	}
}

func TestLinkBadge_WithSourceType(t *testing.T) {
	l := topology.Link{
		SourceType: "snmp",
		Confidence: "high",
	}
	badge := linkBadge(l)
	if badge != "SNMP/high" {
		t.Errorf("expected 'SNMP/high', got %q", badge)
	}
}

func TestLinkSummary_WithEvidence(t *testing.T) {
	l := topology.Link{
		Evidence: "SNMP sysDescr",
	}
	summary := linkSummary(l)
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestTopologyLinkKey_WithDevices(t *testing.T) {
	l := topology.Link{
		Source: &topology.Device{Hostname: "Router1"},
		Target: &topology.Device{IP: "192.168.1.2"},
	}
	key := topologyLinkKey(l)
	if key == "" {
		t.Fatal("expected non-empty key")
	}
}

func TestMatchTopologyNodeFilter_TypeFilterCaseInsensitive(t *testing.T) {
	dev := &topology.Device{Type: topology.DeviceTypeRouter}
	if !matchTopologyNodeFilter(dev, "", "ROUTER") {
		t.Error("expected true for case-insensitive typeFilter")
	}
}

func TestMatchTopologyLinkConfidence_CaseInsensitive(t *testing.T) {
	l := topology.Link{Confidence: topology.LinkConfidenceHigh}
	if !matchTopologyLinkConfidence(l, "HIGH") {
		t.Error("expected true for case-insensitive filter")
	}
}

func TestLimitTopologyKeys_NegativeMax(t *testing.T) {
	keys := []string{"a", "b", "c"}
	result, trimmed := limitTopologyKeys(keys, -1)
	if trimmed {
		t.Error("expected trimmed=false for negative max")
	}
	if len(result) != 3 {
		t.Errorf("expected 3 keys, got %d", len(result))
	}
}

func TestLimitTopologyKeys_ExactMax(t *testing.T) {
	keys := []string{"a", "b", "c"}
	result, trimmed := limitTopologyKeys(keys, 3)
	if trimmed {
		t.Error("expected trimmed=false for exact max")
	}
	if len(result) != 3 {
		t.Errorf("expected 3 keys, got %d", len(result))
	}
}
