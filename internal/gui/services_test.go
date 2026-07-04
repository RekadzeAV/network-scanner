package gui

import (
	"testing"

	"network-scanner/internal/builder"
)

// --- Test AppServices ---

func TestNewAppServices(t *testing.T) {
	container := builder.NewContainer(builder.Config{
		LogLevel: "info",
	})

	services := NewAppServices(container)

	if services == nil {
		t.Fatal("NewAppServices() returned nil")
	}

	if services.Scanner == nil {
		t.Error("Scanner service should not be nil")
	}

	if services.Topology == nil {
		t.Error("Topology service should not be nil")
	}

	if services.Security == nil {
		t.Error("Security service should not be nil")
	}

	if services.RemoteExec == nil {
		t.Error("RemoteExec service should not be nil")
	}

	if services.Inventory == nil {
		t.Error("Inventory service should not be nil")
	}

	if services.ScannerGUI == nil {
		t.Error("ScannerGUI service should not be nil")
	}

	if services.DeviceControlGUI == nil {
		t.Error("DeviceControlGUI service should not be nil")
	}

	if services.Audit == nil {
		t.Error("Audit service should not be nil")
	}

	if services.RiskSignature == nil {
		t.Error("RiskSignature service should not be nil")
	}

	if services.WOL == nil {
		t.Error("WOL service should not be nil")
	}

	if services.NetTools == nil {
		t.Error("NetTools service should not be nil")
	}
}

func TestAppServicesAllNotNil(t *testing.T) {
	container := builder.NewContainer(builder.Config{
		LogLevel: "info",
	})

	services := NewAppServices(container)

	// Проверяем, что все сервисы инициализированы
	servicesToCheck := []struct {
		name string
		ok   bool
	}{
		{"Scanner", services.Scanner != nil},
		{"Topology", services.Topology != nil},
		{"Security", services.Security != nil},
		{"RemoteExec", services.RemoteExec != nil},
		{"Inventory", services.Inventory != nil},
		{"ScannerGUI", services.ScannerGUI != nil},
		{"DeviceControlGUI", services.DeviceControlGUI != nil},
		{"Audit", services.Audit != nil},
		{"RiskSignature", services.RiskSignature != nil},
		{"WOL", services.WOL != nil},
		{"NetTools", services.NetTools != nil},
	}

	for _, s := range servicesToCheck {
		if !s.ok {
			t.Errorf("Service %s should not be nil", s.name)
		}
	}
}

// --- Test ScannerGUIService ---

func TestNewScannerGUIService(t *testing.T) {
	container := builder.NewContainer(builder.Config{
		LogLevel: "info",
	})

	service := NewScannerGUIService(container)

	if service == nil {
		t.Fatal("NewScannerGUIService() returned nil")
	}
}

// --- Test DeviceControlGUIService ---

func TestNewDeviceControlGUIService(t *testing.T) {
	container := builder.NewContainer(builder.Config{
		LogLevel: "info",
	})

	service := NewDeviceControlGUIService(container)

	if service == nil {
		t.Fatal("NewDeviceControlGUIService() returned nil")
	}
}

// --- Test AuditService ---

func TestNewAuditService(t *testing.T) {
	service := NewAuditService()

	if service == nil {
		t.Fatal("NewAuditService() returned nil")
	}
}

// --- Test RiskSignatureService ---

func TestNewRiskSignatureService(t *testing.T) {
	service := NewRiskSignatureService()

	if service == nil {
		t.Fatal("NewRiskSignatureService() returned nil")
	}
}

// --- Test WOLService ---

func TestNewWOLService(t *testing.T) {
	service := NewWOLService()

	if service == nil {
		t.Fatal("NewWOLService() returned nil")
	}
}

// --- Test NetToolsService ---

func TestNewNetToolsService(t *testing.T) {
	service := NewNetToolsService()

	if service == nil {
		t.Fatal("NewNetToolsService() returned nil")
	}
}
