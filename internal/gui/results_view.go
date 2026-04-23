package gui

import (
	"fmt"
	"math"
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

func (a *App) captureHostDetailsSplitOffsetBeforeRebuild() {
	if a == nil || a.resultsMainSplit == nil || a.lastHostDetailsSplitKind == "" {
		return
	}
	o := a.resultsMainSplit.Offset
	switch a.lastHostDetailsSplitKind {
	case "V":
		if o >= 0.28 && o <= 0.92 {
			a.rememberedHostDetailsSplitV = o
		}
	case "H":
		if o >= 0.35 && o <= 0.90 {
			a.rememberedHostDetailsSplitH = o
		}
	}
}

func (a *App) clearResultsMainSplitRef() {
	if a == nil {
		return
	}
	a.resultsMainSplit = nil
	a.lastHostDetailsSplitKind = ""
}

// renderScanResultsView перерисовывает область результатов (таблица или карточки).
func (a *App) renderScanResultsView() {
	if a.resultsBody == nil {
		return
	}
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

func clampFloat32(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampFloat64(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func absFloat32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// layoutAdaptiveMultiplier подбирает коэффициент для минимальных высот панелей по логическому
// размеру окна, «диагонали» (гипотенуза W×H) и масштабу канвы (плотность пикселей / HiDPI).
func layoutAdaptiveMultiplier(canvasW, canvasH, canvasScale float32) float32 {
	if canvasW <= 0 || canvasH <= 0 {
		return 1
	}
	if canvasScale <= 0 {
		canvasScale = 1
	}
	d := float32(math.Hypot(float64(canvasW), float64(canvasH)))
	const refDiag = float32(1150)
	ratio := d / refDiag
	scaleAdj := float32(math.Sqrt(float64(canvasScale)))
	scaleAdj = clampFloat32(scaleAdj, 0.88, 1.38)
	return clampFloat32(ratio/scaleAdj, 0.72, 1.38)
}

func suggestedScanTabOffset(profile string, canvasW, canvasH, canvasScale, layoutMul float32) float32 {
	_ = canvasW
	_ = canvasScale
	if canvasH <= 0 {
		return 0.38
	}
	topFrac := float32(0.36)
	switch profile {
	case "compact":
		topFrac = 0.41
	case "wide":
		topFrac = 0.32
	}
	if canvasH < 680 {
		topFrac -= 0.045
	}
	if canvasH > 920 {
		topFrac += 0.035
	}
	topFrac += (layoutMul - 1) * 0.025
	return clampFloat32(topFrac, 0.26, 0.54)
}

// defaultTopologySplitOffset — стартовая доля верхней панели (превью) для вкладки «Топология».
func defaultTopologySplitOffset(profile string) float64 {
	switch profile {
	case "compact":
		return 0.72
	case "wide":
		return 0.6
	default:
		return 0.62
	}
}

// defaultToolsSplitOffset — стартовая доля верхней зоны (поля + Operations) на вкладке «Инструменты».
func defaultToolsSplitOffset(profile string) float64 {
	switch profile {
	case "compact":
		return 0.48
	case "wide":
		return 0.40
	default:
		return 0.44
	}
}

func (a *App) currentLayoutAdaptiveMultiplier() float32 {
	if a == nil {
		return 1
	}
	w, h := a.lastCanvasSize.Width, a.lastCanvasSize.Height
	s := a.lastCanvasScale
	if w <= 0 || h <= 0 {
		return 1
	}
	if s <= 0 {
		s = 1
	}
	return layoutAdaptiveMultiplier(w, h, s)
}

func (a *App) adaptivePanelMinHeight(base, layoutMul, maxFracWindow, minAbs float32) float32 {
	if a == nil {
		return base
	}
	h := a.lastCanvasSize.Height
	if h <= 0 {
		h = 720
	}
	v := base * layoutMul
	maxH := h * maxFracWindow
	if v > maxH {
		v = maxH
	}
	if v < minAbs {
		v = minAbs
	}
	return v
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

func (a *App) applyAdaptiveToolsScrollMinSizes(profile string, mul float32) {
	if a == nil {
		return
	}
	baseOut := float32(360)
	baseOps := float32(140)
	switch profile {
	case "compact":
		baseOut, baseOps = 280, 110
	case "wide":
		baseOut, baseOps = 420, 170
	}
	if a.toolsOutputScroll != nil {
		a.toolsOutputScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(baseOut, mul, 0.52, 150)))
	}
	if a.operationsOutputScroll != nil {
		a.operationsOutputScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(baseOps, mul, 0.22, 72)))
	}
}

func (a *App) applyResponsiveLayout(profile string) {
	if a == nil {
		return
	}
	a.layoutProfile = profile
	mul := a.currentLayoutAdaptiveMultiplier()
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
			a.scanControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(170, mul, 0.40, 110)))
		}
		if a.resultsScroll != nil {
			a.resultsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(48, mul, 0.34, 32)))
		}
		if a.topologyControlsScroll != nil {
			a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(150, mul, 0.32, 100)))
		}
		if a.topologyScroll != nil {
			a.topologyScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(200, mul, 0.42, 120)))
		}
		if a.toolsControlsScroll != nil {
			a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(180, mul, 0.38, 120)))
		}
		if a.toolButtonsGrid != nil {
			a.toolButtonsGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.toolButtonsGrid.Refresh()
		}
		if a.operationsHeaderGrid != nil {
			a.operationsHeaderGrid.Layout = layout.NewGridLayoutWithColumns(1)
			a.operationsHeaderGrid.Refresh()
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
			a.scanControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(250, mul, 0.42, 140)))
		}
		if a.resultsScroll != nil {
			a.resultsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(70, mul, 0.34, 44)))
		}
		if a.topologyControlsScroll != nil {
			a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(220, mul, 0.34, 140)))
		}
		if a.topologyScroll != nil {
			a.topologyScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(350, mul, 0.48, 180)))
		}
		if a.toolsControlsScroll != nil {
			a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(320, mul, 0.40, 160)))
		}
		if a.toolButtonsGrid != nil {
			a.toolButtonsGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.toolButtonsGrid.Refresh()
		}
		if a.operationsHeaderGrid != nil {
			a.operationsHeaderGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.operationsHeaderGrid.Refresh()
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
			a.scanControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(220, mul, 0.38, 130)))
		}
		if a.resultsScroll != nil {
			a.resultsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(float32(75*0.775), mul, 0.34, 40)))
		}
		if a.topologyControlsScroll != nil {
			a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(200, mul, 0.32, 120)))
		}
		if a.topologyScroll != nil {
			a.topologyScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(300, mul, 0.45, 160)))
		}
		if a.toolsControlsScroll != nil {
			a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(260, mul, 0.38, 140)))
		}
		if a.toolButtonsGrid != nil {
			a.toolButtonsGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.toolButtonsGrid.Refresh()
		}
		if a.operationsHeaderGrid != nil {
			a.operationsHeaderGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.operationsHeaderGrid.Refresh()
		}
	}
	a.applyAdaptiveToolsScrollMinSizes(profile, mul)
	a.clampScanTabMainSplitOffset()
	a.clampTopologyMainSplitOffset()
	a.clampToolsTabMainSplitOffset()
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
		var (
			lastProfile      string
			lastW            float32 = -1
			lastH            float32 = -1
			lastScaleApplied float32 = -1
		)
		const sizeStep = 48
		for range ticker.C {
			size := fyne.NewSize(0, 0)
			scale := float32(1)
			fyne.DoAndWait(func() {
				if a.myWindow == nil || a.myWindow.Canvas() == nil {
					return
				}
				c := a.myWindow.Canvas()
				size = c.Size()
				scale = c.Scale()
			})
			if size.Width <= 0 || size.Height <= 0 {
				continue
			}
			if scale <= 0 {
				scale = 1
			}
			profile := a.detectLayoutProfile(size.Width)
			sizeChanged := lastW < 0 ||
				absFloat32(size.Width-lastW) >= sizeStep ||
				absFloat32(size.Height-lastH) >= sizeStep ||
				absFloat32(scale-lastScaleApplied) >= 0.06
			if profile == lastProfile && !sizeChanged {
				continue
			}
			lastW = size.Width
			lastH = size.Height
			lastScaleApplied = scale
			lastProfile = profile
			fyne.Do(func() {
				a.lastCanvasSize = size
				a.lastCanvasScale = scale
				a.applyResponsiveLayout(profile)
				if a.scanTabMainSplit != nil && !a.scanTabSplitInitialized && size.Height > 200 {
					a.scanTabSplitInitialized = true
					m := layoutAdaptiveMultiplier(size.Width, size.Height, scale)
					a.scanTabMainSplit.Offset = float64(suggestedScanTabOffset(profile, size.Width, size.Height, scale, m))
					a.clampScanTabMainSplitOffset()
				}
				if a.topologyMainSplit != nil && !a.topologySplitInitialized && size.Height > 200 {
					a.topologySplitInitialized = true
					a.topologyMainSplit.Offset = defaultTopologySplitOffset(profile)
					a.clampTopologyMainSplitOffset()
				}
				if a.toolsTabMainSplit != nil && !a.toolsSplitInitialized && size.Height > 200 {
					a.toolsSplitInitialized = true
					a.toolsTabMainSplit.Offset = defaultToolsSplitOffset(profile)
					a.clampToolsTabMainSplitOffset()
				}
				a.maybePersistScanTabSplitOffset()
				a.maybePersistTopologySplitOffset()
				a.maybePersistToolsTabSplitOffset()
				a.maybePersistHostDetailsSplitOffsets()
			})
		}
	}()
}

// applyDefaultSplitOffsetsForProfile выставляет разделители вкладок по профилю и размеру окна
// и записывает их в preferences (используется из меню «Вид» → сброс).
func (a *App) applyDefaultSplitOffsetsForProfile(profile string) {
	if a == nil {
		return
	}
	prof := strings.TrimSpace(profile)
	switch prof {
	case "compact", "wide", "normal":
	default:
		prof = "normal"
	}
	w, h := a.lastCanvasSize.Width, a.lastCanvasSize.Height
	if w <= 0 || h <= 0 {
		w, h = 1280, 720
	}
	s := a.lastCanvasScale
	if s <= 0 {
		s = 1
	}
	m := layoutAdaptiveMultiplier(w, h, s)
	if a.scanTabMainSplit != nil {
		a.scanTabMainSplit.Offset = float64(suggestedScanTabOffset(prof, w, h, s, m))
		a.clampScanTabMainSplitOffset()
		a.scanTabSplitInitialized = true
		a.scanTabSplitPersistPrimed = true
		a.lastPersistedScanSplit = a.scanTabMainSplit.Offset
		if a.myApp != nil {
			a.myApp.Preferences().SetFloat(prefScanTabSplitOffset, a.scanTabMainSplit.Offset)
		}
	}
	if a.topologyMainSplit != nil {
		a.topologyMainSplit.Offset = defaultTopologySplitOffset(prof)
		a.clampTopologyMainSplitOffset()
		a.topologySplitInitialized = true
		a.topologySplitPersistPrimed = true
		a.lastPersistedTopologySplit = a.topologyMainSplit.Offset
		if a.myApp != nil {
			a.myApp.Preferences().SetFloat(prefTopologyMainSplitOffset, a.topologyMainSplit.Offset)
		}
	}
	if a.toolsTabMainSplit != nil {
		a.toolsTabMainSplit.Offset = defaultToolsSplitOffset(prof)
		a.clampToolsTabMainSplitOffset()
		a.toolsSplitInitialized = true
		a.toolsSplitPersistPrimed = true
		a.lastPersistedToolsSplit = a.toolsTabMainSplit.Offset
		if a.myApp != nil {
			a.myApp.Preferences().SetFloat(prefToolsTabSplitOffset, a.toolsTabMainSplit.Offset)
		}
	}
}
