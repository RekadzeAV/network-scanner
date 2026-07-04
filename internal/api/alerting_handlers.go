package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"network-scanner/internal/alerting"
	"network-scanner/internal/inventory"
)

// alertingEngine глобальный движок алертинга (для простоты)
var (
	alertingEng    *alerting.Engine
	alertingEngMu  sync.Mutex
)

// initAlerting инициализирует движок алертинга
func initAlerting(logFile string) {
	alertingEngMu.Lock()
	defer alertingEngMu.Unlock()
	alertingEng = alerting.NewEngine(logFile)
}

// alertsHandler обрабатывает GET /api/v1/alerts
func (h *Handler) alertsHandler(w http.ResponseWriter, r *http.Request) {
	if alertingEng == nil {
		h.writeError(w, http.StatusServiceUnavailable, "alerting not initialized")
		return
	}

	severity := r.URL.Query().Get("severity")
	
	var alerts []alerting.Alert
	if severity != "" {
		alerts = alertingEng.GetAlertsBySeverity(alerting.Severity(severity))
	} else {
		alerts = alertingEng.GetAlerts()
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts":   alerts,
		"count":    len(alerts),
		"severity": severity,
	})
}

// checkAlertsHandler обрабатывает POST /api/v1/alerts/check
func (h *Handler) checkAlertsHandler(w http.ResponseWriter, r *http.Request) {
	if alertingEng == nil {
		h.writeError(w, http.StatusServiceUnavailable, "alerting not initialized")
		return
	}

	var req struct {
		OldHosts []map[string]interface{} `json:"old_hosts"`
		NewHosts []map[string]interface{} `json:"new_hosts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// TODO: преобразовать map в scanner.Result
	// Для демонстрации возвращаем заглушку
	alerts := alertingEng.CheckAlerts(nil, nil)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// clearAlertsHandler обрабатывает DELETE /api/v1/alerts
func (h *Handler) clearAlertsHandler(w http.ResponseWriter, r *http.Request) {
	if alertingEng == nil {
		h.writeError(w, http.StatusServiceUnavailable, "alerting not initialized")
		return
	}

	alertingEng.ClearAlerts()

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "alerts cleared",
	})
}

// triggerAlertHandler обрабатывает POST /api/v1/alerts/trigger/{scan_id_a}/{scan_id_b}
func (h *Handler) triggerAlertHandler(w http.ResponseWriter, r *http.Request) {
	if alertingEng == nil {
		h.writeError(w, http.StatusServiceUnavailable, "alerting not initialized")
		return
	}

	vars := mux.Vars(r)
	scanIDA := vars["id_a"]
	scanIDB := vars["id_b"]

	if scanIDA == "" || scanIDB == "" {
		h.writeError(w, http.StatusBadRequest, "both scan IDs required")
		return
	}

	// Загружаем снапшоты
	store, err := inventory.Open(h.config.InventoryPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("open inventory: %v", err))
		return
	}
	defer store.Close()

	snapA, err := store.LoadSnapshot(scanIDA)
	if err != nil {
		h.writeError(w, http.StatusNotFound, fmt.Sprintf("snapshot A not found: %v", err))
		return
	}

	snapB, err := store.LoadSnapshot(scanIDB)
	if err != nil {
		h.writeError(w, http.StatusNotFound, fmt.Sprintf("snapshot B not found: %v", err))
		return
	}

	alerts := alertingEng.CheckAlerts(snapA.Hosts, snapB.Hosts)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
		"scan_a": scanIDA,
		"scan_b": scanIDB,
	})
}



