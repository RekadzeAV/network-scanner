package gui

import (
	"fmt"
	"net"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

// filteredSortedResults применяет фильтры и сортировку к a.scanResults.
func (a *App) filteredSortedResults() []scanner.Result {
	if len(a.scanResults) == 0 {
		return nil
	}
	base := filterResultsForDisplayAdvanced(
		a.scanResults,
		a.resultsFilterQuery,
		a.selectedTypeFilters(),
		a.onlyWithOpenPorts,
	)
	out := a.applyAdvancedFilters(base)
	return sortedResultsForDisplayWithMode(out, a.resultsSort)
}

func (a *App) currentDisplayedResults() []scanner.Result {
	return a.filteredSortedResults()
}

func (a *App) selectedTypeFilters() []string {
	if a == nil || len(a.quickTypeChecks) == 0 {
		return nil
	}
	out := make([]string, 0, len(a.quickTypeChecks))
	for name, ch := range a.quickTypeChecks {
		if ch != nil && ch.Checked {
			out = append(out, strings.TrimSpace(name))
		}
	}
	return out
}

func (a *App) applyAdvancedFilters(base []scanner.Result) []scanner.Result {
	out := make([]scanner.Result, 0, len(base))
	for _, r := range base {
		if !a.passesCIDRFilter(r) {
			continue
		}
		if !a.passesPortStateMode(r) {
			continue
		}
		out = append(out, r)
	}
	return out
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

	if strings.EqualFold(strings.TrimSpace(a.resultsSubMode), "Security") {
		a.resultsBody.Objects = []fyne.CanvasObject{a.buildSecurityDashboardView(filtered)}
		a.resultsBody.Refresh()
		return
	}

	var mainView fyne.CanvasObject
	if a.resultsMode == "Карточки" {
		mainView = a.buildCardsView(filtered)
	} else {
		mainView = a.buildTableView(filtered)
	}
	mainWithAnalytics := container.NewBorder(
		nil,
		a.buildResultsAnalyticsView(filtered),
		nil,
		nil,
		mainView,
	)
	detailsView := a.buildHostDetailsDrawer(filtered)
	if a.currentLayoutProfile() == "compact" {
		split := container.NewVSplit(mainWithAnalytics, detailsView)
		split.Offset = 0.7
		a.resultsMainSplit = split
		a.resultsBody.Objects = []fyne.CanvasObject{split}
	} else {
		split := container.NewHSplit(mainWithAnalytics, detailsView)
		split.Offset = 0.72
		a.resultsMainSplit = split
		a.resultsBody.Objects = []fyne.CanvasObject{split}
	}
	a.resultsBody.Refresh()
}

func (a *App) buildTableView(data []scanner.Result) fyne.CanvasObject {
	rows := len(data) + 1
	cols := 8
	headers := a.resultsTableHeaders()
	t := widget.NewTable(
		func() (int, int) { return rows, cols },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			l := obj.(*widget.Label)
			l.TextStyle = fyne.TextStyle{}
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
	widths := a.resultsTableColumnWidths()
	for col, width := range widths {
		t.SetColumnWidth(col, width)
	}
	t.OnSelected = func(id widget.TableCellID) {
		if id.Row <= 0 || id.Row-1 >= len(data) {
			return
		}
		a.selectHostForDetails(data[id.Row-1])
	}
	return t
}

func osGuessLine(r scanner.Result) string {
	if strings.TrimSpace(r.GuessOS) != "" {
		label := strings.TrimSpace(r.GuessOS)
		if strings.TrimSpace(r.GuessOSConfidence) != "" {
			label = fmt.Sprintf("%s (%s)", label, strings.TrimSpace(r.GuessOSConfidence))
		}
		if strings.TrimSpace(r.GuessOSReason) != "" {
			label += " — " + strings.TrimSpace(r.GuessOSReason)
		}
		return label
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
		item := r
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
			widget.NewButton("Открыть детали", func() {
				a.selectHostForDetails(item)
			}),
			widget.NewSeparator(),
		)
		bg := canvas.NewRectangle(tableRowBgColor)
		bg.CornerRadius = 4
		objs = append(objs, container.NewMax(bg, container.NewPadded(card)))
	}
	return container.NewVBox(objs...)
}

func (a *App) selectHostForDetails(r scanner.Result) {
	ip := strings.TrimSpace(r.IP)
	if ip == "" {
		return
	}
	a.selectedHostIP = ip
	a.renderScanResultsView()
}

func (a *App) selectedHostFromData(data []scanner.Result) (scanner.Result, bool) {
	if len(data) == 0 {
		return scanner.Result{}, false
	}
	selected := strings.TrimSpace(a.selectedHostIP)
	if selected != "" {
		for _, r := range data {
			if strings.TrimSpace(r.IP) == selected {
				return r, true
			}
		}
	}
	a.selectedHostIP = strings.TrimSpace(data[0].IP)
	return data[0], true
}

func (a *App) buildHostDetailsDrawer(data []scanner.Result) fyne.CanvasObject {
	r, ok := a.selectedHostFromData(data)
	if !ok {
		return widget.NewCard("Host Details", "", widget.NewLabel("Нет данных для отображения деталей."))
	}
	markdown := fmt.Sprintf(
		"### Host Details\n\n- Host: `%s`\n- IP: `%s`\n- MAC: `%s`\n- Type: `%s`\n- Vendor: `%s`\n- OS: `%s`\n- SNMP: `%t`\n- Open ports: `%d`",
		nullDash(r.Hostname),
		nullDash(r.IP),
		nullDash(r.MAC),
		nullDash(r.DeviceType),
		nullDash(r.DeviceVendor),
		osGuessLine(r),
		r.SNMPEnabled,
		countOpenPorts(r.Ports),
	)
	details := widget.NewRichTextFromMarkdown(markdown)
	details.Wrapping = fyne.TextWrapWord
	cols := 2
	if a.currentLayoutProfile() == "compact" {
		cols = 1
	}
	actions := a.buildHostQuickActions(r, cols)
	return widget.NewCard("Host Details Drawer", "Выбранный хост и быстрые действия", container.NewVBox(details, actions))
}

func (a *App) buildHostQuickActions(r scanner.Result, cols int) *fyne.Container {
	if cols <= 0 {
		cols = 1
	}
	return container.NewGridWithColumns(cols,
		widget.NewButton("Ping", func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectTabIndex(2)
			a.runPingTool()
		}),
		widget.NewButton("Traceroute", func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectTabIndex(2)
			a.runTracerouteTool()
		}),
		widget.NewButton("DNS", func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectTabIndex(2)
			a.runDNSTool()
		}),
		widget.NewButton("Whois", func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectTabIndex(2)
			a.runWhoisTool()
		}),
		widget.NewButton("Wake-on-LAN", func() {
			if a.toolsWOLMacEntry != nil {
				a.toolsWOLMacEntry.SetText(strings.TrimSpace(r.MAC))
			}
			a.mainTabs.SelectTabIndex(2)
		}),
	)
}

func countOpenPorts(ports []scanner.PortInfo) int {
	n := 0
	for _, p := range ports {
		if strings.EqualFold(strings.TrimSpace(p.State), "open") {
			n++
		}
	}
	return n
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
		if strings.TrimSpace(p.Version) != "" {
			lbl += " · " + truncateStr(p.Version, 40)
		}
		if a.showRawBanners && strings.TrimSpace(p.Banner) != "" {
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

func (a *App) currentLayoutProfile() string {
	if a == nil {
		return "normal"
	}
	if a.layoutProfile == "" {
		return "normal"
	}
	return a.layoutProfile
}

func (a *App) detectLayoutProfile(width float32) string {
	switch {
	case width <= 1366:
		return "compact"
	case width >= 2200:
		return "wide"
	default:
		return "normal"
	}
}

func (a *App) resultsTableColumnWidths() []float32 {
	profile := a.currentLayoutProfile()
	base := []float32{140, 120, 130, 120, 120, 140, 52, 280}
	if profile == "wide" {
		return []float32{180, 150, 170, 150, 170, 220, 80, 420}
	}
	if profile == "compact" {
		return []float32{110, 100, 96, 96, 110, 120, 52, 180}
	}
	return base
}

func (a *App) resultsTableHeaders() []string {
	if a.currentLayoutProfile() == "compact" {
		return []string{"Host", "IP", "MAC", "Тип", "Вендор", "OS", "SNMP", "Порты"}
	}
	return []string{"Host", "IP", "MAC", "Тип", "Производитель", "ОС (оценка)", "SNMP", "Порты (открытые)"}
}

func (a *App) applyResponsiveLayout(profile string) {
	if a == nil {
		return
	}
	a.layoutProfile = profile
	switch profile {
	case "compact":
		if a.resultsDiagnosticsGrid != nil {
			a.resultsDiagnosticsGrid.Layout = layout.NewGridLayoutWithColumns(1)
			a.resultsDiagnosticsGrid.Refresh()
		}
		if a.resultsSortGrid != nil {
			a.resultsSortGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.resultsSortGrid.Refresh()
		}
		if a.resultsCidrGrid != nil {
			a.resultsCidrGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.resultsCidrGrid.Refresh()
		}
		if a.resultsPresetGrid != nil {
			a.resultsPresetGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.resultsPresetGrid.Refresh()
		}
		if a.scanControlsScroll != nil {
			a.scanControlsScroll.SetMinSize(fyne.NewSize(0, 170))
		}
		if a.resultsScroll != nil {
			a.resultsScroll.SetMinSize(fyne.NewSize(0, 96))
		}
		if a.topologyControlsScroll != nil {
			a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, 150))
		}
		if a.topologyScroll != nil {
			a.topologyScroll.SetMinSize(fyne.NewSize(0, 200))
		}
		if a.toolsControlsScroll != nil {
			a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, 180))
		}
		if a.toolButtonsGrid != nil {
			a.toolButtonsGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.toolButtonsGrid.Refresh()
		}
		if a.operationsHeaderGrid != nil {
			a.operationsHeaderGrid.Layout = layout.NewGridLayoutWithColumns(1)
			a.operationsHeaderGrid.Refresh()
		}
		if a.topologyMainSplit != nil {
			a.topologyMainSplit.Offset = 0.72
		}
	case "wide":
		if a.resultsDiagnosticsGrid != nil {
			a.resultsDiagnosticsGrid.Layout = layout.NewGridLayoutWithColumns(3)
			a.resultsDiagnosticsGrid.Refresh()
		}
		if a.resultsSortGrid != nil {
			a.resultsSortGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.resultsSortGrid.Refresh()
		}
		if a.resultsCidrGrid != nil {
			a.resultsCidrGrid.Layout = layout.NewGridLayoutWithColumns(4)
			a.resultsCidrGrid.Refresh()
		}
		if a.resultsPresetGrid != nil {
			a.resultsPresetGrid.Layout = layout.NewGridLayoutWithColumns(4)
			a.resultsPresetGrid.Refresh()
		}
		if a.scanControlsScroll != nil {
			a.scanControlsScroll.SetMinSize(fyne.NewSize(0, 250))
		}
		if a.resultsScroll != nil {
			a.resultsScroll.SetMinSize(fyne.NewSize(0, 140))
		}
		if a.topologyControlsScroll != nil {
			a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, 220))
		}
		if a.topologyScroll != nil {
			a.topologyScroll.SetMinSize(fyne.NewSize(0, 350))
		}
		if a.toolsControlsScroll != nil {
			a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, 320))
		}
		if a.toolButtonsGrid != nil {
			a.toolButtonsGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.toolButtonsGrid.Refresh()
		}
		if a.operationsHeaderGrid != nil {
			a.operationsHeaderGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.operationsHeaderGrid.Refresh()
		}
		if a.topologyMainSplit != nil {
			a.topologyMainSplit.Offset = 0.6
		}
	default:
		if a.resultsDiagnosticsGrid != nil {
			a.resultsDiagnosticsGrid.Layout = layout.NewGridLayoutWithColumns(3)
			a.resultsDiagnosticsGrid.Refresh()
		}
		if a.resultsSortGrid != nil {
			a.resultsSortGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.resultsSortGrid.Refresh()
		}
		if a.resultsCidrGrid != nil {
			a.resultsCidrGrid.Layout = layout.NewGridLayoutWithColumns(4)
			a.resultsCidrGrid.Refresh()
		}
		if a.resultsPresetGrid != nil {
			a.resultsPresetGrid.Layout = layout.NewGridLayoutWithColumns(4)
			a.resultsPresetGrid.Refresh()
		}
		if a.scanControlsScroll != nil {
			a.scanControlsScroll.SetMinSize(fyne.NewSize(0, 220))
		}
		if a.resultsScroll != nil {
			a.resultsScroll.SetMinSize(fyne.NewSize(0, float32(75*1.55)))
		}
		if a.topologyControlsScroll != nil {
			a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, 200))
		}
		if a.topologyScroll != nil {
			a.topologyScroll.SetMinSize(fyne.NewSize(0, 300))
		}
		if a.toolsControlsScroll != nil {
			a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, 260))
		}
		if a.toolButtonsGrid != nil {
			a.toolButtonsGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.toolButtonsGrid.Refresh()
		}
		if a.operationsHeaderGrid != nil {
			a.operationsHeaderGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.operationsHeaderGrid.Refresh()
		}
		if a.topologyMainSplit != nil {
			a.topologyMainSplit.Offset = 0.62
		}
	}
	if a.resultsBody != nil {
		a.renderScanResultsView()
	}
}

func (a *App) startResultsLayoutWatcher() {
	if a == nil || a.myWindow == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()
		var lastProfile string
		for range ticker.C {
			size := fyne.NewSize(0, 0)
			fyne.DoAndWait(func() {
				if a.myWindow == nil || a.myWindow.Canvas() == nil {
					return
				}
				size = a.myWindow.Canvas().Size()
			})
			if size.Width <= 0 || size.Height <= 0 {
				continue
			}
			profile := a.detectLayoutProfile(size.Width)
			if profile == lastProfile {
				continue
			}
			lastProfile = profile
			fyne.Do(func() {
				a.lastCanvasSize = size
				a.applyResponsiveLayout(profile)
			})
		}
	}()
}
