package scanner

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// ICMP ping timeout
	icmpPingTimeout = 1 * time.Second
)

// ICMPPinger интерфейс для ICMP ping операций
type ICMPPinger interface {
	PingICMP(host string, timeout time.Duration) (bool, error)
}

// DefaultICMPPinger реализует ICMP ping через системные утилиты
type DefaultICMPPinger struct{}

// validateICMPPingHost проверяет, что хост безопасен для передачи системной
// утилите ping через аргументы командной строки: строго IP или FQDN
// ([a-zA-Z0-9._-]), без shell-метасимволов и флагов CLI.
func validateICMPPingHost(host string) (string, error) {
	h := strings.TrimSpace(host)
	if h == "" {
		return "", fmt.Errorf("host is required")
	}
	if len(h) > 253 {
		return "", fmt.Errorf("host too long")
	}
	if strings.ContainsAny(h, " \t\r\n\\/\"'`$;&|<>(){}[]!") {
		return "", fmt.Errorf("host contains forbidden characters")
	}
	if ip := net.ParseIP(h); ip != nil {
		return h, nil
	}
	if strings.HasPrefix(h, "-") {
		return "", fmt.Errorf("host must not start with '-'")
	}
	for _, r := range h {
		isAlnum := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
		if !isAlnum && r != '.' && r != '-' && r != '_' {
			return "", fmt.Errorf("host contains invalid character %q", r)
		}
	}
	if !strings.Contains(h, ".") {
		return "", fmt.Errorf("host must be a FQDN or IP address")
	}
	return h, nil
}

// PingICMP выполняет ICMP ping через системную утилиту ping
func (p *DefaultICMPPinger) PingICMP(host string, timeout time.Duration) (bool, error) {
	validHost, err := validateICMPPingHost(host)
	if err != nil {
		return false, err
	}
	if timeout <= 0 {
		timeout = icmpPingTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Windows: ping -n 1 -w <timeout_ms> <host>
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", strconv.FormatInt(timeout.Milliseconds(), 10), validHost) //nolint:gosec // G204: host валидирован validateICMPPingHost, числа через strconv
	case "darwin":
		// macOS: ping -c 1 -t <timeout_s> <host>
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-t", strconv.FormatFloat(timeout.Seconds(), 'f', -1, 64), validHost) //nolint:gosec // G204: host валидирован
	default:
		// Linux/Unix: ping -c 1 -W <timeout_s> <host>
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-W", strconv.Itoa(int(timeout.Seconds())), validHost) //nolint:gosec // G204: host валидирован
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Проверяем таймаут
		if ctx.Err() == context.DeadlineExceeded {
			return false, fmt.Errorf("icmp ping timeout for %s", validHost)
		}
		// Проверяем результат по коду возврата
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Код возврата 1 обычно означает host unreachable или timeout
			return false, fmt.Errorf("icmp ping failed for %s: exit code %d, output: %s", validHost, exitErr.ExitCode(), string(output))
		}
		return false, fmt.Errorf("icmp ping error for %s: %v, output: %s", validHost, err, string(output))
	}

	// Проверяем вывод на наличие успешного ответа
	outputStr := string(output)
	if icmpContainsString(outputStr, "1 packets transmitted", "1 packets received") ||
		icmpContainsString(outputStr, "1 received") ||
		icmpContainsString(outputStr, "bytes from") {
		return true, nil
	}

	return false, fmt.Errorf("icmp ping inconclusive for %s: %s", validHost, outputStr)
}

// Helper function to check if any of the strings are in the output
func icmpContainsString(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) && icmpContainsSubstring(s, substr) {
			return true
		}
	}
	return false
}

func icmpContainsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ICMPResult содержит результат ICMP ping
type ICMPResult struct {
	Host      string
	Alive     bool
	LatencyMs float64
	Error     error
}

// PingICMPPool выполняет параллельный ICMP ping для нескольких хостов
func PingICMPPool(ctx context.Context, hosts []string, timeout time.Duration) []ICMPResult {
	results := make([]ICMPResult, len(hosts))
	var pinger ICMPPinger = &DefaultICMPPinger{}

	for i, host := range hosts {
		select {
		case <-ctx.Done():
			results[i] = ICMPResult{
				Host:  host,
				Alive: false,
				Error: ctx.Err(),
			}
		default:
			start := time.Now()
			alive, err := pinger.PingICMP(host, timeout)
			latency := float64(time.Since(start).Milliseconds())

			results[i] = ICMPResult{
				Host:      host,
				Alive:     alive,
				LatencyMs: latency,
				Error:     err,
			}
		}
	}

	return results
}

// SetICMPPinger подменяет реализацию ICMP-пинга (тесты, альтернативный пингер).
//
// Пингер используется только когда ICMP-проверка включена через
// SetICMPPingEnabled; nil означает системный DefaultICMPPinger.
func (ns *NetworkScanner) SetICMPPinger(pinger ICMPPinger) {
	ns.icmpPinger = pinger
}

// pingICMP проверяет один хост штатным или подменённым ICMP-пингером.
//
// Таймаут берётся из настроек сканера, но не превышает icmpPingTimeout: ICMP
// отвечает на заведомо мёртвых хостах всё равно только по истечении таймаута.
func (ns *NetworkScanner) pingICMP(host string) (bool, error) {
	pinger := ns.icmpPinger
	if pinger == nil {
		pinger = &DefaultICMPPinger{}
	}

	timeout := ns.timeout
	if timeout <= 0 || timeout > icmpPingTimeout {
		timeout = icmpPingTimeout
	}

	return pinger.PingICMP(host, timeout)
}
