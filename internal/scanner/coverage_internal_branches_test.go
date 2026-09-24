package scanner

import (
	"context"
	"strings"
	"testing"
	"time"

	"network-scanner/internal/contracts"
)

// Внутренние ветки, до которых раньше не доходили тесты: валидация хоста для
// системного ping, ветки адаптивной адаптации, «уже сканируется» в сервисе и
// отмена контекста в потоковых сканерах.

func TestValidateICMPPingHost_Rejects(t *testing.T) {
	cases := []struct {
		name string
		host string
	}{
		{"пустая строка", "   "},
		{"слишком длинный", strings.Repeat("a", 250) + ".example.com"},
		{"shell-метасимволы", "127.0.0.1; rm -rf"},
		{"начинается с флага", "-n"},
		{"недопустимый символ", "host:4444"},
		{"без точки", "localhost"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := validateICMPPingHost(tc.host); err == nil {
				t.Errorf("validateICMPPingHost(%q) должен вернуть ошибку", tc.host)
			}
		})
	}
}

func TestValidateICMPPingHost_Accepts(t *testing.T) {
	for _, host := range []string{" 127.0.0.1 ", "example.com", "host_1.local"} {
		got, err := validateICMPPingHost(host)
		if err != nil {
			t.Errorf("validateICMPPingHost(%q): %v", host, err)
			continue
		}
		if want := strings.TrimSpace(host); got != want {
			t.Errorf("validateICMPPingHost(%q) = %q, want %q", host, got, want)
		}
	}
}

func TestICMPContainsString_Naityvno(t *testing.T) {
	if !icmpContainsString("Pinging 127.0.0.1: 1 received", "1 received") {
		t.Error("icmpContainsString должен находить подстроку")
	}
	if icmpContainsString("no answer", "bytes from") {
		t.Error("icmpContainsString не должен находить отсутствующую подстроку")
	}
	if !icmpContainsSubstring("abc", "bc") || icmpContainsSubstring("abc", "zz") {
		t.Error("icmpContainsSubstring работает неверно")
	}
}

// recorderICMPPinger запоминает таймаут, который передал сканер.
type recorderICMPPinger struct {
	gotTimeout time.Duration
	gotHost    string
}

func (p *recorderICMPPinger) PingICMP(host string, timeout time.Duration) (bool, error) {
	p.gotHost = host
	p.gotTimeout = timeout
	return true, nil
}

func TestNetworkScanner_PingICMP_ClampsTimeout(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 5*time.Second, "1", 1, false)
	defer ns.cancel()

	rec := &recorderICMPPinger{}
	ns.SetICMPPinger(rec)

	alive, err := ns.pingICMP("127.0.0.1")
	if err != nil || !alive {
		t.Fatalf("pingICMP = (%v, %v)", alive, err)
	}
	if rec.gotHost != "127.0.0.1" {
		t.Errorf("хост = %q", rec.gotHost)
	}
	if rec.gotTimeout != icmpPingTimeout {
		t.Errorf("таймаут = %v, want %v", rec.gotTimeout, icmpPingTimeout)
	}

	// Обнулённый таймаут тоже должен приводиться к константе.
	ns.timeout = 0
	if _, err := ns.pingICMP("127.0.0.1"); err != nil {
		t.Fatalf("pingICMP с нулевым таймаутом: %v", err)
	}
	if rec.gotTimeout != icmpPingTimeout {
		t.Errorf("таймаут при timeout=0 = %v, want %v", rec.gotTimeout, icmpPingTimeout)
	}
}

func TestDefaultICMPPinger_PingICMP_Loopback(t *testing.T) {
	// Системный ping по loopback: проверяем только корректность обработки
	// вывода/ошибки — в изолированном окружении хост может и не ответить.
	alive, err := (&DefaultICMPPinger{}).PingICMP("127.0.0.1", 0)
	if alive && err != nil {
		t.Errorf("alive=true вместе с ошибкой %v", err)
	}
	if !alive && err == nil {
		t.Error("alive=false без ошибки")
	}
}

func TestAdaptiveScanner_Adapt_OpenAndClosedRates(t *testing.T) {
	config := AdaptiveConfig{
		MinBudget:      100,
		MaxBudget:      400,
		InitialBudget:  200,
		ErrorThreshold: 0.9,
		AdaptInterval:  time.Nanosecond,
	}
	a := NewAdaptiveScanner(nil, config)

	// Слишком мало probe — адаптация не выполняется.
	a.RecordProbe(true, false)
	a.Adapt()
	if got := a.GetBudget(); got != 200 {
		t.Fatalf("budget после 1 probe = %d, want 200", got)
	}

	// Много открытых портов (openRate > 0.5, probes > 50) — снижение на 10%,
	// но не ниже MinBudget.
	for i := 0; i < 60; i++ {
		a.RecordProbe(true, false)
	}
	a.Adapt()
	if got := a.GetBudget(); got < 100 || got > 200 {
		t.Errorf("budget после серии открытых портов = %d, want в пределах [100,200]", got)
	}

	// Мало открытых портов и probes > 100 — рост budget с ограничением сверху.
	a.SetBudget(400)
	for i := 0; i < 120; i++ {
		a.RecordProbe(false, false)
	}
	a.Adapt()
	if got := a.GetBudget(); got != 400 {
		t.Errorf("budget = %d, want 400 (ограничен MaxBudget)", got)
	}
}

func TestScannerService_Scan_WhenAlreadyRunning(t *testing.T) {
	svc, ok := NewService("error").(*scannerServiceImpl)
	if !ok {
		t.Fatalf("NewService вернул %T", NewService("error"))
	}

	svc.isScanning.Store(true)
	defer svc.isScanning.Store(false)

	if _, err := svc.Scan(context.Background(), contracts.ScanConfig{
		NetworkCIDR: "127.0.0.1/32",
		PortRange:   "1",
	}, nil); err == nil {
		t.Error("Scan при isScanning=true должен отказать")
	}
}

// seedResults наполняет результаты сканера, не запуская реальное сканирование.
func seedResults(ns *NetworkScanner, n int) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	for i := 0; i < n; i++ {
		ns.results = append(ns.results, Result{IP: "127.0.0.1"})
	}
}

// Потоковые сканеры должны обрывать отправку результатов, если контекст
// отменён, а канал событий никто не читает.
func TestIncrementalScanner_StopsSendingWhenContextCanceled(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*IncrementalScanner, context.Context) (<-chan ScanEvent, <-chan error)
	}{
		{
			name: "ScanWithEvents",
			run: func(s *IncrementalScanner, ctx context.Context) (<-chan ScanEvent, <-chan error) {
				return s.ScanWithEvents(ctx, contracts.ScanConfig{NetworkCIDR: "127.0.0.1/32"})
			},
		},
		{
			name: "ScanWithEventsAndConfig",
			run: func(s *IncrementalScanner, ctx context.Context) (<-chan ScanEvent, <-chan error) {
				return s.ScanWithEventsAndConfig(ctx, contracts.ScanConfig{NetworkCIDR: "127.0.0.1/32"})
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ns := NewNetworkScanner("127.0.0.1/32", 10*time.Millisecond, "1", 1, false)
			defer ns.cancel()
			// Буфер канала событий — 100, поэтому 150 результатов гарантированно
			// переполняют его, и выбор в select падает на отменённый контекст.
			seedResults(ns, 150)

			ctx, cancel := context.WithCancel(context.Background())
			events, errCh := tc.run(NewIncrementalScanner(ns), ctx)
			cancel()

			// Канал закрывается только когда горутина завершилась — это и есть
			// ожидаемая ветка «контекст отменён».
			for range events {
			}
			if _, ok := <-errCh; ok {
				t.Error("канал ошибок должен быть закрыт")
			}
		})
	}
}

func TestPrintEventHandler_OpenPortsAndProgressNewline(t *testing.T) {
	res := &Result{
		IP: "127.0.0.1",
		Ports: []PortInfo{
			{Port: 443, Protocol: "tcp", Service: "https", State: "closed"},
			{Port: 80, Protocol: "tcp", Service: "http", State: "open"},
		},
	}

	handlers := map[string]func(ScanEvent) error{
		"verbose": PrintEventHandler(true),
		"quiet":   PrintEventHandler(false),
	}

	for name, handler := range handlers {
		if err := handler(ScanEvent{Type: "progress", Stage: "portscan", Current: 3, Total: 3}); err != nil {
			t.Errorf("%s: progress: %v", name, err)
		}
		if err := handler(ScanEvent{Type: "host", Result: res}); err != nil {
			t.Errorf("%s: host: %v", name, err)
		}
	}
}
