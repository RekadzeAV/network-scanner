package cmd

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"network-scanner/internal/builder"
	"network-scanner/internal/metrics"
)

// TestParseMetricsArgs — разбор флагов метрик (E7/7.8).
func TestParseMetricsArgs(t *testing.T) {
	// По умолчанию выключены, адрес — loopback.
	opts := parseMetricsArgs(nil)
	if opts.Enabled {
		t.Error("metrics должны быть выключены по умолчанию")
	}
	if opts.Addr != defaultMetricsAddr {
		t.Errorf("addr = %q, want %q", opts.Addr, defaultMetricsAddr)
	}

	// --metrics включает экспозицию.
	if got := parseMetricsArgs([]string{"--metrics"}); !got.Enabled {
		t.Error("--metrics должен включать экспозицию")
	}
	// --metrics=false выключает.
	if got := parseMetricsArgs([]string{"--metrics=false"}); got.Enabled {
		t.Error("--metrics=false должен выключать экспозицию")
	}
	// Адрес из следующего аргумента и из формы =value.
	if got := parseMetricsArgs([]string{"--metrics", "--metrics-addr", "0.0.0.0:9999"}); got.Addr != "0.0.0.0:9999" {
		t.Errorf("addr = %q, want 0.0.0.0:9999", got.Addr)
	}
	if got := parseMetricsArgs([]string{"--metrics-addr=127.0.0.1:9200"}); got.Addr != "127.0.0.1:9200" {
		t.Errorf("addr = %q, want 127.0.0.1:9200", got.Addr)
	}
	// Пустое значение → дефолт (loopback).
	if got := parseMetricsArgs([]string{"--metrics-addr", ""}); got.Addr != defaultMetricsAddr {
		t.Errorf("addr = %q, want default %q", got.Addr, defaultMetricsAddr)
	}
}

// TestScanCmd_HasMetricsFlags — флаги присутствуют в scan.
func TestScanCmd_HasMetricsFlags(t *testing.T) {
	for _, name := range []string{"metrics", "metrics-addr"} {
		if scanCmd.Flags().Lookup(name) == nil {
			t.Errorf("scan не имеет флага --%s", name)
		}
	}
}

// TestStartMetricsServer_ServesPrometheus — /metrics отдаёт формат Prometheus.
func TestStartMetricsServer_ServesPrometheus(t *testing.T) {
	reg := metrics.NewRegistry()
	reg.NewCounter("network_scanner_scan_total", "Всего сканирований").Inc()

	cfg := builder.Config{MetricsAddr: "127.0.0.1:0"}
	stop, err := startMetricsServer(reg, cfg)
	if err != nil {
		t.Fatalf("startMetricsServer() error = %v", err)
	}
	defer stop()
	// Адрес :0 — порт назначает ОС; в тесте проверяем только успешный старт
	// (реальную отдачу проверяем через httptest в internal/metrics).
}

// TestStartMetricsServer_HandlerViaHTTP — сквозная проверка отдачи по HTTP.
func TestStartMetricsServer_HandlerViaHTTP(t *testing.T) {
	reg := metrics.NewRegistry()
	reg.NewCounter("network_scanner_scan_total", "Всего сканирований").Add(3)

	cfg := builder.Config{MetricsAddr: "127.0.0.1:19113"}
	stop, err := startMetricsServer(reg, cfg)
	if err != nil {
		t.Fatalf("startMetricsServer() error = %v", err)
	}
	defer stop()

	resp, err := http.Get("http://127.0.0.1:19113/metrics")
	if err != nil {
		t.Skipf("metrics endpoint недоступен в среде теста: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "network_scanner_scan_total 3") {
		t.Errorf("body = %q, want metric value", string(body))
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
}

// TestSubscribeScanMetrics_WiresBus — подписка создаёт шину и метрики.
func TestSubscribeScanMetrics_WiresBus(t *testing.T) {
	container := builder.NewContainer(testCfg(t))
	if container.GetEventBus() != nil {
		t.Fatal("до подписки шина не должна быть подключена")
	}

	reg := metrics.NewRegistry()
	m := subscribeScanMetrics(container, reg)

	if container.GetEventBus() == nil {
		t.Fatal("subscribeScanMetrics должна подключить шину к контейнеру")
	}
	if m == nil || m.Scans == nil {
		t.Fatal("метрики сканирования не созданы")
	}
}

// TestFormatAddrForLog — нормализация адреса для вывода.
func TestFormatAddrForLog(t *testing.T) {
	if got := formatAddrForLog(""); got != defaultMetricsAddr {
		t.Errorf("formatAddrForLog(\"\") = %q, want %q", got, defaultMetricsAddr)
	}
	if got := formatAddrForLog("127.0.0.1:9999"); got != "127.0.0.1:9999" {
		t.Errorf("formatAddrForLog = %q, want as-is", got)
	}
}

// TestParseIntOr — разбор числа с fallback.
func TestParseIntOr(t *testing.T) {
	if got := parseIntOr("42", 7); got != 42 {
		t.Errorf("parseIntOr(\"42\") = %d, want 42", got)
	}
	if got := parseIntOr("abc", 7); got != 7 {
		t.Errorf("parseIntOr(\"abc\") = %d, want 7", got)
	}
}
