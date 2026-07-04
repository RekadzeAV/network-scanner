package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"network-scanner/internal/contracts"
)

func TestInventoryService_SaveAndListSnapshots(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_inventory.db")

	svc := NewInventoryService(dbPath)

	ctx := context.Background()
	testData := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "router",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		},
		{
			IP:       "192.168.1.2",
			Hostname: "server",
			Ports: []contracts.PortInfo{
				{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
			},
		},
	}

	// Сохранение снапшота
	if err := svc.SaveSnapshot(ctx, "test-1", testData); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	// Проверка что файл создан
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}

	// Список снапшотов
	snapshots, err := svc.ListSnapshots(ctx, 10)
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}

	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snapshots))
	}

	if snapshots[0].ID != "test-1" {
		t.Fatalf("expected ID 'test-1', got '%s'", snapshots[0].ID)
	}
}

func TestInventoryService_Diff(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_diff.db")

	svc := NewInventoryService(dbPath)
	ctx := context.Background()

	// Снапшот 1
	data1 := []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.2", Hostname: "host2"},
	}
	if err := svc.SaveSnapshot(ctx, "snap1", data1); err != nil {
		t.Fatalf("SaveSnapshot snap1 failed: %v", err)
	}

	// Снапшот 2 (изменённый)
	data2 := []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "host1-modified"},
		{IP: "192.168.1.3", Hostname: "host3-new"},
	}
	if err := svc.SaveSnapshot(ctx, "snap2", data2); err != nil {
		t.Fatalf("SaveSnapshot snap2 failed: %v", err)
	}

	// Diff
	diff, err := svc.Diff(ctx, "snap1", "snap2")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// host1 изменён
	if len(diff.Changed) != 1 {
		t.Fatalf("expected 1 changed, got %d", len(diff.Changed))
	}

	// host3 новый
	if len(diff.New) != 1 {
		t.Fatalf("expected 1 new, got %d", len(diff.New))
	}

	// host2 отсутствующий
	if len(diff.Missing) != 1 {
		t.Fatalf("expected 1 missing, got %d", len(diff.Missing))
	}
}

func TestInventoryService_Diff_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_notfound.db")

	svc := NewInventoryService(dbPath)
	ctx := context.Background()

	_, err := svc.Diff(ctx, "nonexistent1", "nonexistent2")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshots")
	}
}

func TestInventoryService_ListSnapshots_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_empty.db")

	svc := NewInventoryService(dbPath)
	ctx := context.Background()

	snapshots, err := svc.ListSnapshots(ctx, 10)
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}

	if len(snapshots) != 0 {
		t.Fatalf("expected 0 snapshots, got %d", len(snapshots))
	}
}

func TestNewInventoryService_DefaultPath(t *testing.T) {
	svc := NewInventoryService("")
	if svc.dbPath != "inventory/network_inventory.db" {
		t.Fatalf("expected default path, got '%s'", svc.dbPath)
	}
}

func TestNewInventoryService_CustomPath(t *testing.T) {
	svc := NewInventoryService("/custom/path.db")
	if svc.dbPath != "/custom/path.db" {
		t.Fatalf("expected '/custom/path.db', got '%s'", svc.dbPath)
	}
}

func TestConvertToInternalResults(t *testing.T) {
	contractsData := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "test-host",
			MAC:      "AA:BB:CC:DD:EE:FF",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http", Banner: "Apache", Version: "2.4"},
				{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
			},
			DeviceType:   "server",
			DeviceVendor: "Dell",
			GuessOS:      "Linux",
		},
	}

	internal := ConvertToInternalResults(contractsData)

	if len(internal) != 1 {
		t.Fatalf("expected 1 result, got %d", len(internal))
	}

	r := internal[0]
	if r.IP != "192.168.1.1" {
		t.Fatalf("expected IP '192.168.1.1', got '%s'", r.IP)
	}
	if r.Hostname != "test-host" {
		t.Fatalf("expected Hostname 'test-host', got '%s'", r.Hostname)
	}
	if len(r.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(r.Ports))
	}
	if r.Ports[0].Banner != "Apache" {
		t.Fatalf("expected Banner 'Apache', got '%s'", r.Ports[0].Banner)
	}
	if r.DeviceType != "server" {
		t.Fatalf("expected DeviceType 'server', got '%s'", r.DeviceType)
	}
}

func TestConvertToInternalResults_Empty(t *testing.T) {
	internal := ConvertToInternalResults(nil)
	if len(internal) != 0 {
		t.Fatalf("expected 0 results, got %d", len(internal))
	}
}

func TestConvertToInternalResults_PortConversion(t *testing.T) {
	contractsData := []contracts.ScanResult{
		{
			IP: "10.0.0.1",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "closed", Protocol: "tcp", Service: "ssh"},
				{Port: 53, State: "open", Protocol: "udp", Service: "dns"},
			},
		},
	}

	internal := ConvertToInternalResults(contractsData)

	if len(internal[0].Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(internal[0].Ports))
	}

	if internal[0].Ports[0].Port != 22 {
		t.Fatalf("expected port 22, got %d", internal[0].Ports[0].Port)
	}
	if internal[0].Ports[0].State != "closed" {
		t.Fatalf("expected state 'closed', got '%s'", internal[0].Ports[0].State)
	}
	if internal[0].Ports[1].Protocol != "udp" {
		t.Fatalf("expected protocol 'udp', got '%s'", internal[0].Ports[1].Protocol)
	}
}
