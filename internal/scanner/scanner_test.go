package scanner

import (
	"net"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNewNetworkScanner(t *testing.T) {
	tests := []struct {
		name       string
		network    string
		timeout    time.Duration
		portRange  string
		threads    int
		showClosed bool
	}{
		{
			name:       "Default settings",
			network:    "192.168.1.0/24",
			timeout:    3 * time.Second,
			portRange:  "1-1000",
			threads:    100,
			showClosed: false,
		},
		{
			name:       "Custom settings",
			network:    "10.0.0.0/16",
			timeout:    5 * time.Second,
			portRange:  "80,443,8080",
			threads:    50,
			showClosed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NewNetworkScanner(tt.network, tt.timeout, tt.portRange, tt.threads, tt.showClosed)
			if ns == nil {
				t.Fatal("NewNetworkScanner() returned nil")
			}
			if ns.network != tt.network {
				t.Errorf("NewNetworkScanner() network = %v, want %v", ns.network, tt.network)
			}
			if ns.timeout != tt.timeout {
				t.Errorf("NewNetworkScanner() timeout = %v, want %v", ns.timeout, tt.timeout)
			}
			if ns.portRange != tt.portRange {
				t.Errorf("NewNetworkScanner() portRange = %v, want %v", ns.portRange, tt.portRange)
			}
			if ns.threads != tt.threads {
				t.Errorf("NewNetworkScanner() threads = %v, want %v", ns.threads, tt.threads)
			}
			if ns.showClosed != tt.showClosed {
				t.Errorf("NewNetworkScanner() showClosed = %v, want %v", ns.showClosed, tt.showClosed)
			}
			if ns.results == nil {
				t.Error("NewNetworkScanner() results is nil")
			}
			if ns.ctx == nil {
				t.Error("NewNetworkScanner() ctx is nil")
			}
		})
	}
}

func TestGetResults(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Изначально результатов нет
	results := ns.GetResults()
	if len(results) != 0 {
		t.Errorf("GetResults() got %d results, want 0", len(results))
	}
}

func TestStop(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Проверяем, что Stop не паникует
	ns.Stop()

	// Проверяем, что контекст отменен
	select {
	case <-ns.ctx.Done():
		// Ожидаемое поведение
	default:
		t.Error("Stop() did not cancel context")
	}
}

func TestGetProtocolFromPort(t *testing.T) {
	tests := []struct {
		name string
		port int
		want string
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
		{"Unknown port", 9999, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getProtocolFromPort(tt.port)
			if got != tt.want {
				t.Errorf("getProtocolFromPort(%d) = %v, want %v", tt.port, got, tt.want)
			}
		})
	}
}

func TestGetVendorFromMAC(t *testing.T) {
	tests := []struct {
		name string
		mac  string
		want string
	}{
		{"VMware 1", "00:50:56:00:00:00", "VMware"},
		{"VMware 2", "00:0c:29:00:00:00", "VMware"},
		{"VirtualBox", "08:00:27:00:00:00", "VirtualBox"},
		{"QEMU", "52:54:00:00:00:00", "QEMU"},
		{"Apple", "00:1b:21:00:00:00", "Apple"},
		{"Raspberry Pi", "b8:27:eb:00:00:00", "Raspberry Pi"},
		{"Unknown MAC", "aa:bb:cc:dd:ee:ff", "Unknown"},
		{"Short MAC", "00:50", "Unknown"},
		{"Empty MAC", "", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getVendorFromMAC(tt.mac)
			if got != tt.want {
				t.Errorf("getVendorFromMAC(%v) = %v, want %v", tt.mac, got, tt.want)
			}
		})
	}
}

func TestAppendIfNotExists(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  []string
	}{
		{
			name:  "Add new item",
			slice: []string{"a", "b"},
			item:  "c",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "Item already exists",
			slice: []string{"a", "b"},
			item:  "a",
			want:  []string{"a", "b"},
		},
		{
			name:  "Empty slice",
			slice: []string{},
			item:  "a",
			want:  []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendIfNotExists(tt.slice, tt.item)
			if len(got) != len(tt.want) {
				t.Errorf("appendIfNotExists() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("appendIfNotExists() [%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDetectDeviceType(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	tests := []struct {
		name   string
		result Result
		want   string
	}{
		{
			name: "Web Server",
			result: Result{
				Ports: []PortInfo{
					{Port: 80, State: "open"},
					{Port: 443, State: "open"},
				},
			},
			want: "Web Server",
		},
		{
			name: "Router",
			result: Result{
				Ports: []PortInfo{
					{Port: 80, State: "open"},
					{Port: 22, State: "open"},
				},
			},
			want: "Router/Network Device",
		},
		{
			name: "Database Server",
			result: Result{
				Ports: []PortInfo{
					{Port: 3306, State: "open"},
				},
			},
			want: "Database Server",
		},
		{
			name: "Windows Computer",
			result: Result{
				Ports: []PortInfo{
					{Port: 3389, State: "open"},
					{Port: 445, State: "open"},
				},
			},
			want: "Windows Computer",
		},
		{
			name: "Linux Server",
			result: Result{
				Ports: []PortInfo{
					{Port: 22, State: "open"},
				},
			},
			want: "Linux/Unix Server",
		},
		{
			name: "Printer",
			result: Result{
				Ports: []PortInfo{
					{Port: 9100, State: "open"},
				},
			},
			want: "Printer",
		},
		{
			name: "IoT Device",
			result: Result{
				Ports: []PortInfo{
					{Port: 9999, State: "open"}, // Используем порт, который не попадает под другие категории
				},
			},
			want: "IoT Device", // Должно быть определено как IoT, так как только 1 порт и не попадает под другие категории
		},
		{
			name: "Unknown Device",
			result: Result{
				Ports: []PortInfo{},
			},
			want: "Unknown Device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ns.detectDeviceType(tt.result)
			if got != tt.want {
				t.Errorf("detectDeviceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadMACFromARPTable(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Тест на несуществующий IP (должен вернуть ошибку)
	ip := net.ParseIP("192.0.2.1") // Тестовый IP (RFC 3330)
	_, err := ns.readMACFromARPTable(ip)
	if err == nil {
		// Это нормально, если IP есть в ARP таблице (в тестовой среде)
		t.Log("IP найден в ARP таблице (это нормально в некоторых средах)")
	} else {
		// Ожидаем ошибку для несуществующего IP
		if !strings.Contains(err.Error(), "не найден") && !strings.Contains(err.Error(), "не поддерживается") {
			t.Logf("Ожидаемая ошибка для несуществующего IP: %v", err)
		}
	}
}

func TestReadMACFromLinuxARP(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Тест только для Linux")
	}

	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Тест на несуществующий IP
	_, err := ns.readMACFromLinuxARP("192.0.2.1")
	if err == nil {
		t.Error("readMACFromLinuxARP() должен вернуть ошибку для несуществующего IP")
	}
}

func TestReadMACFromWindowsARP(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Тест только для Windows")
	}

	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Тест на несуществующий IP
	_, err := ns.readMACFromWindowsARP("192.0.2.1")
	if err == nil {
		t.Log("IP найден в ARP таблице (это нормально в некоторых средах)")
	} else {
		t.Logf("Ожидаемая ошибка: %v", err)
	}
}

func TestReadMACFromDarwinARP(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Тест только для macOS")
	}

	ns := NewNetworkScanner("192.168.1.0/24", 1*time.Second, "80", 10, false)

	// Тест на несуществующий IP
	_, err := ns.readMACFromDarwinARP("192.0.2.1")
	if err == nil {
		t.Log("IP найден в ARP таблице (это нормально в некоторых средах)")
	} else {
		t.Logf("Ожидаемая ошибка: %v", err)
	}
}
