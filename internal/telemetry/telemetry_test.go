package telemetry_test

import (
	"testing"
	"time"

	"network-scanner/internal/telemetry"
)

func TestNewTelemetry(t *testing.T) {
	cfg := telemetry.DefaultConfig("1.0.0")
	tel := telemetry.NewTelemetry(cfg)
	if tel == nil {
		t.Fatal("Expected non-nil telemetry")
	}
	if !tel.IsEnabled() {
		t.Error("Expected telemetry to be enabled by default")
	}
}

func TestTelemetry_RecordScan(t *testing.T) {
	cfg := telemetry.DefaultConfig("1.0.0")
	tel := telemetry.NewTelemetry(cfg)
	defer tel.Stop()

	tel.RecordScan(10, 5, 30, "quick")

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 1 {
		t.Errorf("Expected queue size 1, got %d", stats["queue_size"])
	}
}

func TestTelemetry_Disabled(t *testing.T) {
	cfg := telemetry.DefaultConfig("1.0.0")
	cfg.Enabled = false
	tel := telemetry.NewTelemetry(cfg)

	tel.RecordScan(10, 5, 30, "quick")

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 0 {
		t.Errorf("Expected queue size 0 when disabled, got %d", stats["queue_size"])
	}
}

func TestTelemetry_SetEnabled(t *testing.T) {
	cfg := telemetry.DefaultConfig("1.0.0")
	cfg.Enabled = true
	tel := telemetry.NewTelemetry(cfg)

	tel.SetEnabled(false)
	if tel.IsEnabled() {
		t.Error("Expected telemetry to be disabled")
	}

	tel.SetEnabled(true)
	if !tel.IsEnabled() {
		t.Error("Expected telemetry to be enabled")
	}
}

func TestTelemetry_MaxQueueSize(t *testing.T) {
	cfg := telemetry.DefaultConfig("1.0.0")
	cfg.MaxQueueSize = 5
	tel := telemetry.NewTelemetry(cfg)
	defer tel.Stop()

	// Пытаемся добавить больше метрик чем размер очереди
	for i := 0; i < 10; i++ {
		tel.RecordScan(10, 5, 30, "quick")
	}

	stats := tel.GetStats()
	if stats["queue_size"].(int) > 5 {
		t.Errorf("Expected queue size <= 5, got %d", stats["queue_size"])
	}
}

func TestTelemetry_AppStart(t *testing.T) {
	cfg := telemetry.DefaultConfig("1.0.0")
	tel := telemetry.NewTelemetry(cfg)
	defer tel.Stop()

	tel.RecordAppStart()

	stats := tel.GetStats()
	if stats["queue_size"].(int) != 1 {
		t.Errorf("Expected queue size 1, got %d", stats["queue_size"])
	}
}

func TestTelemetry_Stop(t *testing.T) {
	cfg := telemetry.DefaultConfig("1.0.0")
	tel := telemetry.NewTelemetry(cfg)

	tel.RecordScan(10, 5, 30, "quick")
	tel.Stop()

	// После Stop отправка не должна паниковать
}

func TestDefaultConfig(t *testing.T) {
	cfg := telemetry.DefaultConfig("2.0.0")

	if cfg.AppVersion != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", cfg.AppVersion)
	}
	if !cfg.Enabled {
		t.Error("Expected enabled by default")
	}
	if cfg.Interval != 1*time.Hour {
		t.Errorf("Expected interval 1h, got %v", cfg.Interval)
	}
}
