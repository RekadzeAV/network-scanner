package network

import (
	"net"
	"testing"
	"time"
)

func TestParseNetworkRange(t *testing.T) {
	tests := []struct {
		name    string
		network string
		wantErr bool
		minIPs  int
	}{
		{
			name:    "Valid /24 network",
			network: "192.168.1.0/24",
			wantErr: false,
			minIPs:  254, // 256 - 2 (network + broadcast)
		},
		{
			name:    "Valid /16 network",
			network: "10.0.0.0/16",
			wantErr: false,
			minIPs:  65534,
		},
		{
			name:    "Valid /30 network",
			network: "192.168.1.0/30",
			wantErr: false,
			minIPs:  2,
		},
		{
			name:    "Invalid network format",
			network: "192.168.1.0",
			wantErr: true,
		},
		{
			name:    "Invalid CIDR",
			network: "invalid",
			wantErr: true,
		},
		{
			name:    "Empty string",
			network: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := ParseNetworkRange(tt.network)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNetworkRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(ips) < tt.minIPs {
					t.Errorf("ParseNetworkRange() got %d IPs, want at least %d", len(ips), tt.minIPs)
				}
				// Проверяем, что все IP валидны
				for _, ip := range ips {
					if ip == nil {
						t.Error("ParseNetworkRange() returned nil IP")
					}
					if ip.To4() == nil {
						t.Errorf("ParseNetworkRange() returned non-IPv4 IP: %v", ip)
					}
				}
			}
		})
	}
}

func TestParsePortRange(t *testing.T) {
	tests := []struct {
		name      string
		rangeStr  string
		wantErr   bool
		wantLen   int
		wantPorts []int
	}{
		{
			name:      "Single port",
			rangeStr:  "80",
			wantErr:   false,
			wantLen:   1,
			wantPorts: []int{80},
		},
		{
			name:      "Port range",
			rangeStr:  "80-85",
			wantErr:   false,
			wantLen:   6,
			wantPorts: []int{80, 81, 82, 83, 84, 85},
		},
		{
			name:      "Multiple ports",
			rangeStr:  "80,443,8080",
			wantErr:   false,
			wantLen:   3,
			wantPorts: []int{80, 443, 8080},
		},
		{
			name:      "Mixed format",
			rangeStr:  "80,443-445,8080",
			wantErr:   false,
			wantLen:   5,
			wantPorts: []int{80, 443, 444, 445, 8080},
		},
		{
			name:     "Large range",
			rangeStr: "1-100",
			wantErr:  false,
			wantLen:  100,
		},
		{
			name:     "Invalid format - no numbers",
			rangeStr: "abc",
			wantErr:  true,
		},
		{
			name:     "Invalid range - start > end",
			rangeStr: "100-50",
			wantErr:  false, // Парсер не проверяет это
			wantLen:  0,     // Но вернет пустой список
		},
		{
			name:     "Empty string",
			rangeStr: "",
			wantErr:  true, // Пустая строка возвращает ошибку
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := ParsePortRange(tt.rangeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePortRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(ports) != tt.wantLen {
					t.Errorf("ParsePortRange() got %d ports, want %d", len(ports), tt.wantLen)
				}
				if tt.wantPorts != nil {
					for i, wantPort := range tt.wantPorts {
						if i < len(ports) && ports[i] != wantPort {
							t.Errorf("ParsePortRange() ports[%d] = %d, want %d", i, ports[i], wantPort)
						}
					}
				}
				// Проверяем, что все порты в валидном диапазоне
				for _, port := range ports {
					if port < 1 || port > 65535 {
						t.Errorf("ParsePortRange() returned invalid port: %d", port)
					}
				}
			}
		})
	}
}

func TestGetServiceName(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantName string
	}{
		{"HTTP", 80, "HTTP"},
		{"HTTPS", 443, "HTTPS"},
		{"SSH", 22, "SSH"},
		{"FTP", 21, "FTP"},
		{"SMTP", 25, "SMTP"},
		{"DNS", 53, "DNS"},
		{"MySQL", 3306, "MySQL"},
		{"PostgreSQL", 5432, "PostgreSQL"},
		{"RDP", 3389, "RDP"},
		{"Unknown port", 9999, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetServiceName(tt.port)
			if got != tt.wantName {
				t.Errorf("GetServiceName(%d) = %v, want %v", tt.port, got, tt.wantName)
			}
		})
	}
}

func TestIsPortOpen(t *testing.T) {
	// Этот тест требует реального сетевого подключения
	// Тестируем только на localhost, если доступен
	t.Run("Test localhost port 80", func(t *testing.T) {
		// Пропускаем тест, если нет сетевого подключения
		timeout := 1 * time.Second
		result := IsPortOpen("127.0.0.1", 80, timeout)
		// Результат может быть любым, так как зависит от наличия сервера
		_ = result // Просто проверяем, что функция не паникует
	})

	t.Run("Test invalid host", func(t *testing.T) {
		timeout := 100 * time.Millisecond
		result := IsPortOpen("192.0.2.1", 80, timeout) // Тестовый IP (RFC 3330)
		if result {
			t.Error("IsPortOpen() should return false for unreachable host")
		}
	})
}

func TestDetectLocalNetwork(t *testing.T) {
	// Этот тест требует реальной сетевой конфигурации
	network, err := DetectLocalNetwork()
	if err != nil {
		// Это нормально, если нет активных сетевых интерфейсов
		t.Logf("DetectLocalNetwork() returned error (expected in some environments): %v", err)
		return
	}

	// Проверяем, что возвращенная сеть валидна
	if network == "" {
		t.Error("DetectLocalNetwork() returned empty string")
	}

	// Проверяем, что это валидный CIDR
	_, _, err = net.ParseCIDR(network)
	if err != nil {
		t.Errorf("DetectLocalNetwork() returned invalid CIDR: %v", err)
	}
}
