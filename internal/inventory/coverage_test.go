package inventory

import (
	"path/filepath"
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

// ============================================================================
// Open — edge cases
// ============================================================================

func TestOpen_EmptyPath(t *testing.T) {
	_, err := Open("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestOpen_WhitespacePath(t *testing.T) {
	_, err := Open("   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only path")
	}
}

func TestOpen_ValidPath(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
}

// ============================================================================
// Close — edge cases
// ============================================================================

func TestClose_NilStore(t *testing.T) {
	var s *Store
	if err := s.Close(); err != nil {
		t.Fatalf("expected nil error for nil store close, got %v", err)
	}
}

func TestClose_NilDB(t *testing.T) {
	s := &Store{db: nil}
	if err := s.Close(); err != nil {
		t.Fatalf("expected nil error for nil db close, got %v", err)
	}
}

// ============================================================================
// SaveSnapshot — edge cases
// ============================================================================

func TestSaveSnapshot_NilStore(t *testing.T) {
	var s *Store
	err := s.SaveSnapshot("test", time.Now(), nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestSaveSnapshot_EmptyScanID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	err := store.SaveSnapshot("", time.Now(), nil)
	if err == nil {
		t.Fatal("expected error for empty scanID")
	}
}

func TestSaveSnapshot_WhitespaceScanID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	err := store.SaveSnapshot("   ", time.Now(), nil)
	if err == nil {
		t.Fatal("expected error for whitespace scanID")
	}
}

func TestSaveSnapshot_ZeroTime(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	hosts := []scanner.Result{{IP: "10.0.0.1"}}
	err := store.SaveSnapshot("scan-zero", time.Time{}, hosts)
	if err != nil {
		t.Fatalf("expected no error for zero time, got %v", err)
	}
}

// ============================================================================
// LoadSnapshot — edge cases
// ============================================================================

func TestLoadSnapshot_NilStore(t *testing.T) {
	var s *Store
	_, err := s.LoadSnapshot("test")
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestLoadSnapshot_EmptyScanID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	_, err := store.LoadSnapshot("")
	if err == nil {
		t.Fatal("expected error for empty scanID")
	}
}

func TestLoadSnapshot_NotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	_, err := store.LoadSnapshot("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot")
	}
}

// ============================================================================
// ListSnapshots — 0% → 100%
// ============================================================================

func TestListSnapshots_NilStore(t *testing.T) {
	var s *Store
	_, err := s.ListSnapshots(10)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestListSnapshots_Empty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	snaps, err := store.ListSnapshots(0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(snaps) != 0 {
		t.Fatalf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestListSnapshots_WithLimit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	// Save some snapshots
	for i := 0; i < 5; i++ {
		err := store.SaveSnapshot(
			"scan-"+string(rune('a'+i)),
			time.Now().Add(time.Duration(i)*time.Second),
			[]scanner.Result{{IP: "10.0.0.1"}},
		)
		if err != nil {
			t.Fatalf("save snapshot: %v", err)
		}
	}

	snaps, err := store.ListSnapshots(3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(snaps) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(snaps))
	}
}

func TestListSnapshots_NoLimit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	store.SaveSnapshot("scan-a", time.Now(), []scanner.Result{{IP: "10.0.0.1"}})
	store.SaveSnapshot("scan-b", time.Now(), []scanner.Result{{IP: "10.0.0.2"}})

	snaps, err := store.ListSnapshots(0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
}

// ============================================================================
// hostKey — 42.9% → 100%
// ============================================================================

func TestHostKey_MAC(t *testing.T) {
	h := scanner.Result{MAC: "AA:BB:CC:DD:EE:FF"}
	key := hostKey(h)
	if key != "mac:aa:bb:cc:dd:ee:ff" {
		t.Fatalf("expected mac key, got %q", key)
	}
}

func TestHostKey_IPOnly(t *testing.T) {
	h := scanner.Result{IP: "192.168.1.1"}
	key := hostKey(h)
	if key != "ip:192.168.1.1" {
		t.Fatalf("expected ip key, got %q", key)
	}
}

func TestHostKey_Empty(t *testing.T) {
	h := scanner.Result{}
	key := hostKey(h)
	if key != "" {
		t.Fatalf("expected empty key, got %q", key)
	}
}

func TestHostKey_WhitespaceMAC(t *testing.T) {
	h := scanner.Result{MAC: "  ", IP: "10.0.0.1"}
	key := hostKey(h)
	if key != "ip:10.0.0.1" {
		t.Fatalf("expected ip key for whitespace MAC, got %q", key)
	}
}

func TestHostKey_WhitespaceBoth(t *testing.T) {
	h := scanner.Result{MAC: "  ", IP: "  "}
	key := hostKey(h)
	if key != "" {
		t.Fatalf("expected empty key, got %q", key)
	}
}

// ============================================================================
// hostsByKey — edge cases
// ============================================================================

func TestHostsByKey_Empty(t *testing.T) {
	m := hostsByKey([]scanner.Result{})
	if len(m) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(m))
	}
}

func TestHostsByKey_SkipsEmptyKey(t *testing.T) {
	hosts := []scanner.Result{
		{IP: "10.0.0.1"},
		{}, // empty key, should be skipped
	}
	m := hostsByKey(hosts)
	if len(m) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m))
	}
}

// ============================================================================
// changedFields — 62.5% → 100%
// ============================================================================

func TestChangedFields_AllChanged(t *testing.T) {
	a := scanner.Result{IP: "10.0.0.1", MAC: "aa", Hostname: "h1", DeviceType: "router", DeviceVendor: "cisco", GuessOS: "ios", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}}
	b := scanner.Result{IP: "10.0.0.2", MAC: "bb", Hostname: "h2", DeviceType: "switch", DeviceVendor: "juniper", GuessOS: "junos", Ports: []scanner.PortInfo{{Port: 443, Protocol: "tcp", State: "open"}}}
	fields := changedFields(a, b)
	if len(fields) != 7 {
		t.Fatalf("expected 7 changed fields, got %d: %v", len(fields), fields)
	}
}

func TestChangedFields_NoChanges(t *testing.T) {
	a := scanner.Result{IP: "10.0.0.1", MAC: "aa", Hostname: "h1"}
	b := scanner.Result{IP: "10.0.0.1", MAC: "aa", Hostname: "h1"}
	fields := changedFields(a, b)
	if len(fields) != 0 {
		t.Fatalf("expected 0 changed fields, got %d: %v", len(fields), fields)
	}
}

func TestChangedFields_OnlyPorts(t *testing.T) {
	a := scanner.Result{IP: "10.0.0.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}}
	b := scanner.Result{IP: "10.0.0.1", Ports: []scanner.PortInfo{{Port: 443, Protocol: "tcp", State: "open"}}}
	fields := changedFields(a, b)
	if len(fields) != 1 || fields[0] != "ports" {
		t.Fatalf("expected only ports changed, got %v", fields)
	}
}

func TestChangedFields_DeviceTypeChanged(t *testing.T) {
	a := scanner.Result{IP: "10.0.0.1", DeviceType: "router"}
	b := scanner.Result{IP: "10.0.0.1", DeviceType: "switch"}
	fields := changedFields(a, b)
	if len(fields) != 1 || fields[0] != "device_type" {
		t.Fatalf("expected device_type changed, got %v", fields)
	}
}

func TestChangedFields_DeviceVendorChanged(t *testing.T) {
	a := scanner.Result{IP: "10.0.0.1", DeviceVendor: "cisco"}
	b := scanner.Result{IP: "10.0.0.1", DeviceVendor: "juniper"}
	fields := changedFields(a, b)
	if len(fields) != 1 || fields[0] != "device_vendor" {
		t.Fatalf("expected device_vendor changed, got %v", fields)
	}
}

func TestChangedFields_GuessOSChanged(t *testing.T) {
	a := scanner.Result{IP: "10.0.0.1", GuessOS: "linux"}
	b := scanner.Result{IP: "10.0.0.1", GuessOS: "windows"}
	fields := changedFields(a, b)
	if len(fields) != 1 || fields[0] != "guess_os" {
		t.Fatalf("expected guess_os changed, got %v", fields)
	}
}

// ============================================================================
// portsEqual — edge cases
// ============================================================================

func TestPortsEqual_DifferentLength(t *testing.T) {
	a := []scanner.PortInfo{{Port: 80}}
	b := []scanner.PortInfo{{Port: 80}, {Port: 443}}
	if portsEqual(a, b) {
		t.Fatal("expected false for different length ports")
	}
}

func TestPortsEqual_SameUnordered(t *testing.T) {
	a := []scanner.PortInfo{{Port: 80, Protocol: "TCP", State: "OPEN"}, {Port: 443, Protocol: "tcp", State: "open"}}
	b := []scanner.PortInfo{{Port: 443, Protocol: "tcp", State: "open"}, {Port: 80, Protocol: "tcp", State: "open"}}
	if !portsEqual(a, b) {
		t.Fatal("expected true for same ports in different order with case differences")
	}
}

func TestPortsEqual_DifferentPort(t *testing.T) {
	a := []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}
	b := []scanner.PortInfo{{Port: 443, Protocol: "tcp", State: "open"}}
	if portsEqual(a, b) {
		t.Fatal("expected false for different ports")
	}
}

func TestPortsEqual_Empty(t *testing.T) {
	if !portsEqual([]scanner.PortInfo{}, []scanner.PortInfo{}) {
		t.Fatal("expected true for empty ports")
	}
}

// ============================================================================
// GetScanHistory — 0% → 100%
// ============================================================================

func TestGetScanHistory_NilStore(t *testing.T) {
	var s *Store
	_, _, err := s.GetScanHistory(10)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestGetScanHistory_Empty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	history, hosts, err := store.GetScanHistory(10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("expected 0 history entries, got %d", len(history))
	}
	if len(hosts) != 0 {
		t.Fatalf("expected 0 hosts, got %d", len(hosts))
	}
}

func TestGetScanHistory_WithData(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	hosts := []scanner.Result{
		{IP: "10.0.0.1", GuessOS: "linux", DeviceVendor: "dell", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}},
		{IP: "10.0.0.2", GuessOS: "windows", DeviceVendor: "hp", Ports: []scanner.PortInfo{{Port: 443, Protocol: "tcp", State: "closed"}}},
	}
	err := store.SaveSnapshot("scan-1", time.Now(), hosts)
	if err != nil {
		t.Fatalf("save snapshot: %v", err)
	}

	history, allHosts, err := store.GetScanHistory(10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}
	if history[0].HostCount != 2 {
		t.Fatalf("expected 2 hosts in history entry, got %d", history[0].HostCount)
	}
	if len(allHosts) != 2 {
		t.Fatalf("expected 2 total hosts, got %d", len(allHosts))
	}
	// Check OS map
	if history[0].OSMap["linux"] != 1 {
		t.Fatalf("expected linux=1 in OS map, got %d", history[0].OSMap["linux"])
	}
	// Check vendor map
	if history[0].VendorMap["dell"] != 1 {
		t.Fatalf("expected dell=1 in vendor map, got %d", history[0].VendorMap["dell"])
	}
	// Check ports map (only open ports)
	if history[0].Ports["80/tcp"] != 1 {
		t.Fatalf("expected 80/tcp=1 in ports map, got %d", history[0].Ports["80/tcp"])
	}
}

// ============================================================================
// CompareSnapshotsByName — 0% → 100%
// ============================================================================

func TestCompareSnapshotsByName_NilStore(t *testing.T) {
	var s *Store
	_, err := s.CompareSnapshotsByName("a", "b")
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestCompareSnapshotsByName_NotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	_, err := store.CompareSnapshotsByName("nonexistent-a", "nonexistent-b")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshots")
	}
}

func TestCompareSnapshotsByName_Success(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	snapA := []scanner.Result{
		{IP: "10.0.0.1", MAC: "aa:bb:cc:dd:ee:01"},
	}
	snapB := []scanner.Result{
		{IP: "10.0.0.1", MAC: "aa:bb:cc:dd:ee:01"},
		{IP: "10.0.0.2", MAC: "aa:bb:cc:dd:ee:02"},
	}

	store.SaveSnapshot("scan-a", time.Now(), snapA)
	store.SaveSnapshot("scan-b", time.Now(), snapB)

	result, err := store.CompareSnapshotsByName("scan-a", "scan-b")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// ============================================================================
// Diff — edge cases
// ============================================================================

func TestDiff_SnapshotNotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	store.SaveSnapshot("scan-a", time.Now(), []scanner.Result{{IP: "10.0.0.1"}})

	_, err := store.Diff("scan-a", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot B")
	}

	_, err = store.Diff("nonexistent", "scan-a")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot A")
	}
}

func TestDiff_NoChanges(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, _ := Open(dbPath)
	defer store.Close()

	hosts := []scanner.Result{{IP: "10.0.0.1", MAC: "aa:bb:cc:dd:ee:01"}}
	store.SaveSnapshot("scan-a", time.Now(), hosts)
	store.SaveSnapshot("scan-b", time.Now(), hosts)

	diff, err := store.Diff("scan-a", "scan-b")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(diff.New) != 0 {
		t.Fatalf("expected 0 new hosts, got %d", len(diff.New))
	}
	if len(diff.Missing) != 0 {
		t.Fatalf("expected 0 missing hosts, got %d", len(diff.Missing))
	}
	if len(diff.Changed) != 0 {
		t.Fatalf("expected 0 changed hosts, got %d", len(diff.Changed))
	}
}

func TestDiff_NilStore(t *testing.T) {
	var s *Store
	_, err := s.Diff("a", "b")
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}
