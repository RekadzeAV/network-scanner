package gui

import (
	"fmt"
	"strconv"
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
		a.updateResultsCountLabel(0, 0)
		a.resultsBody.Objects = []fyne.CanvasObject{
			buildResultsEmptyState(
				"Ожидание сканирования",
				"Результаты появятся здесь после запуска сканирования.",
				iconScan(),
				"",
				nil,
			),
		}
		a.resultsBody.Refresh()
		a.clearResultsMainSplitRef()
		return
	case resultsStateScanning:
		// R-04/P1-1: live-результаты. Пока не найден ни один хост — показываем
		// заглушку ожидания; с первым найденным устройством область результатов
		// занимает обычная таблица/карточки, которые обновляются по ходу скана.
		if len(a.scanResults) == 0 {
			a.updateResultsCountLabel(0, 0)
			a.resultsBody.Objects = []fyne.CanvasObject{
				buildResultsEmptyState(
					"Сканирование выполняется",
					"Устройства появятся здесь по мере ответа сети.",
					iconScan(),
					"",
					nil,
				),
			}
			a.resultsBody.Refresh()
			a.clearResultsMainSplitRef()
			return
		}
	}

	filtered := a.filteredSortedResults()
	a.lastRenderStats = resultsRenderStats{
		FilteredCount: len(filtered),
		VisibleCount:  len(filtered),
	}
	// R-09: счётчик строк в шапке — сколько показано из отфильтрованного.
	a.updateResultsCountLabel(len(filtered), len(filtered))
	if len(a.scanResults) == 0 {
		a.updateResultsCountLabel(0, 0)
		msg := "Результаты не найдены."
		icon := iconInfo()
		switch a.resultsState {
		case resultsStateStopped:
			msg = "Сканирование остановлено. Результаты могут быть неполными."
			icon = iconCancel()
		case resultsStateTimeout:
			msg = "Сканирование прервано по таймауту."
			icon = iconSecurity()
		}
		a.resultsBody.Objects = []fyne.CanvasObject{
			buildResultsEmptyState("Нет данных", msg, icon, "Повторить", func() {
				if a.scanCtrl != nil {
					a.scanCtrl.StartScan(a.scanResults)
				}
			}),
		}
		a.resultsBody.Refresh()
		a.clearResultsMainSplitRef()
		return
	}

	if len(filtered) == 0 {
		a.updateResultsCountLabel(0, 0)
		a.resultsBody.Objects = []fyne.CanvasObject{
			buildResultsEmptyState(
				"Ничего не найдено",
				"Нет устройств, подходящих под текущие фильтры.",
				iconSearch(),
				"Сбросить фильтры",
				func() {
					if a.resetFiltersBtn != nil {
						a.resetFiltersBtn.OnTapped()
					}
				},
			),
		}
		a.resultsBody.Refresh()
		a.clearResultsMainSplitRef()
		return
	}

	if a.resultsSubMode == resultsSubModeSecurity {
		a.resultsBody.Objects = []fyne.CanvasObject{a.buildSecurityDashboardView(filtered)}
		a.resultsBody.Refresh()
		a.clearResultsMainSplitRef()
		return
	}
	if a.resultsSubMode == resultsSubModeInventory {
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
		return widget.NewCard("Детали хоста", "", widget.NewLabel("Нет данных для отображения деталей."))
	}
	details := a.buildHostDetailsSections(r)
	cols := 2
	if a.currentLayoutProfile() == "compact" {
		cols = 1
	}
	actions := a.buildHostQuickActions(r, cols)
	return widget.NewCard("Детали хоста", "Выбранный хост и быстрые действия", container.NewVBox(details, actions))
}

// buildHostDetailsSections группирует детали хоста по секциям (P2-8).
func (a *App) buildHostDetailsSections(r scanner.Result) fyne.CanvasObject {
	ip := strings.TrimSpace(r.IP)
	if ip == "" {
		return widget.NewLabel("Нет данных")
	}

	// Секция «Идентификация»
	ident := container.NewVBox(
		widget.NewLabelWithStyle("Идентификация", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.buildDetailRow("Имя хоста", nullDash(r.Hostname)),
		a.buildDetailRow("Тип", nullDash(r.DeviceType)),
		a.buildDetailRow("Производитель", nullDash(r.DeviceVendor)),
	)

	// Секция «Сеть»
	net := container.NewVBox(
		widget.NewLabelWithStyle("Сеть", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.buildDetailRow("IP", nullDash(r.IP)),
		a.buildDetailRow("MAC", nullDash(r.MAC)),
		a.buildDetailRow("ОС (оценка)", osGuessLine(r)),
		a.buildDetailRow("SNMP", boolLabel(r.SNMPEnabled)),
	)

	// Секция «Порты»
	openCount := countOpenPorts(r.Ports)
	ports := container.NewVBox(
		widget.NewLabelWithStyle("Порты", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.buildDetailRow("Открытых", strconv.Itoa(openCount)),
	)
	if chips := a.buildPortChips(r); chips != nil {
		ports.Add(chips)
	}

	return container.NewVBox(
		container.NewPadded(ident),
		widget.NewSeparator(),
		container.NewPadded(net),
		widget.NewSeparator(),
		container.NewPadded(ports),
	)
}

// buildDetailRow формирует строку «метка: значение» с кнопкой копирования.
func (a *App) buildDetailRow(label, value string) fyne.CanvasObject {
	lbl := widget.NewLabel(label + ":")
	lbl.Truncation = fyne.TextTruncateEllipsis
	val := widget.NewLabel(value)
	val.Truncation = fyne.TextTruncateEllipsis
	val.Importance = widget.HighImportance
	copyBtn := widget.NewButtonWithIcon("", iconCopy(), func() {
		if a.myApp != nil {
			a.myApp.Clipboard().SetContent(value)
		}
	})
	copyBtn.Importance = widget.LowImportance
	return container.NewBorder(nil, nil, lbl, copyBtn, val)
}

// boolLabel возвращает локализованное булево значение.
func boolLabel(v bool) string {
	if v {
		return "да"
	}
	return "нет"
}

// Индексы вкладок mainTabs (N-06: именованные константы вместо жёстких чисел).
// Порядок задаётся в init_ui.go: mainTabs = AppTabs("Сканирование", "Топология", "Инструменты").
const (
	tabIndexScan     = 0 // resultsTabScanIndex (app_shortcuts.go)
	tabIndexTopology = 1 // topologyTabContent
	tabIndexTools    = 2 // toolsTabContent
)

func (a *App) buildHostQuickActions(r scanner.Result, cols int) *fyne.Container {
	if cols <= 0 {
		cols = 1
	}
	return container.NewGridWithColumns(cols,
		widget.NewButtonWithIcon("Ping", iconRefresh(), func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectIndex(tabIndexTools)
			a.runPingTool()
		}),
		widget.NewButtonWithIcon("Traceroute", iconRoute(), func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectIndex(tabIndexTools)
			a.runTracerouteTool()
		}),
		widget.NewButtonWithIcon("DNS", iconSearch(), func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectIndex(tabIndexTools)
			a.runDNSTool()
		}),
		widget.NewButtonWithIcon("Whois", iconAccount(), func() {
			if a.toolsHostEntry != nil {
				a.toolsHostEntry.SetText(strings.TrimSpace(r.IP))
			}
			a.mainTabs.SelectIndex(tabIndexTools)
			a.toolsCtrl.RunWhoisTool()
		}),
		widget.NewButtonWithIcon("Wake-on-LAN", iconScan(), func() {
			if a.toolsWOLMacEntry != nil {
				a.toolsWOLMacEntry.SetText(strings.TrimSpace(r.MAC))
			}
			a.mainTabs.SelectIndex(tabIndexTools)
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
