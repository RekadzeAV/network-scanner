package gui

import (
	"context"
	"fmt"
	"time"

	"network-scanner/internal/builder"
	"network-scanner/internal/devicecontrol"
)

// DeviceControlGUIService обёртка для управления устройствами
type DeviceControlGUIService struct {
	container *builder.Container
}

// NewDeviceControlGUIService создаёт DeviceControlGUIService
func NewDeviceControlGUIService(container *builder.Container) *DeviceControlGUIService {
	return &DeviceControlGUIService{
		container: container,
	}
}

// DeviceStatusResult результат проверки состояния устройства
type DeviceStatusResult struct {
	Success  bool
	Hostname string
	IP       string
	Status   string
	Error    string
	Duration time.Duration
}

// DeviceRebootResult результат перезагрузки устройства
type DeviceRebootResult struct {
	Success  bool
	Message  string
	Error    string
	Duration time.Duration
}

// GetStatus проверяет статус устройства с реальным вызовом devicecontrol
func (s *DeviceControlGUIService) GetStatus(target, vendor, user, pass string, timeout time.Duration) (*DeviceStatusResult, error) {
	if target == "" {
		return nil, fmt.Errorf("target is required")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	start := time.Now()

	// Убедимся, что target имеет http:// или https://
	if targetURL := target; !isHTTPURL(targetURL) {
		targetURL = "http://" + targetURL
		target = targetURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := devicecontrol.Request{
		Action:    devicecontrol.ActionStatus,
		TargetURL: target,
		Vendor:    vendor,
		Username:  user,
		Password:  pass,
		Timeout:   timeout,
	}

	resp, err := devicecontrol.Execute(ctx, req)
	duration := time.Since(start)

	result := &DeviceStatusResult{
		Success:  resp.Success,
		Hostname: target,
		IP:       target,
		Status:   resp.Message,
		Duration: duration,
	}

	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("device status check failed: %w", err)
	}

	return result, nil
}

// RebootDevice перезагружает устройство с реальным вызовом devicecontrol
func (s *DeviceControlGUIService) RebootDevice(target, vendor, user, pass string, timeout time.Duration) (*DeviceRebootResult, error) {
	if target == "" {
		return nil, fmt.Errorf("target is required")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	start := time.Now()

	// Убедимся, что target имеет http:// или https://
	if targetURL := target; !isHTTPURL(targetURL) {
		targetURL = "http://" + targetURL
		target = targetURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := devicecontrol.Request{
		Action:    devicecontrol.ActionReboot,
		TargetURL: target,
		Vendor:    vendor,
		Username:  user,
		Password:  pass,
		Timeout:   timeout,
	}

	resp, err := devicecontrol.Execute(ctx, req)
	duration := time.Since(start)

	result := &DeviceRebootResult{
		Success:  resp.Success,
		Message:  resp.Message,
		Duration: duration,
	}

	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("device reboot failed: %w", err)
	}

	return result, nil
}

// isHTTPURL проверяет, является ли URL HTTP/HTTPS
func isHTTPURL(url string) bool {
	return len(url) >= 7 && (url[:7] == "http://" || url[:8] == "https://")
}
