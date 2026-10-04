package metrics

import (
	"strings"
	"testing"
	"time"

	"network-scanner/internal/eventbus"
)

// waitForValue ждёт, пока метрика достигнет значения (доставка событий асинхронна).
func waitForValue(t *testing.T, get func() float64, want float64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if get() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("metric = %v, want %v (timeout)", get(), want)
}

// TestScanMetrics_EventFlow — метрики обновляются по событиям сканирования.
func TestScanMetrics_EventFlow(t *testing.T) {
	reg := NewRegistry()
	m := NewScanMetrics(reg)

	bus := eventbus.NewEventBus()
	defer bus.Close()
	m.Register(bus)

	bus.Publish(eventbus.NewScanStartedEvent("192.168.1.0/24", "2s"))
	bus.Publish(eventbus.NewScanCompletedEvent(12, 34, "1500ms"))
	bus.Publish(eventbus.NewBaseEvent("scan.failed"))

	waitForValue(t, m.Scans.Value, 1)
	waitForValue(t, m.Failures.Value, 1)
	waitForValue(t, func() float64 { return float64(m.Duration.Count()) }, 1)
	waitForValue(t, m.Hosts.Value, 12)

	if got := m.OpenPorts.Value(); got != 34 {
		t.Errorf("open_ports = %v, want 34", got)
	}
	if sum := m.Duration.Sum(); sum < 1.4 || sum > 1.6 {
		t.Errorf("duration sum = %v, want ~1.5", sum)
	}
	// started(+1) → completed(-1) → failed(-1) = -1: гейдж отражает
	// дисбаланс событий, что полезно для диагностики.
	if got := m.Active.Value(); got != -1 {
		t.Errorf("active = %v, want -1 при непарных событиях", got)
	}
}

// TestScanMetrics_Render — метрики попадают в экспозицию Prometheus.
func TestScanMetrics_Render(t *testing.T) {
	reg := NewRegistry()
	m := NewScanMetrics(reg)
	m.Scans.Inc()
	m.Hosts.Add(3)

	out := reg.Render()
	for _, want := range []string{
		"# TYPE network_scanner_scan_total counter",
		"network_scanner_scan_total 1",
		"network_scanner_scan_hosts_total 3",
		"# TYPE network_scanner_scan_active gauge",
		"# TYPE network_scanner_scan_duration_seconds summary",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() missing %q\n%s", want, out)
		}
	}
}

// TestScanMetrics_RegisterNilAndIdempotent — nil-шина и повторная регистрация безопасны.
func TestScanMetrics_RegisterNilAndIdempotent(t *testing.T) {
	reg := NewRegistry()
	m := NewScanMetrics(reg)

	m.Register(nil) // не должно паниковать

	bus := eventbus.NewEventBus()
	defer bus.Close()
	m.Register(bus)
	m.Register(bus) // повторная регистрация игнорируется

	bus.Publish(eventbus.NewScanStartedEvent("10.0.0.0/24", "1s"))
	waitForValue(t, m.Scans.Value, 1)
}

// TestParseDurationSeconds — разбор различных представлений длительности.
func TestParseDurationSeconds(t *testing.T) {
	cases := []struct {
		in   interface{}
		want float64
	}{
		{"1500ms", 1.5},
		{"2s", 2},
		{float64(3), 3},
		{int64(4), 4},
		{"not-a-duration", 0},
		{nil, 0},
	}
	for _, c := range cases {
		if got := parseDurationSeconds(c.in); got != c.want {
			t.Errorf("parseDurationSeconds(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestPayloadFloat — извлечение чисел из payload события.
func TestPayloadFloat(t *testing.T) {
	payload := map[string]interface{}{
		"float":  float64(1.5),
		"int":    2,
		"int64":  int64(3),
		"string": "4.5",
		"bad":    "x",
	}
	if got := payloadFloat(payload, "float"); got != 1.5 {
		t.Errorf("float = %v", got)
	}
	if got := payloadFloat(payload, "int"); got != 2 {
		t.Errorf("int = %v", got)
	}
	if got := payloadFloat(payload, "int64"); got != 3 {
		t.Errorf("int64 = %v", got)
	}
	if got := payloadFloat(payload, "string"); got != 4.5 {
		t.Errorf("string = %v", got)
	}
	if got := payloadFloat(payload, "bad"); got != 0 {
		t.Errorf("bad = %v, want 0", got)
	}
	if got := payloadFloat(nil, "float"); got != 0 {
		t.Errorf("nil payload = %v, want 0", got)
	}
}
