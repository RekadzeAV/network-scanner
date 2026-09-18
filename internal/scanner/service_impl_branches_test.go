package scanner

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"network-scanner/internal/contracts"
)

// TestScannerService_Scan_SuccessConversion — успешное завершение Scan
// через сервисный API: покрывает конвертацию результатов (непустой список).
func TestScannerService_Scan_SuccessConversion(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback listen unavailable: %v", err)
	}
	defer ln.Close()
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

	svc := NewService("debug")
	cfg := contracts.ScanConfig{
		NetworkCIDR: "127.0.0.1/30",
		PortRange:   fmt.Sprintf("%d,%d", port, port+1),
		Timeout:     200 * time.Millisecond,
		Threads:     4,
	}

	results, err := svc.Scan(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("expected successful scan, got %v", err)
	}
	if len(results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(results))
	}
	found := false
	for _, r := range results {
		for _, p := range r.Ports {
			if p.State == "open" && p.Port == port {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected open port %d in service results", port)
	}
}

// TestScannerService_Scan_AlreadyRunning — повторный вызов во время
// активного сканирования (ветка ошибки "уже запущено").
func TestScannerService_Scan_AlreadyRunning(t *testing.T) {
	svc := NewService("debug").(*scannerServiceImpl)
	cfg := contracts.ScanConfig{
		NetworkCIDR: "10.99.0.0/24",
		PortRange:   "1-64",
		Timeout:     8 * time.Second,
		Threads:     8,
	}

	firstDone := make(chan struct{})
	go func() {
		_, _ = svc.Scan(context.Background(), cfg, nil)
		close(firstDone)
	}()

	// Ждём устойчивой установки флага isScanning (тест в той же пачке).
	deadline := time.Now().Add(5 * time.Second)
	for !svc.isScanning.Load() && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}
	if !svc.isScanning.Load() {
		svc.Stop()
		<-firstDone
		t.Skip("scan finished before we could observe it as active")
	}

	if _, err := svc.Scan(context.Background(), cfg, nil); err == nil {
		t.Error("expected second concurrent Scan to be rejected")
	}
	svc.Stop()
	select {
	case <-firstDone:
	case <-time.After(30 * time.Second):
		t.Fatal("first scan did not finish after Stop")
	}
}

// TestScannerService_Stop_ActiveScan — Stop во время активного сканирования:
// покрывает cancel + ожидание done + сброс состояния.
func TestScannerService_Stop_ActiveScan(t *testing.T) {
	svc := NewService("debug")
	cfg := contracts.ScanConfig{
		NetworkCIDR: "10.99.1.0/24",
		PortRange:   "1-128",
		Timeout:     8 * time.Second,
		Threads:     16,
	}

	done := make(chan struct{})
	go func() {
		_, _ = svc.Scan(context.Background(), cfg, nil)
		close(done)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := svc.Scan(context.Background(), cfg, nil); err != nil {
			break // сканирование активно
		}
		time.Sleep(10 * time.Millisecond)
	}

	svc.Stop()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("scan did not stop after Stop()")
	}

	// Повторный Stop после завершения — идемпотентен (activeScan уже nil)
	svc.Stop()
}
