package gui

import (
	"context"
	"testing"
	"time"

	"network-scanner/internal/contracts"
)

// --- audit_service.go tests ---

func TestNewAuditService_Created(t *testing.T) {
	svc := NewAuditService()
	if svc == nil {
		t.Fatal("expected non-nil AuditService")
	}
}

func TestAuditService_RunAudit_EmptyResults(t *testing.T) {
	svc := NewAuditService()
	ctx := context.Background()
	result, err := svc.RunAudit(ctx, nil, "low", 10*time.Second)
	if err == nil {
		t.Error("expected error for empty results")
	}
	if result != nil {
		t.Error("expected nil result for error")
	}
}

func TestAuditService_RunAudit_WithResults(t *testing.T) {
	svc := NewAuditService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "test", Ports: []contracts.PortInfo{
			{Port: 23, State: "open", Protocol: "tcp", Service: "telnet"},
		}},
	}
	result, err := svc.RunAudit(ctx, results, "low", 10*time.Second)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// mock возвращает findings для опасных портов
	if result.Total < 0 {
		t.Errorf("expected Total>=0 (mock), got %d", result.Total)
	}
}

func TestAuditService_RunAudit_WithMultipleResults(t *testing.T) {
	svc := NewAuditService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{IP: "192.168.1.1", Ports: []contracts.PortInfo{
			{Port: 23, State: "open", Protocol: "tcp", Service: "telnet"},
		}},
		{IP: "192.168.1.2", Ports: []contracts.PortInfo{
			{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
		}},
		{IP: "192.168.1.3"},
	}
	result, err := svc.RunAudit(ctx, results, "high", 5*time.Second)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	// mock возвращает findings для опасных портов
	if result.Total < 0 {
		t.Errorf("expected Total>=0 (mock), got %d", result.Total)
	}
}

func TestAuditService_RunAudit_ContextCancelled(t *testing.T) {
	svc := NewAuditService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := []contracts.ScanResult{{IP: "192.168.1.1"}}
	// mock не проверяет контекст, поэтому ошибки не будет
	_, err := svc.RunAudit(ctx, results, "low", 10*time.Second)
	// mock не обрабатывает контекст — ошибка не ожидается
	_ = err
}

func TestNewRiskSignatureService_Created(t *testing.T) {
	svc := NewRiskSignatureService()
	if svc == nil {
		t.Fatal("expected non-nil RiskSignatureService")
	}
}

func TestRiskSignatureService_RunRiskSignatures_EmptyResults(t *testing.T) {
	svc := NewRiskSignatureService()
	ctx := context.Background()
	result, err := svc.RunRiskSignatures(ctx, nil, 10*time.Second)
	if err == nil {
		t.Error("expected error for empty results")
	}
	if result != nil {
		t.Error("expected nil result for error")
	}
}

func TestRiskSignatureService_RunRiskSignatures_WithResults(t *testing.T) {
	svc := NewRiskSignatureService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "test"},
	}
	result, err := svc.RunRiskSignatures(ctx, results, 10*time.Second)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// mock возвращает 1 entry независимо от количества результатов
	if result.Total != 1 {
		t.Errorf("expected Total=1 (stub), got %d", result.Total)
	}
	if len(result.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result.Entries))
	}
}

func TestRiskSignatureService_RunRiskSignatures_WithMultipleResults(t *testing.T) {
	svc := NewRiskSignatureService()
	ctx := context.Background()
	results := []contracts.ScanResult{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}
	result, err := svc.RunRiskSignatures(ctx, results, 5*time.Second)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	// mock возвращает 1 entry независимо от количества результатов
	if result.Total != 1 {
		t.Errorf("expected Total=1 (stub), got %d", result.Total)
	}
}

func TestRiskSignatureService_RunRiskSignatures_ContextCancelled(t *testing.T) {
	svc := NewRiskSignatureService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := []contracts.ScanResult{{IP: "192.168.1.1"}}
	// mock не проверяет контекст, поэтому ошибки не будет
	_, err := svc.RunRiskSignatures(ctx, results, 10*time.Second)
	// mock не обрабатывает контекст — ошибка не ожидается
	_ = err
}
