package gui

import (
	"fmt"
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

func BenchmarkFormatResultsForDisplayEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatResultsForDisplay(nil)
	}
}

func BenchmarkFormatResultsForDisplaySmall(b *testing.B) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", MAC: "aa:bb:cc:dd:ee:ff"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatResultsForDisplay(results)
	}
}

func BenchmarkFormatPortsEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatPorts(nil)
	}
}

func BenchmarkEscapeMarkdownBasic(b *testing.B) {
	input := "hello world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		escapeMarkdown(input)
	}
}

func BenchmarkTruncateString(b *testing.B) {
	input := "hello world this is a long string"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateString(input, 10)
	}
}

func BenchmarkSortedResultsForDisplay(b *testing.B) {
	results := []scanner.Result{
		{IP: "192.168.1.20", Hostname: "b"},
		{IP: "192.168.1.3", Hostname: "a"},
		{IP: "10.0.0.2", Hostname: "z"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedResultsForDisplay(results)
	}
}

func BenchmarkFilterResultsForDisplay(b *testing.B) {
	results := []scanner.Result{
		{Hostname: "router-main", IP: "192.168.1.1"},
		{Hostname: "workstation", IP: "192.168.1.10"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterResultsForDisplay(results, "router")
	}
}

func BenchmarkFormatDurationMMSS(b *testing.B) {
	duration := 3661 * time.Second
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatDurationMMSS(duration)
	}
}

func BenchmarkNormalizeDeviceTypes(b *testing.B) {
	raw := map[string]int{
		"Router": 1, "Computer": 2, "Server": 1, "Unknown": 3,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeDeviceTypes(raw)
	}
}

func BenchmarkFormatResultsForDisplayLarge(b *testing.B) {
	results := make([]scanner.Result, 0, 100)
	for i := 0; i < 100; i++ {
		results = append(results, scanner.Result{
			IP:       fmt.Sprintf("192.168.1.%d", i+1),
			Hostname: fmt.Sprintf("host-%d", i+1),
			Ports: []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", State: "open"},
			},
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatResultsForDisplay(results)
	}
}
