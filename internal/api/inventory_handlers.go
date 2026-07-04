package api

import (
	"encoding/json"
	"net/http"
	"time"

	"network-scanner/internal/contracts"

	"github.com/gorilla/mux"
)

// inventoryRequest Р·Р°РїСЂРѕСЃ РЅР° СЃРѕС…СЂР°РЅРµРЅРёРµ СЃРЅР°РїС€РѕС‚Р°
type inventoryRequest struct {
	ID       string                 `json:"id"`
	Results  []contracts.ScanResult `json:"results"`
	Metadata map[string]string      `json:"metadata,omitempty"`
}

// inventoryResponse РѕС‚РІРµС‚
type inventoryResponse struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	HostCount int       `json:"host_count"`
	Message   string    `json:"message"`
}

// inventoryDiffResponse РѕС‚РІРµС‚ РЅР° diff
type inventoryDiffResponse struct {
	ScanIDA string                 `json:"scan_id_a"`
	ScanIDB string                 `json:"scan_id_b"`
	New     []contracts.ScanResult `json:"new"`
	Missing []contracts.ScanResult `json:"missing"`
	Changed []contracts.Change  `json:"changed"`
}

// handleInventoryList РІРѕР·РІСЂР°С‰Р°РµС‚ СЃРїРёСЃРѕРє СЃРЅР°РїС€РѕС‚РѕРІ
func (h *Handler) handleInventoryList(w http.ResponseWriter, r *http.Request) {
	// TODO: integrate with real inventory service
	// For now, return mock data
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"snapshots": []contracts.Snapshot{},
		"message":   "inventory list (mock)",
	})
}

// handleInventorySave СЃРѕС…СЂР°РЅСЏРµС‚ СЃРЅР°РїС€РѕС‚
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

	// TODO: integrate with real inventory service
	// For now, simulate save
	response := inventoryResponse{
		ID:        req.ID,
		Timestamp: time.Now(),
		HostCount: len(req.Results),
		Message:   "snapshot saved successfully",
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// handleInventoryDiff СЃСЂР°РІРЅРёРІР°РµС‚ РґРІР° СЃРЅР°РїС€РѕС‚Р°
func (h *Handler) handleInventoryDiff(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idA := vars["id_a"]
	idB := vars["id_b"]

	if idA == "" || idB == "" {
		h.writeError(w, http.StatusBadRequest, "id_a and id_b are required")
		return
	}

	// TODO: integrate with real inventory service
	// For now, return mock diff
	response := inventoryDiffResponse{
		ScanIDA: idA,
		ScanIDB: idB,
		New:     []contracts.ScanResult{},
		Missing: []contracts.ScanResult{},
		Changed: []contracts.Change{},
	}

	h.writeJSON(w, http.StatusOK, response)
}

