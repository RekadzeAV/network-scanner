// Package plugin определяет интерфейс для плагинов Network Scanner.
//
// Плагины позволяют расширять функциональность без перекомпиляции ядра.
// Поддерживаются типы плагинов:
//   - Filter: фильтрация результатов сканирования
//   - Exporter: экспорт результатов в новые форматы
//   - Scanner: дополнительные сканеры
//   - Reporter: генерация отчетов
//
// # Загрузка плагинов
//
// Плагин должен быть скомпилирован как shared library (.so на Linux,
// .dll на Windows, .dylib на macOS) и помещен в директорию plugins/.
//
// # Пример создания плагина
//
//	package main
//
//	import (
//	    "network-scanner/internal/plugin"
//	    "network-scanner/internal/contracts"
//	)
//
//	func main {}
//
//	// PluginInfo возвращает информацию о плагине
//	func PluginInfo() plugin.Info {
//	    return plugin.Info{
//	        Name:        "MyFilter",
//	        Version:     "1.0.0",
//	        Description: "Мой пользовательский фильтр",
//	        Author:      "Developer",
//	    }
//	}
//
//	// Init инициализирует плагин
//	func Init(cfg map[string]interface{}) error {
//	    // Инициализация
//	    return nil
//	}
//
//	// Run запускает плагин
//	func Run(ctx context.Context, results []contracts.ScanResult) ([]contracts.ScanResult, error) {
//	    // Логика фильтрации
//	    return results, nil
//	}
//
//	// Close осввобождает ресурсы
//	func Close() error {
//	    return nil
//	}
package plugin

import (
	"context"
	"time"

	"network-scanner/internal/contracts"
)

// Info содержит метаинформацию о плагине
type Info struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Type        Type   `json:"type"`
}

// Type определяет тип плагина
type Type string

const (
	TypeFilter   Type = "filter"   // Фильтрация результатов
	TypeExporter Type = "exporter" // Экспорт в новые форматы
	TypeScanner  Type = "scanner"  // Дополнительные сканеры
	TypeReporter Type = "reporter" // Генерация отчетов
)

// Plugin определяет интерфейс для всех плагинов
type Plugin interface {
	// Info возвращает метаинформацию о плагине
	Info() Info

	// Init инициализирует плагин с конфигурацией
	Init(cfg map[string]interface{}) error

	// Run запускает плагин с результатами сканирования
	// Для filter: фильтрует результаты
	// Для exporter: экспортирует результаты
	// Для scanner: запускает дополнительное сканирование
	// Для reporter: генерирует отчет
	Run(ctx context.Context, results []contracts.ScanResult) (interface{}, error)

	// Close освобождает ресурсы плагина
	Close() error
}

// FilterPlugin интерфейс для плагинов фильтрации
type FilterPlugin interface {
	Plugin
	// Filter применяет фильтр к результатам
	Filter(results []contracts.ScanResult) ([]contracts.ScanResult, error)
}

// ExporterPlugin интерфейс для плагинов экспорта
type ExporterPlugin interface {
	Plugin
	// Export экспортирует результаты в файл
	Export(results []contracts.ScanResult, path string) error
	// Format возвращает формат экспорта (csv, json, xml, pdf)
	Format() string
}

// ScannerPlugin интерфейс для дополнительных сканеров
type ScannerPlugin interface {
	Plugin
	// Scan запускает сканирование
	Scan(ctx context.Context, target string) ([]contracts.ScanResult, error)
}

// ReporterPlugin интерфейс для генерации отчетов
type ReporterPlugin interface {
	Plugin
	// Generate генерирует отчет
	Generate(results []contracts.ScanResult, format string) ([]byte, error)
}

// Loader загружает плагины из директории
type Loader interface {
	// Load загружает плагин по пути
	Load(path string) (Plugin, error)
	// LoadAll загружает все плагины из директории
	LoadAll(dir string) ([]Plugin, error)
	// Discover ищет плагины в стандартной директории
	Discover() ([]Plugin, error)
}

// EventBus предоставляет интерфейс для событий плагинов
type EventBus interface {
	// Subscribe подписывается на событие, возвращает ID для отписки
	Subscribe(event string, handler func(interface{})) DefaultEventBusHandler
	// Publish публикует событие
	Publish(event string, data interface{})
	// Unsubscribe отписывается от события по ID
	Unsubscribe(event string, id DefaultEventBusHandler)
}

// DefaultEventBusHandler идентификатор обработчика события
type DefaultEventBusHandler int

// DefaultEventBus реализует базовый EventBus
type DefaultEventBus struct {
	handlers map[string]map[DefaultEventBusHandler]func(interface{})
	nextID   DefaultEventBusHandler
}

// NewDefaultEventBus создает новый EventBus
func NewDefaultEventBus() *DefaultEventBus {
	return &DefaultEventBus{
		handlers: make(map[string]map[DefaultEventBusHandler]func(interface{})),
	}
}

// Subscribe реализует EventBus, возвращает ID для последующей отписки
func (eb *DefaultEventBus) Subscribe(event string, handler func(interface{})) DefaultEventBusHandler {
	if eb.handlers[event] == nil {
		eb.handlers[event] = make(map[DefaultEventBusHandler]func(interface{}))
	}
	eb.nextID++
	eb.handlers[event][eb.nextID] = handler
	return eb.nextID
}

// Publish реализует EventBus
func (eb *DefaultEventBus) Publish(event string, data interface{}) {
	if handlers, ok := eb.handlers[event]; ok {
		for _, handler := range handlers {
			handler(data)
		}
	}
}

// Unsubscribe отписывается от события по ID
func (eb *DefaultEventBus) Unsubscribe(event string, id DefaultEventBusHandler) {
	if handlers, ok := eb.handlers[event]; ok {
		delete(handlers, id)
	}
}

// PluginRegistry регистрирует и управляет плагинами
type PluginRegistry struct {
	plugins  map[string]Plugin
	eventBus EventBus
}

// NewPluginRegistry создает новый регистр плагинов
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins:  make(map[string]Plugin),
		eventBus: NewDefaultEventBus(),
	}
}

// Register регистрирует плагин
func (pr *PluginRegistry) Register(p Plugin) error {
	info := p.Info()
	if _, exists := pr.plugins[info.Name]; exists {
		return nil
	}
	pr.plugins[info.Name] = p
	return nil
}

// Get возвращает плагин по имени
func (pr *PluginRegistry) Get(name string) (Plugin, bool) {
	p, ok := pr.plugins[name]
	return p, ok
}

// GetAll возвращает все зарегистрированные плагины
func (pr *PluginRegistry) GetAll() []Plugin {
	result := make([]Plugin, 0, len(pr.plugins))
	for _, p := range pr.plugins {
		result = append(result, p)
	}
	return result
}

// EventBus возвращает EventBus регистра
func (pr *PluginRegistry) EventBus() EventBus {
	return pr.eventBus
}

// CloseAll закрывает все плагины
func (pr *PluginRegistry) CloseAll() error {
	for _, p := range pr.plugins {
		if err := p.Close(); err != nil {
			return err
		}
	}
	return nil
}

// PluginContext предоставляет контекст для выполнения плагина
type PluginContext struct {
	ScanID        string
	StartTime     time.Time
	CancelContext context.Context
	CancelFunc    context.CancelFunc
}

// NewPluginContext создает новый контекст плагина
func NewPluginContext(scanID string) *PluginContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &PluginContext{
		ScanID:        scanID,
		StartTime:     time.Now(),
		CancelContext: ctx,
		CancelFunc:    cancel,
	}
}
