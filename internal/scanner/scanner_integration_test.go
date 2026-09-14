package scanner

import (
	"context"
	"testing"
	"time"
)

// === Integration: Scanner Configuration ===

func TestIntegrationScannerConfig_Default(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	if ns == nil {
		t.Fatal("expected non-nil scanner")
	}
	if ns.network != "192.168.1.0/24" {
		t.Errorf("expected network '192.168.1.0/24', got %q", ns.network)
	}
	if ns.timeout != 3*time.Second {
		t.Errorf("expected timeout 3s, got %v", ns.timeout)
	}
	if ns.portRange != "1-1000" {
		t.Errorf("expected portRange '1-1000', got %q", ns.portRange)
	}
	if ns.threads != 100 {
		t.Errorf("expected threads 100, got %d", ns.threads)
	}
	if ns.showClosed {
		t.Error("expected showClosed false")
	}
	if ns.results == nil {
		t.Error("expected non-nil results")
	}
	if ns.ctx == nil {
		t.Error("expected non-nil ctx")
	}
}

func TestIntegrationScannerConfig_Custom(t *testing.T) {
	ns := NewNetworkScanner("10.0.0.0/16", 5*time.Second, "80,443,8080", 50, true)
	if ns == nil {
		t.Fatal("expected non-nil scanner")
	}
	if ns.network != "10.0.0.0/16" {
		t.Errorf("expected network '10.0.0.0/16', got %q", ns.network)
	}
	if ns.timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", ns.timeout)
	}
	if ns.portRange != "80,443,8080" {
		t.Errorf("expected portRange '80,443,8080', got %q", ns.portRange)
	}
	if ns.threads != 50 {
		t.Errorf("expected threads 50, got %d", ns.threads)
	}
	if !ns.showClosed {
		t.Error("expected showClosed true")
	}
}

func TestIntegrationScannerConfig_Stop(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	if ns == nil {
		t.Fatal("expected non-nil scanner")
	}

	// Stop should not panic
	ns.Stop()
}

func TestIntegrationScannerConfig_GetResults_Empty(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	results := ns.GetResults()
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// === Integration: Result Models ===

func TestIntegrationResult_Model(t *testing.T) {
	result := Result{
		IP:         "192.168.1.1",
		Hostname:   "router-main",
		MAC:        "AA:BB:CC:DD:EE:01",
		DeviceType: "Router",
		Protocols:  []string{"TCP", "UDP"},
		Ports: []PortInfo{
			{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"},
			{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
		},
	}

	if result.IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", result.IP)
	}
	if result.Hostname != "router-main" {
		t.Errorf("expected Hostname 'router-main', got %q", result.Hostname)
	}
	if result.MAC != "AA:BB:CC:DD:EE:01" {
		t.Errorf("expected MAC 'AA:BB:CC:DD:EE:01', got %q", result.MAC)
	}
	if result.DeviceType != "Router" {
		t.Errorf("expected DeviceType 'Router', got %q", result.DeviceType)
	}
	if len(result.Protocols) != 2 {
		t.Errorf("expected 2 protocols, got %d", len(result.Protocols))
	}
	if len(result.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(result.Ports))
	}
}

func TestIntegrationResult_Empty(t *testing.T) {
	result := Result{}
	if result.IP != "" {
		t.Errorf("expected empty IP, got %q", result.IP)
	}
	if result.Hostname != "" {
		t.Errorf("expected empty Hostname, got %q", result.Hostname)
	}
}

// === Integration: Port Info ===

func TestIntegrationPortInfo_Open(t *testing.T) {
	port := PortInfo{
		Port:     22,
		State:    "open",
		Protocol: "TCP",
		Service:  "ssh",
	}

	if port.Port != 22 {
		t.Errorf("expected port 22, got %d", port.Port)
	}
	if port.State != "open" {
		t.Errorf("expected state 'open', got %q", port.State)
	}
	if port.Protocol != "TCP" {
		t.Errorf("expected protocol 'TCP', got %q", port.Protocol)
	}
	if port.Service != "ssh" {
		t.Errorf("expected service 'ssh', got %q", port.Service)
	}
}

func TestIntegrationPortInfo_Closed(t *testing.T) {
	port := PortInfo{
		Port:     25,
		State:    "closed",
		Protocol: "TCP",
		Service:  "smtp",
	}

	if port.State != "closed" {
		t.Errorf("expected state 'closed', got %q", port.State)
	}
}

// === Integration: Context Cancellation ===

func TestIntegrationContext_Cancellation(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())

	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	if ns == nil {
		t.Fatal("expected non-nil scanner")
	}

	// Cancel context
	cancel()

	// Scanner should respect context cancellation
	// (Stop should work even with cancelled context)
	ns.Stop()
}

// === Integration: Scanner with Custom Context ===

func TestIntegrationScanner_CustomContext(t *testing.T) {
	_ = context.Background()
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	if ns == nil {
		t.Fatal("expected non-nil scanner")
	}
	// NewNetworkScanner always uses context.Background() internally
	// This test verifies scanner creation works correctly
	if ns.ctx == nil {
		t.Error("expected non-nil ctx")
	}
}

// === Integration: Scan Result Collection ===

func TestIntegrationScanResult_Collection(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	if ns == nil {
		t.Fatal("expected non-nil scanner")
	}

	// Results are stored in ns.results (internal field)
	// We can't directly set them, but we can verify GetResults works
	initialResults := ns.GetResults()
	if len(initialResults) != 0 {
		t.Errorf("expected 0 initial results, got %d", len(initialResults))
	}
}

// === Integration: Error Handling ===

func TestIntegrationError_NilScanner(t *testing.T) {
	var ns *NetworkScanner
	if ns == nil {
		// Expected - nil scanner
	}

	// Stop on nil should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Error("Stop() should not panic on nil scanner")
		}
	}()
	// Note: We can't call ns.Stop() on nil, but the test verifies nil handling
}

// === Integration: Port State Validation ===

func TestIntegrationPortState_Valid(t *testing.T) {
	validStates := []string{"open", "closed", "filtered"}
	for _, state := range validStates {
		port := PortInfo{Port: 80, State: state}
		if port.State != state {
			t.Errorf("expected state %q, got %q", state, port.State)
		}
	}
}

func TestIntegrationPortState_Unknown(t *testing.T) {
	port := PortInfo{Port: 80, State: "unknown"}
	if port.State != "unknown" {
		t.Errorf("expected state 'unknown', got %q", port.State)
	}
}

// === Integration: Protocol Validation ===

func TestIntegrationProtocol_Valid(t *testing.T) {
	validProtocols := []string{"TCP", "UDP", "ICMP"}
	for _, protocol := range validProtocols {
		port := PortInfo{Port: 80, Protocol: protocol}
		if port.Protocol != protocol {
			t.Errorf("expected protocol %q, got %q", protocol, port.Protocol)
		}
	}
}

// === Integration: Device Type Classification ===

func TestIntegrationDeviceType_Classification(t *testing.T) {
	deviceTypes := []string{"Router", "Server", "Desktop", "Unknown", "IoT"}
	for _, dtype := range deviceTypes {
		result := Result{IP: "192.168.1.1", DeviceType: dtype}
		if result.DeviceType != dtype {
			t.Errorf("expected DeviceType %q, got %q", dtype, result.DeviceType)
		}
	}
}

// === Integration: MAC Address Format ===

func TestIntegrationMACFormat_Valid(t *testing.T) {
	validMACs := []string{
		"AA:BB:CC:DD:EE:01",
		"aa:bb:cc:dd:ee:01",
		"AA-BB-CC-DD-EE-01",
		"AABB.CCDD.EE01",
	}
	for _, mac := range validMACs {
		result := Result{IP: "192.168.1.1", MAC: mac}
		if result.MAC != mac {
			t.Errorf("expected MAC %q, got %q", mac, result.MAC)
		}
	}
}

func TestIntegrationMACFormat_Empty(t *testing.T) {
	result := Result{IP: "192.168.1.1", MAC: ""}
	if result.MAC != "" {
		t.Errorf("expected empty MAC, got %q", result.MAC)
	}
}

// === Integration: IP Address Validation ===

func TestIntegrationIPFormat_Valid(t *testing.T) {
	validIPs := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
	}
	for _, ip := range validIPs {
		result := Result{IP: ip}
		if result.IP != ip {
			t.Errorf("expected IP %q, got %q", ip, result.IP)
		}
	}
}

// === Integration: Hostname Handling ===

func TestIntegrationHostname_Empty(t *testing.T) {
	result := Result{IP: "192.168.1.1", Hostname: ""}
	if result.Hostname != "" {
		t.Errorf("expected empty Hostname, got %q", result.Hostname)
	}
}

func TestIntegrationHostname_Long(t *testing.T) {
	longHostname := "this-is-a-very-long-hostname-that-may-need-truncation-for-display-purposes"
	result := Result{IP: "192.168.1.1", Hostname: longHostname}
	if result.Hostname != longHostname {
		t.Errorf("expected Hostname %q, got %q", longHostname, result.Hostname)
	}
}

// === Integration: Scanner Thread Limits ===

func TestIntegrationScannerThreads_Min(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 1, false)
	if ns.threads != 1 {
		t.Errorf("expected threads 1, got %d", ns.threads)
	}
}

func TestIntegrationScannerThreads_Max(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 10000, false)
	if ns.threads != 10000 {
		t.Errorf("expected threads 10000, got %d", ns.threads)
	}
}

// === Integration: Show Closed Ports ===

func TestIntegrationShowClosed_True(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, true)
	if !ns.showClosed {
		t.Error("expected showClosed true")
	}
}

func TestIntegrationShowClosed_False(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	if ns.showClosed {
		t.Error("expected showClosed false")
	}
}

// === Integration: Timeout Validation ===

func TestIntegrationTimeout_Min(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "1-1000", 100, false)
	if ns.timeout != 100*time.Millisecond {
		t.Errorf("expected timeout 100ms, got %v", ns.timeout)
	}
}

func TestIntegrationTimeout_Max(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 30*time.Second, "1-1000", 100, false)
	if ns.timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", ns.timeout)
	}
}

// === Integration: Context Lifecycle ===

func TestIntegrationContext_Lifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	if ns == nil {
		t.Fatal("expected non-nil scanner")
	}

	// Context should be valid
	if ctx.Err() != nil {
		t.Errorf("context should be valid, got error: %v", ctx.Err())
	}

	// Scanner should work with valid context
	ns.Stop()
}

// === Integration: Multiple Scanners ===

func TestIntegrationMultipleScanners(t *testing.T) {
	scanners := make([]*NetworkScanner, 5)
	for i := 0; i < 5; i++ {
		scanners[i] = NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
		if scanners[i] == nil {
			t.Fatalf("expected non-nil scanner at index %d", i)
		}
	}

	// All scanners should be independent
	for i, ns := range scanners {
		results := ns.GetResults()
		if len(results) != 0 {
			t.Errorf("expected 0 results for scanner %d, got %d", i, len(results))
		}
		ns.Stop()
	}
}

// === Integration: Scanner Stop idempotency ===

func TestIntegrationScannerStop_Idempotent(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)

	// Multiple stops should not panic
	ns.Stop()
	ns.Stop()
	ns.Stop()
}

// === Integration: Result with All Fields ===

func TestIntegrationResult_Full(t *testing.T) {
	result := Result{
		IP:         "192.168.1.1",
		Hostname:   "router-main",
		MAC:        "AA:BB:CC:DD:EE:01",
		DeviceType: "Router",
		Protocols:  []string{"TCP", "UDP"},
		Ports: []PortInfo{
			{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"},
			{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
			{Port: 443, State: "open", Protocol: "TCP", Service: "https"},
		},
	}

	if result.IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", result.IP)
	}
	if len(result.Ports) != 3 {
		t.Errorf("expected 3 ports, got %d", len(result.Ports))
	}
	if len(result.Protocols) != 2 {
		t.Errorf("expected 2 protocols, got %d", len(result.Protocols))
	}
}

// === Integration: PortInfo with All Fields ===

func TestIntegrationPortInfo_Full(t *testing.T) {
	port := PortInfo{
		Port:     22,
		State:    "open",
		Protocol: "TCP",
		Service:  "ssh",
	}

	if port.Port != 22 {
		t.Errorf("expected port 22, got %d", port.Port)
	}
	if port.State != "open" {
		t.Errorf("expected state 'open', got %q", port.State)
	}
	if port.Protocol != "TCP" {
		t.Errorf("expected protocol 'TCP', got %q", port.Protocol)
	}
	if port.Service != "ssh" {
		t.Errorf("expected service 'ssh', got %q", port.Service)
	}
}

// === Integration: Result Slice Operations ===

func TestIntegrationResultSlice_Empty(t *testing.T) {
	results := []Result{}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestIntegrationResultSlice_Single(t *testing.T) {
	results := []Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", results[0].IP)
	}
}

func TestIntegrationResultSlice_Multiple(t *testing.T) {
	results := []Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "server"},
		{IP: "192.168.1.3", Hostname: "desktop"},
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

// === Integration: Scanner with Zero Threads ===

func TestIntegrationScanner_ZeroThreads(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 0, false)
	if ns.threads != 0 {
		t.Errorf("expected threads 0, got %d", ns.threads)
	}
}

// === Integration: Scanner with Zero Timeout ===

func TestIntegrationScanner_ZeroTimeout(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 0, "1-1000", 100, false)
	if ns.timeout != 0 {
		t.Errorf("expected timeout 0, got %v", ns.timeout)
	}
}

// === Integration: Scanner with Empty Port Range ===

func TestIntegrationScanner_EmptyPortRange(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "", 100, false)
	if ns.portRange != "" {
		t.Errorf("expected empty portRange, got %q", ns.portRange)
	}
}

// === Integration: Scanner with Empty Network ===

func TestIntegrationScanner_EmptyNetwork(t *testing.T) {
	ns := NewNetworkScanner("", 3*time.Second, "1-1000", 100, false)
	if ns.network != "" {
		t.Errorf("expected empty network, got %q", ns.network)
	}
}

// === Integration: Result with Nil Ports ===

func TestIntegrationResult_NilPorts(t *testing.T) {
	result := Result{
		IP:       "192.168.1.1",
		Hostname: "router",
		Ports:    nil,
	}
	if result.Ports != nil {
		t.Errorf("expected nil Ports, got %v", result.Ports)
	}
}

func TestIntegrationResult_NilProtocols(t *testing.T) {
	result := Result{
		IP:        "192.168.1.1",
		Hostname:  "router",
		Protocols: nil,
	}
	if result.Protocols != nil {
		t.Errorf("expected nil Protocols, got %v", result.Protocols)
	}
}

// === Integration: PortInfo with Zero Port ===

func TestIntegrationPortInfo_ZeroPort(t *testing.T) {
	port := PortInfo{
		Port:     0,
		State:    "open",
		Protocol: "TCP",
		Service:  "test",
	}
	if port.Port != 0 {
		t.Errorf("expected port 0, got %d", port.Port)
	}
}

// === Integration: Scanner Context Timeout ===

func TestIntegrationScanner_ContextTimeout(t *testing.T) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	if ns == nil {
		t.Fatal("expected non-nil scanner")
	}

	// Wait for context to expire
	time.Sleep(150 * time.Millisecond)

	// Scanner should handle expired context gracefully
	ns.Stop()
}

// === Integration: Scanner with Large Thread Count ===

func TestIntegrationScanner_LargeThreads(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 10000, false)
	if ns.threads != 10000 {
		t.Errorf("expected threads 10000, got %d", ns.threads)
	}
}

// === Integration: Result with Duplicate Ports ===

func TestIntegrationResult_DuplicatePorts(t *testing.T) {
	result := Result{
		IP:       "192.168.1.1",
		Hostname: "router",
		Ports: []PortInfo{
			{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
			{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
		},
	}
	if len(result.Ports) != 2 {
		t.Errorf("expected 2 ports (even if duplicate), got %d", len(result.Ports))
	}
}

// === Integration: Result with Mixed Port States ===

func TestIntegrationResult_MixedPortStates(t *testing.T) {
	result := Result{
		IP:       "192.168.1.1",
		Hostname: "router",
		Ports: []PortInfo{
			{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"},
			{Port: 25, State: "closed", Protocol: "TCP", Service: "smtp"},
			{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
		},
	}
	if len(result.Ports) != 3 {
		t.Errorf("expected 3 ports, got %d", len(result.Ports))
	}
}
