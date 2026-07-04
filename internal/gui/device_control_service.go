package gui

import (
	"fmt"
	"time"

	"network-scanner/internal/builder"
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

// GetStatus проверяет статус устройства
func (s *DeviceControlGUIService) GetStatus(target, vendor, user, pass string, timeout time.Duration) (*DeviceStatusResult, error) {
	if target == "" {
		return nil, fmt.Errorf("target is required")
	}

	container := s.container
	inventorySvc := container.GetInventory()
	_ = inventorySvc

	// TODO: реальный вызов device-control
	return &DeviceStatusResult{
		Success:  false,
		Hostname: "mock",
		Status:   "stub",
	}, nil
}

// RebootDevice перезагружает устройство
func (s *DeviceControlGUIService) RebootDevice(target, vendor, user, pass string, timeout time.Duration) (*DeviceRebootResult, error) {
	if target == "" {
		return nil, fmt.Errorf("target is required")
	}

	container := s.container
	inventorySvc := container.GetInventory()
	_ = inventorySvc

	// TODO: реальный вызов device-control
	return &DeviceRebootResult{
		Success: false,
		Message: "stub",
	}, nil
}
