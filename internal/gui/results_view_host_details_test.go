package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

// --- results_view.go: host details & selection tests ---

func TestSelectHostForDetails_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать для инициализированного приложения
	a.selectHostForDetails(scanner.Result{IP: "192.168.1.1"})
}

func TestSelectHostForDetails_EmptyIP(t *testing.T) {
	a := &App{}
	// Не должен паниковать, пустой IP игнорируется
	a.selectHostForDetails(scanner.Result{})
}

func TestSelectHostForDetails_WithIP(t *testing.T) {
	a := &App{}
	a.selectHostForDetails(scanner.Result{IP: "192.168.1.1", Hostname: "h1"})
	if a.selectedHostIP != "192.168.1.1" {
		t.Errorf("expected selectedHostIP='192.168.1.1', got %q", a.selectedHostIP)
	}
}

func TestSelectedHostFromData_EmptyData(t *testing.T) {
	a := &App{}
	_, ok := a.selectedHostFromData(nil)
	if ok {
		t.Error("expected ok=false for empty data")
	}
}

func TestSelectedHostFromData_NoSelected(t *testing.T) {
	a := &App{}
	data := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "h1"},
		{IP: "192.168.1.2", Hostname: "h2"},
	}
	r, ok := a.selectedHostFromData(data)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if r.IP != "192.168.1.1" {
		t.Errorf("expected first host selected, got %q", r.IP)
	}
	if a.selectedHostIP != "192.168.1.1" {
		t.Errorf("expected selectedHostIP set to first host, got %q", a.selectedHostIP)
	}
}

func TestSelectedHostFromData_WithSelected(t *testing.T) {
	a := &App{}
	a.selectedHostIP = "192.168.1.2"
	data := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "h1"},
		{IP: "192.168.1.2", Hostname: "h2"},
	}
	r, ok := a.selectedHostFromData(data)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if r.Hostname != "h2" {
		t.Errorf("expected h2 selected, got %q", r.Hostname)
	}
}

func TestSelectedHostFromData_SelectedNotFound(t *testing.T) {
	a := &App{}
	a.selectedHostIP = "10.0.0.99"
	data := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "h1"},
	}
	r, ok := a.selectedHostFromData(data)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if r.IP != "192.168.1.1" {
		t.Errorf("expected fallback to first host, got %q", r.IP)
	}
}

func TestPrimeHostDetailsCache_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.primeHostDetailsCache(scanner.Result{IP: "192.168.1.1"})
}

func TestPrimeHostDetailsCache_CreatesCache(t *testing.T) {
	a := &App{}
	a.primeHostDetailsCache(scanner.Result{IP: "192.168.1.1", Hostname: "h1"})
	a.hostDetailsCacheMu.RLock()
	count := len(a.hostDetailsCache)
	a.hostDetailsCacheMu.RUnlock()
	if count != 1 {
		t.Errorf("expected 1 cache entry, got %d", count)
	}
}

func TestPrefetchHostDetailsNearby_EmptyData(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.prefetchHostDetailsNearby(nil, "")
}

func TestPrefetchHostDetailsNearby_NoSelected(t *testing.T) {
	a := &App{}
	data := []scanner.Result{{IP: "192.168.1.1"}}
	// Не должен паниковать
	a.prefetchHostDetailsNearby(data, "")
}

func TestPrefetchHostDetailsNearby_WithSelected(t *testing.T) {
	a := &App{}
	data := []scanner.Result{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.3"},
	}
	a.prefetchHostDetailsNearby(data, "192.168.1.2")
	// Не должен паниковать (асинхронный prefetch)
}

func TestBuildPortChips_NoOpenPorts(t *testing.T) {
	a := &App{}
	r := scanner.Result{IP: "192.168.1.1", Ports: []scanner.PortInfo{{State: "closed"}}}
	result := a.buildPortChips(r)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildPortChips_WithOpenPorts(t *testing.T) {
	a := &App{}
	a.maxPortChips = 10
	r := scanner.Result{
		IP: "192.168.1.1",
		Ports: []scanner.PortInfo{
			{Port: 22, Protocol: "tcp", State: "open", Service: "ssh"},
			{Port: 80, Protocol: "tcp", State: "open", Service: "http"},
		},
	}
	result := a.buildPortChips(r)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildPortChips_WithBanners(t *testing.T) {
	a := &App{}
	a.showRawBanners = true
	r := scanner.Result{
		IP: "192.168.1.1",
		Ports: []scanner.PortInfo{
			{Port: 80, Protocol: "tcp", State: "open", Service: "http", Banner: "Apache/2.4"},
		},
	}
	result := a.buildPortChips(r)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildPortChips_OverLimit(t *testing.T) {
	a := &App{}
	a.maxPortChips = 2
	ports := make([]scanner.PortInfo, 0, 5)
	for i := 0; i < 5; i++ {
		ports = append(ports, scanner.PortInfo{Port: 1000 + i, Protocol: "tcp", State: "open"})
	}
	r := scanner.Result{IP: "192.168.1.1", Ports: ports}
	result := a.buildPortChips(r)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildHostQuickActions_WithRows(t *testing.T) {
	a := &App{}
	r := scanner.Result{IP: "192.168.1.1", Hostname: "h1"}
	result := a.buildHostQuickActions(r, 1)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildHostQuickActions_ZeroCols(t *testing.T) {
	a := &App{}
	r := scanner.Result{IP: "192.168.1.1"}
	result := a.buildHostQuickActions(r, 0)
	if result == nil {
		t.Fatal("expected non-nil result for cols<=0")
	}
}

func TestBuildHostDetailsDrawer_EmptyData(t *testing.T) {
	a := &App{}
	result := a.buildHostDetailsDrawer(nil)
	if result == nil {
		t.Fatal("expected non-nil result for empty data")
	}
}

func TestBuildHostDetailsDrawer_WithData(t *testing.T) {
	a := &App{}
	data := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "h1", DeviceType: "Router"},
	}
	result := a.buildHostDetailsDrawer(data)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildHostDetailsDrawer_CompactProfile(t *testing.T) {
	a := &App{}
	a.layoutProfile = "compact"
	data := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "h1"},
	}
	result := a.buildHostDetailsDrawer(data)
	if result == nil {
		t.Fatal("expected non-nil result for compact profile")
	}
}

func TestBuildTableView_WithData(t *testing.T) {
	a := &App{}
	data := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "h1", MAC: "AA:BB", DeviceType: "Router"},
	}
	result := a.buildTableView(data)
	if result == nil {
		t.Fatal("expected non-nil table")
	}
}

func TestBuildTableView_EmptyData(t *testing.T) {
	a := &App{}
	result := a.buildTableView(nil)
	if result == nil {
		t.Fatal("expected non-nil table for empty data")
	}
}

func TestBuildCardsView_EmptyData(t *testing.T) {
	a := &App{}
	result := a.buildCardsView(nil)
	if result == nil {
		t.Fatal("expected non-nil cards view for empty data")
	}
}

func TestBuildCardsView_NoLimit(t *testing.T) {
	a := &App{}
	data := []scanner.Result{{IP: "192.168.1.1"}}
	result := a.buildCardsView(data)
	if result == nil {
		t.Fatal("expected non-nil cards view")
	}
}
