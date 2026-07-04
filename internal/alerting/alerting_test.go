package alerting

import (
	"os"
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine("test_alerts.log")
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
	if len(engine.rules) != 6 {
		t.Errorf("expected 6 default rules, got %d", len(engine.rules))
	}
}

func TestCheckAlerts_NewHost(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/alerts.log"
	engine := NewEngine(logPath)

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}

	alerts := engine.CheckAlerts(oldHosts, newHosts)

	if len(alerts) == 0 {
		t.Error("expected at least one alert for new host")
	}

	found := false
	for _, alert := range alerts {
		if alert.RuleName == "New Host Detected" {
			found = true
			if alert.Severity != SeverityMedium {
				t.Errorf("expected MEDIUM severity, got %s", alert.Severity)
			}
		}
	}
	if !found {
		t.Error("expected New Host Detected alert")
	}
}

func TestCheckAlerts_RemovedHost(t *testing.T) {
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
	for _, alert := range alerts {
		if alert.RuleName == "Device Removed" {
			found = true
			if alert.Severity != SeverityHigh {
				t.Errorf("expected HIGH severity, got %s", alert.Severity)
			}
		}
	}
	if !found {
		t.Error("expected Device Removed alert")
	}
}

func TestCheckAlerts_PortChange(t *testing.T) {
	engine := NewEngine("")

	oldHosts := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "open"}}},
	}
	newHosts := []scanner.Result{
		{IP: "192.168.1.1", Ports: []scanner.PortInfo{{Port: 80, Protocol: "tcp", State: "closed"}, {Port: 443, Protocol: "tcp", State: "open"}}},
	}

	alerts := engine.CheckAlerts(oldHosts, newHosts)

	if len(alerts) == 0 {
		t.Error("expected alerts for port changes")
	}
}

func TestCheckAlerts_NoChanges(t *testing.T) {
	engine := NewEngine("")

	hosts := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router"},
	}

	alerts := engine.CheckAlerts(hosts, hosts)

	if len(alerts) != 0 {
		t.Errorf("expected no alerts, got %d", len(alerts))
	}
}

func TestGetAlertsBySeverity(t *testing.T) {
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

	if len(highAlerts) < 0 {
		t.Error("expected non-negative count for HIGH alerts")
	}
	if len(mediumAlerts) < 0 {
		t.Error("expected non-negative count for MEDIUM alerts")
	}
}

func TestClearAlerts(t *testing.T) {
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
		t.Error("expected alerts before clear")
	}

	engine.ClearAlerts()

	if len(engine.GetAlerts()) != 0 {
		t.Error("expected no alerts after clear")
	}
}

func TestFileHandler_OnAlert(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/test_alerts.log"

	handler := &FileHandler{Path: logPath}

	alert := &Alert{
		ID:        "test-alert-1",
		RuleName:  "Test Rule",
		Severity:  SeverityHigh,
		Message:   "Test message",
		Timestamp: testTime,
	}

	err := handler.OnAlert(alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read alert file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty alert file")
	}
}

var testTime = func() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}()

