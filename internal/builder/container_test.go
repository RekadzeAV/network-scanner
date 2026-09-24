package builder

import (
	"testing"

	"network-scanner/internal/eventbus"
)

func TestNewContainer(t *testing.T) {
	cfg := Config{
		LogLevel: "debug",
		DBPath:   "",
	}
	c := NewContainer(cfg)
	if c == nil {
		t.Fatal("expected non-nil Container")
	}
}

func TestContainer_GetScanner(t *testing.T) {
	c := NewContainer(Config{LogLevel: "info"})
	if c.GetScanner() == nil {
		t.Fatal("expected non-nil scanner service")
	}
}

func TestContainer_GetTopology(t *testing.T) {
	c := NewContainer(Config{})
	if c.GetTopology() == nil {
		t.Fatal("expected non-nil topology service")
	}
}

func TestContainer_GetSecurity(t *testing.T) {
	c := NewContainer(Config{})
	if c.GetSecurity() == nil {
		t.Fatal("expected non-nil security service")
	}
}

func TestContainer_GetRemoteExec(t *testing.T) {
	c := NewContainer(Config{})
	if c.GetRemoteExec() == nil {
		t.Fatal("expected non-nil remote exec service")
	}
}

func TestContainer_GetInventory(t *testing.T) {
	c := NewContainer(Config{})
	if c.GetInventory() == nil {
		t.Fatal("expected non-nil inventory service")
	}
}

// TestContainer_WithEventBus — E6: шина событий устанавливается и доступна.
func TestContainer_WithEventBus(t *testing.T) {
	c := NewContainer(Config{LogLevel: "info"})
	if c.GetEventBus() != nil {
		t.Fatal("expected nil event bus before wiring")
	}

	bus := eventbus.NewEventBus()
	defer bus.Close()

	if got := c.WithEventBus(bus); got != c {
		t.Fatal("WithEventBus should return the same container (fluent API)")
	}
	if c.GetEventBus() != bus {
		t.Fatal("expected wired event bus to be returned by GetEventBus")
	}
}

// TestContainer_WithEventBus_ScannerServiceWired — E6: сканер получает шину
// через контейнер (type assertion на интерфейс WithEventBus).
func TestContainer_WithEventBus_ScannerServiceWired(t *testing.T) {
	bus := eventbus.NewEventBus()
	defer bus.Close()

	c := NewContainer(Config{LogLevel: "info"}).WithEventBus(bus)
	if c.GetEventBus() == nil {
		t.Fatal("expected wired event bus")
	}
	if c.GetScanner() == nil {
		t.Fatal("expected non-nil scanner service")
	}
	if _, ok := c.GetScanner().(interface {
		WithEventBus(*eventbus.EventBus)
	}); !ok {
		t.Fatal("expected scanner service to expose WithEventBus(*eventbus.EventBus)")
	}
}
