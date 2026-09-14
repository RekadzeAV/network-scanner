package inventory

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

// === Integration: Store Open/Close ===

func TestIntegrationStore_OpenAndClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "inventory.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestIntegrationStore_OpenEmptyPath(t *testing.T) {
	_, err := Open("")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestIntegrationStore_OpenWhitespacePath(t *testing.T) {
	_, err := Open("   ")
	if err == nil {
		t.Error("expected error for whitespace path")
	}
}

func TestIntegrationStore_CloseNil(t *testing.T) {
	var store *Store
	err := store.Close()
	if err != nil {
		t.Errorf("expected no error for nil store, got %v", err)
	}
}

// === Integration: SaveSnapshot ===

func TestIntegrationSaveSnapshot_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	hosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "test-host", MAC: "aa:bb:cc:dd:ee:ff"},
	}

	err = store.SaveSnapshot("scan-001", time.Now(), hosts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestIntegrationSaveSnapshot_EmptyScanID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	err = store.SaveSnapshot("", time.Now(), nil)
	if err == nil {
		t.Error("expected error for empty scanID")
	}
}

func TestIntegrationSaveSnapshot_ZeroTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	// Zero timestamp should default to now
	err = store.SaveSnapshot("scan-002", time.Time{}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestIntegrationSaveSnapshot_WhitespaceScanID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	err = store.SaveSnapshot("  scan-003  ", time.Now(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestIntegrationSaveSnapshot_NilStore(t *testing.T) {
	var store *Store
	err := store.SaveSnapshot("scan-004", time.Now(), nil)
	if err == nil {
		t.Error("expected error for nil store")
	}
}

// === Integration: LoadSnapshot ===

func TestIntegrationLoadSnapshot_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	hosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "test-host", MAC: "aa:bb:cc:dd:ee:ff"},
	}

	err = store.SaveSnapshot("scan-load-001", time.Now(), hosts)
	if err != nil {
		t.Fatalf("SaveSnapshot error: %v", err)
	}

	snap, err := store.LoadSnapshot("scan-load-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snap.ID != "scan-load-001" {
		t.Errorf("expected ID 'scan-load-001', got %q", snap.ID)
	}
	if len(snap.Hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(snap.Hosts))
	}
}

func TestIntegrationLoadSnapshot_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	_, err = store.LoadSnapshot("non-existent")
	if err == nil {
		t.Error("expected error for non-existent snapshot")
	}
}

func TestIntegrationLoadSnapshot_EmptyScanID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	_, err = store.LoadSnapshot("")
	if err == nil {
		t.Error("expected error for empty scanID")
	}
}

func TestIntegrationLoadSnapshot_NilStore(t *testing.T) {
	var store *Store
	_, err := store.LoadSnapshot("scan-001")
	if err == nil {
		t.Error("expected error for nil store")
	}
}

// === Integration: ListSnapshots ===

func TestIntegrationListSnapshots_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	snapshots, err := store.ListSnapshots(0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snapshots == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(snapshots) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snapshots))
	}
}

func TestIntegrationListSnapshots_WithSnapshots(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	// Save multiple snapshots
	for i := 0; i < 3; i++ {
		err = store.SaveSnapshot(
			"scan-list-00"+string(rune('1'+i)),
			time.Now().Add(time.Duration(i)*time.Hour),
			nil,
		)
		if err != nil {
			t.Fatalf("SaveSnapshot error: %v", err)
		}
	}

	snapshots, err := store.ListSnapshots(0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(snapshots) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(snapshots))
	}
}

func TestIntegrationListSnapshots_WithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	for i := 0; i < 5; i++ {
		err = store.SaveSnapshot(
			"scan-limit-00"+string(rune('1'+i)),
			time.Now(),
			nil,
		)
		if err != nil {
			t.Fatalf("SaveSnapshot error: %v", err)
		}
	}

	snapshots, err := store.ListSnapshots(2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(snapshots) != 2 {
		t.Errorf("expected 2 snapshots with limit, got %d", len(snapshots))
	}
}

// === Integration: Diff ===

func TestIntegrationDiff_NewHost(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	// Save scan A (empty)
	err = store.SaveSnapshot("diff-a", time.Now(), nil)
	if err != nil {
		t.Fatalf("SaveSnapshot A error: %v", err)
	}

	// Save scan B (with new host)
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "new-host", MAC: "aa:bb:cc:dd:ee:ff"},
	}
	err = store.SaveSnapshot("diff-b", time.Now(), hostsB)
	if err != nil {
		t.Fatalf("SaveSnapshot B error: %v", err)
	}

	diff, err := store.Diff("diff-a", "diff-b")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(diff.New) != 1 {
		t.Errorf("expected 1 new host, got %d", len(diff.New))
	}
	if diff.New[0].IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", diff.New[0].IP)
	}
}

func TestIntegrationDiff_MissingHost(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	// Save scan A (with host)
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "old-host", MAC: "aa:bb:cc:dd:ee:ff"},
	}
	err = store.SaveSnapshot("diff-missing-a", time.Now(), hostsA)
	if err != nil {
		t.Fatalf("SaveSnapshot A error: %v", err)
	}

	// Save scan B (empty)
	err = store.SaveSnapshot("diff-missing-b", time.Now(), nil)
	if err != nil {
		t.Fatalf("SaveSnapshot B error: %v", err)
	}

	diff, err := store.Diff("diff-missing-a", "diff-missing-b")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(diff.Missing) != 1 {
		t.Errorf("expected 1 missing host, got %d", len(diff.Missing))
	}
}

func TestIntegrationDiff_ChangedHost(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	// Save scan A
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "old-hostname", MAC: "aa:bb:cc:dd:ee:ff"},
	}
	err = store.SaveSnapshot("diff-changed-a", time.Now(), hostsA)
	if err != nil {
		t.Fatalf("SaveSnapshot A error: %v", err)
	}

	// Save scan B with changed hostname
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "new-hostname", MAC: "aa:bb:cc:dd:ee:ff"},
	}
	err = store.SaveSnapshot("diff-changed-b", time.Now(), hostsB)
	if err != nil {
		t.Fatalf("SaveSnapshot B error: %v", err)
	}

	diff, err := store.Diff("diff-changed-a", "diff-changed-b")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(diff.Changed) != 1 {
		t.Errorf("expected 1 changed host, got %d", len(diff.Changed))
	}
	if len(diff.Changed[0].ChangedField) == 0 {
		t.Error("expected at least one changed field")
	}
}

func TestIntegrationDiff_SameHosts(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	hosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "test", MAC: "aa:bb:cc:dd:ee:ff"},
	}
	err = store.SaveSnapshot("diff-same-a", time.Now(), hosts)
	if err != nil {
		t.Fatalf("SaveSnapshot A error: %v", err)
	}
	err = store.SaveSnapshot("diff-same-b", time.Now(), hosts)
	if err != nil {
		t.Fatalf("SaveSnapshot B error: %v", err)
	}

	diff, err := store.Diff("diff-same-a", "diff-same-b")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(diff.New) != 0 {
		t.Errorf("expected 0 new hosts, got %d", len(diff.New))
	}
	if len(diff.Missing) != 0 {
		t.Errorf("expected 0 missing hosts, got %d", len(diff.Missing))
	}
	if len(diff.Changed) != 0 {
		t.Errorf("expected 0 changed hosts, got %d", len(diff.Changed))
	}
}

func TestIntegrationDiff_NonExistentScanA(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	_, err = store.Diff("non-existent-a", "non-existent-b")
	if err == nil {
		t.Error("expected error for non-existent scan A")
	}
}

func TestIntegrationDiff_NilStore(t *testing.T) {
	var store *Store
	_, err := store.Diff("scan-a", "scan-b")
	if err == nil {
		t.Error("expected error for nil store")
	}
}

// === Integration: Full Inventory Pipeline ===

func TestIntegrationFullInventoryPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "inventory.db")

	// Step 1: Open store
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	// Step 2: Save initial snapshot
	hostsA := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", MAC: "aa:bb:cc:dd:ee:ff"},
		{IP: "192.168.1.2", Hostname: "switch", MAC: "11:22:33:44:55:66"},
	}
	err = store.SaveSnapshot("pipeline-a", time.Now(), hostsA)
	if err != nil {
		t.Fatalf("SaveSnapshot A error: %v", err)
	}

	// Step 3: Load and verify
	snap, err := store.LoadSnapshot("pipeline-a")
	if err != nil {
		t.Fatalf("LoadSnapshot error: %v", err)
	}
	if len(snap.Hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(snap.Hosts))
	}

	// Step 4: Update hosts
	hostsB := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", MAC: "aa:bb:cc:dd:ee:ff"},
		{IP: "192.168.1.2", Hostname: "switch-v2", MAC: "11:22:33:44:55:66"}, // Changed hostname
		{IP: "192.168.1.3", Hostname: "new-host", MAC: "77:88:99:aa:bb:cc"},  // New host
	}
	err = store.SaveSnapshot("pipeline-b", time.Now(), hostsB)
	if err != nil {
		t.Fatalf("SaveSnapshot B error: %v", err)
	}

	// Step 5: Diff
	diff, err := store.Diff("pipeline-a", "pipeline-b")
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}
	if len(diff.New) != 1 {
		t.Errorf("expected 1 new host, got %d", len(diff.New))
	}
	if len(diff.Changed) != 1 {
		t.Errorf("expected 1 changed host, got %d", len(diff.Changed))
	}

	// Step 6: List snapshots
	snapshots, err := store.ListSnapshots(0)
	if err != nil {
		t.Fatalf("ListSnapshots error: %v", err)
	}
	if len(snapshots) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(snapshots))
	}
}

// === Integration: Edge Cases ===

func TestIntegrationSaveSnapshot_WithPorts(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	hosts := []scanner.Result{
		{
			IP: "192.168.1.1",
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		},
	}

	err = store.SaveSnapshot("scan-ports", time.Now(), hosts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snap, err := store.LoadSnapshot("scan-ports")
	if err != nil {
		t.Fatalf("LoadSnapshot error: %v", err)
	}
	if len(snap.Hosts[0].Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(snap.Hosts[0].Ports))
	}
}

func TestIntegrationLoadSnapshot_WhitespaceScanID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	err = store.SaveSnapshot("  scan-ws  ", time.Now(), nil)
	if err != nil {
		t.Fatalf("SaveSnapshot error: %v", err)
	}

	snap, err := store.LoadSnapshot("  scan-ws  ")
	if err != nil {
		t.Fatalf("LoadSnapshot error: %v", err)
	}
	if snap.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestIntegrationListSnapshots_NilStore(t *testing.T) {
	var store *Store
	_, err := store.ListSnapshots(10)
	if err == nil {
		t.Error("expected error for nil store")
	}
}

// === Integration: Snapshot Structure ===

func TestIntegrationSnapshot_Structure(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	expectedHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "test", MAC: "aa:bb:cc:dd:ee:ff", DeviceType: "router"},
	}

	err = store.SaveSnapshot("struct-test", time.Now(), expectedHosts)
	if err != nil {
		t.Fatalf("SaveSnapshot error: %v", err)
	}

	snap, err := store.LoadSnapshot("struct-test")
	if err != nil {
		t.Fatalf("LoadSnapshot error: %v", err)
	}

	if snap.ID != "struct-test" {
		t.Errorf("expected ID 'struct-test', got %q", snap.ID)
	}
	if len(snap.Hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(snap.Hosts))
	}
	if snap.Hosts[0].IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", snap.Hosts[0].IP)
	}
	if snap.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

// === Integration: Diff Sorting ===

func TestIntegrationDiff_Sorting(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	// Scan A: hosts with IPs 3, 2, 1
	hostsA := []scanner.Result{
		{IP: "192.168.1.3", MAC: "aa:bb:cc:dd:ee:01"},
		{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02"},
		{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:03"},
	}
	err = store.SaveSnapshot("sort-a", time.Now(), hostsA)
	if err != nil {
		t.Fatalf("SaveSnapshot A error: %v", err)
	}

	// Scan B: all missing
	err = store.SaveSnapshot("sort-b", time.Now(), nil)
	if err != nil {
		t.Fatalf("SaveSnapshot B error: %v", err)
	}

	diff, err := store.Diff("sort-a", "sort-b")
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	// Missing should be sorted by IP
	if len(diff.Missing) != 3 {
		t.Fatalf("expected 3 missing hosts, got %d", len(diff.Missing))
	}
	if diff.Missing[0].IP != "192.168.1.1" {
		t.Errorf("expected first missing IP '192.168.1.1', got %q", diff.Missing[0].IP)
	}
	if diff.Missing[1].IP != "192.168.1.2" {
		t.Errorf("expected second missing IP '192.168.1.2', got %q", diff.Missing[1].IP)
	}
	if diff.Missing[2].IP != "192.168.1.3" {
		t.Errorf("expected third missing IP '192.168.1.3', got %q", diff.Missing[2].IP)
	}
}

// === Integration: MAC-based Key ===

func TestIntegrationHostKey_MAC(t *testing.T) {
	host := scanner.Result{
		IP:  "192.168.1.1",
		MAC: "aa:bb:cc:dd:ee:ff",
	}

	key := hostKey(host)
	if key != "mac:aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected 'mac:aa:bb:cc:dd:ee:ff', got %q", key)
	}
}

func TestIntegrationHostKey_IP(t *testing.T) {
	host := scanner.Result{
		IP:  "192.168.1.1",
		MAC: "", // No MAC, should use IP key
	}

	key := hostKey(host)
	if key != "ip:192.168.1.1" {
		t.Errorf("expected 'ip:192.168.1.1', got %q", key)
	}
}

func TestIntegrationHostKey_Empty(t *testing.T) {
	host := scanner.Result{
		IP:  "",
		MAC: "",
	}

	key := hostKey(host)
	if key != "" {
		t.Errorf("expected empty key, got %q", key)
	}
}

// === Integration: Changed Fields Detection ===

func TestIntegrationChangedFields_IP(t *testing.T) {
	a := scanner.Result{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff"}
	b := scanner.Result{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:ff"}

	fields := changedFields(a, b)
	if len(fields) == 0 {
		t.Error("expected at least one changed field")
	}
	hasIP := false
	for _, f := range fields {
		if f == "ip" {
			hasIP = true
		}
	}
	if !hasIP {
		t.Error("expected 'ip' in changed fields")
	}
}

func TestIntegrationChangedFields_MAC(t *testing.T) {
	a := scanner.Result{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:ff"}
	b := scanner.Result{IP: "192.168.1.1", MAC: "11:22:33:44:55:66"}

	fields := changedFields(a, b)
	hasMAC := false
	for _, f := range fields {
		if f == "mac" {
			hasMAC = true
		}
	}
	if !hasMAC {
		t.Error("expected 'mac' in changed fields")
	}
}

func TestIntegrationChangedFields_Hostname(t *testing.T) {
	a := scanner.Result{IP: "192.168.1.1", Hostname: "old"}
	b := scanner.Result{IP: "192.168.1.1", Hostname: "new"}

	fields := changedFields(a, b)
	hasHostname := false
	for _, f := range fields {
		if f == "hostname" {
			hasHostname = true
		}
	}
	if !hasHostname {
		t.Error("expected 'hostname' in changed fields")
	}
}

func TestIntegrationChangedFields_NoChange(t *testing.T) {
	a := scanner.Result{IP: "192.168.1.1", Hostname: "test", MAC: "aa:bb:cc:dd:ee:ff"}
	b := scanner.Result{IP: "192.168.1.1", Hostname: "test", MAC: "aa:bb:cc:dd:ee:ff"}

	fields := changedFields(a, b)
	if len(fields) != 0 {
		t.Errorf("expected 0 changed fields, got %d: %v", len(fields), fields)
	}
}

// === Integration: Port Comparison ===

func TestIntegrationPortsEqual_Same(t *testing.T) {
	a := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "tcp"},
		{Port: 80, State: "open", Protocol: "tcp"},
	}
	b := []scanner.PortInfo{
		{Port: 80, State: "open", Protocol: "tcp"},
		{Port: 22, State: "open", Protocol: "tcp"},
	}

	if !portsEqual(a, b) {
		t.Error("expected ports to be equal (order independent)")
	}
}

func TestIntegrationPortsEqual_Different(t *testing.T) {
	a := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "tcp"},
	}
	b := []scanner.PortInfo{
		{Port: 80, State: "open", Protocol: "tcp"},
	}

	if portsEqual(a, b) {
		t.Error("expected ports to be different")
	}
}

func TestIntegrationPortsEqual_DifferentCount(t *testing.T) {
	a := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "tcp"},
	}
	b := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "tcp"},
		{Port: 80, State: "open", Protocol: "tcp"},
	}

	if portsEqual(a, b) {
		t.Error("expected ports to be different (different count)")
	}
}

// === Integration: Subdirectory Creation ===

func TestIntegrationStore_SubdirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "nested", "inventory.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer store.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected database file to exist after Open")
	}
}

// === Integration: Whitespace Handling in ChangedFields ===

func TestIntegrationChangedFields_Whitespace(t *testing.T) {
	a := scanner.Result{IP: "192.168.1.1", Hostname: "  test  "}
	b := scanner.Result{IP: "192.168.1.1", Hostname: "test"}

	fields := changedFields(a, b)
	// Whitespace should be trimmed, so no change detected
	if len(fields) != 0 {
		t.Errorf("expected 0 changed fields (whitespace trimmed), got %d: %v", len(fields), fields)
	}
}
