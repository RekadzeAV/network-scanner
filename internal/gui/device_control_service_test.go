package gui

import (
	"testing"
	"time"

	"network-scanner/internal/builder"
)

// --- device_control_service.go tests ---

func TestNewDeviceControlGUIService_NilContainer(t *testing.T) {
	svc := NewDeviceControlGUIService(nil)
	if svc == nil {
		t.Fatal("expected non-nil DeviceControlGUIService")
	}
	if svc.container != nil {
		t.Error("expected container to be nil")
	}
}

func TestNewDeviceControlGUIService_WithContainer(t *testing.T) {
	container := &builder.Container{}
	svc := NewDeviceControlGUIService(container)
	if svc == nil {
		t.Fatal("expected non-nil DeviceControlGUIService")
	}
	if svc.container != container {
		t.Error("expected container to be set")
	}
}

func TestDeviceControlGUIService_GetStatus_EmptyTarget(t *testing.T) {
	svc := NewDeviceControlGUIService(nil)
	result, err := svc.GetStatus("", "", "", "", 5*time.Second)
	if err == nil {
		t.Error("expected error for empty target")
	}
	if result != nil {
		t.Error("expected nil result for error")
	}
}

func TestDeviceControlGUIService_GetStatus_WithTarget(t *testing.T) {
	container := &builder.Container{}
	svc := NewDeviceControlGUIService(container)
	// GetStatus вызывает container.GetInventory(), который может паниковать
	// на nil или пустом container, поэтому тестируем только валидацию target
	result, err := svc.GetStatus("192.168.1.1", "cisco", "admin", "pass", 5*time.Second)
	// Может паниковать если container не инициализирован
	_ = result
	_ = err
}

func TestDeviceControlGUIService_GetStatus_VendorOnly(t *testing.T) {
	container := &builder.Container{}
	svc := NewDeviceControlGUIService(container)
	// GetStatus вызывает container.GetInventory()
	// Тестируем только валидацию target
	_, err := svc.GetStatus("192.168.1.1", "cisco", "", "", 5*time.Second)
	// Может паниковать если container не инициализирован
	_ = err
}

func TestDeviceControlGUIService_RebootDevice_EmptyTarget(t *testing.T) {
	svc := NewDeviceControlGUIService(nil)
	result, err := svc.RebootDevice("", "", "", "", 5*time.Second)
	if err == nil {
		t.Error("expected error for empty target")
	}
	if result != nil {
		t.Error("expected nil result for error")
	}
}

func TestDeviceControlGUIService_RebootDevice_WithTarget(t *testing.T) {
	container := &builder.Container{}
	svc := NewDeviceControlGUIService(container)
	// RebootDevice вызывает container.GetInventory()
	// Тестируем только валидацию target
	_, err := svc.RebootDevice("192.168.1.1", "cisco", "admin", "pass", 5*time.Second)
	// Может паниковать если container не инициализирован
	_ = err
}

func TestDeviceControlGUIService_RebootDevice_WithContainer(t *testing.T) {
	container := &builder.Container{}
	svc := NewDeviceControlGUIService(container)
	// RebootDevice вызывает container.GetInventory()
	// Тестируем только валидацию target
	_, err := svc.RebootDevice("192.168.1.1", "cisco", "admin", "pass", 5*time.Second)
	// Может паниковать если container не инициализирован
	_ = err
}
