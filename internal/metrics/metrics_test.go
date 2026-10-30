package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- Counter ---

func TestCounter_Inc(t *testing.T) {
	c := &Counter{}
	c.Inc()
	c.Inc()
	if v := c.Value(); v != 2 {
		t.Errorf("Counter.Inc() = %v, want 2", v)
	}
}

func TestCounter_Add(t *testing.T) {
	c := &Counter{}
	c.Add(10)
	c.Add(5)
	if v := c.Value(); v != 15 {
		t.Errorf("Counter.Add() = %v, want 15", v)
	}
}

func TestCounter_Concurrent(t *testing.T) {
	c := &Counter{}
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			c.Inc()
			done <- true
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	if v := c.Value(); v != 100 {
		t.Errorf("concurrent Inc = %v, want 100", v)
	}
}

// --- Gauge ---

func TestGauge_Set(t *testing.T) {
	g := &Gauge{}
	g.Set(42)
	if v := g.Value(); v != 42 {
		t.Errorf("Gauge.Set(42) = %v, want 42", v)
	}
}

func TestGauge_Inc_Dec(t *testing.T) {
	g := &Gauge{}
	g.Set(5)
	g.Inc()
	g.Dec()
	g.Dec()
	if v := g.Value(); v != 4 {
		t.Errorf("after Inc, Dec, Dec from 5 = %v, want 4", v)
	}
}

// --- Histogram ---

func TestHistogram_Observe(t *testing.T) {
	h := &Histogram{
		buckets: []float64{0.1, 0.5, 1},
		counts:  []uint64{0, 0, 0},
	}
	h.Observe(0.05)
	h.Observe(0.2)
	h.Observe(0.8)
	h.Observe(1.5)

	_, cs, sum, cnt := h.Snapshot()
	if len(h.buckets) != 3 {
		t.Fatalf("buckets len = %d, want 3", len(h.buckets))
	}
	if cs[0] != 1 {
		t.Errorf("bucket[0] = %d, want 1", cs[0])
	}
	if cs[1] != 2 {
		t.Errorf("bucket[1] = %d, want 2", cs[1])
	}
	if cs[2] != 3 {
		t.Errorf("bucket[2] = %d, want 3", cs[2])
	}
	if sum != 2.55 {
		t.Errorf("sum = %v, want 2.55", sum)
	}
	if cnt != 4 {
		t.Errorf("count = %d, want 4", cnt)
	}
}

func TestHistogram_SortDedup(t *testing.T) {
	result := sortDedup([]float64{1, 0.5, 0.1, 0.5, 2})
	if len(result) != 4 {
		t.Fatalf("sortDedup = %v, expected 4 unique", result)
	}
	if result[0] != 0.1 || result[1] != 0.5 || result[2] != 1 || result[3] != 2 {
		t.Errorf("sortDedup = %v, want [0.1 0.5 1 2]", result)
	}
}

func TestHistogram_Concurrent(t *testing.T) {
	h := &Histogram{buckets: []float64{0.1, 1}, counts: []uint64{0, 0}}
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			h.Observe(0.05)
			h.Observe(0.5)
			done <- true
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	_, cs, _, cnt := h.Snapshot()
	if cnt != 200 {
		t.Errorf("concurrent Observe count = %d, want 200", cnt)
	}
	if cs[0] != 100 {
		t.Errorf("bucket[0] = %d, want 100", cs[0])
	}
	if cs[1] != 200 {
		t.Errorf("bucket[1] = %d, want 200", cs[1])
	}
}

// --- Registry ---

func TestRegistry_NewCounter_Dedup(t *testing.T) {
	r := NewRegistry()
	c1 := r.NewCounter("test", "help", nil)
	c2 := r.NewCounter("test", "help", nil)
	c1.Inc()
	if c1 != c2 {
		t.Error("NewCounter must return same instance")
	}
}

func TestRegistry_NewGauge_Dedup(t *testing.T) {
	r := NewRegistry()
	g1 := r.NewGauge("test", "help", nil)
	g2 := r.NewGauge("test", "help", nil)
	g1.Set(42)
	if g2.Value() != 42 {
		t.Error("NewGauge must return same instance")
	}
}

func TestRegistry_NewHistogram_Dedup(t *testing.T) {
	r := NewRegistry()
	h1 := r.NewHistogram("test", "help", []float64{0.1, 1}, nil)
	h2 := r.NewHistogram("test", "help", []float64{0.5, 2}, nil)
	h1.Observe(0.05)
	if h1 != h2 {
		t.Error("NewHistogram must return same instance")
	}
}

// --- WriteTo ---

func TestWriteTo_Counter(t *testing.T) {
	r := NewRegistry()
	r.NewCounter("test_counter", "Test counter", nil)
	r.counters["test_counter"].Add(42)

	rr := httptest.NewRecorder()
	r.WriteTo(rr)
	out := rr.Body.String()
	if !strings.Contains(out, "# TYPE test_counter counter") {
		t.Errorf("missing # TYPE: %s", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("missing value 42: %s", out)
	}
}

func TestWriteTo_Gauge(t *testing.T) {
	r := NewRegistry()
	r.NewGauge("test_gauge", "Test gauge", nil)
	r.gauges["test_gauge"].Set(100)

	rr := httptest.NewRecorder()
	r.WriteTo(rr)
	out := rr.Body.String()
	if !strings.Contains(out, "100") {
		t.Errorf("gauge value 100 not found: %s", out)
	}
}

func TestWriteTo_Histogram(t *testing.T) {
	r := NewRegistry()
	r.NewHistogram("test_hist", "Test histogram", []float64{0.1, 1}, nil)
	r.histograms["test_hist"].Observe(0.05)
	r.histograms["test_hist"].Observe(0.5)
	r.histograms["test_hist"].Observe(1.5)

	rr := httptest.NewRecorder()
	r.WriteTo(rr)
	out := rr.Body.String()
	if !strings.Contains(out, "# TYPE test_hist histogram") {
		t.Errorf("missing # TYPE histogram: %s", out)
	}
	if !strings.Contains(out, "test_hist_bucket") {
		t.Error("missing test_hist_bucket")
	}
	if !strings.Contains(out, "test_hist_sum") {
		t.Error("missing test_hist_sum")
	}
	if !strings.Contains(out, "test_hist_count") {
		t.Error("missing test_hist_count")
	}
}

func TestWriteTo_Labels(t *testing.T) {
	r := NewRegistry()
	r.NewCounter("http_requests", "HTTP requests", map[string]string{"method": "GET", "path": "/api"})
	r.counters["http_requests{method=\"GET\",path=\"/api\"}"].Inc()

	rr := httptest.NewRecorder()
	r.WriteTo(rr)
	out := rr.Body.String()
	if !strings.Contains(out, "method=\"GET\"") {
		t.Errorf("missing method label: %s", out)
	}
	if !strings.Contains(out, "path=\"/api\"") {
		t.Errorf("missing path label: %s", out)
	}
	if !strings.Contains(out, "{method=\"GET\",path=\"/api\"}") {
		t.Errorf("labels must be in braces: %s", out)
	}
}

// --- HTTPHandler ---

func TestHandler_ServeHTTP(t *testing.T) {
	NewRegistry().NewCounter("http_requests_total", "HTTP requests", map[string]string{"method": "GET", "path": "/test"})

	h := NewHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/test", nil))

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestWriteTo_HttpHandler(t *testing.T) {
	r := NewRegistry()
	r.NewCounter("http_requests_total", "HTTP requests", nil)
	r.counters["http_requests_total"].Inc()

	rr := httptest.NewRecorder()
	r.WriteTo(rr)

	body := rr.Body.String()
	if !strings.Contains(body, "http_requests_total{") {
		t.Errorf("missing metric in HTTP response: %s", body)
	}
}

// --- InitMetrics ---

func TestInitMetrics(t *testing.T) {
	InitMetrics()
	metrics := Global.Gather()
	if len(metrics) < 10 {
		t.Errorf("expected 10+ metrics, got %d", len(metrics))
	}
}

// --- EscapeLabel ---

func TestEscapeLabel(t *testing.T) {
	cases := []struct {
		input  string
		expect string
	}{
		{"hello", "hello"},
		{`back\slash`, `back\\slash`},
		{`quote"`, `quote\"`},
		{"line\nbreak", "line\\nbreak"},
	}
	for _, tc := range cases {
		if got := escapeLabel(tc.input); got != tc.expect {
			t.Errorf("escapeLabel(%q) = %q, want %q", tc.input, got, tc.expect)
		}
	}
}

// --- labelString ---

func TestLabelString_Sorted(t *testing.T) {
	l := map[string]string{"z": "1", "a": "2"}
	out := labelString(l)
	if !strings.HasPrefix(out, "a=\"2\"") {
		t.Errorf("labels must be sorted: %s", out)
	}
}
