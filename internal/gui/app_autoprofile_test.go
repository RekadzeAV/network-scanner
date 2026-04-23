package gui

import (
	"strings"
	"testing"
)

func TestAutoScanProfile_NoChangesForSmallSubnet(t *testing.T) {
	portRange, threads, note := autoScanProfile("192.168.1.0/24", "1-65535", 120)
	if portRange != "1-65535" {
		t.Fatalf("unexpected port range: %s", portRange)
	}
	if threads != 120 {
		t.Fatalf("unexpected threads: %d", threads)
	}
	if note != "" {
		t.Fatalf("unexpected note: %s", note)
	}
}

func TestAutoScanProfile_CapsLargeSubnet(t *testing.T) {
	portRange, threads, note := autoScanProfile("10.0.0.0/21", "1-65535", 120)
	if portRange != "1-1024" {
		t.Fatalf("expected capped range 1-1024, got %s", portRange)
	}
	if threads != 40 {
		t.Fatalf("expected capped threads 40, got %d", threads)
	}
	if note == "" {
		t.Fatalf("expected non-empty note")
	}
	if !strings.Contains(note, "ports: 1-65535 -> 1-1024") || !strings.Contains(note, "threads: 120 -> 40") {
		t.Fatalf("expected detailed note, got %q", note)
	}
}

func TestAutoScanProfile_CapsVeryLargeSubnet(t *testing.T) {
	portRange, threads, note := autoScanProfile("10.0.0.0/20", "1-65535", 300)
	if portRange != "1-512" {
		t.Fatalf("expected capped range 1-512, got %s", portRange)
	}
	if threads != 24 {
		t.Fatalf("expected capped threads 24, got %d", threads)
	}
	if note == "" {
		t.Fatalf("expected non-empty note")
	}
	if !strings.Contains(note, "ports: 1-65535 -> 1-512") || !strings.Contains(note, "threads: 300 -> 24") {
		t.Fatalf("expected detailed note, got %q", note)
	}
}

