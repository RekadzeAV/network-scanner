package scanner

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"network-scanner/internal/network"
)

// BenchmarkIsHostAlive — проверка живости хоста через TCP probe
func BenchmarkIsHostAlive(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ns.isHostAlive("127.0.0.1")
	}
}

// BenchmarkIsHostAliveContext — проверка живости с контекстом отмены
func BenchmarkIsHostAliveContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		prober := network.DefaultNetworkProber{Timeout: 1 * time.Second}
		_, _ = prober.PingContext("127.0.0.1", ctx.Done())
		cancel()
	}
}

// BenchmarkScanTCPPort — сканирование одного TCP порта
func BenchmarkScanTCPPort(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ns.scanTCPPort("127.0.0.1", 80)
	}
}

// BenchmarkScanHost — сканирование одного хоста (без UDP)
func BenchmarkScanHost(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	ports, _ := network.ParsePortRange("1-100")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.scanHost(net.ParseIP("127.0.0.1"), ports)
	}
}

// BenchmarkGetDiagnosticsSummary — генерация диагностической сводки
func BenchmarkGetDiagnosticsSummary(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	// Имитируем данные
	atomicStoreInt64(&ns.lastPingNs, 500000000)
	atomicStoreInt64(&ns.lastPortscanNs, 1500000000)
	atomicStoreInt64(&ns.lastTotalNs, 2000000000)
	atomicStoreInt64(&ns.tcpProbeTotal, 1000)
	atomicStoreInt64(&ns.tcpProbeOpen, 50)
	atomicStoreInt64(&ns.tcpProbeClosed, 950)
	atomicStoreInt64(&ns.udpProbeTotal, 500)
	atomicStoreInt64(&ns.udpProbeOpen, 10)
	atomicStoreInt64(&ns.udpProbeNoOpen, 490)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ns.GetDiagnosticsSummary()
	}
}

// BenchmarkPortThreadsForHost — расчет потоков для хоста
func BenchmarkPortThreadsForHost(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-1000", 50, false)
	ports := []int{1, 2, 3, 4, 5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ns.portThreadsForHost(len(ports))
	}
}

// BenchmarkDetectDeviceType — определение типа устройства
func BenchmarkDetectDeviceType(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	result := Result{
		IP:        "192.168.1.1",
		Hostname:  "router",
		Ports:     []PortInfo{{Port: 80, Protocol: "tcp", State: "open"}, {Port: 443, Protocol: "tcp", State: "open"}},
		Protocols: []string{"HTTP", "HTTPS"},
		MAC:       "aa:bb:cc:dd:ee:ff",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ns.detectDeviceType(result)
	}
}

// BenchmarkGetVendorFromMAC — определение производителя по MAC
func BenchmarkGetVendorFromMAC(b *testing.B) {
	mac := "aa:bb:cc:dd:ee:ff"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getVendorFromMAC(mac)
	}
}

// BenchmarkAppendIfNotExists — добавление в срез без дубликатов
func BenchmarkAppendIfNotExists(b *testing.B) {
	protocols := []string{"HTTP", "HTTPS", "SSH"}
	protocol := "FTP"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = appendIfNotExists(protocols, protocol)
	}
}

// BenchmarkGetProtocolFromPort — определение протокола по порту
func BenchmarkGetProtocolFromPort(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getProtocolFromPort(80)
	}
}

// BenchmarkHasOpenPort — проверка открытого порта
func BenchmarkHasOpenPort(b *testing.B) {
	ports := []PortInfo{
		{Port: 80, Protocol: "tcp", State: "open"},
		{Port: 443, Protocol: "tcp", State: "open"},
		{Port: 22, Protocol: "tcp", State: "closed"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasOpenPort(ports, 80, "tcp")
	}
}

// BenchmarkShouldGrabBannerPort — проверка необходимости сбора баннера
func BenchmarkShouldGrabBannerPort(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = shouldGrabBannerPort(80)
	}
}

// BenchmarkParseNetworkRange — парсинг CIDR диапазона
func BenchmarkParseNetworkRange(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = network.ParseNetworkRange("192.168.1.0/24")
	}
}

// BenchmarkParsePortRange — парсинг диапазона портов
func BenchmarkParsePortRange(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = network.ParsePortRange("1-1024")
	}
}

// BenchmarkGetServiceName — получение имени сервиса по порту
func BenchmarkGetServiceName(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = network.GetServiceName(80)
	}
}

// BenchmarkIsPortOpen — проверка открытости TCP порта
func BenchmarkIsPortOpen(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = network.IsPortOpen("127.0.0.1", 80, 1*time.Second)
	}
}

// BenchmarkGetDiagnosticsSummaryLarge — генерация сводки с большими данными
func BenchmarkGetDiagnosticsSummaryLarge(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-1000", 100, false)
	atomicStoreInt64(&ns.lastPingNs, 5000000000)
	atomicStoreInt64(&ns.lastPortscanNs, 15000000000)
	atomicStoreInt64(&ns.lastTotalNs, 20000000000)
	atomicStoreInt64(&ns.tcpProbeTotal, 100000)
	atomicStoreInt64(&ns.tcpProbeOpen, 5000)
	atomicStoreInt64(&ns.tcpProbeClosed, 95000)
	atomicStoreInt64(&ns.udpProbeTotal, 50000)
	atomicStoreInt64(&ns.udpProbeOpen, 1000)
	atomicStoreInt64(&ns.udpProbeNoOpen, 49000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ns.GetDiagnosticsSummary()
	}
}

// atomicStoreInt64 helper для установки значений в atomic fields
func atomicStoreInt64(ptr *int64, val int64) {
	*ptr = val
}

// BenchmarkNetworkScannerNew — создание нового сканера
func BenchmarkNetworkScannerNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	}
}

// BenchmarkNetworkScannerNewWithDI — создание сканера с DI
func BenchmarkNetworkScannerNewWithDI(b *testing.B) {
	prober := network.DefaultNetworkProber{Timeout: 2 * time.Second}
	scanner := network.TCPPortScanner{Timeout: 2 * time.Second}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewScanner(
			"192.168.1.0/24",
			2*time.Second,
			"1-100",
			50,
			false,
			prober,
			scanner,
			nil,
		)
	}
}

// BenchmarkScanHostSmall — сканирование хоста с малым диапазоном портов
func BenchmarkScanHostSmall(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-10", 50, false)
	ports, _ := network.ParsePortRange("1-10")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.scanHost(net.ParseIP("127.0.0.1"), ports)
	}
}

// BenchmarkScanHostLarge — сканирование хоста с большим диапазоном портов
func BenchmarkScanHostLarge(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-1000", 50, false)
	ports, _ := network.ParsePortRange("1-1000")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.scanHost(net.ParseIP("127.0.0.1"), ports)
	}
}

// BenchmarkGetResults — получение результатов (копирование с mutex)
func BenchmarkGetResults(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	// Добавляем тестовые результаты
	for i := 0; i < 100; i++ {
		ns.mu.Lock()
		ns.results = append(ns.results, Result{
			IP: fmt.Sprintf("192.168.1.%d", i+1),
		})
		ns.mu.Unlock()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ns.GetResults()
	}
}

// BenchmarkStop — остановка сканирования
func BenchmarkStop(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.Stop()
	}
}

// BenchmarkSetProgressCallback — установка callback
func BenchmarkSetProgressCallback(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	cb := func(stage string, current, total int, message string) {}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.SetProgressCallback(cb)
	}
}

// BenchmarkSetScanUDP — переключение UDP сканирования
func BenchmarkSetScanUDP(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.SetScanUDP(true)
		ns.SetScanUDP(false)
	}
}

// BenchmarkSetGrabBanners — переключение сбора баннеров
func BenchmarkSetGrabBanners(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.SetGrabBanners(true)
		ns.SetGrabBanners(false)
	}
}

// BenchmarkSetOSDetectActive — переключение активного определения ОС
func BenchmarkSetOSDetectActive(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.SetOSDetectActive(true)
		ns.SetOSDetectActive(false)
	}
}

// BenchmarkSetVerbosePortLogs — переключение детальных логов
func BenchmarkSetVerbosePortLogs(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.SetVerbosePortLogs(true)
		ns.SetVerbosePortLogs(false)
	}
}

// BenchmarkScanHostWithBanners — сканирование с включенным сбором баннеров
func BenchmarkScanHostWithBanners(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	ns.SetGrabBanners(true)
	ports, _ := network.ParsePortRange("1-100")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.scanHost(net.ParseIP("127.0.0.1"), ports)
	}
}

// BenchmarkScanHostWithOSDetect — сканирование с активным определением ОС
func BenchmarkScanHostWithOSDetect(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	ns.SetOSDetectActive(true)
	ports, _ := network.ParsePortRange("1-100")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.scanHost(net.ParseIP("127.0.0.1"), ports)
	}
}

// BenchmarkScanHostWithVerboseLogs — сканирование с детальными логами
func BenchmarkScanHostWithVerboseLogs(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	ns.SetVerbosePortLogs(true)
	ports, _ := network.ParsePortRange("1-100")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.scanHost(net.ParseIP("127.0.0.1"), ports)
	}
}

// BenchmarkScanHostWithAllFeatures — сканирование со всеми функциями
func BenchmarkScanHostWithAllFeatures(b *testing.B) {
	ns := NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-100", 50, false)
	ns.SetGrabBanners(true)
	ns.SetOSDetectActive(true)
	ns.SetVerbosePortLogs(true)
	ports, _ := network.ParsePortRange("1-100")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.scanHost(net.ParseIP("127.0.0.1"), ports)
	}
}
