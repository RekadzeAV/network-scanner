package gui

import (
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// --- Декомпозиция buildResultsContainer (R10): монолитная функция разбита
// на функции-секции. Поведение и порядок инициализации сохранены. ---

// initResultsStateAndWidgets задаёт дефолты состояния результатов и создаёт
// заглушку тела результатов.
func (a *App) initResultsStateAndWidgets() {
	a.resultsMode = "Таблица"
	a.resultsSubMode = "Devices"
	a.resultsSort = "IP"
	a.maxPortChips = 24
	a.cardsVisibleCount = 200
	a.showRawBanners = false
	a.resultsState = resultsStateIdle
	a.resultsRenderDebounce = resultsRenderDebounceDefault
	a.resultsBody = container.NewStack(widget.NewLabel("Результаты сканирования появятся здесь после запуска."))
}

// initResultsModeAndInventoryWidgets создаёт переключатели режимов отображения
// и виджеты инвентаризации.
func (a *App) initResultsModeAndInventoryWidgets() {
	a.resultsModeSel = widget.NewRadioGroup([]string{"Таблица", "Карточки"}, func(value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		a.resultsMode = value
		a.saveResultsViewSettings()
		a.renderScanResultsView()
	})
	a.resultsModeSel.Horizontal = true
	a.resultsModeSel.SetSelected(a.resultsMode)
	a.resultsSubModeSel = widget.NewRadioGroup([]string{"Devices", "Security", "Inventory"}, func(value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		a.resultsSubMode = value
		a.saveResultsViewSettings()
		a.updateResultsFiltersVisibility()
		a.renderScanResultsView()
	})
	a.inventoryDBEntry = widget.NewEntry()
	a.inventoryDBEntry.SetText(filepath.Join("inventory", "network_inventory.db"))
	a.inventoryAutoSaveCheck = widget.NewCheck("Автосохранять снапшот после успешного сканирования", nil)
	a.inventoryAutoSaveCheck.SetChecked(true)
	a.inventoryScanASelect = widget.NewSelect([]string{}, func(string) {
		if a.resultsSubMode == "Inventory" {
			a.renderScanResultsView()
		}
	})
	a.inventoryScanASelect.PlaceHolder = "Snapshot A"
	a.inventoryScanBSelect = widget.NewSelect([]string{}, func(string) {
		if a.resultsSubMode == "Inventory" {
			a.renderScanResultsView()
		}
	})
	a.inventoryScanBSelect.PlaceHolder = "Snapshot B"
	a.inventoryStatusLabel = widget.NewLabel("Инвентаризация: выберите режим Inventory")
	a.inventoryStatusLabel.Wrapping = fyne.TextWrapWord
	a.inventoryRefreshBtn = widget.NewButtonWithIcon("Обновить список снапшотов", iconRefresh(), func() {
		a.refreshInventorySnapshots()
		if a.resultsSubMode == "Inventory" {
			a.renderScanResultsView()
		}
	})
	a.resultsSubModeSel.Horizontal = true
	a.resultsSubModeSel.SetSelected(a.resultsSubMode)
}

// initResultsFilterWidgets создаёт виджеты фильтрации и сортировки результатов.
func (a *App) initResultsFilterWidgets() {
	a.resultsSortSel = widget.NewSelect([]string{"IP", "Hostname", "Тип", "Открытые порты"}, func(string) {
		a.saveResultsViewSettings()
		a.scheduleResultsRender(false)
	})
	a.resultsSortSel.SetSelected(a.resultsSort)
	a.resultsFilterEnt = widget.NewEntry()
	a.resultsFilterEnt.SetPlaceHolder("Фильтр: HostName/IP/MAC/тип")
	a.resultsFilterEnt.OnChanged = func(value string) {
		a.resultsFilterQuery = strings.TrimSpace(value)
		a.saveResultsViewSettings()
		a.scheduleResultsRender(false)
	}
	a.clearFilterBtn = widget.NewButtonWithIcon("Очистить", iconClear(), func() {
		a.resultsFilterQuery = ""
		a.resultsFilterEnt.SetText("")
		if a.resultsCidrFilterEnt != nil {
			a.resultsCidrFilterEnt.SetText("")
		}
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Все")
		}
		a.resultsPortStateMode = "all"
		a.saveResultsViewSettings()
		a.scheduleResultsRender(true)
	})
	a.resultsCidrFilterEnt = widget.NewEntry()
	a.resultsCidrFilterEnt.SetPlaceHolder("CIDR фильтр (например 192.168.1.0/24)")
	a.resultsCidrFilterEnt.OnChanged = func(_ string) {
		a.saveResultsViewSettings()
		a.scheduleResultsRender(false)
	}
	a.resultsPortStateMode = "all"
	a.resultsPortStateSel = widget.NewSelect([]string{"Все", "Есть открытые", "Есть закрытые", "Есть фильтруемые"}, func(value string) {
		switch strings.TrimSpace(value) {
		case "Есть открытые":
			a.resultsPortStateMode = "has_open"
		case "Есть закрытые":
			a.resultsPortStateMode = "has_closed"
		case "Есть фильтруемые":
			a.resultsPortStateMode = "has_filtered"
		default:
			a.resultsPortStateMode = "all"
		}
		a.saveResultsViewSettings()
		a.scheduleResultsRender(false)
	})
	a.resultsPortStateSel.SetSelected("Все")
}

// initResultsPresetWidgets создаёт пресеты фильтров, быстрые чекбоксы типов
// и виджеты отображения (чипы портов, raw banner).
func (a *App) initResultsPresetWidgets() {
	a.filtersInfoLabel = widget.NewLabel("Активных фильтров: 0")
	a.filtersInfoLabel.Truncation = fyne.TextTruncateClip
	a.resultsPerfLabel = widget.NewLabel("Рендер: n/a")
	a.resultsPerfLabel.Truncation = fyne.TextTruncateClip
	a.filterPresetSel = widget.NewSelect([]string{"1", "2", "3"}, nil)
	a.filterPresetSel.SetSelected("1")
	a.saveFilterPresetBtn = widget.NewButtonWithIcon("Сохранить пресет", iconSave(), func() {
		a.saveFilterPreset(strings.TrimSpace(a.filterPresetSel.Selected))
	})
	a.applyFilterPresetBtn = widget.NewButtonWithIcon("Применить пресет", iconConfirm(), func() {
		a.applyFilterPreset(strings.TrimSpace(a.filterPresetSel.Selected))
	})
	a.quickTypeChecks = map[string]*widget.Check{}
	typeKeys := []string{"Network Device", "Computer", "Server", "Unknown"}
	a.quickTypeCheckRow = make([]fyne.CanvasObject, 0, len(typeKeys)+2)
	a.quickTypeCheckRow = append(a.quickTypeCheckRow, widget.NewLabel("Быстрые фильтры:"))
	for _, key := range typeKeys {
		label := key
		ch := widget.NewCheck(label, func(_ bool) {
			a.saveResultsViewSettings()
			a.scheduleResultsRender(false)
		})
		a.quickTypeChecks[key] = ch
		a.quickTypeCheckRow = append(a.quickTypeCheckRow, ch)
	}
	a.openPortsOnlyCheck = widget.NewCheck("Только с открытыми портами", func(v bool) {
		a.onlyWithOpenPorts = v
		a.saveResultsViewSettings()
		a.scheduleResultsRender(false)
	})
	a.quickTypeCheckRow = append(a.quickTypeCheckRow, a.openPortsOnlyCheck)
	a.resetFiltersBtn = widget.NewButtonWithIcon("Сбросить фильтры", iconClear(), func() {
		a.resultsFilterQuery = ""
		a.resultsFilterEnt.SetText("")
		a.onlyWithOpenPorts = false
		a.openPortsOnlyCheck.SetChecked(false)
		if a.resultsCidrFilterEnt != nil {
			a.resultsCidrFilterEnt.SetText("")
		}
		a.resultsPortStateMode = "all"
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Все")
		}
		for _, ch := range a.quickTypeChecks {
			ch.SetChecked(false)
		}
		a.saveResultsViewSettings()
		a.scheduleResultsRender(true)
	})
	a.chipLimitSel = widget.NewSelect([]string{"12", "24", "48"}, func(value string) {
		v, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || v <= 0 {
			return
		}
		a.maxPortChips = v
		a.saveResultsViewSettings()
		a.scheduleResultsRender(true)
	})
	a.chipLimitSel.SetSelected(strconv.Itoa(a.maxPortChips))
	a.showRawBannersCheck = widget.NewCheck("Показывать raw banner", func(v bool) {
		a.showRawBanners = v
		a.saveResultsViewSettings()
		a.scheduleResultsRender(true)
	})
}

// initResultsDiagnosticsWidgets создаёт блок диагностики последнего запуска.
func (a *App) initResultsDiagnosticsWidgets() {
	a.diagnosticsLabel = widget.NewLabel("Диагностика последнего запуска: n/a")
	a.diagnosticsLabel.Wrapping = fyne.TextWrapWord
	a.copyDiagnosticsBtn = widget.NewButtonWithIcon("Копировать диагностику", iconCopy(), nil)
	a.copyDiagnosticsBtn.Disable()
	a.saveDiagnosticsBtn = widget.NewButtonWithIcon("Сохранить диагностику", iconSave(), nil)
	a.saveDiagnosticsBtn.Disable()
}

// buildResultsLayout собирает итоговый layout результатов: заголовок,
// сворачиваемую секцию фильтров/режимов (R2) и тело результатов.
func (a *App) buildResultsLayout() *fyne.Container {
	a.resultsScroll = container.NewScroll(a.resultsBody)
	a.resultsScroll.SetMinSize(fyne.NewSize(0, 150))

	resultsLabel := widget.NewLabel("Результаты сканирования:")
	resultsLabel.TextStyle = fyne.TextStyle{Bold: true}
	separator := widget.NewSeparator()

	a.resultsDiagnosticsGrid = container.NewGridWithColumns(3, a.diagnosticsLabel, a.copyDiagnosticsBtn, a.saveDiagnosticsBtn)
	a.resultsSortGrid = container.NewGridWithColumns(
		5,
		widget.NewLabel("Сортировка:"), a.resultsSortSel,
		widget.NewLabel("Чипов портов:"), a.chipLimitSel,
		a.showRawBannersCheck,
	)
	a.resultsCidrGrid = container.NewGridWithColumns(
		4,
		widget.NewLabel("CIDR:"),
		a.resultsCidrFilterEnt,
		widget.NewLabel("Состояние портов:"),
		a.resultsPortStateSel,
	)
	a.resultsPresetGrid = container.NewGridWithColumns(
		4,
		widget.NewLabel("Пресет фильтров:"),
		a.filterPresetSel,
		a.saveFilterPresetBtn,
		a.applyFilterPresetBtn,
	)

	// Секция фильтров (Accordion, R2): режим отображения, сортировка, CIDR/
	// состояние портов, пресеты, быстрые фильтры, настройки Inventory.
	// Внутри «Фильтры и режимы отображения» остаётся быстрый текстовый фильтр
	// и кнопка сброса — они вынесены наружу, так как используются чаще всего.
	filtersSection := container.NewVBox(
		container.NewGridWithColumns(2, widget.NewLabel("Режим отображения:"), a.resultsModeSel),
		a.resultsSortGrid,
		a.resultsCidrGrid,
		a.resultsPresetGrid,
		container.NewHBox(a.quickTypeCheckRow...),
		container.NewGridWithColumns(2, widget.NewLabel("Inventory DB:"), a.inventoryDBEntry),
		container.NewGridWithColumns(2, a.inventoryAutoSaveCheck, a.inventoryRefreshBtn),
	)
	a.resultsFiltersAccordion = widget.NewAccordion(
		widget.NewAccordionItem("Фильтры и режимы отображения", filtersSection),
	)
	a.resultsFiltersAccordion.MultiOpen = false
	if len(a.resultsFiltersAccordion.Items) > 0 {
		a.resultsFiltersAccordion.Items[0].Open = a.resultsFiltersAccordionOpen
	}

	resultsContainer := container.NewBorder(
		container.NewVBox(
			separator,
			resultsLabel,
			a.resultsStateLabel,
			a.autoProfileHeaderLabel,
			a.resultsDiagnosticsGrid,
			container.NewGridWithColumns(2, widget.NewLabel("Подрежим:"), a.resultsSubModeSel),
			a.resultsFiltersAccordion,
			container.NewBorder(
				nil, nil,
				nil,
				container.NewHBox(a.clearFilterBtn, a.filtersInfoLabel, a.resultsPerfLabel),
				a.resultsFilterEnt,
			),
			container.NewHBox(a.resetFiltersBtn),
		),
		nil, nil, nil,
		a.resultsScroll,
	)

	return resultsContainer
}