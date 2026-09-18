package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/scanner"
)

// TestHandleScan_WithDeps — сканирование с внедрённым ScannerService.
func TestHandleScan_WithDeps(t *testing.T) {
	cfg := DefaultConfig()
	mockSvc := newMockScanService()
	deps := ScanDeps{
		ScannerService: mockSvc,
	}
	handler := NewHandlerWithDeps(cfg, deps)

	body, _ := json.Marshal(map[string]interface{}{
		"network":    "192.168.1.0/24",
		"port_range": "1-1000",
	})

	req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleScan(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	var resp scanResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "running" {
		t.Errorf("expected status 'running', got %s", resp.Status)
	}

	// Сканирование уходит в фоновую горутину — ждём фактического вызова,
	// а не читаем флаг сразу после возврата обработчика.
	if !mockSvc.waitForCall(5 * time.Second) {
		t.Fatal("mockScanService.Scan() was not called within 5s")
	}

	if !mockSvc.wasScanned() {
		t.Error("mockScanService.Scan() was not called")
	}

	if err := mockSvc.contextErrAtEntry(); err != nil {
		t.Errorf("scan started with already-canceled context: %v", err)
	}
}

// TestHandleScan_NoBody — пустое тело запроса.
func TestHandleScan_NoBody(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("POST", "/api/v1/scan", http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestHandleScan_NetworkRequired — network required.
func TestHandleScan_NetworkRequired(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"port_range": "1-1000",
	})

	req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestHandleResults_NoResultsYet — результаты отсутствуют.
func TestHandleResults_NoResultsYet(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	// Очистим store перед тестом
	scanStoreInstance.globalMu.Lock()
	scanStoreInstance.mu.Lock()
	scanStoreInstance.scans = make(map[string]*scanState)
	scanStoreInstance.mu.Unlock()
	scanStoreInstance.globalMu.Unlock()

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
		t.Errorf("expected message 'no results available', got %v", resp["message"])
	}
}

// TestHandleScanStatus_CheckValid — валидный статус сканирования.
func TestHandleScanStatus_CheckValid(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	// Сначала запустим сканирование
	body, _ := json.Marshal(map[string]interface{}{
		"network": "192.168.1.0/24",
	})
	req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	var resp scanResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	// Теперь проверим статус
	time.Sleep(100 * time.Millisecond)

	req2 := httptest.NewRequest("GET", "/api/v1/scan/"+resp.ID, nil)
	w2 := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	var status scanStatus
	_ = json.NewDecoder(w2.Body).Decode(&status)

	if status.ID != resp.ID {
		t.Errorf("expected scan ID %s, got %s", resp.ID, status.ID)
	}
}

// TestCorsMiddleware_TrustedOrigin — CORS с разрешённым origin.
func TestCorsMiddleware_TrustedOrigin(t *testing.T) {
	cfg := Config{
		EnableCORS:         true,
		AllowedOrigins:     []string{"http://localhost:3000", "http://example.com"},
		RateLimitPerSecond: 100,
	}
	handler := NewHandler(cfg)

	req := httptest.NewRequest("OPTIONS", "/api/v1/scan", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.corsMiddleware(next).ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://localhost:3000', got %q", origin)
	}

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}

	headers := w.Header().Get("Access-Control-Allow-Headers")
	if headers == "" {
		t.Error("expected Access-Control-Allow-Headers header")
	}
}

// TestCorsMiddleware_WildcardOrigin — wildcard origin behavior.
func TestCorsMiddleware_WildcardOrigin(t *testing.T) {
	cfg := Config{
		EnableCORS:         true,
		AllowedOrigins:     []string{"*"},
		RateLimitPerSecond: 100,
	}
	handler := NewHandler(cfg)

	req := httptest.NewRequest("OPTIONS", "/api/v1/scan", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.corsMiddleware(next).ServeHTTP(w, req)

	// При wildcard AllowedOrigins middleware устанавливает Request Origin
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "" {
		t.Error("expected non-empty Access-Control-Allow-Origin")
	}
}

// TestResponseWriter_SetCode — обёртка responseWriter.
func TestResponseWriter_SetCode(t *testing.T) {
	rrw := &responseWriter{
		ResponseWriter: httptest.NewRecorder(),
		statusCode:     http.StatusOK,
	}

	rrw.WriteHeader(http.StatusCreated)

	if rrw.statusCode != http.StatusCreated {
		t.Errorf("expected statusCode 201, got %d", rrw.statusCode)
	}
}

// TestWriteJSON_Success — запись JSON ответа.
func TestWriteJSON_Success(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	handler.writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("expected Content-Type application/json")
	}

	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp["key"] != "value" {
		t.Errorf("expected key 'value', got %q", resp["key"])
	}
}

// TestWriteJSON_NilData — запись JSON с nil data.
func TestWriteJSON_NilData(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	w := httptest.NewRecorder()

	handler.writeJSON(w, http.StatusOK, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestWriteError_BadRequest — запись ошибки.
func TestWriteError_BadRequest(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	w := httptest.NewRecorder()

	handler.writeError(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp["error"] != "test error" {
		t.Errorf("expected error 'test error', got %q", resp["error"])
	}
}

// TestWriteError_InternalServerError — Internal Server Error.
func TestWriteError_InternalServerError(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	w := httptest.NewRecorder()

	handler.writeError(w, http.StatusInternalServerError, "internal error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// TestHandleHealth_Details — health check extended.
func TestHandleHealth_Details(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	handler.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}

	if resp["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", resp["version"])
	}

	if resp["timestamp"] == nil {
		t.Error("expected non-nil timestamp")
	}
}

// TestScanStore_MultipleScans — несколько сканирований.
func TestScanStore_MultipleScans(t *testing.T) {
	// Очистим store
	scanStoreInstance.globalMu.Lock()
	scanStoreInstance.mu.Lock()
	scanStoreInstance.scans = make(map[string]*scanState)
	scanStoreInstance.mu.Unlock()
	scanStoreInstance.globalMu.Unlock()

	cfg := DefaultConfig()
	router := NewRouter(cfg)

	// Запустим 3 сканирования
	for i := 0; i < 3; i++ {
		body, _ := json.Marshal(map[string]interface{}{
			"network": "192.168.1.0/24",
		})
		req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.GetRouter().ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("scan %d: expected status 202, got %d", i, w.Code)
		}
	}

	// Проверим что все 3 сканирования сохранены
	scanStoreInstance.globalMu.Lock()
	scanStoreInstance.mu.RLock()
	count := len(scanStoreInstance.scans)
	scanStoreInstance.mu.RUnlock()
	scanStoreInstance.globalMu.Unlock()

	if count != 3 {
		t.Errorf("expected 3 scans, got %d", count)
	}
}

// TestScanStore_GetLastCompleted — получение последнего завершённого.
func TestScanStore_GetLastCompleted(t *testing.T) {
	// Очистим store
	scanStoreInstance.globalMu.Lock()
	scanStoreInstance.mu.Lock()
	scanStoreInstance.scans = make(map[string]*scanState)
	scanStoreInstance.mu.Unlock()
	scanStoreInstance.globalMu.Unlock()

	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	// Запустим 2 сканирования
	for i := 0; i < 2; i++ {
		body, _ := json.Marshal(map[string]interface{}{
			"network": "192.168.1.0/24",
		})
		req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.handleScan(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("scan %d: expected status 202, got %d", i, w.Code)
		}
	}

	// Проверим что оба сохранены
	scanStoreInstance.globalMu.Lock()
	scanStoreInstance.mu.RLock()
	count := len(scanStoreInstance.scans)
	scanStoreInstance.mu.RUnlock()
	scanStoreInstance.globalMu.Unlock()

	if count != 2 {
		t.Errorf("expected 2 scans, got %d", count)
	}
}

// mockScanService — mock для тестирования.
//
// handleScan выполняет сканирование в фоновой горутине и сразу отвечает 202,
// поэтому mock сигнализирует о вызове через канал. Читать поле scanned из
// теста напрямую нельзя: это data race и зависимость от планировщика.
type mockScanService struct {
	mu      sync.Mutex
	scanned bool
	ctxErr  error
	calls   chan struct{}
}

func newMockScanService() *mockScanService {
	return &mockScanService{calls: make(chan struct{}, 16)}
}

func (m *mockScanService) Scan(ctx context.Context, cfg contracts.ScanConfig, onProgress contracts.ProgressHandler) ([]contracts.ScanResult, error) {
	m.mu.Lock()
	m.scanned = true
	m.ctxErr = ctx.Err()
	m.mu.Unlock()

	select {
	case m.calls <- struct{}{}:
	default:
	}

	if onProgress != nil {
		onProgress("scanning", 1, 1, "done")
	}

	return []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "test.local", Ports: []contracts.PortInfo{{Port: 80, State: "open"}}},
	}, nil
}

func (m *mockScanService) Stop() {}

// waitForCall — ждёт вызова Scan, возвращает false по таймауту.
func (m *mockScanService) waitForCall(timeout time.Duration) bool {
	select {
	case <-m.calls:
		return true
	case <-time.After(timeout):
		return false
	}
}

// wasScanned — потокобезопасное чтение флага вызова.
func (m *mockScanService) wasScanned() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.scanned
}

// contextErrAtEntry — ошибка контекста на входе в Scan.
// Должна быть nil: отменённый контекст означает, что фоновое сканирование
// убили раньше, чем оно началось.
func (m *mockScanService) contextErrAtEntry() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ctxErr
}

// TestScannerService_NewService — создание ScannerService.
func TestScannerService_NewService(t *testing.T) {
	svc := scanner.NewService("info")
	if svc == nil {
		t.Fatal("NewService() returned nil")
	}

	svc.Stop()
}

// TestDefaultConfig_Check — проверка config по умолчанию.
func TestDefaultConfig_Check(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.EnableCORS {
		t.Error("expected EnableCORS true")
	}

	if cfg.RateLimitPerSecond <= 0 {
		t.Errorf("expected RateLimitPerSecond > 0, got %d", cfg.RateLimitPerSecond)
	}
}

// TestRouter_GetRouter — получение router.
func TestRouter_GetRouter(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	h := router.GetRouter()
	if h == nil {
		t.Fatal("GetRouter() returned nil")
	}
}
