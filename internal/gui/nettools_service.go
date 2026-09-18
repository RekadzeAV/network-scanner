package gui

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"network-scanner/internal/nettools"
)

// PingResult результат ping
type PingResult struct {
	Success  bool
	Output   string
	Duration time.Duration
	Host     string
	Packets  int
	Timeout  time.Duration
	Error    string
}

// TracerouteResult результат traceroute
type TracerouteResult struct {
	Success  bool
	Output   string
	Duration time.Duration
	Host     string
	Hops     int
	Error    string
}

// DNSResult результат DNS-запроса
type DNSResult struct {
	Success  bool
	Output   string
	Duration time.Duration
	Host     string
	Records  []string
	Error    string
}

// WhoisResult результат whois
type WhoisResult struct {
	Success  bool
	Output   string
	Duration time.Duration
	Domain   string
	Error    string
}

// NetToolsService обёртка для сетевых инструментов
type NetToolsService struct {
}

// NewNetToolsService создаёт NetToolsService
func NewNetToolsService() *NetToolsService {
	return &NetToolsService{}
}

// validateNetToolHost проверяет, что хост безопасен для передачи внешнему
// процессу (ping/traceroute/whois) через аргументы командной строки:
//   - для IP: строго net.ParseIP (исключает метасимволы и injection);
//   - для домена: только [a-zA-Z0-9._-], длина ≤253 (RFC 1035);
//   - запрет пробелов, слэшей, shell-метасимволов и ведущего '-' (флаги CLI).
//
// Возвращает нормализованное значение (trimmed).
func validateNetToolHost(host string) (string, error) {
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
		return h, nil // валидный IP (v4/v6)
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

// Ping выполняет ping
func (s *NetToolsService) Ping(ctx context.Context, host string, count int, timeout time.Duration) (*PingResult, error) {
	validHost, err := validateNetToolHost(host)
	if err != nil {
		return nil, err
	}
	if count < 1 {
		count = 1
	}
	if count > 100 {
		count = 100 // ограничение: не превращать ping в DoS-инструмент
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	start := time.Now()
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "ping", "-n", strconv.Itoa(count), "-w", strconv.FormatInt(timeout.Milliseconds(), 10), validHost) //nolint:gosec // G204: host валидирован validateNetToolHost, числа через strconv
	default:
		cmd = exec.CommandContext(ctx, "ping", "-c", strconv.Itoa(count), "-W", strconv.Itoa(int(timeout.Seconds())), validHost) //nolint:gosec // G204: host валидирован validateNetToolHost, числа через strconv
	}

	output, execErr := cmd.CombinedOutput()
	pingErr := ""
	if execErr != nil {
		pingErr = execErr.Error()
	}
	return &PingResult{
		Success:  execErr == nil,
		Output:   string(output),
		Duration: time.Since(start),
		Host:     validHost,
		Packets:  count,
		Timeout:  timeout,
		Error:    pingErr,
	}, nil
}

// Traceroute выполняет traceroute
func (s *NetToolsService) Traceroute(ctx context.Context, host string, maxHops int) (*TracerouteResult, error) {
	validHost, err := validateNetToolHost(host)
	if err != nil {
		return nil, err
	}
	if maxHops < 1 {
		maxHops = 1
	}
	if maxHops > 64 {
		maxHops = 64 // ограничение: RFC 7912 верхняя граница практического TTL
	}

	start := time.Now()
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "tracert", "-h", strconv.Itoa(maxHops), validHost) //nolint:gosec // G204: host валидирован validateNetToolHost, числа через strconv
	default:
		cmd = exec.CommandContext(ctx, "traceroute", "-m", strconv.Itoa(maxHops), validHost) //nolint:gosec // G204: host валидирован validateNetToolHost, числа через strconv
	}

	output, execErr := cmd.CombinedOutput()
	hops := strings.Count(string(output), "\n")
	traceErr := ""
	if execErr != nil {
		traceErr = execErr.Error()
	}
	return &TracerouteResult{
		Success:  execErr == nil,
		Output:   string(output),
		Duration: time.Since(start),
		Host:     validHost,
		Hops:     hops,
		Error:    traceErr,
	}, nil
}

// DNSLookup выполняет DNS-запрос с реальным resolver
func (s *NetToolsService) DNSLookup(ctx context.Context, host string, resolver string) (*DNSResult, error) {
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	start := time.Now()

	// Реальный вызов nettools.LookupDNSWithResolver
	result, err := nettools.LookupDNSWithResolver(ctx, host, resolver)

	records := make([]string, 0)
	if result != nil {
		records = append(records, result.ForwardIPs...)
		records = append(records, result.ReverseNames...)
	}

	return &DNSResult{
		Success:  err == nil,
		Output:   strings.Join(records, "\n"),
		Duration: time.Since(start),
		Host:     host,
		Records:  records,
		Error:    err.Error(),
	}, nil
}

// WhoisLookup выполняет whois-запрос
func (s *NetToolsService) WhoisLookup(ctx context.Context, domain string) (*WhoisResult, error) {
	validDomain, err := validateNetToolHost(domain)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	var output string
	var execErr error

	switch runtime.GOOS {
	case "windows":
		// Windows не имеет whois по умолчанию
		output = "Whois not available on Windows"
	default:
		cmd := exec.CommandContext(ctx, "whois", validDomain) //nolint:gosec // G204: domain валидирован validateNetToolHost
		out, e := cmd.CombinedOutput()
		output = string(out)
		execErr = e
	}

	whoisErr := ""
	if execErr != nil {
		whoisErr = execErr.Error()
	}
	return &WhoisResult{
		Success:  execErr == nil,
		Output:   output,
		Duration: time.Since(start),
		Domain:   validDomain,
		Error:    whoisErr,
	}, nil
}
