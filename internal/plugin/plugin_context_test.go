package plugin

import (
	"fmt"
	"testing"
	"time"
)

// ============================================================================
// PluginContext — полное покрытие
// ============================================================================

func TestNewPluginContext_Basic(t *testing.T) {
	pc := NewPluginContext("scan-123")
	if pc.ScanID != "scan-123" {
		t.Errorf("expected scan-123, got %q", pc.ScanID)
	}
	if pc.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}
	if pc.CancelContext == nil {
		t.Fatal("CancelContext should not be nil")
	}
	if pc.CancelFunc == nil {
		t.Fatal("CancelFunc should not be nil")
	}
}

func TestPluginContext_Cancel(t *testing.T) {
	pc := NewPluginContext("scan-456")
	pc.CancelFunc()

	select {
	case <-pc.CancelContext.Done():
		// OK — контекст отменён
	default:
		t.Error("expected context to be cancelled")
	}
}

func TestPluginContext_Duration(t *testing.T) {
	pc := NewPluginContext("scan-789")
	// Проверим что StartTime установлена
	if pc.StartTime.IsZero() {
		t.Fatal("StartTime should be set")
	}
	// Проверим что время прошло
	elapsed := time.Since(pc.StartTime)
	if elapsed < 0 || elapsed > time.Second {
		t.Errorf("elapsed time should be small, got %v", elapsed)
	}
}

// ============================================================================
// DefaultEventBus — полное покрытие
// ============================================================================

func TestDefaultEventBus_SubscribeAndPublish(t *testing.T) {
	bus := NewDefaultEventBus()
	var received []interface{}
	handler := func(data interface{}) {
		received = append(received, data)
	}

	id := bus.Subscribe("test_event", handler)
	bus.Publish("test_event", "data1")
	bus.Publish("test_event", "data2")

	if len(received) != 2 {
		t.Errorf("expected 2 events, got %d", len(received))
	}
	if received[0] != "data1" || received[1] != "data2" {
		t.Errorf("unexpected data: %v", received)
	}

	// Отписка
	bus.Unsubscribe("test_event", id)
	bus.Publish("test_event", "data3")

	if len(received) != 2 {
		t.Error("should not receive after unsubscribe")
	}
}

func TestDefaultEventBus_MultipleHandlers(t *testing.T) {
	bus := NewDefaultEventBus()
	var count1, count2 int

	handler1 := func(data interface{}) { count1++ }
	handler2 := func(data interface{}) { count2++ }

	id1 := bus.Subscribe("event", handler1)
	id2 := bus.Subscribe("event", handler2)
	_ = id2

	bus.Publish("event", nil)
	bus.Publish("event", nil)

	if count1 != 2 {
		t.Errorf("handler1 called %d times, want 2", count1)
	}
	if count2 != 2 {
		t.Errorf("handler2 called %d times, want 2", count2)
	}

	bus.Unsubscribe("event", id1)
	bus.Publish("event", nil)

	if count1 != 2 {
		t.Error("handler1 should not be called after unsubscribe")
	}
	if count2 != 3 {
		t.Errorf("handler2 called %d times, want 3", count2)
	}
}

func TestDefaultEventBus_UnsubscribeUnknown(t *testing.T) {
	bus := NewDefaultEventBus()
	// Отписка от несуществующего события — не должна паниковать
	bus.Unsubscribe("nonexistent", 999)
}

func TestDefaultEventBus_PublishUnknownEvent(t *testing.T) {
	bus := NewDefaultEventBus()
	// Публикация без подписчиков — не должна паниковать
	bus.Publish("unknown_event", "data")
}

func TestDefaultEventBus_DifferentEvents(t *testing.T) {
	bus := NewDefaultEventBus()
	var eventA, eventB int

	handlerA := func(data interface{}) { eventA++ }
	handlerB := func(data interface{}) { eventB++ }

	bus.Subscribe("event_a", handlerA)
	bus.Subscribe("event_b", handlerB)

	bus.Publish("event_a", nil)
	bus.Publish("event_b", nil)
	bus.Publish("event_a", nil)

	if eventA != 2 {
		t.Errorf("eventA called %d times, want 2", eventA)
	}
	if eventB != 1 {
		t.Errorf("eventB called %d times, want 1", eventB)
	}
}

func TestDefaultEventBus_NilData(t *testing.T) {
	bus := NewDefaultEventBus()
	var received interface{}
	handler := func(data interface{}) { received = data }

	id := bus.Subscribe("event", handler)
	bus.Publish("event", nil)

	if received != nil {
		t.Errorf("expected nil, got %v", received)
	}

	bus.Unsubscribe("event", id)
}

// ============================================================================
// PluginRegistry — дополнительные тесты
// ============================================================================

func TestPluginRegistry_Empty(t *testing.T) {
	registry := NewPluginRegistry()
	all := registry.GetAll()
	if len(all) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(all))
	}

	p, ok := registry.Get("nonexistent")
	if ok {
		t.Error("expected not found")
	}
	if p != nil {
		t.Error("expected nil for not found")
	}
}

func TestPluginRegistry_MultiplePlugins(t *testing.T) {
	registry := NewPluginRegistry()

	plugins := []*simplePlugin{
		{info: Info{Name: "PluginA", Version: "1.0"}},
		{info: Info{Name: "PluginB", Version: "2.0"}},
		{info: Info{Name: "PluginC", Version: "3.0"}},
	}

	for _, p := range plugins {
		err := registry.Register(p)
		if err != nil {
			t.Fatalf("Register(%s) failed: %v", p.info.Name, err)
		}
	}

	all := registry.GetAll()
	if len(all) != 3 {
		t.Errorf("expected 3 plugins, got %d", len(all))
	}
}

func TestPluginRegistry_EventBus_Available(t *testing.T) {
	registry := NewPluginRegistry()
	eventBus := registry.EventBus()
	if eventBus == nil {
		t.Fatal("EventBus should not be nil")
	}

	var received bool
	eventBus.Subscribe("test", func(data interface{}) { received = true })
	eventBus.Publish("test", nil)
	if !received {
		t.Error("event should be received")
	}
}

func TestPluginRegistry_CloseAll_Failing(t *testing.T) {
	registry := NewPluginRegistry()

	// Добавляем плагин с ошибкой при закрытии
	closeErr := fmt.Errorf("close failed")
	failing := &failingPlugin{
		info:     Info{Name: "Failing"},
		closeErr: closeErr,
	}
	registry.Register(failing)

	// CloseAll должен вернуть ошибку
	err := registry.CloseAll()
	if err == nil {
		t.Error("expected error from CloseAll")
	}
	if err != closeErr {
		t.Errorf("expected closeErr, got %v", err)
	}
}

func TestPluginRegistry_CloseAll_Success(t *testing.T) {
	registry := NewPluginRegistry()

	registry.Register(&simplePlugin{info: Info{Name: "Plugin1"}})
	registry.Register(&simplePlugin{info: Info{Name: "Plugin2"}})

	err := registry.CloseAll()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// ============================================================================
// PluginType constants
// ============================================================================

func TestPluginTypeConstants(t *testing.T) {
	types := []Type{TypeFilter, TypeExporter, TypeScanner, TypeReporter}
	for _, tpe := range types {
		if tpe == "" {
			t.Error("plugin type should not be empty")
		}
	}

	if TypeFilter != "filter" {
		t.Errorf("TypeFilter = %q, want filter", TypeFilter)
	}
	if TypeExporter != "exporter" {
		t.Errorf("TypeExporter = %q, want exporter", TypeExporter)
	}
	if TypeScanner != "scanner" {
		t.Errorf("TypeScanner = %q, want scanner", TypeScanner)
	}
	if TypeReporter != "reporter" {
		t.Errorf("TypeReporter = %q, want reporter", TypeReporter)
	}
}

// ============================================================================
// Info struct
// ============================================================================

func TestPluginInfo_JSON(t *testing.T) {
	info := Info{
		Name:        "TestPlugin",
		Version:     "1.0.0",
		Description: "Test description",
		Author:      "TestAuthor",
		Type:        TypeFilter,
	}

	if info.Name != "TestPlugin" {
		t.Errorf("Name = %q, want TestPlugin", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", info.Version)
	}
	if info.Description != "Test description" {
		t.Errorf("Description = %q, want Test description", info.Description)
	}
	if info.Author != "TestAuthor" {
		t.Errorf("Author = %q, want TestAuthor", info.Author)
	}
	if info.Type != TypeFilter {
		t.Errorf("Type = %q, want filter", info.Type)
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkPluginRegistry_Register(b *testing.B) {
	registry := NewPluginRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.Register(&simplePlugin{info: Info{Name: "bench"}})
	}
}

func BenchmarkPluginRegistry_Get(b *testing.B) {
	registry := NewPluginRegistry()
	registry.Register(&simplePlugin{info: Info{Name: "bench"}})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.Get("bench")
	}
}

func BenchmarkEventBus_SubscribePublish(b *testing.B) {
	bus := NewDefaultEventBus()
	handler := func(data interface{}) {}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := bus.Subscribe("bench", handler)
		bus.Publish("bench", nil)
		bus.Unsubscribe("bench", id)
	}
}
