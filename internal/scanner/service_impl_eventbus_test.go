package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/eventbus"
)

// TestScannerService_EventBus_StartedAndCompleted — E6: при заданной шине
// сервис публикует scan.started и scan.completed.
func TestScannerService_EventBus_StartedAndCompleted(t *testing.T) {
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

	bus := eventbus.NewEventBus()
	defer bus.Close()

	var mu sync.Mutex
	var got []string
	bus.SubscribeAny(func(e eventbus.Event) {
		mu.Lock()
		got = append(got, e.Type())
		mu.Unlock()
	})

	svc := NewService("debug")
	svc.(interface{ WithEventBus(*eventbus.EventBus) }).WithEventBus(bus)

	cfg := contracts.ScanConfig{
		NetworkCIDR: "127.0.0.1/30",
		PortRange:   fmt.Sprintf("%d,%d", port, port+1),
		Timeout:     200 * time.Millisecond,
		Threads:     4,
	}

	if _, err := svc.Scan(context.Background(), cfg, nil); err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Публикация асинхронная (горутина шины) — даём ей время.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(got)
		mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	seen := map[string]bool{}
	for _, ty := range got {
		seen[ty] = true
	}
	if !seen["scan.started"] {
		t.Errorf("expected scan.started event, got %v", got)
	}
	if !seen["scan.completed"] {
		t.Errorf("expected scan.completed event, got %v", got)
	}
}

// TestScannerService_EventBus_CompletedPayload — E6: scan.completed содержит
// корректные счётчики хостов/портов и непустую длительность.
func TestScannerService_EventBus_CompletedPayload(t *testing.T) {
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

	bus := eventbus.NewEventBus()
	defer bus.Close()

	completed := make(chan *eventbus.ScanCompletedEvent, 1)
	bus.Subscribe("scan.completed", func(e eventbus.Event) {
		if ev, ok := e.(*eventbus.ScanCompletedEvent); ok {
			select {
			case completed <- ev:
			default:
			}
		}
	})

	svc := NewService("debug")
	svc.(interface{ WithEventBus(*eventbus.EventBus) }).WithEventBus(bus)

	cfg := contracts.ScanConfig{
		NetworkCIDR: "127.0.0.1/30",
		PortRange:   fmt.Sprintf("%d", port),
		Timeout:     200 * time.Millisecond,
		Threads:     4,
	}
	results, err := svc.Scan(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	select {
	case ev := <-completed:
		if ev.HostCount != len(results) {
			t.Errorf("HostCount=%d, want %d", ev.HostCount, len(results))
		}
		if ev.Duration == "" {
			t.Error("expected non-empty duration")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("scan.completed event not received")
	}
}

// TestScannerService_EventBus_FailedOnCancel — E6: отмена сканирования
// публикует scan.failed с причиной.
func TestScannerService_EventBus_FailedOnCancel(t *testing.T) {
	bus := eventbus.NewEventBus()
	defer bus.Close()

	failed := make(chan eventbus.Event, 1)
	bus.Subscribe("scan.failed", func(e eventbus.Event) {
		select {
		case failed <- e:
		default:
		}
	})

	svc := NewService("debug")
	svc.(interface{ WithEventBus(*eventbus.EventBus) }).WithEventBus(bus)

	// Сканирование большого диапазона портов, чтобы гарантированно прервать
	// его вызовом Stop() до завершения.
	cfg := contracts.ScanConfig{
		NetworkCIDR: "10.255.255.0/28",
		PortRange:   "1-1024",
		Timeout:     5 * time.Second,
		Threads:     2,
	}

	done := make(chan error, 1)
	go func() {
		_, err := svc.Scan(context.Background(), cfg, nil)
		done <- err
	}()

	time.Sleep(150 * time.Millisecond)
	svc.Stop()

	select {
	case err := <-done:
		if err == nil {
			t.Skip("scan completed before Stop(); timing-dependent test")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("scan did not return after Stop()")
	}

	select {
	case ev := <-failed:
		if reason, ok := ev.Payload()["reason"]; !ok || reason == "" {
			t.Errorf("expected non-empty reason payload, got %v", ev.Payload())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("scan.failed event not received on cancellation")
	}
}

// TestScannerService_NoEventBus_NoPanic — E6: без шины сервис работает
// штатно и не паникует при публикации событий.
func TestScannerService_NoEventBus_NoPanic(t *testing.T) {
	svc := NewService("debug")
	cfg := contracts.ScanConfig{
		NetworkCIDR: "127.0.0.1/32",
		PortRange:   "1",
		Timeout:     50 * time.Millisecond,
		Threads:     1,
	}
	if _, err := svc.Scan(context.Background(), cfg, nil); err != nil {
		t.Fatalf("scan without event bus should succeed, got %v", err)
	}
}
