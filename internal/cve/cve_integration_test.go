package cve

import (
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

// === Integration: AnalyzeResults with Catalog ===

func TestIntegrationAnalyzeResults_MatchCVE(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:       "192.168.1.1",
			Hostname: "web-server",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Protocol: "TCP", Service: "http", Banner: "nginx/1.25"},
			},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) == 0 {
		t.Error("expected at least 1 match, got 0")
	}
	if len(matches) > 0 {
		if matches[0].Entry.ID != "CVE-2023-44487" {
			t.Errorf("expected CVE-2023-44487, got %s", matches[0].Entry.ID)
		}
		if matches[0].Port != 80 {
			t.Errorf("expected port 80, got %d", matches[0].Port)
		}
	}
}

func TestIntegrationAnalyzeResults_MatchSSH(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:       "192.168.1.2",
			Hostname: "ssh-server",
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Protocol: "TCP", Service: "ssh", Banner: "openssh_9.3"},
			},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) == 0 {
		t.Error("expected at least 1 match, got 0")
	}
	if len(matches) > 0 {
		if matches[0].Entry.ID != "CVE-2023-38408" {
			t.Errorf("expected CVE-2023-38408, got %s", matches[0].Entry.ID)
		}
	}
}

func TestIntegrationAnalyzeResults_NoMatch(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:    "192.168.1.99",
			Ports: []scanner.PortInfo{{Port: 25, State: "open", Protocol: "TCP", Service: "smtp", Banner: "postfix"}},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestIntegrationAnalyzeResults_ClosedPorts(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:    "192.168.1.1",
			Ports: []scanner.PortInfo{{Port: 80, State: "closed", Protocol: "TCP", Service: "http", Banner: "nginx"}},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for closed port, got %d", len(matches))
	}
}

func TestIntegrationAnalyzeResults_EmptyBanner(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:    "192.168.1.1",
			Ports: []scanner.PortInfo{{Port: 80, State: "open", Protocol: "TCP", Service: "http"}},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for empty banner, got %d", len(matches))
	}
}

// === Integration: MinCVSS Filter ===

func TestIntegrationAnalyzeResults_MinCVSS(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:    "192.168.1.1",
			Ports: []scanner.PortInfo{{Port: 80, State: "open", Protocol: "TCP", Service: "http", Banner: "log4j/2.14"}},
		},
	}

	// MinCVSS=8 should only match CVE-2021-44228 (CVSS 10.0)
	matches := AnalyzeResults(results, catalog, Options{MinCVSS: 8.0})
	if len(matches) != 1 {
		t.Errorf("expected 1 match with MinCVSS=8.0, got %d", len(matches))
	}
	if len(matches) > 0 && matches[0].Entry.ID != "CVE-2021-44228" {
		t.Errorf("expected CVE-2021-44228, got %s", matches[0].Entry.ID)
	}
}

func TestIntegrationAnalyzeResults_MinCVSSAll(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP: "192.168.1.1",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Protocol: "TCP", Service: "http", Banner: "log4j/2.14"},
				{Port: 22, State: "open", Protocol: "TCP", Service: "ssh", Banner: "openssh_9.3"},
			},
		},
	}

	// MinCVSS=0 should match all
	matches := AnalyzeResults(results, catalog, Options{MinCVSS: 0.0})
	if len(matches) != 2 {
		t.Errorf("expected 2 matches with MinCVSS=0, got %d", len(matches))
	}
}

// === Integration: MaxAgeDays Filter ===

func TestIntegrationAnalyzeResults_MaxAgeDays(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:    "192.168.1.1",
			Ports: []scanner.PortInfo{{Port: 80, State: "open", Protocol: "TCP", Service: "http", Banner: "log4j/2.14"}},
		},
	}

	// CVE-2021-44228 published 2021-12-10, MaxAgeDays=30 should filter it out
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	matches := AnalyzeResults(results, catalog, Options{MaxAgeDays: 30, Now: now})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches with MaxAgeDays=30, got %d", len(matches))
	}
}

func TestIntegrationAnalyzeResults_MaxAgeDaysNone(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:    "192.168.1.1",
			Ports: []scanner.PortInfo{{Port: 80, State: "open", Protocol: "TCP", Service: "http", Banner: "log4j/2.14"}},
		},
	}

	// MaxAgeDays=0 (unlimited) should match all
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	matches := AnalyzeResults(results, catalog, Options{MaxAgeDays: 0, Now: now})
	if len(matches) != 1 {
		t.Errorf("expected 1 match with MaxAgeDays=0, got %d", len(matches))
	}
}

// === Integration: Multiple Hosts ===

func TestIntegrationAnalyzeResults_MultipleHosts(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:    "192.168.1.1",
			Ports: []scanner.PortInfo{{Port: 80, State: "open", Protocol: "TCP", Service: "http", Banner: "log4j/2.14"}},
		},
		{
			IP:    "192.168.1.2",
			Ports: []scanner.PortInfo{{Port: 22, State: "open", Protocol: "TCP", Service: "ssh", Banner: "openssh_9.3"}},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) != 2 {
		t.Errorf("expected 2 matches for 2 hosts, got %d", len(matches))
	}
}

// === Integration: FormatMatches ===

func TestIntegrationFormatMatches_Empty(t *testing.T) {
	output := FormatMatches([]Match{})
	if output != "CVE анализ: совпадений не найдено." {
		t.Errorf("expected empty message, got %q", output)
	}
}

func TestIntegrationFormatMatches_Single(t *testing.T) {
	matches := []Match{
		{
			HostIP:  "192.168.1.1",
			Port:    80,
			Service: "http",
			Entry:   Entry{ID: "CVE-2021-44228", CVSS: 10.0},
		},
	}
	output := FormatMatches(matches)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if !contains(output, "CVE-2021-44228") {
		t.Error("expected output to contain CVE ID")
	}
	if !contains(output, "192.168.1.1") {
		t.Error("expected output to contain IP")
	}
}

func TestIntegrationFormatMatches_Multiple(t *testing.T) {
	matches := []Match{
		{HostIP: "192.168.1.1", Port: 80, Service: "http", Entry: Entry{ID: "CVE-2021-44228", CVSS: 10.0}},
		{HostIP: "192.168.1.2", Port: 22, Service: "ssh", Entry: Entry{ID: "CVE-2023-38408", CVSS: 9.8}},
	}
	output := FormatMatches(matches)
	if !contains(output, "CVE-2021-44228") {
		t.Error("expected output to contain first CVE")
	}
	if !contains(output, "CVE-2023-38408") {
		t.Error("expected output to contain second CVE")
	}
}

// === Integration: NormalizeService ===

func TestIntegrationNormalizeService_HTTP(t *testing.T) {
	port := scanner.PortInfo{Port: 80, Service: "http"}
	svc := normalizeService(port)
	if svc != "http" {
		t.Errorf("expected http, got %q", svc)
	}
}

func TestIntegrationNormalizeService_HTTPS(t *testing.T) {
	port := scanner.PortInfo{Port: 443, Service: "https"}
	svc := normalizeService(port)
	if svc != "http" {
		t.Errorf("expected http for https, got %q", svc)
	}
}

func TestIntegrationNormalizeService_SSH(t *testing.T) {
	port := scanner.PortInfo{Port: 22, Service: "ssh"}
	svc := normalizeService(port)
	if svc != "ssh" {
		t.Errorf("expected ssh, got %q", svc)
	}
}

func TestIntegrationNormalizeService_PortBased(t *testing.T) {
	port := scanner.PortInfo{Port: 22, Service: "unknown"}
	svc := normalizeService(port)
	if svc != "ssh" {
		t.Errorf("expected ssh (port-based), got %q", svc)
	}
}

func TestIntegrationNormalizeService_NoMatch(t *testing.T) {
	port := scanner.PortInfo{Port: 25, Service: "smtp"}
	svc := normalizeService(port)
	if svc != "" {
		t.Errorf("expected empty for smtp, got %q", svc)
	}
}

// === Integration: Catalog ===

func TestIntegrationNewDefaultCatalog(t *testing.T) {
	catalog := NewDefaultCatalog()
	entries := catalog.entries
	if len(entries) != 3 {
		t.Errorf("expected 3 default entries, got %d", len(entries))
	}
}

func TestIntegrationCatalogEntryFields(t *testing.T) {
	catalog := NewDefaultCatalog()
	for _, e := range catalog.entries {
		if e.ID == "" {
			t.Error("expected non-empty CVE ID")
		}
		if e.Description == "" {
			t.Error("expected non-empty description")
		}
		if e.URL == "" {
			t.Error("expected non-empty URL")
		}
		if e.Service == "" {
			t.Error("expected non-empty service")
		}
	}
}

// === Integration: Match Sorting ===

func TestIntegrationAnalyzeResults_SortedByCVSS(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP: "192.168.1.1",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Protocol: "TCP", Service: "http", Banner: "log4j/2.14"},
				{Port: 22, State: "open", Protocol: "TCP", Service: "ssh", Banner: "openssh_9.3"},
			},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) >= 2 {
		if matches[0].Entry.CVSS < matches[1].Entry.CVSS {
			t.Error("expected matches sorted by CVSS descending")
		}
	}
}

// === Integration: HostName in Match ===

func TestIntegrationAnalyzeResults_WithHostName(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:       "192.168.1.1",
			Hostname: "web-server",
			Ports:    []scanner.PortInfo{{Port: 80, State: "open", Protocol: "TCP", Service: "http", Banner: "log4j/2.14"}},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) > 0 {
		if matches[0].HostName != "web-server" {
			t.Errorf("expected hostname 'web-server', got %q", matches[0].HostName)
		}
	}
}

// === Integration: Port Version in Match ===

func TestIntegrationAnalyzeResults_PortVersionHint(t *testing.T) {
	catalog := NewDefaultCatalog()
	results := []scanner.Result{
		{
			IP:    "192.168.1.1",
			Ports: []scanner.PortInfo{{Port: 80, State: "open", Protocol: "TCP", Service: "http", Version: "1.25"}},
		},
	}

	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) > 0 {
		if matches[0].VersionHint != "1.25" {
			t.Errorf("expected version hint '1.25', got %q", matches[0].VersionHint)
		}
	}
}

// === Integration: Full CVE Pipeline ===

func TestIntegrationFullCVEPipeline(t *testing.T) {
	// Step 1: Сканирование (mock)
	results := []scanner.Result{
		{
			IP:       "192.168.1.1",
			Hostname: "web-server",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Protocol: "TCP", Service: "http", Banner: "nginx/1.25"},
				{Port: 443, State: "open", Protocol: "TCP", Service: "https", Banner: "nginx/1.25"},
			},
		},
		{
			IP:    "192.168.1.2",
			Ports: []scanner.PortInfo{{Port: 22, State: "open", Protocol: "TCP", Service: "ssh", Banner: "openssh_9.3"}},
		},
	}

	// Step 2: CVE анализ
	catalog := NewDefaultCatalog()
	matches := AnalyzeResults(results, catalog, Options{})

	if len(matches) == 0 {
		t.Fatal("expected at least 1 CVE match")
	}

	// Step 3: Фильтрация по CVSS
	highSeverity := AnalyzeResults(results, catalog, Options{MinCVSS: 9.0})
	if len(highSeverity) == 0 {
		t.Error("expected high severity matches")
	}

	// Step 4: Форматирование
	output := FormatMatches(matches)
	if output == "" {
		t.Error("expected non-empty formatted output")
	}
}

// === Integration: Edge Cases ===

func TestIntegrationAnalyzeResults_NilResults(t *testing.T) {
	catalog := NewDefaultCatalog()
	matches := AnalyzeResults(nil, catalog, Options{})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for nil results, got %d", len(matches))
	}
}

func TestIntegrationAnalyzeResults_EmptyResults(t *testing.T) {
	catalog := NewDefaultCatalog()
	matches := AnalyzeResults([]scanner.Result{}, catalog, Options{})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for empty results, got %d", len(matches))
	}
}

func TestIntegrationAnalyzeResults_NilCatalog(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, State: "open", Service: "http", Banner: "nginx"}}},
	}
	var catalog Catalog
	matches := AnalyzeResults(results, catalog, Options{})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for nil catalog, got %d", len(matches))
	}
}

func TestIntegrationFormatMatches_Nil(t *testing.T) {
	output := FormatMatches(nil)
	if output == "" {
		t.Error("expected non-empty output for nil matches")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
