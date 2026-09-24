package builder

import (
	"network-scanner/internal/contracts"
	"network-scanner/internal/eventbus"
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
	eventBus          *eventbus.EventBus
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

// WithEventBus wires an event bus into the container (E6).
//
// Включает публикацию событий scan.started / scan.completed / scan.failed
// из ScannerService. Возвращает контейнер для fluent-настройки:
//
//	c := builder.NewContainer(cfg).WithEventBus(eventbus.NewEventBus())
//
// Шина доступна потребителям через GetEventBus().
func (c *Container) WithEventBus(bus *eventbus.EventBus) *Container {
	c.eventBus = bus
	if s, ok := c.scannerService.(interface {
		WithEventBus(*eventbus.EventBus)
	}); ok {
		s.WithEventBus(bus)
	}
	return c
}

// GetEventBus returns the shared event bus (may be nil if not wired).
func (c *Container) GetEventBus() *eventbus.EventBus {
	return c.eventBus
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
