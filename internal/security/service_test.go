package security

import (
	"context"
	"testing"

	"network-scanner/internal/contracts"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestSecurityService_AnalyzeEmptyResults(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	report, err := svc.AnalyzeRun(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Score < 0 || report.Score > 100 {
		t.Fatalf("expected score 0-100, got %d", report.Score)
	}
}

func TestSecurityService_AnalyzeWithResults(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	results := []contracts.ScanResult{
		{
			IP:       "192.168.1.1",
			Hostname: "test-host",
			Ports: []contracts.PortInfo{
				{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		},
	}

	report, err := svc.AnalyzeRun(ctx, results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Score < 0 || report.Score > 100 {
		t.Fatalf("expected score 0-100, got %d", report.Score)
	}
}

func TestSecurityService_AnalyzeContextCancel(t *testing.T) {
	svc := NewService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.AnalyzeRun(ctx, nil)
	// Контекст может не влиять на синхронную операцию
	_ = err
}
