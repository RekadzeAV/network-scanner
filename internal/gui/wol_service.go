package gui

import (
	"context"
	"fmt"
	"time"
)

// WOLResult результат Wake-on-LAN
type WOLResult struct {
	Success  bool
	Message  string
	Error    string
	Duration time.Duration
}

// WOLService обёртка для Wake-on-LAN
type WOLService struct {
}

// NewWOLService создаёт WOLService
func NewWOLService() *WOLService {
	return &WOLService{}
}

// SendWOL отправляет WoL-магический пакет
func (s *WOLService) SendWOL(ctx context.Context, mac, bcast, iface string, timeout time.Duration) (*WOLResult, error) {
	if mac == "" {
		return nil, fmt.Errorf("MAC address is required")
	}

	start := time.Now()
	// TODO: реальный вызов wol
	return &WOLResult{
		Success:  false,
		Message:  "stub",
		Duration: time.Since(start),
	}, nil
}
