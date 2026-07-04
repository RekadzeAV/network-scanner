package builder

import (
	"network-scanner/internal/contracts"
	"network-scanner/internal/scanner"
	"network-scanner/internal/security"
	"network-scanner/internal/services"
	"network-scanner/internal/topology"
)

// Container provides dependency injection for services.
type Container struct {
	scannerService    contracts.ScannerService
	topologyService   contracts.TopologyService
	securityService   contracts.SecurityService
	remoteExecService contracts.RemoteExecService
	inventoryService  contracts.InventoryService
}

// Config holds configuration for the container.
type Config struct {
	LogLevel string
	DBPath   string // Путь к inventory SQLite базе
	// Add more config fields as needed.
}

// NewContainer creates a new DI container with all services.
func NewContainer(cfg Config) *Container {
	return &Container{
		scannerService:    scanner.NewService(cfg.LogLevel),
		topologyService:   topology.NewService(),
		securityService:   security.NewService(),
		remoteExecService: services.NewRemoteExecService(),
		inventoryService:  services.NewInventoryService(cfg.DBPath),
	}
}

// GetScanner returns the ScannerService instance.
func (c *Container) GetScanner() contracts.ScannerService {
	return c.scannerService
}

// GetTopology returns the TopologyService instance.
func (c *Container) GetTopology() contracts.TopologyService {
	return c.topologyService
}

// GetSecurity returns the SecurityService instance.
func (c *Container) GetSecurity() contracts.SecurityService {
	return c.securityService
}

// GetRemoteExec returns the RemoteExecService instance.
func (c *Container) GetRemoteExec() contracts.RemoteExecService {
	return c.remoteExecService
}

// GetInventory returns the InventoryService instance.
func (c *Container) GetInventory() contracts.InventoryService {
	return c.inventoryService
}
