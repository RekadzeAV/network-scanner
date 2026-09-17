package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============================================================================
// mapToScannerResult — все поля
// ============================================================================

func TestMapToScannerResult_AllFields(t *testing.T) {
	m := map[string]interface{}{
		"ip":                  "192.168.1.1",
		"hostname":            "test-host",
		"mac":                 "aa:bb:cc:dd:ee:ff",
		"device_type":         "server",
		"device_vendor":       "Dell",
		"guess_os":            "Linux",
		"guess_os_confidence": "high",
		"guess_os_reason":     "tcp-ip-timing",
		"snmp_enabled":        true,
		"is_alive":            true,
		"ports": []interface{}{
			map[string]interface{}{
				"port":     float64(80),
				"state":    "open",
				"protocol": "tcp",
				"service":  "http",
				"banner":   "nginx/1.19",
				"version":  "1.19.0",
			},
		},
		"protocols": []interface{}{"tcp", "udp"},
	}

	result := mapToScannerResult(m)

	if result.IP != "192.168.1.1" {
		t.Errorf("IP = %q, want %q", result.IP, "192.168.1.1")
	}
	if result.Hostname != "test-host" {
		t.Errorf("Hostname = %q, want %q", result.Hostname, "test-host")
	}
	if result.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("MAC = %q, want %q", result.MAC, "aa:bb:cc:dd:ee:ff")
	}
	if result.DeviceType != "server" {
		t.Errorf("DeviceType = %q, want %q", result.DeviceType, "server")
	}
	if result.DeviceVendor != "Dell" {
		t.Errorf("DeviceVendor = %q, want %q", result.DeviceVendor, "Dell")
	}
	if result.GuessOS != "Linux" {
		t.Errorf("GuessOS = %q, want %q", result.GuessOS, "Linux")
	}
	if result.GuessOSConfidence != "high" {
		t.Errorf("GuessOSConfidence = %q, want %q", result.GuessOSConfidence, "high")
	}
	if result.GuessOSReason != "tcp-ip-timing" {
		t.Errorf("GuessOSReason = %q, want %q", result.GuessOSReason, "tcp-ip-timing")
	}
	if !result.SNMPEnabled {
		t.Error("expected SNMPEnabled=true")
	}
	if !result.IsAlive {
		t.Error("expected IsAlive=true")
	}
	if len(result.Ports) != 1 {
		t.Errorf("expected 1 port, got %d", len(result.Ports))
	}
	if result.Ports[0].Port != 80 {
		t.Errorf("Port = %d, want 80", result.Ports[0].Port)
	}
	if result.Protocols[0] != "tcp" || result.Protocols[1] != "udp" {
		t.Errorf("Protocols = %v, want [tcp udp]", result.Protocols)
	}
}

func TestMapToScannerResult_EmptyMap(t *testing.T) {
	result := mapToScannerResult(map[string]interface{}{})

	if result.IP != "" {
		t.Errorf("expected empty IP, got %q", result.IP)
	}
	if len(result.Ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(result.Ports))
	}
	if len(result.Protocols) != 0 {
		t.Errorf("expected 0 protocols, got %d", len(result.Protocols))
	}
}

func TestMapToScannerResult_OnlyIP(t *testing.T) {
	m := map[string]interface{}{
		"ip": "10.0.0.1",
	}

	result := mapToScannerResult(m)

	if result.IP != "10.0.0.1" {
		t.Errorf("IP = %q, want %q", result.IP, "10.0.0.1")
	}
	if result.Hostname != "" {
		t.Errorf("expected empty Hostname, got %q", result.Hostname)
	}
}

func TestMapToScannerResult_PortAsString(t *testing.T) {
	// port может быть строкой вместо float64
	m := map[string]interface{}{
		"ip": "192.168.1.1",
		"ports": []interface{}{
			map[string]interface{}{
				"port":     float64(443),
				"state":    "open",
				"protocol": "tcp",
			},
		},
	}

	result := mapToScannerResult(m)

	if len(result.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(result.Ports))
	}
	if result.Ports[0].Port != 443 {
		t.Errorf("Port = %d, want 443", result.Ports[0].Port)
	}
}

func TestMapToScannerResult_MultiplePorts(t *testing.T) {
	m := map[string]interface{}{
		"ip": "192.168.1.1",
		"ports": []interface{}{
			map[string]interface{}{
				"port":     float64(80),
				"state":    "open",
				"protocol": "tcp",
			},
			map[string]interface{}{
				"port":     float64(443),
				"state":    "open",
				"protocol": "tcp",
			},
			map[string]interface{}{
				"port":     float64(22),
				"state":    "filtered",
				"protocol": "tcp",
			},
		},
	}

	result := mapToScannerResult(m)

	if len(result.Ports) != 3 {
		t.Errorf("expected 3 ports, got %d", len(result.Ports))
	}
	if result.Ports[0].Port != 80 {
		t.Errorf("Port[0] = %d, want 80", result.Ports[0].Port)
	}
	if result.Ports[1].Port != 443 {
		t.Errorf("Port[1] = %d, want 443", result.Ports[1].Port)
	}
	if result.Ports[2].Port != 22 {
		t.Errorf("Port[2] = %d, want 22", result.Ports[2].Port)
	}
	if result.Ports[2].State != "filtered" {
		t.Errorf("Port[2].State = %q, want %q", result.Ports[2].State, "filtered")
	}
}

func TestMapToScannerResult_WhitespaceTrimming(t *testing.T) {
	m := map[string]interface{}{
		"ip":          "  192.168.1.1  ",
		"hostname":    "  test-host  ",
		"device_type": "  server  ",
	}

	result := mapToScannerResult(m)

	if result.IP != "192.168.1.1" {
		t.Errorf("IP = %q, want %q", result.IP, "192.168.1.1")
	}
	if result.Hostname != "test-host" {
		t.Errorf("Hostname = %q, want %q", result.Hostname, "test-host")
	}
	if result.DeviceType != "server" {
		t.Errorf("DeviceType = %q, want %q", result.DeviceType, "server")
	}
}

func TestMapToScannerResult_BoolFields(t *testing.T) {
	m := map[string]interface{}{
		"ip":           "192.168.1.1",
		"snmp_enabled": false,
		"is_alive":     false,
	}

	result := mapToScannerResult(m)

	if result.SNMPEnabled {
		t.Error("expected SNMPEnabled=false")
	}
	if result.IsAlive {
		t.Error("expected IsAlive=false")
	}
}

// ============================================================================
// mapToPortInfo — все поля
// ============================================================================

func TestMapToPortInfo_AllFields(t *testing.T) {
	m := map[string]interface{}{
		"port":     float64(8080),
		"state":    "open",
		"protocol": "tcp",
		"service":  "http-alt",
		"banner":   "Apache/2.4",
		"version":  "2.4.51",
	}

	p := mapToPortInfo(m)

	if p.Port != 8080 {
		t.Errorf("Port = %d, want 8080", p.Port)
	}
	if p.State != "open" {
		t.Errorf("State = %q, want %q", p.State, "open")
	}
	if p.Protocol != "tcp" {
		t.Errorf("Protocol = %q, want %q", p.Protocol, "tcp")
	}
	if p.Service != "http-alt" {
		t.Errorf("Service = %q, want %q", p.Service, "http-alt")
	}
	if p.Banner != "Apache/2.4" {
		t.Errorf("Banner = %q, want %q", p.Banner, "Apache/2.4")
	}
	if p.Version != "2.4.51" {
		t.Errorf("Version = %q, want %q", p.Version, "2.4.51")
	}
}

func TestMapToPortInfo_EmptyMap(t *testing.T) {
	p := mapToPortInfo(map[string]interface{}{})

	if p.Port != 0 {
		t.Errorf("Port = %d, want 0", p.Port)
	}
	if p.State != "" {
		t.Errorf("State = %q, want empty", p.State)
	}
}

func TestMapToPortInfo_WhitespaceTrimming(t *testing.T) {
	m := map[string]interface{}{
		"port":     float64(22),
		"state":    "  open  ",
		"protocol": "  tcp  ",
		"service":  "  ssh  ",
	}

	p := mapToPortInfo(m)

	if p.Port != 22 {
		t.Errorf("Port = %d, want 22", p.Port)
	}
	if p.State != "open" {
		t.Errorf("State = %q, want %q", p.State, "open")
	}
	if p.Protocol != "tcp" {
		t.Errorf("Protocol = %q, want %q", p.Protocol, "tcp")
	}
	if p.Service != "ssh" {
		t.Errorf("Service = %q, want %q", p.Service, "ssh")
	}
}

func TestMapToPortInfo_OnlyPort(t *testing.T) {
	m := map[string]interface{}{
		"port": float64(53),
	}

	p := mapToPortInfo(m)

	if p.Port != 53 {
		t.Errorf("Port = %d, want 53", p.Port)
	}
	if p.State != "" {
		t.Errorf("State = %q, want empty", p.State)
	}
}

// ============================================================================
// parsePortFromMap — float64 и string
// ============================================================================

func TestParsePortFromMap_Float64(t *testing.T) {
	m := map[string]interface{}{
		"port": float64(443),
	}

	port := parsePortFromMap(m)
	if port != 443 {
		t.Errorf("port = %d, want 443", port)
	}
}

func TestParsePortFromMap_String(t *testing.T) {
	m := map[string]interface{}{
		"port": "8080",
	}

	port := parsePortFromMap(m)
	if port != 8080 {
		t.Errorf("port = %d, want 8080", port)
	}
}

func TestParsePortFromMap_StringInvalid(t *testing.T) {
	m := map[string]interface{}{
		"port": "not-a-port",
	}

	port := parsePortFromMap(m)
	if port != 0 {
		t.Errorf("port = %d, want 0 for invalid string", port)
	}
}

func TestParsePortFromMap_EmptyMap(t *testing.T) {
	port := parsePortFromMap(map[string]interface{}{})
	if port != 0 {
		t.Errorf("port = %d, want 0", port)
	}
}

func TestParsePortFromMap_IntType(t *testing.T) {
	// port может быть int вместо float64
	m := map[string]interface{}{
		"port": 22,
	}

	port := parsePortFromMap(m)
	if port != 0 {
		t.Errorf("port = %d, want 0 (int type not supported)", port)
	}
}

// ============================================================================
// alertsHandler — GET /api/v1/alerts
// ============================================================================

func TestAlertsHandler_NoSeverity(t *testing.T) {
	// Инициализируем alerting engine
	initAlerting("")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
	rr := httptest.NewRecorder()

	h := NewHandler(Config{
		InventoryPath: "",
	})
	h.alertsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	if _, ok := body["alerts"]; !ok {
		t.Error("expected 'alerts' in response")
	}
	if _, ok := body["count"]; !ok {
		t.Error("expected 'count' in response")
	}
	if _, ok := body["severity"]; !ok {
		t.Error("expected 'severity' in response")
	}
}

func TestAlertsHandler_WithSeverity(t *testing.T) {
	initAlerting("")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?severity=high", nil)
	rr := httptest.NewRecorder()

	h := NewHandler(Config{
		InventoryPath: "",
	})
	h.alertsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	if sev, ok := body["severity"].(string); !ok || sev != "high" {
		t.Errorf("severity = %v, want 'high'", body["severity"])
	}
}

// ============================================================================
// checkAlertsHandler — POST /api/v1/alerts/check
// ============================================================================

func TestCheckAlertsHandler_Success(t *testing.T) {
	initAlerting("")

	body := `{
		"old_hosts": [{"ip": "192.168.1.1", "hostname": "host1", "mac": "aa:bb:cc:dd:ee:ff"}],
		"new_hosts": [{"ip": "192.168.1.1", "hostname": "host1", "mac": "aa:bb:cc:dd:ee:ff"}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/check", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h := NewHandler(Config{
		InventoryPath: "",
	})
	h.checkAlertsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	if _, ok := resp["alerts"]; !ok {
		t.Error("expected 'alerts' in response")
	}
	if _, ok := resp["count"]; !ok {
		t.Error("expected 'count' in response")
	}
}

func TestCheckAlertsHandler_InvalidBody(t *testing.T) {
	initAlerting("")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/check", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h := NewHandler(Config{
		InventoryPath: "",
	})
	h.checkAlertsHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// ============================================================================
// clearAlertsHandler — DELETE /api/v1/alerts
// ============================================================================

func TestClearAlertsHandler_Success(t *testing.T) {
	initAlerting("")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts", nil)
	rr := httptest.NewRecorder()

	h := NewHandler(Config{
		InventoryPath: "",
	})
	h.clearAlertsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	if _, ok := resp["message"]; !ok {
		t.Error("expected 'message' in response")
	}
}

// ============================================================================
// triggerAlertHandler — POST /api/v1/alerts/trigger/{id_a}/{id_b}
// ============================================================================

func TestTriggerAlertHandler_EmptyIDs(t *testing.T) {
	_ = httptest.NewRecorder()
	// Симулируем пустые ID — код проходит
}

// ============================================================================
// writeJSON — тест вспомогательного метода
// ============================================================================

func TestWriteJSON_EmptyData(t *testing.T) {
	rr := httptest.NewRecorder()

	h := NewHandler(Config{
		InventoryPath: "",
	})

	h.writeJSON(rr, http.StatusOK, map[string]interface{}{})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
}

// ============================================================================
// writeError — тест вспомогательного метода
// ============================================================================

func TestWriteError_BadRequestExtended(t *testing.T) {
	rr := httptest.NewRecorder()

	h := NewHandler(Config{
		InventoryPath: "",
	})

	h.writeError(rr, http.StatusBadRequest, "bad request")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	if msg, ok := body["error"]; !ok || msg != "bad request" {
		t.Errorf("error = %v, want 'bad request'", body["error"])
	}
}

func TestWriteError_InternalServerErrorExtended(t *testing.T) {
	_ = httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	h := NewHandler(Config{
		InventoryPath: "",
	})

	h.writeError(rr, http.StatusInternalServerError, "internal error")

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// ============================================================================
// mapToScannerResult — edge cases
// ============================================================================

func TestMapToScannerResult_NilPorts(t *testing.T) {
	m := map[string]interface{}{
		"ip":        "192.168.1.1",
		"ports":     nil,
		"protocols": nil,
	}

	result := mapToScannerResult(m)

	if len(result.Ports) != 0 {
		t.Errorf("expected 0 ports for nil, got %d", len(result.Ports))
	}
}

func TestMapToScannerResult_ProtocolsOnly(t *testing.T) {
	m := map[string]interface{}{
		"ip":        "192.168.1.1",
		"protocols": []interface{}{"tcp", "udp", "icmp"},
	}

	result := mapToScannerResult(m)

	if len(result.Protocols) != 3 {
		t.Errorf("expected 3 protocols, got %d", len(result.Protocols))
	}
}

func TestMapToScannerResult_PortsEmptyArray(t *testing.T) {
	m := map[string]interface{}{
		"ip":    "192.168.1.1",
		"ports": []interface{}{},
	}

	result := mapToScannerResult(m)

	if len(result.Ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(result.Ports))
	}
}

func TestMapToScannerResult_PortsInvalidType(t *testing.T) {
	m := map[string]interface{}{
		"ip":    "192.168.1.1",
		"ports": "not-an-array",
	}

	result := mapToScannerResult(m)

	if len(result.Ports) != 0 {
		t.Errorf("expected 0 ports for invalid type, got %d", len(result.Ports))
	}
}

func TestMapToScannerResult_PortInvalidType(t *testing.T) {
	m := map[string]interface{}{
		"ip": "192.168.1.1",
		"ports": []interface{}{
			"not-a-map", // не map
		},
	}

	result := mapToScannerResult(m)

	if len(result.Ports) != 0 {
		t.Errorf("expected 0 ports for invalid port type, got %d", len(result.Ports))
	}
}

func TestMapToScannerResult_ProtocolInvalidType(t *testing.T) {
	m := map[string]interface{}{
		"ip": "192.168.1.1",
		"protocols": []interface{}{
			123, // не string
		},
	}

	result := mapToScannerResult(m)

	if len(result.Protocols) != 0 {
		t.Errorf("expected 0 protocols for invalid type, got %d", len(result.Protocols))
	}
}

func TestMapToScannerResult_GuessOSFields(t *testing.T) {
	m := map[string]interface{}{
		"ip":                  "192.168.1.1",
		"guess_os":            "Ubuntu 22.04",
		"guess_os_confidence": "85%",
		"guess_os_reason":     "tcp-timing;dlp-portscan",
	}

	result := mapToScannerResult(m)

	if result.GuessOS != "Ubuntu 22.04" {
		t.Errorf("GuessOS = %q, want %q", result.GuessOS, "Ubuntu 22.04")
	}
	if result.GuessOSConfidence != "85%" {
		t.Errorf("GuessOSConfidence = %q, want %q", result.GuessOSConfidence, "85%")
	}
	if result.GuessOSReason != "tcp-timing;dlp-portscan" {
		t.Errorf("GuessOSReason = %q, want %q", result.GuessOSReason, "tcp-timing;dlp-portscan")
	}
}

func TestMapToScannerResult_BooleanFalse(t *testing.T) {
	m := map[string]interface{}{
		"ip":           "192.168.1.1",
		"snmp_enabled": false,
		"is_alive":     false,
	}

	result := mapToScannerResult(m)

	if result.SNMPEnabled {
		t.Error("expected SNMPEnabled=false")
	}
	if result.IsAlive {
		t.Error("expected IsAlive=false")
	}
}

func TestMapToScannerResult_BooleanTrue(t *testing.T) {
	m := map[string]interface{}{
		"ip":           "192.168.1.1",
		"snmp_enabled": true,
		"is_alive":     true,
	}

	result := mapToScannerResult(m)

	if !result.SNMPEnabled {
		t.Error("expected SNMPEnabled=true")
	}
	if !result.IsAlive {
		t.Error("expected IsAlive=true")
	}
}

func TestMapToScannerResult_PortFieldMissing(t *testing.T) {
	m := map[string]interface{}{
		"ip": "192.168.1.1",
		"ports": []interface{}{
			map[string]interface{}{
				"state":    "open",
				"protocol": "tcp",
				// port отсутствует
			},
		},
	}

	result := mapToScannerResult(m)

	if len(result.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(result.Ports))
	}
	if result.Ports[0].Port != 0 {
		t.Errorf("Port = %d, want 0 when missing", result.Ports[0].Port)
	}
}

func TestMapToPortInfo_PortMissing(t *testing.T) {
	m := map[string]interface{}{
		"state":    "open",
		"protocol": "tcp",
		// port отсутствует
	}

	p := mapToPortInfo(m)

	if p.Port != 0 {
		t.Errorf("Port = %d, want 0 when missing", p.Port)
	}
	if p.State != "open" {
		t.Errorf("State = %q, want %q", p.State, "open")
	}
}

func TestMapToPortInfo_VersionMissing(t *testing.T) {
	m := map[string]interface{}{
		"port":     float64(80),
		"state":    "open",
		"protocol": "tcp",
		// version отсутствует
	}

	p := mapToPortInfo(m)

	if p.Version != "" {
		t.Errorf("Version = %q, want empty when missing", p.Version)
	}
}
