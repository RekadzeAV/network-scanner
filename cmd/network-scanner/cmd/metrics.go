// Метрики (E7/7.8): `--metrics` поднимает HTTP-эндпоинт /metrics в формате
// Prometheus и подключает обновление метрик из событий сканирования.
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"network-scanner/internal/builder"
	"network-scanner/internal/eventbus"
	"network-scanner/internal/metrics"
)

// defaultMetricsAddr — адрес эндпоинта метрик по умолчанию: только loopback,
// чтобы экспозиция метрик не была доступна из сети без явного указания адреса.
const defaultMetricsAddr = "127.0.0.1:9101"

// metricsOptions — параметры экспозиции метрик, разобранные из argv.
type metricsOptions struct {
	Enabled bool
	Addr    string
}

// parseMetricsArgs разбирает флаги метрик из argv (--metrics / --metrics-addr).
func parseMetricsArgs(args []string) metricsOptions {
	opts := metricsOptions{Addr: defaultMetricsAddr}
	for i := 0; i < len(args); i++ {
		name, inline, hasInline := splitFlag(args[i])
		switch name {
		case "--metrics":
			opts.Enabled = parseBoolFlag(inline, hasInline, true)
		case "--metrics-addr":
			if hasInline {
				opts.Addr = strings.TrimSpace(inline)
				continue
			}
			if i+1 < len(args) {
				i++
				opts.Addr = strings.TrimSpace(args[i])
			}
		}
	}
	if strings.TrimSpace(opts.Addr) == "" {
		opts.Addr = defaultMetricsAddr
	}
	return opts
}

// startMetricsServer поднимает HTTP-сервер с /metrics.
//
// Возвращает функцию остановки. Сервер запускается в отдельной горутине: ошибка
// прослушивания печатается, но не прерывает сканирование (метрики — вспомогательный
// контур). Таймауты обязательны (gosec G112/G114).
func startMetricsServer(reg *metrics.Registry, cfg builder.Config) (func(), error) {
	addr := strings.TrimSpace(cfg.MetricsAddr)
	if addr == "" {
		addr = defaultMetricsAddr
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", reg.Handler())
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("metrics: сервер остановлен: %v\n", err)
		}
	}()

	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
	return stop, nil
}

// formatAddrForLog нормализует адрес для вывода в лог.
func formatAddrForLog(addr string) string {
	if strings.TrimSpace(addr) == "" {
		return defaultMetricsAddr
	}
	return addr
}

// subscribeScanMetrics подключает шину событий к сканеру и метрикам: события
// scan.started/completed/failed обновляют счётчики.
func subscribeScanMetrics(container *builder.Container, reg *metrics.Registry) *metrics.ScanMetrics {
	m := metrics.NewScanMetrics(reg)
	bus := container.GetEventBus()
	if bus == nil {
		bus = eventbus.NewEventBus()
		container.WithEventBus(bus)
	}
	m.Register(bus)
	return m
}

// parseIntOr возвращает значение флага или defaultValue при ошибке разбора.
func parseIntOr(raw string, defaultValue int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
		return n
	}
	return defaultValue
}
