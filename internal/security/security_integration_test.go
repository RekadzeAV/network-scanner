package security

import (
	"context"
	"testing"

	"network-scanner/internal/contracts"
)

// === Integration: calculateSecurityIndex ===

func TestIntegrationCalculateSecurityIndex_NoFindings(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{})
	if score != 100 {
		t.Errorf("expected 100 for no findings, got %d", score)
	}
}

func TestIntegrationCalculateSecurityIndex_CriticalOnly(t *testing.T) {
	severityCounts := map[string]int{
		"critical": 1,
	}
	score := calculateSecurityIndex(severityCounts)
	if score != 70 {
		t.Errorf("expected 70 (100 - 30), got %d", score)
	}
}

func TestIntegrationCalculateSecurityIndex_MultipleCritical(t *testing.T) {
	severityCounts := map[string]int{
		"critical": 4,
	}
	score := calculateSecurityIndex(severityCounts)
	if score != 0 {
		t.Errorf("expected 0 (clamped), got %d", score)
	}
}

func TestIntegrationCalculateSecurityIndex_HighOnly(t *testing.T) {
	severityCounts := map[string]int{
		"high": 1,
	}
	score := calculateSecurityIndex(severityCounts)
	if score != 80 {
		t.Errorf("expected 80 (100 - 20), got %d", score)
	}
}

func TestIntegrationCalculateSecurityIndex_MediumOnly(t *testing.T) {
	severityCounts := map[string]int{
		"medium": 1,
	}
	score := calculateSecurityIndex(severityCounts)
	if score != 90 {
		t.Errorf("expected 90 (100 - 10), got %d", score)
	}
}

func TestIntegrationCalculateSecurityIndex_LowOnly(t *testing.T) {
	severityCounts := map[string]int{
		"low": 1,
	}
	score := calculateSecurityIndex(severityCounts)
	if score != 95 {
		t.Errorf("expected 95 (100 - 5), got %d", score)
	}
}

func TestIntegrationCalculateSecurityIndex_MixedSeverities(t *testing.T) {
	severityCounts := map[string]int{
		"critical": 1,
		"high":     2,
		"medium":   3,
		"low":      4,
	}
	score := calculateSecurityIndex(severityCounts)
	// 100 - 30 - 40 - 30 - 20 = -20, clamped to 0
	expected := 0
	if score != expected {
		t.Errorf("expected %d (clamped), got %d", expected, score)
	}
}

func TestIntegrationCalculateSecurityIndex_ClampUpper(t *testing.T) {
	// With empty map, score should be 100 (not > 100)
	score := calculateSecurityIndex(map[string]int{})
	if score > 100 {
		t.Errorf("expected max 100, got %d", score)
	}
}

func TestIntegrationCalculateSecurityIndex_Zero(t *testing.T) {
	severityCounts := map[string]int{
		"critical": 0,
		"high":     0,
	}
	score := calculateSecurityIndex(severityCounts)
	if score != 100 {
		t.Errorf("expected 100, got %d", score)
	}
}

// === Integration: SecurityService — Edge Cases ===

func TestIntegrationSecurityService_EmptyResultsSlice(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{}
	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error for empty slice, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Score < 0 || report.Score > 100 {
		t.Errorf("expected score 0-100, got %d", report.Score)
	}
	if len(report.PortAudit) != 0 {
		t.Errorf("expected 0 port audit findings, got %d", len(report.PortAudit))
	}
}

func TestIntegrationSecurityService_ResultWithNoPorts(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "no-ports-host",
			Ports:    []contracts.PortInfo{},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Score < 0 || report.Score > 100 {
		t.Errorf("expected score 0-100, got %d", report.Score)
	}
}

func TestIntegrationSecurityService_ResultWithClosedPorts(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "closed-ports-host",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "closed", Protocol: "tcp", Service: "ssh"},
				{Port: 80, State: "closed", Protocol: "tcp", Service: "http"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_MultipleHosts(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "host-1",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
			},
		},
		{
			IP:       "192.168.1.2",
			Hostname: "host-2",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
				{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
			},
		},
		{
			IP:       "192.168.1.3",
			Hostname: "host-3",
			Ports: []contracts.PortInfo{
				{Port: 21, State: "open", Protocol: "tcp", Service: "ftp"},
				{Port: 23, State: "open", Protocol: "tcp", Service: "telnet"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Score < 0 || report.Score > 100 {
		t.Errorf("expected score 0-100, got %d", report.Score)
	}
	// Multiple hosts with open ports should generate findings
	if len(report.PortAudit) == 0 {
		t.Error("expected port audit findings for multiple hosts")
	}
}

func TestIntegrationSecurityService_ResultWithVersionInfo(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "versioned-host",
			Ports: []contracts.PortInfo{
				{
					Port:    22,
					State:   "open",
					Protocol: "tcp",
					Service: "ssh",
					Version: "OpenSSH_8.9",
					Banner:  "SSH-2.0-OpenSSH_8.9",
				},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_ResultWithDeviceType(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:         "192.168.1.1",
			Hostname:   "switch-host",
			DeviceType: "switch",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_ResultWithMAC(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:   "192.168.1.1",
			MAC:  "aa:bb:cc:dd:ee:ff",
			Ports: []contracts.PortInfo{
				{Port: 445, State: "open", Protocol: "tcp", Service: "smb"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_ResultWithOSGuess(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:      "192.168.1.1",
			GuessOS: "Linux 5.15",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_ResultWithVendor(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:           "192.168.1.1",
			DeviceVendor: "Cisco",
			Ports: []contracts.PortInfo{
				{Port: 161, State: "open", Protocol: "udp", Service: "snmp"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_FindingStructure(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "test-host",
			Ports: []contracts.PortInfo{
				{Port: 21, State: "open", Protocol: "tcp", Service: "ftp"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	// Verify finding structure
	if len(report.PortAudit) > 0 {
		finding := report.PortAudit[0]
		if finding.Host == "" {
			t.Error("expected non-empty finding host")
		}
		if finding.Title == "" {
			t.Error("expected non-empty finding title")
		}
		// Severity can be empty if audit doesn't flag this port
		_ = finding.Severity
		_ = finding.Recommendation
	}
}

func TestIntegrationSecurityService_RiskSigStructure(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "test-host",
			Ports: []contracts.PortInfo{
				{Port: 23, State: "open", Protocol: "tcp", Service: "telnet"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	// RiskSig may or may not have findings depending on risk signature database
	// Just verify the structure is valid
	_ = report.RiskSig
}

func TestIntegrationSecurityService_ContextCancellation(t *testing.T) {
	svc := NewService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Context cancellation should not cause panic
	_, err := svc.AnalyzeRun(ctx, nil)
	// Since AnalyzeRun is synchronous, error may or may not be returned
	_ = err
}

func TestIntegrationSecurityService_LargeResultSet(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := make([]contracts.ScanResult, 50)
	for i := 0; i < 50; i++ {
		results[i] = contracts.ScanResult{
			IP:       "192.168.1.1",
			Hostname: "test-host",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
			},
		}
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error for large result set, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Score < 0 || report.Score > 100 {
		t.Errorf("expected score 0-100, got %d", report.Score)
	}
}

func TestIntegrationSecurityService_PortInfoFields(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "full-info-host",
			Ports: []contracts.PortInfo{
				{
					Port:     443,
					State:    "open",
					Protocol: "tcp",
					Service:  "https",
					Banner:   "HTTP/1.1 200 OK",
					Version:  "Apache/2.4.52",
				},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_InvalidPortState(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "invalid-state-host",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "unknown", Protocol: "tcp", Service: "http"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_UnicodeHostname(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "тест-хост-日本語",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_MultipleProtocols(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "multi-protocol-host",
			Ports: []contracts.PortInfo{
				{Port: 53, State: "open", Protocol: "udp", Service: "dns"},
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
				{Port: 161, State: "open", Protocol: "udp", Service: "snmp"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationSecurityService_ScoreBoundary_CriticalClamp(t *testing.T) {
	// Test that score clamps to 0 even with many critical findings
	score := calculateSecurityIndex(map[string]int{
		"critical": 100,
	})
	if score != 0 {
		t.Errorf("expected score clamped to 0, got %d", score)
	}
}

func TestIntegrationSecurityService_ScoreBoundary_MediumHigh(t *testing.T) {
	// 100 - 3*10 - 2*20 = 100 - 30 - 40 = 30
	score := calculateSecurityIndex(map[string]int{
		"medium": 3,
		"high":   2,
	})
	if score != 30 {
		t.Errorf("expected 30, got %d", score)
	}
}

func TestIntegrationSecurityService_ScoreBoundary_LowMany(t *testing.T) {
	// 100 - 20*5 = 0
	score := calculateSecurityIndex(map[string]int{
		"low": 20,
	})
	if score != 0 {
		t.Errorf("expected 0, got %d", score)
	}
}

// === Integration: NewService ===

func TestIntegrationNewService_ReturnsValid(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestIntegrationNewService_ImplementsInterface(t *testing.T) {
	var _ contracts.SecurityService = NewService()
}

// === Integration: AnalyzeRun Edge Cases ===

func TestIntegrationAnalyzeRun_NilContext(t *testing.T) {
	svc := NewService()
	report, err := svc.AnalyzeRun(nil, []contracts.ScanResult{})
	if err != nil {
		t.Fatalf("expected no error for nil context, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationAnalyzeRun_NilResults(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	report, err := svc.AnalyzeRun(ctx, nil)
	if err != nil {
		t.Fatalf("expected no error for nil results, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationAnalyzeRun_EmptyResults(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	report, err := svc.AnalyzeRun(ctx, []contracts.ScanResult{})
	if err != nil {
		t.Fatalf("expected no error for empty results, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Score != 100 {
		t.Errorf("expected score 100 for empty results, got %d", report.Score)
	}
	if report.PortAudit != nil && len(report.PortAudit) != 0 {
		t.Errorf("expected 0 port audit findings, got %d", len(report.PortAudit))
	}
}

func TestIntegrationAnalyzeRun_AllPortStates(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "all-states",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
				{Port: 80, State: "closed", Protocol: "tcp", Service: "http"},
				{Port: 443, State: "filtered", Protocol: "tcp", Service: "https"},
				{Port: 8080, State: "open", Protocol: "tcp", Service: "http-proxy"},
			},
		},
	}
	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationAnalyzeRun_AllProtocols(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "all-protocols",
			Ports: []contracts.PortInfo{
				{Port: 53, State: "open", Protocol: "udp", Service: "dns"},
				{Port: 67, State: "open", Protocol: "udp", Service: "dhcp"},
				{Port: 68, State: "open", Protocol: "udp", Service: "dhcp"},
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
				{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
				{Port: 123, State: "open", Protocol: "udp", Service: "ntp"},
			},
		},
	}
	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationAnalyzeRun_SeveralHighPorts(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "high-ports",
			Ports: []contracts.PortInfo{
				{Port: 8080, State: "open", Protocol: "tcp", Service: "http-alt"},
				{Port: 8443, State: "open", Protocol: "tcp", Service: "https-alt"},
				{Port: 9090, State: "open", Protocol: "tcp", Service: "prometheus"},
				{Port: 3000, State: "open", Protocol: "tcp", Service: "grafana"},
			},
		},
	}
	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegrationAnalyzeRun_DangerousPorts(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "dangerous-ports",
			Ports: []contracts.PortInfo{
				{Port: 21, State: "open", Protocol: "tcp", Service: "ftp"},
				{Port: 23, State: "open", Protocol: "tcp", Service: "telnet"},
				{Port: 135, State: "open", Protocol: "tcp", Service: "rpc"},
				{Port: 139, State: "open", Protocol: "tcp", Service: "netbios"},
				{Port: 445, State: "open", Protocol: "tcp", Service: "smb"},
				{Port: 3389, State: "open", Protocol: "tcp", Service: "rdp"},
			},
		},
	}
	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(report.PortAudit) == 0 {
		t.Error("expected port audit findings for dangerous ports")
	}
}

func TestIntegrationAnalyzeRun_FindingFields(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{
			IP:       "10.0.0.1",
			Hostname: "finding-test",
			Ports: []contracts.PortInfo{
				{Port: 21, State: "open", Protocol: "tcp", Service: "ftp"},
			},
		},
	}
	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	// Verify finding structure
	for _, f := range report.PortAudit {
		if f.Host == "" {
			t.Error("expected non-empty finding host")
		}
		if f.Title == "" {
			t.Error("expected non-empty finding title")
		}
	}
	for _, f := range report.RiskSig {
		if f.Host == "" {
			t.Error("expected non-empty risk sig host")
		}
		if f.Title == "" {
			t.Error("expected non-empty risk sig title")
		}
	}
}

func TestIntegrationAnalyzeRun_ConcurrentAccess(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "concurrent",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		},
	}
	// Run multiple analyzes concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := svc.AnalyzeRun(ctx, results)
			if err != nil {
				t.Errorf("concurrent analyze error: %v", err)
			}
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

// === Integration: Score Calculation Edge Cases ===

func TestIntegrationScore_NoSeverityKeys(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{})
	if score != 100 {
		t.Errorf("expected 100, got %d", score)
	}
}

func TestIntegrationScore_UnknownSeverity(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{
		"unknown": 5,
	})
	if score != 100 {
		t.Errorf("expected 100 for unknown severity, got %d", score)
	}
}

func TestIntegrationScore_MixedKnownUnknown(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{
		"high":    1,
		"unknown": 5,
	})
	if score != 80 {
		t.Errorf("expected 80 (100 - 20), got %d", score)
	}
}

func TestIntegrationScore_ExactClampAt0(t *testing.T) {
	// 100 - 4*30 = -20, should clamp to 0
	score := calculateSecurityIndex(map[string]int{
		"critical": 4,
	})
	if score != 0 {
		t.Errorf("expected 0, got %d", score)
	}
}

func TestIntegrationScore_ExactClampAt100(t *testing.T) {
	// 100 - 0 = 100, should be exactly 100
	score := calculateSecurityIndex(map[string]int{})
	if score != 100 {
		t.Errorf("expected 100, got %d", score)
	}
}

func TestIntegrationScore_ComplexCalculation(t *testing.T) {
	// 1 critical (30) + 2 high (40) + 3 medium (30) + 1 low (5) = 105
	// 100 - 105 = -5, clamped to 0
	score := calculateSecurityIndex(map[string]int{
		"critical": 1,
		"high":     2,
		"medium":   3,
		"low":      1,
	})
	if score != 0 {
		t.Errorf("expected 0, got %d", score)
	}
}

func TestIntegrationScore_MinimalDeduction(t *testing.T) {
	// 1 low = 5
	score := calculateSecurityIndex(map[string]int{
		"low": 1,
	})
	if score != 95 {
		t.Errorf("expected 95, got %d", score)
	}
}

func TestIntegrationScore_MediumExact(t *testing.T) {
	// 2 medium = 20
	score := calculateSecurityIndex(map[string]int{
		"medium": 2,
	})
	if score != 80 {
		t.Errorf("expected 80, got %d", score)
	}
}

func TestIntegrationScore_HighExact(t *testing.T) {
	// 5 high = 100
	score := calculateSecurityIndex(map[string]int{
		"high": 5,
	})
	if score != 0 {
		t.Errorf("expected 0, got %d", score)
	}
}

func TestIntegrationScore_CriticalExact(t *testing.T) {
	// 4 critical = 120
	score := calculateSecurityIndex(map[string]int{
		"critical": 4,
	})
	if score != 0 {
		t.Errorf("expected 0, got %d", score)
	}
}

func TestIntegrationScore_ZeroValues(t *testing.T) {
	// All zero values
	score := calculateSecurityIndex(map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	})
	if score != 100 {
		t.Errorf("expected 100, got %d", score)
	}
}
