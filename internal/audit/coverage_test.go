package audit

import (
	"strings"
	"testing"

	"network-scanner/internal/scanner"
)

// ============================================================================
// HumanReadable — покрытие всех веток (31.2% → 100%)
// ============================================================================

func TestHumanReadable_EmptyHost(t *testing.T) {
	f := Finding{
		Host:     "",
		Port:     23,
		Protocol: "tcp",
		Title:    "Telnet без шифрования",
	}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "неизвестном устройстве") {
		t.Fatalf("expected 'неизвестном устройстве' for empty host, got: %s", msg)
	}
}

func TestHumanReadable_Telnet(t *testing.T) {
	f := Finding{Host: "10.0.0.1", Title: "Telnet без шифрования"}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "Telnet") {
		t.Fatalf("expected Telnet in message, got: %s", msg)
	}
	if !strings.Contains(msg, "10.0.0.1") {
		t.Fatal("expected host in message")
	}
}

func TestHumanReadable_FTP(t *testing.T) {
	f := Finding{Host: "10.0.0.2", Title: "FTP без шифрования"}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "FTP") {
		t.Fatalf("expected FTP in message, got: %s", msg)
	}
}

func TestHumanReadable_SMB(t *testing.T) {
	f := Finding{Host: "10.0.0.3", Title: "SMB доступен"}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "SMB") {
		t.Fatalf("expected SMB in message, got: %s", msg)
	}
}

func TestHumanReadable_SMBNetBIOS(t *testing.T) {
	f := Finding{Host: "10.0.0.4", Title: "SMB/NetBIOS доступен"}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "SMB") {
		t.Fatalf("expected SMB in message, got: %s", msg)
	}
}

func TestHumanReadable_RDP(t *testing.T) {
	f := Finding{Host: "10.0.0.5", Title: "RDP доступен"}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "RDP") {
		t.Fatalf("expected RDP in message, got: %s", msg)
	}
}

func TestHumanReadable_DefaultWithPort(t *testing.T) {
	f := Finding{
		Host:     "10.0.0.6",
		Port:     6379,
		Protocol: "TCP",
		Title:    "Redis доступен",
	}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "Redis") {
		t.Fatalf("expected Redis in message, got: %s", msg)
	}
	if !strings.Contains(msg, "порт 6379") {
		t.Fatalf("expected port in message, got: %s", msg)
	}
}

func TestHumanReadable_DefaultEmptyTitle(t *testing.T) {
	f := Finding{
		Host:     "10.0.0.7",
		Port:     0,
		Protocol: "",
		Title:    "",
	}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "Обнаружен потенциальный риск") {
		t.Fatalf("expected default title, got: %s", msg)
	}
}

func TestHumanReadable_DefaultNoPort(t *testing.T) {
	f := Finding{
		Host:     "10.0.0.8",
		Port:     0,
		Protocol: "",
		Title:    "Custom risk",
	}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "Custom risk") {
		t.Fatalf("expected custom title, got: %s", msg)
	}
	// Port == 0, so no port part
	if strings.Contains(msg, "порт") {
		t.Fatalf("expected no port info, got: %s", msg)
	}
}

func TestHumanReadable_CaseInsensitive(t *testing.T) {
	f := Finding{Host: "10.0.0.9", Title: "TELNET БЕЗ ШИФРОВАНИЯ"}
	msg := HumanReadable(f)
	if !strings.Contains(msg, "Telnet") {
		t.Fatalf("expected Telnet for case-insensitive match, got: %s", msg)
	}
}

// ============================================================================
// EvaluateOpenPorts — edge cases (81.2% → 100%)
// ============================================================================

func TestEvaluateOpenPorts_EmptyResults(t *testing.T) {
	findings := EvaluateOpenPorts([]scanner.Result{})
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for empty results, got %d", len(findings))
	}
}

func TestEvaluateOpenPorts_NoOpenPorts(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 23, Protocol: "tcp", State: "closed"},
				{Port: 445, Protocol: "tcp", State: "filtered"},
			},
		},
	}
	findings := EvaluateOpenPorts(results)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-open ports, got %d", len(findings))
	}
}

func TestEvaluateOpenPorts_NoRiskyPorts(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", State: "open"},
				{Port: 443, Protocol: "tcp", State: "open"},
			},
		},
	}
	findings := EvaluateOpenPorts(results)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-risky ports, got %d", len(findings))
	}
}

func TestEvaluateOpenPorts_AllRiskyPorts(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 21, Protocol: "tcp", State: "open"},
				{Port: 23, Protocol: "tcp", State: "open"},
				{Port: 139, Protocol: "tcp", State: "open"},
				{Port: 445, Protocol: "tcp", State: "open"},
				{Port: 3389, Protocol: "tcp", State: "open"},
				{Port: 5900, Protocol: "tcp", State: "open"},
				{Port: 6379, Protocol: "tcp", State: "open"},
				{Port: 9200, Protocol: "tcp", State: "open"},
				{Port: 27017, Protocol: "tcp", State: "open"},
				{Port: 11211, Protocol: "udp", State: "open"},
			},
		},
	}
	findings := EvaluateOpenPorts(results)
	if len(findings) != 10 {
		t.Fatalf("expected 10 findings for all risky ports, got %d", len(findings))
	}
}

func TestEvaluateOpenPorts_SortingBySeverity(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 21, Protocol: "tcp", State: "open"},   // medium
				{Port: 23, Protocol: "tcp", State: "open"},   // high
				{Port: 445, Protocol: "tcp", State: "open"},  // high
			},
		},
	}
	findings := EvaluateOpenPorts(results)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
	// high should come before medium
	if findings[0].Severity != "high" {
		t.Fatalf("expected first finding to be high, got %s", findings[0].Severity)
	}
	if findings[2].Severity != "medium" {
		t.Fatalf("expected last finding to be medium, got %s", findings[2].Severity)
	}
}

func TestEvaluateOpenPorts_SortingByHost(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.20",
			Ports: []scanner.PortInfo{
				{Port: 23, Protocol: "tcp", State: "open"}, // high
			},
		},
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 445, Protocol: "tcp", State: "open"}, // high
			},
		},
	}
	findings := EvaluateOpenPorts(results)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	// Same severity, sort by host ascending
	if findings[0].Host != "192.168.1.10" {
		t.Fatalf("expected first host 192.168.1.10, got %s", findings[0].Host)
	}
}

func TestEvaluateOpenPorts_SortingByPort(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 445, Protocol: "tcp", State: "open"},  // high
				{Port: 23, Protocol: "tcp", State: "open"},   // high
			},
		},
	}
	findings := EvaluateOpenPorts(results)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	// Same severity, same host, sort by port ascending
	if findings[0].Port != 23 {
		t.Fatalf("expected first port 23, got %d", findings[0].Port)
	}
}

func TestEvaluateOpenPorts_MultipleHosts(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 23, Protocol: "tcp", State: "open"},
			},
		},
		{
			IP: "192.168.1.20",
			Ports: []scanner.PortInfo{
				{Port: 445, Protocol: "tcp", State: "open"},
			},
		},
		{
			IP: "192.168.1.30",
			Ports: []scanner.PortInfo{
				{Port: 21, Protocol: "tcp", State: "open"},
			},
		},
	}
	findings := EvaluateOpenPorts(results)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
}

func TestEvaluateOpenPorts_WhitespaceIP(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "  192.168.1.10  ",
			Ports: []scanner.PortInfo{
				{Port: 23, Protocol: "  tcp  ", State: "open"},
			},
		},
	}
	findings := EvaluateOpenPorts(results)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Host != "192.168.1.10" {
		t.Fatalf("expected trimmed host, got %q", findings[0].Host)
	}
	if findings[0].Protocol != "tcp" {
		t.Fatalf("expected trimmed protocol, got %q", findings[0].Protocol)
	}
}

// ============================================================================
// FilterByMinSeverity — edge cases (72.7% → 100%)
// ============================================================================

func TestFilterByMinSeverity_All(t *testing.T) {
	findings := []Finding{
		{Host: "h1", Severity: "low"},
		{Host: "h2", Severity: "high"},
	}
	out := FilterByMinSeverity(findings, "all")
	if len(out) != 2 {
		t.Fatalf("expected 2 findings for 'all', got %d", len(out))
	}
}

func TestFilterByMinSeverity_InvalidSeverity(t *testing.T) {
	findings := []Finding{
		{Host: "h1", Severity: "low"},
	}
	out := FilterByMinSeverity(findings, "invalid")
	if len(out) != 1 {
		t.Fatalf("expected 1 finding for invalid severity (returns all), got %d", len(out))
	}
}

func TestFilterByMinSeverity_Critical(t *testing.T) {
	findings := []Finding{
		{Host: "h1", Severity: "low"},
		{Host: "h2", Severity: "medium"},
		{Host: "h3", Severity: "high"},
		{Host: "h4", Severity: "critical"},
	}
	out := FilterByMinSeverity(findings, "critical")
	if len(out) != 1 {
		t.Fatalf("expected 1 finding for critical+, got %d", len(out))
	}
	if out[0].Severity != "critical" {
		t.Fatalf("expected critical, got %s", out[0].Severity)
	}
}

func TestFilterByMinSeverity_Medium(t *testing.T) {
	findings := []Finding{
		{Host: "h1", Severity: "low"},
		{Host: "h2", Severity: "medium"},
		{Host: "h3", Severity: "high"},
	}
	out := FilterByMinSeverity(findings, "medium")
	if len(out) != 2 {
		t.Fatalf("expected 2 findings for medium+, got %d", len(out))
	}
}

func TestFilterByMinSeverity_Low(t *testing.T) {
	findings := []Finding{
		{Host: "h1", Severity: "low"},
		{Host: "h2", Severity: "medium"},
		{Host: "h3", Severity: "high"},
	}
	out := FilterByMinSeverity(findings, "low")
	if len(out) != 3 {
		t.Fatalf("expected 3 findings for low+, got %d", len(out))
	}
}

func TestFilterByMinSeverity_EmptyFindings(t *testing.T) {
	out := FilterByMinSeverity([]Finding{}, "high")
	if len(out) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(out))
	}
}

func TestFilterByMinSeverity_PreservesOrder(t *testing.T) {
	findings := []Finding{
		{Host: "h1", Severity: "high"},
		{Host: "h2", Severity: "low"},
		{Host: "h3", Severity: "high"},
	}
	out := FilterByMinSeverity(findings, "high")
	if len(out) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(out))
	}
	if out[0].Host != "h1" || out[1].Host != "h3" {
		t.Fatal("expected order preserved")
	}
}

// ============================================================================
// sortedHosts — edge cases (62.5% → 100%)
// ============================================================================

func TestSortedHosts_Empty(t *testing.T) {
	out := sortedHosts(map[string]int{})
	if len(out) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(out))
	}
}

func TestSortedHosts_SingleHost(t *testing.T) {
	out := sortedHosts(map[string]int{"10.0.0.1": 3})
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0].host != "10.0.0.1" || out[0].count != 3 {
		t.Fatalf("unexpected entry: %+v", out[0])
	}
}

func TestSortedHosts_SortByCount(t *testing.T) {
	byHost := map[string]int{
		"10.0.0.1": 1,
		"10.0.0.2": 5,
		"10.0.0.3": 3,
	}
	out := sortedHosts(byHost)
	if len(out) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(out))
	}
	// Descending by count
	if out[0].host != "10.0.0.2" || out[0].count != 5 {
		t.Fatalf("expected 10.0.0.2 first, got %+v", out[0])
	}
}

func TestSortedHosts_SortByHostWhenEqualCount(t *testing.T) {
	byHost := map[string]int{
		"10.0.0.3": 2,
		"10.0.0.1": 2,
		"10.0.0.2": 2,
	}
	out := sortedHosts(byHost)
	if len(out) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(out))
	}
	// Same count, sort by host ascending
	if out[0].host != "10.0.0.1" {
		t.Fatalf("expected 10.0.0.1 first, got %s", out[0].host)
	}
	if out[1].host != "10.0.0.2" {
		t.Fatalf("expected 10.0.0.2 second, got %s", out[1].host)
	}
}

// ============================================================================
// severityWeight — edge cases (66.7% → 100%)
// ============================================================================

func TestSeverityWeight_Critical(t *testing.T) {
	if w := severityWeight("critical"); w != 4 {
		t.Fatalf("severityWeight(critical) = %d, want 4", w)
	}
}

func TestSeverityWeight_High(t *testing.T) {
	if w := severityWeight("high"); w != 3 {
		t.Fatalf("severityWeight(high) = %d, want 3", w)
	}
}

func TestSeverityWeight_Medium(t *testing.T) {
	if w := severityWeight("medium"); w != 2 {
		t.Fatalf("severityWeight(medium) = %d, want 2", w)
	}
}

func TestSeverityWeight_Low(t *testing.T) {
	if w := severityWeight("low"); w != 1 {
		t.Fatalf("severityWeight(low) = %d, want 1", w)
	}
}

func TestSeverityWeight_Unknown(t *testing.T) {
	if w := severityWeight("unknown"); w != 0 {
		t.Fatalf("severityWeight(unknown) = %d, want 0", w)
	}
}

func TestSeverityWeight_Empty(t *testing.T) {
	if w := severityWeight(""); w != 0 {
		t.Fatalf("severityWeight(empty) = %d, want 0", w)
	}
}

func TestSeverityWeight_CaseInsensitive(t *testing.T) {
	if w := severityWeight("HIGH"); w != 3 {
		t.Fatalf("severityWeight(HIGH) = %d, want 3", w)
	}
}

func TestSeverityWeight_Whitespace(t *testing.T) {
	if w := severityWeight("  critical  "); w != 4 {
		t.Fatalf("severityWeight(whitespace) = %d, want 4", w)
	}
}

// ============================================================================
// SecurityIndexFromSeverityCounts — edge cases (91.7% → 100%)
// ============================================================================

func TestSecurityIndexFromSeverityCounts_EmptyMap(t *testing.T) {
	score := SecurityIndexFromSeverityCounts(map[string]int{})
	if score != 100 {
		t.Fatalf("expected 100 for empty map, got %d", score)
	}
}

func TestSecurityIndexFromSeverityCounts_OnlyLow(t *testing.T) {
	score := SecurityIndexFromSeverityCounts(map[string]int{"low": 5})
	if score != 75 {
		t.Fatalf("expected 75 for 5 low, got %d", score)
	}
}

func TestSecurityIndexFromSeverityCounts_OnlyMedium(t *testing.T) {
	score := SecurityIndexFromSeverityCounts(map[string]int{"medium": 3})
	if score != 70 {
		t.Fatalf("expected 70 for 3 medium, got %d", score)
	}
}

func TestSecurityIndexFromSeverityCounts_OnlyHigh(t *testing.T) {
	score := SecurityIndexFromSeverityCounts(map[string]int{"high": 2})
	if score != 60 {
		t.Fatalf("expected 60 for 2 high, got %d", score)
	}
}

func TestSecurityIndexFromSeverityCounts_NegativeScore(t *testing.T) {
	score := SecurityIndexFromSeverityCounts(map[string]int{
		"critical": 5, // 150
	})
	if score != 0 {
		t.Fatalf("expected 0 for negative score, got %d", score)
	}
}

// ============================================================================
// FormatFindings — edge cases
// ============================================================================

func TestFormatFindings_MultipleSeverities(t *testing.T) {
	findings := []Finding{
		{Host: "10.0.0.1", Port: 23, Protocol: "tcp", Severity: "high", Title: "Telnet без шифрования", Recommendation: "Disable"},
		{Host: "10.0.0.2", Port: 21, Protocol: "tcp", Severity: "medium", Title: "FTP без шифрования", Recommendation: "Use SFTP"},
	}
	msg := FormatFindings(findings)
	lower := strings.ToLower(msg)
	if !strings.Contains(lower, "high=1") {
		t.Fatalf("expected HIGH=1 in output, got: %s", msg)
	}
	if !strings.Contains(lower, "medium=1") {
		t.Fatalf("expected MEDIUM=1 in output, got: %s", msg)
	}
}

func TestFormatFindings_WithCriticalSeverity(t *testing.T) {
	findings := []Finding{
		{Host: "10.0.0.1", Port: 23, Protocol: "tcp", Severity: "critical", Title: "Critical risk", Recommendation: "Fix now"},
	}
	msg := FormatFindings(findings)
	if !strings.Contains(msg, "CRITICAL=1") {
		t.Fatalf("expected CRITICAL=1 in output, got: %s", msg)
	}
}

func TestFormatFindings_MultipleHosts(t *testing.T) {
	findings := []Finding{
		{Host: "10.0.0.1", Port: 23, Protocol: "tcp", Severity: "high", Title: "Telnet", Recommendation: "Disable"},
		{Host: "10.0.0.2", Port: 445, Protocol: "tcp", Severity: "high", Title: "SMB", Recommendation: "Restrict"},
		{Host: "10.0.0.1", Port: 21, Protocol: "tcp", Severity: "medium", Title: "FTP", Recommendation: "Use SFTP"},
	}
	msg := FormatFindings(findings)
	if !strings.Contains(msg, "10.0.0.1") {
		t.Fatal("expected host 10.0.0.1 in output")
	}
	if !strings.Contains(msg, "10.0.0.2") {
		t.Fatal("expected host 10.0.0.2 in output")
	}
}

// ============================================================================
// BuildSummary — edge cases
// ============================================================================

func TestBuildSummary_EmptyFindings(t *testing.T) {
	s := BuildSummary([]Finding{})
	if s.TotalFindings != 0 {
		t.Fatalf("expected 0 findings, got %d", s.TotalFindings)
	}
	if s.UniqueHosts != 0 {
		t.Fatalf("expected 0 hosts, got %d", s.UniqueHosts)
	}
	if s.OverallRiskScore != 0 {
		t.Fatalf("expected 0 risk score, got %d", s.OverallRiskScore)
	}
}

func TestBuildSummary_CriticalSeverity(t *testing.T) {
	findings := []Finding{
		{Host: "10.0.0.1", Severity: "critical"},
		{Host: "10.0.0.1", Severity: "high"},
	}
	s := BuildSummary(findings)
	if s.HighestSeverity != "critical" {
		t.Fatalf("expected highest=critical, got %q", s.HighestSeverity)
	}
	if s.BySeverity["critical"] != 1 {
		t.Fatalf("expected critical=1, got %d", s.BySeverity["critical"])
	}
}

func TestBuildSummary_WhitespaceHandling(t *testing.T) {
	findings := []Finding{
		{Host: "  10.0.0.1  ", Severity: "  HIGH  "},
	}
	s := BuildSummary(findings)
	if s.HighestSeverity != "high" {
		t.Fatalf("expected normalized high, got %q", s.HighestSeverity)
	}
	if s.ByHost["10.0.0.1"] != 1 {
		t.Fatalf("expected trimmed host counted, got %d", s.ByHost["10.0.0.1"])
	}
}

func TestBuildSummary_RiskScore(t *testing.T) {
	findings := []Finding{
		{Host: "h1", Severity: "critical"}, // 4
		{Host: "h1", Severity: "high"},     // 3
		{Host: "h1", Severity: "medium"},   // 2
		{Host: "h1", Severity: "low"},      // 1
	}
	s := BuildSummary(findings)
	if s.OverallRiskScore != 10 {
		t.Fatalf("expected risk score 10, got %d", s.OverallRiskScore)
	}
}

// ============================================================================
// NormalizeSeverity — edge cases
// ============================================================================

func TestNormalizeSeverity_AllValues(t *testing.T) {
	validValues := []string{"all", "critical", "high", "medium", "low"}
	for _, v := range validValues {
		got, ok := NormalizeSeverity(v)
		if !ok {
			t.Fatalf("NormalizeSeverity(%q) should be valid", v)
		}
		if got != v {
			t.Fatalf("NormalizeSeverity(%q) = %q, want %q", v, got, v)
		}
	}
}

func TestNormalizeSeverity_Uppercase(t *testing.T) {
	got, ok := NormalizeSeverity("CRITICAL")
	if !ok || got != "critical" {
		t.Fatalf("NormalizeSeverity(CRITICAL) = %q ok=%v, want critical true", got, ok)
	}
}

func TestNormalizeSeverity_Whitespace(t *testing.T) {
	got, ok := NormalizeSeverity("  medium  ")
	if !ok || got != "medium" {
		t.Fatalf("NormalizeSeverity(whitespace) = %q ok=%v, want medium true", got, ok)
	}
}

func TestNormalizeSeverity_Empty(t *testing.T) {
	_, ok := NormalizeSeverity("")
	if ok {
		t.Fatal("NormalizeSeverity(empty) should be invalid")
	}
}

func TestNormalizeSeverity_InvalidValues(t *testing.T) {
	invalidValues := []string{"none", "info", "debug", "warn", "error", "123"}
	for _, v := range invalidValues {
		_, ok := NormalizeSeverity(v)
		if ok {
			t.Fatalf("NormalizeSeverity(%q) should be invalid", v)
		}
	}
}
