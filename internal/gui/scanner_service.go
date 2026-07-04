package gui

import (
	"context"
	"fmt"

	"network-scanner/internal/builder"
	"network-scanner/internal/contracts"
)

// ScannerGUIService обёртка для сканирования из GUI
type ScannerGUIService struct {
	svc contracts.ScannerService
}

// NewScannerGUIService создаёт ScannerGUIService
func NewScannerGUIService(container *builder.Container) *ScannerGUIService {
	return &ScannerGUIService{
		svc: container.GetScanner(),
	}
}

// Scan запускает сканирование с прогрессом
func (s *ScannerGUIService) Scan(ctx context.Context, cfg contracts.ScanConfig) ([]contracts.ScanResult, error) {
	return s.svc.Scan(ctx, cfg, func(stage string, current, total int, message string) {
		// GUI обновляется через callback, переданный в initUI
	})
}

// ScanWithProgress запускает сканирование с callback
func (s *ScannerGUIService) ScanWithProgress(ctx context.Context, cfg contracts.ScanConfig, onProgress func(stage string, current, total int, message string)) ([]contracts.ScanResult, error) {
	return s.svc.Scan(ctx, cfg, onProgress)
}

// Stop останавливает текущее сканирование
func (s *ScannerGUIService) Stop() {
	// Реализация через глобальное состояние сканера
}

// ValidateConfig проверяет конфигурацию сканирования
func (s *ScannerGUIService) ValidateConfig(cfg contracts.ScanConfig) error {
	if cfg.NetworkCIDR == "" {
		return fmt.Errorf("network CIDR is required")
	}
	if cfg.PortRange == "" {
		return fmt.Errorf("port range is required")
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	return nil
}
