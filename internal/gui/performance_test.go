package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

// === Performance: Results Model ===

func BenchmarkSortedResultsForDisplay_Empty(b *testing.B) {
	results := []scanner.Result{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedResultsForDisplay(results)
	}
}

func BenchmarkSortedResultsForDisplay_Small(b *testing.B) {
	results := make([]scanner.Result, 10)
	for i := 0; i < 10; i++ {
		results[i] = scanner.Result{
			IP:       "192.168.1." + string(rune('1'+i)),
			Hostname: "host" + string(rune('a'+i)),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedResultsForDisplay(results)
	}
}

func BenchmarkSortedResultsForDisplay_Large(b *testing.B) {
	results := make([]scanner.Result, 1000)
	for i := 0; i < 1000; i++ {
		results[i] = scanner.Result{
			IP:       "192.168." + string(rune('0'+i/256%10)) + "." + string(rune('1'+i%256)),
			Hostname: "host" + string(rune('a'+i%26)),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedResultsForDisplay(results)
	}
}

func BenchmarkSortedResultsForDisplay_ByHostname(b *testing.B) {
	results := make([]scanner.Result, 100)
	for i := 0; i < 100; i++ {
		results[i] = scanner.Result{
			IP:       "192.168.1." + string(rune('1'+i%10)),
			Hostname: "host" + string(rune('a'+i%26)),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedResultsForDisplayWithMode(results, "HostName")
	}
}

func BenchmarkFilterResultsForDisplay_Empty(b *testing.B) {
	results := []scanner.Result{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterResultsForDisplay(results, "")
	}
}

func BenchmarkFilterResultsForDisplay_Small(b *testing.B) {
	results := make([]scanner.Result, 100)
	for i := 0; i < 100; i++ {
		results[i] = scanner.Result{
			IP:       "192.168.1." + string(rune('1'+i%10)),
			Hostname: "host" + string(rune('a'+i%26)),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterResultsForDisplay(results, "192")
	}
}

func BenchmarkFilterResultsForDisplay_Large(b *testing.B) {
	results := make([]scanner.Result, 1000)
	for i := 0; i < 1000; i++ {
		results[i] = scanner.Result{
			IP:       "192.168." + string(rune('0'+i/256%10)) + "." + string(rune('1'+i%256)),
			Hostname: "host" + string(rune('a'+i%26)),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterResultsForDisplay(results, "192")
	}
}

func BenchmarkHasOpenPorts_Empty(b *testing.B) {
	ports := []scanner.PortInfo{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasOpenPorts(ports)
	}
}

func BenchmarkHasOpenPorts_Open(b *testing.B) {
	ports := []scanner.PortInfo{{Port: 22, State: "open"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasOpenPorts(ports)
	}
}

func BenchmarkHasOpenPorts_Closed(b *testing.B) {
	ports := []scanner.PortInfo{{Port: 22, State: "closed"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasOpenPorts(ports)
	}
}

func BenchmarkOpenPortLabels_Empty(b *testing.B) {
	ports := []scanner.PortInfo{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		openPortLabels(ports, 5)
	}
}

func BenchmarkOpenPortLabels_Few(b *testing.B) {
	ports := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"},
		{Port: 80, State: "open", Protocol: "TCP", Service: "http"},
		{Port: 443, State: "open", Protocol: "TCP", Service: "https"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		openPortLabels(ports, 5)
	}
}

func BenchmarkOpenPortLabels_Limit(b *testing.B) {
	ports := make([]scanner.PortInfo, 100)
	for i := 0; i < 100; i++ {
		ports[i] = scanner.PortInfo{Port: i, State: "open", Protocol: "TCP"}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		openPortLabels(ports, 5)
	}
}

func BenchmarkFormatPortNumber_Empty(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatPortNumber(0)
	}
}

func BenchmarkFormatPortNumber_Basic(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatPortNumber(8080)
	}
}

func BenchmarkFormatPortNumber_Large(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatPortNumber(65535)
	}
}

func BenchmarkNormalizeServiceName_Empty(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeServiceName("")
	}
}

func BenchmarkNormalizeServiceName_Basic(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeServiceName("HTTP Server")
	}
}

func BenchmarkCollectAnalytics_Empty(b *testing.B) {
	results := []scanner.Result{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collectAnalytics(results)
	}
}

func BenchmarkCollectAnalytics_Small(b *testing.B) {
	results := make([]scanner.Result, 10)
	for i := 0; i < 10; i++ {
		results[i] = scanner.Result{
			IP:         "192.168.1." + string(rune('1'+i)),
			DeviceType: "Router",
			Protocols:  []string{"TCP", "UDP"},
			Ports:      []scanner.PortInfo{{Port: 22, State: "open", Protocol: "TCP", Service: "ssh"}},
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collectAnalytics(results)
	}
}

func BenchmarkNormalizeDeviceTypes_Empty(b *testing.B) {
	deviceTypes := map[string]int{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeDeviceTypes(deviceTypes)
	}
}

func BenchmarkNormalizeDeviceTypes_Small(b *testing.B) {
	deviceTypes := map[string]int{"Router": 2, "Switch": 3, "Server": 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeDeviceTypes(deviceTypes)
	}
}

func BenchmarkFormatDeviceValue_Empty(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatDeviceValue("")
	}
}

func BenchmarkFormatDeviceValue_NonEmpty(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatDeviceValue("Linux Server 22.04")
	}
}
