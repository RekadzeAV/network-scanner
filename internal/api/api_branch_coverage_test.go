package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"network-scanner/internal/inventory"
	"network-scanner/internal/scanner"
)

// Файл закрывает ветки обработчиков, которые существующие тесты обходят
// стороной: ошибки открытия inventory, выбор снапшота по ID, экспорт
// топологии во все три формата, отмена активного сканирования.

// newRouterWithInventory — роутер с inventory во временном каталоге.
func newRouterWithInventory(t *testing.T, snapshots ...inventory.Snapshot) *Router {
	t.Helper()

	cfg := DefaultConfig()
	cfg.InventoryPath = filepath.Join(t.TempDir(), "inventory.db")
	seedInventory(t, cfg.InventoryPath, snapshots...)

	return NewRouter(cfg)
}

// newRouterWithInventoryPath — роутер с inventory по явному пути.
func newRouterWithInventoryPath(path string) *Router {
	cfg := DefaultConfig()
	cfg.InventoryPath = path
	return NewRouter(cfg)
}

// --- utils.go: нестроковые/небулевы значения ---

func TestMapStringNonStringValue(t *testing.T) {
	if got := mapString(map[string]interface{}{"k": 42}, "k"); got != "" {
		t.Errorf("mapString(int) = %q, want empty", got)
	}
	if got := mapString(map[string]interface{}{"k": "  x  "}, "k"); got != "x" {
		t.Errorf("mapString(spaced) = %q, want %q", got, "x")
	}
}

func TestMapBoolNonBoolValue(t *testing.T) {
	if got := mapBool(map[string]interface{}{"k": "true"}, "k"); got {
		t.Error("mapBool(string) = true, want false")
	}
	if got := mapBool(map[string]interface{}{"k": true}, "k"); !got {
		t.Error("mapBool(true) = false, want true")
	}
}

// --- inventory: ошибки открытия хранилища ---

func TestHandleInventoryList_OpenError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory", nil)
	w := httptest.NewRecorder()

	newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleInventorySave_OpenError(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"id":      "scan-open-fail",
		"results": []map[string]interface{}{{"ip": "127.0.0.1"}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleInventoryDiff_OpenError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/a/diff/b", nil)
	w := httptest.NewRecorder()

	newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleInventoryDiff_MissingHostWithPorts закрывает цикл конвертации
// портов у «пропавших» хостов (diff.Missing) — у новых хостов он покрыт,
// у отсутствующих — нет.
func TestHandleInventoryDiff_MissingHostWithPorts(t *testing.T) {
	now := time.Now().UTC()
	router := newRouterWithInventory(t,
		inventory.Snapshot{
			ID:        "diff-a",
			Timestamp: now.Add(-time.Hour),
			Hosts: []scanner.Result{{
				IP:       "10.10.0.1",
				Hostname: "gone-host",
				Ports: []scanner.PortInfo{
					{Port: 22, State: "open", Protocol: "tcp", Service: "ssh", Banner: "OpenSSH", Version: "8.9"},
					{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
				},
			}},
		},
		inventory.Snapshot{
			ID:        "diff-b",
			Timestamp: now,
			Hosts:     []scanner.Result{{IP: "10.10.0.2", Hostname: "new-host"}},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/diff-a/diff/diff-b", nil)
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp inventoryDiffResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	if len(resp.Missing) != 1 {
		t.Fatalf("expected 1 missing host, got %d: %s", len(resp.Missing), w.Body.String())
	}
	if len(resp.Missing[0].Ports) != 2 {
		t.Fatalf("expected ports of the missing host, got %d", len(resp.Missing[0].Ports))
	}
	if resp.Missing[0].Ports[0].Banner != "OpenSSH" {
		t.Errorf("banner = %q, want %q", resp.Missing[0].Ports[0].Banner, "OpenSSH")
	}
	if len(resp.New) != 1 || resp.New[0].IP != "10.10.0.2" {
		t.Errorf("new hosts = %+v", resp.New)
	}
}

// --- history / compare ---

func TestHistoryHandler_OpenError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/history", nil)
	w := httptest.NewRecorder()

	newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHistoryHandler_LimitIgnoredOnGarbage(t *testing.T) {
	router := newRouterWithInventory(t, inventory.Snapshot{
		ID:        "hist-1",
		Timestamp: time.Now().UTC(),
		Hosts:     []scanner.Result{{IP: "127.0.0.1", GuessOS: "linux", DeviceVendor: "vmware"}},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history?limit=not-a-number", nil)
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCompareHandler_OpenError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/compare/a/b", nil)
	w := httptest.NewRecorder()

	newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestCompareHandler_Success закрывает успешную ветку compareHandler
// (существовал только тест с несуществующим inventory).
func TestCompareHandler_Success(t *testing.T) {
	now := time.Now().UTC()
	router := newRouterWithInventory(t,
		inventory.Snapshot{
			ID:        "cmp-a",
			Timestamp: now.Add(-time.Hour),
			Hosts:     []scanner.Result{{IP: "10.20.0.1", Hostname: "keep", IsAlive: true}},
		},
		inventory.Snapshot{
			ID:        "cmp-b",
			Timestamp: now,
			Hosts: []scanner.Result{
				{IP: "10.20.0.1", Hostname: "keep", IsAlive: true},
				{IP: "10.20.0.2", Hostname: "added", IsAlive: true},
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/compare/cmp-a/cmp-b", nil)
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		ScanIDA   string            `json:"scan_id_a"`
		ScanIDB   string            `json:"scan_id_b"`
		NewHosts  []json.RawMessage `json:"new_hosts"`
		TotalDiff int               `json:"total_diff"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	if resp.ScanIDA != "cmp-a" || resp.ScanIDB != "cmp-b" {
		t.Errorf("echoed ids = %q / %q", resp.ScanIDA, resp.ScanIDB)
	}
	if len(resp.NewHosts) != 1 {
		t.Errorf("new_hosts = %d, want 1 (добавленный хост)", len(resp.NewHosts))
	}
}

// --- alerting ---

func TestTriggerAlertHandler_OpenError(t *testing.T) {
	withAlertingEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/trigger/a/b", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTriggerAlertHandler_SnapshotBMissing(t *testing.T) {
	withAlertingEngine(t)

	now := time.Now().UTC()
	router := newRouterWithInventory(t, inventory.Snapshot{
		ID:        "alert-a",
		Timestamp: now,
		Hosts:     []scanner.Result{{IP: "10.30.0.1", IsAlive: true}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/trigger/alert-a/alert-missing", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- snmp ---

func TestSNMPCollect_OpenError(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{"community": "public"})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snmp/collect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSNMPCollect_NoSnapshots(t *testing.T) {
	router := newRouterWithInventory(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snmp/collect", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	if resp["message"] != "no snapshots found" {
		t.Errorf("message = %v, want %q", resp["message"], "no snapshots found")
	}
}

func TestSNMPCollect_SnapshotByIDNotFound(t *testing.T) {
	router := newRouterWithInventory(t)

	body, _ := json.Marshal(map[string]interface{}{"device_ids": []string{"missing-scan"}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/snmp/collect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// TestSNMPCollect_ByDeviceIDs закрывает ветку загрузки устройств по ID
// снапшотов. Хосты намеренно без флага SNMPEnabled: коллектор коротко
// замыкается на нулевом списке целей, и тест не ждёт сетевых таймаутов.
func TestSNMPCollect_ByDeviceIDs(t *testing.T) {
	router := newRouterWithInventory(t, inventory.Snapshot{
		ID:        "snmp-scan",
		Timestamp: time.Now().UTC(),
		Hosts:     []scanner.Result{{IP: "127.0.0.1", Hostname: "localhost", IsAlive: true}},
	})

	body, _ := json.Marshal(map[string]interface{}{
		"device_ids": []string{"snmp-scan"},
		"community":  "private",
		"timeout":    1,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snmp/collect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		TotalTargets int `json:"total_targets"`
		SNMPDevices  int `json:"snmp_devices"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	if resp.TotalTargets != 0 || resp.SNMPDevices != 0 {
		t.Errorf("collect = %+v, want нулевые счётчики без SNMP-целей", resp)
	}
}

// --- topology ---

func TestTopologyBuild_OpenError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topology/build", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTopologyBuild_NoSnapshots(t *testing.T) {
	router := newRouterWithInventory(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/topology/build", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	if resp["message"] != "no snapshots found" {
		t.Errorf("message = %v, want %q", resp["message"], "no snapshots found")
	}
}

func TestTopologyBuild_SnapshotNotFound(t *testing.T) {
	router := newRouterWithInventory(t)

	body, _ := json.Marshal(map[string]interface{}{"snapshot_id": "absent"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topology/build", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// TestTopologyBuild_BySnapshotIDWithSNMP закрывает загрузку снапшота по ID
// и ветку SNMP-сбора внутри buildTopologyForRequest. Хосты без флага
// SNMPEnabled — коллектор не выполняет реальных опросов, тест остаётся быстрым.
func TestTopologyBuild_BySnapshotIDWithSNMP(t *testing.T) {
	router := newRouterWithInventory(t, inventory.Snapshot{
		ID:        "topo-scan",
		Timestamp: time.Now().UTC(),
		Hosts: []scanner.Result{
			{IP: "127.0.0.1", MAC: "aa:bb:cc:dd:ee:ff", Hostname: "localhost", IsAlive: true},
			{IP: "127.0.0.2", MAC: "aa:bb:cc:dd:ee:11", Hostname: "other", IsAlive: true},
		},
	})

	body, _ := json.Marshal(map[string]interface{}{
		"snapshot_id":    "topo-scan",
		"snmp_enabled":   true,
		"snmp_community": "public",
		"snmp_timeout":   1,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/topology/build", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		DeviceCount int `json:"device_count"`
		LinkCount   int `json:"link_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	if resp.DeviceCount < 2 {
		t.Errorf("device_count = %d, want >= 2", resp.DeviceCount)
	}
	if resp.LinkCount != 0 {
		t.Logf("link_count = %d (без SNMP-данных связей не ожидается)", resp.LinkCount)
	}
}

// TestTopologyExport_AllFormats закрывает экспорт топологии во все три
// поддерживаемых формата — раньше handler выходил на 400 из-за пустого format.
func TestTopologyExport_AllFormats(t *testing.T) {
	now := time.Now().UTC()
	router := newRouterWithInventory(t, inventory.Snapshot{
		ID:        "export-scan",
		Timestamp: now,
		Hosts: []scanner.Result{
			{IP: "10.40.0.1", MAC: "aa:bb:cc:00:00:01", Hostname: "gw", DeviceType: "router", IsAlive: true},
			{IP: "10.40.0.2", MAC: "aa:bb:cc:00:00:02", Hostname: "host", IsAlive: true},
		},
	})

	tests := []struct {
		format      string
		contentType string
	}{
		{format: "json", contentType: "application/json"},
		{format: "dot", contentType: "text/plain"},
		{format: "graphml", contentType: "application/xml"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			body, _ := json.Marshal(map[string]interface{}{"snapshot_id": "export-scan"})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/topology/export/"+tt.format, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.GetRouter().ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}
			if got := w.Header().Get("Content-Type"); got != tt.contentType {
				t.Errorf("Content-Type = %q, want %q", got, tt.contentType)
			}
			if w.Body.Len() == 0 {
				t.Error("empty export body")
			}
		})
	}
}

func TestTopologyExport_LatestSnapshot(t *testing.T) {
	now := time.Now().UTC()
	router := newRouterWithInventory(t,
		inventory.Snapshot{
			ID:        "old-scan",
			Timestamp: now.Add(-2 * time.Hour),
			Hosts:     []scanner.Result{{IP: "10.41.0.1", MAC: "aa:bb:cc:11:00:01", Hostname: "old"}},
		},
		inventory.Snapshot{
			ID:        "new-scan",
			Timestamp: now,
			Hosts:     []scanner.Result{{IP: "10.41.0.2", MAC: "aa:bb:cc:11:00:02", Hostname: "new"}},
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/topology/export/json", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var topo map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &topo); err != nil {
		t.Fatalf("export is not valid JSON: %v", err)
	}
	devices, ok := topo["Devices"].(map[string]interface{})
	if !ok {
		t.Fatalf("Devices missing in %s", w.Body.String())
	}
	if len(devices) != 1 {
		t.Errorf("devices = %d, want 1 (последний снапшот)", len(devices))
	}
}

func TestTopologyExport_OpenError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topology/export/json", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTopologyDOTAndStats(t *testing.T) {
	router := newRouterWithInventory(t, inventory.Snapshot{
		ID:        "dot-scan",
		Timestamp: time.Now().UTC(),
		Hosts:     []scanner.Result{{IP: "10.42.0.1", MAC: "aa:bb:cc:22:00:01", Hostname: "dev", DeviceType: "switch"}},
	})

	t.Run("dot", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{"snapshot_id": "dot-scan"})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/topology/dot", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.GetRouter().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("graph")) {
			t.Errorf("DOT body looks wrong: %s", w.Body.String())
		}
	})

	t.Run("stats", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{"snapshot_id": "dot-scan"})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/topology/stats", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.GetRouter().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			DeviceCount int            `json:"device_count"`
			LinkCount   int            `json:"link_count"`
			TypeStats   map[string]int `json:"type_stats"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response %q: %v", w.Body.String(), err)
		}
		if resp.DeviceCount != 1 || resp.LinkCount != 0 {
			t.Errorf("stats = %+v", resp)
		}
		if resp.TypeStats["switch"] != 1 {
			t.Errorf("type_stats = %v, want switch=1", resp.TypeStats)
		}
	})
}

// TestTopologyHandlers_OpenError закрывает ветки ошибок открытия inventory
// у dot/stats-обработчиков.
func TestTopologyHandlers_OpenError(t *testing.T) {
	for _, path := range []string{"/api/v1/topology/dot", "/api/v1/topology/stats"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer([]byte("{}")))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			newRouterWithInventoryPath("").GetRouter().ServeHTTP(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

// --- loadHostsForSnapshot: ветки ошибок ---

func TestLoadHostsForSnapshot_Errors(t *testing.T) {
	now := time.Now().UTC()

	cases := []struct {
		name    string
		path    string
		req     topologyRequest
		want    int
		wantErr error
	}{
		{name: "open error", path: "", want: http.StatusInternalServerError},
		{
			name: "snapshot not found",
			path: writeInventoryForTest(t, inventory.Snapshot{ID: "lh-1", Timestamp: now}),
			req:  topologyRequest{SnapshotID: "lh-missing"},
			want: http.StatusNotFound,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.InventoryPath = tt.path
			h := NewHandler(cfg)

			_, failure := h.loadHostsForSnapshot(tt.req)
			if failure == nil {
				t.Fatal("expected failure, got nil")
			}
			if failure.code != tt.want {
				t.Errorf("code = %d, want %d (%s)", failure.code, tt.want, failure.msg)
			}
		})
	}
}

// writeInventoryForTest — создание inventory с указанными снапшотами
// (возвращает путь к БД).
func writeInventoryForTest(t *testing.T, snapshots ...inventory.Snapshot) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "inventory.db")
	seedInventory(t, path, snapshots...)
	return path
}

// --- CancelScan: активное сканирование ---

// TestCancelScan_Running закрывает ветки «сканирование найдено и отменено»
// и успешного ответа. Раньше тест обращался к заведомо несуществующему ID.
func TestCancelScan_Running(t *testing.T) {
	resetScanStore()
	t.Cleanup(resetScanStore)

	cfg := DefaultConfig()
	cfg.InventoryPath = filepath.Join(t.TempDir(), "inventory.db")
	router := NewRouter(cfg)

	body, _ := json.Marshal(map[string]interface{}{
		"network": "198.51.100.0/24",
		"ports":   "1-5",
		"timeout": 1,
	})

	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/scan", bytes.NewBuffer(body))
	startReq.Header.Set("Content-Type", "application/json")
	startW := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(startW, startReq)

	if startW.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", startW.Code, startW.Body.String())
	}

	var started scanResponse
	if err := json.Unmarshal(startW.Body.Bytes(), &started); err != nil {
		t.Fatalf("decode start response %q: %v", startW.Body.String(), err)
	}

	cancelReq := mux.SetURLVars(
		httptest.NewRequest(http.MethodDelete, "/api/v1/scan/"+started.ID, nil),
		map[string]string{"id": started.ID},
	)
	cancelW := httptest.NewRecorder()

	NewHandler(cfg).CancelScan(cancelW, cancelReq)

	if cancelW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", cancelW.Code, cancelW.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(cancelW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode cancel response %q: %v", cancelW.Body.String(), err)
	}
	if resp["scan_id"] != started.ID {
		t.Errorf("scan_id = %v, want %v", resp["scan_id"], started.ID)
	}

	scanStoreInstance.mu.RLock()
	state := scanStoreInstance.scans[started.ID]
	status := ""
	if state != nil {
		status = state.Status
	}
	scanStoreInstance.mu.RUnlock()

	if status != "cancelled" {
		t.Errorf("status = %q, want %q", status, "cancelled")
	}
}
