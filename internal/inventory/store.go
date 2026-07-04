package inventory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"network-scanner/internal/comparator"
	"network-scanner/internal/scanner"
)

type Store struct {
	db *sql.DB
}

type Snapshot struct {
	ID        string
	Timestamp time.Time
	Hosts     []scanner.Result
}

type ChangedHost struct {
	Key          string
	Before       scanner.Result
	After        scanner.Result
	ChangedField []string
}

type DiffResult struct {
	ScanIDA string
	ScanIDB string
	New     []scanner.Result
	Missing []scanner.Result
	Changed []ChangedHost
}

func Open(path string) (*Store, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("inventory db path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create inventory dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	s := &Store{db: db}
	if err := s.ensureSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) SaveSnapshot(scanID string, ts time.Time, hosts []scanner.Result) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("inventory store is not initialized")
	}
	scanID = strings.TrimSpace(scanID)
	if scanID == "" {
		return fmt.Errorf("scanID is required")
	}
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	payload, err := json.Marshal(hosts)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT OR REPLACE INTO snapshots(id, created_at, data) VALUES(?, ?, ?)`,
		scanID,
		ts.UTC().Format(time.RFC3339Nano),
		string(payload),
	)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}
	return nil
}

func (s *Store) LoadSnapshot(scanID string) (Snapshot, error) {
	if s == nil || s.db == nil {
		return Snapshot{}, fmt.Errorf("inventory store is not initialized")
	}
	scanID = strings.TrimSpace(scanID)
	if scanID == "" {
		return Snapshot{}, fmt.Errorf("scanID is required")
	}
	var createdAtRaw string
	var payload string
	row := s.db.QueryRow(`SELECT created_at, data FROM snapshots WHERE id = ?`, scanID)
	if err := row.Scan(&createdAtRaw, &payload); err != nil {
		if err == sql.ErrNoRows {
			return Snapshot{}, fmt.Errorf("snapshot %q not found", scanID)
		}
		return Snapshot{}, fmt.Errorf("load snapshot: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtRaw)
	if err != nil {
		createdAt = time.Time{}
	}
	var hosts []scanner.Result
	if err := json.Unmarshal([]byte(payload), &hosts); err != nil {
		return Snapshot{}, fmt.Errorf("decode snapshot payload: %w", err)
	}
	return Snapshot{
		ID:        scanID,
		Timestamp: createdAt,
		Hosts:     hosts,
	}, nil
}

func (s *Store) ListSnapshots(limit int) ([]Snapshot, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("inventory store is not initialized")
	}
	q := `SELECT id, created_at, data FROM snapshots ORDER BY created_at DESC`
	args := make([]interface{}, 0)
	if limit > 0 {
		q += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("query snapshots: %w", err)
	}
	defer rows.Close()
	out := make([]Snapshot, 0)
	for rows.Next() {
		var id, createdAtRaw, payload string
		if err := rows.Scan(&id, &createdAtRaw, &payload); err != nil {
			return nil, fmt.Errorf("scan snapshot row: %w", err)
		}
		t, _ := time.Parse(time.RFC3339Nano, createdAtRaw)
		var hosts []scanner.Result
		if err := json.Unmarshal([]byte(payload), &hosts); err != nil {
			return nil, fmt.Errorf("decode snapshot payload: %w", err)
		}
		out = append(out, Snapshot{ID: id, Timestamp: t, Hosts: hosts})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate snapshots: %w", err)
	}
	return out, nil
}

func (s *Store) Diff(scanIDA, scanIDB string) (DiffResult, error) {
	a, err := s.LoadSnapshot(scanIDA)
	if err != nil {
		return DiffResult{}, err
	}
	b, err := s.LoadSnapshot(scanIDB)
	if err != nil {
		return DiffResult{}, err
	}
	aMap := hostsByKey(a.Hosts)
	bMap := hostsByKey(b.Hosts)

	res := DiffResult{
		ScanIDA: a.ID,
		ScanIDB: b.ID,
		New:     make([]scanner.Result, 0),
		Missing: make([]scanner.Result, 0),
		Changed: make([]ChangedHost, 0),
	}

	for key, hostB := range bMap {
		hostA, ok := aMap[key]
		if !ok {
			res.New = append(res.New, hostB)
			continue
		}
		fields := changedFields(hostA, hostB)
		if len(fields) > 0 {
			res.Changed = append(res.Changed, ChangedHost{
				Key:          key,
				Before:       hostA,
				After:        hostB,
				ChangedField: fields,
			})
		}
	}
	for key, hostA := range aMap {
		if _, ok := bMap[key]; !ok {
			res.Missing = append(res.Missing, hostA)
		}
	}
	sort.Slice(res.New, func(i, j int) bool { return res.New[i].IP < res.New[j].IP })
	sort.Slice(res.Missing, func(i, j int) bool { return res.Missing[i].IP < res.Missing[j].IP })
	sort.Slice(res.Changed, func(i, j int) bool { return res.Changed[i].Key < res.Changed[j].Key })
	return res, nil
}

func hostsByKey(hosts []scanner.Result) map[string]scanner.Result {
	out := make(map[string]scanner.Result, len(hosts))
	for _, h := range hosts {
		key := hostKey(h)
		if key == "" {
			continue
		}
		out[key] = h
	}
	return out
}

func hostKey(h scanner.Result) string {
	mac := strings.TrimSpace(strings.ToLower(h.MAC))
	if mac != "" {
		return "mac:" + mac
	}
	ip := strings.TrimSpace(strings.ToLower(h.IP))
	if ip != "" {
		return "ip:" + ip
	}
	return ""
}

func changedFields(a, b scanner.Result) []string {
	fields := make([]string, 0)
	if strings.TrimSpace(a.IP) != strings.TrimSpace(b.IP) {
		fields = append(fields, "ip")
	}
	if strings.TrimSpace(a.MAC) != strings.TrimSpace(b.MAC) {
		fields = append(fields, "mac")
	}
	if strings.TrimSpace(a.Hostname) != strings.TrimSpace(b.Hostname) {
		fields = append(fields, "hostname")
	}
	if strings.TrimSpace(a.DeviceType) != strings.TrimSpace(b.DeviceType) {
		fields = append(fields, "device_type")
	}
	if strings.TrimSpace(a.DeviceVendor) != strings.TrimSpace(b.DeviceVendor) {
		fields = append(fields, "device_vendor")
	}
	if strings.TrimSpace(a.GuessOS) != strings.TrimSpace(b.GuessOS) {
		fields = append(fields, "guess_os")
	}
	if !portsEqual(a.Ports, b.Ports) {
		fields = append(fields, "ports")
	}
	return fields
}

func portsEqual(a, b []scanner.PortInfo) bool {
	if len(a) != len(b) {
		return false
	}
	normalize := func(in []scanner.PortInfo) []string {
		out := make([]string, 0, len(in))
		for _, p := range in {
			out = append(out, fmt.Sprintf("%d/%s/%s", p.Port, strings.ToLower(p.Protocol), strings.ToLower(p.State)))
		}
		sort.Strings(out)
		return out
	}
	na := normalize(a)
	nb := normalize(b)
	for i := range na {
		if na[i] != nb[i] {
			return false
		}
	}
	return true
}

func (s *Store) ensureSchema() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS snapshots (
	id TEXT PRIMARY KEY,
	created_at TEXT NOT NULL,
	data TEXT NOT NULL
);
`)
	if err != nil {
		return fmt.Errorf("create inventory schema: %w", err)
	}
	return nil
}



// GetScanHistory возвращает историю сканирований с metadata
func (s *Store) GetScanHistory(limit int) ([]comparator.ScanHistoryEntry, []scanner.Result, error) {
	if s == nil || s.db == nil {
		return nil, nil, fmt.Errorf("inventory store is not initialized")
	}
	snapshots, err := s.ListSnapshots(limit)
	if err != nil {
		return nil, nil, err
	}
	history := make([]comparator.ScanHistoryEntry, 0, len(snapshots))
	allHosts := make([]scanner.Result, 0)
	for _, snap := range snapshots {
		entry := comparator.ScanHistoryEntry{
			ID:        snap.ID,
			HostCount: len(snap.Hosts),
			StartedAt: snap.Timestamp,
			Completed: snap.Timestamp,
			Ports:     make(map[string]int),
			OSMap:     make(map[string]int),
			VendorMap: make(map[string]int),
		}
		for _, h := range snap.Hosts {
			if h.GuessOS != "" {
				entry.OSMap[h.GuessOS]++
			}
			if h.DeviceVendor != "" {
				entry.VendorMap[h.DeviceVendor]++
			}
			for _, p := range h.Ports {
				if strings.EqualFold(p.State, "open") {
					key := fmt.Sprintf("%d/%s", p.Port, p.Protocol)
					entry.Ports[key]++
				}
			}
		}
		history = append(history, entry)
		allHosts = append(allHosts, snap.Hosts...)
	}
	return history, allHosts, nil
}

// CompareSnapshotsByName сравнивает два снапшота по ID и возвращает ComparisonResult
func (s *Store) CompareSnapshotsByName(scanIDA, scanIDB string) (*comparator.ComparisonResult, error) {
	a, err := s.LoadSnapshot(scanIDA)
	if err != nil {
		return nil, fmt.Errorf("load snapshot A: %w", err)
	}
	b, err := s.LoadSnapshot(scanIDB)
	if err != nil {
		return nil, fmt.Errorf("load snapshot B: %w", err)
	}
	return comparator.CompareSnapshots(scanIDA, scanIDB, a.Hosts, b.Hosts), nil
}
