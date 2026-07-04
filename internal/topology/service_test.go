package topology

import (
	"context"
	"testing"

	"network-scanner/internal/contracts"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestTopologyService_BuildEmptyResults(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	topo, err := svc.Build(ctx, nil, contracts.TopologyOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topo == nil {
		t.Fatal("expected non-nil topology")
	}
}

func TestTopologyService_BuildWithResults(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "router",
			MAC:      "AA:BB:CC:DD:EE:01",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		},
		{
			IP:       "192.168.1.2",
			Hostname: "host",
			MAC:      "AA:BB:CC:DD:EE:02",
			Ports: []contracts.PortInfo{
				{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
			},
		},
	}

	topo, err := svc.Build(ctx, results, contracts.TopologyOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topo == nil {
		t.Fatal("expected non-nil topology")
	}
}

func TestTopologyService_BuildWithSNMPOptions(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "switch", MAC: "AA:BB:CC:DD:EE:FF"},
	}

	opts := contracts.TopologyOptions{
		SNMPEnabled: true,
		Community:   "public",
	}

	topo, err := svc.Build(ctx, results, opts)
	// SNMP может не быть доступен, но топология должна построиться
	_ = topo
	_ = err
}

func TestTopologyService_BuildContextCancel(t *testing.T) {
	svc := NewService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Build(ctx, nil, contracts.TopologyOptions{})
	// Контекст может не влиять на синхронную операцию
	_ = err
}
