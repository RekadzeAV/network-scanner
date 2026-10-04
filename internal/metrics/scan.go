package metrics

import (
	"strconv"
	"time"

	"network-scanner/internal/eventbus"
)

// Имена метрик сканирования (E7/7.8).
const (
	MetricScanTotal     = "network_scanner_scan_total"
	MetricScanFailures  = "network_scanner_scan_failures_total"
	MetricScanActive    = "network_scanner_scan_active"
	MetricScanHosts     = "network_scanner_scan_hosts_total"
	MetricScanOpenPorts = "network_scanner_scan_open_ports_total"
	MetricScanDuration  = "network_scanner_scan_duration_seconds"
	MetricHTTPServerUp  = "network_scanner_metrics_server_up"
	MetricScansRunning  = "network_scanner_scans_running"
)

// ScanMetrics — набор метрик сканирования, обновляемых из eventbus.
type ScanMetrics struct {
	Scans       *Counter
	Failures    *Counter
	Hosts       *Counter
	OpenPorts   *Counter
	Active      *Gauge
	Duration    *Summary
	Registry    *Registry
	subscribed  bool
	lastStarted time.Time
}

// NewScanMetrics создаёт и регистрирует метрики сканирования в реестре.
func NewScanMetrics(reg *Registry) *ScanMetrics {
	return &ScanMetrics{
		Scans:     reg.NewCounter(MetricScanTotal, "Всего запусков сканирования"),
		Failures:  reg.NewCounter(MetricScanFailures, "Отменённые/неуспешные сканирования"),
		Hosts:     reg.NewCounter(MetricScanHosts, "Найдено хостов (суммарно)"),
		OpenPorts: reg.NewCounter(MetricScanOpenPorts, "Открытых портов (суммарно)"),
		Active:    reg.NewGauge(MetricScanActive, "Активных сканирований сейчас"),
		Duration:  reg.NewSummary(MetricScanDuration, "Длительность сканирования (секунды)"),
		Registry:  reg,
	}
}

// Register подписывает метрики на события шины (scan.started/completed/failed).
//
// Публикация событий opt-in (E6): если шина не подключена к сканеру, метрики
// остаются нулевыми и ни на что не влияют.
func (m *ScanMetrics) Register(bus *eventbus.EventBus) {
	if bus == nil || m.subscribed {
		return
	}
	bus.Subscribe("scan.started", func(e eventbus.Event) {
		m.Scans.Inc()
		m.Active.Inc()
		m.lastStarted = e.Timestamp()
	})
	bus.Subscribe("scan.completed", func(e eventbus.Event) {
		m.Active.Dec()
		// Событие предоставляет типизированные поля (E6), но подписчик может
		// получить и базовое событие — тогда читаем payload.
		if completed, ok := e.(*eventbus.ScanCompletedEvent); ok {
			m.Hosts.Add(float64(completed.HostCount))
			m.OpenPorts.Add(float64(completed.OpenPorts))
			m.Duration.Observe(parseDurationSeconds(completed.Duration))
			return
		}
		payload := e.Payload()
		m.Hosts.Add(payloadFloat(payload, "host_count"))
		m.OpenPorts.Add(payloadFloat(payload, "open_ports"))
		m.Duration.Observe(parseDurationSeconds(payload["duration"]))
	})
	bus.Subscribe("scan.failed", func(eventbus.Event) {
		m.Active.Dec()
		m.Failures.Inc()
	})
	m.subscribed = true
}

// payloadFloat извлекает числовое значение из payload события.
func payloadFloat(payload map[string]interface{}, key string) float64 {
	if payload == nil {
		return 0
	}
	switch v := payload[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

// parseDurationSeconds разбирает длительность из payload (формат Go или число).
func parseDurationSeconds(v interface{}) float64 {
	switch d := v.(type) {
	case time.Duration:
		return d.Seconds()
	case string:
		if parsed, err := time.ParseDuration(d); err == nil {
			return parsed.Seconds()
		}
		if f, err := strconv.ParseFloat(d, 64); err == nil {
			return f
		}
	case float64:
		return d
	case int64:
		return float64(d)
	}
	return 0
}
