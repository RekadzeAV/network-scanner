package gui

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// initScanUI инициализирует UI сканирования
func (a *App) initScanUI() {
	// Поле ввода сети
	networkLabel := widget.NewLabel("Сеть (CIDR, например 192.168.1.0/24):")
	networkLabel.Wrapping = fyne.TextWrapWord
	a.networkEntry = widget.NewEntry()
	a.networkEntry.SetPlaceHolder("Оставьте пустым для автоматического определения")
	a.portRangeEntry = widget.NewEntry()
	a.portRangeEntry.SetPlaceHolder("1-65535")
	a.portRangeEntry.SetText("1-65535")
	a.scanTCPPortsCheck = widget.NewCheck("Сканировать TCP порты", func(v bool) {
		a.setPortRangeControlsEnabled(v)
		a.saveScanSettings()
	})
	a.portWellKnownBtn = widget.NewButton("Системные (Well-Known): 0–1023", nil)
	a.portRegisteredBtn = widget.NewButton("Зарегистрированные: 1024–49151", nil)
	a.portDynamicBtn = widget.NewButton("Динамические / частные: 49152–65535", nil)
	a.timeoutEntry = widget.NewEntry()
	a.timeoutEntry.SetText("2")
	a.threadsEntry = widget.NewEntry()
	a.threadsEntry.SetText("50")
	a.scanUDPCheck = widget.NewCheck("Включить UDP сканирование", nil)
	a.scanBannersCheck = widget.NewCheck("Собирать баннеры/версии служб (медленнее)", nil)
	a.scanOSActiveCheck = widget.NewCheck("Активные эвристики определения ОС (может замедлить)", nil)
	a.scanVerboseLogsCheck = widget.NewCheck("Детальные логи по портам (debug, шумно)", nil)
	a.scanVerboseInfoBtn = widget.NewButton("Подробнее", nil)
	a.autoProfileCheck = widget.NewCheck("Автопрофиль сканирования (рекомендуется)", nil)
	a.autoProfileCheck.SetChecked(true)
	a.autoProfileInfoBtn = widget.NewButton("Почему изменены параметры?", nil)
	a.autoProfileStateText = canvas.NewText("", color.RGBA{R: 60, G: 170, B: 80, A: 255})
	a.autoProfileStateText.TextSize = 13
	a.autoProfileHint = widget.NewLabel(fmt.Sprintf(
		"Для больших подсетей автопрофиль ограничивает нагрузку: от ~%d хостов снижает threads/диапазон портов (пороги: %d/%d/%d хостов).",
		autoProfileHostWarn,
		autoProfileHostLarge,
		autoProfileHostXLarge,
		autoProfileHostXXLarge,
	))
	a.autoProfileHint.Wrapping = fyne.TextWrapWord
	// SetChecked после создания полей, которые читает saveScanSettings (колбэк срабатывает сразу).
	a.scanTCPPortsCheck.SetChecked(true)
	a.presetQuickBtn = widget.NewButton("Быстро", nil)
	a.presetBalBtn = widget.NewButton("Баланс", nil)
	a.presetDeepBtn = widget.NewButton("Глубоко", nil)
	a.recommendedProfileBtn = widget.NewButton("Рекомендуемые настройки", nil)
	a.recommendedProfileInfoBtn = widget.NewButton("Почему?", nil)
	a.recommendedProfileBadge = canvas.NewText("Профиль: не выбран", color.RGBA{R: 110, G: 110, B: 110, A: 255})
	a.recommendedProfileBadge.TextSize = 12

	// Кнопка сканирования
	a.scanButton = widget.NewButton("Запустить сканирование", nil)
	a.scanButton.Importance = widget.HighImportance
	a.stopButton = widget.NewButton("Стоп сканирование", nil)
	a.stopButton.Disable()

	// Кнопка сохранения
	a.saveButton = widget.NewButton("Сохранить результаты", nil)
	a.saveButton.Disable()

	// Поля SNMP/топологии
	a.snmpCommEntry = widget.NewEntry()
	a.snmpCommEntry.SetText("public")
	a.snmpTimeoutEnt = widget.NewEntry()
	a.snmpTimeoutEnt.SetText("2")
	a.buildTopoBtn = widget.NewButton("Построить топологию", nil)
	a.buildTopoBtn.Disable()
	a.stopTopoBtn = widget.NewButton("Стоп топологию", nil)
	a.stopTopoBtn.Disable()
	a.saveTopoBtn = widget.NewButton("Сохранить топологию", nil)
	a.saveTopoBtn.Disable()
	a.copyPerfBtn = widget.NewButton("Копировать отчет производительности", nil)
	a.copyPerfBtn.Disable()
	a.savePerfBtn = widget.NewButton("Сохранить отчет производительности", nil)
	a.savePerfBtn.Disable()

	// Статус
	a.statusLabel = widget.NewLabel("Готов к сканированию")
	a.statusLabel.Wrapping = fyne.TextWrapWord
	a.resultsStateLabel = widget.NewLabel("Результаты еще не получены")
	a.resultsStateLabel.Wrapping = fyne.TextWrapWord
	a.autoProfileHeaderLabel = widget.NewLabel("")
	a.autoProfileHeaderLabel.Wrapping = fyne.TextWrapWord
	a.diagnosticsLabel = widget.NewLabel("Диагностика последнего запуска: n/a")
	a.diagnosticsLabel.Wrapping = fyne.TextWrapWord
	a.copyDiagnosticsBtn = widget.NewButton("Копировать диагностику", nil)
	a.copyDiagnosticsBtn.Disable()
	a.saveDiagnosticsBtn = widget.NewButton("Сохранить диагностику", nil)
	a.saveDiagnosticsBtn.Disable()

	// Метка этапа сканирования
	a.stageLabel = widget.NewLabel("")
	a.stageLabel.Wrapping = fyne.TextWrapWord
	a.stageLabel.Hide()

	// Прогресс-бар
	a.progressBar = widget.NewProgressBar()
	a.progressBar.Hide()
}

// buildScanControlsContainer создаёт контейнер с настройками сканирования
func (a *App) buildScanControlsContainer() *container.Scroll {
	portClassHint := widget.NewLabel("Системные (Well-Known) 0–1023 — резерв под известные и системные службы; для части портов нужны права администратора. Примеры: 21 FTP, 22 SSH, 25 SMTP, 53 DNS, 80 HTTP, 443 HTTPS. " +
		"Зарегистрированные (Registered) 1024–49151 — назначения IANA для приложений (например 1433 MSSQL, 3306 MySQL, 8080 HTTP-alt). " +
		"Динамические/частные (Dynamic/Private) 49152–65535 — эфемерные и частные порты.")
	portClassHint.Wrapping = fyne.TextWrapWord

	scanControlsContainer := container.NewVBox(
		widget.NewLabel("Сеть (CIDR, например 192.168.1.0/24):"),
		a.networkEntry,
		a.scanTCPPortsCheck,
		widget.NewLabel("Диапазон TCP портов (например 1-65535 или 80,443):"),
		a.portRangeEntry,
		portClassHint,
		container.NewVBox(
			a.portWellKnownBtn,
			a.portRegisteredBtn,
			a.portDynamicBtn,
		),
		widget.NewLabel("Пресет:"),
		container.NewGridWithColumns(
			3,
			a.presetQuickBtn,
			a.presetBalBtn,
			a.presetDeepBtn,
		),
		widget.NewLabel("Онбординг:"),
		container.NewGridWithColumns(
			2,
			a.recommendedProfileBtn,
			a.recommendedProfileInfoBtn,
		),
		a.recommendedProfileBadge,
		container.NewGridWithColumns(
			2,
			widget.NewLabel("Таймаут (сек):"),
			a.timeoutEntry,
			widget.NewLabel("Потоки:"),
			a.threadsEntry,
		),
		a.scanUDPCheck,
		a.scanBannersCheck,
		a.scanOSActiveCheck,
		container.NewGridWithColumns(2, a.scanVerboseLogsCheck, a.scanVerboseInfoBtn),
		container.NewGridWithColumns(2, a.autoProfileCheck, a.autoProfileInfoBtn),
		a.autoProfileStateText,
		a.autoProfileHint,
		container.NewGridWithColumns(3, a.scanButton, a.stopButton, a.saveButton),
		a.statusLabel,
		a.stageLabel,
		a.progressBar,
	)

	return container.NewVScroll(scanControlsContainer)
}

// buildResultsContainer создаёт контейнер с результатами сканирования
func (a *App) buildResultsContainer() *fyne.Container {
	// Область результатов с прокруткой
	a.resultsMode = "Таблица"
	a.resultsSubMode = "Devices"
	a.resultsSort = "IP"
	a.maxPortChips = 24
	a.cardsVisibleCount = 200
	a.showRawBanners = false
	a.resultsState = resultsStateIdle
	a.resultsRenderDebounce = 180 * time.Millisecond
	a.resultsBody = container.NewMax(widget.NewLabel("Результаты сканирования появятся здесь после запуска."))
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
	a.inventoryRefreshBtn = widget.NewButton("Обновить список снапшотов", func() {
		a.refreshInventorySnapshots()
		if a.resultsSubMode == "Inventory" {
			a.renderScanResultsView()
		}
	})
	a.resultsSubModeSel.Horizontal = true
	a.resultsSubModeSel.SetSelected(a.resultsSubMode)
	a.resultsSortSel = widget.NewSelect([]string{"IP", "HostName"}, func(value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		a.resultsSort = value
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
	a.clearFilterBtn = widget.NewButton("Очистить", func() {
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
	a.filtersInfoLabel = widget.NewLabel("Активных фильтров: 0")
	a.filtersInfoLabel.Wrapping = fyne.TextTruncate
	a.resultsPerfLabel = widget.NewLabel("Рендер: n/a")
	a.resultsPerfLabel.Wrapping = fyne.TextTruncate
	a.filterPresetSel = widget.NewSelect([]string{"1", "2", "3"}, nil)
	a.filterPresetSel.SetSelected("1")
	a.saveFilterPresetBtn = widget.NewButton("Сохранить пресет", func() {
		a.saveFilterPreset(strings.TrimSpace(a.filterPresetSel.Selected))
	})
	a.applyFilterPresetBtn = widget.NewButton("Применить пресет", func() {
		a.applyFilterPreset(strings.TrimSpace(a.filterPresetSel.Selected))
	})
	a.quickTypeChecks = map[string]*widget.Check{}
	typeKeys := []string{"Network Device", "Computer", "Server", "Unknown"}
	typeCheckRow := make([]fyne.CanvasObject, 0, len(typeKeys)+2)
	typeCheckRow = append(typeCheckRow, widget.NewLabel("Быстрые фильтры:"))
	for _, key := range typeKeys {
		label := key
		ch := widget.NewCheck(label, func(_ bool) {
			a.saveResultsViewSettings()
			a.scheduleResultsRender(false)
		})
		a.quickTypeChecks[key] = ch
		typeCheckRow = append(typeCheckRow, ch)
	}
	a.openPortsOnlyCheck = widget.NewCheck("Только с открытыми портами", func(v bool) {
		a.onlyWithOpenPorts = v
		a.saveResultsViewSettings()
		a.scheduleResultsRender(false)
	})
	typeCheckRow = append(typeCheckRow, a.openPortsOnlyCheck)
	a.resetFiltersBtn = widget.NewButton("Сбросить фильтры", func() {
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

	// Создаем прокручиваемый контейнер для результатов
	a.resultsScroll = container.NewScroll(a.resultsBody)
	a.resultsScroll.SetMinSize(fyne.NewSize(0, 150))

	// Заголовок результатов
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

	resultsContainer := container.NewBorder(
		container.NewVBox(
			separator,
			resultsLabel,
			a.resultsStateLabel,
			a.autoProfileHeaderLabel,
			a.resultsDiagnosticsGrid,
			container.NewGridWithColumns(2, widget.NewLabel("Подрежим:"), a.resultsSubModeSel),
			container.NewGridWithColumns(2, widget.NewLabel("Inventory DB:"), a.inventoryDBEntry),
			container.NewGridWithColumns(2, a.inventoryAutoSaveCheck, a.inventoryRefreshBtn),
			container.NewGridWithColumns(2, widget.NewLabel("Режим отображения:"), a.resultsModeSel),
			a.resultsSortGrid,
			container.NewBorder(
				nil, nil,
				nil,
				container.NewHBox(a.clearFilterBtn, a.filtersInfoLabel, a.resultsPerfLabel),
				a.resultsFilterEnt,
			),
			a.resultsCidrGrid,
			a.resultsPresetGrid,
			container.NewHBox(typeCheckRow...),
			container.NewHBox(a.resetFiltersBtn),
		),
		nil, nil, nil,
		a.resultsScroll,
	)

	return resultsContainer
}

// buildScanTabContent создаёт содержимое вкладки сканирования
func (a *App) buildScanTabContent() fyne.CanvasObject {
	scanControlsScroll := a.buildScanControlsContainer()
	scanControlsScroll.SetMinSize(fyne.NewSize(0, 180))

	resultsContainer := a.buildResultsContainer()

	// Вкладка сканирования: верх/низ с перетаскиваемой границей
	a.scanTabMainSplit = container.NewVSplit(scanControlsScroll, resultsContainer)
	a.scanTabMainSplit.Offset = 0.35

	return a.scanTabMainSplit
}
