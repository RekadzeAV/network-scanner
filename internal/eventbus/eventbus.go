// Package eventbus предоставляет легковесную систему событий для decoupling компонентов.
//
// # Основные компоненты
//
// EventBus — центральный шина событий для публикации и подписки
//
//	NewEventBus() — создает новую шину событий
//	Subscribe() — подписка на события с фильтром по типу
//	Publish() — публикация события всем подписчикам
//	HandleOnce() — обработка события один раз
//	Close() — остановка всех goroutine
//
// Event — базовый интерфейс для всех событий
//
//	Type() string — уникальный тип события
//	Timestamp() time.Time — время возникновения
//
// # Типы событий
//
// ScannerEvents:
//   - ScanStarted — начало сканирования
//   - ScanCompleted — завершение сканирования
//   - HostFound — найден новый хост
//   - PortOpened — открыт порт
//   - DeviceDetected — обнаружено устройство
//
// PluginEvents:
//   - ProbeStarted — начало выполнения probe
//   - ProbeCompleted — завершение probe
//   - ProbeFailed — ошибка probe
//
// GUIEvents:
//   - ProgressUpdated — обновление прогресса
//   - ThemeChanged — изменение темы
//   - SettingsChanged — изменение настроек
//
// # Пример использования
//
//	bus := eventbus.NewEventBus()
//	defer bus.Close()
//
//	// Подписка на события
//	bus.Subscribe("scan.started", func(e Event) {
//	    fmt.Println("Scan started!")
//	})
//
//	// Публикация события
//	bus.Publish(&ScanStartedEvent{})

package eventbus

import (
	"sync"
	"time"
)

// Event определяет базовый интерфейс для всех событий
type Event interface {
	// Type возвращает уникальный тип события
	Type() string

	// Timestamp возвращает время возникновения события
	Timestamp() time.Time

	// Payload возвращает дополнительные данные события (опционально)
	Payload() map[string]interface{}
}

// EventHandler определяет функцию-обработчик события
type EventHandler func(event Event)

// subscription — зарегистрированный обработчик с устойчивым идентификатором.
//
// Идентификатор нужен для корректного удаления: удаление по индексу теряет
// смысл при конкурентных Subscribe/HandleOnce и сдвигает другие обработчики.
type subscription struct {
	id      uint64
	handler EventHandler
}

// EventBus — центральный шина событий для публикации и подписки
type EventBus struct {
	mu                 sync.RWMutex
	subscribers        map[string][]subscription
	defaultSubscribers []subscription
	nextSubscriptionID uint64
	closed             bool
	done               chan struct{}
}

// NewEventBus создает новую шину событий
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]subscription),
		done:        make(chan struct{}),
	}
}

// Subscribe подписывает обработчик на события указанного типа
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return
	}

	eb.nextSubscriptionID++
	eb.subscribers[eventType] = append(eb.subscribers[eventType], subscription{
		id:      eb.nextSubscriptionID,
		handler: handler,
	})
}

// SubscribeAny подписывает обработчик на все события
func (eb *EventBus) SubscribeAny(handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return
	}

	eb.nextSubscriptionID++
	eb.defaultSubscribers = append(eb.defaultSubscribers, subscription{
		id:      eb.nextSubscriptionID,
		handler: handler,
	})
}

// HandleOnce подписывает обработчик который сработает один раз
//
// Однократность обеспечивает sync.Once: между публикацией и удалением
// обработчика следующее событие успевает его застать, и без once колбэк
// вызывался бы несколько раз.
func (eb *EventBus) HandleOnce(eventType string, handler EventHandler) {
	var once sync.Once

	eb.mu.Lock()
	if eb.closed {
		eb.mu.Unlock()
		return
	}
	eb.nextSubscriptionID++
	id := eb.nextSubscriptionID

	eb.subscribers[eventType] = append(eb.subscribers[eventType], subscription{
		id: id,
		handler: func(event Event) {
			once.Do(func() {
				eb.removeByID(eventType, id)
				handler(event)
			})
		},
	})
	eb.mu.Unlock()
}

// removeByID удаляет обработчик по идентификатору.
//
// Новый слайс вместо append(h[:i], h[i+1:]...): мутация «на месте» конкурирует
// с итерацией по тому же слайсу в Publish и приводит к пропущенным либо
// повторным вызовам обработчиков.
func (eb *EventBus) removeByID(eventType string, id uint64) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers, ok := eb.subscribers[eventType]
	if !ok {
		return
	}

	remaining := make([]subscription, 0, len(handlers))
	for _, sub := range handlers {
		if sub.id != id {
			remaining = append(remaining, sub)
		}
	}

	if len(remaining) == 0 {
		delete(eb.subscribers, eventType)
		return
	}
	eb.subscribers[eventType] = remaining
}

// Publish публикует событие всем подписчикам
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	if eb.closed {
		eb.mu.RUnlock()
		return
	}

	// Снимки списков: подписчики добавляются и удаляются конкурентно,
	// итерация по исходным слайсам дала бы гонку.
	eventType := event.Type()
	typed := make([]subscription, len(eb.subscribers[eventType]))
	copy(typed, eb.subscribers[eventType])

	any := make([]subscription, len(eb.defaultSubscribers))
	copy(any, eb.defaultSubscribers)
	eb.mu.RUnlock()

	// Вызываем специфичных подписчиков
	for _, sub := range typed {
		go eb.callHandler(sub.handler, event)
	}

	// Вызываем общих подписчиков
	for _, sub := range any {
		go eb.callHandler(sub.handler, event)
	}
}

// callHandler вызывает обработчик с recover для panic protection
func (eb *EventBus) callHandler(handler EventHandler, event Event) {
	defer func() {
		if r := recover(); r != nil {
			// Логирование panic в обработчике события
			_ = r
		}
	}()

	handler(event)
}

// PublishBatch публикует несколько событий batch-обработкой
func (eb *EventBus) PublishBatch(events []Event) {
	for _, event := range events {
		eb.Publish(event)
	}
}

// Unsubscribe удаляет всех подписчиков для указанного типа события
func (eb *EventBus) Unsubscribe(eventType string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	delete(eb.subscribers, eventType)
}

// UnsubscribeHandler удаляет конкретного подписчика
// Примечание: из-за ограничений Go (func нельзя сравнивать),
// этот метод удаляет ВСЕХ подписчиков для указанного типа.
// Используйте Unsubscribe(eventType) для явного удаления.
func (eb *EventBus) UnsubscribeHandler(eventType string, handler EventHandler) {
	// Игнорируем handler — удаляем все подписчики для типа
	eb.Unsubscribe(eventType)
}

// Close останавливает шину событий
func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.closed = true
	close(eb.done)
}

// IsClosed возвращает true если шина закрыта
func (eb *EventBus) IsClosed() bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.closed
}

// SubscriberCount возвращает количество подписчиков для типа события
func (eb *EventBus) SubscriberCount(eventType string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscribers[eventType])
}

// TotalSubscriberCount возвращает общее количество подписчиков
func (eb *EventBus) TotalSubscriberCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	count := len(eb.defaultSubscribers)
	for _, handlers := range eb.subscribers {
		count += len(handlers)
	}
	return count
}

// ============================================================================
// Базовые события
// ============================================================================

// BaseEvent — базовая реализация Event
type BaseEvent struct {
	eventType string
	timestamp time.Time
	payload   map[string]interface{}
}

// NewBaseEvent создает новое базовое событие
func NewBaseEvent(eventType string) *BaseEvent {
	return &BaseEvent{
		eventType: eventType,
		timestamp: time.Now(),
		payload:   make(map[string]interface{}),
	}
}

func (e *BaseEvent) Type() string {
	return e.eventType
}

func (e *BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

func (e *BaseEvent) Payload() map[string]interface{} {
	return e.payload
}

// SetPayload устанавливает дополнительные данные события
func (e *BaseEvent) SetPayload(key string, value interface{}) {
	e.payload[key] = value
}

// GetPayload возвращает дополнительные данные события
func (e *BaseEvent) GetPayload(key string) (interface{}, bool) {
	val, ok := e.payload[key]
	return val, ok
}

// ============================================================================
// Scanner Events
// ============================================================================

// ScanStartedEvent — начало сканирования
type ScanStartedEvent struct {
	*BaseEvent
	Network string
	Timeout string
}

// NewScanStartedEvent создает событие начала сканирования
func NewScanStartedEvent(network string, timeout string) *ScanStartedEvent {
	event := NewBaseEvent("scan.started")
	return &ScanStartedEvent{
		BaseEvent: event,
		Network:   network,
		Timeout:   timeout,
	}
}

// ScanCompletedEvent — завершение сканирования
type ScanCompletedEvent struct {
	*BaseEvent
	HostCount int
	OpenPorts int
	Duration  string
}

// NewScanCompletedEvent создает событие завершения сканирования
func NewScanCompletedEvent(hostCount int, openPorts int, duration string) *ScanCompletedEvent {
	event := NewBaseEvent("scan.completed")
	return &ScanCompletedEvent{
		BaseEvent: event,
		HostCount: hostCount,
		OpenPorts: openPorts,
		Duration:  duration,
	}
}

// HostFoundEvent — найден новый хост
type HostFoundEvent struct {
	*BaseEvent
	IP       string
	MAC      string
	Hostname string
}

// NewHostFoundEvent создает событие обнаружения хоста
func NewHostFoundEvent(ip string, mac string, hostname string) *HostFoundEvent {
	event := NewBaseEvent("host.found")
	return &HostFoundEvent{
		BaseEvent: event,
		IP:        ip,
		MAC:       mac,
		Hostname:  hostname,
	}
}

// PortOpenedEvent — открыт порт
type PortOpenedEvent struct {
	*BaseEvent
	IP       string
	Port     int
	Protocol string
	Service  string
}

// NewPortOpenedEvent создает событие открытия порта
func NewPortOpenedEvent(ip string, port int, protocol string, service string) *PortOpenedEvent {
	event := NewBaseEvent("port.opened")
	return &PortOpenedEvent{
		BaseEvent: event,
		IP:        ip,
		Port:      port,
		Protocol:  protocol,
		Service:   service,
	}
}

// DeviceDetectedEvent — обнаружено устройство
type DeviceDetectedEvent struct {
	*BaseEvent
	IP           string
	DeviceType   string
	DeviceVendor string
	OS           string
}

// NewDeviceDetectedEvent создает событие обнаружения устройства
func NewDeviceDetectedEvent(ip, deviceType, deviceVendor, os string) *DeviceDetectedEvent {
	event := NewBaseEvent("device.detected")
	return &DeviceDetectedEvent{
		BaseEvent:    event,
		IP:           ip,
		DeviceType:   deviceType,
		DeviceVendor: deviceVendor,
		OS:           os,
	}
}

// ============================================================================
// Plugin Events
// ============================================================================

// ProbeStartedEvent — начало выполнения probe
type ProbeStartedEvent struct {
	*BaseEvent
	ProbeName string
	Host      string
	Port      int
}

// NewProbeStartedEvent создает событие начала probe
func NewProbeStartedEvent(probeName string, host string, port int) *ProbeStartedEvent {
	event := NewBaseEvent("probe.started")
	return &ProbeStartedEvent{
		BaseEvent: event,
		ProbeName: probeName,
		Host:      host,
		Port:      port,
	}
}

// ProbeCompletedEvent — завершение probe
type ProbeCompletedEvent struct {
	*BaseEvent
	ProbeName string
	Host      string
	Port      int
	Result    string
}

// NewProbeCompletedEvent создает событие завершения probe
func NewProbeCompletedEvent(probeName string, host string, port int, result string) *ProbeCompletedEvent {
	event := NewBaseEvent("probe.completed")
	return &ProbeCompletedEvent{
		BaseEvent: event,
		ProbeName: probeName,
		Host:      host,
		Port:      port,
		Result:    result,
	}
}

// ProbeFailedEvent — ошибка probe
type ProbeFailedEvent struct {
	*BaseEvent
	ProbeName string
	Host      string
	Port      int
	Error     string
}

// NewProbeFailedEvent создает событие ошибки probe
func NewProbeFailedEvent(probeName string, host string, port int, err string) *ProbeFailedEvent {
	event := NewBaseEvent("probe.failed")
	return &ProbeFailedEvent{
		BaseEvent: event,
		ProbeName: probeName,
		Host:      host,
		Port:      port,
		Error:     err,
	}
}

// ============================================================================
// GUI Events
// ============================================================================

// ProgressUpdatedEvent — обновление прогресса
type ProgressUpdatedEvent struct {
	*BaseEvent
	Stage   string
	Current int
	Total   int
	Message string
}

// NewProgressUpdatedEvent создает событие обновления прогресса
func NewProgressUpdatedEvent(stage string, current int, total int, message string) *ProgressUpdatedEvent {
	event := NewBaseEvent("progress.updated")
	return &ProgressUpdatedEvent{
		BaseEvent: event,
		Stage:     stage,
		Current:   current,
		Total:     total,
		Message:   message,
	}
}

// ThemeChangedEvent — изменение темы
type ThemeChangedEvent struct {
	*BaseEvent
	Mode string
}

// NewThemeChangedEvent создает событие изменения темы
func NewThemeChangedEvent(mode string) *ThemeChangedEvent {
	event := NewBaseEvent("theme.changed")
	return &ThemeChangedEvent{
		BaseEvent: event,
		Mode:      mode,
	}
}

// SettingsChangedEvent — изменение настроек
type SettingsChangedEvent struct {
	*BaseEvent
	Key   string
	Value interface{}
}

// NewSettingsChangedEvent создает событие изменения настроек
func NewSettingsChangedEvent(key string, value interface{}) *SettingsChangedEvent {
	event := NewBaseEvent("settings.changed")
	return &SettingsChangedEvent{
		BaseEvent: event,
		Key:       key,
		Value:     value,
	}
}
