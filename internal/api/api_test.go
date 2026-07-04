package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}
}

func TestHandleScan(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"network": "192.168.1.0/24",
		"port_range": "1-1000",
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

	if resp.Status != "running" {
		t.Errorf("expected status 'running', got %s", resp.Status)
	}
}

func TestHandleScan_MissingNetwork(t *testing.T) {
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

func TestHandleScanStatus_NotFound(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/scan/non-existent", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleInventoryList(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/v1/inventory", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleDocs(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCorsMiddleware(t *testing.T) {
	cfg := DefaultConfig()
	handler := NewHandler(cfg)

	req := httptest.NewRequest("OPTIONS", "/api/v1/scan", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	// Create a simple handler to test middleware
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.corsMiddleware(next).ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected Access-Control-Allow-Origin header")
	}
}
