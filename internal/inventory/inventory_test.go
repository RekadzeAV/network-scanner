package inventory

import (
	"path/filepath"
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

func TestSaveLoadAndDiff(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "inventory.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	snapA := []scanner.Result{
		{IP: "192.168.1.10", MAC: "AA:AA:AA:AA:AA:10", Hostname: "cam-1", Ports: []scanner.PortInfo{{Port: 554, Protocol: "tcp", State: "open"}}},
		{IP: "192.168.1.20", MAC: "AA:AA:AA:AA:AA:20", Hostname: "pc-1", Ports: []scanner.PortInfo{{Port: 445, Protocol: "tcp", State: "open"}}},
	}
	snapB := []scanner.Result{
		{IP: "192.168.1.10", MAC: "AA:AA:AA:AA:AA:10", Hostname: "cam-1-renamed", Ports: []scanner.PortInfo{{Port: 554, Protocol: "tcp", State: "open"}}},
		{IP: "192.168.1.30", MAC: "AA:AA:AA:AA:AA:30", Hostname: "new-host", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}},
	}

	if err := store.SaveSnapshot("scan-a", time.Now().UTC(), snapA); err != nil {
		t.Fatalf("save snapshot A: %v", err)
	}
	if err := store.SaveSnapshot("scan-b", time.Now().UTC(), snapB); err != nil {
		t.Fatalf("save snapshot B: %v", err)
	}

	loadedA, err := store.LoadSnapshot("scan-a")
	if err != nil {
		t.Fatalf("load snapshot A: %v", err)
	}
	if len(loadedA.Hosts) != 2 {
		t.Fatalf("expected 2 hosts in snapshot A, got %d", len(loadedA.Hosts))
	}

	diff, err := store.Diff("scan-a", "scan-b")
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if len(diff.New) != 1 {
		t.Fatalf("expected 1 new host, got %d", len(diff.New))
	}
	if len(diff.Missing) != 1 {
		t.Fatalf("expected 1 missing host, got %d", len(diff.Missing))
	}
	if len(diff.Changed) != 1 {
		t.Fatalf("expected 1 changed host, got %d", len(diff.Changed))
	}
}
