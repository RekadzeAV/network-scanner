package benchmark

import (
	"testing"
	"time"

	"network-scanner/internal/configvalidation"
	"network-scanner/internal/eventbus"
	"network-scanner/internal/scanner/plugin"
)

// TestBenchmarkPackageSanity — проверка что весь пайплайн работает корректно.
// Гарантирует что бенчмарки измеряют рабочий код, а не ошибку.
func TestBenchmarkPackageSanity(t *testing.T) {
	// 1. Config validation работает
	schema := configvalidation.Schema{
		"network": configvalidation.FieldRule{Required: true},
	}
	if errs := configvalidation.ValidateConfig(map[string]interface{}{"network": "192.168.1.0/24"}, schema); !errs.IsValid() {
		t.Fatalf("config validation failed: %v", errs)
	}

	// 2. Event bus публикует события
	bus := eventbus.NewEventBus()
	defer bus.Close()
	done := make(chan struct{}, 1)
	bus.Subscribe("scan.started", func(e eventbus.Event) {
		select {
		case done <- struct{}{}:
		default:
		}
	})
	bus.Publish(eventbus.NewScanStartedEvent("192.168.1.0/24", "10s"))
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("event not received")
	}

	// 3. Plugin registry зарегистрирован
	registry := plugin.NewRegistry()
	plugin.RegisterDefaultPlugins(registry)
	if registry.Count() != 5 {
		t.Fatalf("expected 5 plugins, got %d", registry.Count())
	}
}
