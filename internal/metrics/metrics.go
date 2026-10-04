// Package metrics предоставляет минимальный потокобезопасный реестр метрик
// и HTTP-эндпоинт экспорта в текстовом формате Prometheus
// (text/plain; version=0.0.4).
//
// Пакет намеренно не тянет внешних зависимостей (prometheus/client_golang):
// текстовая экспозиция стандартизирована и без библиотеки, а репозиторий
// придерживается правила «только зависимости из go.mod».
//
// Пример:
//
//	reg := metrics.NewRegistry()
//	scans := reg.NewCounter("network_scanner_scan_total", "Всего сканирований")
//	scans.Inc()
//	http.Handle("/metrics", reg.Handler())
package metrics

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Registry — набор метрик с конкурентным доступом.
type Registry struct {
	mu        sync.RWMutex
	counters  map[string]*Counter
	gauges    map[string]*Gauge
	summaries map[string]*Summary
}

// NewRegistry создаёт пустой реестр.
func NewRegistry() *Registry {
	return &Registry{
		counters:  make(map[string]*Counter),
		gauges:    make(map[string]*Gauge),
		summaries: make(map[string]*Summary),
	}
}

// NewCounter возвращает счётчик (создаётся при первом обращении).
func (r *Registry) NewCounter(name, help string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &Counter{name: name, help: help}
	r.counters[name] = c
	return c
}

// NewGauge возвращает gauge-метрику.
func (r *Registry) NewGauge(name, help string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.gauges[name]; ok {
		return g
	}
	g := &Gauge{name: name, help: help}
	r.gauges[name] = g
	return g
}

// NewSummary возвращает summary-метрику (сумма + количество).
func (r *Registry) NewSummary(name, help string) *Summary {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.summaries[name]; ok {
		return s
	}
	s := &Summary{name: name, help: help}
	r.summaries[name] = s
	return s
}

// Counter — монотонно возрастающее значение.
type Counter struct {
	name  string
	help  string
	mu    sync.Mutex
	value float64
}

// Inc увеличивает счётчик на 1.
func (c *Counter) Inc() { c.Add(1) }

// Add увеличивает счётчик на delta (отрицательные игнорируются).
func (c *Counter) Add(delta float64) {
	if delta < 0 {
		return
	}
	c.mu.Lock()
	c.value += delta
	c.mu.Unlock()
}

// Value возвращает текущее значение.
func (c *Counter) Value() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Name возвращает имя метрики.
func (c *Counter) Name() string { return c.name }

// Gauge — значение, которое может возрастать и убывать.
type Gauge struct {
	name  string
	help  string
	mu    sync.Mutex
	value float64
}

// Set устанавливает значение.
func (g *Gauge) Set(v float64) {
	g.mu.Lock()
	g.value = v
	g.mu.Unlock()
}

// Inc увеличивает на 1.
func (g *Gauge) Inc() { g.Add(1) }

// Dec уменьшает на 1.
func (g *Gauge) Dec() { g.Add(-1) }

// Add изменяет значение на delta.
func (g *Gauge) Add(delta float64) {
	g.mu.Lock()
	g.value += delta
	g.mu.Unlock()
}

// Value возвращает текущее значение.
func (g *Gauge) Value() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

// Name возвращает имя метрики.
func (g *Gauge) Name() string { return g.name }

// Summary — распределение наблюдений (количество + сумма).
type Summary struct {
	name  string
	help  string
	mu    sync.Mutex
	count uint64
	sum   float64
}

// Observe добавляет наблюдение (например, длительность в секундах).
func (s *Summary) Observe(v float64) {
	s.mu.Lock()
	s.count++
	s.sum += v
	s.mu.Unlock()
}

// ObserveDuration добавляет длительность в секундах.
func (s *Summary) ObserveDuration(d time.Duration) {
	s.Observe(d.Seconds())
}

// Count возвращает число наблюдений.
func (s *Summary) Count() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

// Sum возвращает сумму наблюдений.
func (s *Summary) Sum() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sum
}

// Name возвращает имя метрики.
func (s *Summary) Name() string { return s.name }

// Render формирует текстовую экспозицию Prometheus.
func (r *Registry) Render() string {
	r.mu.RLock()
	counters := make([]*Counter, 0, len(r.counters))
	for _, c := range r.counters {
		counters = append(counters, c)
	}
	gauges := make([]*Gauge, 0, len(r.gauges))
	for _, g := range r.gauges {
		gauges = append(gauges, g)
	}
	summaries := make([]*Summary, 0, len(r.summaries))
	for _, s := range r.summaries {
		summaries = append(summaries, s)
	}
	r.mu.RUnlock()

	sort.Slice(counters, func(i, j int) bool { return counters[i].name < counters[j].name })
	sort.Slice(gauges, func(i, j int) bool { return gauges[i].name < gauges[j].name })
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].name < summaries[j].name })

	var b strings.Builder
	for _, c := range counters {
		writeHelp(&b, c.name, c.help, "counter")
		fmt.Fprintf(&b, "%s %s\n", c.name, formatFloat(c.Value()))
	}
	for _, g := range gauges {
		writeHelp(&b, g.name, g.help, "gauge")
		fmt.Fprintf(&b, "%s %s\n", g.name, formatFloat(g.Value()))
	}
	for _, s := range summaries {
		writeHelp(&b, s.name, s.help, "summary")
		fmt.Fprintf(&b, "%s_sum %s\n", s.name, formatFloat(s.Sum()))
		fmt.Fprintf(&b, "%s_count %d\n", s.name, s.Count())
	}
	return b.String()
}

// Handler возвращает http.Handler с экспозицией в формате Prometheus.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(r.Render()))
	})
}

// writeHelp добавляет строки # HELP / # TYPE, если описание задано.
func writeHelp(b *strings.Builder, name, help, typ string) {
	if strings.TrimSpace(help) != "" {
		fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	}
	fmt.Fprintf(b, "# TYPE %s %s\n", name, typ)
}

// formatFloat форматирует значение без экспоненты и без лишних нулей.
func formatFloat(v float64) string {
	if math.IsInf(v, 1) {
		return "+Inf"
	}
	if math.IsInf(v, -1) {
		return "-Inf"
	}
	if math.IsNaN(v) {
		return "NaN"
	}
	return fmt.Sprintf("%g", v)
}
