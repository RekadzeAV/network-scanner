// Package plugin предоставляет модульную архитектуру для probe handlers.
//
// # Основные компоненты
//
// Probe — интерфейс для всех типов probe (ICMP, SNMP, SSH, HTTP и т.д.)
//
//	Execute(ctx, host) — выполнение probe для хоста
//	Name() string — имя probe для логирования
//	Priority() int — порядок выполнения (меньше = раньше)
//
// ProbeRegistry — реестр для регистрации и управления плагинами
//
//	Register() — регистрация probe
//	GetByName() — получение probe по имени
//	ExecuteAll() — последовательное выполнение всех probe
//
// # Порядок выполнения
//
//  1. HostDiscovery (priority 100) — ICMP ping, ARP, port check
//  2. PortScan (priority 200) — TCP/UDP port scanning
//  3. ServiceProbe (priority 300) — SNMP, SSH banner, HTTP fingerprint
//  4. DeviceInfo (priority 400) — OS detection, vendor lookup, MAC OUI
//
// # Пример добавления нового probe
//
//	type SSHBannerProbe struct{}
//
//	func (p *SSHBannerProbe) Name() string { return "ssh-banner" }
//	func (p *SSHBannerProbe) Priority() int { return 310 }
//	func (p *SSHBannerProbe) Execute(ctx context.Context, host string, port int) (*ProbeResult, error) {
//	    // SSH banner grabbing logic
//	    return &ProbeResult{}, nil
//	}
//
//	// Регистрация
//	registry := plugin.NewRegistry()
//	registry.Register(&SSHBannerProbe{})

package plugin

import (
	"context"
	"fmt"
	"sync"
)

// ProbePhase определяет фазу выполнения probe
type ProbePhase int

const (
	// PhaseHostDiscovery — обнаружение хостов (ICMP, ARP, port check)
	PhaseHostDiscovery ProbePhase = 100
	// PhasePortScan — сканирование портов (TCP/UDP)
	PhasePortScan ProbePhase = 200
	// PhaseServiceProbe — анализ служб (SNMP, SSH, HTTP)
	PhaseServiceProbe ProbePhase = 300
	// PhaseDeviceInfo — информация об устройстве (OS, vendor, MAC)
	PhaseDeviceInfo ProbePhase = 400
)

// ProbeResult содержит результаты выполнения probe
type ProbeResult struct {
	// Service определяет тип службы (http, ssh, snmp, ftp и т.д.)
	Service string
	// Version — версия службы (опционально)
	Version string
	// Banner — сырой баннер/ответ службы
	Banner string
	// Extra — дополнительные данные (key-value)
	Extra map[string]string
	// Error — ошибка probe (если была)
	Error error
}

// NewProbeResult создает новый пустой результат probe
func NewProbeResult() *ProbeResult {
	return &ProbeResult{
		Extra: make(map[string]string),
	}
}

// Probe определяет интерфейс для всех типов probe
type Probe interface {
	// Name возвращает уникальное имя probe
	Name() string

	// Phase возвращает фазу выполнения probe
	Phase() ProbePhase

	// Priority определяет порядок выполнения (меньше = раньше)
	// Внутри одной фазы probe выполняются в порядке регистрации
	Priority() int

	// Execute выполняет probe для указанного хоста и порта
	// host — IP адрес или hostname
	// port — порт для probe (0 для ICMP/ARP)
	// ctx — контекст для отмены операции
	Execute(ctx context.Context, host string, port int) (*ProbeResult, error)
}

// ProbeHandler обрабатывает результаты probe и обновляет состояние
type ProbeHandler func(probeName string, result *ProbeResult, host string, port int) error

// Registry управляет реестром probe плагинов
type Registry struct {
	mu       sync.RWMutex
	probes   []Probe
	handlers []ProbeHandler
}

// NewRegistry создает новый реестр probe
func NewRegistry() *Registry {
	return &Registry{
		probes:   make([]Probe, 0),
		handlers: make([]ProbeHandler, 0),
	}
}

// Register регистрирует новый probe в реестре
func (r *Registry) Register(probe Probe) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем дубликат имени
	for _, p := range r.probes {
		if p.Name() == probe.Name() {
			return // Уже зарегистрирован
		}
	}

	r.probes = append(r.probes, probe)
}

// RegisterHandler регистрирует обработчик результатов probe
func (r *Registry) RegisterHandler(handler ProbeHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers = append(r.handlers, handler)
}

// GetByName возвращает probe по имени
func (r *Registry) GetByName(name string) Probe {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.probes {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// GetAll возвращает все зарегистрированные probe, отсортированные по фазе и приоритету
func (r *Registry) GetAll() []Probe {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Копируем слайс для безопасной сортировки
	result := make([]Probe, len(r.probes))
	copy(result, r.probes)

	// Сортируем по фазе и приоритету
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Phase() > result[j].Phase() ||
				(result[i].Phase() == result[j].Phase() && result[i].Priority() > result[j].Priority()) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// GetByPhase возвращает probe для указанной фазы
func (r *Registry) GetByPhase(phase ProbePhase) []Probe {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Probe
	for _, p := range r.probes {
		if p.Phase() == phase {
			result = append(result, p)
		}
	}

	// Сортируем по приоритету
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Priority() > result[j].Priority() {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// ExecuteAll выполняет все probe последовательно для указанного хоста и порта
func (r *Registry) ExecuteAll(ctx context.Context, host string, port int) ([]*ProbeResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allResults []*ProbeResult
	var firstErr error

	for _, probe := range r.probes {
		result, err := probe.Execute(ctx, host, port)
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("probe %s failed: %w", probe.Name(), err)
		}

		if result != nil {
			allResults = append(allResults, result)

			// Вызываем обработчики результатов
			for _, handler := range r.handlers {
				if err := handler(probe.Name(), result, host, port); err != nil {
					return allResults, fmt.Errorf("handler for %s failed: %w", probe.Name(), err)
				}
			}
		}
	}

	return allResults, firstErr
}

// ExecutePhase выполняет все probe для указанной фазы
func (r *Registry) ExecutePhase(ctx context.Context, phase ProbePhase, host string, port int) ([]*ProbeResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	probes := r.GetByPhase(phase)
	var allResults []*ProbeResult
	var firstErr error

	for _, probe := range probes {
		result, err := probe.Execute(ctx, host, port)
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("probe %s (phase %d) failed: %w", probe.Name(), phase, err)
		}

		if result != nil {
			allResults = append(allResults, result)

			// Вызываем обработчики результатов
			for _, handler := range r.handlers {
				if err := handler(probe.Name(), result, host, port); err != nil {
					return allResults, fmt.Errorf("handler for %s failed: %w", probe.Name(), err)
				}
			}
		}
	}

	return allResults, firstErr
}

// Count возвращает количество зарегистрированных probe
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.probes)
}

// CountByPhase возвращает количество probe для указанной фазы
func (r *Registry) CountByPhase(phase ProbePhase) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, p := range r.probes {
		if p.Phase() == phase {
			count++
		}
	}
	return count
}
