package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"network-scanner/internal/contracts"
)

func TestNew(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if p == nil {
		t.Fatal("New() returned nil")
	}
	info := p.Info()
	if info.Name != "JSONExporter" {
		t.Errorf("info.Name = %q, want %q", info.Name, "JSONExporter")
	}
	if info.Type != "exporter" {
		t.Errorf("info.Type = %q, want %q", info.Type, "exporter")
	}
}

func TestJSONExporter_Export(t *testing.T) {
	e := &JSONExporter{}
	results := []contracts.ScanResult{
		{
			IP: "10.0.0.1", Hostname: "server1", DeviceType: "server",
			GuessOS: "Linux", Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		},
	}
	tmp := t.TempDir()
	path := tmp + "/out.jsonl"
	if err := e.Export(results, path); err != nil {
		t.Fatalf("Export error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	var row map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &row); err != nil {
		t.Fatalf("JSON decode error = %v", err)
	}
	if row["ip"] != "10.0.0.1" {
		t.Errorf("ip = %v, want 10.0.0.1", row["ip"])
	}
}

func TestJSONExporter_Format(t *testing.T) {
	e := &JSONExporter{}
	if got := e.Format(); got != "jsonl" {
		t.Errorf("Format() = %q, want %q", got, "jsonl")
	}
}
