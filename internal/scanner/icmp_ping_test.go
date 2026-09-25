package scanner

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// M6.1: Тесты для ICMP ping probe (Этап 6 / E7 — P3-3)
//
// Живые прогоны (PingICMP к localhost/192.0.2.254) не детерминированы: ICMP
// может быть заблокирован firewall'ом, а код возврата ping различается по ОС.
// Ниже — детерминированные тесты через fake-пингер (SetICMPPinger) и точные
// проверки парсера/валидатора, дающие стабильный результат в CI.
// ============================================================================

// fakeICMPPinger — подменяемый ICMP-пингер (объявлен в scanner_loopback_cov_test.go:
// поля calls/lastTTL используются в тестах ветки ICMP).

// TestPingICMP_ClampsTimeoutToICMPPingTimeout — таймаут ICMP не превышает icmpPingTimeout.
func TestPingICMP_ClampsTimeoutToICMPPingTimeout(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/24", 5*time.Second, "", 10, false)
	fake := &fakeICMPPinger{alive: true}
	ns.SetICMPPinger(fake)

	alive, err := ns.pingICMP("192.0.2.10")
	if err != nil {
		t.Fatalf("pingICMP() error = %v", err)
	}
	if !alive {
		t.Fatal("pingICMP() alive = false, want true")
	}
	if fake.calls != 1 {
		t.Fatalf("PingICMP calls = %d, want 1", fake.calls)
	}
	if fake.lastTTL != icmpPingTimeout {
		t.Errorf("timeout = %v, want %v (clamped)", fake.lastTTL, icmpPingTimeout)
	}
}

// TestPingICMP_KeepsSmallerTimeout — меньший таймаут сканера не увеличивается.
func TestPingICMP_KeepsSmallerTimeout(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/24", 200*time.Millisecond, "", 10, false)
	fake := &fakeICMPPinger{alive: false}
	ns.SetICMPPinger(fake)

	if _, err := ns.pingICMP("192.0.2.10"); err != nil {
		t.Fatalf("pingICMP() error = %v", err)
	}
	if fake.lastTTL != 200*time.Millisecond {
		t.Errorf("timeout = %v, want 200ms", fake.lastTTL)
	}
}

// TestPingICMP_PropagatesError — ошибка пингера доходит до вызывающего кода.
func TestPingICMP_PropagatesError(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/24", icmpPingTimeout, "", 10, false)
	wantErr := errors.New("icmp failed")
	ns.SetICMPPinger(&fakeICMPPinger{err: wantErr})

	if _, err := ns.pingICMP("192.0.2.10"); !errors.Is(err, wantErr) {
		t.Fatalf("pingICMP() error = %v, want %v", err, wantErr)
	}
}

// TestIsHostAlive_ICMPAliveShortCircuits — живой ICMP-хост считается активным без TCP-probe.
func TestIsHostAlive_ICMPAliveShortCircuits(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/24", icmpPingTimeout, "", 10, false)
	ns.SetICMPPingEnabled(true)
	ns.SetICMPPinger(&fakeICMPPinger{alive: true})

	if !ns.isHostAlive("192.0.2.10") {
		t.Fatal("isHostAlive() = false, want true при живом ICMP")
	}
}

// TestIsHostAlive_ICMPDeadFallsBackToTCP — мёртвый ICMP не отменяет TCP-probe fallback.
func TestIsHostAlive_ICMPDeadFallsBackToTCP(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/24", 100*time.Millisecond, "", 10, false)
	ns.SetICMPPingEnabled(true)
	fake := &fakeICMPPinger{alive: false, err: errors.New("unreachable")}
	ns.SetICMPPinger(fake)

	// У 192.0.2.254 заведомо нет открытых commonHostPorts → результат false,
	// но ICMP-пингер должен быть вызван (fallback-ветка не пропускает его).
	if ns.isHostAlive("192.0.2.254") {
		t.Fatal("isHostAlive() = true, want false для TEST-NET адреса")
	}
	if fake.calls == 0 {
		t.Fatal("ICMP pinger not called: ICMP-ветка пропущена")
	}
}

// TestDefaultICMPPinger_Ping_Success — ветка: успешный ICMP ping
func TestDefaultICMPPinger_Ping_Success(t *testing.T) {
	pinger := &DefaultICMPPinger{}

	// Ping на localhost должен работать на большинстве систем
	alive, err := pinger.PingICMP("127.0.0.1", icmpPingTimeout)
	if err != nil {
		t.Logf("ICMP ping to localhost failed (may be blocked): %v", err)
	}
	t.Logf("ICMP ping to 127.0.0.1 alive: %v", alive)
}

// TestDefaultICMPPinger_Ping_Failure — ветка: неудачный ICMP ping
func TestDefaultICMPPinger_Ping_Failure(t *testing.T) {
	pinger := &DefaultICMPPinger{}

	// Ping на несуществующий IP должен вернуть false
	alive, err := pinger.PingICMP("192.0.2.254", icmpPingTimeout)
	if alive {
		t.Error("expected host 192.0.2.254 to be unreachable")
	}
	t.Logf("ICMP ping to 192.0.2.254: alive=%v, err=%v", alive, err)
}

// TestDefaultICMPPinger_Ping_Timeout — ветка: таймаут ICMP ping
func TestDefaultICMPPinger_Ping_Timeout(t *testing.T) {
	pinger := &DefaultICMPPinger{}

	// Короткий таймаут для несуществующего хоста
	alive, err := pinger.PingICMP("192.0.2.254", 100*time.Millisecond)
	if alive {
		t.Error("expected host 192.0.2.254 to be unreachable")
	}
	t.Logf("ICMP ping to 192.0.2.254 with 100ms timeout: alive=%v, err=%v", alive, err)
}

// TestDefaultICMPPinger_Ping_InvalidHost — ветка: невалидный хост
func TestDefaultICMPPinger_Ping_InvalidHost(t *testing.T) {
	pinger := &DefaultICMPPinger{}

	// Невалидный hostname
	alive, err := pinger.PingICMP("invalid-host-name-12345", icmpPingTimeout)
	if alive {
		t.Error("expected invalid hostname to be unreachable")
	}
	t.Logf("ICMP ping to invalid hostname: alive=%v, err=%v", alive, err)
}

// TestPingICMPPool_Empty — ветка: пустой список хостов
func TestPingICMPPool_Empty(t *testing.T) {
	ctx := context.Background()
	hosts := []string{}

	results := PingICMPPool(ctx, hosts, icmpPingTimeout)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestPingICMPPool_SingleHost — ветка: один хост
func TestPingICMPPool_SingleHost(t *testing.T) {
	ctx := context.Background()
	hosts := []string{"192.0.2.254"}

	results := PingICMPPool(ctx, hosts, icmpPingTimeout)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Результат должен быть unreachable
	if results[0].Alive {
		t.Error("expected host 192.0.2.254 to be unreachable")
	}
	t.Logf("ICMP pool ping: alive=%v, latency=%vms, err=%v",
		results[0].Alive, results[0].LatencyMs, results[0].Error)
}

// TestPingICMPPool_MultipleHosts — ветка: несколько хостов
func TestPingICMPPool_MultipleHosts(t *testing.T) {
	ctx := context.Background()
	hosts := []string{"192.0.2.254", "192.0.2.253", "192.0.2.252"}

	results := PingICMPPool(ctx, hosts, icmpPingTimeout)
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Все должны быть unreachable
	for i, result := range results {
		if result.Alive {
			t.Errorf("expected host %s to be unreachable", hosts[i])
		}
		if result.Error == nil && !result.Alive {
			t.Logf("host %s: alive=%v, latency=%vms", hosts[i], result.Alive, result.LatencyMs)
		}
	}
}

// TestPingICMPPool_ContextCancelled — ветка: отмена контекста
func TestPingICMPPool_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем сразу

	hosts := []string{"192.0.2.254", "192.0.2.253"}

	results := PingICMPPool(ctx, hosts, icmpPingTimeout)
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Все результаты должны содержать ошибку отмены
	for i, result := range results {
		if result.Error == nil {
			t.Errorf("expected error for host %s due to context cancellation", hosts[i])
		}
	}
}

// TestNetworkScanner_SetICMPPingEnabled — ветка: включение ICMP ping
func TestNetworkScanner_SetICMPPingEnabled(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/24", icmpPingTimeout, "", 10, false)

	// Проверяем начальное состояние
	if ns.icmpPingEnabled {
		t.Error("expected icmpPingEnabled to be false by default")
	}

	// Включаем ICMP ping
	ns.SetICMPPingEnabled(true)
	if !ns.icmpPingEnabled {
		t.Error("expected icmpPingEnabled to be true after SetICMPPingEnabled(true)")
	}

	// Выключаем
	ns.SetICMPPingEnabled(false)
	if ns.icmpPingEnabled {
		t.Error("expected icmpPingEnabled to be false after SetICMPPingEnabled(false)")
	}
}

// TestNetworkScanner_ICMPPing_Integration — ветка: интеграция с isHostAlive
func TestNetworkScanner_ICMPPing_Integration(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/24", icmpPingTimeout, "", 10, false)
	ns.SetICMPPingEnabled(true)

	// isHostAlive должен использовать ICMP ping
	alive := ns.isHostAlive("192.0.2.254")
	if alive {
		t.Error("expected host 192.0.2.254 to be unreachable")
	}
	t.Logf("isHostAlive for 192.0.2.254: alive=%v", alive)
}

// TestNetworkScanner_ICMPPing_Localhost — ветка: ICMP ping на localhost
func TestNetworkScanner_ICMPPing_Localhost(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/24", icmpPingTimeout, "", 10, false)
	ns.SetICMPPingEnabled(true)

	// localhost должен быть доступен через ICMP (если не заблокирован)
	alive := ns.isHostAlive("127.0.0.1")
	t.Logf("isHostAlive for 127.0.0.1: alive=%v", alive)
}

// TestDefaultICMPPinger_Ping_DifferentOS — ветка: тест на разных ОС
func TestDefaultICMPPinger_Ping_DifferentOS(t *testing.T) {
	// Тест проверяет, что ICMP ping работает на текущей ОС
	pinger := &DefaultICMPPinger{}

	alive, err := pinger.PingICMP("127.0.0.1", icmpPingTimeout)
	t.Logf("OS-specific ICMP ping: alive=%v, err=%v", alive, err)

	// Не делаем жесткой проверки, так как ICMP может быть заблокирован
	// на некоторых системах (например, Windows Firewall)
}

// TestICMPResult_Structure — ветка: структура ICMPResult
func TestICMPResult_Structure(t *testing.T) {
	result := ICMPResult{
		Host:      "192.0.2.1",
		Alive:     true,
		LatencyMs: 15.5,
		Error:     nil,
	}

	if result.Host != "192.0.2.1" {
		t.Errorf("expected Host to be '192.0.2.1', got '%s'", result.Host)
	}
	if !result.Alive {
		t.Error("expected Alive to be true")
	}
	if result.LatencyMs <= 0 {
		t.Errorf("expected LatencyMs > 0, got %v", result.LatencyMs)
	}
	if result.Error != nil {
		t.Errorf("expected Error to be nil, got %v", result.Error)
	}
}

// TestValidateICMPPingHost_Table — таблица валидации хоста (shell-injection guard).
func TestValidateICMPPingHost_Table(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"ipv4", "192.168.1.1", false},
		{"ipv6", "::1", false},
		{"fqdn", "example.com", false},
		{"trimmed", "  example.com  ", false},
		{"empty", "", true},
		{"space", "example com", true},
		{"semicolon", "example.com;rm -rf /", true},
		{"backtick", "example.com`whoami`", true},
		{"pipe", "example.com|cat", true},
		{"leading-dash", "-c100", true},
		{"dollar", "example.com$(id)", true},
		{"no-dot", "localhost", true},
		{"too-long", strings.Repeat("a", 254) + ".com", true},
		{"quote", "example.com\"", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateICMPPingHost(tt.host)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("validateICMPPingHost(%q) err = nil, want error", tt.host)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateICMPPingHost(%q) unexpected error: %v", tt.host, err)
			}
			if got != strings.TrimSpace(tt.host) {
				t.Errorf("validateICMPPingHost(%q) = %q, want trimmed original", tt.host, got)
			}
		})
	}
}

// TestICMPContainsString_Table — точные проверки разбора вывода ping.
func TestICMPContainsString_Table(t *testing.T) {
	linuxOutput := "1 packets transmitted, 1 received, 0% packet loss, time 0ms"
	failOutput := "1 packets transmitted, 0 received, 100% packet loss, time 0ms"
	winOutput := "Reply from 127.0.0.1: bytes=32 time<1ms TTL=128"

	tests := []struct {
		name string
		s    string
		subs []string
		want bool
	}{
		{"linux-ok", linuxOutput, []string{"1 packets transmitted", "1 received"}, true},
		// Реализация — «любая из подстрок»: failOutput содержит
		// "1 packets transmitted" → true. Точная семантика успеха/неуспеха
		// проверяется в PingICMP (нужны ОБЕ подстроки).
		{"linux-fail-any-substring", failOutput, []string{"1 packets transmitted", "1 received"}, true},
		{"windows-lowercase", winOutput, []string{"bytes=32"}, true},
		{"windows-linux-phrase-not-matched", winOutput, []string{"bytes from"}, false},
		{"substring-hit", "1 received", []string{"1 received"}, true},
		{"shorter-than-substr", "abc", []string{"abcdef"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := icmpContainsString(tt.s, tt.subs...); got != tt.want {
				t.Errorf("icmpContainsString(%q, %v) = %v, want %v", tt.s, tt.subs, got, tt.want)
			}
		})
	}
}

// TestPingICMPPool_ContextCancelledShape — форма результата пула при отмене контекста.
func TestPingICMPPool_ContextCancelledShape(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := PingICMPPool(ctx, []string{"192.0.2.254", "192.0.2.253"}, icmpPingTimeout)
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	for _, r := range results {
		if r.Alive {
			t.Errorf("host %s: Alive = true, want false при отменённом контексте", r.Host)
		}
		if r.Error == nil {
			t.Errorf("host %s: Error = nil, want context error", r.Host)
		}
	}
}

// TestPingICMPPool_NilHosts — nil-срез возвращает пустой результат без паники.
func TestPingICMPPool_NilHosts(t *testing.T) {
	if got := PingICMPPool(context.Background(), nil, icmpPingTimeout); len(got) != 0 {
		t.Fatalf("PingICMPPool(nil) len = %d, want 0", len(got))
	}
}
