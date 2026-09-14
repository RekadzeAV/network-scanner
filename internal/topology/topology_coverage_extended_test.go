package topology

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"network-scanner/internal/contracts"
)

// ============================================================================
// ValidateJSONSchema — ветки с nil devices, empty type, invalid values
// ============================================================================

func TestValidateJSONSchema_NilDeviceInMap(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"key": nil,
		},
		Links: []Link{},
	}
	check := topo.ValidateJSONSchema()
	if check.Valid {
		t.Error("expected invalid for nil device in map")
	}
	if !containsStr(check.Errors, "nil device in devices map") {
		t.Errorf("expected 'nil device in devices map' in errors, got %v", check.Errors)
	}
}

func TestValidateJSONSchema_EmptyDeviceType(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "test",
				Type:     "", // пустой тип
			},
		},
		Links: []Link{},
	}
	check := topo.ValidateJSONSchema()
	if check.Valid {
		t.Error("expected invalid for empty device type")
	}
	found := false
	for _, e := range check.Errors {
		if strings.Contains(e, "empty type") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'empty type' in errors, got %v", check.Errors)
	}
}

func TestValidateJSONSchema_NilSourceInLink(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
		},
		Links: []Link{
			{
				Source: nil, // nil source
				Target: &Device{IP: "192.168.1.2", MAC: "11:22:33:44:55:66", Hostname: "host1", Type: DeviceTypeHost},
			},
		},
	}
	check := topo.ValidateJSONSchema()
	if check.Valid {
		t.Error("expected invalid for nil source in link")
	}
	found := false
	for _, e := range check.Errors {
		if strings.Contains(e, "link[0] has nil endpoint") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'nil endpoint' in errors, got %v", check.Errors)
	}
}

func TestValidateJSONSchema_SourceNotInDevices(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
		},
		Links: []Link{
			{
				Source: &Device{IP: "10.0.0.1", MAC: "00:11:22:33:44:55", Hostname: "unknown", Type: DeviceTypeHost}, // не в devices map
				Target: &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff", Hostname: "switch1", Type: DeviceTypeSwitch},
			},
		},
	}
	check := topo.ValidateJSONSchema()
	if check.Valid {
		t.Error("expected invalid for source not in devices map")
	}
	found := false
	for _, e := range check.Errors {
		if strings.Contains(e, "not in devices") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'not in devices' in errors, got %v", check.Errors)
	}
}

func TestValidateJSONSchema_InvalidSourceType(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
			"mac_11_22_33_44_55_66": {
				IP:       "192.168.1.2",
				MAC:      "11:22:33:44:55:66",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff", Hostname: "switch1", Type: DeviceTypeSwitch},
				Target:     &Device{IP: "192.168.1.2", MAC: "11:22:33:44:55:66", Hostname: "host1", Type: DeviceTypeHost},
				SourceType: "invalid_source_type",
				Confidence: LinkConfidenceHigh,
			},
		},
	}
	check := topo.ValidateJSONSchema()
	if check.Valid {
		t.Error("expected invalid for invalid source_type")
	}
	found := false
	for _, e := range check.Errors {
		if strings.Contains(e, "invalid source_type") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'invalid source_type' in errors, got %v", check.Errors)
	}
}

func TestValidateJSONSchema_InvalidConfidence(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
			"mac_11_22_33_44_55_66": {
				IP:       "192.168.1.2",
				MAC:      "11:22:33:44:55:66",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff", Hostname: "switch1", Type: DeviceTypeSwitch},
				Target:     &Device{IP: "192.168.1.2", MAC: "11:22:33:44:55:66", Hostname: "host1", Type: DeviceTypeHost},
				SourceType: LinkSourceLLDP,
				Confidence: "invalid_confidence",
			},
		},
	}
	check := topo.ValidateJSONSchema()
	if check.Valid {
		t.Error("expected invalid for invalid confidence")
	}
	found := false
	for _, e := range check.Errors {
		if strings.Contains(e, "invalid confidence") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'invalid confidence' in errors, got %v", check.Errors)
	}
}

func TestValidateJSONSchema_DeviceTypesCollected(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
			"mac_11_22_33_44_55_66": {
				IP:       "192.168.1.2",
				MAC:      "11:22:33:44:55:66",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
			"mac_22_33_44_55_66_77": {
				IP:       "192.168.1.3",
				MAC:      "22:33:44:55:66:77",
				Hostname: "router1",
				Type:     DeviceTypeRouter,
			},
		},
		Links: []Link{},
	}
	check := topo.ValidateJSONSchema()
	if len(check.DeviceTypes) != 3 {
		t.Errorf("expected 3 device types, got %d: %v", len(check.DeviceTypes), check.DeviceTypes)
	}
}

func TestValidateJSONSchema_SourceTypesCollected(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
			"mac_11_22_33_44_55_66": {
				IP:       "192.168.1.2",
				MAC:      "11:22:33:44:55:66",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff", Hostname: "switch1", Type: DeviceTypeSwitch},
				Target:     &Device{IP: "192.168.1.2", MAC: "11:22:33:44:55:66", Hostname: "host1", Type: DeviceTypeHost},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff", Hostname: "switch1", Type: DeviceTypeSwitch},
				Target:     &Device{IP: "192.168.1.2", MAC: "11:22:33:44:55:66", Hostname: "host1", Type: DeviceTypeHost},
				SourceType: LinkSourceFDB,
				Confidence: LinkConfidenceMedium,
			},
		},
	}
	check := topo.ValidateJSONSchema()
	if len(check.SourceTypes) != 2 {
		t.Errorf("expected 2 source types, got %d: %v", len(check.SourceTypes), check.SourceTypes)
	}
	if len(check.ConfidenceLvl) != 2 {
		t.Errorf("expected 2 confidence levels, got %d: %v", len(check.ConfidenceLvl), check.ConfidenceLvl)
	}
}

// ============================================================================
// ToJSONSchema — marshal error
// ============================================================================

func TestToJSONSchema_Success(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
		},
		Links: []Link{},
	}

	data, err := topo.ToJSONSchema()
	if err != nil {
		t.Fatalf("ToJSONSchema() error = %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"$schema_version"`) {
		t.Error("expected $schema_version in output")
	}
	if !strings.Contains(jsonStr, "switch1") {
		t.Error("expected 'switch1' in output")
	}
}

// ============================================================================
// MarshalJSONToJSON — validation error
// ============================================================================

func TestMarshalJSONToJSON_ValidationFailure(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"key": nil, // nil device
		},
		Links: []Link{},
	}

	data, err := topo.MarshalJSONToJSON()
	if err == nil {
		t.Fatal("expected error for invalid topology")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected 'validation failed' in error, got %v", err)
	}
	if data != nil {
		t.Error("expected nil data on error")
	}
}

func TestMarshalJSONToJSON_Success(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
		},
		Links: []Link{},
	}

	data, err := topo.MarshalJSONToJSON()
	if err != nil {
		t.Fatalf("MarshalJSONToJSON() error = %v", err)
	}
	if !strings.Contains(string(data), "switch1") {
		t.Error("expected 'switch1' in JSON output")
	}
}

// ============================================================================
// GraphMLEquivalence — все ветки
// ============================================================================

func TestGraphMLEquivalence_NilTopology(t *testing.T) {
	var topo *Topology
	check := topo.GraphMLEquivalence()
	if check.Match {
		t.Error("expected Match=false for nil topology")
	}
	if !containsStr(check.Errors, "topology is nil") {
		t.Errorf("expected 'topology is nil' in errors, got %v", check.Errors)
	}
}

func TestGraphMLEquivalence_Match(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
			"mac_11_22_33_44_55_66": {
				IP:       "192.168.1.2",
				MAC:      "11:22:33:44:55:66",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff", Hostname: "switch1", Type: DeviceTypeSwitch},
				Target:     &Device{IP: "192.168.1.2", MAC: "11:22:33:44:55:66", Hostname: "host1", Type: DeviceTypeHost},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	check := topo.GraphMLEquivalence()
	if !check.Match {
		t.Errorf("expected Match=true, errors: %v", check.Errors)
	}
	if check.JSONDevices != check.GraphMLDevices {
		t.Errorf("device count mismatch: json=%d, graphml=%d", check.JSONDevices, check.GraphMLDevices)
	}
	if check.JSONLinks != check.GraphMLLinks {
		t.Errorf("link count mismatch: json=%d, graphml=%d", check.JSONLinks, check.GraphMLLinks)
	}
}

func TestGraphMLEquivalence_EmptyTopology(t *testing.T) {
	topo := &Topology{
		Devices: make(map[string]*Device),
		Links:   []Link{},
	}

	check := topo.GraphMLEquivalence()
	if !check.Match {
		t.Errorf("expected Match=true for empty topology, errors: %v", check.Errors)
	}
	if check.JSONDevices != 0 {
		t.Errorf("expected 0 JSON devices, got %d", check.JSONDevices)
	}
}

// ============================================================================
// sortStrings — сортировка
// ============================================================================

func TestSortStrings_Empty(t *testing.T) {
	s := []string{}
	sortStrings(s)
	if len(s) != 0 {
		t.Errorf("expected empty slice, got %v", s)
	}
}

func TestSortStrings_Single(t *testing.T) {
	s := []string{"b"}
	sortStrings(s)
	if len(s) != 1 || s[0] != "b" {
		t.Errorf("expected ['b'], got %v", s)
	}
}

func TestSortStrings_AlreadySorted(t *testing.T) {
	s := []string{"a", "b", "c"}
	sortStrings(s)
	expected := []string{"a", "b", "c"}
	for i := range s {
		if s[i] != expected[i] {
			t.Errorf("expected %v, got %v", expected, s)
			return
		}
	}
}

func TestSortStrings_Reversed(t *testing.T) {
	s := []string{"c", "b", "a"}
	sortStrings(s)
	expected := []string{"a", "b", "c"}
	for i := range s {
		if s[i] != expected[i] {
			t.Errorf("expected %v, got %v", expected, s)
			return
		}
	}
}

func TestSortStrings_Unsorted(t *testing.T) {
	s := []string{"zebra", "apple", "mango"}
	sortStrings(s)
	expected := []string{"apple", "mango", "zebra"}
	for i := range s {
		if s[i] != expected[i] {
			t.Errorf("expected %v, got %v", expected, s)
			return
		}
	}
}

// ============================================================================
// parseGraphMLOrder — парсинг GraphML
// ============================================================================

func TestParseGraphMLOrder_WithNodes(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<graphml xmlns="http://graphml.graphdrawing.org/xmlns">
  <graph edgedefault="undirected">
    <node id="mac_aa_bb_cc_dd_ee_ff">
      <data key="hostname">switch1</data>
    </node>
    <node id="mac_11_22_33_44_55_66">
      <data key="hostname">host1</data>
    </node>
    <edge source="mac_aa_bb_cc_dd_ee_ff" target="mac_11_22_33_44_55_66"/>
    <edge source="mac_11_22_33_44_55_66" target="mac_aa_bb_cc_dd_ee_ff"/>
  </graph>
</graphml>`

	devices, links, errors := parseGraphMLOrder([]byte(xml))
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d: %v", len(devices), devices)
	}
	if links != 2 {
		t.Errorf("expected 2 links, got %d", links)
	}
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestParseGraphMLOrder_Empty(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<graphml xmlns="http://graphml.graphdrawing.org/xmlns">
  <graph edgedefault="undirected">
  </graph>
</graphml>`

	devices, links, errors := parseGraphMLOrder([]byte(xml))
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
	if links != 0 {
		t.Errorf("expected 0 links, got %d", links)
	}
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestParseGraphMLOrder_EmptyBytes(t *testing.T) {
	devices, links, errors := parseGraphMLOrder([]byte{})
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
	if links != 0 {
		t.Errorf("expected 0 links, got %d", links)
	}
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// Export — все форматы
// ============================================================================

func TestExport_JSON_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topology.json")

	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
		},
		Links: []*contracts.Link{},
	}

	svc := NewService()
	err := svc.Export(contract, "json", path)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "switch1") {
		t.Error("expected 'switch1' in JSON output")
	}
}

func TestExport_GraphML_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topology.graphml")

	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
		},
		Links: []*contracts.Link{},
	}

	svc := NewService()
	err := svc.Export(contract, "graphml", path)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "switch1") {
		t.Error("expected 'switch1' in GraphML output")
	}
}

func TestExport_DOT_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topology.dot")

	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
		},
		Links: []*contracts.Link{},
	}

	svc := NewService()
	err := svc.Export(contract, "dot", path)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "switch1") {
		t.Error("expected 'switch1' in DOT output")
	}
	if !strings.Contains(string(data), "graph network") {
		t.Error("expected 'graph network' in DOT output")
	}
}

func TestExport_Text_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topology.txt")

	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
		},
		Links: []*contracts.Link{},
	}

	svc := NewService()
	err := svc.Export(contract, "text", path)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "switch1") {
		t.Error("expected 'switch1' in text output")
	}
}

func TestExport_XML_Format(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topology.xml")

	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
		},
		Links: []*contracts.Link{},
	}

	svc := NewService()
	err := svc.Export(contract, "xml", path)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "switch1") {
		t.Error("expected 'switch1' in XML output")
	}
}

func TestExport_TXT_Format(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topology.txt")

	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
		},
		Links: []*contracts.Link{},
	}

	svc := NewService()
	err := svc.Export(contract, "txt", path)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "switch1") {
		t.Error("expected 'switch1' in text output")
	}
}

func TestExport_NilTopology(t *testing.T) {
	var contract *contracts.Topology

	svc := NewService()
	err := svc.Export(contract, "json", "/tmp/test.json")
	if err == nil {
		t.Fatal("expected error for nil topology")
	}
	if !strings.Contains(err.Error(), "topology is nil") {
		t.Errorf("expected 'topology is nil' in error, got %v", err)
	}
}

func TestExport_EmptyPath(t *testing.T) {
	contract := &contracts.Topology{
		Devices: []*contracts.Device{},
		Links:   []*contracts.Link{},
	}

	svc := NewService()
	err := svc.Export(contract, "json", "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !strings.Contains(err.Error(), "export path is empty") {
		t.Errorf("expected 'export path is empty' in error, got %v", err)
	}
}

func TestExport_UnsupportedFormat(t *testing.T) {
	contract := &contracts.Topology{
		Devices: []*contracts.Device{},
		Links:   []*contracts.Link{},
	}

	svc := NewService()
	err := svc.Export(contract, "csv", "/tmp/test.csv")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported export format") {
		t.Errorf("expected 'unsupported export format' in error, got %v", err)
	}
}

func TestExport_CreateFileFailure(t *testing.T) {
	contract := &contracts.Topology{
		Devices: []*contracts.Device{},
		Links:   []*contracts.Link{},
	}

	svc := NewService()
	err := svc.Export(contract, "dot", "/nonexistent/dir/topology.dot")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
	if !strings.Contains(err.Error(), "create file") {
		t.Errorf("expected 'create file' in error, got %v", err)
	}
}

// ============================================================================
// Build — создание топологии из результатов
// ============================================================================

func TestBuild_EmptyResults(t *testing.T) {
	svc := NewService()
	topo, err := svc.Build(context.Background(), []contracts.ScanResult{}, contracts.TopologyOptions{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if topo == nil {
		t.Fatal("expected non-nil topology for empty results")
	}
	if len(topo.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(topo.Devices))
	}
	if len(topo.Links) != 0 {
		t.Errorf("expected 0 links, got %d", len(topo.Links))
	}
}

func TestBuild_SingleResult(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP:         "192.168.1.1",
			Hostname:   "test-host",
			MAC:        "aa:bb:cc:dd:ee:ff",
			DeviceType: "host",
		},
	}

	topo, err := svc.Build(context.Background(), results, contracts.TopologyOptions{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if topo == nil {
		t.Fatal("expected non-nil topology")
	}
	if len(topo.Devices) == 0 {
		t.Error("expected at least 1 device")
	}
}

func TestBuild_ContextSuccess(t *testing.T) {
	ctx := context.Background()

	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP:         "192.168.1.1",
			Hostname:   "test-host",
			MAC:        "aa:bb:cc:dd:ee:ff",
			DeviceType: "host",
		},
	}

	topo, err := svc.Build(ctx, results, contracts.TopologyOptions{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if topo == nil {
		t.Fatal("expected non-nil topology")
	}
}

// ============================================================================
// convertFromContractTopology — конвертация
// ============================================================================

func TestConvertFromContractTopology_Nil(t *testing.T) {
	result := convertFromContractTopology(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestConvertFromContractTopology_WithDevice(t *testing.T) {
	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
			{IP: "192.168.1.2", Hostname: "host1", Type: "host"},
			{IP: "192.168.1.3", Hostname: "router1", Type: "router"},
		},
		Links: []*contracts.Link{},
	}

	result := convertFromContractTopology(contract)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Devices) != 3 {
		t.Errorf("expected 3 devices, got %d", len(result.Devices))
	}
}

func TestConvertFromContractTopology_WithDeviceTypeMapping(t *testing.T) {
	tests := []struct {
		contractType string
		wantInternal DeviceType
	}{
		{"router", DeviceTypeRouter},
		{"Router", DeviceTypeRouter},
		{"switch", DeviceTypeSwitch},
		{"network", DeviceTypeSwitch},
		{"host", DeviceTypeHost},
		{"server", DeviceTypeHost},
		{"computer", DeviceTypeHost},
		{"unknown", DeviceTypeUnknown},
		{"", DeviceTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.contractType, func(t *testing.T) {
			contract := &contracts.Topology{
				Devices: []*contracts.Device{
					{IP: "192.168.1.1", Type: tt.contractType},
				},
				Links: []*contracts.Link{},
			}

			result := convertFromContractTopology(contract)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			for _, d := range result.Devices {
				if d.Type != tt.wantInternal {
					t.Errorf("contract type %q -> internal type %q, want %q", tt.contractType, d.Type, tt.wantInternal)
				}
			}
		})
	}
}

func TestConvertFromContractTopology_LinkSourceTypeMapping(t *testing.T) {
	tests := []struct {
		contractType string
		wantInternal LinkSourceType
	}{
		{"lldp", LinkSourceLLDP},
		{"LLDP", LinkSourceLLDP},
		{"fdb", LinkSourceFDB},
		{"FDB", LinkSourceFDB},
		{"inferred", LinkSourceInferred},
		{"", LinkSourceInferred},
	}

	for _, tt := range tests {
		t.Run(tt.contractType, func(t *testing.T) {
			contract := &contracts.Topology{
				Devices: []*contracts.Device{
					{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
					{IP: "192.168.1.2", Hostname: "host1", Type: "host"},
				},
				Links: []*contracts.Link{
					{
						Source:     &contracts.Device{IP: "192.168.1.1"},
						Target:     &contracts.Device{IP: "192.168.1.2"},
						SourceType: tt.contractType,
					},
				},
			}

			result := convertFromContractTopology(contract)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if len(result.Links) != 1 {
				t.Fatalf("expected 1 link, got %d", len(result.Links))
			}
			if result.Links[0].SourceType != tt.wantInternal {
				t.Errorf("contract source type %q -> internal type %q, want %q", tt.contractType, result.Links[0].SourceType, tt.wantInternal)
			}
		})
	}
}

func TestConvertFromContractTopology_LinkConfidenceMapping(t *testing.T) {
	tests := []struct {
		contractConf string
		wantInternal LinkConfidence
	}{
		{"high", LinkConfidenceHigh},
		{"HIGH", LinkConfidenceHigh},
		{"medium", LinkConfidenceMedium},
		{"MEDIUM", LinkConfidenceMedium},
		{"low", LinkConfidenceLow},
		{"", LinkConfidenceLow},
	}

	for _, tt := range tests {
		t.Run(tt.contractConf, func(t *testing.T) {
			contract := &contracts.Topology{
				Devices: []*contracts.Device{
					{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
					{IP: "192.168.1.2", Hostname: "host1", Type: "host"},
				},
				Links: []*contracts.Link{
					{
						Source:     &contracts.Device{IP: "192.168.1.1"},
						Target:     &contracts.Device{IP: "192.168.1.2"},
						Confidence: tt.contractConf,
					},
				},
			}

			result := convertFromContractTopology(contract)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if len(result.Links) != 1 {
				t.Fatalf("expected 1 link, got %d", len(result.Links))
			}
			if result.Links[0].Confidence != tt.wantInternal {
				t.Errorf("contract confidence %q -> internal %q, want %q", tt.contractConf, result.Links[0].Confidence, tt.wantInternal)
			}
		})
	}
}

func TestConvertFromContractTopology_WithPort(t *testing.T) {
	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
			{IP: "192.168.1.2", Hostname: "host1", Type: "host"},
		},
		Links: []*contracts.Link{
			{
				Source:     &contracts.Device{IP: "192.168.1.1"},
				SourcePort: "Gi0/1",
				Target:     &contracts.Device{IP: "192.168.1.2"},
				TargetPort: "",
			},
		},
	}

	result := convertFromContractTopology(contract)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(result.Links))
	}
	if result.Links[0].SourcePort == nil || result.Links[0].SourcePort.Name != "Gi0/1" {
		t.Error("expected source port name 'Gi0/1'")
	}
}

func TestConvertFromContractTopology_NilLinkSkipped(t *testing.T) {
	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
		},
		Links: []*contracts.Link{
			nil, // должен быть пропущен
		},
	}

	result := convertFromContractTopology(contract)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Links) != 0 {
		t.Errorf("expected 0 links (nil skipped), got %d", len(result.Links))
	}
}

func TestConvertFromContractTopology_NilSourceTargetSkipped(t *testing.T) {
	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "192.168.1.1", Hostname: "switch1", Type: "switch"},
		},
		Links: []*contracts.Link{
			{
				Source: nil, // должен быть пропущен
			},
		},
	}

	result := convertFromContractTopology(contract)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Links) != 0 {
		t.Errorf("expected 0 links (nil source skipped), got %d", len(result.Links))
	}
}

func TestConvertFromContractTopology_DeviceKeyFallback(t *testing.T) {
	// Устройство без IP, hostname, MAC — ключ должен быть "unknown"
	contract := &contracts.Topology{
		Devices: []*contracts.Device{
			{IP: "", Hostname: "", MAC: ""},
		},
		Links: []*contracts.Link{},
	}

	result := convertFromContractTopology(contract)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result.Devices["unknown"]; !ok {
		t.Errorf("expected device with key 'unknown', got keys: %v", func() []string {
			keys := make([]string, 0, len(result.Devices))
			for k := range result.Devices {
				keys = append(keys, k)
			}
			return keys
		}())
	}
}

// ============================================================================
// convertToContractTopology — конвертация
// ============================================================================

func TestConvertToContractTopology_Nil(t *testing.T) {
	result := convertToContractTopology(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestConvertToContractTopology_WithData(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"mac_aa_bb_cc_dd_ee_ff": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", Hostname: "switch1"},
				Target:     &Device{IP: "192.168.1.2", Hostname: "host1"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	result := convertToContractTopology(topo)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(result.Devices))
	}
	if len(result.Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(result.Links))
	}
	if result.Links[0].Confidence != string(LinkConfidenceHigh) {
		t.Errorf("expected confidence %q, got %q", LinkConfidenceHigh, result.Links[0].Confidence)
	}
}

// ============================================================================
// convertToDevice — конвертация устройства
// ============================================================================

func TestConvertToDevice_Nil(t *testing.T) {
	result := convertToDevice(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestConvertToDevice_WithData(t *testing.T) {
	dev := &Device{
		IP:       "192.168.1.1",
		Hostname: "switch1",
		Type:     DeviceTypeSwitch,
	}

	result := convertToDevice(dev)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", result.IP)
	}
	if result.Hostname != "switch1" {
		t.Errorf("expected Hostname 'switch1', got %q", result.Hostname)
	}
	if result.Type != string(DeviceTypeSwitch) {
		t.Errorf("expected Type %q, got %q", DeviceTypeSwitch, result.Type)
	}
}

// ============================================================================
// Вспомогательная функция
// ============================================================================

func containsStr(slice []string, s string) bool {
	for _, item := range slice {
		if strings.Contains(item, s) {
			return true
		}
	}
	return false
}
