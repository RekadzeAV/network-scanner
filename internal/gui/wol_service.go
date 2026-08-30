package gui

import (
	"context"
	"fmt"
	"time"

	"network-scanner/internal/wol"
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

// SendWOL отправляет WoL-магический пакет с реальным вызовом wol
func (s *WOLService) SendWOL(ctx context.Context, mac, bcast, iface string, timeout time.Duration) (*WOLResult, error) {
	if mac == "" {
		return nil, fmt.Errorf("MAC address is required")
	}

	if bcast == "" {
		bcast = "255.255.255.255"
	}

	start := time.Now()

	_, err := wol.SendMagicPacketWithInterface(mac, bcast, iface)
	duration := time.Since(start)

	result := &WOLResult{
		Success:  err == nil,
		Message:  "Magic packet sent successfully",
		Duration: duration,
	}

	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("WOL failed: %w", err)
	}

	return result, nil
}
