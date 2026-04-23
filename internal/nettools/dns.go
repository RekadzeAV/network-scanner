package nettools

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// DNSResult прямое и обратное разрешение.
type DNSResult struct {
	Query        string
	ForwardIPs   []string
	ReverseNames []string
}

// LookupDNS выполняет LookupHost и при IP — LookupAddr.
func LookupDNS(ctx context.Context, host string) (*DNSResult, error) {
	return LookupDNSWithResolver(ctx, host, "")
}

// LookupDNSWithResolver выполняет LookupHost/LookupAddr с опциональным DNS-сервером (например "1.1.1.1:53").
func LookupDNSWithResolver(ctx context.Context, host string, resolverAddr string) (*DNSResult, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("пустой запрос")
	}
	resolverAddr = strings.TrimSpace(resolverAddr)
	r := &net.Resolver{}
	r.PreferGo = true
	if resolverAddr != "" {
		if !strings.Contains(resolverAddr, ":") {
			resolverAddr += ":53"
		}
		d := &net.Dialer{Timeout: 5 * time.Second}
		r.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
			return d.DialContext(ctx, "udp", resolverAddr)
		}
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	out := &DNSResult{Query: host}
	if ip := net.ParseIP(host); ip != nil {
		names, err := r.LookupAddr(lookupCtx, ip.String())
		if err != nil {
			return out, normalizeDNSError(err)
		}
		out.ReverseNames = names
		return out, nil
	}
	ips, err := r.LookupHost(lookupCtx, host)
	if err != nil {
		return out, normalizeDNSError(err)
	}
	out.ForwardIPs = ips
	return out, nil
}

func normalizeDNSError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return newToolError("dns", ToolErrorTimeout, "DNS lookup превысил таймаут", err)
	}
	var nerr net.Error
	if errors.As(err, &nerr) && nerr.Timeout() {
		return newToolError("dns", ToolErrorTimeout, "DNS lookup превысил таймаут", err)
	}
	var derr *net.DNSError
	if errors.As(err, &derr) {
		if derr.IsTimeout {
			return newToolError("dns", ToolErrorTimeout, "DNS lookup превысил таймаут", err)
		}
		if derr.IsNotFound {
			return newToolError("dns", ToolErrorNetwork, "запись не найдена", err)
		}
	}
	return newToolError("dns", ToolErrorNetwork, "ошибка DNS lookup", err)
}
