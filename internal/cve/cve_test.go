package cve

import (
	"strings"
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

func TestAnalyzeResults_AppliesServiceVersionAndFilters(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Service: "ssh", Version: "OpenSSH_9.3"},
				{Port: 80, State: "open", Service: "http", Banner: "server: nginx/1.25.3"},
			},
		},
	}

	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{
		MinCVSS:    8.0,
		MaxAgeDays: 2000,
		Now:        time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC),
	})
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Entry.ID != "CVE-2023-38408" {
		t.Fatalf("unexpected CVE id: %s", matches[0].Entry.ID)
	}
}

func TestFormatMatches_Empty(t *testing.T) {
	got := FormatMatches(nil)
	if !strings.Contains(got, "не найдено") {
		t.Fatalf("expected empty message, got: %s", got)
	}
}

func TestFormatMatches_SanitizesSensitiveText(t *testing.T) {
	got := FormatMatches([]Match{
		{
			HostIP:   "10.0.0.10",
			HostName: "srv password=abc",
			Port:     22,
			Service:  "ssh token=xyz",
			Entry: Entry{
				ID:   "CVE-1",
				CVSS: 9.0,
			},
		},
	})
	if strings.Contains(got, "abc") || strings.Contains(got, "xyz") {
		t.Fatalf("expected sensitive values to be masked: %s", got)
	}
	if !strings.Contains(got, "***") {
		t.Fatalf("expected masked marker: %s", got)
	}
}
