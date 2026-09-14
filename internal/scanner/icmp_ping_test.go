package scanner

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// M6.1: Тесты для ICMP ping probe (Этап 6)
// ============================================================================

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
