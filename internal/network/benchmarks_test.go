package network

import (
	"net"
	"testing"
	"time"
)

// BenchmarkDetectLocalNetwork — определение локальной сети
func BenchmarkDetectLocalNetwork(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DetectLocalNetwork()
	}
}

// BenchmarkParseNetworkRange — парсинг CIDR
func BenchmarkParseNetworkRange(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseNetworkRange("192.168.1.0/24")
	}
}

// BenchmarkParseNetworkRangeLarge — парсинг большого CIDR
func BenchmarkParseNetworkRangeLarge(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseNetworkRange("10.0.0.0/16")
	}
}

// BenchmarkParsePortRange — парсинг диапазона портов
func BenchmarkParsePortRange(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParsePortRange("1-1024")
	}
}

// BenchmarkParsePortRangeLarge — парсинг большого диапазона портов
func BenchmarkParsePortRangeLarge(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParsePortRange("1-65535")
	}
}

// BenchmarkParsePortRangeList — парсинг списка портов
func BenchmarkParsePortRangeList(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParsePortRange("22,80,443,8080,8443")
	}
}

// BenchmarkParsePortRangeMixed — парсинг смешанного диапазона
func BenchmarkParsePortRangeMixed(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParsePortRange("22,80,443,8080-8090,9000")
	}
}

// BenchmarkIsPortOpen — проверка TCP порта
func BenchmarkIsPortOpen(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsPortOpen("127.0.0.1", 80, 1*time.Second)
	}
}

// BenchmarkIsUDPPortOpen — проверка UDP порта
func BenchmarkIsUDPPortOpen(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsUDPPortOpen("127.0.0.1", 53, 1*time.Second)
	}
}

// BenchmarkGetServiceName — получение имени сервиса
func BenchmarkGetServiceName(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetServiceName(80)
	}
}

// BenchmarkGetServiceNameNotFound — сервис не найден
func BenchmarkGetServiceNameNotFound(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetServiceName(12345)
	}
}

// BenchmarkDefaultNetworkProberPing — проверка живости через prober
func BenchmarkDefaultNetworkProberPing(b *testing.B) {
	prober := DefaultNetworkProber{Timeout: 1 * time.Second}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = prober.Ping("127.0.0.1")
	}
}

// BenchmarkDefaultNetworkProberPingContext — проверка живости с контекстом
func BenchmarkDefaultNetworkProberPingContext(b *testing.B) {
	prober := DefaultNetworkProber{Timeout: 1 * time.Second}
	done := make(chan struct{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = prober.PingContext("127.0.0.1", done)
	}
}

// BenchmarkTCPPortScannerScanPort — сканирование TCP порта
func BenchmarkTCPPortScannerScanPort(b *testing.B) {
	scanner := TCPPortScanner{Timeout: 1 * time.Second}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.ScanPort("127.0.0.1", 80, "tcp")
	}
}

// BenchmarkUDPPortScannerScanPort — сканирование UDP порта
func BenchmarkUDPPortScannerScanPort(b *testing.B) {
	scanner := UDPPortScanner{Timeout: 1 * time.Second}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.ScanPort("127.0.0.1", 53, "udp")
	}
}

// BenchmarkResolveMAC — разрешение MAC адреса
func BenchmarkResolveMAC(b *testing.B) {
	prober := DefaultNetworkProber{Timeout: 1 * time.Second}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = prober.ResolveMAC("127.0.0.1")
	}
}

// BenchmarkParseIP — парсинг IP адреса
func BenchmarkParseIP(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = net.ParseIP("192.168.1.1")
	}
}

// BenchmarkJoinHostPort — объединение хоста и порта
func BenchmarkJoinHostPort(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = net.JoinHostPort("192.168.1.1", "80")
	}
}

// BenchmarkDetectLocalNetworkMultiple — определение нескольких сетей
func BenchmarkDetectLocalNetworkMultiple(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DetectLocalNetwork()
		_, _ = DetectLocalNetwork()
	}
}

// BenchmarkParseNetworkRangeEdge — парсинг edge cases
func BenchmarkParseNetworkRangeEdge(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseNetworkRange("192.168.1.128/25")
	}
}

// BenchmarkParsePortRangeEdge — парсинг edge cases
func BenchmarkParsePortRangeEdge(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParsePortRange("0-65535")
	}
}

// BenchmarkIsPortOpenMultiple — проверка нескольких портов
func BenchmarkIsPortOpenMultiple(b *testing.B) {
	ports := []int{80, 443, 22, 8080, 8443}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, port := range ports {
			_ = IsPortOpen("127.0.0.1", port, 1*time.Second)
		}
	}
}

// BenchmarkGetServiceNameMultiple — получение имен нескольких сервисов
func BenchmarkGetServiceNameMultiple(b *testing.B) {
	ports := []int{21, 22, 25, 53, 80, 110, 143, 443, 993, 995}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, port := range ports {
			_ = GetServiceName(port)
		}
	}
}

// BenchmarkParseNetworkRangeStress — стресс-тест парсинга сети
func BenchmarkParseNetworkRangeStress(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseNetworkRange("172.16.0.0/12")
	}
}

// BenchmarkParsePortRangeStress — стресс-тест парсинга портов
func BenchmarkParsePortRangeStress(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParsePortRange("1-2048")
	}
}

// BenchmarkDefaultNetworkProberPingStress — стресс-тест проверки живости
func BenchmarkDefaultNetworkProberPingStress(b *testing.B) {
	prober := DefaultNetworkProber{Timeout: 500 * time.Millisecond}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = prober.Ping("127.0.0.1")
	}
}
