package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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

// portChipSize — размер ячейки чипа в GridWrap (фиксированный: Fyne не умеет
// flow-wrap для объектов переменной ширины, фикс. ячейка даёт перенос по ширине
// без горизонтального «хвоста» за карточку — критерий P1-4).
var portChipSize = fyne.NewSize(200, 30)

// buildPortChips — чипы портов как семантический компонент (R-05/P1-4):
// открытые первыми, затем closed/filtered (если скан их возвращал); состояние
// различимо маркером формы и палитрой состояний, GridWrap переносит чипы.
func (a *App) buildPortChips(r scanner.Result) fyne.CanvasObject {
	chips := portChipOrder(r.Ports)
	if len(chips) == 0 {
		return widget.NewLabel("нет открытых")
	}
	limit := a.maxPortChips
	if limit <= 0 {
		limit = 24
	}
	cells := make([]fyne.CanvasObject, 0, limit+1)
	for i, p := range chips {
		if i >= limit {
			cells = append(cells, widget.NewLabel(fmt.Sprintf("… +%d", len(chips)-limit)))
			break
		}
		cells = append(cells, a.buildPortChip(p))
	}
	return container.NewGridWrap(portChipSize, cells...)
}

// portChipOrder — порядок чипов: сначала open (стабильно, в исходном порядке),
// затем остальные состояния. Хост без портов даёт пустой срез.
func portChipOrder(ports []scanner.PortInfo) []scanner.PortInfo {
	open := make([]scanner.PortInfo, 0, len(ports))
	other := make([]scanner.PortInfo, 0, len(ports))
	for _, p := range ports {
		if portStateKey(p.State) == "open" {
			open = append(open, p)
			continue
		}
		other = append(other, p)
	}
	return append(open, other...)
}

// buildPortChip — чип одного порта: подложка и текст из палитры состояний,
// рамка border-strong 1 px, маркер формы состояния (WCAG 1.4.1).
func (a *App) buildPortChip(p scanner.PortInfo) fyne.CanvasObject {
	st := portStateStyleFor(p.State)
	label := portChipLabel(p, a.showRawBanners)
	if st.Marker != "" {
		label = st.Marker + " " + label
	}
	txt := canvas.NewText(label, st.Text)
	txt.TextSize = theme.CaptionTextSize()
	bg := canvas.NewRectangle(st.Fill)
	bg.CornerRadius = 3
	bg.StrokeColor = paletteBorderStrong(currentThemeIsDark())
	bg.StrokeWidth = 1
	return container.NewStack(bg, container.NewCenter(txt))
}

// portChipLabel — текст чипа: сервис (или порт/протокол), версия и, по
// настройке, баннер; всё усечено по рунам (R-11) до ширины ячейки.
func portChipLabel(p scanner.PortInfo, showBanner bool) string {
	lbl := p.Service
	if lbl == "" || lbl == "Unknown" {
		lbl = fmt.Sprintf("%d/%s", p.Port, p.Protocol)
	} else {
		lbl = fmt.Sprintf("%d %s", p.Port, lbl)
	}
	if v := strings.TrimSpace(p.Version); v != "" {
		lbl += " · " + truncateStr(v, 24)
	}
	if showBanner {
		if b := strings.TrimSpace(p.Banner); b != "" {
			lbl += " · " + truncateStr(b, 24)
		}
	}
	return lbl
}

// truncateStr — rune-safe усечение (делегрует truncateString, R-11).
func truncateStr(s string, n int) string {
	return truncateString(strings.TrimSpace(s), n)
}
