package controller

import (
	"net"
	"sort"
	"strconv"
	"strings"

	"network-scanner/internal/scanner"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// ResultsUI предоставляет доступ к виджетам результатов.
type ResultsUI struct {
	ResultsModeSel       *widget.RadioGroup
	ResultsSubModeSel    *widget.RadioGroup
	ResultsSortSel       *widget.Select
	ResultsFilterEnt     *widget.Entry
	ResultsCidrFilterEnt *widget.Entry
	ResultsPortStateSel  *widget.Select
	ChipLimitSel         *widget.Select
	ShowRawBannersCheck  *widget.Check
	OpenPortsOnlyCheck   *widget.Check
	QuickTypeChecks      map[string]*widget.Check
	StatusLabel          *widget.Label
}

// ResultsController управляет отображением и фильтрацией результатов.
type ResultsController struct {
	app fyne.App
	ui  *ResultsUI
}

// NewResultsController создает контроллер.
func NewResultsController(app fyne.App, ui *ResultsUI) *ResultsController {
	return &ResultsController{app: app, ui: ui}
}

// LoadSettings загружает настройки отображения.
func (c *ResultsController) LoadSettings() {
	if c.app == nil || c.ui == nil {
		return
	}
	p := c.app.Preferences()
	viewMode := p.String("scan.results_view_mode")
	if viewMode == "Таблица" || viewMode == "Карточки" {
		if c.ui.ResultsModeSel != nil {
			c.ui.ResultsModeSel.SetSelected(viewMode)
		}
	}
	subMode := p.String("scan.results_submode")
	if subMode == "Devices" || subMode == "Security" || subMode == "Inventory" {
		if c.ui.ResultsSubModeSel != nil {
			c.ui.ResultsSubModeSel.SetSelected(subMode)
		}
	}
	sortMode := p.String("scan.results_sort_mode")
	if sortMode == "IP" || sortMode == "HostName" {
		if c.ui.ResultsSortSel != nil {
			c.ui.ResultsSortSel.SetSelected(sortMode)
		}
	}
	if v, err := strconv.Atoi(p.String("scan.results_chip_limit")); err == nil && v > 0 {
		if c.ui.ChipLimitSel != nil {
			c.ui.ChipLimitSel.SetSelected(strconv.Itoa(v))
		}
	}
	c.ui.ShowRawBannersCheck.SetChecked(p.String("scan.results_show_raw_banners") == "true")
	c.ui.ResultsFilterEnt.SetText(p.String("scan.results_filter_query"))
	c.ui.OpenPortsOnlyCheck.SetChecked(p.String("scan.results_only_open_ports") == "true")
	if rawTypes := p.String("scan.results_type_filters"); rawTypes != "" {
		for _, typeName := range strings.Split(rawTypes, ",") {
			name := strings.TrimSpace(typeName)
			if check, ok := c.ui.QuickTypeChecks[name]; ok && check != nil {
				check.SetChecked(true)
			}
		}
	}
	if v := p.String("scan.results_cidr_filter"); v != "" {
		c.ui.ResultsCidrFilterEnt.SetText(v)
	}
	mode := p.String("scan.results_port_state_mode")
	switch mode {
	case "has_open":
		if c.ui.ResultsPortStateSel != nil {
			c.ui.ResultsPortStateSel.SetSelected("Есть открытые")
		}
	case "has_closed":
		if c.ui.ResultsPortStateSel != nil {
			c.ui.ResultsPortStateSel.SetSelected("Есть закрытые")
		}
	case "has_filtered":
		if c.ui.ResultsPortStateSel != nil {
			c.ui.ResultsPortStateSel.SetSelected("Есть фильтруемые")
		}
	default:
		if c.ui.ResultsPortStateSel != nil {
			c.ui.ResultsPortStateSel.SetSelected("Все")
		}
	}
}

// SaveSettings сохраняет настройки отображения.
func (c *ResultsController) SaveSettings() {
	if c.app == nil || c.ui == nil {
		return
	}
	p := c.app.Preferences()
	if c.ui.ResultsModeSel != nil {
		p.SetString("scan.results_view_mode", c.ui.ResultsModeSel.Selected)
	}
	if c.ui.ResultsSubModeSel != nil {
		p.SetString("scan.results_submode", c.ui.ResultsSubModeSel.Selected)
	}
	if c.ui.ResultsSortSel != nil {
		p.SetString("scan.results_sort_mode", c.ui.ResultsSortSel.Selected)
	}
	if c.ui.ChipLimitSel != nil {
		p.SetString("scan.results_chip_limit", c.ui.ChipLimitSel.Selected)
	}
	if c.ui.ShowRawBannersCheck != nil {
		if c.ui.ShowRawBannersCheck.Checked {
			p.SetString("scan.results_show_raw_banners", "true")
		} else {
			p.SetString("scan.results_show_raw_banners", "false")
		}
	}
	p.SetString("scan.results_filter_query", c.ui.ResultsFilterEnt.Text)
	if c.ui.OpenPortsOnlyCheck != nil {
		if c.ui.OpenPortsOnlyCheck.Checked {
			p.SetString("scan.results_only_open_ports", "true")
		} else {
			p.SetString("scan.results_only_open_ports", "false")
		}
	}
	selectedTypes := make([]string, 0)
	for typeName, check := range c.ui.QuickTypeChecks {
		if check != nil && check.Checked {
			selectedTypes = append(selectedTypes, typeName)
		}
	}
	if len(selectedTypes) > 1 {
		sort.Strings(selectedTypes)
	}
	p.SetString("scan.results_type_filters", strings.Join(selectedTypes, ","))
	if c.ui.ResultsCidrFilterEnt != nil {
		p.SetString("scan.results_cidr_filter", c.ui.ResultsCidrFilterEnt.Text)
	}
	if c.ui.ResultsPortStateSel != nil {
		p.SetString("scan.results_port_state_mode", c.ui.ResultsPortStateSel.Selected)
	}
}

// FilterResults применяет фильтры к результатам.
func (c *ResultsController) FilterResults(results []scanner.Result) []scanner.Result {
	query := strings.ToLower(c.ui.ResultsFilterEnt.Text)
	cidr := strings.TrimSpace(c.ui.ResultsCidrFilterEnt.Text)
	onlyOpen := c.ui.OpenPortsOnlyCheck != nil && c.ui.OpenPortsOnlyCheck.Checked
	filtered := make([]scanner.Result, 0, len(results))

	for _, r := range results {
		if query != "" {
			match := false
			match = match || strings.Contains(strings.ToLower(r.Hostname), query)
			match = match || strings.Contains(strings.ToLower(r.IP), query)
			match = match || strings.Contains(strings.ToLower(r.MAC), query)
			if !match {
				continue
			}
		}
		if cidr != "" {
			_, ipnet, err := net.ParseCIDR(cidr)
			if err == nil {
				ip := net.ParseIP(strings.TrimSpace(r.IP))
				if ip == nil || !ipnet.Contains(ip) {
					continue
				}
			}
		}
		if onlyOpen {
			hasOpen := false
			for _, p := range r.Ports {
				if p.State == "open" {
					hasOpen = true
					break
				}
			}
			if !hasOpen {
				continue
			}
		}
		filtered = append(filtered, r)
	}
	return filtered
}

// SortResults сортирует результаты.
func (c *ResultsController) SortResults(results []scanner.Result, mode string) {
	switch mode {
	case "IP":
		sort.Slice(results, func(i, j int) bool {
			return results[i].IP < results[j].IP
		})
	case "HostName":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Hostname < results[j].Hostname
		})
	}
}
