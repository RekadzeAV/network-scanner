package comparator

import (
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

// ============================================================================
// BuildHistoryEntry — полное покрытие
// ============================================================================

func TestBuildHistoryEntry_EmptyHosts(t *testing.T) {
	entry := BuildHistoryEntry("scan-002", "10.0.0.0/24", []scanner.Result{}, time.Now(), time.Now())
	if entry.ID != "scan-002" {
		t.Errorf("expected scan-002, got %s", entry.ID)
	}
	if entry.HostCount != 0 {
		t.Errorf("expected 0 hosts, got %d", entry.HostCount)
	}
	if len(entry.Ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(entry.Ports))
	}
}

func TestBuildHistoryEntry_MultipleOS(t *testing.T) {
	hosts := []scanner.Result{
		{IP: "192.168.1.1", GuessOS: "Linux"},
		{IP: "192.168.1.2", GuessOS: "Windows"},
		{IP: "192.168.1.3", GuessOS: "Linux"},
	}
	entry := BuildHistoryEntry("scan-003", "192.168.1.0/24", hosts, time.Now(), time.Now())
	if entry.OSMap["Linux"] != 2 {
		t.Errorf("expected 2 Linux, got %d", entry.OSMap["Linux"])
	}
	if entry.OSMap["Windows"] != 1 {
		t.Errorf("expected 1 Windows, got %d", entry.OSMap["Windows"])
	}
}

func TestBuildHistoryEntry_MultipleVendors(t *testing.T) {
	hosts := []scanner.Result{
		{IP: "192.168.1.1", DeviceVendor: "Cisco"},
		{IP: "192.168.1.2", DeviceVendor: "TP-Link"},
		{IP: "192.168.1.3", DeviceVendor: "Cisco"},
	}
	entry := BuildHistoryEntry("scan-004", "192.168.1.0/24", hosts, time.Now(), time.Now())
	if entry.VendorMap["Cisco"] != 2 {
		t.Errorf("expected 2 Cisco, got %d", entry.VendorMap["Cisco"])
	}
	if entry.VendorMap["TP-Link"] != 1 {
		t.Errorf("expected 1 TP-Link, got %d", entry.VendorMap["TP-Link"])
	}
}

func TestBuildHistoryEntry_OpenPorts(t *testing.T) {
	hosts := []scanner.Result{
		{
			IP: "192.168.1.1",
			Ports: []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", State: "open"},
				{Port: 443, Protocol: "tcp", State: "open"},
				{Port: 22, Protocol: "tcp", State: "closed"},
			},
		},
	}
	entry := BuildHistoryEntry("scan-005", "192.168.1.0/24", hosts, time.Now(), time.Now())
	if entry.Ports["80/tcp"] != 1 {
		t.Errorf("expected 1 open port 80/tcp, got %d", entry.Ports["80/tcp"])
	}
	if entry.Ports["443/tcp"] != 1 {
		t.Errorf("expected 1 open port 443/tcp, got %d", entry.Ports["443/tcp"])
	}
	if _, exists := entry.Ports["22/tcp"]; exists {
		t.Error("closed port 22/tcp should not be in entry")
	}
}

func TestBuildHistoryEntry_CaseInsensitiveState(t *testing.T) {
	hosts := []scanner.Result{
		{
			IP: "192.168.1.1",
			Ports: []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", State: "Open"},
				{Port: 443, Protocol: "tcp", State: "OPEN"},
			},
		},
	}
	entry := BuildHistoryEntry("scan-006", "192.168.1.0/24", hosts, time.Now(), time.Now())
	if entry.Ports["80/tcp"] != 1 {
		t.Errorf("expected 1 open port 80/tcp, got %d", entry.Ports["80/tcp"])
	}
	if entry.Ports["443/tcp"] != 1 {
		t.Errorf("expected 1 open port 443/tcp, got %d", entry.Ports["443/tcp"])
	}
}

// ============================================================================
// CompareSnapshots — полное покрытие
// ============================================================================

func TestCompareSnapshots_EmptyBoth(t *testing.T) {
	result := CompareSnapshots("scan-a", "scan-b", []scanner.Result{}, []scanner.Result{})
	if result.TotalDiff != 0 {
		t.Errorf("expected 0 total diff, got %d", result.TotalDiff)
	}
}

func TestCompareSnapshots_AllNew(t *testing.T) {
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.2", Hostname: "host2"},
	}
	result := CompareSnapshots("scan-a", "scan-b", []scanner.Result{}, hostsB)
	if len(result.NewHosts) != 2 {
		t.Errorf("expected 2 new hosts, got %d", len(result.NewHosts))
	}
	if len(result.RemovedHosts) != 0 {
		t.Errorf("expected 0 removed hosts, got %d", len(result.RemovedHosts))
	}
}

func TestCompareSnapshots_AllRemoved(t *testing.T) {
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.2", Hostname: "host2"},
	}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, []scanner.Result{})
	if len(result.RemovedHosts) != 2 {
		t.Errorf("expected 2 removed hosts, got %d", len(result.RemovedHosts))
	}
}

func TestCompareSnapshots_ChangedHostname(t *testing.T) {
	hostsA := []scanner.Result{{IP: "192.168.1.1", Hostname: "old"}}
	hostsB := []scanner.Result{{IP: "192.168.1.1", Hostname: "new"}}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	if len(result.ChangedHosts) != 1 {
		t.Errorf("expected 1 changed host, got %d", len(result.ChangedHosts))
	}
	if result.ChangedHosts[0].Hostname != "old" {
		t.Errorf("expected hostname old, got %s", result.ChangedHosts[0].Hostname)
	}
}

func TestCompareSnapshots_ChangedOS(t *testing.T) {
	hostsA := []scanner.Result{{IP: "192.168.1.1", GuessOS: "Linux"}}
	hostsB := []scanner.Result{{IP: "192.168.1.1", GuessOS: "Windows"}}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	if len(result.ChangedHosts) != 1 {
		t.Errorf("expected 1 changed host, got %d", len(result.ChangedHosts))
	}
}

func TestCompareSnapshots_ChangedVendor(t *testing.T) {
	hostsA := []scanner.Result{{IP: "192.168.1.1", DeviceVendor: "Cisco"}}
	hostsB := []scanner.Result{{IP: "192.168.1.1", DeviceVendor: "TP-Link"}}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	if len(result.ChangedHosts) != 1 {
		t.Errorf("expected 1 changed host, got %d", len(result.ChangedHosts))
	}
}

func TestCompareSnapshots_ChangedDeviceType(t *testing.T) {
	hostsA := []scanner.Result{{IP: "192.168.1.1", DeviceType: "Router"}}
	hostsB := []scanner.Result{{IP: "192.168.1.1", DeviceType: "Switch"}}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	if len(result.ChangedHosts) != 1 {
		t.Errorf("expected 1 changed host, got %d", len(result.ChangedHosts))
	}
}

func TestCompareSnapshots_PortOpened(t *testing.T) {
	hostsA := []scanner.Result{{IP: "192.168.1.1"}}
	hostsB := []scanner.Result{{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}}}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	if len(result.PortChanges) != 1 {
		t.Errorf("expected 1 port change, got %d", len(result.PortChanges))
	}
	if result.PortChanges[0].ChangedFrom != "closed" {
		t.Errorf("expected changedFrom closed, got %s", result.PortChanges[0].ChangedFrom)
	}
}

func TestCompareSnapshots_PortClosed(t *testing.T) {
	hostsA := []scanner.Result{{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}}}
	hostsB := []scanner.Result{{IP: "192.168.1.1"}}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	if len(result.PortChanges) != 1 {
		t.Errorf("expected 1 port change, got %d", len(result.PortChanges))
	}
	if result.PortChanges[0].ChangedTo != "closed" {
		t.Errorf("expected changedTo closed, got %s", result.PortChanges[0].ChangedTo)
	}
}

func TestCompareSnapshots_PortStateChange(t *testing.T) {
	hostsA := []scanner.Result{{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}}}
	hostsB := []scanner.Result{{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "closed"}}}}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	if len(result.PortChanges) != 1 {
		t.Errorf("expected 1 port change, got %d", len(result.PortChanges))
	}
	if result.PortChanges[0].ChangedFrom != "open" {
		t.Errorf("expected changedFrom open, got %s", result.PortChanges[0].ChangedFrom)
	}
}

func TestCompareSnapshots_MultipleChanges(t *testing.T) {
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "old", GuessOS: "Linux", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}},
	}
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "new", GuessOS: "Windows", Ports: []scanner.PortInfo{{Port: 443, Protocol: "tcp", State: "open"}}},
	}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	if len(result.ChangedHosts) != 1 {
		t.Errorf("expected 1 changed host, got %d", len(result.ChangedHosts))
	}
	if len(result.ChangedHosts[0].ChangedIn) < 3 {
		t.Errorf("expected at least 3 change categories, got %d", len(result.ChangedHosts[0].ChangedIn))
	}
	if len(result.PortChanges) != 2 {
		t.Errorf("expected 2 port changes, got %d", len(result.PortChanges))
	}
}

func TestCompareSnapshots_SortedNewHosts(t *testing.T) {
	hostsB := []scanner.Result{
		{IP: "192.168.1.3", Hostname: "host3"},
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.2", Hostname: "host2"},
	}
	result := CompareSnapshots("scan-a", "scan-b", []scanner.Result{}, hostsB)
	if result.NewHosts[0].IP != "192.168.1.1" {
		t.Errorf("expected first host 192.168.1.1, got %s", result.NewHosts[0].IP)
	}
}

func TestCompareSnapshots_SortedRemovedHosts(t *testing.T) {
	hostsA := []scanner.Result{
		{IP: "192.168.1.3", Hostname: "host3"},
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.2", Hostname: "host2"},
	}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, []scanner.Result{})
	if result.RemovedHosts[0].IP != "192.168.1.1" {
		t.Errorf("expected first removed host 192.168.1.1, got %s", result.RemovedHosts[0].IP)
	}
}

func TestCompareSnapshots_SortedChangedHosts(t *testing.T) {
	hostsA := []scanner.Result{
		{IP: "192.168.1.3", Hostname: "host3"},
		{IP: "192.168.1.1", Hostname: "host1"},
	}
	hostsB := []scanner.Result{
		{IP: "192.168.1.3", Hostname: "host3-new"},
		{IP: "192.168.1.1", Hostname: "host1-new"},
	}
	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	if result.ChangedHosts[0].IP != "192.168.1.1" {
		t.Errorf("expected first changed host 192.168.1.1, got %s", result.ChangedHosts[0].IP)
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkBuildHistoryEntry(b *testing.B) {
	hosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", GuessOS: "Linux", DeviceVendor: "Cisco"},
		{IP: "192.168.1.2", Hostname: "switch", GuessOS: "Windows", DeviceVendor: "TP-Link"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildHistoryEntry("scan-001", "192.168.1.0/24", hosts, time.Now(), time.Now())
	}
}

func BenchmarkCompareSnapshots(b *testing.B) {
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", GuessOS: "Linux"},
		{IP: "192.168.1.2", Hostname: "switch", GuessOS: "Windows"},
	}
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router-new", GuessOS: "Linux"},
		{IP: "192.168.1.3", Hostname: "server", GuessOS: "Linux"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)
	}
}
