package plugin

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// ============================================================================
// C1: Тесты для Plugin System
// ============================================================================

// mockProbe — тестовая реализация Probe для тестирования
type mockProbe struct {
	name     string
	phase    ProbePhase
	priority int
	execute  func(ctx context.Context, host string, port int) (*ProbeResult, error)
}

func (m *mockProbe) Name() string {
	return m.name
}

func (m *mockProbe) Phase() ProbePhase {
	return m.phase
}

func (m *mockProbe) Priority() int {
	return m.priority
}

func (m *mockProbe) Execute(ctx context.Context, host string, port int) (*ProbeResult, error) {
	if m.execute != nil {
		return m.execute(ctx, host, port)
	}
	return NewProbeResult(), nil
}

// TestRegistry_New — ветка: создание реестра
func TestRegistry_New(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
	if registry.Count() != 0 {
		t.Errorf("expected 0 probes, got %d", registry.Count())
	}
}

// TestRegistry_Register — ветка: регистрация probe
func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	probe := &mockProbe{
		name:     "test-probe",
		phase:    PhaseHostDiscovery,
		priority: 100,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			return NewProbeResult(), nil
		},
	}

	registry.Register(probe)

	if registry.Count() != 1 {
		t.Errorf("expected 1 probe, got %d", registry.Count())
	}

	// Проверяем что probe можно получить по имени
	found := registry.GetByName("test-probe")
	if found == nil {
		t.Fatal("expected to find probe by name")
	}
	if found.Name() != "test-probe" {
		t.Errorf("expected name test-probe, got %s", found.Name())
	}
}

// TestRegistry_Register_Duplicate — ветка: дубликаты игнорируются
func TestRegistry_Register_Duplicate(t *testing.T) {
	registry := NewRegistry()

	probe := &mockProbe{name: "duplicate", phase: PhasePortScan, priority: 200}
	registry.Register(probe)
	registry.Register(probe) // Повторная регистрация

	if registry.Count() != 1 {
		t.Errorf("expected 1 probe (duplicate ignored), got %d", registry.Count())
	}
}

// TestRegistry_GetByName_NotFound — ветка: probe не найден
func TestRegistry_GetByName_NotFound(t *testing.T) {
	registry := NewRegistry()

	found := registry.GetByName("nonexistent")
	if found != nil {
		t.Error("expected nil for nonexistent probe")
	}
}

// TestRegistry_GetAll_Sorted — ветка: сортировка по фазе и приоритету
func TestRegistry_GetAll_Sorted(t *testing.T) {
	registry := NewRegistry()

	// Регистрируем в случайном порядке
	registry.Register(&mockProbe{name: "high-priority", phase: PhaseHostDiscovery, priority: 50})
	registry.Register(&mockProbe{name: "low-priority", phase: PhaseHostDiscovery, priority: 200})
	registry.Register(&mockProbe{name: "service-probe", phase: PhaseServiceProbe, priority: 300})

	all := registry.GetAll()

	if len(all) != 3 {
		t.Fatalf("expected 3 probes, got %d", len(all))
	}

	// Проверяем порядок: сначала фаза, потом приоритет
	if all[0].Name() != "high-priority" {
		t.Errorf("expected high-priority first, got %s", all[0].Name())
	}
	if all[1].Name() != "low-priority" {
		t.Errorf("expected low-priority second, got %s", all[1].Name())
	}
	if all[2].Name() != "service-probe" {
		t.Errorf("expected service-probe third, got %s", all[2].Name())
	}
}

// TestRegistry_GetByPhase — ветка: получение probe по фазе
func TestRegistry_GetByPhase(t *testing.T) {
	registry := NewRegistry()

	registry.Register(&mockProbe{name: "icmp", phase: PhaseHostDiscovery, priority: 100})
	registry.Register(&mockProbe{name: "arp", phase: PhaseHostDiscovery, priority: 200})
	registry.Register(&mockProbe{name: "tcp-scan", phase: PhasePortScan, priority: 100})

	// Получаем probe для фазы HostDiscovery
	discoveryProbes := registry.GetByPhase(PhaseHostDiscovery)
	if len(discoveryProbes) != 2 {
		t.Errorf("expected 2 discovery probes, got %d", len(discoveryProbes))
	}

	// Проверяем порядок по приоритету
	if discoveryProbes[0].Name() != "icmp" {
		t.Errorf("expected icmp first, got %s", discoveryProbes[0].Name())
	}

	// Получаем probe для фазы PortScan
	portProbes := registry.GetByPhase(PhasePortScan)
	if len(portProbes) != 1 {
		t.Errorf("expected 1 port probe, got %d", len(portProbes))
	}

	// Получаем probe для несуществующей фазы
	emptyProbes := registry.GetByPhase(PhaseDeviceInfo)
	if len(emptyProbes) != 0 {
		t.Errorf("expected 0 device info probes, got %d", len(emptyProbes))
	}
}

// TestRegistry_ExecuteAll — ветка: выполнение всех probe
func TestRegistry_ExecuteAll(t *testing.T) {
	registry := NewRegistry()

	executeCount := 0
	registry.Register(&mockProbe{
		name:     "probe-1",
		phase:    PhaseHostDiscovery,
		priority: 100,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			executeCount++
			return NewProbeResult(), nil
		},
	})

	registry.Register(&mockProbe{
		name:     "probe-2",
		phase:    PhasePortScan,
		priority: 200,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			executeCount++
			return NewProbeResult(), nil
		},
	})

	ctx := context.Background()
	results, err := registry.ExecuteAll(ctx, "192.168.1.1", 80)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if executeCount != 2 {
		t.Errorf("expected 2 executions, got %d", executeCount)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

// TestRegistry_ExecutePhase — ветка: выполнение probe для фазы
func TestRegistry_ExecutePhase(t *testing.T) {
	registry := NewRegistry()

	registry.Register(&mockProbe{
		name:     "icmp",
		phase:    PhaseHostDiscovery,
		priority: 100,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			return NewProbeResult(), nil
		},
	})

	registry.Register(&mockProbe{
		name:     "tcp-scan",
		phase:    PhasePortScan,
		priority: 100,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			return NewProbeResult(), nil
		},
	})

	ctx := context.Background()
	results, err := registry.ExecutePhase(ctx, PhaseHostDiscovery, "192.168.1.1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for discovery phase, got %d", len(results))
	}

	// Проверка что port-scan probe не выполнен
	portResults, err := registry.ExecutePhase(ctx, PhasePortScan, "192.168.1.1", 80)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(portResults) != 1 {
		t.Errorf("expected 1 result for port phase, got %d", len(portResults))
	}
}

// TestRegistry_Handler — ветка: обработчики результатов
func TestRegistry_Handler(t *testing.T) {
	registry := NewRegistry()

	handlerCalled := false
	registry.RegisterHandler(func(probeName string, result *ProbeResult, host string, port int) error {
		handlerCalled = true
		if probeName != "test-probe" {
			return fmt.Errorf("unexpected probe name: %s", probeName)
		}
		return nil
	})

	registry.Register(&mockProbe{
		name:     "test-probe",
		phase:    PhaseHostDiscovery,
		priority: 100,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			return NewProbeResult(), nil
		},
	})

	ctx := context.Background()
	_, err := registry.ExecuteAll(ctx, "192.168.1.1", 80)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

// TestRegistry_Handler_Error — ветка: ошибка в обработчике
func TestRegistry_Handler_Error(t *testing.T) {
	registry := NewRegistry()

	registry.RegisterHandler(func(probeName string, result *ProbeResult, host string, port int) error {
		return fmt.Errorf("handler error")
	})

	registry.Register(&mockProbe{
		name:     "test-probe",
		phase:    PhaseHostDiscovery,
		priority: 100,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			return NewProbeResult(), nil
		},
	})

	ctx := context.Background()
	_, err := registry.ExecuteAll(ctx, "192.168.1.1", 80)
	if err == nil {
		t.Error("expected error from handler")
	}
}

// TestRegistry_Execute_ProbeError — ветка: ошибка в probe
func TestRegistry_Execute_ProbeError(t *testing.T) {
	registry := NewRegistry()

	registry.Register(&mockProbe{
		name:     "failing-probe",
		phase:    PhaseHostDiscovery,
		priority: 100,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			return nil, fmt.Errorf("probe failed")
		},
	})

	registry.Register(&mockProbe{
		name:     "success-probe",
		phase:    PhaseHostDiscovery,
		priority: 200,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			return NewProbeResult(), nil
		},
	})

	ctx := context.Background()
	results, err := registry.ExecuteAll(ctx, "192.168.1.1", 80)
	if err == nil {
		t.Error("expected error from failing probe")
	}

	// Второй probe должен выполниться несмотря на ошибку первого
	if len(results) != 1 {
		t.Errorf("expected 1 result (second probe), got %d", len(results))
	}
}

// TestRegistry_ContextCancellation — ветка: отмена контекста
func TestRegistry_ContextCancellation(t *testing.T) {
	registry := NewRegistry()

	registry.Register(&mockProbe{
		name:     "slow-probe",
		phase:    PhaseHostDiscovery,
		priority: 100,
		execute: func(ctx context.Context, host string, port int) (*ProbeResult, error) {
			select {
			case <-time.After(100 * time.Millisecond):
				return NewProbeResult(), nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Немедленная отмена

	_, err := registry.ExecuteAll(ctx, "192.168.1.1", 80)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}

// TestProbeResult_New — ветка: создание результата probe
func TestProbeResult_New(t *testing.T) {
	result := NewProbeResult()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Extra == nil {
		t.Error("expected non-nil Extra map")
	}
}

// TestProbeResult_SetExtra — ветка: дополнительные данные
func TestProbeResult_SetExtra(t *testing.T) {
	result := NewProbeResult()
	result.Service = "http"
	result.Version = "nginx/1.21"
	result.Extra["os"] = "linux"
	result.Extra["vendor"] = "cisco"

	if result.Service != "http" {
		t.Errorf("expected service http, got %s", result.Service)
	}
	if result.Version != "nginx/1.21" {
		t.Errorf("expected version nginx/1.21, got %s", result.Version)
	}
	if result.Extra["os"] != "linux" {
		t.Errorf("expected os linux, got %s", result.Extra["os"])
	}
}

// TestRegistry_CountByPhase — ветка: подсчёт probe по фазе
func TestRegistry_CountByPhase(t *testing.T) {
	registry := NewRegistry()

	registry.Register(&mockProbe{name: "icmp", phase: PhaseHostDiscovery, priority: 100})
	registry.Register(&mockProbe{name: "arp", phase: PhaseHostDiscovery, priority: 200})
	registry.Register(&mockProbe{name: "tcp", phase: PhasePortScan, priority: 100})
	registry.Register(&mockProbe{name: "snmp", phase: PhaseServiceProbe, priority: 100})

	if registry.CountByPhase(PhaseHostDiscovery) != 2 {
		t.Errorf("expected 2 discovery probes")
	}
	if registry.CountByPhase(PhasePortScan) != 1 {
		t.Errorf("expected 1 port probe")
	}
	if registry.CountByPhase(PhaseServiceProbe) != 1 {
		t.Errorf("expected 1 service probe")
	}
	if registry.CountByPhase(PhaseDeviceInfo) != 0 {
		t.Errorf("expected 0 device info probes")
	}
}

// TestProbe_PhaseConstants — ветка: константы фаз
func TestProbe_PhaseConstants(t *testing.T) {
	if PhaseHostDiscovery != 100 {
		t.Error("expected PhaseHostDiscovery = 100")
	}
	if PhasePortScan != 200 {
		t.Error("expected PhasePortScan = 200")
	}
	if PhaseServiceProbe != 300 {
		t.Error("expected PhaseServiceProbe = 300")
	}
	if PhaseDeviceInfo != 400 {
		t.Error("expected PhaseDeviceInfo = 400")
	}
}
