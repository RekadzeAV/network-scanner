package security

import (
	"context"
	"fmt"
	"testing"

	"network-scanner/internal/contracts"
)

// ============================================================================
// AnalyzeRun — полное покрытие
// ============================================================================

func TestAnalyzeRun_EmptyResults(t *testing.T) {
	svc := NewService()
	report, err := svc.AnalyzeRun(context.Background(), []contracts.ScanResult{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Score != 100 {
		t.Errorf("expected score 100 for empty, got %d", report.Score)
	}
	if len(report.PortAudit) != 0 {
		t.Errorf("expected 0 port audit findings, got %d", len(report.PortAudit))
	}
	if len(report.RiskSig) != 0 {
		t.Errorf("expected 0 risk sig findings, got %d", len(report.RiskSig))
	}
}

func TestAnalyzeRun_SingleHost(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "router",
			MAC:      "aa:bb:cc:dd:ee:ff",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
				{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
			},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Score < 0 || report.Score > 100 {
		t.Errorf("expected score between 0-100, got %d", report.Score)
	}
}

func TestAnalyzeRun_MultipleHosts(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP:    "192.168.1.1",
			Ports: []contracts.PortInfo{{Port: 80, State: "open"}},
		},
		{
			IP:    "192.168.1.2",
			Ports: []contracts.PortInfo{{Port: 22, State: "open"}},
		},
		{
			IP:    "192.168.1.3",
			Ports: []contracts.PortInfo{{Port: 443, State: "open"}},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_WithDeviceType(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP:           "192.168.1.1",
			DeviceType:   "Router",
			DeviceVendor: "Cisco",
			GuessOS:      "Cisco IOS",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "open", Service: "ssh"},
			},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_WithVersion(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP: "192.168.1.1",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open", Service: "http", Version: "nginx/1.25.0", Banner: "Server: nginx"},
			},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_ClosedPorts(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP: "192.168.1.1",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "closed", Protocol: "tcp"},
				{Port: 443, State: "filtered", Protocol: "tcp"},
			},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_UDP(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP: "192.168.1.1",
			Ports: []contracts.PortInfo{
				{Port: 53, State: "open", Protocol: "udp"},
			},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_ContextCancelled(t *testing.T) {
	svc := NewService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel

	// Даже с отменённым контекстом функция должна завершиться без паники
	report, err := svc.AnalyzeRun(ctx, []contracts.ScanResult{
		{IP: "192.168.1.1"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_WithHostname(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "my-router.local",
			MAC:      "aa:bb:cc:dd:ee:ff",
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_WithOS(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP:      "192.168.1.1",
			GuessOS: "Ubuntu 22.04",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "open", Service: "ssh"},
			},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_WithVendor(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP:           "192.168.1.1",
			DeviceVendor: "TP-Link",
			DeviceType:   "Switch",
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_PortsWithoutService(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP: "192.168.1.1",
			Ports: []contracts.PortInfo{
				{Port: 8080, State: "open", Protocol: "tcp"},
			},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_MixedPortStates(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP: "192.168.1.1",
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open"},
				{Port: 443, State: "closed"},
				{Port: 22, State: "filtered"},
			},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_WithBanner(t *testing.T) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP: "192.168.1.1",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "open", Service: "ssh", Banner: "SSH-2.0-OpenSSH_9.3"},
			},
		},
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyzeRun_ScoreRange(t *testing.T) {
	svc := NewService()
	// Тест с множеством хостов и портов — должен вернуть score в диапазоне 0-100
	results := make([]contracts.ScanResult, 50)
	for i := range results {
		results[i] = contracts.ScanResult{
			IP: fmt.Sprintf("192.168.1.%d", i+1),
			Ports: []contracts.PortInfo{
				{Port: 80, State: "open"},
				{Port: 443, State: "open"},
				{Port: 22, State: "open"},
			},
		}
	}
	report, err := svc.AnalyzeRun(context.Background(), results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report.Score < 0 || report.Score > 100 {
		t.Errorf("expected score 0-100, got %d", report.Score)
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkAnalyzeRun(b *testing.B) {
	svc := NewService()
	results := []contracts.ScanResult{
		{
			IP:    "192.168.1.1",
			Ports: []contracts.PortInfo{{Port: 80, State: "open"}},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.AnalyzeRun(context.Background(), results)
	}
}
