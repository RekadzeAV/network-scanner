package topology

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// --- Тесты JSON Schema Validation ---

func TestValidateJSONSchemaValid(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:01",
				Hostname: "router",
				Type:     DeviceTypeRouter,
			},
			"dev2": {
				IP:       "192.168.1.2",
				MAC:      "aa:bb:cc:dd:ee:02",
				Hostname: "switch",
				Type:     DeviceTypeSwitch,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "router"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "switch"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	check := topo.ValidateJSONSchema()

	if !check.Valid {
		t.Errorf("expected valid schema, got errors: %v", check.Errors)
	}
	if check.DeviceCount != 2 {
		t.Errorf("expected 2 devices, got %d", check.DeviceCount)
	}
	if check.LinkCount != 1 {
		t.Errorf("expected 1 link, got %d", check.LinkCount)
	}
	if len(check.DeviceTypes) != 2 {
		t.Errorf("expected 2 device types, got %d", len(check.DeviceTypes))
	}
	if len(check.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", check.Errors)
	}
}

func TestValidateJSONSchemaEmpty(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{},
		Links:   []Link{},
	}

	check := topo.ValidateJSONSchema()

	if check.Valid {
		t.Error("expected invalid for empty topology")
	}
	if !strings.Contains(strings.Join(check.Errors, " "), "empty") {
		t.Error("expected 'empty' error message")
	}
}

func TestValidateJSONSchemaNil(t *testing.T) {
	var topo *Topology

	check := topo.ValidateJSONSchema()

	if check.Valid {
		t.Error("expected invalid for nil topology")
	}
	if !strings.Contains(strings.Join(check.Errors, " "), "nil") {
		t.Error("expected 'nil' error message")
	}
}

func TestValidateJSONSchemaInvalidSource(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Type: DeviceTypeHost},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1"},
				Target:     &Device{IP: "192.168.1.2"},
				SourceType: "invalid_type",
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	check := topo.ValidateJSONSchema()

	if check.Valid {
		t.Error("expected invalid for bad source_type")
	}
	if !strings.Contains(strings.Join(check.Errors, " "), "invalid source_type") {
		t.Error("expected 'invalid source_type' error")
	}
}

func TestValidateJSONSchemaInvalidConfidence(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Type: DeviceTypeHost},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1"},
				Target:     &Device{IP: "192.168.1.2"},
				SourceType: LinkSourceLLDP,
				Confidence: "invalid_conf",
			},
		},
	}

	check := topo.ValidateJSONSchema()

	if check.Valid {
		t.Error("expected invalid for bad confidence")
	}
	if !strings.Contains(strings.Join(check.Errors, " "), "invalid confidence") {
		t.Error("expected 'invalid confidence' error")
	}
}

func TestValidateJSONSchemaMissingDeviceInLink(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Type: DeviceTypeHost},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1"},
				Target:     &Device{IP: "10.0.0.1"}, // Not in devices
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	check := topo.ValidateJSONSchema()

	if check.Valid {
		t.Error("expected invalid for link to missing device")
	}
	if !strings.Contains(strings.Join(check.Errors, " "), "not in devices") {
		t.Error("expected 'not in devices' error")
	}
}

// --- Тесты ToJSONSchema ---

func TestToJSONSchema(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Type: DeviceTypeHost},
		},
		Links: []Link{},
	}

	data, err := topo.ToJSONSchema()
	if err != nil {
		t.Fatalf("ToJSONSchema error: %v", err)
	}

	// Парсим JSON
	var schema JSONSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("Failed to unmarshal JSON schema: %v", err)
	}

	if schema.SchemaVersion != JSONSchemaVersion {
		t.Errorf("expected schema version %s, got %s", JSONSchemaVersion, schema.SchemaVersion)
	}
	if !schema.Validation.Valid {
		t.Errorf("expected valid validation, got errors: %v", schema.Validation.Errors)
	}
}

// --- Тесты GraphMLEquivalence ---

func TestGraphMLEquivalenceValid(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "host1", Type: DeviceTypeHost},
			"dev2": {IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host2", Type: DeviceTypeHost},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "host1"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host2"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	check := topo.GraphMLEquivalence()

	if !check.Match {
		t.Errorf("expected equivalence match, got errors: %v", check.Errors)
	}
	if check.JSONDevices != check.GraphMLDevices {
		t.Errorf("device count mismatch: json=%d, graphml=%d", check.JSONDevices, check.GraphMLDevices)
	}
	if check.JSONLinks != check.GraphMLLinks {
		t.Errorf("link count mismatch: json=%d, graphml=%d", check.JSONLinks, check.GraphMLLinks)
	}
}

func TestGraphMLEquivalenceNil(t *testing.T) {
	var topo *Topology

	check := topo.GraphMLEquivalence()

	if check.Match {
		t.Error("expected non-match for nil topology")
	}
	if !strings.Contains(strings.Join(check.Errors, " "), "nil") {
		t.Error("expected 'nil' error message")
	}
}

func TestGraphMLEquivalenceEmpty(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{},
		Links:   []Link{},
	}

	check := topo.GraphMLEquivalence()

	// Пустая топология должна быть эквивалентна (0=0)
	if !check.Match {
		t.Errorf("expected equivalence match for empty topology, got errors: %v", check.Errors)
	}
}

// --- Тесты SaveJSON + smoke ---

func TestSaveJSONSmoke(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Type: DeviceTypeHost},
		},
		Links: []Link{},
	}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "topology.json")

	err := topo.SaveJSON(filename)
	if err != nil {
		t.Fatalf("SaveJSON error: %v", err)
	}

	// Проверяем, что файл создан
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if len(content) == 0 {
		t.Error("saved JSON file is empty")
	}

	// Проверяем, что это валидный JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("Saved content is not valid JSON: %v", err)
	}
}

func TestSaveJSONInvalidTopology(t *testing.T) {
	// Топология с невалидным устройством
	topo := &Topology{
		Devices: map[string]*Device{
			"key": {IP: "", MAC: "", Hostname: ""}, // Нет стабильного ID
		},
		Links: []Link{},
	}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "topology.json")

	err := topo.SaveJSON(filename)
	if err == nil {
		t.Error("SaveJSON should fail for invalid topology")
	}
}

// --- Тесты SaveGraphML smoke ---

func TestSaveGraphMLSmoke(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Type: DeviceTypeHost},
		},
		Links: []Link{},
	}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "topology.graphml")

	err := topo.SaveGraphML(filename)
	if err != nil {
		t.Fatalf("SaveGraphML error: %v", err)
	}

	// Проверяем, что файл создан
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if len(content) == 0 {
		t.Error("saved GraphML file is empty")
	}

	// Проверяем, что это валидный XML (начинается с <?xml)
	if !strings.Contains(string(content), "<?xml") {
		t.Error("saved GraphML should start with XML declaration")
	}
	if !strings.Contains(string(content), "<graphml") {
		t.Error("saved GraphML should contain <graphml> root element")
	}
}

// --- Тесты MarshalJSONToJSON ---

func TestMarshalJSONToJSON(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Type: DeviceTypeHost},
		},
		Links: []Link{},
	}

	data, err := topo.MarshalJSONToJSON()
	if err != nil {
		t.Fatalf("MarshalJSONToJSON error: %v", err)
	}

	// Проверяем, что это валидный JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Marshaled JSON is not valid: %v", err)
	}
}

// --- Тесты SaveGraphMLToBytes ---

func TestSaveGraphMLToBytes(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Type: DeviceTypeHost},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1"},
				Target:     &Device{IP: "192.168.1.2"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	data, err := topo.SaveGraphMLToBytes()
	if err != nil {
		t.Fatalf("SaveGraphMLToBytes error: %v", err)
	}

	if len(data) == 0 {
		t.Error("SaveGraphMLToBytes returned empty data")
	}

	if !strings.Contains(string(data), "<?xml") {
		t.Error("SaveGraphMLToBytes should start with XML declaration")
	}
	if !strings.Contains(string(data), `<node id="`) {
		t.Error("SaveGraphMLToBytes should contain nodes")
	}
	if !strings.Contains(string(data), `<edge `) {
		t.Error("SaveGraphMLToBytes should contain edges")
	}
}

// --- Тесты parseGraphMLOrder ---

func TestParseGraphMLOrder(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<graphml xmlns="http://graphml.graphdrawing.org/xmlns">
  <graph id="network" edgedefault="undirected">
    <node id="mac_aa_bb_cc">
      <data key="label">test-host</data>
    </node>
    <node id="mac_dd_ee_ff">
      <data key="label">another-host</data>
    </node>
    <edge source="mac_aa_bb_cc" target="mac_dd_ee_ff">
      <data key="source_type">lldp</data>
    </edge>
  </graph>
</graphml>`)

	devices, links, errors := parseGraphMLOrder(xmlData)

	if len(errors) > 0 {
		t.Errorf("unexpected errors: %v", errors)
	}
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
	if links != 1 {
		t.Errorf("expected 1 link, got %d", links)
	}
	if !devices["mac_aa_bb_cc"] {
		t.Error("expected device 'mac_aa_bb_cc' not found")
	}
	if !devices["mac_dd_ee_ff"] {
		t.Error("expected device 'mac_dd_ee_ff' not found")
	}
}

// --- Тесты sortStrings ---

func TestSortStrings(t *testing.T) {
	tests := []struct {
		input  []string
		output []string
	}{
		{[]string{"c", "a", "b"}, []string{"a", "b", "c"}},
		{[]string{"z"}, []string{"z"}},
		{[]string{}, []string{}},
		{[]string{"b", "a"}, []string{"a", "b"}},
	}

	for _, tt := range tests {
		sortStrings(tt.input)
		if !reflect.DeepEqual(tt.input, tt.output) {
			t.Errorf("sortStrings(%v) = %v, want %v", tt.output, tt.input, tt.output)
		}
	}
}
