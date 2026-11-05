package controller

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

// TestResultsController_LoadSettings_nilApp проверяет, что LoadSettings
// безопасно завершается при nil app.
func TestResultsController_LoadSettings_nilApp(t *testing.T) {
	ui := &ResultsUI{}
	ctrl := NewResultsController(nil, ui)
	ctrl.LoadSettings() // не должен паниковать
}

// TestResultsController_LoadSettings_nilUI проверяет, что LoadSettings
// безопасно завершается при nil ui.
func TestResultsController_LoadSettings_nilUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	ctrl := NewResultsController(a, nil)
	ctrl.LoadSettings() // не должен паниковать
}

// TestResultsController_LoadSettings_full загружает все настройки из Preferences.
func TestResultsController_LoadSettings_full(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()
	p.SetString("scan.results_view_mode", "Карточки")
	p.SetString("scan.results_submode", "Безопасность")
	p.SetString("scan.results_sort_mode", "HostName")
	p.SetString("scan.results_chip_limit", "50")
	p.SetString("scan.results_show_raw_banners", "true")
	p.SetString("scan.results_filter_query", "test")
	p.SetString("scan.results_only_open_ports", "true")
	p.SetString("scan.results_type_filters", "router,switch")
	p.SetString("scan.results_cidr_filter", "192.168.1.0/24")
	p.SetString("scan.results_port_state_mode", "has_open")

	ui := &ResultsUI{
		ResultsModeSel:       widget.NewRadioGroup([]string{"Таблица", "Карточки"}, nil),
		ResultsSubModeSel:    widget.NewRadioGroup([]string{"Устройства", "Безопасность"}, nil),
		ResultsSortSel:       widget.NewSelect([]string{"IP", "HostName"}, nil),
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		ResultsPortStateSel:  widget.NewSelect([]string{"Все", "Есть открытые"}, nil),
		ChipLimitSel:         widget.NewSelect([]string{"10", "50"}, nil),
		ShowRawBannersCheck:  widget.NewCheck("Raw", nil),
		OpenPortsOnlyCheck:   widget.NewCheck("Open", nil),
		QuickTypeChecks:      map[string]*widget.Check{"router": widget.NewCheck("", nil), "switch": widget.NewCheck("", nil)},
	}

	ctrl := NewResultsController(a, ui)
	ctrl.LoadSettings()

	if ui.ResultsModeSel.Selected != "Карточки" {
		t.Fatalf("view_mode = %q, want %q", ui.ResultsModeSel.Selected, "Карточки")
	}
	if ui.ResultsSubModeSel.Selected != "Безопасность" {
		t.Fatalf("submode = %q, want %q", ui.ResultsSubModeSel.Selected, "Безопасность")
	}
	if ui.ResultsSortSel.Selected != "HostName" {
		t.Fatalf("sort_mode = %q, want %q", ui.ResultsSortSel.Selected, "HostName")
	}
	if ui.ResultsFilterEnt.Text != "test" {
		t.Fatalf("filter_query = %q, want %q", ui.ResultsFilterEnt.Text, "test")
	}
	if ui.ResultsCidrFilterEnt.Text != "192.168.1.0/24" {
		t.Fatalf("cidr_filter = %q, want %q", ui.ResultsCidrFilterEnt.Text, "192.168.1.0/24")
	}
	if !ui.ShowRawBannersCheck.Checked {
		t.Fatal("ShowRawBannersCheck should be checked")
	}
	if !ui.OpenPortsOnlyCheck.Checked {
		t.Fatal("OpenPortsOnlyCheck should be checked")
	}
	if ui.QuickTypeChecks["router"] == nil || !ui.QuickTypeChecks["router"].Checked {
		t.Fatal("router check should be checked")
	}
	if ui.QuickTypeChecks["switch"] == nil || !ui.QuickTypeChecks["switch"].Checked {
		t.Fatal("switch check should be checked")
	}
}

// TestResultsController_LoadSettings_unknownValues проверяет, что неизвестные
// значения Preferences игнорируются.
func TestResultsController_LoadSettings_unknownValues(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()
	p.SetString("scan.results_view_mode", "НеизвестныйРежим")
	p.SetString("scan.results_submode", "НеизвестныйПод")
	p.SetString("scan.results_sort_mode", "НеизвестнаяСортировка")
	p.SetString("scan.results_port_state_mode", "has_nonexistent")
	p.SetString("scan.results_chip_limit", "0")

	ui := &ResultsUI{
		ResultsModeSel:       widget.NewRadioGroup([]string{"Таблица", "Карточки"}, nil),
		ResultsSubModeSel:    widget.NewRadioGroup([]string{"Devices"}, nil),
		ResultsSortSel:       widget.NewSelect([]string{"IP", "HostName"}, nil),
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		ResultsPortStateSel:  widget.NewSelect([]string{"Все", "Есть открытые"}, nil),
		ChipLimitSel:         widget.NewSelect([]string{"10", "20"}, nil),
		ShowRawBannersCheck:  widget.NewCheck("Raw", nil),
		OpenPortsOnlyCheck:   widget.NewCheck("Open", nil),
	}

	ctrl := NewResultsController(a, ui)
	ctrl.LoadSettings()

	// Неизвестные значения не должны применяться
	if ui.ResultsModeSel.Selected != "" {
		t.Fatalf("view_mode should not be set for unknown value: got %q", ui.ResultsModeSel.Selected)
	}
	if ui.ResultsSortSel.Selected != "" {
		t.Fatalf("sort_mode should not be set for unknown value: got %q", ui.ResultsSortSel.Selected)
	}
	// Неизвестный port_state_mode → default-ветка → SetSelected("Все")
	if ui.ResultsPortStateSel.Selected != "Все" {
		t.Fatalf("port_state should default to 'Все' for unknown value: got %q", ui.ResultsPortStateSel.Selected)
	}
}

// TestResultsController_SaveSettings_full сохраняет все настройки UI в Preferences.
func TestResultsController_SaveSettings_full(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()

	ui := &ResultsUI{
		ResultsModeSel:       widget.NewRadioGroup([]string{"Таблица", "Карточки"}, func(s string) {}),
		ResultsSubModeSel:    widget.NewRadioGroup([]string{"Устройства", "Безопасность"}, func(s string) {}),
		ResultsSortSel:       widget.NewSelect([]string{"IP", "HostName"}, func(s string) {}),
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		ResultsPortStateSel:  widget.NewSelect([]string{"Все", "Есть открытые"}, func(s string) {}),
		ChipLimitSel:         widget.NewSelect([]string{"10", "50"}, func(s string) {}),
		ShowRawBannersCheck:  widget.NewCheck("Raw", nil),
		OpenPortsOnlyCheck:   widget.NewCheck("Open", nil),
		QuickTypeChecks:      map[string]*widget.Check{"router": widget.NewCheck("", nil)},
	}

	ui.ResultsModeSel.SetSelected("Карточки")
	ui.ResultsSubModeSel.SetSelected("Безопасность")
	ui.ResultsSortSel.SetSelected("IP")
	ui.ResultsFilterEnt.SetText("filter123")
	ui.ResultsCidrFilterEnt.SetText("10.0.0.0/8")
	ui.ResultsPortStateSel.SetSelected("Есть открытые")
	ui.ChipLimitSel.SetSelected("50")
	ui.ShowRawBannersCheck.SetChecked(true)
	ui.OpenPortsOnlyCheck.SetChecked(true)
	ui.QuickTypeChecks["router"].SetChecked(true)

	ctrl := NewResultsController(a, ui)
	ctrl.SaveSettings()

	if p.String("scan.results_view_mode") != "Карточки" {
		t.Fatalf("view_mode = %q, want %q", p.String("scan.results_view_mode"), "Карточки")
	}
	if p.String("scan.results_submode") != "Безопасность" {
		t.Fatalf("submode = %q, want %q", p.String("scan.results_submode"), "Безопасность")
	}
	if p.String("scan.results_sort_mode") != "IP" {
		t.Fatalf("sort_mode = %q, want %q", p.String("scan.results_sort_mode"), "IP")
	}
	if p.String("scan.results_filter_query") != "filter123" {
		t.Fatalf("filter_query = %q, want %q", p.String("scan.results_filter_query"), "filter123")
	}
	if p.String("scan.results_cidr_filter") != "10.0.0.0/8" {
		t.Fatalf("cidr_filter = %q, want %q", p.String("scan.results_cidr_filter"), "10.0.0.0/8")
	}
	if p.String("scan.results_show_raw_banners") != "true" {
		t.Fatalf("show_raw_banners = %q, want %q", p.String("scan.results_show_raw_banners"), "true")
	}
	if p.String("scan.results_only_open_ports") != "true" {
		t.Fatalf("only_open_ports = %q, want %q", p.String("scan.results_only_open_ports"), "true")
	}
	if p.String("scan.results_type_filters") != "router" {
		t.Fatalf("type_filters = %q, want %q", p.String("scan.results_type_filters"), "router")
	}
}

// TestResultsController_SaveSettings_nilUI проверяет, что SaveSettings
// безопасно завершается при nil ui.
func TestResultsController_SaveSettings_nilUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	ctrl := NewResultsController(a, nil)
	ctrl.SaveSettings() // не должен паниковать
}

// TestResultsController_SaveSettings_nilChecks проверяет ветки с nil-чекбоксами.
func TestResultsController_SaveSettings_nilChecks(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()

	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		// ShowRawBannersCheck, OpenPortsOnlyCheck, QuickTypeChecks — nil
	}
	ui.ResultsFilterEnt.SetText("test")
	ui.ResultsCidrFilterEnt.SetText("192.168.0.0/16")

	ctrl := NewResultsController(a, ui)
	ctrl.SaveSettings()

	if p.String("scan.results_filter_query") != "test" {
		t.Fatalf("filter_query = %q, want %q", p.String("scan.results_filter_query"), "test")
	}
	if p.String("scan.results_cidr_filter") != "192.168.0.0/16" {
		t.Fatalf("cidr_filter = %q, want %q", p.String("scan.results_cidr_filter"), "192.168.0.0/16")
	}
}

// TestResultsController_FilterResults_queryIP проверяет фильтрацию по IP.
func TestResultsController_FilterResults_queryIP(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
	}
	// FilterResults читает текст из UI-поля, а не из Preferences
	ui.ResultsFilterEnt.SetText("10.0.0.1")

	ctrl := NewResultsController(a, ui)

	results := []scanner.Result{
		{IP: "10.0.0.1", Hostname: "server1"},
		{IP: "192.168.1.1", Hostname: "router"},
	}

	filtered := ctrl.FilterResults(results)
	if len(filtered) != 1 {
		t.Fatalf("filtered count = %d, want 1", len(filtered))
	}
	if filtered[0].IP != "10.0.0.1" {
		t.Fatalf("filtered IP = %q, want %q", filtered[0].IP, "10.0.0.1")
	}
}

// TestResultsController_FilterResults_queryHostname проверяет фильтрацию по hostname.
func TestResultsController_FilterResults_queryHostname(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
	}
	ui.ResultsFilterEnt.SetText("server")

	ctrl := NewResultsController(a, ui)

	results := []scanner.Result{
		{IP: "10.0.0.1", Hostname: "web-server"},
		{IP: "10.0.0.2", Hostname: "db-server"},
		{IP: "10.0.0.3", Hostname: "gateway"},
	}

	filtered := ctrl.FilterResults(results)
	if len(filtered) != 2 {
		t.Fatalf("filtered count = %d, want 2", len(filtered))
	}
}

// TestResultsController_FilterResults_caseInsensitive проверяет регистронезависимость.
func TestResultsController_FilterResults_caseInsensitive(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
	}
	ui.ResultsFilterEnt.SetText("SERVER")

	ctrl := NewResultsController(a, ui)

	results := []scanner.Result{
		{IP: "10.0.0.1", Hostname: "web-server"},
	}

	filtered := ctrl.FilterResults(results)
	if len(filtered) != 1 {
		t.Fatal("filter should be case-insensitive")
	}
}

// TestResultsController_FilterResults_emptyQuery проверяет, что при пустом
// query возвращаются все результаты.
func TestResultsController_FilterResults_emptyQuery(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
	}

	ctrl := NewResultsController(a, ui)

	results := []scanner.Result{
		{IP: "10.0.0.1", Hostname: "server1"},
		{IP: "10.0.0.2", Hostname: "server2"},
	}

	filtered := ctrl.FilterResults(results)
	if len(filtered) != 2 {
		t.Fatalf("filtered count = %d, want 2", len(filtered))
	}
}

// TestResultsController_FilterResults_cidr проверяет фильтрацию по CIDR.
func TestResultsController_FilterResults_cidr(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
	}
	ui.ResultsCidrFilterEnt.SetText("10.0.0.0/24")

	ctrl := NewResultsController(a, ui)

	results := []scanner.Result{
		{IP: "10.0.0.1", Hostname: "server1"},
		{IP: "10.0.1.1", Hostname: "server2"},
		{IP: "192.168.1.1", Hostname: "gateway"},
	}

	filtered := ctrl.FilterResults(results)
	if len(filtered) != 1 {
		t.Fatalf("filtered count = %d, want 1", len(filtered))
	}
	if filtered[0].IP != "10.0.0.1" {
		t.Fatalf("filtered IP = %q, want %q", filtered[0].IP, "10.0.0.1")
	}
}

// TestResultsController_FilterResults_openPortsOnly проверяет фильтр только
// открытых портов.
func TestResultsController_FilterResults_openPortsOnly(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ResultsUI{
		ResultsFilterEnt:     widget.NewEntry(),
		ResultsCidrFilterEnt: widget.NewEntry(),
		OpenPortsOnlyCheck:   widget.NewCheck("", nil),
	}
	ui.OpenPortsOnlyCheck.SetChecked(true)

	ctrl := NewResultsController(a, ui)

	results := []scanner.Result{
		{IP: "10.0.0.1", Ports: []scanner.PortInfo{{State: "open", Port: 80}}},
		{IP: "10.0.0.2", Ports: []scanner.PortInfo{{State: "closed", Port: 22}}},
		{IP: "10.0.0.3", Ports: []scanner.PortInfo{{State: "open", Port: 443}, {State: "closed", Port: 22}}},
	}

	filtered := ctrl.FilterResults(results)
	if len(filtered) != 2 {
		t.Fatalf("filtered count = %d, want 2", len(filtered))
	}
}

// TestResultsController_SortResults_emptyResults проверяет сортировку пустого slice.
func TestResultsController_SortResults_emptyResults(t *testing.T) {
	results := []scanner.Result{}

	ctrl := &ResultsController{}
	ctrl.SortResults(results, "IP") // не должен паниковать
}

// TestResultsController_SortResults_nilResults проверяет сортировку nil slice.
func TestResultsController_SortResults_nilResults(t *testing.T) {
	var results []scanner.Result

	ctrl := &ResultsController{}
	ctrl.SortResults(results, "IP") // не должен паниковать
}

// TestResultsController_SortResults_unknownMode проверяет, что неизвестный
// режим сортировки не вызывает панику.
func TestResultsController_SortResults_unknownMode(t *testing.T) {
	results := []scanner.Result{
		{IP: "10.0.0.3"},
		{IP: "10.0.0.1"},
	}

	ctrl := &ResultsController{}
	ctrl.SortResults(results, "UnknownSort") // должен проигнорировать
}
