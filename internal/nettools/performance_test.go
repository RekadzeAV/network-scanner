package nettools

import (
	"context"
	"testing"
	"time"
)

// === Performance: Ping ===

func BenchmarkBuildPingArgs_Windows(b *testing.B) {
	host := "192.168.1.1"
	count := 4
	goos := "windows"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildPingArgs(host, count, goos)
	}
}

func BenchmarkBuildPingArgs_Linux(b *testing.B) {
	host := "192.168.1.1"
	count := 4
	goos := "linux"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildPingArgs(host, count, goos)
	}
}

// === Performance: DNS ===

func BenchmarkNormalizeDNSError(b *testing.B) {
	errs := []error{
		context.DeadlineExceeded,
		context.Canceled,
		nil,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, e := range errs {
			_ = normalizeDNSError(e)
		}
	}
}

// === Performance: Whois ===

func BenchmarkBuildRDAPURL_Domain(b *testing.B) {
	query := "example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildRDAPURL(query)
	}
}

func BenchmarkBuildRDAPURL_IP(b *testing.B) {
	query := "8.8.8.8"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildRDAPURL(query)
	}
}

func BenchmarkResolveRDAPBaseURL(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolveRDAPBaseURL()
	}
}

// === Performance: Parse ===

func BenchmarkParsePingStats_Windows(b *testing.B) {
	raw := "Pinging 192.168.1.1 with 32 bytes of data:\nReply from 192.168.1.1: bytes=32 time=1ms TTL=128\nLost = 0 (0% loss)\nMinimum = 1ms, Maximum = 1ms, Average = 1ms"
	count := 4
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsePingStats(raw, count)
	}
}

func BenchmarkParsePingStats_Unix(b *testing.B) {
	raw := "PING 192.168.1.1 (192.168.1.1) 56(84) bytes of data.\n\n--- 192.168.1.1 ping statistics ---\n4 packets transmitted, 4 received, 0% packet loss"
	count := 4
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsePingStats(raw, count)
	}
}

func BenchmarkParseTraceroute(b *testing.B) {
	raw := " 1  192.168.1.1  1.234 ms  1.123 ms  1.098 ms\n 2  10.0.0.1  5.678 ms  5.432 ms  5.321 ms\n 3  * * *"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseTraceroute(raw)
	}
}

func BenchmarkParseFloatMs(b *testing.B) {
	values := []string{"1.234ms", "0ms", "100.000ms", ""}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range values {
			_ = parseFloatMs(v)
		}
	}
}

// === Performance: WiFi ===

func BenchmarkParseWindowsNetsh(b *testing.B) {
	raw := `
SSID 1 : MyNetwork
   Type                   : Managed
   Description            : Managed network
   Channel                : 6
   RSSI                   : 80
   Authenticated          : Yes
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseWindowsNetsh(raw)
	}
}

func BenchmarkParseLinuxNmcli(b *testing.B) {
	raw := `
SSID: MyNetwork
BSSID: AA:BB:CC:DD:EE:FF
MODE: Infra
CHAN: 6
RATE: 144 Mbit/s
SIGNAL: 80
SECURITY: WPA2
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseLinuxNmcli(raw)
	}
}

func BenchmarkParseDarwinAirport(b *testing.B) {
	raw := `
AirPort: On
Wireless Network: MyNetwork
Status: Connected
SSID: MyNetwork
Channel: 6
Signal Strength: 80%
Security: WPA2 Personal
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseDarwinAirport(raw)
	}
}

// === Performance: Build Traceroute Args ===

func BenchmarkBuildTracerouteArgs_Windows(b *testing.B) {
	host := "192.168.1.1"
	maxHops := 30
	goos := "windows"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildTracerouteArgs(host, maxHops, goos)
	}
}

func BenchmarkBuildTracerouteArgs_Linux(b *testing.B) {
	host := "192.168.1.1"
	maxHops := 30
	goos := "linux"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildTracerouteArgs(host, maxHops, goos)
	}
}

// === Performance: Context ===

func BenchmarkContextWithTimeout(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cancel()
	}
}
