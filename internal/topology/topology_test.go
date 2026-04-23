package topology

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

func TestBuildTopologyLowersConfidenceForPartialSNMPDevice(t *testing.T) {
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
			LldpNeighbors: []*LldpNeighbor{
				{LocalIfIndex: 10, RemoteChassisID: "ba:bb:bb:bb:bb:bb", RemotePortID: "Gi0/1", RemoteSysName: "sw2"},
			},
		},
	}
	topo, err := BuildTopologyWithOptions(results, snmp, BuildOptions{
		PartialSNMPKeys: map[string]struct{}{
			"ip:192.168.1.1": {},
		},
	})
	if err != nil {
		t.Fatalf("BuildTopologyWithOptions error: %v", err)
	}
	if len(topo.Links) != 1 {
		t.Fatalf("expected one link, got %d", len(topo.Links))
	}
	if topo.Links[0].Confidence != LinkConfidenceMedium {
		t.Fatalf("expected downgraded medium confidence, got %s", topo.Links[0].Confidence)
	}
}

func TestBuildTopologyIncludesDetailedEvidence(t *testing.T) {
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
		t.Fatalf("expected one link, got %d", len(topo.Links))
	}
	got := topo.Links[0].Evidence
	if !strings.Contains(got, "lldp_neighbor_match") || !strings.Contains(got, "local_if=10") || !strings.Contains(got, "remote_port=Gi0/1") {
		t.Fatalf("expected detailed LLDP evidence, got %q", got)
	}
}

func TestValidateRejectsBrokenTopology(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"sw1": {IP: "192.168.1.1"},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.2"},
				Target:     &Device{IP: "192.168.1.1"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}
	if err := topo.Validate(); err == nil {
		t.Fatalf("expected validation error for missing source device")
	}
}

func TestJSONAndGraphMLEquivalenceCounts(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "sw1", SNMPEnabled: true},
		{IP: "192.168.1.2", MAC: "ba:bb:bb:bb:bb:bb", Hostname: "sw2", SNMPEnabled: true},
		{IP: "192.168.1.10", MAC: "ca:cc:cc:cc:cc:cc", Hostname: "host1"},
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
				"ca:cc:cc:cc:cc:cc": 11,
			},
		},
	}
	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "topology.json")
	graphmlPath := filepath.Join(dir, "topology.graphml")
	if err := topo.SaveJSON(jsonPath); err != nil {
		t.Fatalf("SaveJSON error: %v", err)
	}
	if err := topo.SaveGraphML(graphmlPath); err != nil {
		t.Fatalf("SaveGraphML error: %v", err)
	}

	rawJSON, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json file: %v", err)
	}
	var decodedJSON struct {
		Devices map[string]json.RawMessage `json:"Devices"`
		Links   []json.RawMessage          `json:"Links"`
	}
	if err := json.Unmarshal(rawJSON, &decodedJSON); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}

	rawGraphML, err := os.ReadFile(graphmlPath)
	if err != nil {
		t.Fatalf("read graphml file: %v", err)
	}
	var decodedGraphML struct {
		Graph struct {
			Nodes []struct {
				ID   string `xml:"id,attr"`
				Data []struct {
					Key   string `xml:"key,attr"`
					Value string `xml:",chardata"`
				} `xml:"data"`
			} `xml:"node"`
			Edges []struct {
				Source string `xml:"source,attr"`
				Target string `xml:"target,attr"`
			} `xml:"edge"`
		} `xml:"graph"`
	}
	if err := xml.Unmarshal(rawGraphML, &decodedGraphML); err != nil {
		t.Fatalf("unmarshal graphml: %v", err)
	}

	if len(decodedJSON.Devices) != len(decodedGraphML.Graph.Nodes) {
		t.Fatalf("device count mismatch json=%d graphml=%d", len(decodedJSON.Devices), len(decodedGraphML.Graph.Nodes))
	}
	if len(decodedJSON.Links) != len(decodedGraphML.Graph.Edges) {
		t.Fatalf("link count mismatch json=%d graphml=%d", len(decodedJSON.Links), len(decodedGraphML.Graph.Edges))
	}

	type jsonDevice struct {
		IP       string `json:"IP"`
		MAC      string `json:"MAC"`
		Hostname string `json:"Hostname"`
	}
	type jsonLink struct {
		Source struct {
			IP       string `json:"IP"`
			MAC      string `json:"MAC"`
			Hostname string `json:"Hostname"`
		} `json:"Source"`
		Target struct {
			IP       string `json:"IP"`
			MAC      string `json:"MAC"`
			Hostname string `json:"Hostname"`
		} `json:"Target"`
	}
	var fullJSON struct {
		Devices map[string]jsonDevice `json:"Devices"`
		Links   []jsonLink            `json:"Links"`
	}
	if err := json.Unmarshal(rawJSON, &fullJSON); err != nil {
		t.Fatalf("unmarshal full json: %v", err)
	}

	jsonNodes := make([]string, 0, len(fullJSON.Devices))
	for _, d := range fullJSON.Devices {
		jsonNodes = append(jsonNodes, normalizeNodeIdentity(d.Hostname, d.IP, d.MAC))
	}
	sort.Strings(jsonNodes)

	graphmlNodes := make([]string, 0, len(decodedGraphML.Graph.Nodes))
	for _, n := range decodedGraphML.Graph.Nodes {
		var label string
		for _, data := range n.Data {
			if data.Key == "label" {
				label = strings.TrimSpace(data.Value)
				break
			}
		}
		graphmlNodes = append(graphmlNodes, strings.ToLower(strings.TrimSpace(label)))
	}
	sort.Strings(graphmlNodes)
	if fmt.Sprintf("%v", jsonNodes) != fmt.Sprintf("%v", graphmlNodes) {
		t.Fatalf("node identity mismatch json=%v graphml=%v", jsonNodes, graphmlNodes)
	}

	jsonEdges := make([]string, 0, len(fullJSON.Links))
	for _, l := range fullJSON.Links {
		jsonEdges = append(jsonEdges, normalizeUndirectedEdge(
			normalizeNodeIdentity(l.Source.Hostname, l.Source.IP, l.Source.MAC),
			normalizeNodeIdentity(l.Target.Hostname, l.Target.IP, l.Target.MAC),
		))
	}
	sort.Strings(jsonEdges)

	graphmlLabelByID := make(map[string]string, len(decodedGraphML.Graph.Nodes))
	for _, n := range decodedGraphML.Graph.Nodes {
		label := strings.TrimSpace(n.ID)
		for _, data := range n.Data {
			if data.Key == "label" {
				label = strings.TrimSpace(data.Value)
				break
			}
		}
		graphmlLabelByID[strings.TrimSpace(n.ID)] = strings.ToLower(label)
	}
	graphmlEdges := make([]string, 0, len(decodedGraphML.Graph.Edges))
	for _, e := range decodedGraphML.Graph.Edges {
		src := graphmlLabelByID[strings.TrimSpace(e.Source)]
		dst := graphmlLabelByID[strings.TrimSpace(e.Target)]
		graphmlEdges = append(graphmlEdges, normalizeUndirectedEdge(src, dst))
	}
	sort.Strings(graphmlEdges)
	if fmt.Sprintf("%v", jsonEdges) != fmt.Sprintf("%v", graphmlEdges) {
		t.Fatalf("edge identity mismatch json=%v graphml=%v", jsonEdges, graphmlEdges)
	}
}

func TestBuildTopologySkipsLoopLikeLLDPRelation(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "sw-core", SNMPEnabled: true},
	}
	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "192.168.1.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "sw-core",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			LldpNeighbors: []*LldpNeighbor{
				// Loop-like case: remote sysname points to the same device.
				{LocalIfIndex: 1, RemoteChassisID: "", RemotePortID: "Gi0/1", RemoteSysName: "SW-CORE"},
			},
		},
	}
	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}
	if len(topo.Links) != 0 {
		t.Fatalf("expected loop-like LLDP relation to be skipped, got %d links", len(topo.Links))
	}
}

func TestBuildTopologyMixedVendorUsesRemoteSysNameMatch(t *testing.T) {
	results := []scanner.Result{
		{IP: "10.0.0.1", MAC: "aa:aa:aa:aa:aa:aa", Hostname: "core-switch", SNMPEnabled: true},
		{IP: "10.0.0.2", MAC: "bb:bb:bb:bb:bb:bb", Hostname: "Edge-SW-01", SNMPEnabled: true},
	}
	snmp := map[string]*Device{
		"aa:aa:aa:aa:aa:aa": {
			IP:          "10.0.0.1",
			MAC:         "aa:aa:aa:aa:aa:aa",
			Hostname:    "core-switch",
			Type:        DeviceTypeSwitch,
			SNMPEnabled: true,
			LldpNeighbors: []*LldpNeighbor{
				// Mixed-vendor case: chassis ID is non-MAC string, fallback to RemoteSysName.
				{LocalIfIndex: 12, RemoteChassisID: "Port-Channel1", RemotePortID: "Eth1/12", RemoteSysName: " edge-sw-01 "},
			},
		},
	}
	topo, err := BuildTopology(results, snmp)
	if err != nil {
		t.Fatalf("BuildTopology error: %v", err)
	}
	if len(topo.Links) != 1 {
		t.Fatalf("expected one link from RemoteSysName fallback, got %d", len(topo.Links))
	}
	if topo.Links[0].SourceType != LinkSourceLLDP {
		t.Fatalf("expected LLDP source type, got %s", topo.Links[0].SourceType)
	}
}

func TestSaveGraphMLIncludesCompatibilityKeysAndEvidence(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"sw1": {IP: "192.168.1.1", Hostname: "sw1", Type: DeviceTypeSwitch},
			"sw2": {IP: "192.168.1.2", Hostname: "sw2", Type: DeviceTypeSwitch},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", Hostname: "sw1"},
				Target:     &Device{IP: "192.168.1.2", Hostname: "sw2"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
				Evidence:   "lldp_neighbor_match;local_if=1;remote_port=Gi0/1",
			},
		},
	}
	path := filepath.Join(t.TempDir(), "topology.graphml")
	if err := topo.SaveGraphML(path); err != nil {
		t.Fatalf("SaveGraphML error: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read graphml: %v", err)
	}
	got := string(raw)
	for _, expected := range []string{
		`<key id="label" for="node"`,
		`<key id="type" for="node"`,
		`<key id="source_type" for="edge"`,
		`<key id="confidence" for="edge"`,
		`<key id="evidence" for="edge"`,
		`<data key="evidence">lldp_neighbor_match;local_if=1;remote_port=Gi0/1</data>`,
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected GraphML to contain %q", expected)
		}
	}
}

func normalizeNodeIdentity(hostname, ip, mac string) string {
	hn := strings.ToLower(strings.TrimSpace(hostname))
	if hn != "" {
		return hn
	}
	trimIP := strings.ToLower(strings.TrimSpace(ip))
	if trimIP != "" {
		return trimIP
	}
	return strings.ToLower(strings.TrimSpace(mac))
}

func normalizeUndirectedEdge(a, b string) string {
	left := strings.ToLower(strings.TrimSpace(a))
	right := strings.ToLower(strings.TrimSpace(b))
	if left > right {
		left, right = right, left
	}
	return left + "<->" + right
}
