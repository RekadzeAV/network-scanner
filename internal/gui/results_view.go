package gui

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

// filteredSortedResults применяет фильтры и сортировку к a.scanResults.
func (a *App) filteredSortedResults() []scanner.Result {
	if len(a.scanResults) == 0 {
		return nil
	}
	out := make([]scanner.Result, 0, len(a.scanResults))
	for _, r := range a.scanResults {
		if !a.passesTextFilter(r) {
			continue
		}
		if !a.passesTypeFilters(r) {
			continue
		}
		if a.onlyWithOpenPorts && !deviceHasOpenPort(r) {
			continue
		}
		if !a.passesCIDRFilter(r) {
			continue
		}
		if !a.passesPortStateMode(r) {
			continue
		}
		out = append(out, r)
	}
	sortResultsSlice(out, a.resultsSort)
	return out
}

func (a *App) currentDisplayedResults() []scanner.Result {
	return a.filteredSortedResults()
}

func deviceHasOpenPort(r scanner.Result) bool {
	for _, p := range r.Ports {
		if p.State == "open" {
			return true
		}
	}
	return false
}

func (a *App) passesTextFilter(r scanner.Result) bool {
	q := strings.TrimSpace(strings.ToLower(a.resultsFilterQuery))
	if q == "" {
		return true
	}
	hay := strings.ToLower(strings.Join([]string{
		r.IP, r.MAC, r.Hostname, r.DeviceType, r.DeviceVendor,
	}, " "))
	return strings.Contains(hay, q)
}

func normalizedDeviceCategory(dt string) string {
	typeMapping := map[string]string{
		"Router/Network Device": "Network Device",
		"Network Device":        "Network Device",
		"Router":                "Network Device",
		"Printer":               "Network Device",
		"IoT Device":            "Network Device",
		"IoT":                   "Network Device",
		"Windows Computer":      "Computer",
		"Computer":              "Computer",
		"Windows":               "Computer",
		"PC":                    "Computer",
		"Desktop":               "Computer",
		"Laptop":                "Computer",
		"Web Server":            "Server",
		"Database Server":       "Server",
		"Linux/Unix Server":     "Server",
		"Server":                "Server",
		"Linux Server":          "Server",
		"Unix Server":           "Server",
		"Linux":                 "Server",
		"Unix":                  "Server",
		"Unknown Device":        "Unknown",
		"Unknown":               "Unknown",
	}
	if v, ok := typeMapping[strings.TrimSpace(dt)]; ok {
		return v
	}
	return strings.TrimSpace(dt)
}

func (a *App) passesTypeFilters(r scanner.Result) bool {
	any := false
	for _, ch := range a.quickTypeChecks {
		if ch != nil && ch.Checked {
			any = true
			break
		}
	}
	if !any {
		return true
	}
	cat := normalizedDeviceCategory(r.DeviceType)
	for name, ch := range a.quickTypeChecks {
		if ch != nil && ch.Checked && name == cat {
			return true
		}
	}
	return false
}

func (a *App) passesCIDRFilter(r scanner.Result) bool {
	if a.resultsCidrFilterEnt == nil {
		return true
	}
	cidr := strings.TrimSpace(a.resultsCidrFilterEnt.Text)
	if cidr == "" {
		return true
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return true
	}
	ip := net.ParseIP(strings.TrimSpace(r.IP))
	if ip == nil {
		return false
	}
	return ipnet.Contains(ip)
}

func (a *App) passesPortStateMode(r scanner.Result) bool {
	mode := strings.TrimSpace(a.resultsPortStateMode)
	if mode == "" || mode == "all" {
		return true
	}
	switch mode {
	case "has_open":
		for _, p := range r.Ports {
			if p.State == "open" {
				return true
			}
		}
		return false
	case "has_closed":
		for _, p := range r.Ports {
			if p.State == "closed" {
				return true
			}
		}
		return false
	case "has_filtered":
		for _, p := range r.Ports {
			if p.State == "filtered" {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func sortResultsSlice(results []scanner.Result, mode string) {
	sort.SliceStable(results, func(i, j int) bool {
		if strings.TrimSpace(mode) == "HostName" {
			hi := strings.ToLower(strings.TrimSpace(results[i].Hostname))
			hj := strings.ToLower(strings.TrimSpace(results[j].Hostname))
			if hi != hj {
				return hi < hj
			}
		}
		a1 := net.ParseIP(strings.TrimSpace(results[i].IP))
		a2 := net.ParseIP(strings.TrimSpace(results[j].IP))
		if a1 != nil && a1.To4() != nil && a2 != nil && a2.To4() != nil {
			return bytes.Compare(a1.To4(), a2.To4()) < 0
		}
		return strings.TrimSpace(results[i].IP) < strings.TrimSpace(results[j].IP)
	})
}

func (a *App) activeFilterCount() int {
	n := 0
	if strings.TrimSpace(a.resultsFilterQuery) != "" {
		n++
	}
	for _, ch := range a.quickTypeChecks {
		if ch != nil && ch.Checked {
			n++
		}
	}
	if a.onlyWithOpenPorts {
		n++
	}
	if a.resultsCidrFilterEnt != nil && strings.TrimSpace(a.resultsCidrFilterEnt.Text) != "" {
		n++
	}
	if a.resultsPortStateMode != "" && a.resultsPortStateMode != "all" {
		n++
	}
	return n
}

func (a *App) updateFiltersInfoLabel() {
	if a.filtersInfoLabel == nil {
		return
	}
	a.filtersInfoLabel.SetText(fmt.Sprintf("Активных фильтров: %d", a.activeFilterCount()))
}

// renderScanResultsView перерисовывает область результатов (таблица или карточки).
func (a *App) renderScanResultsView() {
	if a.resultsBody == nil {
		return
	}
	a.updateFiltersInfoLabel()

	switch a.resultsState {
	case resultsStateIdle:
		a.resultsBody.Objects = []fyne.CanvasObject{
			container.NewCenter(widget.NewLabel("Результаты сканирования появятся здесь после запуска.")),
		}
		a.resultsBody.Refresh()
		return
	case resultsStateScanning:
		a.resultsBody.Objects = []fyne.CanvasObject{
			container.NewCenter(widget.NewLabel("Сканирование...")),
		}
		a.resultsBody.Refresh()
		return
	}

	filtered := a.filteredSortedResults()
	if len(a.scanResults) == 0 {
		msg := "Результаты не найдены."
		switch a.resultsState {
		case resultsStateStopped:
			msg = "Сканирование остановлено. Результаты могут быть неполными."
		case resultsStateTimeout:
			msg = "Сканирование прервано по таймауту."
		}
		a.resultsBody.Objects = []fyne.CanvasObject{container.NewCenter(widget.NewLabel(msg))}
		a.resultsBody.Refresh()
		return
	}

	if len(filtered) == 0 {
		a.resultsBody.Objects = []fyne.CanvasObject{
			container.NewCenter(widget.NewLabel("Нет устройств, подходящих под текущие фильтры.")),
		}
		a.resultsBody.Refresh()
		return
	}

	if a.resultsMode == "Карточки" {
		a.resultsBody.Objects = []fyne.CanvasObject{a.buildCardsView(filtered)}
	} else {
		a.resultsBody.Objects = []fyne.CanvasObject{a.buildTableView(filtered)}
	}
	a.resultsBody.Refresh()
}

func (a *App) buildTableView(data []scanner.Result) fyne.CanvasObject {
	rows := len(data) + 1
	cols := 8
	t := widget.NewTable(
		func() (int, int) { return rows, cols },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			l := obj.(*widget.Label)
			l.TextStyle = fyne.TextStyle{}
			headers := []string{"Host", "IP", "MAC", "Тип", "Производитель", "ОС (оценка)", "SNMP", "Порты (открытые)"}
			if id.Row == 0 {
				l.TextStyle = fyne.TextStyle{Bold: true}
				if id.Col < len(headers) {
					l.SetText(headers[id.Col])
				}
				return
			}
			r := data[id.Row-1]
			switch id.Col {
			case 0:
				l.SetText(nullDash(r.Hostname))
			case 1:
				l.SetText(r.IP)
			case 2:
				l.SetText(nullDash(r.MAC))
			case 3:
				l.SetText(nullDash(r.DeviceType))
			case 4:
				l.SetText(nullDash(r.DeviceVendor))
			case 5:
				l.SetText(osGuessLine(r))
			case 6:
				if r.SNMPEnabled {
					l.SetText("да")
				} else {
					l.SetText("нет")
				}
			case 7:
				l.SetText(formatPorts(r.Ports))
			default:
				l.SetText("")
			}
		},
	)
	t.SetColumnWidth(0, 140)
	t.SetColumnWidth(1, 120)
	t.SetColumnWidth(2, 130)
	t.SetColumnWidth(3, 120)
	t.SetColumnWidth(4, 120)
	t.SetColumnWidth(5, 140)
	t.SetColumnWidth(6, 52)
	t.SetColumnWidth(7, 280)
	return t
}

func osGuessLine(r scanner.Result) string {
	if strings.TrimSpace(r.GuessOS) != "" {
		if strings.TrimSpace(r.GuessOSConfidence) != "" {
			return fmt.Sprintf("%s (%s)", strings.TrimSpace(r.GuessOS), strings.TrimSpace(r.GuessOSConfidence))
		}
		return strings.TrimSpace(r.GuessOS)
	}
	return "-"
}

func nullDash(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	return s
}

func (a *App) buildCardsView(data []scanner.Result) fyne.CanvasObject {
	objs := make([]fyne.CanvasObject, 0, len(data))
	for _, r := range data {
		title := strings.TrimSpace(r.Hostname)
		if title == "" {
			title = r.IP
		}
		head := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		sub := widget.NewLabel(fmt.Sprintf("%s · %s · %s", r.IP, nullDash(r.MAC), nullDash(r.DeviceType)))
		sub.Wrapping = fyne.TextWrapWord
		chipRow := a.buildPortChips(r)
		card := container.NewVBox(
			head,
			sub,
			widget.NewLabel(fmt.Sprintf("Производитель: %s", nullDash(r.DeviceVendor))),
			widget.NewLabel(fmt.Sprintf("ОС (оценка): %s", osGuessLine(r))),
			widget.NewLabel("Порты:"),
			chipRow,
			widget.NewSeparator(),
		)
		bg := canvas.NewRectangle(tableRowBgColor)
		bg.CornerRadius = 4
		objs = append(objs, container.NewMax(bg, container.NewPadded(card)))
	}
	return container.NewVBox(objs...)
}

func (a *App) buildPortChips(r scanner.Result) fyne.CanvasObject {
	var open []scanner.PortInfo
	for _, p := range r.Ports {
		if p.State == "open" {
			open = append(open, p)
		}
	}
	if len(open) == 0 {
		return widget.NewLabel("нет открытых")
	}
	limit := a.maxPortChips
	if limit <= 0 {
		limit = 24
	}
	row := make([]fyne.CanvasObject, 0)
	for i, p := range open {
		if i >= limit {
			row = append(row, widget.NewLabel(fmt.Sprintf("… +%d", len(open)-limit)))
			break
		}
		lbl := p.Service
		if lbl == "" || lbl == "Unknown" {
			lbl = fmt.Sprintf("%d/%s", p.Port, p.Protocol)
		} else {
			lbl = fmt.Sprintf("%d %s", p.Port, lbl)
		}
		if strings.TrimSpace(p.Banner) != "" {
			lbl += " · " + truncateStr(p.Banner, 40)
		}
		t := widget.NewLabel(lbl)
		bg := canvas.NewRectangle(chipBgColor)
		bg.CornerRadius = 3
		row = append(row, container.NewMax(bg, container.NewPadded(t)))
	}
	return container.NewHBox(row...)
}

func truncateStr(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

// startResultsLayoutWatcher зарезервирован для будущей подписки на resize/тему; прокрутка Fyne уже адаптивна.
func (a *App) startResultsLayoutWatcher() {}
