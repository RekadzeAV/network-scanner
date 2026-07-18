package topology

import (
	"strings"
	"testing"
)

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"AA:BB:CC:DD:EE:FF", "aa:bb:cc:dd:ee:ff"},
		{"aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:ff"},
		{"AA-BB-CC-DD-EE-FF", "aa:bb:cc:dd:ee:ff"},
		{"a:b:c:d:e:f", "0a:0b:0c:0d:0e:0f"},
		{"", ""},
		{"invalid", ""},
		{"AA:BB:CC", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeMAC(tt.input)
			if got != tt.want {
				t.Errorf("normalizeMAC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsBroadcastOrMulticast(t *testing.T) {
	tests := []struct {
		mac  string
		want bool
	}{
		{"ff:ff:ff:ff:ff:ff", true},
		{"01:00:5e:00:00:01", true},
		{"33:33:00:00:00:01", true},
		{"aa:bb:cc:dd:ee:ff", false},
		{"00:00:00:00:00:00", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.mac, func(t *testing.T) {
			got := isBroadcastOrMulticast(tt.mac)
			if got != tt.want {
				t.Errorf("isBroadcastOrMulticast(%q) = %v, want %v", tt.mac, got, tt.want)
			}
		})
	}
}

func TestIsZeroMAC(t *testing.T) {
	tests := []struct {
		mac  string
		want bool
	}{
		{"00:00:00:00:00:00", true},
		{"00:00:00:00:00:01", false},
		{"aa:bb:cc:dd:ee:ff", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.mac, func(t *testing.T) {
			got := isZeroMAC(tt.mac)
			if got != tt.want {
				t.Errorf("isZeroMAC(%q) = %v, want %v", tt.mac, got, tt.want)
			}
		})
	}
}

func TestParseMACBytes(t *testing.T) {
	tests := []struct {
		mac   string
		valid bool
	}{
		{"aa:bb:cc:dd:ee:ff", true},
		{"00:00:00:00:00:00", true},
		{"", false},
		{"invalid", false},
		{"aa:bb:cc", false},
	}

	for _, tt := range tests {
		t.Run(tt.mac, func(t *testing.T) {
			_, valid := parseMACBytes(tt.mac)
			if valid != tt.valid {
				t.Errorf("parseMACBytes(%q) valid = %v, want %v", tt.mac, valid, tt.valid)
			}
		})
	}
}

func TestConfidenceRank(t *testing.T) {
	tests := []struct {
		conf LinkConfidence
		want int
	}{
		{LinkConfidenceHigh, 3},
		{LinkConfidenceMedium, 2},
		{LinkConfidenceLow, 1},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.conf), func(t *testing.T) {
			got := confidenceRank(tt.conf)
			if got != tt.want {
				t.Errorf("confidenceRank(%q) = %d, want %d", tt.conf, got, tt.want)
			}
		})
	}
}

func TestNodeID(t *testing.T) {
	tests := []struct {
		name   string
		device *Device
		want   string
	}{
		{"MAC priority", &Device{MAC: "aa:bb:cc:dd:ee:ff"}, "mac_aa_bb_cc_dd_ee_ff"},
		{"IP fallback", &Device{IP: "192.168.1.1"}, "ip_192_168_1_1"},
		{"Hostname fallback", &Device{Hostname: "my-host"}, "hn_my-host"},
		{"Unknown", &Device{}, "unknown"},
		{"Nil device", nil, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nodeID(tt.device)
			if got != tt.want {
				t.Errorf("nodeID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeviceDisplayName(t *testing.T) {
	tests := []struct {
		name   string
		device *Device
		want   string
	}{
		{"Hostname priority", &Device{Hostname: "my-host"}, "my-host"},
		{"IP fallback", &Device{IP: "192.168.1.1"}, "192.168.1.1"},
		{"MAC fallback", &Device{MAC: "aa:bb:cc:dd:ee:ff"}, "aa:bb:cc:dd:ee:ff"},
		{"Unknown", &Device{}, "unknown"},
		{"Nil device", nil, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deviceDisplayName(tt.device)
			if got != tt.want {
				t.Errorf("deviceDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPortLabel(t *testing.T) {
	tests := []struct {
		name string
		port *Port
		want string
	}{
		{"By name", &Port{Name: "Gi0/1"}, "Gi0/1"},
		{"By index", &Port{Index: 5}, "if5"},
		{"Empty", &Port{}, ""},
		{"Nil port", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := portLabel(tt.port)
			if got != tt.want {
				t.Errorf("portLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLinkKey(t *testing.T) {
	// linkKey should be symmetric (order-independent)
	key1 := linkKey("node-a", "port1", "node-b", "port2")
	key2 := linkKey("node-b", "port2", "node-a", "port1")
	if key1 != key2 {
		t.Errorf("linkKey should be symmetric: key1=%q, key2=%q", key1, key2)
	}
}

func TestClassifyFromScannerResult(t *testing.T) {
	tests := []struct {
		input string
		want  DeviceType
	}{
		{"Router", DeviceTypeRouter},
		{"router/switch", DeviceTypeRouter},
		{"Switch", DeviceTypeSwitch},
		{"network device", DeviceTypeSwitch},
		{"Server", DeviceTypeHost},
		{"computer", DeviceTypeHost},
		{"host", DeviceTypeHost},
		{"unknown device", DeviceTypeUnknown},
		{"", DeviceTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := classifyFromScannerResult(tt.input)
			if got != tt.want {
				t.Errorf("classifyFromScannerResult(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFindNeighbor(t *testing.T) {
	byMAC := map[string]*Device{
		"aa:bb:cc:dd:ee:ff": {MAC: "aa:bb:cc:dd:ee:ff", IP: "192.168.1.1"},
	}
	byHostname := map[string]*Device{
		"my-host": {Hostname: "my-host", IP: "192.168.1.2"},
	}

	tests := []struct {
		name        string
		neighbor    *LldpNeighbor
		byMAC       map[string]*Device
		byHostname  map[string]*Device
		expectMatch bool
	}{
		{"Match by MAC", &LldpNeighbor{RemoteChassisID: "aa:bb:cc:dd:ee:ff"}, byMAC, byHostname, true},
		{"Match by hostname", &LldpNeighbor{RemoteSysName: "my-host"}, byMAC, byHostname, true},
		{"No match", &LldpNeighbor{RemoteChassisID: "unknown"}, byMAC, byHostname, false},
		{"Nil neighbor", nil, byMAC, byHostname, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findNeighbor(tt.byMAC, tt.byHostname, tt.neighbor)
			if tt.expectMatch && result == nil {
				t.Error("expected neighbor match, got nil")
			}
			if !tt.expectMatch && result != nil {
				t.Error("expected no match, got non-nil")
			}
		})
	}
}

func TestEnsurePort(t *testing.T) {
	dev := &Device{
		Ports: []Port{
			{Index: 1, Name: "Gi0/1"},
		},
	}

	// Should find existing port by index
	p1 := ensurePort(dev, 1, "")
	if p1 == nil || p1.Name != "Gi0/1" {
		t.Errorf("ensurePort should find existing port by index")
	}

	// Should find existing port by name
	p2 := ensurePort(dev, 0, "Gi0/1")
	if p2 == nil || p2.Name != "Gi0/1" {
		t.Errorf("ensurePort should find existing port by name")
	}

	// Should create new port
	p3 := ensurePort(dev, 5, "")
	if p3 == nil || p3.Name != "if5" {
		t.Errorf("ensurePort should create new port with if%d name, got %q", p3.Index, p3.Name)
	}

	// Should create new port with explicit name
	p4 := ensurePort(dev, 0, "Eth0/1")
	if p4 == nil || p4.Name != "Eth0/1" {
		t.Errorf("ensurePort should create new port with explicit name, got %q", p4.Name)
	}
}

func TestValidateNilTopology(t *testing.T) {
	var topo *Topology
	if err := topo.Validate(); err == nil {
		t.Error("Validate() should return error for nil topology")
	}

	topo = &Topology{}
	if err := topo.Validate(); err == nil {
		t.Error("Validate() should return error for nil devices map")
	}
}

func TestValidateNilDevice(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"key": nil,
		},
	}
	if err := topo.Validate(); err == nil {
		t.Error("Validate() should return error for nil device")
	}
}

func TestValidateNilLinkEndpoint(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1"},
		},
		Links: []Link{
			{
				Source:     nil,
				Target:     &Device{IP: "192.168.1.1"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}
	if err := topo.Validate(); err == nil {
		t.Error("Validate() should return error for nil link endpoint")
	}
}

func TestValidateEmptySourceType(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1"},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1"},
				Target:     &Device{IP: "192.168.1.2"},
				SourceType: "",
				Confidence: LinkConfidenceHigh,
			},
		},
	}
	if err := topo.Validate(); err == nil {
		t.Error("Validate() should return error for empty source_type")
	}
}

func TestValidateEmptyConfidence(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1"},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1"},
				Target:     &Device{IP: "192.168.1.2"},
				SourceType: LinkSourceLLDP,
				Confidence: "",
			},
		},
	}
	if err := topo.Validate(); err == nil {
		t.Error("Validate() should return error for empty confidence")
	}
}

func TestValidateUnknownDeviceID(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"key": {IP: "", MAC: "", Hostname: ""},
		},
	}
	if err := topo.Validate(); err == nil {
		t.Error("Validate() should return error for device with no stable identifier")
	}
}

func TestToDOTWithLinks(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Hostname: "host1"},
			"dev2": {IP: "192.168.1.2", Hostname: "host2"},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", Hostname: "host1"},
				Target:     &Device{IP: "192.168.1.2", Hostname: "host2"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	var buf strings.Builder
	if err := topo.ToDOT(&buf); err != nil {
		t.Fatalf("ToDOT error: %v", err)
	}

	dot := buf.String()
	if !strings.Contains(dot, "graph network") {
		t.Error("ToDOT should contain 'graph network'")
	}
	if !strings.Contains(dot, "host1") {
		t.Error("ToDOT should contain hostname")
	}
	// Edge label may be empty if ports are not set
	if !strings.Contains(dot, "host1") || !strings.Contains(dot, "host2") {
		t.Error("ToDOT should contain both hostnames in nodes")
	}
}

func TestNormalizedKey(t *testing.T) {
	tests := []struct {
		mac  string
		ip   string
		want string
	}{
		{"aa:bb:cc:dd:ee:ff", "192.168.1.1", "aa:bb:cc:dd:ee:ff"},
		{"", "192.168.1.1", "192.168.1.1"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.mac+tt.ip, func(t *testing.T) {
			got := normalizedKey(tt.mac, tt.ip)
			if got != tt.want {
				t.Errorf("normalizedKey(%q, %q) = %q, want %q", tt.mac, tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsPartialDevice(t *testing.T) {
	opts := BuildOptions{
		PartialSNMPKeys: map[string]struct{}{
			"ip:192.168.1.1": {},
		},
	}

	tests := []struct {
		name   string
		device *Device
		want   bool
	}{
		{"Match by IP", &Device{IP: "192.168.1.1"}, true},
		{"Match by MAC", &Device{MAC: "aa:bb:cc:dd:ee:ff"}, false},
		{"Match by hostname", &Device{Hostname: "my-host"}, false},
		{"No match", &Device{IP: "10.0.0.1"}, false},
		{"Nil device", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPartialDevice(tt.device, opts)
			if got != tt.want {
				t.Errorf("isPartialDevice(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}

	// Test with empty PartialSNMPKeys
	optsEmpty := BuildOptions{}
	if isPartialDevice(&Device{IP: "192.168.1.1"}, optsEmpty) {
		t.Error("isPartialDevice should return false with empty PartialSNMPKeys")
	}
}

func TestDeviceKeys(t *testing.T) {
	dev := &Device{
		IP:       "192.168.1.1",
		MAC:      "aa:bb:cc:dd:ee:ff",
		Hostname: "my-host",
	}

	keys := deviceKeys(dev)
	if len(keys) != 3 {
		t.Errorf("deviceKeys() returned %d keys, want 3", len(keys))
	}

	expectedPrefixes := []string{"ip:192.168.1.1", "mac:aa:bb:cc:dd:ee:ff", "hostname:my-host"}
	for _, expected := range expectedPrefixes {
		found := false
		for _, key := range keys {
			if key == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("deviceKeys() missing key %q", expected)
		}
	}

	// Test nil device
	if deviceKeys(nil) != nil {
		t.Error("deviceKeys(nil) should return nil")
	}
}

func TestMaybeLowerConfidence(t *testing.T) {
	opts := BuildOptions{
		PartialSNMPKeys: map[string]struct{}{
			"ip:192.168.1.1": {},
		},
	}

	tests := []struct {
		name     string
		base     LinkConfidence
		a        *Device
		b        *Device
		opts     BuildOptions
		expected LinkConfidence
	}{
		{"High to Medium", LinkConfidenceHigh, &Device{IP: "192.168.1.1"}, &Device{IP: "192.168.1.2"}, opts, LinkConfidenceMedium},
		{"Medium to Low", LinkConfidenceMedium, &Device{IP: "192.168.1.1"}, &Device{IP: "192.168.1.2"}, opts, LinkConfidenceLow},
		{"Low stays Low", LinkConfidenceLow, &Device{IP: "192.168.1.1"}, &Device{IP: "192.168.1.2"}, opts, LinkConfidenceLow},
		{"High stays High (no partial)", LinkConfidenceHigh, &Device{IP: "10.0.0.1"}, &Device{IP: "10.0.0.2"}, opts, LinkConfidenceHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maybeLowerConfidence(tt.base, tt.a, tt.b, tt.opts)
			if got != tt.expected {
				t.Errorf("maybeLowerConfidence(%q) = %q, want %q", tt.base, got, tt.expected)
			}
		})
	}
}
