package services

import (
	"context"
	"fmt"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/inventory"
	"network-scanner/internal/scanner"
)

// InventoryService реализация InventoryService
type InventoryService struct {
	dbPath string
}

func (s *InventoryService) SaveSnapshot(ctx context.Context, id string, data []contracts.ScanResult) error {
	// Конвертация в internal формат
	internalResults := make([]scanner.Result, 0, len(data))
	for _, r := range data {
		ports := make([]scanner.PortInfo, 0, len(r.Ports))
		for _, p := range r.Ports {
			ports = append(ports, scanner.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}

		internalResults = append(internalResults, scanner.Result{
			IP:           r.IP,
			Hostname:     r.Hostname,
			MAC:          r.MAC,
			Ports:        ports,
			DeviceType:   r.DeviceType,
			DeviceVendor: r.DeviceVendor,
			GuessOS:      r.GuessOS,
		})
	}

	store, err := inventory.Open(s.dbPath)
	if err != nil {
		return fmt.Errorf("открытие inventory DB: %w", err)
	}
	defer store.Close()

	if err := store.SaveSnapshot(id, time.Now().UTC(), internalResults); err != nil {
		return fmt.Errorf("сохранение снапшота: %w", err)
	}

	return nil
}

func (s *InventoryService) ListSnapshots(ctx context.Context, limit int) ([]contracts.Snapshot, error) {
	store, err := inventory.Open(s.dbPath)
	if err != nil {
		return nil, fmt.Errorf("открытие inventory DB: %w", err)
	}
	defer store.Close()

	snapshots, err := store.ListSnapshots(limit)
	if err != nil {
		return nil, fmt.Errorf("список снапшотов: %w", err)
	}

	result := make([]contracts.Snapshot, 0, len(snapshots))
	for _, snap := range snapshots {
		result = append(result, contracts.Snapshot{
			ID:        snap.ID,
			Timestamp: snap.Timestamp,
		})
	}

	return result, nil
}

func (s *InventoryService) Diff(ctx context.Context, idA, idB string) (*contracts.Diff, error) {
	store, err := inventory.Open(s.dbPath)
	if err != nil {
		return nil, fmt.Errorf("открытие inventory DB: %w", err)
	}
	defer store.Close()

	diff, err := store.Diff(idA, idB)
	if err != nil {
		return nil, fmt.Errorf("вычисление diff: %w", err)
	}

	newResults := make([]contracts.ScanResult, 0, len(diff.New))
	for _, r := range diff.New {
		newResults = append(newResults, contracts.ScanResult{
			IP:       r.IP,
			Hostname: r.Hostname,
		})
	}

	missingResults := make([]contracts.ScanResult, 0, len(diff.Missing))
	for _, r := range diff.Missing {
		missingResults = append(missingResults, contracts.ScanResult{
			IP:       r.IP,
			Hostname: r.Hostname,
		})
	}

	changedList := make([]contracts.Change, 0, len(diff.Changed))
	for _, c := range diff.Changed {
		changedList = append(changedList, contracts.Change{
			Key:          c.Key,
			ChangedField: c.ChangedField,
		})
	}

	return &contracts.Diff{
		ScanIDA: diff.ScanIDA,
		ScanIDB: diff.ScanIDB,
		New:     newResults,
		Missing: missingResults,
		Changed: changedList,
	}, nil
}
