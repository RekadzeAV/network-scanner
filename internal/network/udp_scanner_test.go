package network

import (
	"net"
	"testing"
	"time"
)

func TestUDPScanState_Values(t *testing.T) {
	states := map[string]UDPScanState{
		"open":          UDPOpen,
		"closed":        UDPClosed,
		"filtered":      UDPFiltered,
		"open|filtered": UDPOpenFiltered,
	}

	for name, state := range states {
		if string(state) != name {
			t.Errorf("UDPScanState %q = %q, want %q", name, state, name)
		}
	}
}

func TestUDPAdvancedScanner_DefaultTimeout(t *testing.T) {
	s := UDPAdvancedScanner{}
	if s.Timeout != 0 {
		t.Fatalf("expected zero Timeout, got %v", s.Timeout)
	}
	res := s.ScanPort("192.0.2.1", 9)
	if res.State == UDPOpenFiltered {
		t.Logf("non-routable host: state=%v detail=%q", res.State, res.Detail)
	}
}

func TestUDPAdvancedScanner_ScanPort_Closed(t *testing.T) {
	s := UDPAdvancedScanner{Timeout: 50 * time.Millisecond}
	res := s.ScanPort("127.0.0.1", 1) // порт 1 обычно закрыт на локальном хосте
	if res.State != UDPClosed && res.State != UDPOpenFiltered {
		t.Fatalf("expected closed or open|filtered for port 1 on localhost, got %v (%q)", res.State, res.Detail)
	}
}

func TestUDPAdvancedScanner_ScanPorts(t *testing.T) {
	s := UDPAdvancedScanner{Timeout: 50 * time.Millisecond}
	ports := []int{1, 2, 3}
	res := s.ScanPorts("127.0.0.1", ports)
	if len(res) != len(ports) {
		t.Fatalf("expected %d results, got %d", len(ports), len(res))
	}
	for i, r := range res {
		if r.Port != ports[i] {
			t.Fatalf("result %d port = %d, want %d", i, r.Port, ports[i])
		}
	}
}

func TestUDPAdvancedScanner_ScanPortsParallel(t *testing.T) {
	s := UDPAdvancedScanner{Timeout: 50 * time.Millisecond, Threads: 10}
	ports := []int{1, 2, 3, 4, 5}
	res := s.ScanPortsParallel("127.0.0.1", ports)
	if len(res) != len(ports) {
		t.Fatalf("expected %d results, got %d", len(ports), len(res))
	}
	for i, r := range res {
		if r.Port != ports[i] {
			t.Fatalf("result %d port = %d, want %d", i, r.Port, ports[i])
		}
	}
}

func TestUDPAdvancedScanner_ThreadsDefault(t *testing.T) {
	s := UDPAdvancedScanner{Timeout: 50 * time.Millisecond}
	_ = s.ScanPortsParallel("127.0.0.1", []int{1})
}

func TestIsUDPPortOpenAdvanced_Alias(t *testing.T) {
	s := UDPAdvancedScanner{Timeout: 50 * time.Millisecond}
	res := s.ScanPort("127.0.0.1", 1)
	expected := res.State == UDPOpen || res.State == UDPOpenFiltered
	if IsUDPPortOpenAdvanced("127.0.0.1", 1, 50*time.Millisecond) != expected {
		t.Fatalf("IsUDPPortOpenAdvanced mismatch (state=%v)", res.State)
	}
}

func TestUDPAdvancedScanner_InvalidHost(t *testing.T) {
	s := UDPAdvancedScanner{Timeout: 50 * time.Millisecond}
	res := s.ScanPort("256.256.256.256", 80)
	if res.State != UDPClosed {
		t.Fatalf("expected closed for invalid host, got %v (%q)", res.State, res.Detail)
	}
}

func TestJoinHostPortIPv6(t *testing.T) {
	addr := net.JoinHostPort("::1", "53")
	if addr != "[::1]:53" {
		t.Fatalf("expected [::1]:53, got %q", addr)
	}
}
