package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// MetricType тип метрики в формате Prometheus Exposition Format.
type MetricType string

const (
	MetricCounter   MetricType = "counter"
	MetricGauge     MetricType = "gauge"
	MetricHistogram MetricType = "histogram"
)

// Metric описывает одну метрику.
type Metric struct {
	Name   string
	Type   MetricType
	Value  float64
	Labels map[string]string
	Help   string
}

// Registry хранит все зарегистрированные метрики.
type Registry struct {
	mu         sync.RWMutex
	metrics    map[string]*Metric
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
}

// Counter счётчик (может только расти).
type Counter struct {
	mu     sync.Mutex
	val    float64
	help   string
	labels map[string]string
}

// Gauge измеритель (может расти и падать).
type Gauge struct {
	mu     sync.Mutex
	val    float64
	help   string
	labels map[string]string
}

// Histogram распределение значений с бакетами (Prometheus _bucket/_sum/_count).
type Histogram struct {
	mu      sync.Mutex
	buckets []float64 // верхние границы, возрастающие
	counts  []uint64  // накопленные счётчики по бакетам
	sum     float64
	count   uint64
	help    string
	labels  map[string]string
}

// DefaultHistogramBuckets стандартные границы для длительностей (секунды).
var DefaultHistogramBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// NewRegistry создаёт новый реестр метрик.
func NewRegistry() *Registry {
	return &Registry{
		metrics:    make(map[string]*Metric),
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
	}
}

// NewCounter создаёт или возвращает существующий счётчик.
func (r *Registry) NewCounter(name, help string, labels map[string]string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := metricKey(name, labels)
	if c, ok := r.counters[key]; ok {
		return c
	}
	c := &Counter{help: help, labels: labels}
	r.counters[key] = c
	r.metrics[key] = &Metric{Name: name, Type: MetricCounter, Labels: labels, Help: help}
	return c
}

// NewGauge создаёт или возвращает существующий измеритель.
func (r *Registry) NewGauge(name, help string, labels map[string]string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := metricKey(name, labels)
	if g, ok := r.gauges[key]; ok {
		return g
	}
	g := &Gauge{help: help, labels: labels}
	r.gauges[key] = g
	r.metrics[key] = &Metric{Name: name, Type: MetricGauge, Labels: labels, Help: help}
	return g
}

// NewHistogram создаёт или возвращает существующую гистограмму.
// buckets — верхние границы бакетов; сортируются и дедуплицируются автоматически.
func (r *Registry) NewHistogram(name, help string, buckets []float64, labels map[string]string) *Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := metricKey(name, labels)
	if h, ok := r.histograms[key]; ok {
		return h
	}
	bs := sortDedup(buckets)
	h := &Histogram{buckets: bs, counts: make([]uint64, len(bs)), help: help, labels: labels}
	r.histograms[key] = h
	r.metrics[key] = &Metric{Name: name, Type: MetricHistogram, Labels: labels, Help: help}
	return h
}

// Inc увеличивает счётчик на 1.
func (c *Counter) Inc() {
	c.mu.Lock()
	c.val++
	c.mu.Unlock()
}

// Add увеличивает счётчик на значение.
func (c *Counter) Add(v float64) {
	c.mu.Lock()
	c.val += v
	c.mu.Unlock()
}

// Value возвращает текущее значение счётчика.
func (c *Counter) Value() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.val
}

// Set устанавливает значение измерителя.
func (g *Gauge) Set(v float64) {
	g.mu.Lock()
	g.val = v
	g.mu.Unlock()
}

// Inc увеличивает измеритель на 1.
func (g *Gauge) Inc() {
	g.mu.Lock()
	g.val++
	g.mu.Unlock()
}

// Dec уменьшает измеритель на 1.
func (g *Gauge) Dec() {
	g.mu.Lock()
	g.val--
	g.mu.Unlock()
}

// Value возвращает текущее значение измерителя.
func (g *Gauge) Value() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.val
}

// Observe добавляет наблюдение в гистограмму.
func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sum += v
	h.count++
	for i, b := range h.buckets {
		if v <= b {
			h.counts[i]++
		}
	}
}

// Snapshot возвращает копию состояния гистограммы для вывода.
func (h *Histogram) Snapshot() (buckets []float64, counts []uint64, sum float64, count uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	b := make([]float64, len(h.buckets))
	copy(b, h.buckets)
	c := make([]uint64, len(h.counts))
	copy(c, h.counts)
	return b, c, h.sum, h.count
}

// metricKey формирует ключ для хранения метрики.
func metricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	return name + "{" + labelString(labels) + "}"
}

// labelString форматирует метки в строку Prometheus.
func labelString(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := ""
	for i, k := range keys {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf(`%s="%s"`, k, escapeLabel(labels[k]))
	}
	return out
}

// escapeLabel экранирует значение метки для Prometheus.
func escapeLabel(v string) string {
	out := ""
	for _, ch := range v {
		switch ch {
		case '\\':
			out += `\\`
		case '"':
			out += `\"`
		case '\n':
			out += `\n`
		default:
			out += string(ch)
		}
	}
	return out
}

// sortDedup сортирует и дедуплицирует границы бакетов по возрастанию.
func sortDedup(in []float64) []float64 {
	cp := make([]float64, len(in))
	copy(cp, in)
	sort.Float64s(cp)
	out := cp[:0]
	for i, v := range cp {
		if i == 0 || v != cp[i-1] {
			out = append(out, v)
		}
	}
	return out
}

// Gather собирает все метрики в формате Prometheus Exposition Format.
func (r *Registry) Gather() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Metric, 0, len(r.metrics))
	for _, m := range r.metrics {
		out = append(out, *m)
	}
	return out
}

// WriteTo выводит метрики в формате Prometheus Exposition Format.
func (r *Registry) WriteTo(w http.ResponseWriter) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// HELP и TYPE (один раз на имя метрики).
	seen := make(map[string]bool)
	for _, m := range r.metrics {
		if !seen[m.Name] {
			fmt.Fprintf(w, "# HELP %s %s\n", m.Name, m.Help)
			fmt.Fprintf(w, "# TYPE %s %s\n", m.Name, string(m.Type))
			seen[m.Name] = true
		}
	}

	// Counter и Gauge значения.
	for _, m := range r.metrics {
		val := 0.0
		key := metricKey(m.Name, m.Labels)
		switch m.Type {
		case MetricCounter:
			if c, ok := r.counters[key]; ok {
				val = c.Value()
			}
		case MetricGauge:
			if g, ok := r.gauges[key]; ok {
				val = g.Value()
			}
		}
		fmt.Fprintf(w, "%s{%s} %v\n", m.Name, labelString(m.Labels), val)
	}

	// Histogram _bucket, _sum, _count.
	for _, m := range r.metrics {
		if m.Type != MetricHistogram {
			continue
		}
		key := metricKey(m.Name, m.Labels)
		if h, ok := r.histograms[key]; ok {
			bs, cs, sum, cnt := h.Snapshot()
			labels := labelString(m.Labels)
			for i, b := range bs {
				if labels != "" {
					fmt.Fprintf(w, "%s_bucket{le=\"%s\",%s} %d\n", m.Name, formatBucket(b), labels, cs[i])
				} else {
					fmt.Fprintf(w, "%s_bucket{le=\"%s\"} %d\n", m.Name, formatBucket(b), cs[i])
				}
			}
			if labels != "" {
				fmt.Fprintf(w, "%s_bucket{le=\"+Inf\",%s} %d\n", m.Name, labels, cnt)
				fmt.Fprintf(w, "%s_sum{%s} %v\n", m.Name, labels, sum)
				fmt.Fprintf(w, "%s_count{%s} %d\n", m.Name, labels, cnt)
			} else {
				fmt.Fprintf(w, "%s_bucket{le=\"+Inf\"} %d\n", m.Name, cnt)
				fmt.Fprintf(w, "%s_sum{} %v\n", m.Name, sum)
				fmt.Fprintf(w, "%s_count{} %d\n", m.Name, cnt)
			}
		}
	}
}

// formatBucket форматирует верхнюю границу бакета.
func formatBucket(b float64) string {
	if b == 0 {
		return "0"
	}
	s := fmt.Sprintf("%g", b)
	return s
}

// Global — глобальный реестр метрик приложения.
var Global = NewRegistry()

// InitMetrics регистрирует базовые метрики приложения.
func InitMetrics() {
	Global.NewCounter("network_scanner_scans_total", "Общее количество запусков сканирования", nil)
	Global.NewCounter("network_scanner_scans_completed_total", "Количество успешно завершённых сканирований", nil)
	Global.NewCounter("network_scanner_scans_failed_total", "Количество неудачных сканирований", nil)
	Global.NewCounter("network_scanner_hosts_found_total", "Общее количество найденных хостов", nil)
	Global.NewCounter("network_scanner_ports_open_total", "Общее количество открытых портов", nil)
	Global.NewGauge("network_scanner_active_scans", "Количество активных сканирований", nil)
	Global.NewCounter("network_scanner_http_requests_total", "HTTP запросы", nil)
	Global.NewCounter("network_scanner_http_request_duration_seconds_sum", "Суммарное время HTTP-запросов", nil)
	Global.NewCounter("network_scanner_http_request_duration_seconds_count", "Количество HTTP-запросов для вычисления средней длительности", nil)
	Global.NewHistogram("network_scanner_http_request_duration_seconds", "Длительность HTTP-запросов (сек)", DefaultHistogramBuckets, nil)
}

// HTTPRequestStarted фиксирует начало HTTP-запроса.
func HTTPRequestStarted(method, path string) {
	c := Global.NewCounter("network_scanner_http_requests_total", "HTTP запросы", map[string]string{"method": method, "path": path})
	c.Inc()
}

// HTTPRequestFinished фиксирует завершение HTTP-запроса с длительностью.
func HTTPRequestFinished(method, path string, duration time.Duration) {
	c := Global.NewCounter("network_scanner_http_requests_total", "HTTP запросы", map[string]string{"method": method, "path": path})
	c.Inc()
	sum := Global.NewCounter("network_scanner_http_request_duration_seconds_sum", "Суммарное время HTTP-запросов", map[string]string{"method": method, "path": path})
	sum.Add(duration.Seconds())
	cnt := Global.NewCounter("network_scanner_http_request_duration_seconds_count", "Количество HTTP-запросов", map[string]string{"method": method, "path": path})
	cnt.Inc()
	h := Global.NewHistogram("network_scanner_http_request_duration_seconds", "Длительность HTTP-запросов (сек)", DefaultHistogramBuckets, map[string]string{"method": method, "path": path})
	h.Observe(duration.Seconds())
}

// Handler — HTTP middleware для сбора метрик запросов.
type Handler struct {
	next http.Handler
}

// NewHandler создаёт middleware для сбора метрик HTTP-запросов.
func NewHandler(next http.Handler) *Handler {
	return &Handler{next: next}
}

// ServeHTTP оборачивает следующий handler, собирая метрики.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	wrapped := &statusRecorder{ResponseWriter: w, status: 200}
	h.next.ServeHTTP(wrapped, r)
	duration := time.Since(start)
	HTTPRequestFinished(r.Method, r.URL.Path, duration)
}

// statusRecorder отслеживает HTTP статус-код.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader перехватывает установку статус-кода.
func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}
