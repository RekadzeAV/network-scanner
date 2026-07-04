package gui

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// PingResult результат ping
type PingResult struct {
	Success    bool
	Output     string
	Duration   time.Duration
	Host       string
	Packets    int
	Timeout    time.Duration
	Error      string
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

// Ping выполняет ping
func (s *NetToolsService) Ping(ctx context.Context, host string, count int, timeout time.Duration) (*PingResult, error) {
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	start := time.Now()
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "ping", "-n", fmt.Sprintf("%d", count), "-w", fmt.Sprintf("%d", timeout.Milliseconds()), host)
	default:
		cmd = exec.CommandContext(ctx, "ping", "-c", fmt.Sprintf("%d", count), "-W", fmt.Sprintf("%d", int(timeout.Seconds())), host)
	}

	output, err := cmd.CombinedOutput()
	return &PingResult{
		Success:  err == nil,
		Output:   string(output),
		Duration: time.Since(start),
		Host:     host,
		Packets:  count,
		Timeout:  timeout,
		Error:    err.Error(),
	}, nil
}

// Traceroute выполняет traceroute
func (s *NetToolsService) Traceroute(ctx context.Context, host string, maxHops int) (*TracerouteResult, error) {
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	start := time.Now()
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "tracert", "-h", fmt.Sprintf("%d", maxHops), host)
	default:
		cmd = exec.CommandContext(ctx, "traceroute", "-m", fmt.Sprintf("%d", maxHops), host)
	}

	output, err := cmd.CombinedOutput()
	hops := strings.Count(string(output), "\n")
	return &TracerouteResult{
		Success:  err == nil,
		Output:   string(output),
		Duration: time.Since(start),
		Host:     host,
		Hops:     hops,
		Error:    err.Error(),
	}, nil
}

// DNSLookup выполняет DNS-запрос
func (s *NetToolsService) DNSLookup(ctx context.Context, host string, resolver string) (*DNSResult, error) {
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	start := time.Now()
	var records []string
	var err error
	
	if resolver != "" {
		// Используем указанный resolver
		records, err = net.LookupHost(host)
		_ = resolver // TODO: передать resolver в resolver
	} else {
		records, err = net.LookupHost(host)
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
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	start := time.Now()
	var output string
	var err error
	
	switch runtime.GOOS {
	case "windows":
		// Windows не имеет whois по умолчанию
		output = "Whois not available on Windows"
	default:
		cmd := exec.CommandContext(ctx, "whois", domain)
		out, e := cmd.CombinedOutput()
		output = string(out)
		err = e
	}

	return &WhoisResult{
		Success:  err == nil,
		Output:   output,
		Duration: time.Since(start),
		Domain:   domain,
		Error:    err.Error(),
	}, nil
}
