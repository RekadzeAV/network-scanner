package nettools

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// RunPing запускает системную утилиту ping (ICMP недоступен без raw-сокетов; это ожидаемый fallback).
func RunPing(ctx context.Context, host string, count int, timeout time.Duration) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", fmt.Errorf("пустой хост")
	}
	if count < 1 {
		count = 4
	}
	if count > 50 {
		count = 50
	}
	args := buildPingArgs(host, count, runtimeGOOS())
	return runCmd(ctx, args, timeout)
}

// RunPingStructured выполняет ping и возвращает сырой вывод + распарсенные метрики.
func RunPingStructured(ctx context.Context, host string, count int, timeout time.Duration) (*PingResult, error) {
	raw, err := RunPing(ctx, host, count, timeout)
	if err != nil {
		return nil, err
	}
	return &PingResult{
		RawOutput: raw,
		Stats:     parsePingStats(raw, count),
	}, nil
}

func buildPingArgs(host string, count int, goos string) []string {
	if strings.EqualFold(strings.TrimSpace(goos), "windows") {
		return []string{"ping", "-n", strconv.Itoa(count), host}
	}
	return []string{"ping", "-c", strconv.Itoa(count), host}
}

func runtimeGOOS() string {
	return runtime.GOOS
}
