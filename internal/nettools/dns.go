package nettools

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// DNSResult прямое и обратное разрешение.
type DNSResult struct {
	Query       string
	ForwardIPs  []string
	ReverseNames []string
}

// LookupDNS выполняет LookupHost и при IP — LookupAddr.
func LookupDNS(ctx context.Context, host string) (*DNSResult, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("пустой запрос")
	}
	r := &net.Resolver{}
	r.PreferGo = true
	lookupCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	out := &DNSResult{Query: host}
	if ip := net.ParseIP(host); ip != nil {
		names, err := r.LookupAddr(lookupCtx, ip.String())
		if err != nil {
			return out, err
		}
		out.ReverseNames = names
		return out, nil
	}
	ips, err := r.LookupHost(lookupCtx, host)
	if err != nil {
		return out, err
	}
	out.ForwardIPs = ips
	return out, nil
}
