package gui

import (
	"strings"
	"testing"

	"network-scanner/internal/topology"
)

func TestMatchTopologyNodeFilter(t *testing.T) {
	dev := &topology.Device{
		IP:       "192.168.1.10",
		MAC:      "AA:BB:CC:DD:EE:FF",
		Hostname: "core-switch",
		Type:     topology.DeviceTypeSwitch,
	}
	if !matchTopologyNodeFilter(dev, "", "all") {
		t.Fatalf("expected empty query/all type to pass")
	}
	if !matchTopologyNodeFilter(dev, "core", "switch") {
		t.Fatalf("expected hostname query + type switch to pass")
	}
	if matchTopologyNodeFilter(dev, "router", "switch") {
		t.Fatalf("expected unmatched query to fail")
	}
	if matchTopologyNodeFilter(dev, "core", "router") {
		t.Fatalf("expected mismatched type to fail")
	}
}

func TestMatchTopologyLinkConfidence(t *testing.T) {
	link := topology.Link{
		Confidence: topology.LinkConfidenceMedium,
	}
	if !matchTopologyLinkConfidence(link, "all") {
		t.Fatalf("expected all filter to pass")
	}
	if !matchTopologyLinkConfidence(link, "medium") {
		t.Fatalf("expected medium filter to pass")
	}
	if matchTopologyLinkConfidence(link, "high") {
		t.Fatalf("expected high filter to fail for medium link")
	}
}

func TestLinkSummaryIncludesEvidence(t *testing.T) {
	link := topology.Link{
		Source:     &topology.Device{IP: "192.168.1.1", Hostname: "r1"},
		Target:     &topology.Device{IP: "192.168.1.2", Hostname: "sw1"},
		SourceType: topology.LinkSourceLLDP,
		Confidence: topology.LinkConfidenceHigh,
		Evidence:   "lldp_neighbor_match",
	}
	got := linkSummary(link)
	if !strings.Contains(got, "evidence=lldp_neighbor_match") {
		t.Fatalf("expected evidence in summary, got: %s", got)
	}
}

func TestLimitTopologyKeys(t *testing.T) {
	keys := []string{"a", "b", "c", "d"}
	limited, trimmed := limitTopologyKeys(keys, 2)
	if !trimmed {
		t.Fatalf("expected trimmed=true")
	}
	if len(limited) != 2 || limited[0] != "a" || limited[1] != "b" {
		t.Fatalf("unexpected limited keys: %#v", limited)
	}
}
