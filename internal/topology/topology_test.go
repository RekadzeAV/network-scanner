package topology

import (
	"bytes"
	"strings"
	"testing"

	"network-scanner/internal/scanner"
)

func TestBuildTopology(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "sw1", SNMPEnabled: true},
		{IP: "192.168.1.10", MAC: "ba:bb:bb:bb:bb:bb", Hostname: "pc1"},
	}
	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "192.168.1.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "sw1",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			MacTable: map[string]int{
				"ba:bb:bb:bb:bb:bb": 2,
			},
		},
	}
	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}
	if len(topo.Devices) < 2 {
		t.Fatalf("expected at least 2 devices, got %d", len(topo.Devices))
	}
	if len(topo.Links) == 0 {
		t.Fatalf("expected links from MAC table, got none")
	}
	if topo.Links[0].SourceType != LinkSourceFDB {
		t.Fatalf("expected FDB source type, got %s", topo.Links[0].SourceType)
	}
	if topo.Links[0].Confidence != LinkConfidenceMedium {
		t.Fatalf("expected medium confidence, got %s", topo.Links[0].Confidence)
	}
}

func TestToDOT(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"a": {IP: "192.168.1.1"},
		},
	}
	var b bytes.Buffer
	if err := topo.ToDOT(&b); err != nil {
		t.Fatalf("ToDOT error: %v", err)
	}
	dot := b.String()
	if !strings.Contains(dot, "graph network") {
		t.Fatalf("expected DOT header, got: %s", dot)
	}
}

func TestBuildTopologySkipsBroadcastAndMulticastFromFDB(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "sw1", SNMPEnabled: true},
	}
	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "192.168.1.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "sw1",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			MacTable: map[string]int{
				"ff:ff:ff:ff:ff:ff": 1,
				"01:00:5e:00:00:01": 2,
				"00:00:00:00:00:00": 3,
			},
		},
	}
	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}
	if len(topo.Links) != 0 {
		t.Fatalf("expected no links for broadcast/multicast/zero mac, got %d", len(topo.Links))
	}
}

func TestBuildTopologyPrefersLLDPOverFDBForSameEndpoints(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "sw1", SNMPEnabled: true},
		{IP: "192.168.1.2", MAC: "ba:bb:bb:bb:bb:bb", Hostname: "sw2", SNMPEnabled: true},
	}
	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "192.168.1.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "sw1",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			MacTable: map[string]int{
				"ba:bb:bb:bb:bb:bb": 10,
			},
			LldpNeighbors: []*LldpNeighbor{
				{LocalIfIndex: 10, RemoteChassisID: "ba:bb:bb:bb:bb:bb", RemotePortID: "Gi0/1", RemoteSysName: "sw2"},
			},
		},
	}
	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}
	if len(topo.Links) != 1 {
		t.Fatalf("expected a single deduplicated link, got %d", len(topo.Links))
	}
	if topo.Links[0].SourceType != LinkSourceLLDP {
		t.Fatalf("expected LLDP to win over FDB, got %s", topo.Links[0].SourceType)
	}
	if topo.Links[0].Confidence != LinkConfidenceHigh {
		t.Fatalf("expected high confidence, got %s", topo.Links[0].Confidence)
	}
}
