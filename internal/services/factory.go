package services

// NewRemoteExecService создаёт RemoteExecService
func NewRemoteExecService() *RemoteExecService {
	return &RemoteExecService{}
}

// NewInventoryService создаёт InventoryService
func NewInventoryService(dbPath string) *InventoryService {
	if dbPath == "" {
		dbPath = "inventory/network_inventory.db"
	}
	return &InventoryService{dbPath: dbPath}
}
