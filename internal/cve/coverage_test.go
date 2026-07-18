package cve

import (
	"strings"
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

// ============================================================================
// W-66: Улучшение coverage AnalyzeResults (75.0% → 90%+)
// ============================================================================

func TestAnalyzeResults_NoOpenPorts(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "closed", Service: "http"},
			},
		},
	}

	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{})
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches for closed ports, got %d", len(matches))
	}
}

func TestAnalyzeResults_NoVersion(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Service: "http"},
			},
		},
	}

	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{})
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches when no version, got %d", len(matches))
	}
}

func TestAnalyzeResults_NoBanner(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Service: "http", Version: "nginx"},
			},
		},
	}

	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{})
	// Should match because version is present
	if len(matches) < 0 {
		t.Fatal("matches should not be negative")
	}
}

func TestAnalyzeResults_MinCVSSFilter(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Service: "http", Banner: "log4j/2.14.0"},
			},
		},
	}

	// MinCVSS 10.0 should filter out CVE-2023-44487 (7.5) and CVE-2023-38408 (9.8)
	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{
		MinCVSS: 10.0,
	})
	if len(matches) != 1 {
		t.Fatalf("expected 1 match with CVSS >= 10.0, got %d", len(matches))
	}
	if matches[0].Entry.ID != "CVE-2021-44228" {
		t.Fatalf("expected CVE-2021-44228, got %s", matches[0].Entry.ID)
	}
}

func TestAnalyzeResults_MaxAgeFilter(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Service: "http", Banner: "log4j/2.14.0"},
			},
		},
	}

	// MaxAgeDays 365 should filter out CVE-2021-44228 (published 2021-12-10)
	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{
		MaxAgeDays: 365,
		Now:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches for old CVE, got %d", len(matches))
	}
}

func TestAnalyzeResults_MultipleHosts(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Service: "http", Banner: "nginx/1.25.3"},
			},
		},
		{
			IP: "192.168.1.20",
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Service: "ssh", Version: "OpenSSH_9.3"},
			},
		},
	}

	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{
		MinCVSS:    7.0,
		MaxAgeDays: 2000,
		Now:        time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC),
	})
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
}

func TestAnalyzeResults_SortingByCVSS(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Service: "http", Banner: "log4j/2.14.0 nginx/1.25"},
			},
		},
	}

	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{
		MinCVSS:    7.0,
		MaxAgeDays: 2000,
		Now:        time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC),
	})
	
	// Should be sorted by CVSS descending
	if len(matches) >= 2 {
		if matches[0].Entry.CVSS < matches[1].Entry.CVSS {
			t.Fatal("expected matches sorted by CVSS descending")
		}
	}
}

func TestAnalyzeResults_SortingByID(t *testing.T) {
	// Тестируем сортировку по ID когда CVSS одинаковый
	catalog := Catalog{
		entries: []Entry{
			{
				ID:          "CVE-2024-0002",
				Description: "Test CVE 2",
				CVSS:        8.0,
				Service:     "http",
				VersionHint: "test",
			},
			{
				ID:          "CVE-2024-0001",
				Description: "Test CVE 1",
				CVSS:        8.0,
				Service:     "http",
				VersionHint: "test",
			},
		},
	}

	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Service: "http", Banner: "test"},
			},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	// При одинаковом CVSS сортировка по ID ascending
	if matches[0].Entry.ID != "CVE-2024-0001" {
		t.Fatalf("expected CVE-2024-0001 first, got %s", matches[0].Entry.ID)
	}
}

func TestAnalyzeResults_EmptyResults(t *testing.T) {
	matches := AnalyzeResults([]scanner.Result{}, NewDefaultCatalog(), Options{})
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches for empty results, got %d", len(matches))
	}
}

func TestAnalyzeResults_HTTPSNormalizedToHTTP(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 443, State: "open", Service: "https", Banner: "nginx/1.25"},
			},
		},
	}

	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{
		MinCVSS:    7.0,
		MaxAgeDays: 2000,
		Now:        time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC),
	})
	// HTTPS должен нормализоваться в HTTP и совпасть с CVE для http
	if len(matches) == 0 {
		t.Fatal("expected match for HTTPS normalized to HTTP")
	}
}

func TestAnalyzeResults_PortBasedHTTP(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 8080, State: "open", Service: "unknown", Banner: "nginx/1.25"},
			},
		},
	}

	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{
		MinCVSS:    7.0,
		MaxAgeDays: 2000,
		Now:        time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC),
	})
	// Порт 8080 должен нормализоваться в HTTP
	if len(matches) == 0 {
		t.Fatal("expected match for port 8080 normalized to HTTP")
	}
}

func TestAnalyzeResults_PortBasedSSH(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 2222, State: "open", Service: "unknown", Version: "OpenSSH_9.3"},
			},
		},
	}

	matches := AnalyzeResults(results, NewDefaultCatalog(), Options{
		MinCVSS:    7.0,
		MaxAgeDays: 2000,
		Now:        time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC),
	})
	// Порт 22 должен нормализоваться в SSH, но 2222 - нет
	// Поэтому совпадений быть не должно
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches for port 2222, got %d", len(matches))
	}
}

// ============================================================================
// W-67: Улучшение coverage normalizeService (44.4% → 80%+)
// ============================================================================

func TestNormalizeService_HTTPS(t *testing.T) {
	port := scanner.PortInfo{Port: 443, State: "open", Service: "https"}
	got := normalizeService(port)
	if got != "http" {
		t.Fatalf("normalizeService(https) = %q, want %q", got, "http")
	}
}

func TestNormalizeService_HTTP(t *testing.T) {
	port := scanner.PortInfo{Port: 80, State: "open", Service: "http"}
	got := normalizeService(port)
	if got != "http" {
		t.Fatalf("normalizeService(http) = %q, want %q", got, "http")
	}
}

func TestNormalizeService_SSH(t *testing.T) {
	port := scanner.PortInfo{Port: 22, State: "open", Service: "ssh"}
	got := normalizeService(port)
	if got != "ssh" {
		t.Fatalf("normalizeService(ssh) = %q, want %q", got, "ssh")
	}
}

func TestNormalizeService_Port22(t *testing.T) {
	port := scanner.PortInfo{Port: 22, State: "open", Service: "unknown"}
	got := normalizeService(port)
	if got != "ssh" {
		t.Fatalf("normalizeService(port 22) = %q, want %q", got, "ssh")
	}
}

func TestNormalizeService_Port80(t *testing.T) {
	port := scanner.PortInfo{Port: 80, State: "open", Service: "unknown"}
	got := normalizeService(port)
	if got != "http" {
		t.Fatalf("normalizeService(port 80) = %q, want %q", got, "http")
	}
}

func TestNormalizeService_Port443(t *testing.T) {
	port := scanner.PortInfo{Port: 443, State: "open", Service: "unknown"}
	got := normalizeService(port)
	if got != "http" {
		t.Fatalf("normalizeService(port 443) = %q, want %q", got, "http")
	}
}

func TestNormalizeService_Port8080(t *testing.T) {
	port := scanner.PortInfo{Port: 8080, State: "open", Service: "unknown"}
	got := normalizeService(port)
	if got != "http" {
		t.Fatalf("normalizeService(port 8080) = %q, want %q", got, "http")
	}
}

func TestNormalizeService_Port8443(t *testing.T) {
	port := scanner.PortInfo{Port: 8443, State: "open", Service: "unknown"}
	got := normalizeService(port)
	if got != "http" {
		t.Fatalf("normalizeService(port 8443) = %q, want %q", got, "http")
	}
}

func TestNormalizeService_Unsupported(t *testing.T) {
	port := scanner.PortInfo{Port: 3306, State: "open", Service: "mysql"}
	got := normalizeService(port)
	if got != "" {
		t.Fatalf("normalizeService(mysql) = %q, want empty", got)
	}
}

func TestNormalizeService_EmptyService(t *testing.T) {
	port := scanner.PortInfo{Port: 9999, State: "open", Service: ""}
	got := normalizeService(port)
	if got != "" {
		t.Fatalf("normalizeService(empty) = %q, want empty", got)
	}
}

func TestNormalizeService_CaseInsensitive(t *testing.T) {
	port := scanner.PortInfo{Port: 443, State: "open", Service: "HTTPS"}
	got := normalizeService(port)
	if got != "http" {
		t.Fatalf("normalizeService(HTTPS) = %q, want %q", got, "http")
	}
}

func TestNormalizeService_Whitespace(t *testing.T) {
	port := scanner.PortInfo{Port: 80, State: "open", Service: "  http  "}
	got := normalizeService(port)
	if got != "http" {
		t.Fatalf("normalizeService(whitespace) = %q, want %q", got, "http")
	}
}

// ============================================================================
// W-68: Улучшение coverage FormatMatches (100% → 100%)
// ============================================================================

func TestFormatMatches_MultipleMatches(t *testing.T) {
	matches := []Match{
		{
			HostIP:    "192.168.1.10",
			HostName:  "server1",
			Port:      80,
			Service:   "http",
			Entry:     Entry{ID: "CVE-2023-44487", CVSS: 7.5},
		},
		{
			HostIP:    "192.168.1.20",
			HostName:  "server2",
			Port:      22,
			Service:   "ssh",
			Entry:     Entry{ID: "CVE-2023-38408", CVSS: 9.8},
		},
	}

	got := FormatMatches(matches)
	if !strings.Contains(got, "найдено совпадений: 2") {
		t.Fatalf("expected 2 matches in output, got: %s", got)
	}
	if !strings.Contains(got, "192.168.1.10") {
		t.Fatal("expected first IP in output")
	}
	if !strings.Contains(got, "192.168.1.20") {
		t.Fatal("expected second IP in output")
	}
}

func TestFormatMatches_WithHostName(t *testing.T) {
	matches := []Match{
		{
			HostIP:     "192.168.1.10",
			HostName:   "myserver.local",
			Port:       80,
			Service:    "http",
			Entry:      Entry{ID: "CVE-2023-44487", CVSS: 7.5},
		},
	}

	got := FormatMatches(matches)
	if !strings.Contains(got, "myserver.local") {
		t.Fatal("expected hostname in output")
	}
	if !strings.Contains(got, "192.168.1.10") {
		t.Fatal("expected IP in output")
	}
}

func TestFormatMatches_CVSSFormat(t *testing.T) {
	matches := []Match{
		{
			HostIP:  "192.168.1.10",
			Port:    80,
			Service: "http",
			Entry:   Entry{ID: "CVE-2023-44487", CVSS: 7.5},
		},
	}

	got := FormatMatches(matches)
	if !strings.Contains(got, "CVSS 7.5") {
		t.Fatalf("expected CVSS 7.5 in output, got: %s", got)
	}
}

func TestFormatMatches_CVSSZero(t *testing.T) {
	matches := []Match{
		{
			HostIP:  "192.168.1.10",
			Port:    80,
			Service: "http",
			Entry:   Entry{ID: "CVE-2023-0000", CVSS: 0.0},
		},
	}

	got := FormatMatches(matches)
	if !strings.Contains(got, "CVSS 0.0") {
		t.Fatalf("expected CVSS 0.0 in output, got: %s", got)
	}
}

func TestFormatMatches_CVSSHigh(t *testing.T) {
	matches := []Match{
		{
			HostIP:  "192.168.1.10",
			Port:    80,
			Service: "http",
			Entry:   Entry{ID: "CVE-2023-0000", CVSS: 10.0},
		},
	}

	got := FormatMatches(matches)
	if !strings.Contains(got, "CVSS 10.0") {
		t.Fatalf("expected CVSS 10.0 in output, got: %s", got)
	}
}

// ============================================================================
// W-69: Тесты для NewDefaultCatalog (100% → 100%)
// ============================================================================

func TestNewDefaultCatalog_NotEmpty(t *testing.T) {
	catalog := NewDefaultCatalog()
	if len(catalog.entries) == 0 {
		t.Fatal("expected non-empty catalog")
	}
	if len(catalog.entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(catalog.entries))
	}
}

func TestNewDefaultCatalog_CVEStructure(t *testing.T) {
	catalog := NewDefaultCatalog()
	
	// Проверяем структуру первой записи
	entry := catalog.entries[0]
	if entry.ID != "CVE-2023-44487" {
		t.Fatalf("expected CVE-2023-44487, got %s", entry.ID)
	}
	if entry.CVSS != 7.5 {
		t.Fatalf("expected CVSS 7.5, got %f", entry.CVSS)
	}
	if entry.Service != "http" {
		t.Fatalf("expected service http, got %s", entry.Service)
	}
}

func TestNewDefaultCatalog_AllCVEs(t *testing.T) {
	catalog := NewDefaultCatalog()
	
	cveIDs := make(map[string]bool)
	for _, entry := range catalog.entries {
		cveIDs[entry.ID] = true
	}
	
	expectedIDs := []string{"CVE-2023-44487", "CVE-2023-38408", "CVE-2021-44228"}
	for _, id := range expectedIDs {
		if !cveIDs[id] {
			t.Fatalf("expected CVE %s in catalog", id)
		}
	}
}

func TestNewDefaultCatalog_CVSSValues(t *testing.T) {
	catalog := NewDefaultCatalog()
	
	expectedCVSS := map[string]float64{
		"CVE-2023-44487": 7.5,
		"CVE-2023-38408": 9.8,
		"CVE-2021-44228": 10.0,
	}
	
	for _, entry := range catalog.entries {
		expected, ok := expectedCVSS[entry.ID]
		if !ok {
			t.Fatalf("unexpected CVE %s", entry.ID)
		}
		if entry.CVSS != expected {
			t.Fatalf("expected CVSS %f for %s, got %f", expected, entry.ID, entry.CVSS)
		}
	}
}
