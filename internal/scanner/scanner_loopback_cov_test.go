package scanner

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// fakeProber возвращает детерминированные ответы живости/MAC,
// чтобы полный флоу Scan() покрывался без реальной сети.
type fakeProber struct {
	alive bool
	mac   net.HardwareAddr
	err   error
}

func (p *fakeProber) Ping(ip string) (bool, error) {
	if p.err != nil {
		return false, p.err
	}
	return p.alive, nil
}

func (p *fakeProber) ResolveMAC(ip string) (net.HardwareAddr, error) {
	if p.mac != nil {
		return p.mac, nil
	}
	return nil, fmt.Errorf("no mac for %s", ip)
}

// fakeICMPPinger — подменяемый ICMP-пингер.
//
// calls/lastTTL фиксируют факт вызова и таймаут, переданный пингеру
// (используются в тестах ветки ICMP в icmp_ping_test.go).
type fakeICMPPinger struct {
	alive   bool
	err     error
	calls   int
	lastTTL time.Duration
}

func (p *fakeICMPPinger) PingICMP(host string, timeout time.Duration) (bool, error) {
	p.calls++
	p.lastTTL = timeout
	return p.alive, p.err
}

// openListenerOn находит свободный порт и слушает его.
func openListenerOn(t *testing.T) (net.Listener, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback listen unavailable: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	return ln, port
}

func newLoopbackScanner(port int, prober NetworkProber) *NetworkScanner {
	ns := NewScanner("127.0.0.1/30", 200*time.Millisecond, fmt.Sprintf("%d,%d", port, port+1), 4, true,
		prober, nil, nil)
	return ns
}

// TestScan_FullFlowLoopback — полный флоу Scan() на loopback:
// alive-ветка, горутинный обход хостов, определение протокола по открытому порту,
// сбор MAC и hostname.
func TestScan_FullFlowLoopback(t *testing.T) {
	ln, port := openListenerOn(t)
	defer ln.Close()

	prober := &fakeProber{
		alive: true,
		mac:   net.HardwareAddr{0x00, 0x50, 0x56, 0xAA, 0xBB, 0xCC}, // VMware OUI
	}
	ns := newLoopbackScanner(port, prober)

	progressCalls := 0
	ns.SetProgressCallback(func(stage string, current, total int, message string) {
		progressCalls++
	})

	done := make(chan struct{})
	go func() {
		ns.Scan()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("Scan() did not finish in time")
	}

	results := ns.GetResults()
	if len(results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(results))
	}
	hasOpen := false
	for _, r := range results {
		for _, p := range r.Ports {
			if p.State == "open" && p.Port == port {
				hasOpen = true
			}
		}
	}
	if !hasOpen {
		t.Errorf("expected open port %d in results", port)
	}
	if progressCalls == 0 {
		t.Error("expected progress callback invocations")
	}
}

// TestScan_LoopbackUDPEnabled — ветка UDP-сканирования хоста.
func TestScan_LoopbackUDPEnabled(t *testing.T) {
	ln, port := openListenerOn(t)
	defer ln.Close()

	prober := &fakeProber{alive: true}
	ns := newLoopbackScanner(port, prober)
	ns.SetScanUDP(true)

	done := make(chan struct{})
	go func() {
		ns.Scan()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("Scan() with UDP did not finish in time")
	}
	if len(ns.GetResults()) < 1 {
		t.Errorf("expected results with UDP, got %d", len(ns.GetResults()))
	}
}

// TestScan_LoopbackBannersOSDetect — сбор баннеров и активный OS-detect.
func TestScan_LoopbackBannersOSDetect(t *testing.T) {
	ln, port := openListenerOn(t)
	defer ln.Close()

	prober := &fakeProber{alive: true}
	ns := newLoopbackScanner(port, prober)
	ns.SetGrabBanners(true)
	ns.SetOSDetectActive(true)
	ns.Scan()
	if len(ns.GetResults()) < 1 {
		t.Errorf("expected results with banners, got %d", len(ns.GetResults()))
	}
}

// TestScan_CancelDuringPortScan — отмена сканирования во время обхода портов.
func TestScan_CancelDuringPortScan(t *testing.T) {
	ln, _ := openListenerOn(t)
	defer ln.Close()

	prober := &fakeProber{alive: true}
	ns := NewScanner("127.0.0.1/30", 100*time.Millisecond, "1-256", 16, false,
		prober, nil, nil)

	done := make(chan struct{})
	go func() {
		ns.Scan()
		close(done)
	}()
	time.Sleep(80 * time.Millisecond)
	ns.Stop()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("cancelled Scan() did not stop")
	}
	// Повторный Stop идемпотентен
	ns.Stop()
}

// TestIsHostAlive_ICMPFakePinger — ветка живости через подменный ICMP-пингер.
func TestIsHostAlive_ICMPFakePinger(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/30", 100*time.Millisecond, "80", 2, false)
	ns.SetICMPPingEnabled(true)
	ns.SetICMPPinger(&fakeICMPPinger{alive: true})
	if !ns.isHostAlive("127.0.0.1") {
		t.Error("expected alive via fake ICMP pinger")
	}
	ns.SetICMPPinger(&fakeICMPPinger{alive: false})
	if ns.isHostAlive("127.0.0.1") {
		t.Error("expected dead via fake ICMP pinger")
	}
	ns.SetICMPPinger(&fakeICMPPinger{err: fmt.Errorf("icmp blocked")})
	// Ошибка ICMP — фолбэк на порты, вердикт не должен паниковать
	_ = ns.isHostAlive("127.0.0.1")
}

// TestIsHostAlive_FakeProberBranches — ветки кастомного NetworkProber
// (context-aware и обычный) plus ошибка пробера.
func TestIsHostAlive_FakeProberBranches(t *testing.T) {
	ns := NewScanner("127.0.0.1/30", 100*time.Millisecond, "80", 2, false,
		&fakeProber{alive: true}, nil, nil)
	if !ns.isHostAlive("127.0.0.1") {
		t.Error("expected alive via fake prober")
	}

	failing := NewScanner("127.0.0.1/30", 50*time.Millisecond, "80", 2, false,
		&fakeProber{err: fmt.Errorf("probe down")}, nil, nil)
	_ = failing.isHostAlive("127.0.0.1") // фолбэк на commonPorts, не паникует
}

// TestScanHost_UDPOnlyBranch — прямой вызов scanHost с включенным UDP.
func TestScanHost_UDPOnlyBranch(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/30", 50*time.Millisecond, "80", 2, false)
	ns.SetScanUDP(true)
	ns.scanHost(net.ParseIP("127.0.0.1"), []int{80})
	if len(ns.GetResults()) != 1 {
		t.Fatalf("expected 1 stored result, got %d", len(ns.GetResults()))
	}
}
