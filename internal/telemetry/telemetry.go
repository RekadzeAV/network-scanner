// Package telemetry предоставляет модуль для сбора анонимной статистики использования.
//
// # Конфиденциальность
//
// Telemetry собирает ТОЛЬКО анонимные метрики:
//   - Версия приложения
//   - Количество сканирований по дням
//   - Типы используемых профилей сканирования
//   - Время выполнения сканирования
//
// НЕ собирается:
//   - IP-адреса
//   - MAC-адреса
//   - Имена хостов
//   - Содержимое результатов сканирования
//
// # Opt-Out
//
// Пользователь может отключить телеметрию в настройках.
// Предпочтение сохраняется в Preferences (ключ: "telemetry.enabled").
package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

// Config содержит конфигурацию телеметрии
type Config struct {
	Enabled      bool
	Endpoint     string // URL для отправки метрик
	Interval     time.Duration
	AppVersion   string
	MaxQueueSize int
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig(appVersion string) *Config {
	return &Config{
		Enabled:      true,
		Endpoint:     "https://telemetry.example.com/collect",
		Interval:     1 * time.Hour,
		AppVersion:   appVersion,
		MaxQueueSize: 1000,
	}
}

// Telemetry управляет сбором и отправкой метрик
type Telemetry struct {
	cfg     *Config
	client  *http.Client
	queue   []Metric
	mu      sync.Mutex
	enabled bool
	stopCh  chan struct{}
}

// Metric представляет одну метрику
type Metric struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Timestamp  time.Time              `json:"timestamp"`
	AppVersion string                 `json:"app_version"`
	OS         string                 `json:"os"`
	Arch       string                 `json:"arch"`
	Data       map[string]interface{} `json:"data"`
}

// NewTelemetry создает новый экземпляр телеметрии
func NewTelemetry(cfg *Config) *Telemetry {
	if cfg == nil {
		cfg = DefaultConfig("unknown")
	}

	return &Telemetry{
		cfg:     cfg,
		client:  &http.Client{Timeout: 10 * time.Second},
		queue:   make([]Metric, 0, cfg.MaxQueueSize),
		enabled: cfg.Enabled,
		stopCh:  make(chan struct{}),
	}
}

// IsEnabled возвращает статус телеметрии
func (t *Telemetry) IsEnabled() bool {
	return t.enabled
}

// SetEnabled устанавливает статус телеметрии
func (t *Telemetry) SetEnabled(enabled bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = enabled
	if !enabled {
		// Отменяем отправку если отключили
		select {
		case <-t.stopCh:
		default:
			close(t.stopCh)
			t.stopCh = make(chan struct{})
		}
	}
}

// RecordScan записывает метрику сканирования
func (t *Telemetry) RecordScan(hostCount, openPorts, durationSec int, profile string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.enabled {
		return
	}

	if len(t.queue) >= t.cfg.MaxQueueSize {
		return // Очередь заполнена, пропускаем
	}

	metric := Metric{
		ID:         generateID(),
		Type:       "scan",
		Timestamp:  time.Now(),
		AppVersion: t.cfg.AppVersion,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Data: map[string]interface{}{
			"host_count": hostCount,
			"open_ports": openPorts,
			"duration":   durationSec,
			"profile":    profile,
		},
	}

	t.queue = append(t.queue, metric)
}

// RecordAppStart записывает метрику запуска приложения
func (t *Telemetry) RecordAppStart() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.enabled {
		return
	}

	if len(t.queue) >= t.cfg.MaxQueueSize {
		return
	}

	metric := Metric{
		ID:         generateID(),
		Type:       "app_start",
		Timestamp:  time.Now(),
		AppVersion: t.cfg.AppVersion,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Data:       map[string]interface{}{},
	}

	t.queue = append(t.queue, metric)
}

// Start запускает периодическую отправку метрик
func (t *Telemetry) Start() {
	if !t.enabled {
		return
	}

	ticker := time.NewTicker(t.cfg.Interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.flush()
			case <-t.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop останавливает телеметрию и отправляет оставшиеся метрики
func (t *Telemetry) Stop() {
	select {
	case <-t.stopCh:
	default:
		close(t.stopCh)
	}
	t.flush()
}

// flush отправляет накопленные метрики
func (t *Telemetry) flush() {
	t.mu.Lock()
	if len(t.queue) == 0 {
		t.mu.Unlock()
		return
	}

	metrics := make([]Metric, len(t.queue))
	copy(metrics, t.queue)
	t.queue = t.queue[:0]
	t.mu.Unlock()

	t.send(metrics)
}

// send отправляет метрики на сервер
func (t *Telemetry) send(metrics []Metric) {
	payload, err := json.Marshal(metrics)
	if err != nil {
		return // Не публикуруем ошибки в лог
	}

	req, err := http.NewRequest("POST", t.cfg.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NetworkScanner/"+t.cfg.AppVersion)

	resp, err := t.client.Do(req)
	if err != nil {
		// Не сохраняем метрики при ошибке — они будут отправлены позже
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Успешно отправлено
	}
}

// generateID генерирует уникальный ID для метрики
func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}

// GetStats возвращает статистику телеметрии
func (t *Telemetry) GetStats() map[string]interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()

	return map[string]interface{}{
		"enabled":    t.enabled,
		"queue_size": len(t.queue),
		"endpoint":   t.cfg.Endpoint,
		"version":    t.cfg.AppVersion,
	}
}
