package audit

import (
	"strings"
	"testing"

	"network-scanner/internal/scanner"
)

func TestEvaluateOpenPortsFindsKnownRisks(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.10",
			Ports: []scanner.PortInfo{
				{Port: 23, Protocol: "tcp", State: "open"},
				{Port: 445, Protocol: "tcp", State: "open"},
				{Port: 53, Protocol: "udp", State: "open"},
			},
		},
	}

	findings := EvaluateOpenPorts(results)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestFormatFindingsEmpty(t *testing.T) {
	msg := FormatFindings(nil)
	if !strings.Contains(strings.ToLower(msg), "рисков") {
		t.Fatalf("expected no-risk message, got %q", msg)
	}
}

func TestBuildSummary(t *testing.T) {
	findings := []Finding{
		{Host: "192.168.1.10", Severity: "high"},
		{Host: "192.168.1.10", Severity: "medium"},
		{Host: "192.168.1.20", Severity: "high"},
	}
	s := BuildSummary(findings)
	if s.TotalFindings != 3 {
		t.Fatalf("expected total=3, got %d", s.TotalFindings)
	}
	if s.UniqueHosts != 2 {
		t.Fatalf("expected hosts=2, got %d", s.UniqueHosts)
	}
	if s.HighestSeverity != "high" {
		t.Fatalf("expected highest=high, got %q", s.HighestSeverity)
	}
	if s.BySeverity["high"] != 2 {
		t.Fatalf("expected high=2, got %d", s.BySeverity["high"])
	}
}

func TestFormatFindingsContainsSummary(t *testing.T) {
	findings := []Finding{
		{Host: "192.168.1.10", Port: 23, Protocol: "tcp", Severity: "high", Title: "Telnet", Recommendation: "Disable"},
	}
	msg := FormatFindings(findings)
	lower := strings.ToLower(msg)
	if !strings.Contains(lower, "risk score") {
		t.Fatalf("expected risk score in output, got %q", msg)
	}
	if !strings.Contains(lower, "хосты с рисками") {
		t.Fatalf("expected host summary in output, got %q", msg)
	}
}

func TestFilterByMinSeverity(t *testing.T) {
	findings := []Finding{
		{Host: "h1", Severity: "low"},
		{Host: "h1", Severity: "medium"},
		{Host: "h1", Severity: "high"},
	}
	out := FilterByMinSeverity(findings, "high")
	if len(out) != 1 {
		t.Fatalf("expected 1 finding for high+, got %d", len(out))
	}
	if out[0].Severity != "high" {
		t.Fatalf("expected high severity, got %q", out[0].Severity)
	}
}

func TestNormalizeSeverity(t *testing.T) {
	if v, ok := NormalizeSeverity(" HIGH "); !ok || v != "high" {
		t.Fatalf("expected normalized high, got %q ok=%v", v, ok)
	}
	if _, ok := NormalizeSeverity("oops"); ok {
		t.Fatal("expected invalid severity")
	}
}
