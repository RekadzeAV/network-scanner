package gui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

func (a *App) renderScanResultsView() {
	if a.resultsBody == nil {
		return
	}
	switch a.resultsState {
	case resultsStateScanning:
		a.resultsStateLabel.SetText("Сканирование выполняется...")
		a.resultsBody.Objects = []fyne.CanvasObject{
			widget.NewCard("Сканирование", "", widget.NewLabel("Идет сбор данных, пожалуйста подождите.")),
		}
		a.resultsBody.Refresh()
		return
	case resultsStateStopped:
		a.resultsStateLabel.SetText("Сканирование остановлено пользователем")
	case resultsStateTimeout:
		a.resultsStateLabel.SetText("Сканирование завершилось по таймауту")
	case resultsStateDone:
		a.resultsStateLabel.SetText(fmt.Sprintf("Результаты получены: %d устройств", len(a.scanResults)))
	default:
		a.resultsStateLabel.SetText("Результаты еще не получены")
	}

	if len(a.scanResults) == 0 {
		a.resultsBody.Objects = []fyne.CanvasObject{
			widget.NewCard("Результаты сканирования", "", widget.NewLabel("Результаты сканирования не найдены.")),
		}
		a.resultsBody.Refresh()
		return
	}
	selectedTypes := a.selectedQuickTypes()
	if a.filtersInfoLabel != nil {
		activeFilters := len(selectedTypes)
		if strings.TrimSpace(a.resultsFilterQuery) != "" {
			activeFilters++
		}
		if a.onlyWithOpenPorts {
			activeFilters++
		}
		a.filtersInfoLabel.SetText(fmt.Sprintf("Активных фильтров: %d", activeFilters))
	}
	results := a.currentDisplayedResults()
	if len(results) == 0 {
		a.resultsBody.Objects = []fyne.CanvasObject{
			widget.NewCard("Результаты сканирования", "", widget.NewLabel("По текущему фильтру ничего не найдено.")),
		}
		a.resultsBody.Refresh()
		return
	}
	if a.resultsMode == "Карточки" {
		a.resultsBody.Objects = []fyne.CanvasObject{a.buildCardsResultsView(results)}
	} else {
		a.resultsBody.Objects = []fyne.CanvasObject{a.buildTableResultsView(results)}
	}
	a.resultsBody.Refresh()
}

func (a *App) selectedQuickTypes() []string {
	selectedTypes := make([]string, 0)
	for typeName, ch := range a.quickTypeChecks {
		if ch != nil && ch.Checked {
			selectedTypes = append(selectedTypes, typeName)
		}
	}
	return selectedTypes
}

func (a *App) currentDisplayedResults() []scanner.Result {
	selectedTypes := a.selectedQuickTypes()
	filtered := filterResultsForDisplayAdvanced(a.scanResults, a.resultsFilterQuery, selectedTypes, a.onlyWithOpenPorts)
	return sortedResultsForDisplayWithMode(filtered, a.resultsSort)
}

func (a *App) startResultsLayoutWatcher() {
	if a == nil || a.myWindow == nil {
		return
	}
	a.lastCanvasSize = a.myWindow.Canvas().Size()
	go func() {
		ticker := time.NewTicker(350 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if a.myWindow == nil {
				return
			}
			size := a.myWindow.Canvas().Size()
			if size == a.lastCanvasSize {
				continue
			}
			a.lastCanvasSize = size
			if a.resultsMode != "Карточки" || len(a.scanResults) == 0 {
				continue
			}
			fyne.Do(func() {
				a.renderScanResultsView()
				a.resultsScroll.Refresh()
			})
		}
	}()
}

func (a *App) buildTableResultsView(results []scanner.Result) fyne.CanvasObject {
	const (
		colHost  = float32(220)
		colIP    = float32(170)
		colMAC   = float32(190)
		colPorts = float32(560)
	)

	makeCell := func(text string, width float32, bold bool) fyne.CanvasObject {
		label := widget.NewLabel(text)
		label.Wrapping = fyne.TextTruncate
		label.TextStyle = fyne.TextStyle{Bold: bold}
		bgColor := tableRowBgColor
		if bold {
			bgColor = tableHeaderBgColor
		}
		bg := canvas.NewRectangle(bgColor)
		content := container.NewPadded(label)
		cell := container.NewStack(bg, content)
		return container.NewGridWrap(fyne.NewSize(width, 44), cell)
	}

	makePortCell := func(ports []scanner.PortInfo, width float32, bold bool) fyne.CanvasObject {
		if bold {
			return makeCell("Порты", width, true)
		}
		portLabels := openPortLabels(ports, a.maxPortChips)
		if len(portLabels) == 0 {
			return makeCell("-", width, false)
		}
		rows := a.makeChipRows(portLabels, width-24)
		bg := canvas.NewRectangle(color.RGBA{R: 247, G: 248, B: 250, A: 255})
		box := container.NewStack(bg, container.NewPadded(container.NewVBox(rows...)))
		minH := float32(48 + (len(rows)-1)*30)
		return container.NewGridWrap(fyne.NewSize(width, minH), box)
	}

	header := container.NewHBox(
		makeCell("HostName", colHost, true),
		makeCell("IP", colIP, true),
		makeCell("MAC", colMAC, true),
		makePortCell(nil, colPorts, true),
	)

	rows := []fyne.CanvasObject{header}
	for _, r := range results {
		hostname := formatDeviceValue(r.Hostname)
		ip := formatDeviceValue(r.IP)
		mac := formatDeviceValue(r.MAC)
		rows = append(rows, container.NewHBox(
			makeCell(hostname, colHost, false),
			makeCell(ip, colIP, false),
			makeCell(mac, colMAC, false),
			makePortCell(r.Ports, colPorts, false),
		))
	}

	tableWidth := colHost + colIP + colMAC + colPorts + 12
	tableBody := container.NewVBox(rows...)
	tableFixed := container.NewGridWrap(fyne.NewSize(tableWidth, tableBody.MinSize().Height), tableBody)
	tableScroll := container.NewScroll(tableFixed)
	tableScroll.SetMinSize(fyne.NewSize(0, 340))

	protocols, types := collectAnalytics(results)
	analyticsCols := 2
	if a.myWindow != nil && a.myWindow.Canvas().Size().Width < 900 {
		analyticsCols = 1
	}
	analytics := container.NewGridWithColumns(analyticsCols,
		a.buildSimpleStatsTable("Протоколы", protocols),
		a.buildSimpleStatsTable("Типы устройств", normalizeDeviceTypes(types)),
	)
	return container.NewVBox(
		widget.NewCard("Текстовый режим: Таблица", "", tableScroll),
		widget.NewLabel("Аналитика"),
		analytics,
	)
}

func (a *App) buildCardsResultsView(results []scanner.Result) fyne.CanvasObject {
	width := a.myWindow.Canvas().Size().Width
	cols := 3
	if width < 1200 {
		cols = 2
	}
	if width < 768 {
		cols = 1
	}
	cardWidth := float32(320)
	if cols == 2 {
		cardWidth = 360
	}
	if cols == 1 {
		cardWidth = width - 72
		if cardWidth < 280 {
			cardWidth = 280
		}
	}
	cardHeight := float32(220)

	cards := make([]fyne.CanvasObject, 0, len(results))
	for _, r := range results {
		hostname := formatDeviceValue(r.Hostname)
		ip := formatDeviceValue(r.IP)
		mac := formatDeviceValue(r.MAC)
		portLabels := openPortLabels(r.Ports, a.maxPortChips)
		rows := a.makeChipRows(portLabels, 300)
		if len(rows) == 0 {
			rows = []fyne.CanvasObject{widget.NewLabel("-")}
		}
		body := container.NewVBox(
			widget.NewLabelWithStyle(hostname, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("IP: "+ip),
			widget.NewLabel("MAC: "+mac),
			widget.NewSeparator(),
			container.NewVBox(rows...),
		)
		card := widget.NewCard("", "", body)
		cards = append(cards, container.NewGridWrap(fyne.NewSize(cardWidth, cardHeight), card))
	}

	protocols, types := collectAnalytics(results)
	charts := container.NewGridWithColumns(2,
		a.buildPieChart("Протоколы", protocols),
		a.buildPieChart("Типы устройств", normalizeDeviceTypes(types)),
	)
	if width < 768 {
		charts = container.NewGridWithColumns(1,
			a.buildPieChart("Протоколы", protocols),
			a.buildPieChart("Типы устройств", normalizeDeviceTypes(types)),
		)
	}

	return container.NewVBox(
		widget.NewCard("Текстовый режим: Карточки", "", container.NewGridWithColumns(cols, cards...)),
		widget.NewLabel("Аналитика"),
		charts,
	)
}

func (a *App) makeChipRows(chips []string, maxWidth float32) []fyne.CanvasObject {
	if len(chips) == 0 {
		return nil
	}
	rows := make([]fyne.CanvasObject, 0)
	currentRow := make([]fyne.CanvasObject, 0)
	currentWidth := float32(0)
	for _, chipText := range chips {
		approxWidth := estimateChipWidth(chipText)
		if currentWidth > 0 && currentWidth+approxWidth > maxWidth {
			rows = append(rows, container.NewHBox(currentRow...))
			currentRow = make([]fyne.CanvasObject, 0)
			currentWidth = 0
		}
		chip := makeDecorativeChip(chipText, approxWidth)
		currentRow = append(currentRow, chip)
		currentWidth += approxWidth + 8
	}
	if len(currentRow) > 0 {
		rows = append(rows, container.NewHBox(currentRow...))
	}
	return rows
}

func (a *App) buildSimpleStatsTable(title string, data map[string]int) fyne.CanvasObject {
	if len(data) == 0 {
		return widget.NewCard(title, "", widget.NewLabel("Нет данных"))
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := make([]fyne.CanvasObject, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, container.NewGridWithColumns(2, widget.NewLabel(k), widget.NewLabel(fmt.Sprintf("%d", data[k]))))
	}
	return widget.NewCard(title, "", container.NewVBox(rows...))
}

func estimateChipWidth(text string) float32 {
	w := float32(len([]rune(text))*7 + 28)
	if w < 72 {
		return 72
	}
	if w > 260 {
		return 260
	}
	return w
}

func makeDecorativeChip(text string, width float32) fyne.CanvasObject {
	bg := canvas.NewRectangle(chipBgColor)
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextTruncate
	label.TextStyle = fyne.TextStyle{Bold: true}
	chip := container.NewStack(bg, container.NewPadded(label))
	return container.NewGridWrap(fyne.NewSize(width, 30), chip)
}
