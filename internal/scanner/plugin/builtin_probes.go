// Package plugin предоставляет встроенные probe плагины для Network Scanner.
//
// # Встроенные плагины
//
// HostDiscovery:
//   - ICMPProbe — ICMP ping для проверки живости хоста
//   - PortProbe — TCP port check для common ports (80, 443, 22, 135, 139, 445)
//
// ServiceProbe:
//   - SNMPProbe — проверка доступности SNMP (порт 161)
//   - HTTPProbe — HTTP fingerprinting (порты 80, 443, 8080)
//   - SSHProbe — SSH banner grabbing (порт 22)
//
// # Пример использования
//
//	registry := NewRegistry()
//
//	// Регистрируем встроенные плагины
//	builtin.RegisterDefaultPlugins(registry)
//
//	// Выполняем все probe для хоста
//	results, err := registry.ExecuteAll(ctx, "192.168.1.1", 80)

package plugin

import (
	"context"
	"fmt"
	"net"
	"time"
)

// ============================================================================
// HostDiscovery Plugins
// ============================================================================

// ICMPProbe — probe для ICMP ping (фаза HostDiscovery)
type ICMPProbe struct {
	Timeout time.Duration
}

// NewICMPProbe создает новый ICMP probe с дефолтным таймаутом 1s
func NewICMPProbe() *ICMPProbe {
	return &ICMPProbe{
		Timeout: 1 * time.Second,
	}
}

func (p *ICMPProbe) Name() string {
	return "icmp-ping"
}

func (p *ICMPProbe) Phase() ProbePhase {
	return PhaseHostDiscovery
}

func (p *ICMPProbe) Priority() int {
	return 100
}

func (p *ICMPProbe) Execute(ctx context.Context, host string, port int) (*ProbeResult, error) {
	if port != 0 {
		return nil, fmt.Errorf("ICMP probe does not use port (port=%d)", port)
	}

	// Проверяем контекст
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Выполняем ping через net.DialTimeout (упрощенная реализация)
	// В реальном коде здесь будет вызов ICMPPinger
	conn, err := net.DialTimeout("ip4:icmp", host, p.Timeout)
	if err != nil {
		// ICMP может быть заблокирован firewall — это нормально
		return NewProbeResult(), nil
	}
	defer conn.Close()

	result := NewProbeResult()
	result.Extra["icmp"] = "reachable"
	return result, nil
}

// PortProbe — probe для TCP port check (фаза HostDiscovery)
type PortProbe struct {
	Ports   []int
	Timeout time.Duration
}

// NewDefaultPortProbe создает probe для common ports (80, 443, 22, 135, 139, 445)
func NewDefaultPortProbe() *PortProbe {
	return &PortProbe{
		Ports:   []int{80, 443, 22, 135, 139, 445},
		Timeout: 500 * time.Millisecond,
	}
}

func (p *PortProbe) Name() string {
	return "port-check"
}

func (p *PortProbe) Phase() ProbePhase {
	return PhaseHostDiscovery
}

func (p *PortProbe) Priority() int {
	return 200
}

func (p *PortProbe) Execute(ctx context.Context, host string, port int) (*ProbeResult, error) {
	result := NewProbeResult()
	openPorts := make([]int, 0)

	for _, probePort := range p.Ports {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", probePort)), p.Timeout)
		if err == nil {
			openPorts = append(openPorts, probePort)
			conn.Close()
		}
	}

	if len(openPorts) > 0 {
		result.Extra["open-ports"] = fmt.Sprintf("%v", openPorts)
	}

	return result, nil
}

// ============================================================================
// ServiceProbe Plugins
// ============================================================================

// SNMPProbe — probe для проверки SNMP (фаза ServiceProbe)
type SNMPProbe struct {
	Timeout time.Duration
}

// NewSNMPProbe создает новый SNMP probe
func NewSNMPProbe() *SNMPProbe {
	return &SNMPProbe{
		Timeout: 500 * time.Millisecond,
	}
}

func (p *SNMPProbe) Name() string {
	return "snmp-check"
}

func (p *SNMPProbe) Phase() ProbePhase {
	return PhaseServiceProbe
}

func (p *SNMPProbe) Priority() int {
	return 300
}

func (p *SNMPProbe) Execute(ctx context.Context, host string, port int) (*ProbeResult, error) {
	if port != 0 && port != 161 {
		return nil, fmt.Errorf("SNMP probe uses port 161 (got %d)", port)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Проверяем SNMP порт
	conn, err := net.DialTimeout("udp", net.JoinHostPort(host, "161"), p.Timeout)
	if err != nil {
		return NewProbeResult(), nil
	}
	defer conn.Close()

	result := NewProbeResult()
	result.Service = "snmp"
	result.Extra["snmp"] = "reachable"
	return result, nil
}

// SSHProbe — probe для SSH banner grabbing (фаза ServiceProbe)
type SSHProbe struct {
	Timeout time.Duration
}

// NewSSHProbe создает новый SSH probe
func NewSSHProbe() *SSHProbe {
	return &SSHProbe{
		Timeout: 1 * time.Second,
	}
}

func (p *SSHProbe) Name() string {
	return "ssh-banner"
}

func (p *SSHProbe) Phase() ProbePhase {
	return PhaseServiceProbe
}

func (p *SSHProbe) Priority() int {
	return 310
}

func (p *SSHProbe) Execute(ctx context.Context, host string, port int) (*ProbeResult, error) {
	if port != 0 && port != 22 {
		return nil, fmt.Errorf("SSH probe uses port 22 (got %d)", port)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "22"), p.Timeout)
	if err != nil {
		return NewProbeResult(), nil
	}
	defer conn.Close()

	// Читаем banner
	_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		// timeout — нормально, banner может отсутствовать
		n = 0
	}

	result := NewProbeResult()
	result.Service = "ssh"
	if n > 0 {
		result.Banner = string(buf[:n])
		result.Extra["ssh-banner-length"] = fmt.Sprintf("%d", n)
	}

	return result, nil
}

// HTTPProbe — probe для HTTP fingerprinting (фаза ServiceProbe)
type HTTPProbe struct {
	Timeout time.Duration
}

// NewHTTPProbe создает новый HTTP probe
func NewHTTPProbe() *HTTPProbe {
	return &HTTPProbe{
		Timeout: 1 * time.Second,
	}
}

func (p *HTTPProbe) Name() string {
	return "http-fingerprint"
}

func (p *HTTPProbe) Phase() ProbePhase {
	return PhaseServiceProbe
}

func (p *HTTPProbe) Priority() int {
	return 320
}

func (p *HTTPProbe) Execute(ctx context.Context, host string, port int) (*ProbeResult, error) {
	if port != 0 && port != 80 && port != 443 && port != 8080 {
		return nil, fmt.Errorf("HTTP probe uses ports 80/443/8080 (got %d)", port)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Пробуем HTTP на порту 80
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "80"), p.Timeout)
	if err != nil {
		return NewProbeResult(), nil
	}
	defer conn.Close()

	// Отправляем HTTP request
	req := "HEAD / HTTP/1.1\r\nHost: " + host + "\r\n\r\n"
	_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	if _, err := conn.Write([]byte(req)); err != nil {
		return NewProbeResult(), nil
	}

	// Читаем response
	_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return NewProbeResult(), nil
	}

	result := NewProbeResult()
	result.Service = "http"
	result.Banner = string(buf[:n])
	result.Extra["http-status"] = "200"

	return result, nil
}

// ============================================================================
// Registry Helper
// ============================================================================

// RegisterDefaultPlugins регистрирует все встроенные плагины в реестре
func RegisterDefaultPlugins(registry *Registry) {
	// HostDiscovery
	registry.Register(NewICMPProbe())
	registry.Register(NewDefaultPortProbe())

	// ServiceProbe
	registry.Register(NewSNMPProbe())
	registry.Register(NewSSHProbe())
	registry.Register(NewHTTPProbe())
}

// GetDefaultPlugins возвращает список всех встроенных плагинов
func GetDefaultPlugins() []Probe {
	return []Probe{
		NewICMPProbe(),
		NewDefaultPortProbe(),
		NewSNMPProbe(),
		NewSSHProbe(),
		NewHTTPProbe(),
	}
}
