package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

func TestSortedResultsForDisplay(t *testing.T) {
	in := []scanner.Result{
		{IP: "192.168.1.20", Hostname: "b"},
		{IP: "192.168.1.3", Hostname: "a"},
		{IP: "10.0.0.2", Hostname: "z"},
	}
	got := sortedResultsForDisplay(in)
	if len(got) != 3 {
		t.Fatalf("expected 3 results, got %d", len(got))
	}
	if got[0].IP != "10.0.0.2" || got[1].IP != "192.168.1.3" || got[2].IP != "192.168.1.20" {
		t.Fatalf("unexpected sort order: %#v", []string{got[0].IP, got[1].IP, got[2].IP})
	}
	// Ensure input slice is not mutated.
	if in[0].IP != "192.168.1.20" {
		t.Fatalf("input slice was mutated")
	}
}

func TestOpenPortLabels(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 22, Protocol: "tcp", Service: "ssh", State: "open"},
		{Port: 53, Protocol: "udp", Service: "domain", State: "open"},
		{Port: 80, Protocol: "tcp", Service: "http", State: "closed"},
	}
	labels := openPortLabels(ports, 24)
	if len(labels) != 2 {
		t.Fatalf("expected 2 open labels, got %d", len(labels))
	}
	if labels[0] != "22/TCP ssh" {
		t.Fatalf("unexpected first label: %q", labels[0])
	}
	if labels[1] != "53/UDP domain" {
		t.Fatalf("unexpected second label: %q", labels[1])
	}
}

func TestOpenPortLabelsLimit(t *testing.T) {
	ports := make([]scanner.PortInfo, 0, 30)
	for i := 1; i <= 30; i++ {
		ports = append(ports, scanner.PortInfo{
			Port:     1000 + i,
			Protocol: "tcp",
			Service:  "svc",
			State:    "open",
		})
	}
	labels := openPortLabels(ports, 24)
	if len(labels) != 25 {
		t.Fatalf("expected 25 labels (24 + overflow), got %d", len(labels))
	}
	if labels[len(labels)-1] != "+6" {
		t.Fatalf("expected overflow label +6, got %q", labels[len(labels)-1])
	}
}

func TestSortedResultsForDisplayWithModeHostName(t *testing.T) {
	in := []scanner.Result{
		{IP: "192.168.1.20", Hostname: "zeta"},
		{IP: "192.168.1.3", Hostname: "alpha"},
		{IP: "10.0.0.2", Hostname: "beta"},
	}
	got := sortedResultsForDisplayWithMode(in, "HostName")
	if got[0].Hostname != "alpha" || got[1].Hostname != "beta" || got[2].Hostname != "zeta" {
		t.Fatalf("unexpected hostname sort order: %#v", []string{got[0].Hostname, got[1].Hostname, got[2].Hostname})
	}
}

func TestFilterResultsForDisplay(t *testing.T) {
	in := []scanner.Result{
		{Hostname: "router-main", IP: "192.168.1.1", MAC: "aa:bb", DeviceType: "Network Device"},
		{Hostname: "workstation", IP: "192.168.1.10", MAC: "cc:dd", DeviceType: "Computer"},
	}
	got := filterResultsForDisplay(in, "router")
	if len(got) != 1 || got[0].Hostname != "router-main" {
		t.Fatalf("unexpected filter result: %#v", got)
	}
	got = filterResultsForDisplay(in, "192.168.1.10")
	if len(got) != 1 || got[0].Hostname != "workstation" {
		t.Fatalf("unexpected filter by IP result: %#v", got)
	}
}

func TestFilterResultsForDisplayAdvanced(t *testing.T) {
	in := []scanner.Result{
		{
			Hostname:   "router-main",
			IP:         "192.168.1.1",
			DeviceType: "Router",
			Ports:      []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}},
		},
		{
			Hostname:   "pc-1",
			IP:         "192.168.1.10",
			DeviceType: "Windows Computer",
			Ports:      []scanner.PortInfo{{Port: 445, Protocol: "tcp", State: "closed"}},
		},
	}

	got := filterResultsForDisplayAdvanced(in, "", []string{"Network Device"}, false)
	if len(got) != 1 || got[0].Hostname != "router-main" {
		t.Fatalf("unexpected type-filter result: %#v", got)
	}

	got = filterResultsForDisplayAdvanced(in, "", nil, true)
	if len(got) != 1 || got[0].Hostname != "router-main" {
		t.Fatalf("unexpected open-port-filter result: %#v", got)
	}

	got = filterResultsForDisplayAdvanced(in, "pc", []string{"Computer"}, false)
	if len(got) != 1 || got[0].Hostname != "pc-1" {
		t.Fatalf("unexpected combined-filter result: %#v", got)
	}
}

func TestNormalizeDeviceTypes(t *testing.T) {
	raw := map[string]int{
		"Router":           1,
		"Network Device":   2,
		"Windows Computer": 2,
		"Linux Server":     1,
		"Unknown Device":   3,
		"NAS":              4,
	}
	got := normalizeDeviceTypes(raw)
	if got["Network Device"] != 3 {
		t.Fatalf("expected Network Device=3, got %d", got["Network Device"])
	}
	if got["Computer"] != 2 {
		t.Fatalf("expected Computer=2, got %d", got["Computer"])
	}
	if got["Server"] != 1 {
		t.Fatalf("expected Server=1, got %d", got["Server"])
	}
	if got["Unknown"] != 3 {
		t.Fatalf("expected Unknown=3, got %d", got["Unknown"])
	}
	if got["NAS"] != 4 {
		t.Fatalf("expected NAS=4, got %d", got["NAS"])
	}
}
