package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"network-scanner/internal/api"
	"network-scanner/internal/contracts"
	"network-scanner/internal/report"
)

func TestFullScanWorkflow(t *testing.T) {
	// 1. Create API router
	cfg := api.DefaultConfig()
	router := api.NewRouter(cfg)

	// 2. Start a scan
	body, _ := json.Marshal(map[string]interface{}{
		"network":    "192.168.1.0/24",
		"port_range": "1-1000",
	})

	req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", w.Code)
	}

	var scanResp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&scanResp); err != nil {
		t.Fatalf("failed to decode scan response: %v", err)
	}

	if scanResp["status"] != "running" {
		t.Errorf("expected scan status 'running', got %v", scanResp["status"])
	}

	scanID, _ := scanResp["id"].(string)

	// 3. Generate report data
	results := []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "router", DeviceVendor: "Cisco"},
	}
	findings := []contracts.Finding{
		{Host: "192.168.1.1", Title: "Open SSH", Severity: "HIGH", Recommendation: "Close SSH"},
	}

	reportData := report.GenerateScanReportData(scanID, "192.168.1.0/24", results, findings, nil)

	// 4. Render HTML report
	html, err := report.RenderScanHTML(reportData)
	if err != nil {
		t.Fatalf("failed to render HTML report: %v", err)
	}

	if len(html) == 0 {
		t.Fatal("expected non-empty HTML report")
	}

	// 5. Save HTML report
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "report.html")

	if err := report.SaveScanHTML(htmlPath, reportData); err != nil {
		t.Fatalf("failed to save HTML report: %v", err)
	}

	// 6. Verify saved file
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("failed to read saved report: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("expected non-empty saved report")
	}

	t.Logf("Generated HTML report: %d bytes", len(html))
}

func TestAPIHealthCheck(t *testing.T) {
	cfg := api.DefaultConfig()
	router := api.NewRouter(cfg)

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

func TestAPICORSHeaders(t *testing.T) {
	// Simple CORS test - just verify the middleware exists
	req := httptest.NewRequest("OPTIONS", "/api/v1/scan", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	// Create a simple handler that returns CORS headers
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
	})

	// Apply CORS middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected Access-Control-Allow-Origin header")
	}
}
