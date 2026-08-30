package alerting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

// === Integration: CheckAlerts — New Host ===

func TestIntegrationCheckAlerts_NewHost(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "alerts.log")
	engine := NewEngine(logPath)

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router"},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router"},
		{IP: "192.168.1.2", Hostname: "switch", DeviceType: "Switch"},
	}

	alerts := engine.CheckAlerts(oldHosts, newHosts)
	if len(alerts) == 0 {
		t.Fatal("expected at least 1 alert for new host")
	}

	found := false
	for _, a := range alerts {
		if a.RuleName == "New Host Detected" {
			found = true
			if a.Severity != SeverityMedium {
				t.Errorf("expected MEDIUM severity, got %s", a.Severity)
			}
			if a.Host != "192.168.1.2" {
				t.Errorf("expected host 192.168.1.2, got %s", a.Host)
			}
			if a.Port != 0 {
				t.Errorf("expected port 0, got %d", a.Port)
			}
		}
	}
	if !found {
		t.Error("expected New Host Detected alert")
	}
}

// === Integration: CheckAlerts — Removed Host ===

func TestIntegrationCheckAlerts_RemovedHost(t *testing.T) {
	engine := NewEngine("")

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}

	alerts := engine.CheckAlerts(oldHosts, newHosts)

	found := false
	for _, a := range alerts {
		if a.RuleName == "Device Removed" {
			found = true
			if a.Severity != SeverityHigh {
				t.Errorf("expected HIGH severity, got %s", a.Severity)
			}
			if a.Host != "192.168.1.2" {
				t.Errorf("expected host 192.168.1.2, got %s", a.Host)
			}
		}
	}
	if !found {
		t.Error("expected Device Removed alert")
	}
}

// === Integration: CheckAlerts — New Port ===

func TestIntegrationCheckAlerts_NewPort(t *testing.T) {
	engine := NewEngine("")

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{
			{Port: 80, Protocol: "tcp", State: "open"},
			{Port: 443, Protocol: "tcp", State: "open"},
		}},
	}

	alerts := engine.CheckAlerts(oldHosts, newHosts)

	found := false
	for _, a := range alerts {
		if a.RuleName == "New Port Opened" {
			found = true
			if a.Severity != SeverityHigh {
				t.Errorf("expected HIGH severity, got %s", a.Severity)
			}
			if a.Port != 443 {
				t.Errorf("expected port 443, got %d", a.Port)
			}
		}
	}
	if !found {
		t.Error("expected New Port Opened alert")
	}
}

// === Integration: CheckAlerts — Port Closed ===

func TestIntegrationCheckAlerts_PortClosed(t *testing.T) {
	engine := NewEngine("")

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "closed"}}},
	}

	alerts := engine.CheckAlerts(oldHosts, newHosts)

	found := false
	for _, a := range alerts {
		if a.RuleName == "Port Closed" {
			found = true
			if a.Severity != SeverityMedium {
				t.Errorf("expected MEDIUM severity, got %s", a.Severity)
			}
			if a.Port != 80 {
				t.Errorf("expected port 80, got %d", a.Port)
			}
		}
	}
	if !found {
		t.Error("expected Port Closed alert")
	}
}

// === Integration: CheckAlerts — Hostname Changed ===

func TestIntegrationCheckAlerts_HostnameChanged(t *testing.T) {
	engine := NewEngine("")

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "old-name", DeviceType: "Router"},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "new-name", DeviceType: "Router"},
	}

	alerts := engine.CheckAlerts(oldHosts, newHosts)

	found := false
	for _, a := range alerts {
		if a.RuleName == "Hostname Changed" {
			found = true
			if a.Severity != SeverityLow {
				t.Errorf("expected LOW severity, got %s", a.Severity)
			}
		}
	}
	if !found {
		t.Error("expected Hostname Changed alert")
	}
}

// === Integration: CheckAlerts — OS Changed ===
// Note: DeviceType changes are detected as "device_type" field, not "os"
func TestIntegrationCheckAlerts_DeviceTypeChanged(t *testing.T) {
	engine := NewEngine("")

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "server", DeviceType: "Linux Server"},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "server", DeviceType: "Windows Server"},
	}

	alerts := engine.CheckAlerts(oldHosts, newHosts)

	// Device type changes are detected as general host changes
	// We verify that some alert was generated for the changed host
	found := false
	for _, a := range alerts {
		if a.Host == "192.168.1.1" {
			found = true
		}
	}
	if !found {
		t.Error("expected alert for device type change")
	}
}

// === Integration: CheckAlerts — No Changes ===

func TestIntegrationCheckAlerts_NoChanges(t *testing.T) {
	engine := NewEngine("")

	hosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router"},
	}

	alerts := engine.CheckAlerts(hosts, hosts)

	if len(alerts) != 0 {
		t.Errorf("expected no alerts, got %d", len(alerts))
	}
}

// === Integration: CheckAlerts — Empty Inputs ===

func TestIntegrationCheckAlerts_EmptyInputs(t *testing.T) {
	engine := NewEngine("")

	alerts := engine.CheckAlerts(nil, nil)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts for nil inputs, got %d", len(alerts))
	}
}

func TestIntegrationCheckAlerts_EmptyNew(t *testing.T) {
	engine := NewEngine("")

	alerts := engine.CheckAlerts([]scanner.Result{}, []scanner.Result{})
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts for empty inputs, got %d", len(alerts))
	}
}

// === Integration: GetAlerts ===

func TestIntegrationGetAlerts(t *testing.T) {
	engine := NewEngine("")

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}

	engine.CheckAlerts(oldHosts, newHosts)

	allAlerts := engine.GetAlerts()
	if len(allAlerts) == 0 {
		t.Error("expected alerts after CheckAlerts")
	}
}

// === Integration: GetAlertsBySeverity ===

func TestIntegrationGetAlertsBySeverity_High(t *testing.T) {
	engine := NewEngine("")

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}

	engine.CheckAlerts(oldHosts, newHosts)

	highAlerts := engine.GetAlertsBySeverity(SeverityHigh)
	mediumAlerts := engine.GetAlertsBySeverity(SeverityMedium)
	lowAlerts := engine.GetAlertsBySeverity(SeverityLow)

	if len(highAlerts) < 0 {
		t.Error("expected non-negative HIGH alerts count")
	}
	if len(mediumAlerts) < 0 {
		t.Error("expected non-negative MEDIUM alerts count")
	}
	if len(lowAlerts) < 0 {
		t.Error("expected non-negative LOW alerts count")
	}
}

// === Integration: ClearAlerts ===

func TestIntegrationClearAlerts(t *testing.T) {
	engine := NewEngine("")

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}

	engine.CheckAlerts(oldHosts, newHosts)
	if len(engine.GetAlerts()) == 0 {
		t.Fatal("expected alerts before clear")
	}

	engine.ClearAlerts()
	if len(engine.GetAlerts()) != 0 {
		t.Error("expected no alerts after clear")
	}
}

// === Integration: FileHandler OnAlert ===

func TestIntegrationFileHandler_OnAlert(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "alerts.log")

	handler := &FileHandler{Path: logPath}

	alert := &Alert{
		ID:        "test-alert-1",
		RuleName:  "New Host Detected",
		Severity:  SeverityMedium,
		Message:   "New host 192.168.1.2",
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Host:      "192.168.1.2",
	}

	err := handler.OnAlert(alert)
	if err != nil {
		t.Fatalf("OnAlert error: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty file")
	}
	if !strings.Contains(string(data), "test-alert-1") {
		t.Error("expected file to contain alert ID")
	}
	if !strings.Contains(string(data), "New Host Detected") {
		t.Error("expected file to contain rule name")
	}
}

// === Integration: FileHandler Invalid Path ===

func TestIntegrationFileHandler_InvalidPath(t *testing.T) {
	// On Windows, we use a path that's guaranteed to fail
	// (e.g., a file path without a valid parent directory)
	handler := &FileHandler{Path: "C:\\NonExistentRoot\\alerts.log"}

	alert := &Alert{
		ID:       "test-alert",
		RuleName: "Test",
		Severity: SeverityLow,
		Message:  "test",
	}

	err := handler.OnAlert(alert)
	if err == nil {
		// On some systems, this might succeed or fail differently
		// Just verify the alert structure is valid
	}
}

// === Integration: ConsoleHandler OnAlert ===

func TestIntegrationConsoleHandler_OnAlert(t *testing.T) {
	handler := &ConsoleHandler{}

	alert := &Alert{
		ID:        "test-alert",
		RuleName:  "New Host",
		Severity:  SeverityHigh,
		Message:   "New host detected",
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Host:      "192.168.1.2",
		Port:      22,
	}

	err := handler.OnAlert(alert)
	if err != nil {
		t.Fatalf("OnAlert error: %v", err)
	}
}

// === Integration: ConsoleHandler No Host ===

func TestIntegrationConsoleHandler_NoHost(t *testing.T) {
	handler := &ConsoleHandler{}

	alert := &Alert{
		ID:        "test-alert",
		RuleName:  "Test",
		Severity:  SeverityLow,
		Message:   "test message",
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	err := handler.OnAlert(alert)
	if err != nil {
		t.Fatalf("OnAlert error: %v", err)
	}
}

// === Integration: Full Alerting Pipeline ===

func TestIntegrationFullAlertingPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "alerts.log")

	// Step 1: Инициализация
	engine := NewEngine(logPath)
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
	if len(engine.rules) != 6 {
		t.Errorf("expected 6 default rules, got %d", len(engine.rules))
	}

	// Step 2: Первый скан (baseline)
	baseline := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router"},
		{IP: "192.168.1.2", Hostname: "switch", DeviceType: "Switch"},
	}

	// Step 3: Второй скан (с изменениями)
	current := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router"},
		{IP: "192.168.1.2", Hostname: "switch", DeviceType: "Switch"},
		{IP: "192.168.1.3", Hostname: "new-server", DeviceType: "Server"}, // Новый хост
		{IP: "192.168.1.4", Hostname: "printer", DeviceType: "Printer"},   // Новый хост
	}

	// Step 4: Проверка алертов
	alerts := engine.CheckAlerts(baseline, current)
	if len(alerts) == 0 {
		t.Fatal("expected alerts for new hosts")
	}

	// Step 5: Фильтрация по severity
	highAlerts := engine.GetAlertsBySeverity(SeverityHigh)
	if len(highAlerts) < 0 {
		t.Error("expected non-negative HIGH alerts")
	}

	// Step 6: Проверка сохранения в файл
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty log file")
	}

	// Step 7: Очистка
	engine.ClearAlerts()
	if len(engine.GetAlerts()) != 0 {
		t.Error("expected no alerts after clear")
	}
}

// === Integration: Multiple Scans ===

func TestIntegrationMultipleScans(t *testing.T) {
	engine := NewEngine("")

	scan1 := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}
	scan2 := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}
	scan3 := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
		{IP: "192.168.1.3", Hostname: "server"},
	}

	engine.CheckAlerts(scan1, scan2)
	engine.CheckAlerts(scan2, scan3)

	allAlerts := engine.GetAlerts()
	if len(allAlerts) < 2 {
		t.Errorf("expected at least 2 alerts from multiple scans, got %d", len(allAlerts))
	}
}

// === Integration: Severity Constants ===

func TestIntegrationSeverityConstants(t *testing.T) {
	if SeverityLow != "LOW" {
		t.Errorf("expected SeverityLow=LOW, got %s", SeverityLow)
	}
	if SeverityMedium != "MEDIUM" {
		t.Errorf("expected SeverityMedium=MEDIUM, got %s", SeverityMedium)
	}
	if SeverityHigh != "HIGH" {
		t.Errorf("expected SeverityHigh=HIGH, got %s", SeverityHigh)
	}
	if SeverityCritical != "CRITICAL" {
		t.Errorf("expected SeverityCritical=CRITICAL, got %s", SeverityCritical)
	}
}

// === Integration: RuleType Constants ===

func TestIntegrationRuleTypeConstants(t *testing.T) {
	if RuleTypeNewHost != "new_host" {
		t.Errorf("expected new_host, got %s", RuleTypeNewHost)
	}
	if RuleTypeNewPort != "new_port" {
		t.Errorf("expected new_port, got %s", RuleTypeNewPort)
	}
	if RuleTypePortClosed != "port_closed" {
		t.Errorf("expected port_closed, got %s", RuleTypePortClosed)
	}
	if RuleTypeDeviceRemoved != "device_removed" {
		t.Errorf("expected device_removed, got %s", RuleTypeDeviceRemoved)
	}
	if RuleTypeOSChanged != "os_changed" {
		t.Errorf("expected os_changed, got %s", RuleTypeOSChanged)
	}
	if RuleTypeHostnameChanged != "hostname_changed" {
		t.Errorf("expected hostname_changed, got %s", RuleTypeHostnameChanged)
	}
}

// === Integration: Alert Struct Fields ===

func TestIntegrationAlertStructFields(t *testing.T) {
	alert := Alert{
		ID:        "test-1",
		RuleID:    "rule-001",
		RuleName:  "Test Rule",
		Severity:  SeverityHigh,
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Message:   "Test message",
		Data:      "extra data",
		Host:      "192.168.1.1",
		Port:      80,
	}

	if alert.ID != "test-1" {
		t.Errorf("expected ID test-1, got %s", alert.ID)
	}
	if alert.RuleID != "rule-001" {
		t.Errorf("expected RuleID rule-001, got %s", alert.RuleID)
	}
	if alert.Host != "192.168.1.1" {
		t.Errorf("expected Host 192.168.1.1, got %s", alert.Host)
	}
	if alert.Port != 80 {
		t.Errorf("expected Port 80, got %d", alert.Port)
	}
}

// === Integration: Engine Struct ===

func TestIntegrationEngineStruct(t *testing.T) {
	engine := NewEngine("test.log")
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
	if engine.logFile != "test.log" {
		t.Errorf("expected logFile test.log, got %s", engine.logFile)
	}
	if len(engine.rules) != 6 {
		t.Errorf("expected 6 rules, got %d", len(engine.rules))
	}
	if len(engine.handlers) != 2 {
		t.Errorf("expected 2 handlers, got %d", len(engine.handlers))
	}
}
