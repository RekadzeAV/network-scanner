package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"network-scanner/internal/inventory"
	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
)

// snmpCollectHandler обрабатывает POST /api/v1/snmp/collect
func (h *Handler) snmpCollectHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceIDs []string `json:"device_ids"`
		Community string   `json:"community"`
		Timeout   int      `json:"timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Community == "" {
		req.Community = "public"
	}
	if req.Timeout <= 0 {
		req.Timeout = 2
	}

	// Загружаем устройства из inventory
	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("open inventory: %v", err))
		return
	}
	defer store.Close()

	// Получаем все устройства или по ID
	var devices []scanner.Result
	if len(req.DeviceIDs) > 0 {
		for _, id := range req.DeviceIDs {
			snap, err := store.LoadSnapshot(id)
			if err != nil {
				h.writeError(w, http.StatusNotFound, fmt.Sprintf("snapshot %s not found: %v", id, err))
				return
			}
			devices = append(devices, snap.Hosts...)
		}
	} else {
		// Получаем последнее сканирование
		snapshots, err := store.ListSnapshots(1)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("list snapshots: %v", err))
			return
		}
		if len(snapshots) == 0 {
			h.writeJSON(w, http.StatusOK, map[string]interface{}{
				"message": "no snapshots found",
			})
			return
		}
		lastSnap, err := store.LoadSnapshot(snapshots[len(snapshots)-1].ID)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("load last snapshot: %v", err))
			return
		}
		devices = lastSnap.Hosts
	}

	// Запускаем SNMP опрос
	fmt.Printf("SNMP опрос: %d устройств\n", len(devices))
	snmpDevices, report, err := snmpcollector.CollectWithReport(devices, []string{req.Community}, req.Timeout)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("snmp collect error: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_targets":    report.TotalSNMPTargets,
		"connected":        report.Connected,
		"partial":          report.Partial,
		"failed":           report.Failed,
		"snmp_devices":     len(snmpDevices),
		"device_summaries": report.DeviceSummaries,
		"failures":         report.Failures,
	})
}
