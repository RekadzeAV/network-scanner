package plugin_test

import (
	"context"
	"testing"

	"network-scanner/internal/contracts"
	"network-scanner/internal/plugin"
)

func TestNewPluginRegistry(t *testing.T) {
	registry := plugin.NewPluginRegistry()
	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}
}

func TestPluginRegistry_RegisterAndGet(t *testing.T) {
	registry := plugin.NewPluginRegistry()

	// Создаем мок-плагин для теста
	mockPlugin := &mockPlugin{
		info: plugin.Info{
			Name:        "TestPlugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Type:        plugin.TypeFilter,
		},
	}

	err := registry.Register(mockPlugin)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	p, ok := registry.Get("TestPlugin")
	if !ok {
		t.Fatal("Expected plugin to be found")
	}

	if p.Info().Name != "TestPlugin" {
		t.Errorf("Expected name 'TestPlugin', got '%s'", p.Info().Name)
	}
}

func TestPluginRegistry_GetNonExistent(t *testing.T) {
	registry := plugin.NewPluginRegistry()

	_, ok := registry.Get("NonExistent")
	if ok {
		t.Error("Expected plugin not to be found")
	}
}

func TestPluginRegistry_GetAll(t *testing.T) {
	registry := plugin.NewPluginRegistry()

	plugin1 := &mockPlugin{info: plugin.Info{Name: "Plugin1"}}
	plugin2 := &mockPlugin{info: plugin.Info{Name: "Plugin2"}}

	registry.Register(plugin1)
	registry.Register(plugin2)

	all := registry.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(all))
	}
}

func TestDefaultEventBus(t *testing.T) {
	bus := plugin.NewDefaultEventBus()

	received := make([]interface{}, 0)
	id := bus.Subscribe("test_event", func(data interface{}) {
		received = append(received, data)
	})

	bus.Publish("test_event", "test_data")

	if len(received) != 1 {
		t.Errorf("Expected 1 received event, got %d", len(received))
	}

	if received[0] != "test_data" {
		t.Errorf("Expected 'test_data', got '%v'", received[0])
	}

	bus.Unsubscribe("test_event", id)
}

func TestDefaultEventBus_Unsubscribe(t *testing.T) {
	bus := plugin.NewDefaultEventBus()

	id := bus.Subscribe("test_event", func(data interface{}) {})
	bus.Unsubscribe("test_event", id)

	// После отписки событие не должно обрабатываться
	bus.Publish("test_event", "data")
	// Если код дошел сюда без panic — тест пройден
}

// mockPlugin — мок для тестирования
type mockPlugin struct {
	info plugin.Info
}

func (m *mockPlugin) Info() plugin.Info {
	return m.info
}

func (m *mockPlugin) Init(cfg map[string]interface{}) error {
	return nil
}

func (m *mockPlugin) Run(_ context.Context, _ []contracts.ScanResult) (interface{}, error) {
	return nil, nil
}

func (m *mockPlugin) Close() error {
	return nil
}
