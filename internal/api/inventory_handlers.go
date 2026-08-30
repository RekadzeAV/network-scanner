package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/inventory"
	"network-scanner/internal/scanner"

	"github.com/gorilla/mux"
)

// inventoryRequest запрос на сохранение снапшота
type inventoryRequest struct {
	ID       string                 `json:"id"`
	Results  []contracts.ScanResult `json:"results"`
	Metadata map[string]string      `json:"metadata,omitempty"`
}

// inventoryResponse ответ
type inventoryResponse struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	HostCount int       `json:"host_count"`
	Message   string    `json:"message"`
}

// inventoryDiffResponse ответ на diff
type inventoryDiffResponse struct {
	ScanIDA string                 `json:"scan_id_a"`
	ScanIDB string                 `json:"scan_id_b"`
	New     []contracts.ScanResult `json:"new"`
	Missing []contracts.ScanResult `json:"missing"`
	Changed []contracts.Change     `json:"changed"`
}

// handleInventoryList возвращает список снапшотов с реального инвентаря
func (h *Handler) handleInventoryList(w http.ResponseWriter, r *http.Request) {
	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "open inventory store")
		return
	}
	defer store.Close()

	snapshots, err := store.ListSnapshots(100)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "list snapshots")
		return
	}

	out := make([]contracts.Snapshot, 0, len(snapshots))
	for _, s := range snapshots {
		out = append(out, contracts.Snapshot{
			ID:        s.ID,
			Timestamp: s.Timestamp,
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"snapshots": out,
		"count":     len(out),
		"message":   "inventory list ok",
	})
}

// handleInventorySave сохраняет снапшот в реальный инвентарь
func (h *Handler) handleInventorySave(w http.ResponseWriter, r *http.Request) {
	var req inventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ID == "" {
		h.writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	if len(req.Results) == 0 {
		h.writeError(w, http.StatusBadRequest, "results are required")
		return
	}

	// Конвертируем contracts.ScanResult в scanner.Result
	scanResults := make([]scanner.Result, 0, len(req.Results))
	for _, cr := range req.Results {
		ports := make([]scanner.PortInfo, 0, len(cr.Ports))
		for _, p := range cr.Ports {
			ports = append(ports, scanner.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}
		scanResults = append(scanResults, scanner.Result{
			IP:           cr.IP,
			Hostname:     cr.Hostname,
			MAC:          cr.MAC,
			Ports:        ports,
			DeviceType:   cr.DeviceType,
			DeviceVendor: cr.DeviceVendor,
			GuessOS:      cr.GuessOS,
		})
	}

	// Сохраняем в реальный инвентарь
	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "open inventory store")
		return
	}
	defer store.Close()

	timestamp := time.Now()
	err = store.SaveSnapshot(req.ID, timestamp, scanResults)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "save snapshot")
		return
	}

	response := inventoryResponse{
		ID:        req.ID,
		Timestamp: timestamp,
		HostCount: len(scanResults),
		Message:   "snapshot saved successfully",
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// handleInventoryDiff сравнивает два снапшота с реального инвентаря
func (h *Handler) handleInventoryDiff(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idA := vars["id_a"]
	idB := vars["id_b"]

	if idA == "" || idB == "" {
		h.writeError(w, http.StatusBadRequest, "id_a and id_b are required")
		return
	}

	// Загружаем и сравниваем снапшоты из реального инвентаря
	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "open inventory store")
		return
	}
	defer store.Close()

	diff, err := store.Diff(idA, idB)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "diff snapshots")
		return
	}

	// Конвертируем scanner.Result в contracts.ScanResult
	newResults := make([]contracts.ScanResult, 0, len(diff.New))
	for _, host := range diff.New {
		ports := make([]contracts.PortInfo, 0, len(host.Ports))
		for _, p := range host.Ports {
			ports = append(ports, contracts.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}
		newResults = append(newResults, contracts.ScanResult{
			IP:         host.IP,
			Hostname:   host.Hostname,
			MAC:        host.MAC,
			Ports:      ports,
			DeviceType: host.DeviceType,
		})
	}

	missingResults := make([]contracts.ScanResult, 0, len(diff.Missing))
	for _, host := range diff.Missing {
		ports := make([]contracts.PortInfo, 0, len(host.Ports))
		for _, p := range host.Ports {
			ports = append(ports, contracts.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}
		missingResults = append(missingResults, contracts.ScanResult{
			IP:         host.IP,
			Hostname:   host.Hostname,
			MAC:        host.MAC,
			Ports:      ports,
			DeviceType: host.DeviceType,
		})
	}

	changedResults := make([]contracts.Change, 0, len(diff.Changed))
	for _, ch := range diff.Changed {
		changedResults = append(changedResults, contracts.Change{
			Key:          ch.Key,
			ChangedField: ch.ChangedField,
		})
	}

	response := inventoryDiffResponse{
		ScanIDA: idA,
		ScanIDB: idB,
		New:     newResults,
		Missing: missingResults,
		Changed: changedResults,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// contractPortToScanner преобразует contracts.PortInfo в scanner.PortInfo
func contractPortToScanner(p contracts.PortInfo) scanner.PortInfo {
	return scanner.PortInfo{
		Port:     p.Port,
		State:    strings.ToLower(strings.TrimSpace(p.State)),
		Protocol: strings.ToLower(strings.TrimSpace(p.Protocol)),
		Service:  strings.TrimSpace(p.Service),
		Banner:   p.Banner,
		Version:  p.Version,
	}
}
