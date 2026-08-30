package gui

import (
	"context"
	"testing"

	"network-scanner/internal/contracts"
)

// --- scanner_service.go: Stop (пустая реализация) ---

// TestScannerGUIService_Stop_Duplicate проверяет что Stop не паникует.
// (основной тест в app_model_scanner_topology_test.go: TestScannerGUIService_Stop_NilService)
func TestScannerGUIService_Stop_Duplicate(t *testing.T) {
	s := &ScannerGUIService{svc: nil}
	// Пустая реализация — не должен паниковать
	s.Stop()
}

// --- scanner_service.go: Scan, ScanWithProgress (delegate to service) ---

// TestScannerGUIService_Scan_NilService_Panics проверяет что Scan паникует при nil service.
// (Scan делегирует вызов svc.Scan, который паникует при nil interface)
func TestScannerGUIService_Scan_NilService_Panics(t *testing.T) {
	s := &ScannerGUIService{svc: nil}
	// Scan паникует при nil service — это ожидаемое поведение
	defer func() {
		if r := recover(); r == nil {
			t.Error("ожидалась паника при вызове Scan с nil service")
		}
	}()
	ctx := context.Background()
	cfg := contracts.ScanConfig{
		NetworkCIDR: "192.168.1.0/24",
		PortRange:   "1-1000",
		Timeout:     2,
	}
	_, _ = s.Scan(ctx, cfg)
}

func TestScannerGUIService_ScanWithProgress_NilService_Panics(t *testing.T) {
	s := &ScannerGUIService{svc: nil}
	// ScanWithProgress паникует при nil service — это ожидаемое поведение
	defer func() {
		if r := recover(); r == nil {
			t.Error("ожидалась паника при вызове ScanWithProgress с nil service")
		}
	}()
	ctx := context.Background()
	cfg := contracts.ScanConfig{
		NetworkCIDR: "192.168.1.0/24",
		PortRange:   "1-1000",
		Timeout:     2,
	}
	_, _ = s.ScanWithProgress(ctx, cfg, nil)
}
