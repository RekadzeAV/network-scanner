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
	var args []string
	switch runtime.GOOS {
	case "windows":
		args = []string{"ping", "-n", strconv.Itoa(count), host}
	default:
		args = []string{"ping", "-c", strconv.Itoa(count), host}
	}
	return runCmd(ctx, args, timeout)
}
