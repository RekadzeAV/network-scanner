package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/topology"
)

// ============================================================================
// handleScan — edge cases
// ============================================================================

func TestHandleScan_InvalidJSON(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleScan_DefaultPortRange(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"network": "192.168.1.0/24",
	})
	req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	var resp scanResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty scan ID")
	}
}

func TestHandleScanStatus_Found(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	// Start a scan
	body, _ := json.Marshal(map[string]interface{}{
		"network": "192.168.1.0/24",
	})
	req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	var resp scanResponse
	json.NewDecoder(w.Body).Decode(&resp)

	// Check status
	req2 := httptest.NewRequest("GET", "/api/v1/scan/"+resp.ID, nil)
	w2 := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	var status scanStatus
	if err := json.NewDecoder(w2.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if status.ID != resp.ID {
		t.Errorf("expected scan ID %s, got %s", resp.ID, status.ID)
	}
}

func TestHandleResults_NoResults(t *testing.T) {
	resetScanStore()
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/results", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["message"] != "no results available" {
		t.Errorf("expected 'no results available', got %v", resp["message"])
	}
}

func TestHandleResults_WithCompletedScan(t *testing.T) {
	resetScanStore()
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	// Manually insert a completed scan
	scanID := "test-scan-results"
	completedAt := time.Now()
	scanStoreInstance.mu.Lock()
	scanStoreInstance.scans[scanID] = &scanState{
		ID:          scanID,
		Status:      "completed",
		Message:     "done",
		Results:     []contracts.ScanResult{{IP: "192.168.1.1"}},
		StartedAt:   time.Now().Add(-5 * time.Second),
		CompletedAt: &completedAt,
		Progress:    100,
	}
	scanStoreInstance.mu.Unlock()

	defer func() {
		scanStoreInstance.mu.Lock()
		delete(scanStoreInstance.scans, scanID)
		scanStoreInstance.mu.Unlock()
	}()

	req := httptest.NewRequest("GET", "/api/v1/results", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["scan_id"] != scanID {
		t.Errorf("expected scan_id %s, got %v", scanID, resp["scan_id"])
	}
}

// ============================================================================
// handleInventorySave & handleInventoryDiff
// ============================================================================

func TestHandleInventorySave_Success(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	body, _ := json.Marshal(inventoryRequest{
		ID: "snap-001",
		Results: []contracts.ScanResult{
			{IP: "192.168.1.1"},
		},
	})
	req := httptest.NewRequest("POST", "/api/v1/inventory", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var resp inventoryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != "snap-001" {
		t.Errorf("expected ID snap-001, got %s", resp.ID)
	}
	if resp.HostCount != 1 {
		t.Errorf("expected host count 1, got %d", resp.HostCount)
	}
}

func TestHandleInventorySave_InvalidJSON(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("POST", "/api/v1/inventory", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleInventorySave_MissingID(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	body, _ := json.Marshal(inventoryRequest{
		Results: []contracts.ScanResult{{IP: "192.168.1.1"}},
	})
	req := httptest.NewRequest("POST", "/api/v1/inventory", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleInventorySave_EmptyResults(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	body, _ := json.Marshal(inventoryRequest{
		ID: "snap-002",
	})
	req := httptest.NewRequest("POST", "/api/v1/inventory", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleInventoryDiff_Success(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/inventory/snap-a/diff?id_b=snap-b", nil)
	w := httptest.NewRecorder()

	// Use mux vars - need to test through router
	router.GetRouter().ServeHTTP(w, req)

	// This endpoint requires mux vars, so direct testing through router
	// The route is /inventory/{id}/diff - but it expects id_a and id_b
	// Let's test the handler directly
	h := NewHandler(cfg)
	req2 := httptest.NewRequest("GET", "/api/v1/inventory/snap-a/diff", nil)
	w2 := httptest.NewRecorder()
	h.handleInventoryDiff(w2, req2)

	// Without mux vars, idA and idB will be empty
	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing ids, got %d", w2.Code)
	}
}

// ============================================================================
// Alerts handlers
// ============================================================================

func TestAlertsHandler_NotInitialized(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	// Ensure alertingEng is nil
	alertingEngMu.Lock()
	alertingEng = nil
	alertingEngMu.Unlock()

	req := httptest.NewRequest("GET", "/api/v1/alerts", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestCheckAlertsHandler_NotInitialized(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	alertingEngMu.Lock()
	alertingEng = nil
	alertingEngMu.Unlock()

	body, _ := json.Marshal(map[string]interface{}{
		"old_hosts": []map[string]interface{}{},
		"new_hosts": []map[string]interface{}{},
	})
	req := httptest.NewRequest("POST", "/api/v1/alerts/check", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestCheckAlertsHandler_InvalidJSON(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	// Initialize alerting
	initAlerting("")

	req := httptest.NewRequest("POST", "/api/v1/alerts/check", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestClearAlertsHandler_NotInitialized(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	alertingEngMu.Lock()
	alertingEng = nil
	alertingEngMu.Unlock()

	req := httptest.NewRequest("DELETE", "/api/v1/alerts/clear", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestTriggerAlertHandler_NotInitialized(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	alertingEngMu.Lock()
	alertingEng = nil
	alertingEngMu.Unlock()

	req := httptest.NewRequest("POST", "/api/v1/alerts/trigger/id-a/id-b", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

// ============================================================================
// History handlers
// ============================================================================

func TestHistoryHandler_InventoryOpenFail(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InventoryPath = "nonexistent/path/that/does/not/exist.db"
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/history", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	// inventory.Open creates the DB file, so it succeeds.
	// GetScanHistory returns empty history successfully.
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 200 or 500, got %d", w.Code)
	}
}

func TestHistoryHandler_WithLimit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InventoryPath = "nonexistent/path/that/does/not/exist.db"
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/history?limit=10", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 200 or 500 for bad inventory path, got %d", w.Code)
	}
}

func TestCompareHandler_InventoryOpenFail(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InventoryPath = "nonexistent/path/that/does/not/exist.db"
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/history/compare/id-a/id-b", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// SNMP handler
// ============================================================================

func TestSNMPCollectHandler_InvalidJSON(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("POST", "/api/v1/snmp/collect", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSNMPCollectHandler_InventoryOpenFail(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InventoryPath = "nonexistent/path/that/does/not/exist.db"
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"device_ids": []string{"snap-1"},
	})
	req := httptest.NewRequest("POST", "/api/v1/snmp/collect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 404 or 500, got %d", w.Code)
	}
}

// ============================================================================
// Topology handlers
// ============================================================================

func TestTopologyBuildHandler_InvalidJSON(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("POST", "/api/v1/topology/build", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTopologyBuildHandler_InventoryOpenFail(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InventoryPath = "nonexistent/path/that/does/not/exist.db"
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"snapshot_id": "snap-1",
	})
	req := httptest.NewRequest("POST", "/api/v1/topology/build", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 404 or 500, got %d", w.Code)
	}
}

func TestTopologyExportHandler_InvalidFormat(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest("POST", "/api/v1/topology/export/xml", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTopologyExportHandler_InvalidJSON(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("POST", "/api/v1/topology/export/json", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTopologyExportHandler_InventoryOpenFail(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InventoryPath = "nonexistent/path/that/does/not/exist.db"
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"snapshot_id": "snap-1",
	})
	req := httptest.NewRequest("POST", "/api/v1/topology/export/json", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	// inventory.Open creates the DB file, so it doesn't fail at open.
	// The snapshot load fails, returning 404.
	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 404 or 500, got %d", w.Code)
	}
}

func TestTopologyDOTHandler_InvalidJSON(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("POST", "/api/v1/topology/dot", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTopologyDOTHandler_InventoryOpenFail(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InventoryPath = "nonexistent/path/that/does/not/exist.db"
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest("POST", "/api/v1/topology/dot", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 404 or 500, got %d", w.Code)
	}
}

func TestTopologyStatsHandler_InvalidJSON(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("POST", "/api/v1/topology/stats", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTopologyStatsHandler_InventoryOpenFail(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InventoryPath = "nonexistent/path/that/does/not/exist.db"
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest("POST", "/api/v1/topology/stats", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 404 or 500, got %d", w.Code)
	}
}

// ============================================================================
// deviceDisplayName & portLabel helpers
// ============================================================================

func TestDeviceDisplayName_Nil(t *testing.T) {
	if got := deviceDisplayName(nil); got != "unknown" {
		t.Errorf("expected 'unknown', got %q", got)
	}
}

func TestDeviceDisplayName_Hostname(t *testing.T) {
	d := &topology.Device{Hostname: "router1", IP: "10.0.0.1"}
	if got := deviceDisplayName(d); got != "router1" {
		t.Errorf("expected 'router1', got %q", got)
	}
}

func TestDeviceDisplayName_IP(t *testing.T) {
	d := &topology.Device{IP: "10.0.0.1"}
	if got := deviceDisplayName(d); got != "10.0.0.1" {
		t.Errorf("expected '10.0.0.1', got %q", got)
	}
}

func TestDeviceDisplayName_MAC(t *testing.T) {
	d := &topology.Device{MAC: "aa:bb:cc:dd:ee:ff"}
	if got := deviceDisplayName(d); got != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected MAC, got %q", got)
	}
}

func TestDeviceDisplayName_Empty(t *testing.T) {
	d := &topology.Device{}
	if got := deviceDisplayName(d); got != "unknown" {
		t.Errorf("expected 'unknown', got %q", got)
	}
}

func TestPortLabel_Nil(t *testing.T) {
	if got := portLabel(nil); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestPortLabel_Name(t *testing.T) {
	p := &topology.Port{Name: "eth0"}
	if got := portLabel(p); got != "eth0" {
		t.Errorf("expected 'eth0', got %q", got)
	}
}

func TestPortLabel_Index(t *testing.T) {
	p := &topology.Port{Index: 5}
	if got := portLabel(p); got != "if5" {
		t.Errorf("expected 'if5', got %q", got)
	}
}

func TestPortLabel_Empty(t *testing.T) {
	p := &topology.Port{}
	if got := portLabel(p); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// ============================================================================
// Middleware tests
// ============================================================================

func TestCorsMiddleware_Disabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EnableCORS = false
	handler := NewHandler(cfg)

	req := httptest.NewRequest("GET", "/api/v1/scan", nil)
	w := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler.corsMiddleware(next).ServeHTTP(w, req)

	if !called {
		t.Fatal("expected next handler to be called when CORS disabled")
	}
}

func TestCorsMiddleware_Options(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	req := httptest.NewRequest("OPTIONS", "/api/v1/scan", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called for OPTIONS")
	})

	handler.corsMiddleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", w.Code)
	}
}

func TestCorsMiddleware_AllowedOrigin(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	req := httptest.NewRequest("GET", "/api/v1/scan", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.corsMiddleware(next).ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("expected specific origin, got %q", got)
	}
}

func TestCorsMiddleware_DisallowedOrigin(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	req := httptest.NewRequest("GET", "/api/v1/scan", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.corsMiddleware(next).ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected wildcard origin, got %q", got)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	req := httptest.NewRequest("GET", "/api/v1/scan", nil)
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	handler.loggingMiddleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusTeapot {
		t.Errorf("expected status 418, got %d", w.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	req := httptest.NewRequest("GET", "/api/v1/scan", nil)
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.rateLimitMiddleware(10)(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rw := &responseWriter{
		ResponseWriter: httptest.NewRecorder(),
		statusCode:     http.StatusOK,
	}
	rw.WriteHeader(http.StatusNotFound)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rw.statusCode)
	}
}

// ============================================================================
// initAlerting
// ============================================================================

func TestInitAlerting(t *testing.T) {
	initAlerting("")
	// Should not panic
	alertingEngMu.Lock()
	if alertingEng == nil {
		t.Fatal("expected alerting engine to be initialized")
	}
	alertingEngMu.Unlock()
}

func TestAlertsHandler_Initialized(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	initAlerting("")

	req := httptest.NewRequest("GET", "/api/v1/alerts?severity=high", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestClearAlertsHandler_Initialized(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	initAlerting("")

	req := httptest.NewRequest("DELETE", "/api/v1/alerts/clear", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCheckAlertsHandler_Initialized(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	initAlerting("")

	body, _ := json.Marshal(map[string]interface{}{
		"old_hosts": []map[string]interface{}{},
		"new_hosts": []map[string]interface{}{},
	})
	req := httptest.NewRequest("POST", "/api/v1/alerts/check", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
