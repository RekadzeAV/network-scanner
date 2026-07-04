package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"network-scanner/internal/contracts"
)

// scanRequest запрос на сканирование
type scanRequest struct {
	NetworkCIDR  string            `json:"network"`
	PortRange    string            `json:"port_range"`
	Timeout      int               `json:"timeout"`
	Threads      int               `json:"threads"`
	ScanUDP      bool              `json:"scan_udp"`
	GrabBanners  bool              `json:"grab_banners"`
	OSActive     bool              `json:"os_active"`
	VerboseLogs  bool              `json:"verbose_logs"`
	Security     bool              `json:"security"`
	Topology     bool              `json:"topology"`
}

// scanResponse ответ на сканирование
type scanResponse struct {
	ID          string             `json:"id"`
	Status      string             `json:"status"`
	Message     string             `json:"message"`
	Results     []contracts.ScanResult `json:"results,omitempty"`
	StartedAt   time.Time          `json:"started_at"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`
}

// scanStatus статус сканирования
type scanStatus struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	Progress    int       `json:"progress"`
	Results     int       `json:"results_count,omitempty"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// scanStore хранит состояния сканирований
type scanStore struct {
	mu       sync.RWMutex
	scans    map[string]*scanState
}

type scanState struct {
	ID          string
	Status      string
	Message     string
	Results     []contracts.ScanResult
	StartedAt   time.Time
	CompletedAt *time.Time
	Progress    int
}

var scanStoreInstance = &scanStore{
	scans: make(map[string]*scanState),
}

// handleScan запускает сканирование
func (h *Handler) handleScan(w http.ResponseWriter, r *http.Request) {
	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.NetworkCIDR == "" {
		h.writeError(w, http.StatusBadRequest, "network is required")
		return
	}
	if req.PortRange == "" {
		req.PortRange = "1-1000"
	}
	if req.Timeout <= 0 {
		req.Timeout = 2
	}
	if req.Threads <= 0 {
		req.Threads = 50
	}

	// Create scan ID
	scanID := generateScanID()

	// Store scan state
	scanStoreInstance.mu.Lock()
	scanStoreInstance.scans[scanID] = &scanState{
		ID:        scanID,
		Status:    "running",
		Message:   "scan started",
		StartedAt: time.Now(),
		Progress:  0,
	}
	scanStoreInstance.mu.Unlock()

	// Run scan in background
	go func() {
		// TODO: integrate with real scanner service
		// For now, simulate scan
		time.Sleep(2 * time.Second)

		results := []contracts.ScanResult{
			{
				IP:       "192.168.1.1",
				Hostname: "router",
				Ports: []contracts.PortInfo{
					{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
				},
			},
		}

		completedAt := time.Now()

		scanStoreInstance.mu.Lock()
		scanStoreInstance.scans[scanID] = &scanState{
			ID:          scanID,
			Status:      "completed",
			Message:     "scan completed successfully",
			Results:     results,
			StartedAt:   time.Now().Add(-2 * time.Second),
			CompletedAt: &completedAt,
			Progress:    100,
		}
		scanStoreInstance.mu.Unlock()
	}()

	// Return immediate response
	h.writeJSON(w, http.StatusAccepted, scanResponse{
		ID:        scanID,
		Status:    "running",
		Message:   "scan started",
		StartedAt: time.Now(),
	})
}

// handleScanStatus возвращает статус сканирования
func (h *Handler) handleScanStatus(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	vars := mux.Vars(r)
	scanID := vars["id"]

	scanStoreInstance.mu.RLock()
	scan, exists := scanStoreInstance.scans[scanID]
	scanStoreInstance.mu.RUnlock()

	if !exists {
		h.writeError(w, http.StatusNotFound, "scan not found")
		return
	}

	h.writeJSON(w, http.StatusOK, scanStatus{
		ID:          scan.ID,
		Status:      scan.Status,
		Message:     scan.Message,
		Progress:    scan.Progress,
		Results:     len(scan.Results),
		StartedAt:   scan.StartedAt,
		CompletedAt: scan.CompletedAt,
	})
}

// handleResults возвращает результаты сканирования
func (h *Handler) handleResults(w http.ResponseWriter, r *http.Request) {
	// Get last completed scan
	scanStoreInstance.mu.RLock()
	var lastScan *scanState
	for _, scan := range scanStoreInstance.scans {
		if scan.Status == "completed" {
			if lastScan == nil || scan.StartedAt.After(lastScan.StartedAt) {
				lastScan = scan
			}
		}
	}
	scanStoreInstance.mu.RUnlock()

	if lastScan == nil {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"results": []contracts.ScanResult{},
			"message": "no results available",
		})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": lastScan.Results,
		"scan_id": lastScan.ID,
	})
}
