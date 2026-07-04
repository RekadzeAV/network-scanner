package comparator

import (
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

func TestBuildHistoryEntry(t *testing.T) {
	hosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", GuessOS: "Cisco IOS", DeviceVendor: "Cisco"},
		{IP: "192.168.1.2", Hostname: "switch", GuessOS: "Linux", DeviceVendor: "TP-Link"},
	}
	started := time.Now()
	completed := started.Add(5 * time.Minute)

	entry := BuildHistoryEntry("scan-001", "192.168.1.0/24", hosts, started, completed)

	if entry.ID != "scan-001" {
		t.Errorf("expected ID 'scan-001', got '%s'", entry.ID)
	}
	if entry.HostCount != 2 {
		t.Errorf("expected 2 hosts, got %d", entry.HostCount)
	}
	if entry.OSMap["Cisco IOS"] != 1 {
		t.Error("expected 1 Cisco IOS host")
	}
	if entry.VendorMap["Cisco"] != 1 {
		t.Error("expected 1 Cisco vendor")
	}
}

func TestCompareSnapshots_NewHost(t *testing.T) {
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}

	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)

	if len(result.NewHosts) != 1 {
		t.Errorf("expected 1 new host, got %d", len(result.NewHosts))
	}
	if result.NewHosts[0].IP != "192.168.1.2" {
		t.Errorf("expected new host IP 192.168.1.2, got %s", result.NewHosts[0].IP)
	}
}

func TestCompareSnapshots_RemovedHost(t *testing.T) {
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}

	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)

	if len(result.RemovedHosts) != 1 {
		t.Errorf("expected 1 removed host, got %d", len(result.RemovedHosts))
	}
	if result.RemovedHosts[0].IP != "192.168.1.2" {
		t.Errorf("expected removed host IP 192.168.1.2, got %s", result.RemovedHosts[0].IP)
	}
}

func TestCompareSnapshots_ChangedHost(t *testing.T) {
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", GuessOS: "Cisco IOS"},
	}
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router-new", GuessOS: "IOS"},
	}

	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)

	if len(result.ChangedHosts) != 1 {
		t.Fatalf("expected 1 changed host, got %d", len(result.ChangedHosts))
	}
	if len(result.ChangedHosts[0].ChangedIn) == 0 {
		t.Error("expected some changes, got none")
	}
}

func TestCompareSnapshots_PortChange(t *testing.T) {
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}},
	}
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "closed"}, {Port: 443, Protocol: "tcp", State: "open"}}},
	}

	result := CompareSnapshots("scan-a", "scan-b", hostsA, hostsB)

	if len(result.PortChanges) == 0 {
		t.Error("expected port changes, got none")
	}
}

func TestCompareSnapshots_NoChanges(t *testing.T) {
	hosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}

	result := CompareSnapshots("scan-a", "scan-b", hosts, hosts)

	if result.TotalDiff != 0 {
		t.Errorf("expected 0 total diff, got %d", result.TotalDiff)
	}
	if len(result.NewHosts) != 0 || len(result.RemovedHosts) != 0 || len(result.ChangedHosts) != 0 {
		t.Error("expected no changes")
	}
}
