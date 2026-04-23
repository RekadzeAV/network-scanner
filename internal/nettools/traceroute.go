package nettools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// RunTraceroute вызывает tracert (Windows) или traceroute (Unix).
func RunTraceroute(ctx context.Context, host string, timeout time.Duration) (string, error) {
	return RunTracerouteWithMaxHops(ctx, host, timeout, 30)
}

// RunTracerouteWithMaxHops вызывает tracert/traceroute с настраиваемым числом прыжков.
func RunTracerouteWithMaxHops(ctx context.Context, host string, timeout time.Duration, maxHops int) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", fmt.Errorf("пустой хост")
	}
	if maxHops <= 0 {
		maxHops = 30
	}
	if maxHops > 64 {
		maxHops = 64
	}
	args := buildTracerouteArgs(host, maxHops, runtimeGOOS())
	return runCmd(ctx, args, timeout)
}

// RunTracerouteStructured выполняет traceroute и возвращает сырой вывод + hops.
func RunTracerouteStructured(ctx context.Context, host string, timeout time.Duration) (*TracerouteResult, error) {
	return RunTracerouteStructuredWithMaxHops(ctx, host, timeout, 30)
}

func buildTracerouteArgs(host string, maxHops int, goos string) []string {
	if strings.EqualFold(strings.TrimSpace(goos), "windows") {
		return []string{"tracert", "-d", "-h", strconv.Itoa(maxHops), host}
	}
	return []string{"traceroute", "-m", strconv.Itoa(maxHops), "-n", host}
}

// RunTracerouteStructuredWithMaxHops выполняет traceroute с max hops и возвращает сырой вывод + hops.
func RunTracerouteStructuredWithMaxHops(ctx context.Context, host string, timeout time.Duration, maxHops int) (*TracerouteResult, error) {
	raw, err := RunTracerouteWithMaxHops(ctx, host, timeout, maxHops)
	if err != nil {
		return nil, err
	}
	return &TracerouteResult{
		RawOutput: raw,
		Hops:      parseTraceroute(raw),
	}, nil
}
