package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

// --- security_view.go tests ---

func TestBuildSecurityDashboardView_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.buildSecurityDashboardView(nil)
}

func TestBuildSecurityDashboardView_EmptyData(t *testing.T) {
	a := &App{}
	result := a.buildSecurityDashboardView(nil)
	if result == nil {
		t.Fatal("expected non-nil result for empty data")
	}
}

func TestBuildSecurityDashboardView_WithData(t *testing.T) {
	a := &App{}
	data := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Hostname:   "router",
			DeviceType: "Router",
			Ports: []scanner.PortInfo{
				{Port: 22, Protocol: "tcp", State: "open"},
				{Port: 80, Protocol: "tcp", State: "open"},
			},
		},
	}
	result := a.buildSecurityDashboardView(data)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildSecurityDashboardView_MultipleDevices(t *testing.T) {
	a := &App{}
	data := []scanner.Result{
		{
			IP: "192.168.1.1",
			Ports: []scanner.PortInfo{
				{Port: 22, Protocol: "tcp", State: "open"},
			},
		},
		{
			IP: "192.168.1.2",
			Ports: []scanner.PortInfo{
				{Port: 443, Protocol: "tcp", State: "open"},
			},
		},
	}
	result := a.buildSecurityDashboardView(data)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildSecurityFindingsTable_NilApp(t *testing.T) {
	a := &App{}
	result := a.buildSecurityFindingsTable(nil, nil)
	if result == nil {
		t.Fatal("expected non-nil result for empty findings")
	}
}

func TestExportSecurityDashboardReport_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.exportSecurityDashboardReport(nil, nil)
}

func TestExportSecurityDashboardReport_NilWindow(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.exportSecurityDashboardReport(nil, nil)
}
