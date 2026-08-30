package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"network-scanner/internal/alerting"
	"network-scanner/internal/inventory"
	"network-scanner/internal/scanner"

	"github.com/gorilla/mux"
)

// alertingEngine глобальный движок алертинга (для простоты)
var (
	alertingEng   *alerting.Engine
	alertingEngMu sync.Mutex
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

// mapToScannerResult преобразует map[string]interface{} в scanner.Result
func mapToScannerResult(m map[string]interface{}) scanner.Result {
	result := scanner.Result{}

	if v, ok := m["ip"].(string); ok {
		result.IP = strings.TrimSpace(v)
	}
	if v, ok := m["hostname"].(string); ok {
		result.Hostname = strings.TrimSpace(v)
	}
	if v, ok := m["mac"].(string); ok {
		result.MAC = strings.TrimSpace(v)
	}
	if v, ok := m["device_type"].(string); ok {
		result.DeviceType = strings.TrimSpace(v)
	}
	if v, ok := m["device_vendor"].(string); ok {
		result.DeviceVendor = strings.TrimSpace(v)
	}
	if v, ok := m["guess_os"].(string); ok {
		result.GuessOS = strings.TrimSpace(v)
	}
	if v, ok := m["guess_os_confidence"].(string); ok {
		result.GuessOSConfidence = strings.TrimSpace(v)
	}
	if v, ok := m["guess_os_reason"].(string); ok {
		result.GuessOSReason = strings.TrimSpace(v)
	}
	if v, ok := m["snmp_enabled"].(bool); ok {
		result.SNMPEnabled = v
	}
	if v, ok := m["is_alive"].(bool); ok {
		result.IsAlive = v
	}

	// Конвертируем порты
	if portsRaw, ok := m["ports"].([]interface{}); ok {
		for _, pRaw := range portsRaw {
			if pMap, ok := pRaw.(map[string]interface{}); ok {
				port := scanner.PortInfo{}
				if v, ok := pMap["port"].(float64); ok {
					port.Port = int(v)
				}
				if v, ok := pMap["state"].(string); ok {
					port.State = strings.TrimSpace(v)
				}
				if v, ok := pMap["protocol"].(string); ok {
					port.Protocol = strings.TrimSpace(v)
				}
				if v, ok := pMap["service"].(string); ok {
					port.Service = strings.TrimSpace(v)
				}
				if v, ok := pMap["banner"].(string); ok {
					port.Banner = v
				}
				if v, ok := pMap["version"].(string); ok {
					port.Version = v
				}
				result.Ports = append(result.Ports, port)
			}
		}
	}

	// Конвертируем протоколы
	if protosRaw, ok := m["protocols"].([]interface{}); ok {
		for _, pRaw := range protosRaw {
			if s, ok := pRaw.(string); ok {
				result.Protocols = append(result.Protocols, strings.TrimSpace(s))
			}
		}
	}

	return result
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

	// Преобразуем map в scanner.Result
	oldResults := make([]scanner.Result, 0, len(req.OldHosts))
	for _, m := range req.OldHosts {
		oldResults = append(oldResults, mapToScannerResult(m))
	}

	newResults := make([]scanner.Result, 0, len(req.NewHosts))
	for _, m := range req.NewHosts {
		newResults = append(newResults, mapToScannerResult(m))
	}

	alerts := alertingEng.CheckAlerts(oldResults, newResults)

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

// mapToPortInfo преобразует map[string]interface{} в scanner.PortInfo
func mapToPortInfo(m map[string]interface{}) scanner.PortInfo {
	p := scanner.PortInfo{}
	if v, ok := m["port"].(float64); ok {
		p.Port = int(v)
	}
	if v, ok := m["state"].(string); ok {
		p.State = strings.TrimSpace(v)
	}
	if v, ok := m["protocol"].(string); ok {
		p.Protocol = strings.TrimSpace(v)
	}
	if v, ok := m["service"].(string); ok {
		p.Service = strings.TrimSpace(v)
	}
	if v, ok := m["banner"].(string); ok {
		p.Banner = v
	}
	if v, ok := m["version"].(string); ok {
		p.Version = v
	}
	return p
}

// parsePortFromMap парсит port из map (для обратной совместимости)
func parsePortFromMap(m map[string]interface{}) int {
	if v, ok := m["port"].(float64); ok {
		return int(v)
	}
	if v, ok := m["port"].(string); ok {
		p, err := strconv.Atoi(v)
		if err == nil {
			return p
		}
	}
	return 0
}
