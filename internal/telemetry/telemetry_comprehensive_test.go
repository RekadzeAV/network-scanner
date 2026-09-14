package telemetry

import (
	"sync"
	"testing"
	"time"
)

// ============================================================================
// TestTelemetry_MultipleRecords — покрытие веток с несколькими метриками
// ============================================================================

func TestTelemetry_MultipleScanRecords(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	// Добавляем несколько метрик
	for i := 0; i < 10; i++ {
		tel.RecordScan(i, i*2, 10+i, "test")
	}

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 10 {
		t.Errorf("Expected queue size 10, got %d", stats["queue_size"])
	}
}

func TestTelemetry_MixedRecords(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	tel.RecordScan(10, 5, 30, "quick")
	tel.RecordAppStart()
	tel.RecordScan(20, 10, 60, "full")

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 3 {
		t.Errorf("Expected queue size 3, got %d", stats["queue_size"])
	}
}

// ============================================================================
// TestTelemetry_EdgeCases — граничные значения
// ============================================================================

func TestTelemetry_ZeroValues(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	tel.RecordScan(0, 0, 0, "")

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 1 {
		t.Errorf("Expected queue size 1, got %d", stats["queue_size"])
	}
}

func TestTelemetry_NegativeValues(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	tel.RecordScan(-1, -5, -30, "negative")

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 1 {
		t.Errorf("Expected queue size 1, got %d", stats["queue_size"])
	}
}

func TestTelemetry_LargeValues(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	tel.RecordScan(100000, 50000, 3600, "large")

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 1 {
		t.Errorf("Expected queue size 1, got %d", stats["queue_size"])
	}
}

// ============================================================================
// TestTelemetry_QueueFull — очередь заполнена
// ============================================================================

func TestTelemetry_QueueFull_DropsMetrics(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	cfg.MaxQueueSize = 3
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	// Заполняем очередь
	for i := 0; i < 3; i++ {
		tel.RecordScan(i, i, 10, "test")
	}

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 3 {
		t.Errorf("Expected queue size 3, got %d", stats["queue_size"])
	}

	// Пытаемся добавить ещё — должно быть проигнорировано
	tel.RecordScan(100, 100, 100, "overflow")

	stats = tel.GetStats()
	if stats["queue_size"].(int) != 3 {
		t.Errorf("Expected queue size still 3, got %d", stats["queue_size"])
	}
}

// ============================================================================
// TestTelemetry_EmptyConfig — пустая конфигурация
// ============================================================================

func TestTelemetry_EmptyAppVersion(t *testing.T) {
	cfg := DefaultConfig("")
	if cfg.AppVersion != "" {
		t.Errorf("Expected empty app version, got %q", cfg.AppVersion)
	}
}

func TestTelemetry_CustomEndpoint(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	cfg.Endpoint = "https://custom.example.com/collect"
	tel := NewTelemetry(cfg)

	stats := tel.GetStats()
	if stats["endpoint"] != "https://custom.example.com/collect" {
		t.Errorf("Expected custom endpoint, got %v", stats["endpoint"])
	}
}

func TestTelemetry_CustomInterval(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	cfg.Interval = 5 * time.Minute
	tel := NewTelemetry(cfg)

	stats := tel.GetStats()
	if stats["endpoint"] == "" {
		t.Error("Expected non-empty endpoint")
	}
	_ = tel // cfg.Interval не доступен извне, но должен устанавливаться
}

func TestTelemetry_CustomMaxQueueSize(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	cfg.MaxQueueSize = 100
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	for i := 0; i < 100; i++ {
		tel.RecordScan(i, i, 10, "test")
	}

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 100 {
		t.Errorf("Expected queue size 100, got %d", stats["queue_size"])
	}
}

// ============================================================================
// TestTelemetry_Concurrent — потокобезопасность
// ============================================================================

func TestTelemetry_ConcurrentRecords(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tel.RecordScan(id, id*2, 10, "concurrent")
		}(i)
	}
	wg.Wait()

	stats := tel.GetStats()
	queueSize := stats["queue_size"].(int)
	if queueSize != 10 {
		t.Errorf("Expected queue size 10, got %d", queueSize)
	}
}

func TestTelemetry_ConcurrentEnableDisable(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			tel.RecordScan(i, i, 10, "test")
		}()
		go func() {
			defer wg.Done()
			tel.SetEnabled(i%2 == 0)
		}()
	}
	wg.Wait()

	// Должно быть без паники
	stats := tel.GetStats()
	if stats == nil {
		t.Error("Expected non-nil stats")
	}
}

// ============================================================================
// TestTelemetry_SetEnabled_Twice — двойное переключение
// ============================================================================

func TestTelemetry_SetEnabled_Twice(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)

	tel.SetEnabled(false)
	if tel.IsEnabled() {
		t.Error("Expected disabled")
	}

	tel.SetEnabled(false) // Повторное выключение
	if tel.IsEnabled() {
		t.Error("Expected still disabled")
	}

	tel.SetEnabled(true)
	if !tel.IsEnabled() {
		t.Error("Expected enabled")
	}

	tel.SetEnabled(true) // Повторное включение
	if !tel.IsEnabled() {
		t.Error("Expected still enabled")
	}
}

// ============================================================================
// TestTelemetry_Stop_ThenRecord — запись после остановки
// ============================================================================

func TestTelemetry_Stop_ThenRecord(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)

	tel.RecordScan(10, 5, 30, "before")
	tel.Stop()
	tel.RecordScan(20, 10, 60, "after")

	stats := tel.GetStats()
	// После Stop метрики не добавляются
	if stats["queue_size"].(int) != 1 {
		t.Errorf("Expected queue size 1 (only before Stop), got %d", stats["queue_size"])
	}
}

// ============================================================================
// TestTelemetry_Start_Stop — запуск и остановка
// ============================================================================

func TestTelemetry_Start_Stop(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)

	// Start должен работать без паники
	tel.Start()

	// Stop должен остановить
	tel.Stop()

	// Повторный Stop не должен паниковать
	tel.Stop()
}

func TestTelemetry_Start_Disabled(t *testing.T) {
	cfg := DefaultConfig("1.0.0")
	cfg.Enabled = false
	tel := NewTelemetry(cfg)

	// Start на отключённой телеметрии должен быть no-op
	tel.Start()

	// Stop должен работать
	tel.Stop()
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkRecordScan(b *testing.B) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tel.RecordScan(10, 5, 30, "benchmark")
	}
}

func BenchmarkRecordAppStart(b *testing.B) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tel.RecordAppStart()
	}
}

func BenchmarkGetStats(b *testing.B) {
	cfg := DefaultConfig("1.0.0")
	tel := NewTelemetry(cfg)
	defer tel.Stop()

	// Добавляем метрики
	for i := 0; i < 100; i++ {
		tel.RecordScan(i, i, 10, "test")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tel.GetStats()
	}
}
