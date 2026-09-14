package network

import (
	"net"
	"testing"
)

// TestIsIPv4 проверяет определение IPv4.
func TestIsIPv4(t *testing.T) {
	tests := []struct {
		ip     string
		expect bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"127.0.0.1", true},
		{"::1", false},
		{"2001:db8::1", false},
		{"fe80::1", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		result := IsIPv4(tt.ip)
		if result != tt.expect {
			t.Errorf("IsIPv4(%q) = %v, want %v", tt.ip, result, tt.expect)
		}
	}
}

// TestIsIPv6 проверяет определение IPv6.
func TestIsIPv6(t *testing.T) {
	tests := []struct {
		ip     string
		expect bool
	}{
		{"::1", true},
		{"2001:db8::1", true},
		{"fe80::1", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"192.168.1.1", false},
		{"10.0.0.1", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		result := IsIPv6(tt.ip)
		if result != tt.expect {
			t.Errorf("IsIPv6(%q) = %v, want %v", tt.ip, result, tt.expect)
		}
	}
}

// TestIsIP проверяет валидацию IP.
func TestIsIP(t *testing.T) {
	tests := []struct {
		ip     string
		expect bool
	}{
		{"192.168.1.1", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"", false},
		{"invalid", false},
		{"256.256.256.256", false},
	}

	for _, tt := range tests {
		result := IsIP(tt.ip)
		if result != tt.expect {
			t.Errorf("IsIP(%q) = %v, want %v", tt.ip, result, tt.expect)
		}
	}
}

// TestIsCIDR проверяет валидацию CIDR.
func TestIsCIDR(t *testing.T) {
	tests := []struct {
		cidr   string
		expect bool
	}{
		{"192.168.1.0/24", true},
		{"10.0.0.0/8", true},
		{"2001:db8::/32", true},
		{"fe80::/10", true},
		{"192.168.1.1/32", true},
		{"::1/128", true},
		{"invalid", false},
		{"192.168.1.0/33", false},
		{"2001:db8::/129", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsCIDR(tt.cidr)
		if result != tt.expect {
			t.Errorf("IsCIDR(%q) = %v, want %v", tt.cidr, result, tt.expect)
		}
	}
}

// TestPreferredNetwork проверяет выбор предпочтительной сети.
func TestPreferredNetwork(t *testing.T) {
	tests := []struct {
		ipv4   []string
		ipv6   []string
		expect string
	}{
		{[]string{"192.168.1.0/24"}, nil, "192.168.1.0/24"},
		{nil, []string{"2001:db8::/64"}, "2001:db8::/64"},
		{[]string{"192.168.1.0/24"}, []string{"2001:db8::/64"}, "192.168.1.0/24"},
		{nil, nil, ""},
	}

	for _, tt := range tests {
		result := PreferredNetwork(tt.ipv4, tt.ipv6)
		if result != tt.expect {
			t.Errorf("PreferredNetwork(%v, %v) = %q, want %q", tt.ipv4, tt.ipv6, result, tt.expect)
		}
	}
}

// TestFormatIPForDisplay проверяет форматирование IP для отображения.
func TestFormatIPForDisplay(t *testing.T) {
	tests := []struct {
		ip     string
		expect string
	}{
		{"192.168.1.1", "192.168.1.1"},
		{"10.0.0.1", "10.0.0.1"},
		{"::1", "[::1]"},
		{"2001:db8::1", "[2001:db8::1]"},
		{"", ""},
	}

	for _, tt := range tests {
		result := FormatIPForDisplay(tt.ip)
		if result != tt.expect {
			t.Errorf("FormatIPForDisplay(%q) = %q, want %q", tt.ip, result, tt.expect)
		}
	}
}

// TestParseIPv6NetworkRange проверяет парсинг IPv6 диапазонов.
func TestParseIPv6NetworkRange(t *testing.T) {
	tests := []struct {
		cidr      string
		wantCount int
		wantFirst string
	}{
		{"2001:db8::/120", 256, "2001:db8::"},
		{"::1/128", 1, "::1"},
		{"fe80::/128", 1, "fe80::"},
	}

	for _, tt := range tests {
		ips, err := ParseNetworkRange(tt.cidr)
		if err != nil {
			t.Errorf("ParseNetworkRange(%q) error: %v", tt.cidr, err)
			continue
		}
		if len(ips) != tt.wantCount {
			t.Errorf("ParseNetworkRange(%q) count = %d, want %d", tt.cidr, len(ips), tt.wantCount)
		}
		if len(ips) > 0 {
			first := ips[0].String()
			if first != tt.wantFirst {
				t.Errorf("ParseNetworkRange(%q) first = %q, want %q", tt.cidr, first, tt.wantFirst)
			}
		}
	}
}

// TestParseIPv4NetworkRange проверяет парсинг IPv4 диапазонов.
func TestParseIPv4NetworkRange(t *testing.T) {
	tests := []struct {
		cidr      string
		wantCount int
		wantFirst string
	}{
		{"192.168.1.0/24", 254, "192.168.1.1"},
		{"10.0.0.0/24", 254, "10.0.0.1"},
		{"192.168.1.1/32", 0, ""}, // /32: network=broadcast, 0 адресов
	}

	for _, tt := range tests {
		ips, err := ParseNetworkRange(tt.cidr)
		if err != nil {
			t.Errorf("ParseNetworkRange(%q) error: %v", tt.cidr, err)
			continue
		}
		if len(ips) != tt.wantCount {
			t.Errorf("ParseNetworkRange(%q) count = %d, want %d", tt.cidr, len(ips), tt.wantCount)
		}
		if len(ips) > 0 {
			first := ips[0].String()
			if first != tt.wantFirst {
				t.Errorf("ParseNetworkRange(%q) first = %q, want %q", tt.cidr, first, tt.wantFirst)
			}
		}
	}
}

// TestDetectLocalNetworkDual проверяет dual-stack detection.
func TestDetectLocalNetworkDual(t *testing.T) {
	ipv4, ipv6, err := DetectLocalNetworkDual()
	if err != nil {
		t.Skipf("DetectLocalNetworkDual failed (expected in some environments): %v", err)
	}

	// Хотя бы одна сеть должна быть обнаружена
	if len(ipv4) == 0 && len(ipv6) == 0 {
		t.Error("Expected at least one network (IPv4 or IPv6) to be detected")
	}

	// Проверим валидность обнаруженных CIDR
	for _, cidr := range ipv4 {
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			t.Errorf("Detected invalid IPv4 CIDR: %q", cidr)
		}
	}
	for _, cidr := range ipv6 {
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			t.Errorf("Detected invalid IPv6 CIDR: %q", cidr)
		}
	}
}

// TestEstimateHostCountIPv6 проверяет оценку хостов для IPv6.
func TestEstimateHostCountIPv6(t *testing.T) {
	tests := []struct {
		cidr      string
		wantCount int
		wantErr   bool
	}{
		{"2001:db8::/64", 0, true},       // слишком большой (hostBits=64 > 30)
		{"2001:db8::/96", 0, true},       // hostBits=32 > 30
		{"2001:db8::/112", 65536, false}, // hostBits=16, 1<<16=65536
		{"::1/128", 1, false},
	}

	for _, tt := range tests {
		count, err := EstimateHostCount(tt.cidr)
		if tt.wantErr && err == nil {
			t.Errorf("EstimateHostCount(%q) expected error, got none", tt.cidr)
			continue
		}
		if !tt.wantErr && err != nil {
			t.Errorf("EstimateHostCount(%q) unexpected error: %v", tt.cidr, err)
			continue
		}
		if count != tt.wantCount {
			t.Errorf("EstimateHostCount(%q) = %d, want %d", tt.cidr, count, tt.wantCount)
		}
	}
}
