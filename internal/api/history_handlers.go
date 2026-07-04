package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"network-scanner/internal/inventory"
)

// historyHandler обрабатывает GET /api/v1/history
func (h *Handler) historyHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to open inventory")
		return
	}
	defer store.Close()

	history, _, err := store.GetScanHistory(limit)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}

// compareHandler обрабатывает GET /api/v1/history/compare/{id_a}/{id_b}
func (h *Handler) compareHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanIDA := vars["id_a"]
	scanIDB := vars["id_b"]

	if scanIDA == "" || scanIDB == "" {
		h.writeError(w, http.StatusBadRequest, "both scan IDs required")
		return
	}

	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to open inventory")
		return
	}
	defer store.Close()

	result, err := store.CompareSnapshotsByName(scanIDA, scanIDB)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to compare: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}


