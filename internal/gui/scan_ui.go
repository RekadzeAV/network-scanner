package gui

import (
	"fmt"
	"image/color"
	"network-scanner/internal/logger"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// resultsRenderDebounceDefault — задержка дебаунса перерисовки результатов:
// сглаживает поток обновлений прогресса, не давая GUI перерисовываться на
// каждое событие сканера.
const resultsRenderDebounceDefault = 180 * time.Millisecond

// initScanUI инициализирует UI сканирования
func (a *App) initScanUI() {
	logger.LogDebug("[initScanUI] Начало")
	defer func() {
		if rec := recover(); rec != nil {
			logger.LogError(fmt.Errorf("PANIC в initScanUI: %v", rec), "initScanUI")
		}
	}()
	// Дефолт debounce рендера результатов задаётся сразу: он нужен даже если
	// buildResultsContainer ещё не вызывался (тесты, headless-режим).
	a.resultsRenderDebounce = resultsRenderDebounceDefault
	// Поле ввода сети
	logger.LogDebug("[initScanUI] Создаю networkEntry")
	networkLabel := widget.NewLabel("Сеть (CIDR, например 192.168.1.0/24):")
	networkLabel.Wrapping = fyne.TextWrapWord
	a.networkEntry = widget.NewEntry()
	a.networkEntry.SetPlaceHolder("Оставьте пустым для автоматического определения")
	a.networkEntry.OnChanged = func(v string) {
		a.showFieldErrorDelayed(a.networkEntryMsg, validateCIDRField(v))
		a.saveScanSettings()
	}
	a.networkEntryBox, a.networkEntryMsg = makeValidatedEntry(a.networkEntry)
	logger.LogDebug("[initScanUI] Создаю portRangeEntry")
	a.portRangeEntry = widget.NewEntry()
	a.portRangeEntry.SetPlaceHolder("1-65535")
	a.portRangeEntry.SetText("1-65535")
	a.portRangeEntry.OnChanged = func(v string) {
		a.showFieldErrorDelayed(a.portRangeEntryMsg, validatePortRangeField(v))
		a.saveScanSettings()
	}
	a.portRangeEntryBox, a.portRangeEntryMsg = makeValidatedEntry(a.portRangeEntry)
	logger.LogDebug("[initScanUI] Создаю scanTCPPortsCheck")
	a.scanTCPPortsCheck = widget.NewCheck("Сканировать TCP порты", func(v bool) {
		if a != nil {
			logger.LogDebug("[scanTCPPortsCheck callback] v=%v", v)
			a.setPortRangeControlsEnabled(v)
			a.saveScanSettings()
		}
	})
	logger.LogDebug("[initScanUI] Создаю portWellKnownBtn")
	a.portWellKnownBtn = widget.NewButton("Системные (Well-Known): 0–1023", nil)
	logger.LogDebug("[initScanUI] Создаю portRegisteredBtn")
	a.portRegisteredBtn = widget.NewButtonWithIcon("Зарегистрированные: 1024–49151", iconDocument(), nil)
	logger.LogDebug("[initScanUI] Создаю portDynamicBtn")
	a.portDynamicBtn = widget.NewButtonWithIcon("Динамические / частные: 49152–65535", iconMore(), nil)
	a.timeoutEntry = widget.NewEntry()
	a.timeoutEntry.SetText("2")
	a.timeoutEntry.OnChanged = func(v string) {
		a.showFieldErrorDelayed(a.timeoutEntryMsg, validateIntField(v, 1, 60))
		a.saveScanSettings()
	}
	a.timeoutEntryBox, a.timeoutEntryMsg = makeValidatedEntry(a.timeoutEntry)
	a.threadsEntry = widget.NewEntry()
	a.threadsEntry.SetText("50")
	a.threadsEntry.OnChanged = func(v string) {
		a.showFieldErrorDelayed(a.threadsEntryMsg, validateIntField(v, 1, 512))
		a.saveScanSettings()
	}
	a.threadsEntryBox, a.threadsEntryMsg = makeValidatedEntry(a.threadsEntry)
	a.scanUDPCheck = widget.NewCheck("Включить UDP сканирование", nil)
	a.scanBannersCheck = widget.NewCheck("Собирать баннеры/версии служб (медленнее)", nil)
	a.scanOSActiveCheck = widget.NewCheck("Активные эвристики определения ОС (может замедлить)", nil)
	a.scanVerboseLogsCheck = widget.NewCheck("Детальные логи по портам (debug, шумно)", nil)
	a.scanVerboseInfoBtn = widget.NewButtonWithIcon("Подробнее", iconHelp(), nil)
	a.autoProfileCheck = widget.NewCheck("Автопрофиль сканирования (рекомендуется)", nil)
	a.autoProfileCheck.SetChecked(true)
	a.autoProfileInfoBtn = widget.NewButtonWithIcon("Почему изменены параметры?", iconInfo(), nil)
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
	logger.LogDebug("[initScanUI] SetChecked true на scanTCPPortsCheck")
	logger.LogDebug("[initScanUI] Создаю presetQuickBtn")
	a.presetQuickBtn = widget.NewButtonWithIcon("Быстро", iconFast(), nil)
	logger.LogDebug("[initScanUI] Создаю presetBalBtn")
	a.presetBalBtn = widget.NewButtonWithIcon("Баланс", iconScan(), nil)
	logger.LogDebug("[initScanUI] Создаю presetDeepBtn")
	a.presetDeepBtn = widget.NewButtonWithIcon("Глубоко", iconDeep(), nil)
	a.recommendedProfileBtn = widget.NewButtonWithIcon("Рекомендуемые настройки", iconConfirm(), nil)
	a.recommendedProfileInfoBtn = widget.NewButtonWithIcon("Почему?", iconHelp(), nil)
	a.recommendedProfileBadge = canvas.NewText("Профиль: не выбран", themeColorDisabled())
	a.recommendedProfileBadge.TextSize = 12

	// Кнопка сканирования
	a.scanButton = widget.NewButtonWithIcon("Запустить сканирование", iconScan(), nil)
	a.scanButton.Importance = widget.HighImportance
	a.stopButton = widget.NewButtonWithIcon("Стоп сканирование", iconStop(), nil)
	a.stopButton.Disable()

	// Кнопка сохранения
	a.saveButton = widget.NewButtonWithIcon("Сохранить результаты", iconSave(), nil)
	a.saveButton.Disable()

	// Поля SNMP/топологии
	a.snmpCommEntry = widget.NewEntry()
	a.snmpCommEntry.SetText("public")
	a.snmpTimeoutEnt = widget.NewEntry()
	a.snmpTimeoutEnt.SetText("2")
	a.buildTopoBtn = widget.NewButtonWithIcon("Построить топологию", iconTopology(), nil)
	a.buildTopoBtn.Disable()
	a.stopTopoBtn = widget.NewButtonWithIcon("Стоп топологию", iconStop(), nil)
	a.stopTopoBtn.Disable()
	a.saveTopoBtn = widget.NewButtonWithIcon("Сохранить топологию", iconSave(), nil)
	a.saveTopoBtn.Disable()
	a.copyPerfBtn = widget.NewButtonWithIcon("Копировать отчёт", iconCopy(), nil)
	a.copyPerfBtn.Disable()
	a.savePerfBtn = widget.NewButtonWithIcon("Сохранить отчёт", iconSave(), nil)
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
	a.copyDiagnosticsBtn = widget.NewButtonWithIcon("Копировать диагностику", iconCopy(), nil)
	a.copyDiagnosticsBtn.Disable()
	a.saveDiagnosticsBtn = widget.NewButtonWithIcon("Сохранить диагностику", iconSave(), nil)
	a.saveDiagnosticsBtn.Disable()

	// Метка этапа сканирования
	a.stageLabel = widget.NewLabel("")
	a.stageLabel.Wrapping = fyne.TextWrapWord
	a.stageLabel.Hide()

	// Прогресс-бар
	a.progressBar = widget.NewProgressBar()
	a.progressBar.Hide()
}

// buildScanControlsContainer создаёт контейнер с настройками сканирования.
// Панель сгруппирована: «Основное» (сеть/порты/запуск) всегда видимо,
// остальные параметры — в сворачиваемых секциях (widget.Accordion).
func (a *App) buildScanControlsContainer() *container.Scroll {
	defer func() {
		if rec := recover(); rec != nil {
			logger.LogError(fmt.Errorf("PANIC в buildScanControlsContainer: %v", rec), "buildScanControlsContainer")
		}
	}()
	portClassHint := widget.NewLabel("Системные (Well-Known) 0–1023 — резерв под известные и системные службы; для части портов нужны права администратора. Примеры: 21 FTP, 22 SSH, 25 SMTP, 53 DNS, 80 HTTP, 443 HTTPS. " +
		"Зарегистрированные (Registered) 1024–49151 — назначения IANA для приложений (например 1433 MSSQL, 3306 MySQL, 8080 HTTP-alt). " +
		"Динамические/частные (Dynamic/Private) 49152–65535 — эфемерные и частные порты.")
	portClassHint.Wrapping = fyne.TextWrapWord

	// --- Секция «Основное»: всегда видима ---
	mainSection := container.NewVBox(
		widget.NewLabel("Сеть (CIDR, например 192.168.1.0/24):"),
		a.networkEntryBox,
		a.scanTCPPortsCheck,
		widget.NewLabel("Диапазон TCP портов (например 1-65535 или 80,443):"),
		a.portRangeEntryBox,
		portClassHint,
		container.NewGridWithColumns(
			3,
			a.portWellKnownBtn,
			a.portRegisteredBtn,
			a.portDynamicBtn,
		),
		container.NewGridWithColumns(3, a.scanButton, a.stopButton, a.saveButton),
		a.statusLabel,
		a.stageLabel,
		a.progressBar,
	)

	// --- Секция «Пресеты и профиль» ---
	profilesSection := container.NewVBox(
		widget.NewLabel("Пресет:"),
		container.NewGridWithColumns(3, a.presetQuickBtn, a.presetBalBtn, a.presetDeepBtn),
		container.NewGridWithColumns(2, a.recommendedProfileBtn, a.recommendedProfileInfoBtn),
		a.recommendedProfileBadge,
		container.NewGridWithColumns(2, a.autoProfileCheck, a.autoProfileInfoBtn),
		a.autoProfileStateText,
		a.autoProfileHint,
	)

	// --- Секция «Производительность и опции» ---
	optionsSection := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("Таймаут (сек):"), a.timeoutEntryBox,
			widget.NewLabel("Потоки:"), a.threadsEntryBox,
		),
		a.scanUDPCheck,
		a.scanBannersCheck,
		a.scanOSActiveCheck,
		container.NewGridWithColumns(2, a.scanVerboseLogsCheck, a.scanVerboseInfoBtn),
	)

	a.scanAdvancedAccordion = widget.NewAccordion(
		widget.NewAccordionItem("Пресеты и профиль сканирования", profilesSection),
		widget.NewAccordionItem("Производительность и опции", optionsSection),
	)
	// По умолчанию все секции свёрнуты — видна только «Основное».
	for i := range a.scanAdvancedAccordion.Items {
		a.scanAdvancedAccordion.Items[i].Open = false
	}
	a.scanAdvancedAccordion.MultiOpen = true
	a.scanAdvancedOpen = false

	scanControlsContainer := container.NewVBox(
		mainSection,
		a.scanAdvancedAccordion,
	)

	return container.NewVScroll(scanControlsContainer)
}

// buildResultsContainer создаёт контейнер с результатами сканирования.
// Декомпозирован (R10): виджеты создаются в results_ui_sections.go по секциям,
// здесь — только их вызов (порядок инициализации сохранён).
func (a *App) buildResultsContainer() *fyne.Container {
	a.initResultsStateAndWidgets()
	a.initResultsModeAndInventoryWidgets()
	a.initResultsFilterWidgets()
	a.initResultsPresetWidgets()
	a.initResultsDiagnosticsWidgets()
	return a.buildResultsLayout()
}

// buildScanTabContent создаёт содержимое вкладки сканирования
func (a *App) buildScanTabContent() fyne.CanvasObject {
	logger.LogDebug("[buildScanTabContent] Начало, defer установлен")
	logger.LogDebug("[buildScanTabContent] Вызываю buildScanControlsContainer")
	scanControlsScroll := a.buildScanControlsContainer()
	logger.LogDebug("[buildScanTabContent] buildScanControlsContainer завершена")
	logger.LogDebug("[buildScanTabContent] Вызываю buildResultsContainer")
	resultsContainer := a.buildResultsContainer()
	logger.LogDebug("[buildScanTabContent] buildResultsContainer завершена")
	logger.LogDebug("[buildScanTabContent] Создаю scanTabMainSplit")
	// Вкладка сканирования: верх/низ с перетаскиваемой границей
	a.scanTabMainSplit = container.NewVSplit(scanControlsScroll, resultsContainer)
	logger.LogDebug("[buildScanTabContent] scanTabMainSplit создан")
	a.scanTabMainSplit.Offset = 0.35
	logger.LogDebug("[buildScanTabContent] offset установлен")

	return a.scanTabMainSplit
}
