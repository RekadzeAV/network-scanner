package audit

import (
	"fmt"
	"testing"

	"network-scanner/internal/scanner"
)

// === Performance: EvaluateOpenPorts ===

func BenchmarkEvaluateOpenPorts_Empty(b *testing.B) {
	results := []scanner.Result{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluateOpenPorts(results)
	}
}

func BenchmarkEvaluateOpenPorts_Few(b *testing.B) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{
			{Port: 22, State: "open", Protocol: "TCP"},
			{Port: 80, State: "open", Protocol: "TCP"},
		}},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluateOpenPorts(results)
	}
}

func BenchmarkEvaluateOpenPorts_Many(b *testing.B) {
	results := make([]scanner.Result, 100)
	ports := []scanner.PortInfo{
		{Port: 21, State: "open", Protocol: "TCP"},
		{Port: 22, State: "open", Protocol: "TCP"},
		{Port: 23, State: "open", Protocol: "TCP"},
		{Port: 80, State: "open", Protocol: "TCP"},
		{Port: 443, State: "open", Protocol: "TCP"},
		{Port: 445, State: "open", Protocol: "TCP"},
		{Port: 3389, State: "open", Protocol: "TCP"},
		{Port: 5900, State: "open", Protocol: "TCP"},
		{Port: 6379, State: "open", Protocol: "TCP"},
		{Port: 9200, State: "open", Protocol: "TCP"},
		{Port: 11211, State: "open", Protocol: "TCP"},
		{Port: 27017, State: "open", Protocol: "TCP"},
	}
	for i := 0; i < 100; i++ {
		results[i] = scanner.Result{
			IP:    fmt.Sprintf("192.168.1.%d", i+1),
			Ports: append([]scanner.PortInfo(nil), ports...),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluateOpenPorts(results)
	}
}

// === Performance: BuildSummary ===

func BenchmarkBuildSummary_Empty(b *testing.B) {
	findings := []Finding{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildSummary(findings)
	}
}

func BenchmarkBuildSummary_Few(b *testing.B) {
	findings := []Finding{
		{Host: "192.168.1.1", Port: 22, Severity: "high"},
		{Host: "192.168.1.2", Port: 80, Severity: "medium"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildSummary(findings)
	}
}

func BenchmarkBuildSummary_Many(b *testing.B) {
	findings := make([]Finding, 100)
	severities := []string{"critical", "high", "medium", "low"}
	for i := 0; i < 100; i++ {
		findings[i] = Finding{
			Host:     fmt.Sprintf("192.168.1.%d", i+1),
			Port:     22 + i,
			Severity: severities[i%4],
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildSummary(findings)
	}
}

// === Performance: FormatFindings ===

func BenchmarkFormatFindings_Empty(b *testing.B) {
	findings := []Finding{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatFindings(findings)
	}
}

func BenchmarkFormatFindings_Few(b *testing.B) {
	findings := []Finding{
		{Host: "192.168.1.1", Port: 22, Severity: "high", Title: "Telnet", Recommendation: "Use SSH"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatFindings(findings)
	}
}

// === Performance: FormatFindingCSV (alias for FormatFindings) ===

func BenchmarkFormatFindingsCSV_Empty(b *testing.B) {
	findings := []Finding{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatFindings(findings)
	}
}

func BenchmarkFormatFindingsCSV_Few(b *testing.B) {
	findings := []Finding{
		{Host: "192.168.1.1", Port: 22, Severity: "high", Title: "Telnet", Recommendation: "Use SSH"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatFindings(findings)
	}
}
