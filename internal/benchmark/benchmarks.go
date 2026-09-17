// Package benchmark предоставляет benchmark тесты для CI perf regression.
//
// # Основные бенчмарки
//
// Plugin Benchmarks:
//   - BenchmarkPluginRegistry_Register — скорость регистрации probe
//   - BenchmarkPluginRegistry_ExecuteAll — скорость выполнения всех probe
//   - BenchmarkPluginRegistry_ExecutePhase — скорость выполнения probe по фазе
//
// EventBus Benchmarks:
//   - BenchmarkEventBus_Subscribe — скорость подписки
//   - BenchmarkEventBus_Publish — скорость публикации
//   - BenchmarkEventBus_PublishBatch — скорость batch публикации
//
// ConfigValidation Benchmarks:
//   - BenchmarkConfigValidation_Schema — скорость валидации по схеме
//   - BenchmarkConfigValidation_Specific — скорость специфичных валидаторов
//
// Integration Benchmarks:
//   - BenchmarkFullPipeline — полный пайплайн (config validation + event + plugin)
//
// Memory Benchmarks:
//   - BenchmarkPluginRegistry_MemoryAlloc — аллокации памяти plugin registry
//   - BenchmarkEventBus_MemoryAlloc — аллокации памяти event bus
//   - BenchmarkConfigValidation_MemoryAlloc — аллокации памяти config validation
//
// # Использование в CI
//
// Запуск бенчмарков:
//   go test -bench=. -benchmem ./internal/benchmark/
//
// Сохранение baseline:
//   go test -bench=. -benchmem -run=^$ ./internal/benchmark/ > benchmarks_baseline.txt
//
// Сравнение с baseline:
//   go test -bench=. -benchmem ./internal/benchmark/ | diff benchmarks_baseline.txt -

package benchmark

import (
	"context"
	"testing"
	"time"

	"network-scanner/internal/configvalidation"
	"network-scanner/internal/eventbus"
	"network-scanner/internal/scanner/plugin"
)

// ============================================================================
// Plugin Benchmarks
// ============================================================================

// BenchmarkPluginRegistry_Register — бенчмарк регистрации probe
func BenchmarkPluginRegistry_Register(b *testing.B) {
	registry := plugin.NewRegistry()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		registry.Register(plugin.NewICMPProbe())
		registry.Register(plugin.NewDefaultPortProbe())
		registry.Register(plugin.NewSNMPProbe())
		registry.Register(plugin.NewSSHProbe())
		registry.Register(plugin.NewHTTPProbe())
	}
}

// BenchmarkPluginRegistry_ExecuteAll — бенчмарк выполнения всех probe
func BenchmarkPluginRegistry_ExecuteAll(b *testing.B) {
	registry := plugin.NewRegistry()
	plugin.RegisterDefaultPlugins(registry)
	b.ResetTimer()

	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = registry.ExecuteAll(ctx, "127.0.0.1", 0)
	}
}

// BenchmarkPluginRegistry_ExecutePhase — бенчмарк выполнения probe по фазе
func BenchmarkPluginRegistry_ExecutePhase(b *testing.B) {
	registry := plugin.NewRegistry()
	plugin.RegisterDefaultPlugins(registry)
	b.ResetTimer()

	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = registry.ExecutePhase(ctx, plugin.PhaseHostDiscovery, "127.0.0.1", 0)
	}
}

// BenchmarkPluginRegistry_GetAll — бенчмарк получения всех probe
func BenchmarkPluginRegistry_GetAll(b *testing.B) {
	registry := plugin.NewRegistry()
	plugin.RegisterDefaultPlugins(registry)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = registry.GetAll()
	}
}

// ============================================================================
// EventBus Benchmarks
// ============================================================================

// BenchmarkEventBus_Subscribe — бенчмарк подписки
func BenchmarkEventBus_Subscribe(b *testing.B) {
	bus := eventbus.NewEventBus()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.Subscribe("test.event", func(e eventbus.Event) {})
	}
}

// BenchmarkEventBus_Publish — бенчмарк публикации
func BenchmarkEventBus_Publish(b *testing.B) {
	bus := eventbus.NewEventBus()

	// Подписываем обработчики
	for i := 0; i < 10; i++ {
		bus.Subscribe("test.event", func(e eventbus.Event) {
			_ = e.Type()
		})
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.Publish(eventbus.NewBaseEvent("test.event"))
	}
}

// BenchmarkEventBus_PublishBatch — бенчмарк batch публикации
func BenchmarkEventBus_PublishBatch(b *testing.B) {
	bus := eventbus.NewEventBus()
	bus.Subscribe("batch.event", func(e eventbus.Event) {})

	events := make([]eventbus.Event, 10)
	for i := range events {
		events[i] = eventbus.NewBaseEvent("batch.event")
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.PublishBatch(events)
	}
}

// BenchmarkEventBus_SubscribeAny — бенчмарк подписки на все события
func BenchmarkEventBus_SubscribeAny(b *testing.B) {
	bus := eventbus.NewEventBus()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.SubscribeAny(func(e eventbus.Event) {})
	}
}

// BenchmarkEventBus_ConcurrentPublish — бенчмарк конкурентной публикации
func BenchmarkEventBus_ConcurrentPublish(b *testing.B) {
	bus := eventbus.NewEventBus()
	bus.Subscribe("concurrent.event", func(e eventbus.Event) {})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish(eventbus.NewBaseEvent("concurrent.event"))
		}
	})
}

// ============================================================================
// ConfigValidation Benchmarks
// ============================================================================

// BenchmarkConfigValidation_Schema — бенчмарк валидации по схеме
func BenchmarkConfigValidation_Schema(b *testing.B) {
	schema := configvalidation.Schema{
		"port":     configvalidation.FieldRule{Required: true, Min: 1, Max: 65535},
		"host":     configvalidation.FieldRule{Required: true},
		"cidr":     configvalidation.FieldRule{Required: false},
		"logLevel": configvalidation.FieldRule{Allowed: []string{"debug", "info", "warn", "error"}},
	}

	config := map[string]interface{}{
		"port":     8080,
		"host":     "localhost",
		"cidr":     "192.168.1.0/24",
		"logLevel": "info",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = configvalidation.ValidateConfig(config, schema)
	}
}

// BenchmarkConfigValidation_Specific — бенчмарк специфичных валидаторов
func BenchmarkConfigValidation_Specific(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = configvalidation.ValidatePort(8080)
		_ = configvalidation.ValidateHost("localhost")
		_ = configvalidation.ValidateCIDR("192.168.1.0/24")
		_ = configvalidation.ValidateDuration(10 * time.Second)
		_ = configvalidation.ValidatePortRange("1-1024")
	}
}

// BenchmarkConfigValidation_InvalidConfig — бенчмарк валидации невалидной config
func BenchmarkConfigValidation_InvalidConfig(b *testing.B) {
	schema := configvalidation.Schema{
		"port": configvalidation.FieldRule{Required: true, Min: 1, Max: 65535},
		"host": configvalidation.FieldRule{Required: true},
	}

	config := map[string]interface{}{
		"port": 99999,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errs := configvalidation.ValidateConfig(config, schema)
		if errs.IsValid() {
			b.Fatal("expected validation errors")
		}
	}
}

// BenchmarkConfigValidation_EmptySchema — бенчмарк пустой схемы
func BenchmarkConfigValidation_EmptySchema(b *testing.B) {
	schema := configvalidation.Schema{}
	config := map[string]interface{}{"any": "value"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = configvalidation.ValidateConfig(config, schema)
	}
}

// ============================================================================
// Integration Benchmarks
// ============================================================================

// BenchmarkFullPipeline — бенчмарк полного пайплайна
func BenchmarkFullPipeline(b *testing.B) {
	// Инициализация
	pluginRegistry := plugin.NewRegistry()
	plugin.RegisterDefaultPlugins(pluginRegistry)

	eventBus := eventbus.NewEventBus()
	eventBus.Subscribe("scan.started", func(e eventbus.Event) {})
	eventBus.Subscribe("scan.completed", func(e eventbus.Event) {})

	schema := configvalidation.Schema{
		"network": configvalidation.FieldRule{Required: true},
		"timeout": configvalidation.FieldRule{Required: false},
	}
	config := map[string]interface{}{
		"network": "192.168.1.0/24",
		"timeout": "10s",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 1. Валидация config
		errs := configvalidation.ValidateConfig(config, schema)
		if !errs.IsValid() {
			b.Fatal("validation failed")
		}

		// 2. Публикация события начала сканирования
		eventBus.Publish(eventbus.NewScanStartedEvent("192.168.1.0/24", "10s"))

		// 3. Выполнение probe
		ctx := context.Background()
		_, _ = pluginRegistry.ExecutePhase(ctx, plugin.PhaseHostDiscovery, "127.0.0.1", 0)

		// 4. Публикация события завершения
		eventBus.Publish(eventbus.NewScanCompletedEvent(1, 0, "10ms"))
	}
}

// ============================================================================
// Memory Benchmarks
// ============================================================================

// BenchmarkPluginRegistry_MemoryAlloc — бенчмарк аллокаций памяти
func BenchmarkPluginRegistry_MemoryAlloc(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		registry := plugin.NewRegistry()
		plugin.RegisterDefaultPlugins(registry)
		_ = registry.GetAll()
	}
}

// BenchmarkEventBus_MemoryAlloc — бенчмарк аллокаций памяти
func BenchmarkEventBus_MemoryAlloc(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus := eventbus.NewEventBus()
		bus.Subscribe("test.event", func(e eventbus.Event) {})
		bus.Publish(eventbus.NewBaseEvent("test.event"))
		bus.Close()
	}
}

// BenchmarkConfigValidation_MemoryAlloc — бенчмарк аллокаций памяти
func BenchmarkConfigValidation_MemoryAlloc(b *testing.B) {
	b.ReportAllocs()

	schema := configvalidation.Schema{
		"port": configvalidation.FieldRule{Required: true, Min: 1, Max: 65535},
	}
	config := map[string]interface{}{"port": 8080}

	for i := 0; i < b.N; i++ {
		_ = configvalidation.ValidateConfig(config, schema)
	}
}
