package main

import (
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
	if info.Name != "AlertReport" {
		t.Errorf("info.Name = %q, want %q", info.Name, "AlertReport")
	}
	if info.Type != "reporter" {
		t.Errorf("info.Type = %q, want %q", info.Type, "reporter")
	}
}

func TestAlertReport_GenerateText(t *testing.T) {
	r := &AlertReport{}
	results := []contracts.ScanResult{
		{
			IP: "10.0.0.1", Hostname: "server1",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
				{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
				{Port: 22, State: "closed", Protocol: "tcp", Service: "ssh"},
			},
		},
		{
			IP: "10.0.0.2", Hostname: "workstation1",
			Ports: []contracts.PortInfo{
				{Port: 3389, State: "open", Protocol: "tcp", Service: "rdp"},
			},
		},
		{
			IP: "10.0.0.3", Hostname: "db1",
			Ports: []contracts.PortInfo{
				{Port: 5432, State: "open", Protocol: "tcp", Service: "postgresql"},
			},
		},
	}

	data, err := r.Generate(results, "text")
	if err != nil {
		t.Fatalf("Generate error = %v", err)
	}
	report := string(data)
	if len(report) == 0 {
		t.Fatal("empty report")
	}
	if !strings.Contains(report, "10.0.0.1") {
		t.Error("report missing 10.0.0.1")
	}
	if !strings.Contains(report, "10.0.0.2") {
		t.Error("report missing 10.0.0.2")
	}
	// 10.0.0.3 has no critical ports, should NOT appear
	if strings.Contains(report, "10.0.0.3") {
		t.Error("report should not include 10.0.0.3 (no critical ports)")
	}
}

func TestAlertReport_Generate_UnsupportedFormat(t *testing.T) {
	r := &AlertReport{}
	_, err := r.Generate(nil, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}
