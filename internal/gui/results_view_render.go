package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

// renderScanResultsView перерисовывает область результатов (таблица или карточки).
func (a *App) renderScanResultsView() {
	if a.resultsBody == nil {
		return
	}
	started := time.Now()
	defer func() {
		a.lastRenderStats.Duration = time.Since(started)
		a.updateResultsPerfLabel(a.lastRenderStats)
	}()
	a.updateFiltersInfoLabel()
	a.captureHostDetailsSplitOffsetBeforeRebuild()

	switch a.resultsState {
	case resultsStateIdle:
		a.resultsBody.Objects = []fyne.CanvasObject{
			container.NewCenter(widget.NewLabel("Результаты сканирования появятся здесь после запуска.")),
		}
		a.resultsBody.Refresh()
		a.clearResultsMainSplitRef()
		return
	case resultsStateScanning:
		a.resultsBody.Objects = []fyne.CanvasObject{
			container.NewCenter(widget.NewLabel("Сканирование...")),
		}
		a.resultsBody.Refresh()
		a.clearResultsMainSplitRef()
		return
	}

	filtered := a.filteredSortedResults()
	a.lastRenderStats = resultsRenderStats{
		FilteredCount: len(filtered),
		VisibleCount:  len(filtered),
	}
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
		a.clearResultsMainSplitRef()
		return
	}

	if len(filtered) == 0 {
		a.resultsBody.Objects = []fyne.CanvasObject{
			container.NewCenter(widget.NewLabel("Нет устройств, подходящих под текущие фильтры.")),
		}
		a.resultsBody.Refresh()
		a.clearResultsMainSplitRef()
		return
	}

	if strings.EqualFold(strings.TrimSpace(a.resultsSubMode), "Security") {
		a.resultsBody.Objects = []fyne.CanvasObject{a.buildSecurityDashboardView(filtered)}
		a.resultsBody.Refresh()
		a.clearResultsMainSplitRef()
		return
	}
	if strings.EqualFold(strings.TrimSpace(a.resultsSubMode), "Inventory") {
		a.resultsBody.Objects = []fyne.CanvasObject{a.buildInventoryDashboardView()}
		a.resultsBody.Refresh()
		a.clearResultsMainSplitRef()
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
		offV := 0.7
		if a.rememberedHostDetailsSplitV > 0 {
			offV = clampFloat64(a.rememberedHostDetailsSplitV, 0.28, 0.92)
		}
		split.Offset = offV
		a.resultsMainSplit = split
		a.lastHostDetailsSplitKind = "V"
		a.resultsBody.Objects = []fyne.CanvasObject{split}
	} else {
		split := container.NewHSplit(mainWithAnalytics, detailsView)
		offH := 0.72
		if a.rememberedHostDetailsSplitH > 0 {
			offH = clampFloat64(a.rememberedHostDetailsSplitH, 0.35, 0.90)
		}
		split.Offset = offH
		a.resultsMainSplit = split
		a.lastHostDetailsSplitKind = "H"
		a.resultsBody.Objects = []fyne.CanvasObject{split}
	}
	a.resultsBody.Refresh()
}

func (a *App) buildHostDetailsDrawer(data []scanner.Result) fyne.CanvasObject {
	r, ok := a.selectedHostFromData(data)
	if !ok {
		return widget.NewCard("Host Details", "", widget.NewLabel("Нет данных для отображения деталей."))
	}
	markdown := a.hostDetailsMarkdown(r)
	details := widget.NewRichTextFromMarkdown(markdown)
	details.Wrapping = fyne.TextWrapWord
	cols := 2
	if a.currentLayoutProfile() == "compact" {
		cols = 1
	}
	actions := a.buildHostQuickActions(r, cols)
	a.prefetchHostDetailsNearby(data, strings.TrimSpace(r.IP))
	return widget.NewCard("Host Details Drawer", "Выбранный хост и быстрые действия", container.NewVBox(details, actions))
}

func (a *App) hostDetailsMarkdown(r scanner.Result) string {
	ip := strings.TrimSpace(r.IP)
	if ip == "" {
		return "### Host Details\n\n- Нет данных"
	}
	a.hostDetailsCacheMu.RLock()
	if v, ok := a.hostDetailsCache[ip]; ok {
		a.hostDetailsCacheMu.RUnlock()
		return v
	}
	a.hostDetailsCacheMu.RUnlock()
	md := fmt.Sprintf(
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
	a.hostDetailsCacheMu.Lock()
	if a.hostDetailsCache == nil {
		a.hostDetailsCache = make(map[string]string)
	}
	a.hostDetailsCache[ip] = md
	a.hostDetailsCacheMu.Unlock()
	return md
}

func (a *App) primeHostDetailsCache(r scanner.Result) {
	_ = a.hostDetailsMarkdown(r)
}

func (a *App) prefetchHostDetailsNearby(data []scanner.Result, selectedIP string) {
	if len(data) == 0 || strings.TrimSpace(selectedIP) == "" {
		return
	}
	idx := -1
	for i, r := range data {
		if strings.TrimSpace(r.IP) == selectedIP {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	start := idx - 3
	if start < 0 {
		start = 0
	}
	end := idx + 3
	if end >= len(data) {
		end = len(data) - 1
	}
	snapshot := append([]scanner.Result(nil), data[start:end+1]...)
	go func(items []scanner.Result) {
		for _, item := range items {
			a.primeHostDetailsCache(item)
		}
	}(snapshot)
}

func (a *App) buildHostQuickActions(r scanner.Result, cols int) *fyne.Container {
	if cols <= 0 {
		cols = 1
	}
	return container.NewGridWithColumns(cols,
		widget.NewButtonWithIcon("Ping", iconRefresh(), func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectIndex(2)
			a.runPingTool()
		}),
		widget.NewButtonWithIcon("Traceroute", iconRoute(), func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectIndex(2)
			a.runTracerouteTool()
		}),
		widget.NewButtonWithIcon("DNS", iconSearch(), func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectIndex(2)
			a.runDNSTool()
		}),
		widget.NewButtonWithIcon("Whois", iconAccount(), func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectIndex(2)
			a.toolsCtrl.RunWhoisTool()
		}),
		widget.NewButtonWithIcon("Wake-on-LAN", iconScan(), func() {
			if a.toolsWOLMacEntry != nil {
				a.toolsWOLMacEntry.SetText(strings.TrimSpace(r.MAC))
			}
			a.mainTabs.SelectIndex(2)
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
		bg := canvas.NewRectangle(a.chipBackground())
		bg.CornerRadius = 3
		row = append(row, container.NewStack(bg, container.NewPadded(t)))
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
