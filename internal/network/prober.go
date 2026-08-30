package network

import (
	"context"
	"fmt"
	"net"
	"time"
)

// DefaultNetworkProber is the default implementation for liveness and MAC probes.
type DefaultNetworkProber struct {
	Timeout  time.Duration
	arpCache *ARPCache
}

// NewDefaultNetworkProber создаёт новый пробер с инициализированным ARP-кэшем.
func NewDefaultNetworkProber(timeout time.Duration) *DefaultNetworkProber {
	return &DefaultNetworkProber{
		Timeout:  timeout,
		arpCache: NewDefaultARPCache(5 * time.Minute),
	}
}

// SetARPCache устанавливает пользовательский ARP-кэш.
func (p *DefaultNetworkProber) SetARPCache(cache *ARPCache) {
	p.arpCache = cache
}

// Ping checks host availability using a short TCP probe set.
func (p DefaultNetworkProber) Ping(ip string) (bool, error) {
	return p.PingContext(ip, nil)
}

// PingContext checks host availability and supports cancellation via done channel.
func (p DefaultNetworkProber) PingContext(ip string, done <-chan struct{}) (bool, error) {
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = time.Second
	}

	ports := []int{80, 443, 22, 135, 139, 445}
	probeTimeout := timeout / 3
	if probeTimeout < 150*time.Millisecond {
		probeTimeout = 150 * time.Millisecond
	}
	if probeTimeout > 800*time.Millisecond {
		probeTimeout = 800 * time.Millisecond
	}

	baseCtx := context.Background()
	if done != nil {
		var cancelBase context.CancelFunc
		baseCtx, cancelBase = context.WithCancel(baseCtx)
		defer cancelBase()
		go func() {
			select {
			case <-done:
				cancelBase()
			case <-baseCtx.Done():
			}
		}()
	}
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	results := make(chan bool, len(ports))
	for _, port := range ports {
		go func(port int) {
			dialer := &net.Dialer{Timeout: probeTimeout}
			conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
			if err == nil {
				if conn != nil {
					_ = conn.Close()
				}
				select {
				case results <- true:
				default:
				}
				cancel()
				return
			}
			select {
			case results <- false:
			default:
			}
		}(port)
	}

	for i := 0; i < len(ports); i++ {
		if <-results {
			return true, nil
		}
	}

	return false, nil
}

// ResolveMAC attempts to resolve MAC from parsed IP.
// Uses ARP cache for fast resolution.
func (p *DefaultNetworkProber) ResolveMAC(ip string) (net.HardwareAddr, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil, fmt.Errorf("invalid IP: %s", ip)
	}

	// Если ARP-кэш не инициализирован — возвращаем ошибку
	if p.arpCache == nil {
		return nil, fmt.Errorf("ARP cache not initialized")
	}

	macStr, err := p.arpCache.Get(ip)
	if err != nil {
		return nil, fmt.Errorf("MAC resolution failed for %s: %w", ip, err)
	}

	mac, err := net.ParseMAC(macStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MAC address %q for %s: %w", macStr, ip, err)
	}

	return mac, nil
}
