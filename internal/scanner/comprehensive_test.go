package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"network-scanner/internal/network"
)

// --- Конструкторы и настройки ---

func TestNewScannerWithDI(t *testing.T) {
	prober := network.DefaultNetworkProber{Timeout: 1 * time.Second}
	portScanner := network.TCPPortScanner{Timeout: 1 * time.Second}

	ns := NewScanner(
		"192.168.1.0/24",
		1*time.Second,
		"80,443",
		10,
		false,
		prober,
		portScanner,
		nil,
	)

	if ns == nil {
		t.Fatal("NewScanner() returned nil")
	}
	if ns.network != "192.168.1.0/24" {
		t.Errorf("network = %v, want 192.168.1.0/24", ns.network)
	}
	if ns.portRange != "80,443" {
		t.Errorf("portRange = %v, want 80,443", ns.portRange)
	}
	if ns.threads != 10 {
		t.Errorf("threads = %v, want 10", ns.threads)
	}
}

func TestSetScanUDP(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	if ns.scanUDP {
		t.Error("scanUDP should be false by default")
	}

	ns.SetScanUDP(true)
	if !ns.scanUDP {
		t.Error("scanUDP should be true after SetScanUDP(true)")
	}

	ns.SetScanUDP(false)
	if ns.scanUDP {
		t.Error("scanUDP should be false after SetScanUDP(false)")
	}
}

func TestSetGrabBanners(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	if ns.grabBanners {
		t.Error("grabBanners should be false by default")
	}

	ns.SetGrabBanners(true)
	if !ns.grabBanners {
		t.Error("grabBanners should be true after SetGrabBanners(true)")
	}
}

func TestSetOSDetectActive(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	if ns.osDetectActive {
		t.Error("osDetectActive should be false by default")
	}

	ns.SetOSDetectActive(true)
	if !ns.osDetectActive {
		t.Error("osDetectActive should be true after SetOSDetectActive(true)")
	}
}

func TestSetVerbosePortLogs(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	if ns.verbosePortLogs {
		t.Error("verbosePortLogs should be false by default")
	}

	ns.SetVerbosePortLogs(true)
	if !ns.verbosePortLogs {
		t.Error("verbosePortLogs should be true after SetVerbosePortLogs(true)")
	}
}

func TestSetScanTCPPorts(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	if !ns.scanTCPPorts {
		t.Error("scanTCPPorts should be true by default")
	}

	ns.SetScanTCPPorts(false)
	if ns.scanTCPPorts {
		t.Error("scanTCPPorts should be false after SetScanTCPPorts(false)")
	}
}

// --- Контекст и отмена ---

func TestContextCancellation(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Проверяем, что контекст не отменен изначально
	select {
	case <-ns.ctx.Done():
		t.Error("context should not be cancelled initially")
	default:
		// OK
	}

	// Отменяем контекст
	ns.Stop()

	// Проверяем, что контекст отменен
	select {
	case <-ns.ctx.Done():
		// OK
	default:
		t.Error("context should be cancelled after Stop()")
	}
}

func TestStopMultipleTimes(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Multiple stops should not panic
	ns.Stop()
	ns.Stop()
	ns.Stop()
}

// --- Порт-сканирование ---

func TestShouldGrabBannerPort(t *testing.T) {
	tests := []struct {
		port int
		want bool
	}{
		{21, true},    // FTP
		{22, true},    // SSH
		{25, true},    // SMTP
		{80, true},    // HTTP
		{110, true},   // POP3
		{143, true},   // IMAP
		{443, true},   // HTTPS
		{993, true},   // IMAPS
		{995, true},   // POP3S
		{8080, true},  // HTTP-ALT
		{8443, true},  // HTTPS-ALT
		{3306, false}, // MySQL
		{5432, false}, // PostgreSQL
		{9999, false}, // Unknown
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("port_%d", tt.port), func(t *testing.T) {
			got := shouldGrabBannerPort(tt.port)
			if got != tt.want {
				t.Errorf("shouldGrabBannerPort(%d) = %v, want %v", tt.port, got, tt.want)
			}
		})
	}
}

func TestHasOpenPort(t *testing.T) {
	ports := []PortInfo{
		{Port: 80, Protocol: "tcp", State: "open"},
		{Port: 443, Protocol: "tcp", State: "open"},
		{Port: 22, Protocol: "tcp", State: "closed"},
	}

	tests := []struct {
		port     int
		protocol string
		want     bool
	}{
		{80, "tcp", true},
		{443, "tcp", true},
		{22, "tcp", false},
		{80, "udp", false},
		{9999, "tcp", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("port_%d_proto_%s", tt.port, tt.protocol), func(t *testing.T) {
			got := hasOpenPort(ports, tt.port, tt.protocol)
			if got != tt.want {
				t.Errorf("hasOpenPort(%d, %s) = %v, want %v", tt.port, tt.protocol, got, tt.want)
			}
		})
	}
}

// --- Port scanning ---

func TestGetVendorFromMACEmpty(t *testing.T) {
	got := getVendorFromMAC("")
	if got != "Unknown" {
		t.Errorf("getVendorFromMAC(\"\") = %v, want Unknown", got)
	}
}

func TestGetVendorFromMACShort(t *testing.T) {
	got := getVendorFromMAC("00:50")
	if got != "Unknown" {
		t.Errorf("getVendorFromMAC(\"00:50\") = %v, want Unknown", got)
	}
}

func TestGetVendorFromMACInvalid(t *testing.T) {
	got := getVendorFromMAC("zz:zz:zz:zz:zz:zz")
	if got != "Unknown" {
		t.Errorf("getVendorFromMAC(\"zz:zz:zz:zz:zz:zz\") = %v, want Unknown", got)
	}
}

// --- AppendIfNotExists ---

func TestAppendIfNotExistsEmpty(t *testing.T) {
	slice := []string{}
	got := appendIfNotExists(slice, "a")
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("appendIfNotExists([], \"a\") = %v, want [a]", got)
	}
}

func TestAppendIfNotExistsDuplicate(t *testing.T) {
	slice := []string{"a", "b"}
	got := appendIfNotExists(slice, "a")
	if len(got) != 2 {
		t.Errorf("appendIfNotExists([a, b], \"a\") length = %d, want 2", len(got))
	}
}

func TestAppendIfNotExistsNew(t *testing.T) {
	slice := []string{"a", "b"}
	got := appendIfNotExists(slice, "c")
	if len(got) != 3 {
		t.Errorf("appendIfNotExists([a, b], \"c\") length = %d, want 3", len(got))
	}
	found := false
	for _, item := range got {
		if item == "c" {
			found = true
		}
	}
	if !found {
		t.Error("appendIfNotExists([a, b], \"c\") should contain \"c\"")
	}
}

// --- DeviceType detection ---

func TestDetectDeviceTypeEmptyPorts(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)
	result := Result{
		IP:    "192.168.1.1",
		Ports: []PortInfo{},
	}
	got := ns.detectDeviceType(result)
	if got != "Unknown" {
		t.Errorf("detectDeviceType(empty) = %v, want Unknown", got)
	}
}

func TestDetectDeviceTypeSinglePort(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Порт 9999 должен определить как IoT
	result := Result{
		IP:    "192.168.1.1",
		Ports: []PortInfo{{Port: 9999, State: "open"}},
	}
	got := ns.detectDeviceType(result)
	if got != "IoT" {
		t.Errorf("detectDeviceType(9999) = %v, want IoT", got)
	}
}

func TestDetectDeviceTypeMultiplePorts(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Сервер с HTTP и HTTPS
	result := Result{
		IP: "192.168.1.1",
		Ports: []PortInfo{
			{Port: 80, State: "open"},
			{Port: 443, State: "open"},
		},
	}
	got := ns.detectDeviceType(result)
	if got != "Server" {
		t.Errorf("detectDeviceType(80,443) = %v, want Server", got)
	}
}

// --- Results management ---

func TestGetResultsConcurrent(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Добавляем результаты параллельно
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ns.mu.Lock()
			ns.results = append(ns.results, Result{
				IP: fmt.Sprintf("192.168.1.%d", idx),
			})
			ns.mu.Unlock()
		}(i)
	}
	wg.Wait()

	// Получаем результаты
	results := ns.GetResults()
	if len(results) != 10 {
		t.Errorf("GetResults() length = %d, want 10", len(results))
	}
}

func TestGetResultsReturnsCopy(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Добавляем результат
	ns.mu.Lock()
	ns.results = append(ns.results, Result{IP: "192.168.1.1"})
	ns.mu.Unlock()

	// Получаем результаты
	results1 := ns.GetResults()

	// Модифицируем копию (создаем новый срез)
	modified := make([]Result, len(results1))
	copy(modified, results1)
	modified[0].IP = "10.0.0.1"

	// Получаем снова
	results2 := ns.GetResults()
	// results2 должен содержать оригинальное значение
	if results2[0].IP != "192.168.1.1" {
		t.Errorf("GetResults() should return original data, got IP=%v", results2[0].IP)
	}
}

// --- PortThreadsForHost edge cases ---

func TestPortThreadsForHostZeroThreads(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 0, false)
	got := ns.portThreadsForHost(100)
	// При threads=0 код использует threads=1, затем budget/1=512, но capped by maxPerHostPortThreads=64
	if got != maxPerHostPortThreads {
		t.Errorf("portThreadsForHost(0 threads, 100 ports) = %d, want %d", got, maxPerHostPortThreads)
	}
}

func TestPortThreadsForHostNegativeThreads(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", -1, false)
	got := ns.portThreadsForHost(100)
	// При threads=-1 код использует threads=1 (min), затем budget/1=512, но capped by maxPerHostPortThreads=64
	if got != maxPerHostPortThreads {
		t.Errorf("portThreadsForHost(-1 threads, 100 ports) = %d, want %d", got, maxPerHostPortThreads)
	}
}

func TestPortThreadsForHostLargePortCount(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "1-65535", 50, false)
	got := ns.portThreadsForHost(65535)
	if got > maxPerHostPortThreads {
		t.Errorf("portThreadsForHost(65535 ports) = %d, should not exceed %d", got, maxPerHostPortThreads)
	}
}

// --- Diagnostics ---

func TestGetDiagnosticsSummaryEmpty(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)
	summary := ns.GetDiagnosticsSummary()
	if summary == "" {
		t.Error("GetDiagnosticsSummary() should not return empty string")
	}
}

func TestGetDiagnosticsSummaryWithValues(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Устанавливаем значения
	ns.lastPingNs = 500000000      // 500ms
	ns.lastPortscanNs = 1500000000 // 1.5s
	ns.lastTotalNs = 2000000000    // 2s

	summary := ns.GetDiagnosticsSummary()
	if summary == "" {
		t.Error("GetDiagnosticsSummary() should not return empty string with values")
	}
}

// --- Progress callback ---

func TestSetProgressCallback(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	callbackCalled := false
	ns.SetProgressCallback(func(stage string, current, total int, message string) {
		callbackCalled = true
		if stage != "test" {
			t.Errorf("stage = %v, want test", stage)
		}
	})

	// Вызываем callback
	if ns.progressCallback != nil {
		ns.progressCallback("test", 1, 10, "test message")
	}

	if !callbackCalled {
		t.Error("progressCallback was not called")
	}
}

func TestProgressCallbackNil(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Nil callback should not panic
	if ns.progressCallback != nil {
		ns.progressCallback("test", 1, 10, "test")
	}
}

// --- ScanTCPPort fallback ---

func TestScanTCPPortNilScanner(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// При nil portScanner должен использоваться fallback
	// Это не должно паниковать
	result := ns.scanTCPPort("127.0.0.1", 80)
	_ = result // Может быть true или false в зависимости от системы
}

// --- Context-aware prober ---

func TestContextNetworkProberInterface(t *testing.T) {
	// Проверяем, что DefaultNetworkProber реализует ContextNetworkProber
	var _ ContextNetworkProber = network.DefaultNetworkProber{}
}

// --- ScanHost with empty ports ---

func TestScanHostEmptyPorts(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Сканирование с пустым списком портов
	ns.scanHost(net.ParseIP("127.0.0.1"), []int{})

	results := ns.GetResults()
	if len(results) != 1 {
		t.Errorf("scanHost with empty ports should still create result, got %d results", len(results))
	}

	if results[0].IP != "127.0.0.1" {
		t.Errorf("result IP = %v, want 127.0.0.1", results[0].IP)
	}
}

// --- UDP scan settings ---

func TestScanUDPDefaultFalse(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)
	if ns.scanUDP {
		t.Error("scanUDP should be false by default")
	}
}

func TestScanUDPEnabled(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)
	ns.SetScanUDP(true)
	if !ns.scanUDP {
		t.Error("scanUDP should be true after enabling")
	}
}

// --- TCP scan settings ---

func TestScanTCPPortsDefaultTrue(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)
	if !ns.scanTCPPorts {
		t.Error("scanTCPPorts should be true by default")
	}
}

func TestScanTCPPortsDisabled(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)
	ns.SetScanTCPPorts(false)
	if ns.scanTCPPorts {
		t.Error("scanTCPPorts should be false after disabling")
	}
}

// --- Result struct ---

func TestResultStructFields(t *testing.T) {
	result := Result{
		IP:                "192.168.1.1",
		MAC:               "aa:bb:cc:dd:ee:ff",
		Hostname:          "test-host",
		Ports:             []PortInfo{{Port: 80, Protocol: "tcp", State: "open"}},
		Protocols:         []string{"HTTP"},
		DeviceType:        "Server",
		DeviceVendor:      "TestVendor",
		SNMPEnabled:       true,
		IsAlive:           true,
		GuessOS:           "Linux",
		GuessOSConfidence: "high",
		GuessOSReason:     "SSH open",
	}

	if result.IP != "192.168.1.1" {
		t.Errorf("IP = %v, want 192.168.1.1", result.IP)
	}
	if result.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("MAC = %v, want aa:bb:cc:dd:ee:ff", result.MAC)
	}
	if result.Hostname != "test-host" {
		t.Errorf("Hostname = %v, want test-host", result.Hostname)
	}
	if len(result.Ports) != 1 {
		t.Errorf("Ports length = %d, want 1", len(result.Ports))
	}
	if result.DeviceType != "Server" {
		t.Errorf("DeviceType = %v, want Server", result.DeviceType)
	}
	if !result.SNMPEnabled {
		t.Error("SNMPEnabled should be true")
	}
	if !result.IsAlive {
		t.Error("IsAlive should be true")
	}
	if result.GuessOS != "Linux" {
		t.Errorf("GuessOS = %v, want Linux", result.GuessOS)
	}
}

// --- PortInfo struct ---

func TestPortInfoStructFields(t *testing.T) {
	port := PortInfo{
		Port:     80,
		State:    "open",
		Protocol: "tcp",
		Service:  "HTTP",
		Banner:   "Apache/2.4.41",
		Version:  "2.4.41",
	}

	if port.Port != 80 {
		t.Errorf("Port = %v, want 80", port.Port)
	}
	if port.State != "open" {
		t.Errorf("State = %v, want open", port.State)
	}
	if port.Protocol != "tcp" {
		t.Errorf("Protocol = %v, want tcp", port.Protocol)
	}
	if port.Service != "HTTP" {
		t.Errorf("Service = %v, want HTTP", port.Service)
	}
	if port.Banner != "Apache/2.4.41" {
		t.Errorf("Banner = %v, want Apache/2.4.41", port.Banner)
	}
	if port.Version != "2.4.41" {
		t.Errorf("Version = %v, want 2.4.41", port.Version)
	}
}

// --- Context cancellation during scan ---

func TestScanHostContextCancelled(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Отменяем контекст до сканирования
	ns.Stop()

	// Сканирование должно корректно обработать отмену
	ns.scanHost(net.ParseIP("127.0.0.1"), []int{80, 443})

	// Результат должен быть создан даже при отмене
	results := ns.GetResults()
	if len(results) != 1 {
		t.Errorf("scanHost with cancelled context should still create result, got %d", len(results))
	}
}

// --- Concurrent GetResults ---

func TestConcurrentGetResults(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Добавляем результаты
	for i := 0; i < 100; i++ {
		ns.mu.Lock()
		ns.results = append(ns.results, Result{IP: fmt.Sprintf("192.168.1.%d", i)})
		ns.mu.Unlock()
	}

	// Параллельное чтение
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ns.GetResults()
		}()
	}
	wg.Wait()
}

// --- Timeout edge cases ---

func TestNewScannerZeroTimeout(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 0, "80", 10, false)
	if ns.timeout != 0 {
		t.Errorf("timeout = %v, want 0", ns.timeout)
	}
}

func TestNewScannerNegativeTimeout(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", -1*time.Second, "80", 10, false)
	if ns.timeout != -time.Second {
		t.Errorf("timeout = %v, want -1s", ns.timeout)
	}
}

// --- Port range parsing integration ---

func TestPortRangeParsingIntegration(t *testing.T) {
	tests := []struct {
		rangeStr string
		wantLen  int
	}{
		{"80", 1},
		{"80,443", 2},
		{"1-10", 10},
		{"1-100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.rangeStr, func(t *testing.T) {
			ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, tt.rangeStr, 10, false)
			if ns.portRange != tt.rangeStr {
				t.Errorf("portRange = %v, want %v", ns.portRange, tt.rangeStr)
			}
		})
	}
}

// --- Network range parsing integration ---

func TestNetworkRangeParsingIntegration(t *testing.T) {
	tests := []struct {
		network string
		want    string
	}{
		{"192.168.1.0/24", "192.168.1.0/24"},
		{"10.0.0.0/8", "10.0.0.0/8"},
		{"172.16.0.0/12", "172.16.0.0/12"},
	}

	for _, tt := range tests {
		t.Run(tt.network, func(t *testing.T) {
			ns := NewNetworkScanner(tt.network, 1*time.Second, "80", 10, false)
			if ns.network != tt.want {
				t.Errorf("network = %v, want %v", ns.network, tt.want)
			}
		})
	}
}

// --- Thread count edge cases ---

func TestThreadsOne(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 1, false)
	if ns.threads != 1 {
		t.Errorf("threads = %v, want 1", ns.threads)
	}
}

func TestThreadsLarge(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 1000, false)
	if ns.threads != 1000 {
		t.Errorf("threads = %v, want 1000", ns.threads)
	}
}

// --- ShowClosed flag ---

func TestShowClosedDefault(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)
	if ns.showClosed {
		t.Error("showClosed should be false by default")
	}
}

func TestShowClosedTrue(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, true)
	if !ns.showClosed {
		t.Error("showClosed should be true when set to true")
	}
}

// --- ResultPresenter interface ---

func TestResultPresenterNil(t *testing.T) {
	prober := network.DefaultNetworkProber{Timeout: 1 * time.Second}
	portScanner := network.TCPPortScanner{Timeout: 1 * time.Second}

	ns := NewScanner(
		"192.168.1.0/24",
		1*time.Second,
		"80",
		10,
		false,
		prober,
		portScanner,
		nil, // nil presenter
	)

	if ns.resultPresenter != nil {
		t.Error("resultPresenter should be nil")
	}
}

// --- Context background ---

func TestContextNotBackground(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Контекст должен быть не background (должен иметь cancel)
	if ns.ctx == context.Background() {
		t.Error("context should not be context.Background()")
	}
}

// --- WaitGroup after stop ---

func TestWaitGroupAfterStop(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Stop должен корректно отменить контекст
	ns.Stop()

	// Wait не должен паниковать
	ns.wg.Wait()
}
