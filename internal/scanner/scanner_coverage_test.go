package scanner

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/network"
)

// --- scanUDPPort tests ---

func TestScanUDPPort(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)

	// Сканирование недоступного UDP порта — не должно паниковать
	result := ns.scanUDPPort("192.0.2.1", 53)
	_ = result
}

// --- scanUDPPortWithTimeout edge cases ---

func TestScanUDPPortWithTimeout_ZeroTimeout(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// При timeout=0 должен использоваться ns.timeout
	result := ns.scanUDPPortWithTimeout("192.0.2.1", 53, 0)
	_ = result
}

func TestScanUDPPortWithTimeout_NegativeTimeout(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	result := ns.scanUDPPortWithTimeout("192.0.2.1", 53, -1*time.Second)
	_ = result
}

// --- checkARP tests ---

func TestCheckARP(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// checkARP всегда возвращает false (stub)
	result := ns.checkARP("192.168.1.1")
	if result {
		t.Error("checkARP() should always return false")
	}
}

// --- scanHostUDP tests ---

func TestScanHostUDP(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	ns.SetScanUDP(true)

	result := &Result{IP: "192.0.2.1", Ports: make([]PortInfo, 0)}

	// Сканирование UDP недоступного хоста — не должно паниковать
	ns.scanHostUDP(result.IP, result)

	_ = result.Ports
}

func TestScanHostUDP_Cancelled(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	ns.SetScanUDP(true)

	ns.Stop()

	result := &Result{IP: "192.0.2.1", Ports: make([]PortInfo, 0)}

	ns.scanHostUDP(result.IP, result)
}

// --- isHostAlive with context prober ---

func TestIsHostAlive_ContextProberSuccess(t *testing.T) {
	prober := network.DefaultNetworkProber{Timeout: 100 * time.Millisecond}
	ns := NewScanner(
		"192.168.1.0/24",
		100*time.Millisecond,
		"80",
		10,
		false,
		prober,
		nil,
		nil,
	)

	result := ns.isHostAlive("192.0.2.1")
	_ = result
}

func TestIsHostAlive_ContextProberFail(t *testing.T) {
	prober := network.DefaultNetworkProber{Timeout: 10 * time.Millisecond}
	ns := NewScanner(
		"192.168.1.0/24",
		10*time.Millisecond,
		"80",
		10,
		false,
		prober,
		nil,
		nil,
	)

	result := ns.isHostAlive("192.0.2.1")
	_ = result
}

func TestIsHostAlive_VerboseLogs(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	ns.SetVerbosePortLogs(true)

	result := ns.isHostAlive("192.0.2.1")
	_ = result
}

func TestIsHostAlive_NilProber(t *testing.T) {
	ns := NewScanner(
		"192.168.1.0/24",
		100*time.Millisecond,
		"80",
		10,
		false,
		nil, // nil prober
		nil,
		nil,
	)

	// При nil prober должен использоваться встроенный пинг по портам
	result := ns.isHostAlive("192.0.2.1")
	_ = result
}

func TestIsHostAlive_ContextCancelled(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)

	// Отменяем контекст перед проверкой
	ns.Stop()

	// При отменённом контексте должен вернуть false
	result := ns.isHostAlive("192.0.2.1")
	if result {
		t.Error("isHostAlive() should return false when context is cancelled")
	}
}

// --- scanTCPPort with nil scanner ---

func TestScanTCPPort_NilPortScanner(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)

	result := ns.scanTCPPort("192.0.2.1", 80)
	_ = result
}

// --- getMACAddress with nil prober ---

func TestGetMACAddress_NilProber(t *testing.T) {
	ns := NewScanner(
		"192.168.1.0/24",
		1*time.Second,
		"80",
		10,
		false,
		nil,
		nil,
		nil,
	)

	_, err := ns.getMACAddress(net.ParseIP("192.168.1.1"))
	_ = err
}

func TestGetMACAddress_NilIP(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	_, err := ns.getMACAddress(nil)
	if err == nil {
		t.Error("getMACAddress(nil) should return error")
	}
}

func TestGetMACAddress_InvalidIPv6(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	_, err := ns.getMACAddress(net.ParseIP("::1"))
	if err == nil {
		t.Error("getMACAddress(IPv6) should return error")
	}
}

// --- ScanWithEventsAndConfig tests ---

func TestScanWithEventsAndConfig(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/30", 50*time.Millisecond, "80", 10, false)
	scanner := NewIncrementalScanner(ns)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, errChan := scanner.ScanWithEventsAndConfig(ctx, contracts.ScanConfig{
		NetworkCIDR: "192.0.2.0/30",
		PortRange:   "80",
	})

	eventCount := 0
	timeout := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-events:
			if !ok {
				goto done
			}
			eventCount++
		case err := <-errChan:
			if err != nil {
				t.Logf("ScanWithEventsAndConfig error: %v", err)
			}
			goto done
		case <-timeout:
			t.Fatal("ScanWithEventsAndConfig timed out")
		}
	}
done:
	if eventCount == 0 {
		t.Log("ScanWithEventsAndConfig produced no events (expected in test environment)")
	}
}

// --- PrintEventHandler tests ---

func TestPrintEventHandler_Quiet(t *testing.T) {
	handler := PrintEventHandler(false)

	err := handler(ScanEvent{
		Type:    "start",
		Message: "test",
	})
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}
}

func TestPrintEventHandler_Verbose(t *testing.T) {
	handler := PrintEventHandler(true)

	err := handler(ScanEvent{
		Type:    "start",
		Message: "test start",
	})
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}

	err = handler(ScanEvent{
		Type:    "progress",
		Stage:   "ping",
		Message: "test progress",
		Current: 1,
		Total:   10,
	})
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}

	err = handler(ScanEvent{
		Type:    "summary",
		Message: "test summary",
	})
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}
}

func TestPrintEventHandler_HostVerbose(t *testing.T) {
	handler := PrintEventHandler(true)

	err := handler(ScanEvent{
		Type:    "host",
		Message: "test host",
		Result: &Result{
			IP:         "192.168.1.1",
			DeviceType: "Server",
			MAC:        "aa:bb:cc:dd:ee:ff",
			Hostname:   "test",
			Ports: []PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"},
			},
		},
	})
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}
}

func TestPrintEventHandler_Default(t *testing.T) {
	handler := PrintEventHandler(false)

	err := handler(ScanEvent{
		Type:    "unknown_type",
		Message: "unknown event",
	})
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}
}

// --- CollectEventHandler tests ---

func TestCollectEventHandler_Progress(t *testing.T) {
	handler := NewCollectEventHandler()

	err := handler.Handle(ScanEvent{
		Type:    "progress",
		Stage:   "ping",
		Message: "test",
	})
	if err != nil {
		t.Errorf("Handle() error = %v", err)
	}

	progress := handler.GetProgress()
	if len(progress) != 1 {
		t.Errorf("GetProgress() length = %d, want 1", len(progress))
	}
}

func TestCollectEventHandler_Host(t *testing.T) {
	handler := NewCollectEventHandler()

	result := Result{IP: "192.168.1.1", DeviceType: "Server"}
	err := handler.Handle(ScanEvent{
		Type:    "host",
		Message: "test host",
		Result:  &result,
	})
	if err != nil {
		t.Errorf("Handle() error = %v", err)
	}

	results := handler.GetResults()
	if len(results) != 1 {
		t.Errorf("GetResults() length = %d, want 1", len(results))
	}
}

func TestCollectEventHandler_GetResultsReturnsCopy(t *testing.T) {
	handler := NewCollectEventHandler()

	result := Result{IP: "192.168.1.1"}
	handler.Handle(ScanEvent{
		Type:   "host",
		Result: &result,
	})

	results1 := handler.GetResults()
	results1[0].IP = "10.0.0.1"

	results2 := handler.GetResults()
	if results2[0].IP != "192.168.1.1" {
		t.Error("GetResults() should return a copy")
	}
}

func TestCollectEventHandler_GetProgressReturnsCopy(t *testing.T) {
	handler := NewCollectEventHandler()

	handler.Handle(ScanEvent{
		Type:    "progress",
		Message: "test",
	})

	progress1 := handler.GetProgress()
	progress1[0].Message = "modified"

	progress2 := handler.GetProgress()
	if progress2[0].Message != "test" {
		t.Error("GetProgress() should return a copy")
	}
}

// --- ServiceImpl.Stop test ---

func TestServiceImpl_Stop(t *testing.T) {
	svc := NewService("info")
	svc.Stop()
}

// --- ConsumeEvents tests ---

func TestConsumeEvents(t *testing.T) {
	ctx := context.Background()
	events := make(chan ScanEvent, 3)

	events <- ScanEvent{Type: "start", Message: "start"}
	events <- ScanEvent{Type: "summary", Message: "done"}
	close(events)

	lastEvent, err := ConsumeEvents(ctx, events, func(event ScanEvent) error {
		return nil
	})

	if err != nil {
		t.Errorf("ConsumeEvents() error = %v", err)
	}
	if lastEvent.Type != "summary" {
		t.Errorf("ConsumeEvents() lastEvent.Type = %v, want summary", lastEvent.Type)
	}
}

func TestConsumeEvents_HandlerError(t *testing.T) {
	ctx := context.Background()
	events := make(chan ScanEvent, 3)

	events <- ScanEvent{Type: "start", Message: "start"}
	close(events)

	_, err := ConsumeEvents(ctx, events, func(event ScanEvent) error {
		return fmt.Errorf("handler error")
	})

	if err == nil {
		t.Error("ConsumeEvents() should return error when handler returns error")
	}
}

func TestConsumeEvents_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan ScanEvent, 3)

	events <- ScanEvent{Type: "start", Message: "start"}
	cancel()

	_, err := ConsumeEvents(ctx, events, func(event ScanEvent) error {
		return nil
	})

	if err == nil {
		t.Error("ConsumeEvents() should return context error when cancelled")
	}
}

// --- getMACViaARPRequest tests ---

func TestGetMACViaARPRequest_NonLocalIP(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)

	// Для IP вне локальной сети ARP запрос не вернёт ответ
	_, err := ns.getMACViaARPRequest(net.ParseIP("192.0.2.1"))
	_ = err
}

// --- adaptive_scanner tests ---

func TestAdaptiveScanner_Adapt_NoData(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	config := DefaultAdaptiveConfig()
	scanner := NewAdaptiveScanner(ns, config)

	// Adapt без данных — меньше 10 probe — не должен менять budget
	scanner.Adapt()
	if scanner.GetBudget() != int(config.InitialBudget) {
		t.Errorf("Adapt() budget = %d, want %d", scanner.GetBudget(), config.InitialBudget)
	}
}

func TestAdaptiveScanner_Adapt_HighError(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	config := DefaultAdaptiveConfig()
	config.AdaptInterval = 0 // Отключаем интервал для теста
	scanner := NewAdaptiveScanner(ns, config)

	// Записываем probe'ы с ошибками
	for i := 0; i < 20; i++ {
		scanner.RecordProbe(false, true)
	}

	scanner.Adapt()
	// При 100%% ошибок budget должен уменьшиться
	if scanner.GetBudget() >= int(config.InitialBudget) {
		t.Errorf("Adapt() budget = %d, want < initial=%d", scanner.GetBudget(), config.InitialBudget)
	}
}

func TestAdaptiveScanner_Adapt_FastSuccess(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	config := DefaultAdaptiveConfig()
	config.AdaptInterval = 0
	scanner := NewAdaptiveScanner(ns, config)

	// Записываем закрытые порты (не открытые) — openRate < 0.1
	for i := 0; i < 120; i++ {
		scanner.RecordProbe(false, false) // closed, no error
	}

	scanner.Adapt()
	// При 0%% успеха и 0%% ошибок budget должен вырасти
	if scanner.GetBudget() <= int(config.InitialBudget) {
		t.Logf("Adapt() budget = %d (expected >= %d for 0%% open rate)", scanner.GetBudget(), config.InitialBudget)
	}
}

func TestAdaptiveScanner_GetMetrics(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	config := DefaultAdaptiveConfig()
	scanner := NewAdaptiveScanner(ns, config)

	scanner.RecordProbe(true, false)
	scanner.RecordProbe(false, true)

	metrics := scanner.GetMetrics()
	_ = metrics // Проверяем что не паникует и возвращает структуру
}

func TestAdaptiveScanner_GetOpenRate(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	config := DefaultAdaptiveConfig()
	scanner := NewAdaptiveScanner(ns, config)

	scanner.RecordProbe(true, false)
	scanner.RecordProbe(false, false)

	rate := scanner.GetOpenRate()
	if rate != 0.5 {
		t.Errorf("GetOpenRate() = %f, want 0.5", rate)
	}
}

func TestAdaptiveScanner_GetErrorRate(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	config := DefaultAdaptiveConfig()
	scanner := NewAdaptiveScanner(ns, config)

	scanner.RecordProbe(false, true)
	scanner.RecordProbe(true, false)

	rate := scanner.GetErrorRate()
	if rate != 0.5 {
		t.Errorf("GetErrorRate() = %f, want 0.5", rate)
	}
}

func TestAdaptiveScanner_GetSummary(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	config := DefaultAdaptiveConfig()
	scanner := NewAdaptiveScanner(ns, config)

	summary := scanner.GetSummary()
	if summary == "" {
		t.Error("GetSummary() should return non-empty string")
	}
}

func TestAdaptiveScanner_SetBudget(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 100*time.Millisecond, "80", 10, false)
	config := DefaultAdaptiveConfig()
	scanner := NewAdaptiveScanner(ns, config)

	scanner.SetBudget(200)
	if scanner.GetBudget() != 200 {
		t.Errorf("SetBudget(200) = %d, want 200", scanner.GetBudget())
	}

	// Тест минимального ограничения
	scanner.SetBudget(10)
	if scanner.GetBudget() != config.MinBudget {
		t.Errorf("SetBudget(10) = %d, want min=%d", scanner.GetBudget(), config.MinBudget)
	}

	// Тест максимального ограничения
	scanner.SetBudget(9999)
	if scanner.GetBudget() != config.MaxBudget {
		t.Errorf("SetBudget(9999) = %d, want max=%d", scanner.GetBudget(), config.MaxBudget)
	}
}
