package scanner

import (
	"net"
	"testing"
	"time"
)

// TestPortThreadsForHost_Budget — проверка расчёта threads.
func TestPortThreadsForHost_Budget(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-65535", 128, false)
	threads := ns.portThreadsForHost(1000)
	if threads > maxPerHostPortThreads {
		t.Errorf("portThreadsForHost(1000) = %d, want <= %d", threads, maxPerHostPortThreads)
	}
	if threads < minPerHostPortThreads {
		t.Errorf("portThreadsForHost(1000) = %d, want >= %d", threads, minPerHostPortThreads)
	}
}

// TestPortThreadsForHost_Cap — threads не превышает max.
func TestPortThreadsForHost_Cap(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-65535", 1, false)
	threads := ns.portThreadsForHost(10000)
	if threads != maxPerHostPortThreads {
		t.Errorf("portThreadsForHost(10000) = %d, want %d", threads, maxPerHostPortThreads)
	}
}

// TestPortThreadsForHost_PortCountCapped — threads не выше portCount.
func TestPortThreadsForHost_PortCountCapped(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-65535", 256, false)
	// 5 портов → threads = min(calculated, 5) = 5
	threads := ns.portThreadsForHost(5)
	if threads != 5 {
		t.Errorf("portThreadsForHost(5) = %d, want 5", threads)
	}
}

// TestScanConfig_BannersEnabled — проверка флага banner grabbing.
func TestScanConfig_BannersEnabled(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "80,443", 10, false)
	ns.SetGrabBanners(true)
	if !ns.grabBanners {
		t.Error("SetGrabBanners(true) not applied")
	}
}

// TestScanConfig_UDPEnabled — проверка UDP флага.
func TestScanConfig_UDPEnabled(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "53,67,68", 10, false)
	ns.SetScanUDP(true)
	if !ns.scanUDP {
		t.Error("SetScanUDP(true) not applied")
	}
}

// TestScanConfig_TCPEnabled — проверка TCP флага.
func TestScanConfig_TCPEnabled(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 10, false)
	ns.SetScanTCPPorts(true)
	if !ns.scanTCPPorts {
		t.Error("SetScanTCPPorts(true) not applied")
	}
}

// TestScanConfig_OSDetect — проверка OS detect active.
func TestScanConfig_OSDetect(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 10, false)
	ns.SetOSDetectActive(true)
	if !ns.osDetectActive {
		t.Error("SetOSDetectActive(true) not applied")
	}
}

// TestScanConfig_Verbose — проверка verbose logs.
func TestScanConfig_Verbose(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 10, false)
	ns.SetVerbosePortLogs(true)
	if !ns.verbosePortLogs {
		t.Error("SetVerbosePortLogs(true) not applied")
	}
}

// TestNetworkScanner_GetDiagnosticsSummary — проверка диагностики.
func TestNetworkScanner_GetDiagnosticsSummary(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	diag := ns.GetDiagnosticsSummary()
	if diag == "" {
		t.Error("GetDiagnosticsSummary() returned empty string")
	}
}

// TestNetworkScanner_Stop — остановка сканера.
func TestNetworkScanner_Stop(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	ns.Stop()
}

// TestNetworkScanner_GetResults_Empty — пустые результаты.
func TestNetworkScanner_GetResults_Empty(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", 3*time.Second, "1-1000", 100, false)
	results := ns.GetResults()
	if len(results) != 0 {
		t.Errorf("GetResults() = %d results, want 0", len(results))
	}
}

// TestParseIP_Valid — парсинг валидного IP.
func TestParseIP_Valid(t *testing.T) {
	tests := []string{
		"192.168.1.1",
		"10.0.0.1",
		"255.255.255.255",
		"0.0.0.0",
	}

	for _, ipStr := range tests {
		t.Run(ipStr, func(t *testing.T) {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				t.Errorf("net.ParseIP(%s) = nil", ipStr)
			}
		})
	}
}

// TestParseIP_Invalid — парсинг невалидного IP.
func TestParseIP_Invalid(t *testing.T) {
	ip := net.ParseIP("invalid-ip")
	if ip != nil {
		t.Errorf("net.ParseIP(invalid-ip) = %v, want nil", ip)
	}
}

// TestIsIPInCIDR — проверка IP в CIDR.
func TestIsIPInCIDR(t *testing.T) {
	tests := []struct {
		ip   string
		cidr string
		want bool
	}{
		{"192.168.1.1", "192.168.1.0/24", true},
		{"192.168.1.255", "192.168.1.0/24", true},
		{"192.168.2.1", "192.168.1.0/24", false},
		{"10.0.0.1", "10.0.0.0/16", true},
		{"10.0.255.255", "10.0.0.0/16", true},
		{"10.1.0.1", "10.0.0.0/16", false}, // /16 = 10.0.0.0-10.0.255.255
		{"10.1.0.1", "10.1.0.0/16", true},
		{"10.2.0.1", "10.0.0.0/16", false},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			_, cidrNet, _ := net.ParseCIDR(tt.cidr)
			got := cidrNet.Contains(ip)
			if got != tt.want {
				t.Errorf("CIDR.Contains(%s, %s) = %v, want %v", tt.cidr, tt.ip, got, tt.want)
			}
		})
	}
}
