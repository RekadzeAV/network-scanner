package main

import (
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
	if info.Name != "OnlyUp" {
		t.Errorf("info.Name = %q, want %q", info.Name, "OnlyUp")
	}
	if info.Type != "filter" {
		t.Errorf("info.Type = %q, want %q", info.Type, "filter")
	}
}

func TestOnlyUpFilter_Filter(t *testing.T) {
	f := &OnlyUpFilter{}
	results := []contracts.ScanResult{
		{
			IP: "10.0.0.1", Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
				{Port: 443, State: "closed", Protocol: "tcp", Service: "https"},
			},
		},
		{
			IP: "10.0.0.2", Ports: []contracts.PortInfo{
				{Port: 22, State: "closed", Protocol: "tcp", Service: "ssh"},
				{Port: 8080, State: "closed", Protocol: "tcp", Service: "http-alt"},
			},
		},
		{
			IP: "10.0.0.3", Ports: []contracts.PortInfo{
				{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
			},
		},
		{
			IP: "10.0.0.4", Ports: []contracts.PortInfo{},
		},
	}

	filtered, err := f.Filter(results)
	if err != nil {
		t.Fatalf("Filter error = %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 results, got %d", len(filtered))
	}
	if filtered[0].IP != "10.0.0.1" {
		t.Errorf("filtered[0].IP = %q, want %q", filtered[0].IP, "10.0.0.1")
	}
	if filtered[1].IP != "10.0.0.3" {
		t.Errorf("filtered[1].IP = %q, want %q", filtered[1].IP, "10.0.0.3")
	}
}

func TestOnlyUpFilter_Filter_Empty(t *testing.T) {
	f := &OnlyUpFilter{}
	results := []contracts.ScanResult{}
	filtered, err := f.Filter(results)
	if err != nil {
		t.Fatalf("Filter error = %v", err)
	}
	if len(filtered) != 0 {
		t.Fatalf("expected 0 results, got %d", len(filtered))
	}
}
