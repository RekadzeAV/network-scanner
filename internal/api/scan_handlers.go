package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/scanner"

	"github.com/gorilla/mux"
)

// scanRequest запрос на сканирование
type scanRequest struct {
	NetworkCIDR string `json:"network"`
	PortRange   string `json:"port_range"`
	Timeout     int    `json:"timeout"`
	Threads     int    `json:"threads"`
	ScanUDP     bool   `json:"scan_udp"`
	GrabBanners bool   `json:"grab_banners"`
	OSActive    bool   `json:"os_active"`
	VerboseLogs bool   `json:"verbose_logs"`
	Security    bool   `json:"security"`
	Topology    bool   `json:"topology"`
}

// scanResponse ответ на сканирование
type scanResponse struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Results     []contracts.ScanResult `json:"results,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// scanStatus статус сканирования
type scanStatus struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	Message     string     `json:"message"`
	Progress    int        `json:"progress"`
	Results     int        `json:"results_count,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// scanStore хранит состояния сканирований
type scanStore struct {
	mu       sync.RWMutex
	scans    map[string]*scanState
	resetMu  sync.Mutex // Separate mutex for reset to avoid deadlock
	globalMu sync.Mutex // Global mutex for test isolation
}

type scanState struct {
	ID          string
	Status      string
	Message     string
	Results     []contracts.ScanResult
	StartedAt   time.Time
	CompletedAt *time.Time
	Progress    int
	cancel      context.CancelFunc
}

var scanStoreInstance = &scanStore{
	scans:   make(map[string]*scanState),
	resetMu: sync.Mutex{},
}

// handleScan запускает сканирование с реального ScannerService
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

	// Генерируем уникальный ID для сканирования
	scanID := generateScanID()

	// Создаём cancellable контекст с таймаутом 10 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Конвертируем запрос в ScanConfig
	cfg := contracts.ScanConfig{
		NetworkCIDR: req.NetworkCIDR,
		PortRange:   req.PortRange,
		Timeout:     time.Duration(req.Timeout) * time.Second,
		Threads:     req.Threads,
		ScanUDP:     req.ScanUDP,
		GrabBanners: req.GrabBanners,
		OSActive:    req.OSActive,
		VerboseLogs: req.VerboseLogs,
	}

	// Создаём ScannerService из deps или создаём новый
	var svc contracts.ScannerService
	if h.deps.ScannerService != nil {
		svc = h.deps.ScannerService
	} else {
		svc = scanner.NewService("info")
	}

	// Сохраняем scan state
	scanStoreInstance.globalMu.Lock()
	scanStoreInstance.mu.Lock()
	scanStoreInstance.scans[scanID] = &scanState{
		ID:        scanID,
		Status:    "running",
		Message:   "scan started",
		StartedAt: time.Now(),
		Progress:  0,
		cancel:    cancel,
	}
	scanStoreInstance.mu.Unlock()
	scanStoreInstance.globalMu.Unlock()

	// Запускаем сканирование в фоне
	go func() {
		results, _ := svc.Scan(ctx, cfg, func(stage string, current, total int, message string) {
			// Обновляем прогресс
			progress := 0
			if total > 0 {
				progress = int(float64(current) / float64(total) * 100)
			}

			scanStoreInstance.mu.Lock()
			if s, exists := scanStoreInstance.scans[scanID]; exists {
				s.Progress = progress
				s.Message = message
			}
			scanStoreInstance.mu.Unlock()
		})

		completedAt := time.Now()

		scanStoreInstance.mu.Lock()
		if s, exists := scanStoreInstance.scans[scanID]; exists {
			s.Status = "completed"
			s.Message = "scan completed successfully"
			s.Results = results
			s.CompletedAt = &completedAt
			s.Progress = 100
		}
		scanStoreInstance.mu.Unlock()
	}()

	// Возвращаем immediate response
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

	scanStoreInstance.globalMu.Lock()
	scanStoreInstance.mu.RLock()
	scan, exists := scanStoreInstance.scans[scanID]
	scanStoreInstance.mu.RUnlock()
	scanStoreInstance.globalMu.Unlock()

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
	scanStoreInstance.globalMu.Lock()
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
	scanStoreInstance.globalMu.Unlock()

	if lastScan == nil {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"results": []contracts.ScanResult{},
			"message": "no results available",
			"scan_id": "",
		})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": lastScan.Results,
		"scan_id": lastScan.ID,
		"message": "results ok",
	})
}

// CancelScan отменяет активное сканирование
func (h *Handler) CancelScan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID := vars["id"]

	scanStoreInstance.globalMu.Lock()
	scanStoreInstance.mu.Lock()
	scan, exists := scanStoreInstance.scans[scanID]
	if exists && scan.Status == "running" && scan.cancel != nil {
		scan.cancel()
		scan.Status = "cancelled"
		scan.Message = "scan cancelled by user"
	}
	scanStoreInstance.mu.Unlock()
	scanStoreInstance.globalMu.Unlock()

	if !exists {
		h.writeError(w, http.StatusNotFound, "scan not found")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "scan cancelled",
		"scan_id": scanID,
	})
}
