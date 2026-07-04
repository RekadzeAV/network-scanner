package gui

import (
	"network-scanner/internal/builder"
	"network-scanner/internal/contracts"
)

// AppServices агрегация сервисов для GUI
type AppServices struct {
	Scanner          contracts.ScannerService
	Topology         contracts.TopologyService
	Security         contracts.SecurityService
	RemoteExec       contracts.RemoteExecService
	Inventory        contracts.InventoryService
	ScannerGUI       *ScannerGUIService
	DeviceControlGUI *DeviceControlGUIService
	Audit            *AuditService
	RiskSignature    *RiskSignatureService
	WOL              *WOLService
	NetTools         *NetToolsService
}

// NewAppServices создаёт AppServices из контейнера
func NewAppServices(container *builder.Container) *AppServices {
	return &AppServices{
		Scanner:          container.GetScanner(),
		Topology:         container.GetTopology(),
		Security:         container.GetSecurity(),
		RemoteExec:       container.GetRemoteExec(),
		Inventory:        container.GetInventory(),
		ScannerGUI:       NewScannerGUIService(container),
		DeviceControlGUI: NewDeviceControlGUIService(container),
		Audit:            NewAuditService(),
		RiskSignature:    NewRiskSignatureService(),
		WOL:              NewWOLService(),
		NetTools:         NewNetToolsService(),
	}
}
