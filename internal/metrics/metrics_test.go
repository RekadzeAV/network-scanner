package metrics

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCounter_IncAddValue(t *testing.T) {
	reg := NewRegistry()
	c := reg.NewCounter("test_total", "тестовый счётчик")
	c.Inc()
	c.Add(4)
	if got := c.Value(); got != 5 {
		t.Fatalf("Value() = %v, want 5", got)
	}
	// Отрицательные значения игнорируются (counter монотонен).
	c.Add(-10)
	if got := c.Value(); got != 5 {
		t.Fatalf("Value() after negative Add = %v, want 5", got)
	}
	// Повторное получение возвращает тот же счётчик.
	if reg.NewCounter("test_total", "другое описание") != c {
		t.Fatal("NewCounter должен возвращать существующий счётчик")
	}
}

func TestGauge_SetIncDec(t *testing.T) {
	g := NewRegistry().NewGauge("test_active", "активные")
	g.Set(2)
	g.Inc()
	g.Dec()
	g.Dec()
	if got := g.Value(); got != 1 {
		t.Fatalf("Value() = %v, want 1", got)
	}
}

func TestSummary_ObserveAndDuration(t *testing.T) {
	s := NewRegistry().NewSummary("test_duration_seconds", "длительность")
	s.Observe(0.5)
	s.ObserveDuration(1500 * time.Millisecond)
	if s.Count() != 2 {
		t.Fatalf("Count() = %d, want 2", s.Count())
	}
	if sum := s.Sum(); sum != 2.0 {
		t.Fatalf("Sum() = %v, want 2.0", sum)
	}
}

func TestRender_PrometheusFormat(t *testing.T) {
	reg := NewRegistry()
	reg.NewCounter("b_total", "счётчик b").Inc()
	reg.NewCounter("a_total", "счётчик a").Add(3)
	g := reg.NewGauge("g_current", "")
	g.Set(7)
	reg.NewSummary("d_seconds", "длительность").Observe(2)

	out := reg.Render()

	// Типы и значения.
	for _, want := range []string{
		"# TYPE a_total counter",
		"a_total 3",
		"b_total 1",
		"# TYPE g_current gauge",
		"g_current 7",
		"# TYPE d_seconds summary",
		"d_seconds_sum 2",
		"d_seconds_count 1",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() missing %q\noutput:\n%s", want, out)
		}
	}
	// Метрики отсортированы по имени (a раньше b).
	if strings.Index(out, "a_total") > strings.Index(out, "b_total") {
		t.Errorf("метрики должны быть отсортированы:\n%s", out)
	}
	// HELP добавляется только при непустом описании.
	if strings.Contains(out, "# HELP g_current") {
		t.Errorf("пустое HELP не должно попадать в вывод:\n%s", out)
	}
}

func TestHandler_ContentTypeAndBody(t *testing.T) {
	reg := NewRegistry()
	reg.NewCounter("hits_total", "хиты").Inc()

	rec := httptest.NewRecorder()
	reg.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") || !strings.Contains(ct, "version=0.0.4") {
		t.Errorf("Content-Type = %q, want Prometheus text format", ct)
	}
	if !strings.Contains(rec.Body.String(), "hits_total 1") {
		t.Errorf("body = %q, want hits_total 1", rec.Body.String())
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewRegistry()
	c := reg.NewCounter("conc_total", "конкурентный")
	g := reg.NewGauge("conc_gauge", "конкурентный gauge")
	s := reg.NewSummary("conc_seconds", "конкурентная длительность")

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.Inc()
				g.Inc()
				if j%2 == 0 {
					g.Dec()
				}
				s.Observe(0.001)
				_ = reg.Render()
			}
		}()
	}
	wg.Wait()

	if got := c.Value(); got != 1600 {
		t.Errorf("counter = %v, want 1600", got)
	}
	if got := s.Count(); got != 1600 {
		t.Errorf("summary count = %d, want 1600", got)
	}
}
