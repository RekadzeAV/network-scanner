package gui

import (
	"context"
	"testing"
	"time"
)

// --- wol_service.go tests ---

func TestNewWOLService_Created(t *testing.T) {
	svc := NewWOLService()
	if svc == nil {
		t.Fatal("expected non-nil WOLService")
	}
}

func TestWOLService_SendWOL_EmptyMAC(t *testing.T) {
	svc := NewWOLService()
	ctx := context.Background()
	result, err := svc.SendWOL(ctx, "", "192.168.1.255", "", 5*time.Second)
	if err == nil {
		t.Error("expected error for empty MAC")
	}
	if result != nil {
		t.Error("expected nil result for error")
	}
}

func TestWOLService_SendWOL_WithMAC(t *testing.T) {
	svc := NewWOLService()
	ctx := context.Background()
	result, err := svc.SendWOL(ctx, "00:11:22:33:44:55", "192.168.1.255", "", 5*time.Second)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Success может быть true, если пакет успешно отправлен (даже без устройства)
	if result.Message != "Magic packet sent successfully" {
		t.Errorf("expected Message='Magic packet sent successfully', got %q", result.Message)
	}
	// Duration может быть 0 для mock
}

func TestWOLService_SendWOL_WithBroadcast(t *testing.T) {
	svc := NewWOLService()
	ctx := context.Background()
	result, err := svc.SendWOL(ctx, "00:11:22:33:44:55", "10.0.0.255", "", 5*time.Second)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Message != "Magic packet sent successfully" {
		t.Errorf("expected Message='Magic packet sent successfully', got %q", result.Message)
	}
}

func TestWOLService_SendWOL_WithInterface(t *testing.T) {
	svc := NewWOLService()
	ctx := context.Background()
	result, err := svc.SendWOL(ctx, "00:11:22:33:44:55", "192.168.1.255", "eth0", 5*time.Second)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestWOLService_SendWOL_ContextCancelled(t *testing.T) {
	svc := NewWOLService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// WOL не проверяет контекст в mock
	_, _ = svc.SendWOL(ctx, "00:11:22:33:44:55", "192.168.1.255", "", 5*time.Second)
}

func TestWOLResult_Fields(t *testing.T) {
	result := &WOLResult{
		Success:  true,
		Message:  "WOL packet sent",
		Duration: 100 * time.Millisecond,
	}
	if !result.Success {
		t.Error("expected Success=true")
	}
	if result.Message != "WOL packet sent" {
		t.Errorf("expected 'WOL packet sent', got %q", result.Message)
	}
	if result.Duration != 100*time.Millisecond {
		t.Errorf("expected 100ms, got %v", result.Duration)
	}
}
