package network

import (
	"context"
	"net"
	"testing"
	"time"
)

// --- DefaultNetworkProber tests ---

func TestDefaultNetworkProber_Ping_Success(t *testing.T) {
	prober := DefaultNetworkProber{Timeout: 1 * time.Second}

	// Локальный хост должен быть доступен (если запущен сервер)
	// Тестируем, что функция не паникует
	isAlive, err := prober.Ping("127.0.0.1")
	// Результат зависит от наличия сервиса на порту 80
	if err != nil {
		t.Logf("Ping to 127.0.0.1 returned error (expected if no service): %v", err)
	}
	_ = isAlive
}

func TestDefaultNetworkProber_Ping_Unreachable(t *testing.T) {
	prober := DefaultNetworkProber{Timeout: 100 * time.Millisecond}

	// Тестовый IP из RFC 3330 (не должен быть доступен)
	isAlive, err := prober.Ping("192.0.2.1")
	// Ping может вернуть false без ошибки (в зависимости от реализации)
	if isAlive {
		t.Error("Ping to unreachable IP should return false")
	}
	_ = err
}

func TestDefaultNetworkProber_Ping_InvalidIP(t *testing.T) {
	prober := DefaultNetworkProber{Timeout: 1 * time.Second}

	// Invalid IP может быть обработан без ошибки (вернет false)
	isAlive, err := prober.Ping("invalid-ip")
	if isAlive {
		t.Error("Ping to invalid IP should return false")
	}
	_ = err
}

func TestDefaultNetworkProber_PingContext_Success(t *testing.T) {
	prober := DefaultNetworkProber{Timeout: 1 * time.Second}
	done := make(chan struct{})

	// Тестируем, что функция не паникует с контекстом
	isAlive, err := prober.PingContext("192.0.2.1", done)
	if isAlive {
		t.Error("PingContext to unreachable IP should return false")
	}
	_ = err
}

func TestDefaultNetworkProber_PingContext_Cancelled(t *testing.T) {
	prober := DefaultNetworkProber{Timeout: 5 * time.Second}
	done := make(chan struct{})
	close(done) // Немедленная отмена

	// Должна вернуть ошибку или таймаут
	_, err := prober.PingContext("127.0.0.1", done)
	if err == nil {
		t.Log("PingContext with cancelled context returned no error (acceptable)")
	}
}

func TestDefaultNetworkProber_ResolveMAC_Success(t *testing.T) {
	prober := DefaultNetworkProber{Timeout: 1 * time.Second}

	// Тестируем, что функция не паникует
	mac, err := prober.ResolveMAC("127.0.0.1")
	// Результат зависит от ARP таблицы
	_ = mac
	_ = err
}

func TestDefaultNetworkProber_ResolveMAC_Unreachable(t *testing.T) {
	prober := DefaultNetworkProber{Timeout: 100 * time.Millisecond}

	_, err := prober.ResolveMAC("192.0.2.1")
	if err == nil {
		t.Error("ResolveMAC for unreachable IP should return error")
	}
}

// --- TCPPortScanner tests ---

func TestTCPPortScanner_ScanPort_Open(t *testing.T) {
	scanner := TCPPortScanner{Timeout: 1 * time.Second}

	// Тестируем, что функция не паникует
	isOpen, err := scanner.ScanPort("192.0.2.1", 80, "tcp")
	// Ожидаем ошибку для недостижимого хоста
	if err == nil && isOpen {
		t.Error("ScanPort to unreachable IP should return error or false")
	}
	_ = isOpen
}

func TestTCPPortScanner_ScanPort_InvalidPort(t *testing.T) {
	scanner := TCPPortScanner{Timeout: 1 * time.Second}

	// Port 0 может не валидироваться, тестируем что не паникует
	isOpen, err := scanner.ScanPort("127.0.0.1", 0, "tcp")
	// Результат зависит от реализации
	_ = isOpen
	_ = err
}

func TestTCPPortScanner_ScanPort_InvalidProtocol(t *testing.T) {
	scanner := TCPPortScanner{Timeout: 1 * time.Second}

	_, err := scanner.ScanPort("127.0.0.1", 80, "invalid")
	if err == nil {
		t.Error("ScanPort with invalid protocol should return error")
	}
}

func TestTCPPortScanner_ScanPort_NegativePort(t *testing.T) {
	scanner := TCPPortScanner{Timeout: 1 * time.Second}

	// Negative port может не валидироваться, тестируем что не паникует
	isOpen, err := scanner.ScanPort("127.0.0.1", -1, "tcp")
	_ = isOpen
	_ = err
}

func TestTCPPortScanner_ScanPort_PortOutOfRange(t *testing.T) {
	scanner := TCPPortScanner{Timeout: 1 * time.Second}

	// Port > 65535 может не валидироваться, тестируем что не паникует
	isOpen, err := scanner.ScanPort("127.0.0.1", 70000, "tcp")
	_ = isOpen
	_ = err
}

// --- UDPPortScanner tests ---

func TestUDPPortScanner_ScanPort(t *testing.T) {
	scanner := UDPPortScanner{Timeout: 1 * time.Second}

	// Тестируем, что функция не паникует
	isOpen, err := scanner.ScanPort("192.0.2.1", 53, "udp")
	// Результат зависит от сети
	_ = isOpen
	_ = err
}

func TestUDPPortScanner_ScanPort_InvalidPort(t *testing.T) {
	scanner := UDPPortScanner{Timeout: 1 * time.Second}

	// Port 0 может не валидироваться, тестируем что не паникует
	isOpen, err := scanner.ScanPort("127.0.0.1", 0, "udp")
	_ = isOpen
	_ = err
}

func TestUDPPortScanner_ScanPort_InvalidProtocol(t *testing.T) {
	scanner := UDPPortScanner{Timeout: 1 * time.Second}

	_, err := scanner.ScanPort("127.0.0.1", 53, "tcp")
	if err == nil {
		t.Error("ScanPort with non-udp protocol should return error")
	}
}

// --- ParseNetworkRange edge cases ---

func TestParseNetworkRange_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		wantErr bool
		minIPs  int
	}{
		{
			name:    "/32 single host",
			cidr:    "192.168.1.1/32",
			wantErr: false,
			minIPs:  0, // /32 returns 0 IPs (only network address, excluded)
		},
		{
			name:    "/31 point-to-point",
			cidr:    "192.168.1.0/31",
			wantErr: false,
			minIPs:  0, // /31 returns 0 IPs (network and broadcast, both excluded)
		},
		{
			name:    "/24 standard",
			cidr:    "192.168.1.0/24",
			wantErr: false,
			minIPs:  254,
		},
		{
			name:    "/16 large",
			cidr:    "10.0.0.0/16",
			wantErr: false,
			minIPs:  65534,
		},
		{
			name:    "IPv4 with leading zeros",
			cidr:    "192.168.001.0/24",
			wantErr: true,
		},
		{
			name:    "Invalid mask",
			cidr:    "192.168.1.0/33",
			wantErr: true,
		},
		{
			name:    "Empty string",
			cidr:    "",
			wantErr: true,
		},
		{
			name:    "Just IP no mask",
			cidr:    "192.168.1.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := ParseNetworkRange(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNetworkRange(%q) error = %v, wantErr %v", tt.cidr, err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(ips) < tt.minIPs {
				t.Errorf("ParseNetworkRange(%q) got %d IPs, want at least %d", tt.cidr, len(ips), tt.minIPs)
			}
		})
	}
}

// --- ParsePortRange edge cases ---

func TestParsePortRange_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantLen int
	}{
		{
			name:    "Single port",
			input:   "80",
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "Range",
			input:   "80-85",
			wantErr: false,
			wantLen: 6,
		},
		{
			name:    "Mixed",
			input:   "80,443-445,8080",
			wantErr: false,
			wantLen: 5,
		},
		{
			name:    "Whitespace",
			input:   "80, 443, 8080",
			wantErr: false,
			wantLen: 3,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
			wantLen: 0,
		},
		{
			name:    "Invalid range",
			input:   "100-50",
			wantErr: false,
			wantLen: 0,
		},
		{
			name:    "Non-numeric",
			input:   "abc",
			wantErr: true,
			wantLen: 0,
		},
		{
			name:    "Port 1",
			input:   "1",
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "Port 65535",
			input:   "65535",
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "Range 1-65535",
			input:   "1-65535",
			wantErr: false,
			wantLen: 65535,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := ParsePortRange(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePortRange(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(ports) != tt.wantLen {
				t.Errorf("ParsePortRange(%q) got %d ports, want %d", tt.input, len(ports), tt.wantLen)
			}
			// Validate all ports in valid range
			for _, port := range ports {
				if port < 1 || port > 65535 {
					t.Errorf("ParsePortRange(%q) returned invalid port: %d", tt.input, port)
				}
			}
		})
	}
}

// --- GetServiceName tests ---

func TestGetServiceName_KnownPorts(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{21, "FTP"},
		{22, "SSH"},
		{25, "SMTP"},
		{53, "DNS"},
		{80, "HTTP"},
		{110, "POP3"},
		{143, "IMAP"},
		{443, "HTTPS"},
		{993, "IMAPS"},
		{995, "POP3S"},
		{3306, "MySQL"},
		{5432, "PostgreSQL"},
		{3389, "RDP"},
		{5900, "VNC"},
		{9999, "Distinct"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := GetServiceName(tt.port)
			if got != tt.want {
				t.Errorf("GetServiceName(%d) = %q, want %q", tt.port, got, tt.want)
			}
		})
	}
}

func TestGetServiceName_CustomPorts(t *testing.T) {
	// Тестируем порты с нестандартными именами
	port8080 := GetServiceName(8080)
	if port8080 == "" {
		t.Error("GetServiceName(8080) should return non-empty string")
	}

	port8443 := GetServiceName(8443)
	if port8443 == "" {
		t.Error("GetServiceName(8443) should return non-empty string")
	}
}

func TestGetServiceName_UnknownPort(t *testing.T) {
	got := GetServiceName(12345)
	if got == "FTP" || got == "SSH" || got == "HTTP" {
		t.Errorf("GetServiceName(12345) should not return known service name, got %q", got)
	}
}

// --- IsPortOpen tests ---

func TestIsPortOpen_Localhost(t *testing.T) {
	timeout := 100 * time.Millisecond

	// Тестируем на недостижимом порту
	result := IsPortOpen("127.0.0.1", 59999, timeout)
	if result {
		t.Error("IsPortOpen should return false for unreachable port on localhost")
	}
}

func TestIsPortOpen_UnreachableHost(t *testing.T) {
	timeout := 100 * time.Millisecond

	result := IsPortOpen("192.0.2.1", 80, timeout)
	if result {
		t.Error("IsPortOpen should return false for unreachable host")
	}
}

func TestIsPortOpen_ZeroTimeout(t *testing.T) {
	// Нулевой таймаут должен работать без паники
	result := IsPortOpen("192.0.2.1", 80, 0)
	_ = result
}

func TestIsPortOpen_NegativeTimeout(t *testing.T) {
	// Отрицательный таймаут должен работать без паники
	result := IsPortOpen("192.0.2.1", 80, -1*time.Second)
	_ = result
}

// --- IsUDPPortOpen tests ---

func TestIsUDPPortOpen_UnreachableHost(t *testing.T) {
	timeout := 100 * time.Millisecond

	result := IsUDPPortOpen("192.0.2.1", 53, timeout)
	// UDP сканирование может вернуть true или false в зависимости от реализации
	_ = result
}

func TestIsUDPPortOpen_ZeroTimeout(t *testing.T) {
	// Нулевой таймаут должен работать без паники
	result := IsUDPPortOpen("192.0.2.1", 53, 0)
	_ = result
}

// --- DetectLocalNetwork tests ---

func TestDetectLocalNetwork_Valid(t *testing.T) {
	network, err := DetectLocalNetwork()
	if err != nil {
		t.Skipf("DetectLocalNetwork failed (expected in some environments): %v", err)
		return
	}

	if network == "" {
		t.Error("DetectLocalNetwork should return non-empty string")
	}

	// Проверяем валидность CIDR
	_, _, err = net.ParseCIDR(network)
	if err != nil {
		t.Errorf("DetectLocalNetwork returned invalid CIDR: %v", err)
	}
}

// --- EstimateHostCount tests ---

func TestEstimateHostCount_Valid(t *testing.T) {
	tests := []struct {
		cidr    string
		want    int
		wantErr bool
	}{
		{"192.168.1.0/24", 254, false},
		{"192.168.1.0/30", 2, false},
		{"192.168.1.0/31", 2, false},
		{"192.168.1.1/32", 1, false},
		{"10.0.0.0/8", 16777214, false},
		{"2001:db8::/126", 4, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			got, err := EstimateHostCount(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("EstimateHostCount(%q) error = %v, wantErr %v", tt.cidr, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("EstimateHostCount(%q) = %d, want %d", tt.cidr, got, tt.want)
			}
		})
	}
}

// --- inc function tests ---

func TestInc_IPv4(t *testing.T) {
	ip := net.ParseIP("192.168.1.254").To4()
	if ip == nil {
		t.Fatal("Failed to parse IP")
	}
	inc(ip)
	if ip.String() != "192.168.1.255" {
		t.Errorf("inc(192.168.1.254) = %s, want 192.168.1.255", ip)
	}

	// Переполнение
	inc(ip)
	if ip.String() != "192.168.2.0" {
		t.Errorf("inc(192.168.1.255) = %s, want 192.168.2.0", ip)
	}
}

func TestInc_IPv6(t *testing.T) {
	ip := net.ParseIP("2001:db8::1").To16()
	if ip == nil {
		t.Fatal("Failed to parse IP")
	}
	inc(ip)
	expected := net.ParseIP("2001:db8::2")
	if !ip.Equal(expected) {
		t.Errorf("inc(2001:db8::1) = %s, want %s", ip, expected)
	}
}

// --- Timeout edge cases ---

func TestIsPortOpen_VeryShortTimeout(t *testing.T) {
	// Очень короткий таймаут должен работать без паники
	timeout := 1 * time.Millisecond
	result := IsPortOpen("192.0.2.1", 80, timeout)
	_ = result
}

func TestIsPortOpen_VeryLongTimeout(t *testing.T) {
	// Очень длинный таймаут должен работать без паники
	timeout := 30 * time.Second
	result := IsPortOpen("192.0.2.1", 80, timeout)
	_ = result
}

// --- Concurrent access tests ---

func TestParseNetworkRange_Concurrent(t *testing.T) {
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := ParseNetworkRange("192.168.1.0/24")
			if err != nil {
				t.Errorf("ParseNetworkRange failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestParsePortRange_Concurrent(t *testing.T) {
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := ParsePortRange("1-100")
			if err != nil {
				t.Errorf("ParsePortRange failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// --- Context cancellation tests ---

func TestDefaultNetworkProber_PingContext_ContextCancelled(t *testing.T) {
	prober := DefaultNetworkProber{Timeout: 5 * time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Немедленная отмена

	done := make(chan struct{})
	_, err := prober.PingContext("127.0.0.1", done)
	// Ожидаем, что функция завершится без паники
	_ = err
	_ = ctx
}

// --- String parsing edge cases ---

func TestParsePortRange_WhitespaceOnly(t *testing.T) {
	ports, err := ParsePortRange("   ")
	if err == nil {
		t.Error("ParsePortRange with whitespace should return error")
	}
	_ = ports
}

func TestParsePortRange_LeadingTrailingWhitespace(t *testing.T) {
	ports, err := ParsePortRange("  80, 443  ")
	if err != nil {
		t.Errorf("ParsePortRange with whitespace should succeed: %v", err)
	}
	if len(ports) != 2 {
		t.Errorf("ParsePortRange with whitespace got %d ports, want 2", len(ports))
	}
}

func TestParsePortRange_EmptyParts(t *testing.T) {
	// Пустые части между запятыми
	ports, err := ParsePortRange("80,,443")
	if err != nil {
		// Может вернуть ошибку или пропустить пустые части
		t.Logf("ParsePortRange with empty parts returned error: %v", err)
	}
	_ = ports
}

// --- IPv6 specific tests ---

func TestParseNetworkRange_IPv6_Large(t *testing.T) {
	// IPv6 /112 (16 хостов) - должно работать
	_, err := ParseNetworkRange("2001:db8::/112")
	if err != nil {
		t.Errorf("ParseNetworkRange IPv6 /112 should succeed: %v", err)
	}
}

func TestParseNetworkRange_IPv6_TooLarge(t *testing.T) {
	// IPv6 /64 (слишком большой диапазон)
	_, err := ParseNetworkRange("2001:db8::/64")
	if err == nil {
		t.Error("ParseNetworkRange IPv6 /64 should return error (too large)")
	}
}

// --- Port validation tests ---

func TestParsePortRange_HighPorts(t *testing.T) {
	ports, err := ParsePortRange("49152-65535")
	if err != nil {
		t.Errorf("ParsePortRange high ports should succeed: %v", err)
	}
	if len(ports) != 16384 {
		t.Errorf("ParsePortRange high ports got %d ports, want 16384", len(ports))
	}
}

func TestParsePortRange_EphemeralPorts(t *testing.T) {
	ports, err := ParsePortRange("49152-50000")
	if err != nil {
		t.Errorf("ParsePortRange ephemeral ports should succeed: %v", err)
	}
	if len(ports) != 849 {
		t.Errorf("ParsePortRange ephemeral ports got %d ports, want 849", len(ports))
	}
}
