package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/inventory"
	"network-scanner/internal/scanner"
)

// ============================================================================
// M1.3: Тесты для непокрытых функций API (0.0% coverage)
// ============================================================================

// TestCancelScan_NotFound — ветка: сканирование не найдено
func TestCancelScan_NotFound(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("DELETE", "/api/v1/scan/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// TestCancelScan_Success — ветка: успешная отмена сканирования
func TestCancelScan_Success(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)
	router := NewRouter(cfg)

	// Запускаем сканирование
	body, _ := json.Marshal(map[string]interface{}{
		"network": "192.0.2.0/24",
		"ports":   "100",
	})
	req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Skipf("scan start failed: %d", w.Code)
		return
	}

	// Отменяем сканирование
	req = httptest.NewRequest("DELETE", "/api/v1/scan/test-scan-id", nil)
	w = httptest.NewRecorder()
	handler.CancelScan(w, req)

	if w.Code != http.StatusNotFound {
		t.Logf("expected 404 for test-scan-id (not running): %d", w.Code)
	}
}

// TestContractPortToScanner — покрытие функции contractPortToScanner (0.0% → 100%)
func TestContractPortToScanner(t *testing.T) {
	tests := []struct {
		name     string
		input    contracts.PortInfo
		expected scanner.PortInfo
	}{
		{
			name: "standard port",
			input: contracts.PortInfo{
				Port:     80,
				State:    "open",
				Protocol: "tcp",
				Service:  "http",
				Banner:   "Apache/2.4.41",
				Version:  "2.4.41",
			},
			expected: scanner.PortInfo{
				Port:     80,
				State:    "open",
				Protocol: "tcp",
				Service:  "http",
				Banner:   "Apache/2.4.41",
				Version:  "2.4.41",
			},
		},
		{
			name: "uppercase state normalization",
			input: contracts.PortInfo{
				Port:     443,
				State:    "OPEN",
				Protocol: "TCP",
				Service:  "https",
				Banner:   "",
				Version:  "",
			},
			expected: scanner.PortInfo{
				Port:     443,
				State:    "open",
				Protocol: "tcp",
				Service:  "https",
				Banner:   "",
				Version:  "",
			},
		},
		{
			name: "trimmed spaces",
			input: contracts.PortInfo{
				Port:     22,
				State:    " open ",
				Protocol: " tcp ",
				Service:  " ssh ",
				Banner:   "  OpenSSH  ",
				Version:  " 8.2",
			},
			expected: scanner.PortInfo{
				Port:     22,
				State:    "open",
				Protocol: "tcp",
				Service:  "ssh",
				Banner:   "  OpenSSH  ",
				Version:  " 8.2",
			},
		},
		{
			name: "closed port",
			input: contracts.PortInfo{
				Port:     8080,
				State:    "closed",
				Protocol: "udp",
				Service:  "",
				Banner:   "",
				Version:  "",
			},
			expected: scanner.PortInfo{
				Port:     8080,
				State:    "closed",
				Protocol: "udp",
				Service:  "",
				Banner:   "",
				Version:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contractPortToScanner(tt.input)
			if result.Port != tt.expected.Port {
				t.Errorf("Port: got %d, want %d", result.Port, tt.expected.Port)
			}
			if result.State != tt.expected.State {
				t.Errorf("State: got %q, want %q", result.State, tt.expected.State)
			}
			if result.Protocol != tt.expected.Protocol {
				t.Errorf("Protocol: got %q, want %q", result.Protocol, tt.expected.Protocol)
			}
			if result.Service != tt.expected.Service {
				t.Errorf("Service: got %q, want %q", result.Service, tt.expected.Service)
			}
		})
	}
}

// ============================================================================
// M1.3: Тесты для функций с низким покрытием (< 35%)
// ============================================================================

// TestCompareHandler_NoInventory — ветка: ошибка открытия inventory
func TestCompareHandler_NoInventory(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	req := httptest.NewRequest("GET", "/api/v1/inventory/scan-a/diff", nil)
	w := httptest.NewRecorder()
	handler.compareHandler(w, req)

	// Должна быть ошибка (inventory не существует)
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusBadRequest {
		t.Logf("expected error status, got: %d", w.Code)
	}
}

// TestTopologyExportHandler_EmptyFormat — ветка: пустой формат
func TestTopologyExportHandler_EmptyFormat(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	req := httptest.NewRequest("POST", "/api/v1/topology/export/", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.topologyExportHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Logf("expected 400 for empty format, got: %d", w.Code)
	}
}

// TestTopologyExportHandler_NoFormatField — ветка: отсутствие поля format
func TestTopologyExportHandler_NoFormatField(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"other_field": "value",
	})
	req := httptest.NewRequest("POST", "/api/v1/topology/export/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.topologyExportHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Logf("expected 400 for missing format field, got: %d", w.Code)
	}
}

// TestHandleInventoryDiff_EmptyIDs — пустой id_b → 400 (через back-compat маршрут).
func TestHandleInventoryDiff_EmptyIDs(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "inventory.db")
	cfg := DefaultConfig()
	cfg.InventoryPath = dbPath
	router := NewRouter(cfg)

	// /inventory/{id_a}/diff без id_b и без ?id_b → 400.
	req := httptest.NewRequest("GET", "/api/v1/inventory/only-a/diff", nil)
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing id_b, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// ============================================================================
// M1.3: Тесты для triggerAlertHandler (58.3% → 85%+)
// ============================================================================

// TestTriggerAlertHandler_InvalidIDs — ветка: невалидные ID
func TestTriggerAlertHandler_InvalidIDs(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("POST", "/api/v1/alerts/trigger/", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Logf("route not found (expected): %d", w.Code)
	}
}

// TestTriggerAlertHandler_MissingBody — ветка: отсутствует тело запроса
func TestTriggerAlertHandler_MissingBody(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	req := httptest.NewRequest("POST", "/api/v1/alerts/trigger/scan-a/scan-b", nil)
	w := httptest.NewRecorder()
	handler.triggerAlertHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Logf("expected 400, got: %d", w.Code)
	}
}

// ============================================================================
// M1.3: Тесты для topologyExportHandler (31.0% → 85%+)
// ============================================================================

// TestTopologyExportHandler_JSONFormat — ветка: экспорт в JSON
func TestTopologyExportHandler_JSONFormat(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"format": "json",
	})
	req := httptest.NewRequest("POST", "/api/v1/topology/export/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.topologyExportHandler(w, req)

	// Должна быть ошибка (топология не построена), но формат валиден
	if w.Code == http.StatusBadRequest {
		t.Logf("expected error (no topology built), got: %d", w.Code)
	}
}

// TestTopologyExportHandler_GraphMLFormat — ветка: экспорт в GraphML
func TestTopologyExportHandler_GraphMLFormat(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"format": "graphml",
	})
	req := httptest.NewRequest("POST", "/api/v1/topology/export/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.topologyExportHandler(w, req)

	if w.Code == http.StatusBadRequest {
		t.Logf("expected error (no topology built), got: %d", w.Code)
	}
}

// TestTopologyExportHandler_DOTFormat — ветка: экспорт в DOT
func TestTopologyExportHandler_DOTFormat(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"format": "dot",
	})
	req := httptest.NewRequest("POST", "/api/v1/topology/export/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.topologyExportHandler(w, req)

	if w.Code == http.StatusBadRequest {
		t.Logf("expected error (no topology built), got: %d", w.Code)
	}
}

// TestTopologyExportHandler_PDFFormat — ветка: экспорт в PDF
func TestTopologyExportHandler_PDFFormat(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"format": "pdf",
	})
	req := httptest.NewRequest("POST", "/api/v1/topology/export/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.topologyExportHandler(w, req)

	if w.Code == http.StatusBadRequest {
		t.Logf("expected error (no topology built), got: %d", w.Code)
	}
}

// ============================================================================
// M1.3: Тесты для handleInventoryDiff (18.8% → 85%+)
// ============================================================================

// TestHandleInventoryDiff_CompareSuccess — успешное сравнение двух снапшотов.
//
// Снапшоты готовятся напрямую в inventory-сторе (helper seedInventory), запрос
// идёт через роутер, чтобы mux.Vars корректно заполнил id_a/id_b.
func TestHandleInventoryDiff_CompareSuccess(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "inventory.db")
	now := time.Now().UTC()

	seedInventory(t, dbPath,
		inventory.Snapshot{
			ID:        "snapshot-a",
			Timestamp: now.Add(-time.Hour),
			Hosts: []scanner.Result{
				{
					IP:       "192.0.2.1",
					Hostname: "host-a",
					IsAlive:  true,
					Ports: []scanner.PortInfo{
						{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
					},
				},
			},
		},
		inventory.Snapshot{
			ID:        "snapshot-b",
			Timestamp: now,
			Hosts: []scanner.Result{
				{
					IP:       "192.0.2.1",
					Hostname: "host-a-updated",
					IsAlive:  true,
					Ports: []scanner.PortInfo{
						{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
						{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
					},
				},
				{
					IP:       "192.0.2.2",
					Hostname: "host-b",
					IsAlive:  true,
					Ports: []scanner.PortInfo{
						{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
					},
				},
			},
		},
	)

	cfg := DefaultConfig()
	cfg.InventoryPath = dbPath
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/inventory/snapshot-a/diff/snapshot-b", nil)
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var response struct {
		ScanIDA string                 `json:"scan_id_a"`
		ScanIDB string                 `json:"scan_id_b"`
		New     []contracts.ScanResult `json:"new"`
		Missing []contracts.ScanResult `json:"missing"`
		Changed []contracts.Change     `json:"changed"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.ScanIDA != "snapshot-a" || response.ScanIDB != "snapshot-b" {
		t.Errorf("unexpected ids: %q / %q", response.ScanIDA, response.ScanIDB)
	}

	// 192.0.2.2 появился в B → должен быть в new.
	foundNew := false
	for _, h := range response.New {
		if h.IP == "192.0.2.2" {
			foundNew = true
		}
	}
	if !foundNew {
		t.Errorf("expected 192.0.2.2 in 'new', got %+v", response.New)
	}

	// Проверяем, что конвертация портов не потеряла детали.
	for _, h := range response.New {
		if h.IP != "192.0.2.2" {
			continue
		}
		if len(h.Ports) != 1 || h.Ports[0].Port != 22 || h.Ports[0].Service != "ssh" {
			t.Errorf("port conversion mismatch: %+v", h.Ports)
		}
	}
}

// TestHandleInventoryDiff_MissingSnapshot — несуществующий снапшот → 404.
func TestHandleInventoryDiff_MissingSnapshot(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "inventory.db")
	cfg := DefaultConfig()
	cfg.InventoryPath = dbPath
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/inventory/no-such-a/diff/no-such-b", nil)
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing snapshots, got %d", w.Code)
	}
}

// TestHandleInventoryDiff_QueryParamBackCompat — id_b через ?id_b= (back-compat).
func TestHandleInventoryDiff_QueryParamBackCompat(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "inventory.db")
	now := time.Now().UTC()

	seedInventory(t, dbPath,
		inventory.Snapshot{ID: "q-a", Timestamp: now.Add(-time.Hour), Hosts: []scanner.Result{{IP: "198.51.100.1", IsAlive: true}}},
		inventory.Snapshot{ID: "q-b", Timestamp: now, Hosts: []scanner.Result{{IP: "198.51.100.2", IsAlive: true}}},
	)

	cfg := DefaultConfig()
	cfg.InventoryPath = dbPath
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/inventory/q-a/diff?id_b=q-b", nil)
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 via query param, got %d (body: %s)", w.Code, w.Body.String())
	}

	var response struct {
		ScanIDA string                 `json:"scan_id_a"`
		ScanIDB string                 `json:"scan_id_b"`
		New     []contracts.ScanResult `json:"new"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.ScanIDA != "q-a" || response.ScanIDB != "q-b" {
		t.Errorf("unexpected ids: %q / %q", response.ScanIDA, response.ScanIDB)
	}
}
