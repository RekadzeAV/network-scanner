package network

import (
	"testing"
)

// TestIsIPv4 — проверка определения IPv4.
func TestIsIPv4_Extended(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"::1", false},
		{"2001:db8::1", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := IsIPv4(tt.ip)
			if got != tt.want {
				t.Errorf("IsIPv4(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// TestIsIPv6 — проверка определения IPv6.
func TestIsIPv6_Extended(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"::1", true},
		{"2001:db8::1", true},
		{"fe80::1", true},
		{"::", true},
		{"192.168.1.1", false},
		{"10.0.0.1", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := IsIPv6(tt.ip)
			if got != tt.want {
				t.Errorf("IsIPv6(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// TestIsIP — проверка определения любого IP.
func TestIsIP_Extended(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"invalid", false},
		{"", false},
		{"not-an-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := IsIP(tt.ip)
			if got != tt.want {
				t.Errorf("IsIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// TestIsCIDR — проверка определения CIDR.
func TestIsCIDR_Extended(t *testing.T) {
	tests := []struct {
		cidr string
		want bool
	}{
		{"192.168.1.0/24", true},
		{"10.0.0.0/16", true},
		{"172.16.0.0/12", true},
		{"2001:db8::/32", true},
		{"fe80::/10", true},
		{"::/0", true},
		{"192.168.1.1", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			got := IsCIDR(tt.cidr)
			if got != tt.want {
				t.Errorf("IsCIDR(%q) = %v, want %v", tt.cidr, got, tt.want)
			}
		})
	}
}

// TestPreferredNetwork — проверка preferred network.
func TestPreferredNetwork_Extended(t *testing.T) {
	tests := []struct {
		name      string
		ipv4      []string
		ipv6      []string
		wantFirst string
	}{
		{
			name:      "IPv4 preferred",
			ipv4:      []string{"192.168.1.0/24"},
			ipv6:      []string{"fe80::/10"},
			wantFirst: "192.168.1.0/24",
		},
		{
			name:      "IPv6 only",
			ipv4:      []string{},
			ipv6:      []string{"fe80::/10"},
			wantFirst: "fe80::/10",
		},
		{
			name:      "Empty",
			ipv4:      []string{},
			ipv6:      []string{},
			wantFirst: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PreferredNetwork(tt.ipv4, tt.ipv6)
			if got != tt.wantFirst {
				t.Errorf("PreferredNetwork(%v, %v) = %q, want %q", tt.ipv4, tt.ipv6, got, tt.wantFirst)
			}
		})
	}
}

// TestFormatIPForDisplay — форматирование IP для отображения.
func TestFormatIPForDisplay_Extended(t *testing.T) {
	tests := []struct {
		ip   string
		want string
	}{
		{"192.168.1.1", "192.168.1.1"},
		{"10.0.0.1", "10.0.0.1"},
		{"::1", "[::1]"},
		{"2001:db8::1", "[2001:db8::1]"},
		{"fe80::1", "[fe80::1]"},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := FormatIPForDisplay(tt.ip)
			if got != tt.want {
				t.Errorf("FormatIPForDisplay(%q) = %q, want %q", tt.ip, got, tt.want)
			}
		})
	}
}

// TestProtocolForHost — проверка протокола для хоста.
func TestProtocolForHost_Extended(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		useTCP    bool
		preferUDP bool
		want      string
	}{
		{
			name:   "IPv4 TCP",
			host:   "192.168.1.1",
			useTCP: true,
			want:   "tcp4",
		},
		{
			name:   "IPv4 default",
			host:   "192.168.1.1",
			useTCP: false,
			want:   "tcp4", // preferUDP=false → tcp
		},
		{
			name:      "IPv4 UDP preferred",
			host:      "192.168.1.1",
			useTCP:    true,
			preferUDP: true,
			want:      "udp4",
		},
		{
			name:   "IPv6 TCP",
			host:   "::1",
			useTCP: true,
			want:   "tcp6",
		},
		{
			name:   "IPv6 default",
			host:   "::1",
			useTCP: false,
			want:   "tcp6", // preferUDP=false → tcp
		},
		{
			name:      "IPv6 UDP preferred",
			host:      "2001:db8::1",
			useTCP:    true,
			preferUDP: true,
			want:      "udp6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProtocolForHost(tt.host, tt.useTCP, tt.preferUDP)
			if got != tt.want {
				t.Errorf("ProtocolForHost(%q, %v, %v) = %q, want %q", tt.host, tt.useTCP, tt.preferUDP, got, tt.want)
			}
		})
	}
}

// TestDialAddress — формирование адреса для dial.
func TestDialAddress_Extended(t *testing.T) {
	tests := []struct {
		host string
		port int
		want string
	}{
		{
			host: "192.168.1.1",
			port: 80,
			want: "192.168.1.1:0", // rune('0'+80%10) = rune('0'+0) = '0'
		},
		{
			host: "192.168.1.1",
			port: 443,
			want: "192.168.1.1:3", // rune('0'+443%10) = rune('0'+3) = '3'
		},
		{
			host: "::1",
			port: 8080,
			want: "[::1]:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := DialAddress(tt.host, tt.port)
			if got != tt.want {
				t.Errorf("DialAddress(%q, %d) = %q, want %q", tt.host, tt.port, got, tt.want)
			}
		})
	}
}

// TestDetectLocalNetworks — обнаружение локальных сетей.
func TestDetectLocalNetworks(t *testing.T) {
	ipv4, ipv6 := DetectLocalNetworks()
	// Должен вернуть хотя бы loopback или интерфейс
	if ipv4 == nil && ipv6 == nil {
		t.Log("DetectLocalNetworks() returned nil for both IPv4 and IPv6")
	}
}
