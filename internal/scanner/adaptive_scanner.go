package scanner

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// AdaptiveConfig конфигурация адаптивного сканирования.
type AdaptiveConfig struct {
	// MinBudget минимальный budget probe (не может быть меньше)
	MinBudget int
	// MaxBudget максимальный budget probe (не может быть больше)
	MaxBudget int
	// InitialBudget начальный budget
	InitialBudget int
	// SpeedThreshold порог скорости (мс на probe) для адаптации
	SpeedThreshold time.Duration
	// ErrorThreshold порог ошибок (процент) для снижения budget
	ErrorThreshold float64
	// AdaptInterval интервал адаптации
	AdaptInterval time.Duration
}

// DefaultAdaptiveConfig возвращает конфигурацию по умолчанию.
func DefaultAdaptiveConfig() AdaptiveConfig {
	return AdaptiveConfig{
		MinBudget:      64,
		MaxBudget:      1024,
		InitialBudget:  512,
		SpeedThreshold: 50 * time.Millisecond,
		ErrorThreshold: 0.3, // 30% ошибок
		AdaptInterval:  1 * time.Second,
	}
}

// AdaptiveScanner обёртка над NetworkScanner с адаптивными лимитами.
type AdaptiveScanner struct {
	config        AdaptiveConfig
	scanner       *NetworkScanner
	metrics       *ScanMetrics
	adaptMu       sync.Mutex
	currentBudget int64
}

// ScanMetrics метрики сканирования для адаптации.
type ScanMetrics struct {
	probesTotal      int64
	probesOpen       int64
	probesClosed     int64
	probesError      int64
	startTime        time.Time
	lastAdaptTime    time.Time
	probesSinceAdapt int64
	openSinceAdapt   int64
}

// NewAdaptiveScanner создаёт новый адаптивный сканер.
func NewAdaptiveScanner(ns *NetworkScanner, config AdaptiveConfig) *AdaptiveScanner {
	if config.MinBudget <= 0 {
		config.MinBudget = 64
	}
	if config.MaxBudget <= 0 {
		config.MaxBudget = 1024
	}
	if config.InitialBudget <= 0 {
		config.InitialBudget = 512
	}

	return &AdaptiveScanner{
		config:        config,
		scanner:       ns,
		metrics:       &ScanMetrics{startTime: time.Now()},
		currentBudget: int64(config.InitialBudget),
	}
}

// GetBudget возвращает текущий budget probe.
func (a *AdaptiveScanner) GetBudget() int {
	return int(atomic.LoadInt64(&a.currentBudget))
}

// SetBudget устанавливает новый budget probe.
func (a *AdaptiveScanner) SetBudget(budget int) {
	a.adaptMu.Lock()
	defer a.adaptMu.Unlock()

	if budget < a.config.MinBudget {
		budget = a.config.MinBudget
	}
	if budget > a.config.MaxBudget {
		budget = a.config.MaxBudget
	}

	atomic.StoreInt64(&a.currentBudget, int64(budget))
}

// RecordProbe записывает метрику probe.
func (a *AdaptiveScanner) RecordProbe(isOpen bool, isError bool) {
	atomic.AddInt64(&a.metrics.probesTotal, 1)
	atomic.AddInt64(&a.metrics.probesSinceAdapt, 1)

	if isError {
		atomic.AddInt64(&a.metrics.probesError, 1)
	} else if isOpen {
		atomic.AddInt64(&a.metrics.probesOpen, 1)
		atomic.AddInt64(&a.metrics.openSinceAdapt, 1)
	} else {
		atomic.AddInt64(&a.metrics.probesClosed, 1)
	}
}

// Adapt адаптирует budget на основе метрик.
func (a *AdaptiveScanner) Adapt() {
	a.adaptMu.Lock()
	defer a.adaptMu.Unlock()

	// Проверяем, нужно ли адаптировать
	if time.Since(a.metrics.lastAdaptTime) < a.config.AdaptInterval {
		return
	}

	probesSinceAdapt := atomic.LoadInt64(&a.metrics.probesSinceAdapt)
	if probesSinceAdapt < 10 { // Минимум 10 probe для адаптации
		return
	}

	openSinceAdapt := atomic.LoadInt64(&a.metrics.openSinceAdapt)
	errorSinceAdapt := atomic.LoadInt64(&a.metrics.probesError)

	// Рассчитываем процент ошибок
	errorRate := float64(errorSinceAdapt) / float64(probesSinceAdapt)

	// Рассчитываем процент открытых портов
	openRate := float64(openSinceAdapt) / float64(probesSinceAdapt)

	currentBudget := atomic.LoadInt64(&a.currentBudget)
	newBudget := currentBudget

	// Если высокий процент ошибок — снижаем budget
	if errorRate > a.config.ErrorThreshold {
		newBudget = int64(float64(currentBudget) * 0.7) // Снижаем на 30%
	} else if openRate > 0.5 && probesSinceAdapt > 50 {
		// Если много открытых портов и высокая нагрузка — немного снижаем
		newBudget = int64(float64(currentBudget) * 0.9)
	} else if openRate < 0.1 && probesSinceAdapt > 100 {
		// Если мало открытых портов и мало ошибок — увеличиваем budget
		newBudget = int64(float64(currentBudget) * 1.2) // Увеличиваем на 20%
	}

	// Применяем ограничения
	if newBudget < int64(a.config.MinBudget) {
		newBudget = int64(a.config.MinBudget)
	}
	if newBudget > int64(a.config.MaxBudget) {
		newBudget = int64(a.config.MaxBudget)
	}

	if newBudget != currentBudget {
		atomic.StoreInt64(&a.currentBudget, newBudget)
	}

	// Сбрасываем счётчики для следующего интервала
	atomic.StoreInt64(&a.metrics.probesSinceAdapt, 0)
	atomic.StoreInt64(&a.metrics.openSinceAdapt, 0)
	atomic.StoreInt64(&a.metrics.probesError, 0)

	a.metrics.lastAdaptTime = time.Now()
}

// GetMetrics возвращает текущие метрики.
func (a *AdaptiveScanner) GetMetrics() ScanMetrics {
	return ScanMetrics{
		probesTotal:      atomic.LoadInt64(&a.metrics.probesTotal),
		probesOpen:       atomic.LoadInt64(&a.metrics.probesOpen),
		probesClosed:     atomic.LoadInt64(&a.metrics.probesClosed),
		probesError:      atomic.LoadInt64(&a.metrics.probesError),
		startTime:        a.metrics.startTime,
		lastAdaptTime:    a.metrics.lastAdaptTime,
		probesSinceAdapt: atomic.LoadInt64(&a.metrics.probesSinceAdapt),
		openSinceAdapt:   atomic.LoadInt64(&a.metrics.openSinceAdapt),
	}
}

// GetOpenRate возвращает процент открытых портов.
func (a *AdaptiveScanner) GetOpenRate() float64 {
	total := atomic.LoadInt64(&a.metrics.probesTotal)
	if total == 0 {
		return 0
	}
	open := atomic.LoadInt64(&a.metrics.probesOpen)
	return float64(open) / float64(total)
}

// GetErrorRate возвращает процент ошибок.
func (a *AdaptiveScanner) GetErrorRate() float64 {
	total := atomic.LoadInt64(&a.metrics.probesTotal)
	if total == 0 {
		return 0
	}
	errors := atomic.LoadInt64(&a.metrics.probesError)
	return float64(errors) / float64(total)
}

// GetDuration возвращает время работы сканирования.
func (a *AdaptiveScanner) GetDuration() time.Duration {
	return time.Since(a.metrics.startTime)
}

// GetSummary возвращает сводку метрик.
func (a *AdaptiveScanner) GetSummary() string {
	metrics := a.GetMetrics()
	duration := time.Since(metrics.startTime)
	openRate := a.GetOpenRate()
	errorRate := a.GetErrorRate()
	budget := a.GetBudget()

	return fmt.Sprintf("Адаптивное сканирование:\n"+
		"  Budget: %d\n"+
		"  Probe всего: %d\n"+
		"  Open: %d (%.1f%%)\n"+
		"  Closed: %d\n"+
		"  Error: %d (%.1f%%)\n"+
		"  Duration: %v",
		budget,
		metrics.probesTotal,
		metrics.probesOpen, openRate*100,
		metrics.probesClosed,
		metrics.probesError, errorRate*100,
		duration)
}
