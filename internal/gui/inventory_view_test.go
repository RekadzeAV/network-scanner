package gui

import (
	"testing"
	"time"

	"network-scanner/internal/inventory"
)

// --- inventory_view.go tests ---

func TestSnapshotOptionLabel_WithTimestamp(t *testing.T) {
	// Используем локальное время
	ts := time.Date(2025, 8, 14, 10, 30, 0, 0, time.Local)
	snap := inventory.Snapshot{
		ID:        "snap-001",
		Timestamp: ts,
	}
	label := snapshotOptionLabel(snap)
	// Формат: "ID (YYYY-MM-DD HH:MM:SS)"
	if len(label) < 20 || label[0:8] != "snap-001" {
		t.Errorf("expected label starting with 'snap-001', got %q", label)
	}
}

func TestSnapshotOptionLabel_WithoutTimestamp(t *testing.T) {
	snap := inventory.Snapshot{
		ID:        "snap-002",
		Timestamp: time.Time{},
	}
	label := snapshotOptionLabel(snap)
	if label != "snap-002" {
		t.Errorf("expected 'snap-002', got %q", label)
	}
}

func TestSnapshotOptionLabel_EmptyID(t *testing.T) {
	snap := inventory.Snapshot{}
	label := snapshotOptionLabel(snap)
	if label != "" {
		t.Errorf("expected empty, got %q", label)
	}
}

func TestSnapshotOptionLabel_NilSafe(t *testing.T) {
	// Snapshot — struct, не pointer, nil невозможен
	snap := inventory.Snapshot{ID: "snap-003"}
	label := snapshotOptionLabel(snap)
	if label == "" {
		t.Error("expected non-empty label")
	}
}

func TestParseSnapshotID_Basic(t *testing.T) {
	option := "snap-001 (2025-08-14 13:30:00)"
	id := parseSnapshotID(option)
	if id != "snap-001" {
		t.Errorf("expected 'snap-001', got %q", id)
	}
}

func TestParseSnapshotID_NoTimestamp(t *testing.T) {
	option := "snap-002"
	id := parseSnapshotID(option)
	if id != "snap-002" {
		t.Errorf("expected 'snap-002', got %q", id)
	}
}

func TestParseSnapshotID_Empty(t *testing.T) {
	id := parseSnapshotID("")
	if id != "" {
		t.Errorf("expected empty, got %q", id)
	}
}

func TestParseSnapshotID_WhitespaceBeforeTimestamp(t *testing.T) {
	option := "snap-003  (2025-08-14 13:30:00)"
	id := parseSnapshotID(option)
	if id != "snap-003" {
		t.Errorf("expected 'snap-003', got %q", id)
	}
}

func TestParseSnapshotID_IDWithParenthesis(t *testing.T) {
	option := "snap (test) (2025-08-14 13:30:00)"
	id := parseSnapshotID(option)
	// parseSnapshotID берёт первую часть до " ("
	if id != "snap" {
		t.Errorf("expected 'snap', got %q", id)
	}
}

func TestParseSnapshotID_MultipleTimestamps(t *testing.T) {
	option := "snap-004 (2025-08-14 13:30:00) extra"
	id := parseSnapshotID(option)
	// Должен взять первую часть до " ("
	if id != "snap-004" {
		t.Errorf("expected 'snap-004', got %q", id)
	}
}
