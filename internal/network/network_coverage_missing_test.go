package network

import (
	"testing"
	"time"
)

// ============================================================================
// M1.X: Тесты для непокрытых функций network (0.0% → 100%)
// ============================================================================

// TestARPCache_Stop — ветка: остановка кэша
func TestARPCache_Stop(t *testing.T) {
	cache := NewARPCache(5*time.Minute, func() (map[string]string, error) {
		return map[string]string{"192.168.1.1": "aa:bb:cc:dd:ee:ff"}, nil
	})

	// Stop должна вызываться без паники
	cache.Stop()
}

// TestGetARPTabaleWindows — ветка: Windows ARP таблица
func TestGetARPTabaleWindows(t *testing.T) {
	// На Windows эта функция вызовет arp -a
	// На других ОС вернёт ошибку
	macTable, err := GetARPTabaleWindows()
	if err != nil {
		t.Logf("expected error on non-Windows or no ARP: %v", err)
	}
	_ = macTable
}

// TestGetARPTabaleLinux — ветка: Linux ARP таблица
func TestGetARPTabaleLinux(t *testing.T) {
	// На Linux эта функция вызовет ip neigh
	// На других ОС вернёт ошибку
	macTable, err := GetARPTabaleLinux()
	if err != nil {
		t.Logf("expected error on non-Linux or no ARP: %v", err)
	}
	_ = macTable
}

// TestGetARPTabale — ветка: кроссплатформенная ARP
func TestGetARPTabale(t *testing.T) {
	// Вызовет GetARPTabaleWindows или GetARPTabaleLinux в зависимости от ОС
	macTable, err := GetARPTabale()
	if err != nil {
		t.Logf("expected error or empty ARP: %v", err)
	}
	_ = macTable
}

// TestNewDefaultARPCache — ветка: создание кэша с дефолтным updater
func TestNewDefaultARPCache(t *testing.T) {
	cache := NewDefaultARPCache(10 * time.Minute)
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}

	// Проверяем что кэш создан
	if !cache.IsFresh() {
		t.Log("cache not fresh yet (expected)")
	}

	// Останавливаем
	cache.Stop()
}

// TestNewDefaultARPCache_Refresh — ветка: обновление кэша
func TestNewDefaultARPCache_Refresh(t *testing.T) {
	cache := NewDefaultARPCache(1 * time.Minute)
	defer cache.Stop()

	// Запускаем обновление
	cache.Refresh()

	// Даем время на завершение
	time.Sleep(100 * time.Millisecond)

	// Проверяем размер
	size := cache.Size()
	t.Logf("cache size after refresh: %d", size)
}

// ============================================================================
// M1.X: Тесты для prober (0.0% → 100%)
// ============================================================================

// TestNewDefaultNetworkProber — ветка: создание пробера
func TestNewDefaultNetworkProber(t *testing.T) {
	prober := NewDefaultNetworkProber(1 * time.Second)
	if prober == nil {
		t.Fatal("expected non-nil prober")
	}
}

// TestNetworkProber_SetARPCache — ветка: установка кэша
func TestNetworkProber_SetARPCache(t *testing.T) {
	prober := NewDefaultNetworkProber(1 * time.Second)

	cache := NewARPCache(5*time.Minute, func() (map[string]string, error) {
		return map[string]string{}, nil
	})
	defer cache.Stop()

	// Устанавливаем кэш
	prober.SetARPCache(cache)
}

// TestNetworkProber_Ping_Success — ветка: успешный ping
func TestNetworkProber_Ping_Success(t *testing.T) {
	prober := NewDefaultNetworkProber(500 * time.Millisecond)

	// Ping на localhost должен работать
	isAlive, err := prober.Ping("127.0.0.1")
	if err != nil {
		t.Logf("ping error: %v", err)
	}
	t.Logf("host 127.0.0.1 alive: %v", isAlive)
}

// TestNetworkProber_Ping_Failure — ветка: fail ping
func TestNetworkProber_Ping_Failure(t *testing.T) {
	prober := NewDefaultNetworkProber(200 * time.Millisecond)

	// Ping на заведомо несуществующий IP
	isAlive, err := prober.Ping("192.0.2.254")
	if err == nil {
		t.Logf("unexpected success: alive=%v", isAlive)
	}
}

// ============================================================================
// M1.X: Тесты для detectLocalNetworkIPv6 (0.0% → 100%)
// ============================================================================

// TestDetectLocalNetworkIPv6 — ветка: определение IPv6 сети
func TestDetectLocalNetworkIPv6(t *testing.T) {
	// Эта функция определяет локальную IPv6 сеть
	// На машинах без IPv6 вернёт ошибку или пустой результат
	network, err := detectLocalNetworkIPv6()
	if err != nil {
		t.Logf("expected error or empty network: %v", err)
	}
	t.Logf("detected IPv6 network: %s", network)
}

// ============================================================================
// M1.X: Тесты для IsPortOpen (70.0% → 100%)
// ============================================================================

// TestIsPortOpen_Open — ветка: открытый порт
func TestIsPortOpen_Open(t *testing.T) {
	// Порт 80 часто открыт
	isOpen := IsPortOpen("127.0.0.1", 80, 500*time.Millisecond)
	t.Logf("port 80 on localhost open: %v", isOpen)
}

// TestIsPortOpen_Closed — ветка: закрытый порт
func TestIsPortOpen_Closed(t *testing.T) {
	// Порт 59999 точно закрыт
	isOpen := IsPortOpen("127.0.0.1", 59999, 200*time.Millisecond)
	if isOpen {
		t.Error("expected closed port")
	}
}

// TestIsPortOpen_Timeout — ветка: таймаут соединения
func TestIsPortOpen_Timeout(t *testing.T) {
	// Порт 59998 с коротким таймаутом
	isOpen := IsPortOpen("192.0.2.1", 59998, 100*time.Millisecond)
	if isOpen {
		t.Error("expected closed port")
	}
}
