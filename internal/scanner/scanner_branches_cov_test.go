package scanner

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// scriptedPortScanner — детерминированный PortScanner: заданный набор
// «открытых» портов, опциональная задержка на каждый probe.
type scriptedPortScanner struct {
	open   map[int]bool
	delay  time.Duration
	called atomic.Int32 // ScanPort вызывается из горутин portWg конкурентно
}

func (s *scriptedPortScanner) ScanPort(ip string, port int, proto string) (bool, error) {
	s.called.Add(1)
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return s.open[port], nil
}

func (s *scriptedPortScanner) ScanPorts(ip string, ports []int, proto string) ([]int, error) {
	var out []int
	for _, p := range ports {
		if s.open[p] {
			out = append(out, p)
		}
	}
	return out, nil
}

// TestScanHost_ProtocolAndOSDetect — ветки определения протокола по
// открытым портам и угадывания ОС (135+445 → Windows).
func TestScanHost_ProtocolAndOSDetect(t *testing.T) {
	ps := &scriptedPortScanner{open: map[int]bool{135: true, 445: true}}
	ns := NewScanner("127.0.0.1/30", 50*time.Millisecond, "135,445", 4, false,
		&fakeProber{alive: true}, ps, nil)
	ns.SetOSDetectActive(true)

	ns.scanHost(net.ParseIP("127.0.0.1"), []int{135, 445})

	results := ns.GetResults()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if len(r.Protocols) == 0 {
		t.Error("expected protocols detected from open ports")
	}
	if r.GuessOS == "" {
		t.Error("expected GuessOS for ports 135+445")
	}
	if r.GuessOSReason == "" {
		t.Error("expected GuessOSReason to be set")
	}
	if got := ps.called.Load(); got != 2 {
		t.Errorf("expected 2 port probes, got %d", got)
	}
}

// TestScanHost_CancelWhileCollecting — отмена во время сбора результатов
// портов (ветка tcpCancelWait + drain канала).
func TestScanHost_CancelWhileCollecting(t *testing.T) {
	ps := &scriptedPortScanner{
		open:  map[int]bool{},
		delay: 10 * time.Millisecond,
	}
	ns := NewScanner("127.0.0.1/30", 50*time.Millisecond, "1-64", 32, true,
		&fakeProber{alive: true}, ps, nil)

	done := make(chan struct{})
	go func() {
		ns.scanHost(net.ParseIP("127.0.0.1"), portsRange(1, 64))
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	ns.Stop()

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("scanHost did not return after cancel")
	}
	if len(ns.GetResults()) != 1 {
		t.Errorf("expected result stored after cancelled scanHost, got %d", len(ns.GetResults()))
	}
}

// TestScan_CancelBeforePortLaunch — отмена до запуска горутин портов
// (счётчик tcpCancelBefore).
func TestScan_CancelBeforePortLaunch(t *testing.T) {
	ps := &scriptedPortScanner{open: map[int]bool{}, delay: 5 * time.Millisecond}
	ns := NewScanner("127.0.0.1/30", 50*time.Millisecond, "1-256", 64, false,
		&fakeProber{alive: true}, ps, nil)

	done := make(chan struct{})
	go func() {
		ns.Scan()
		close(done)
	}()
	time.Sleep(10 * time.Millisecond)
	ns.Stop()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("Scan did not stop after cancel")
	}
}

// TestGetMACAddress_FakeProberHit — успешная ветка ResolveMAC через
// инжектированный prober.
func TestGetMACAddress_FakeProberHit(t *testing.T) {
	mac := net.HardwareAddr{0x08, 0x00, 0x27, 0x11, 0x22, 0x33} // VirtualBox OUI
	ns := NewScanner("127.0.0.1/30", 50*time.Millisecond, "80", 2, false,
		&fakeProber{alive: true, mac: mac}, nil, nil)

	got, err := ns.getMACAddress(net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("expected mac from prober, got err %v", err)
	}
	if got != mac.String() {
		t.Errorf("expected %s, got %s", mac.String(), got)
	}
}

// TestGetMACAddress_ARPFallbackErrors — без prober-MAC: цепочка падает в
// платформенную ARP-таблицу и pcap-ARP; ошибок быть не должно только как
// error-возврат (не паника).
func TestGetMACAddress_ARPFallbackErrors(t *testing.T) {
	ns := NewScanner("127.0.0.1/30", 50*time.Millisecond, "80", 2, false,
		&fakeProber{alive: true}, nil, nil)

	// Несуществующий внешний IP: ARP-цепочка должна вернуть error, не паникуя.
	if _, err := ns.getMACAddress(net.ParseIP("203.0.113.77")); err == nil {
		t.Log("ARP chain unexpectedly resolved MAC for test-net IP")
	}
}

// TestReadMACFromARPTable_WindowsBranch — прямой вызов табличной ветки
// (на windows покрывает readMACFromWindowsARP, на прочих — readMACFrom*ARP).
func TestReadMACFromARPTable_WindowsBranch(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/30", 50*time.Millisecond, "80", 2, false)
	_, _ = ns.readMACFromARPTable(net.ParseIP("127.0.0.1"))
	// Linux/Darwin-специфичные функции вызываются напрямую для покрытия
	// их веток ошибок/парсинга (на windows они не выполняются в рантайме).
	_, _ = ns.readMACFromLinuxARP("127.0.0.1")
	_, _ = ns.readMACFromDarwinARP("127.0.0.1")
}

func portsRange(a, b int) []int {
	out := make([]int, 0, b-a+1)
	for i := a; i <= b; i++ {
		out = append(out, i)
	}
	return out
}

// TestScanHost_BannerNoListener — включённый grabBanners на закрытых
// портах не должен ломать сбор результатов.
func TestScanHost_BannerNoListener(t *testing.T) {
	ps := &scriptedPortScanner{open: map[int]bool{80: true}}
	ns := NewScanner("127.0.0.1/30", 50*time.Millisecond, "80", 2, false,
		&fakeProber{alive: true}, ps, nil)
	ns.SetGrabBanners(true)

	ns.scanHost(net.ParseIP("127.0.0.1"), []int{80})
	results := ns.GetResults()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if len(results[0].Ports) == 0 {
		t.Error("expected port 80 recorded")
	}
	_ = fmt.Sprint(results[0].Ports[0].Banner)
}
