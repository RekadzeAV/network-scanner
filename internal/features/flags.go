package features

import (
	"sync"
	"sync/atomic"
)

// Flag представляет один feature-flag с именем, описанием и состоянием.
type Flag struct {
	name        string
	description string
	enabled     atomic.Bool
	defaultEnabled bool
}

// Manager управляет набором feature-flags.
type Manager struct {
	mu   sync.RWMutex
	flags map[string]*Flag
}

// NewManager создаёт новый менеджер feature-flags.
func NewManager() *Manager {
	return &Manager{
		flags: make(map[string]*Flag),
	}
}

// Register регистрирует новый feature-flag.
func (m *Manager) Register(name, description string, defaultEnabled bool) *Flag {
	m.mu.Lock()
	defer m.mu.Unlock()

	f := &Flag{
		name:           name,
		description:    description,
		defaultEnabled: defaultEnabled,
	}
	f.enabled.Store(defaultEnabled)
	m.flags[name] = f
	return f
}

// IsEnabled возвращает текущее состояние флага.
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	f, ok := m.flags[name]
	m.mu.RUnlock()
	if !ok {
		return false
	}
	return f.enabled.Load()
}

// SetEnabled устанавливает состояние флага.
func (m *Manager) SetEnabled(name string, enabled bool) {
	m.mu.RLock()
	f, ok := m.flags[name]
	m.mu.RUnlock()
	if !ok {
		return
	}
	f.enabled.Store(enabled)
}

// Toggle переключает состояние флага.
func (m *Manager) Toggle(name string) bool {
	m.mu.RLock()
	f, ok := m.flags[name]
	m.mu.RUnlock()
	if !ok {
		return false
	}
	newState := !f.enabled.Load()
	f.enabled.Store(newState)
	return newState
}

// Flags возвращает список всех зарегистрированных флагов.
func (m *Manager) Flags() []*Flag {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]*Flag, 0, len(m.flags))
	for _, f := range m.flags {
		out = append(out, f)
	}
	return out
}

// StatusReport возвращает строковое представление всех флагов.
func (m *Manager) StatusReport() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report := "Feature Flags Status:\n"
	for _, f := range m.flags {
		state := "OFF"
		if f.enabled.Load() {
			state = "ON"
		}
		report += "  [" + state + "] " + f.name + ": " + f.description + "\n"
	}
	return report
}

// --- Предопределённые флаги для D-трека ---

// DefaultManager — глобальный экземпляр менеджера флагов.
var DefaultManager = NewManager()

// D1TopologyHardening — включение улучшенной дедупликации LLDP/FDB связей.
var D1TopologyHardening = DefaultManager.Register(
	"d1.topology.hardening",
	"Улучшенная дедупликация LLDP/FDB связей с confidence rules",
	true, // включён по умолчанию
)

// D1TopologyFallback — fallback режим без Graphviz (текстовый вывод).
var D1TopologyFallback = DefaultManager.Register(
	"d1.topology.fallback",
	"Показать текстовую топологию при отсутствии Graphviz",
	true, // включён по умолчанию
)

// D2ExportSchemaValidation — валидация JSON schema при экспорте топологии.
var D2ExportSchemaValidation = DefaultManager.Register(
	"d2.export.schema_validation",
	"Строгая валидация JSON schema при экспорте топологии",
	true, // включён по умолчанию
)

// D2GraphMLEquivalence — проверка эквивалентности json/graphml экспорта.
var D2GraphMLEquivalence = DefaultManager.Register(
	"d2.export.graphml_equivalence",
	"Проверка эквивалентности графа между json и graphml форматами",
	false, // отключён по умолчанию (тяжёлый тест)
)

// D3GUIResponsive — responsive-тесты для карточек/таблиц.
var D3GUIResponsive = DefaultManager.Register(
	"d3.gui.responsive",
	"Responsive-адаптация UI для разных размеров окна",
	true, // включён по умолчанию
)

// D3PerfBudget — проверка perf-budget при рендере результатов.
var D3PerfBudget = DefaultManager.Register(
	"d3.gui.perf_budget",
	"Ограничение времени перерисовки результатов (perf-budget)",
	false, // отключён по умолчанию (профилирование)
)

// D4RollbackEnabled — включение rollback-механизма для нестабильных изменений.
var D4RollbackEnabled = DefaultManager.Register(
	"d4.rollback.enabled",
	"Включение rollback-механизма для отката нестабильных изменений",
	false, // отключён по умолчанию
)

// InitDefaultFlags инициализирует все предопределённые флаги.
// Вызывается при старте приложения.
func InitDefaultFlags() {
	// Флаги уже зарегистрированы при инициализации переменных.
	// Эта функция может использоваться для дополнительной настройки.
}
