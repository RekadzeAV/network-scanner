package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/devicecontrol"
)

// initUI инициализирует весь UI приложения.
func (a *App) initUI() {
	// Инициализация UI сканирования и результатов
	a.initScanUI()

	// Создание вкладки сканирования
	scanTabContent := a.buildScanTabContent()

	// Вкладка топологии
	a.topologyText = widget.NewRichText()
	a.topologyText.Wrapping = fyne.TextWrapWord
	a.topologyText.ParseMarkdown("## Топология сети\n\nСначала выполните сканирование, затем нажмите **Построить топологию**.")
	a.topologyScroll = container.NewScroll(a.topologyText)
	a.topologyScroll.SetMinSize(fyne.NewSize(0, 250))
	a.topologyGraphStatus = widget.NewLabel("Интерактивная карта: нет данных")
	a.topologyGraphStatus.Wrapping = fyne.TextWrapWord
	a.topologySearchEntry = widget.NewEntry()
	a.topologySearchEntry.SetPlaceHolder("Поиск узла (IP / hostname / MAC)")
	a.topologyTypeFilterSel = widget.NewSelect([]string{"all", "router", "switch", "host", "unknown"}, nil)
	a.topologyTypeFilterSel.SetSelected("all")
	a.topologyConfidenceFilterSel = widget.NewSelect([]string{"all", "high", "medium", "low"}, nil)
	a.topologyConfidenceFilterSel.SetSelected("all")
	a.topologyResetMapBtn = widget.NewButton("Сброс карты", nil)
	a.topologyGraphBox = container.NewWithoutLayout()
	a.topologyGraphBox.Resize(fyne.NewSize(1200, 800))
	a.topologyGraphScroll = container.NewScroll(a.topologyGraphBox)
	a.topologyGraphScroll.SetMinSize(fyne.NewSize(0, 240))
	a.topologyImage = canvas.NewImageFromResource(nil)
	a.topologyImage.FillMode = canvas.ImageFillContain
	a.topologyImage.SetMinSize(fyne.NewSize(0, 200))
	a.topologyImgBox = container.NewStack(a.topologyImage)
	a.topologyImgScroll = container.NewScroll(a.topologyImgBox)
	a.zoomSelect = widget.NewSelect([]string{"Fit", "100%", "150%", "200%"}, nil)
	a.zoomSelect.SetSelected("Fit")
	a.refreshPreviewBtn = widget.NewButton("Обновить превью", nil)
	a.refreshPreviewBtn.Disable()
	a.openPreviewBtn = widget.NewButton("Открыть PNG во внешнем окне", nil)
	a.openPreviewBtn.Disable()
	a.topologyStatus = widget.NewLabel("Топология не построена")
	a.topologyStatus.Wrapping = fyne.TextWrapWord
	a.snmpStageLabel = widget.NewLabel("")
	a.snmpStageLabel.Wrapping = fyne.TextWrapWord
	a.snmpStageLabel.Hide()
	a.snmpProgress = widget.NewProgressBar()
	a.snmpProgress.Hide()

	topologyControls := container.NewVBox(
		widget.NewLabel("SNMP community (через запятую):"),
		a.snmpCommEntry,
		widget.NewLabel("SNMP timeout (сек):"),
		a.snmpTimeoutEnt,
		container.NewHBox(a.buildTopoBtn, a.stopTopoBtn, a.saveTopoBtn),
		container.NewHBox(a.copyPerfBtn, a.savePerfBtn),
		container.NewHBox(widget.NewLabel("Масштаб превью:"), a.zoomSelect, a.refreshPreviewBtn),
		a.openPreviewBtn,
		widget.NewLabel("Интерактивная карта:"),
		a.topologySearchEntry,
		container.NewHBox(
			widget.NewLabel("Тип:"), a.topologyTypeFilterSel,
			widget.NewLabel("Confidence:"), a.topologyConfidenceFilterSel,
			a.topologyResetMapBtn,
		),
		a.topologyGraphStatus,
		a.snmpStageLabel,
		a.snmpProgress,
		a.topologyStatus,
	)
	a.topologyControlsScroll = container.NewVScroll(topologyControls)
	a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, 160))
	a.topologyMainSplit = container.NewVSplit(a.topologyGraphScroll, a.topologyScroll)
	a.topologyMainSplit.Offset = 0.55
	topologyTabContent := container.NewBorder(
		a.topologyControlsScroll,
		nil,
		nil,
		nil,
		a.topologyMainSplit,
	)
	a.toolsHostEntry = widget.NewEntry()
	a.toolsHostEntry.SetPlaceHolder("Хост или IP")
	a.toolsPingCountEnt = widget.NewEntry()
	a.toolsPingCountEnt.SetText("4")
	a.toolsTimeoutEnt = widget.NewEntry()
	a.toolsTimeoutEnt.SetText("60")
	a.toolsTraceHopsEnt = widget.NewEntry()
	a.toolsTraceHopsEnt.SetText("30")
	a.toolsDNSResolverEnt = widget.NewEntry()
	a.toolsDNSResolverEnt.SetPlaceHolder("DNS сервер (опционально, например 1.1.1.1:53)")
	a.toolsWOLMacEntry = widget.NewEntry()
	a.toolsWOLMacEntry.SetPlaceHolder("MAC для Wake-on-LAN (например aa:bb:cc:dd:ee:ff)")
	a.toolsWOLBcastEntry = widget.NewEntry()
	a.toolsWOLBcastEntry.SetPlaceHolder("Broadcast (опционально, например 192.168.1.255:9)")
	a.toolsWOLIfaceEntry = widget.NewEntry()
	a.toolsWOLIfaceEntry.SetPlaceHolder("Интерфейс (опц., если broadcast пуст)")
	a.toolsAuditMinSeveritySel = widget.NewSelect([]string{"all", "critical", "high", "medium", "low"}, func(_ string) {
		a.saveScanSettings()
	})
	a.toolsAuditMinSeveritySel.SetSelected("low")
	a.toolsDeviceTargetEntry = widget.NewEntry()
	a.toolsDeviceTargetEntry.SetPlaceHolder("Device target URL (например http://192.168.1.1)")
	a.toolsDeviceVendorEntry = widget.NewSelect(deviceControlVendors, func(_ string) {
		a.saveScanSettings()
	})
	a.toolsDeviceVendorEntry.SetSelected(devicecontrol.VendorGenericHTTP)
	a.toolsDeviceUserEntry = widget.NewEntry()
	a.toolsDeviceUserEntry.SetPlaceHolder("Username (опционально)")
	a.toolsDevicePassEntry = widget.NewPasswordEntry()
	a.toolsDevicePassEntry.SetPlaceHolder("Password (опционально)")
	a.toolsPingBtn = widget.NewButton("Ping", nil)
	a.toolsTraceBtn = widget.NewButton("Traceroute", nil)
	a.toolsDNSBtn = widget.NewButton("DNS", nil)
	a.toolsWhoisBtn = widget.NewButton("Whois", nil)
	a.toolsWiFiBtn = widget.NewButton("Wi-Fi", nil)
	a.toolsAuditBtn = widget.NewButton("Аудит портов", nil)
	a.toolsRiskBtn = widget.NewButton("Risk Signatures", nil)
	a.toolsWOLBtn = widget.NewButton("Wake-on-LAN", nil)
	a.toolsDeviceStatusBtn = widget.NewButton("Device Status", nil)
	a.toolsDeviceRebootBtn = widget.NewButton("Device Reboot", nil)
	a.toolsOutput = widget.NewRichText()
	a.toolsOutput.Wrapping = fyne.TextWrapWord
	a.toolsOutput.ParseMarkdown("Введите хост/IP и выберите инструмент.")
	a.operationsOutput = widget.NewRichText()
	a.operationsOutput.Wrapping = fyne.TextWrapWord
	a.operationsOutput.ParseMarkdown("### Operations Center\n\nИстория операций появится после запуска задач.")
	a.operationsSelectMap = make(map[string]string)
	a.operationsSelect = widget.NewSelect([]string{}, func(value string) {
		if a == nil {
			return
		}
		id := strings.TrimSpace(a.operationsSelectMap[strings.TrimSpace(value)])
		a.selectedOperationID = id
		a.refreshOperationActionsState()
	})
	a.operationsSelect.PlaceHolder = "Выберите операцию"
	a.operationsRetryBtn = widget.NewButton("Retry", func() {
		id := strings.TrimSpace(a.selectedOperationID)
		if id == "" || a.operations == nil {
			return
		}
		if _, ok := a.operations.Retry(id); !ok {
			if a.statusLabel != nil {
				a.statusLabel.SetText("Retry недоступен для выбранной операции")
			}
			return
		}
		if a.statusLabel != nil {
			a.statusLabel.SetText("Операция отправлена в retry")
		}
	})
	a.operationsCancelBtn = widget.NewButton("Cancel", func() {
		id := strings.TrimSpace(a.selectedOperationID)
		if id == "" || a.operations == nil {
			return
		}
		if !a.operations.Cancel(id) {
			if a.statusLabel != nil {
				a.statusLabel.SetText("Cancel недоступен для выбранной операции")
			}
			return
		}
		if a.statusLabel != nil {
			a.statusLabel.SetText("Операция отменена")
		}
	})
	a.operationsRetryBtn.Disable()
	a.operationsCancelBtn.Disable()
	a.toolsOutputScroll = container.NewScroll(a.toolsOutput)
	a.toolsOutputScroll.SetMinSize(fyne.NewSize(0, 280))
	a.operationsOutputScroll = container.NewScroll(a.operationsOutput)
	a.operationsOutputScroll.SetMinSize(fyne.NewSize(0, 120))
	a.toolButtonsGrid = container.NewGridWithColumns(
		5,
		a.toolsPingBtn,
		a.toolsTraceBtn,
		a.toolsDNSBtn,
		a.toolsWhoisBtn,
		a.toolsWiFiBtn,
		a.toolsWOLBtn,
		a.toolsAuditBtn,
		a.toolsRiskBtn,
		a.toolsDeviceStatusBtn,
		a.toolsDeviceRebootBtn,
	)
	a.toolsControlsScroll = container.NewVScroll(container.NewVBox(
		widget.NewLabel("Хост/IP:"),
		a.toolsHostEntry,
		a.toolsDNSResolverEnt,
		widget.NewLabel("Wake-on-LAN:"),
		a.toolsWOLMacEntry,
		a.toolsWOLBcastEntry,
		a.toolsWOLIfaceEntry,
		widget.NewLabel("Device Control (HTTP API):"),
		a.toolsDeviceTargetEntry,
		a.toolsDeviceVendorEntry,
		widget.NewLabel("Профили: generic-http -> /api/{status|reboot}; tp-link-http -> /api/system/{status|reboot}."),
		container.NewGridWithColumns(2, a.toolsDeviceUserEntry, a.toolsDevicePassEntry),
		container.NewGridWithColumns(
			2,
			widget.NewLabel("Audit min severity:"),
			a.toolsAuditMinSeveritySel,
		),
		container.NewGridWithColumns(
			2,
			widget.NewLabel("Ping пакетов:"),
			a.toolsPingCountEnt,
			widget.NewLabel("Timeout (сек):"),
			a.toolsTimeoutEnt,
			widget.NewLabel("Traceroute hops:"),
			a.toolsTraceHopsEnt,
		),
		a.toolButtonsGrid,
	))
	a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, 200))
	a.operationsHeaderGrid = container.New(layout.NewGridLayoutWithColumns(2),
		widget.NewLabel("Operations:"),
		a.operationsSelect,
		a.operationsRetryBtn,
		a.operationsCancelBtn,
	)
	toolsUpper := container.NewVBox(
		a.toolsControlsScroll,
		container.NewVBox(
			a.operationsHeaderGrid,
			a.operationsOutputScroll,
		),
	)
	a.toolsTabMainSplit = container.NewVSplit(toolsUpper, a.toolsOutputScroll)
	a.toolsTabMainSplit.Offset = 0.40
	toolsTabContent := a.toolsTabMainSplit

	a.mainTabs = container.NewAppTabs(
		container.NewTabItem("Сканирование", scanTabContent),
		container.NewTabItem("Топология", topologyTabContent),
		container.NewTabItem("Инструменты", toolsTabContent),
	)
	a.mainTabs.OnSelected = func(item *container.TabItem) {
		if item != nil && item.Text == "Сканирование" {
			a.renderScanResultsView()
		}
	}

	// Тулбар с основными действиями
	a.mainToolbar = container.NewHBox(
		widget.NewSeparator(),
		widget.NewButton("▶ Сканирование", func() {
			if a.scanCtrl != nil {
				a.scanCtrl.StartScan(a.scanResults)
			}
		}),
		widget.NewButton("⏹ Стоп", func() {
			if a.scanCtrl != nil {
				a.scanCtrl.StopScan()
			}
		}),
		widget.NewSeparator(),
		widget.NewButton("💾 Сохранить", func() {
			a.saveResults()
		}),
		widget.NewButton("🗺 Топология", func() {
			if a.topoCtrl != nil {
				a.topoCtrl.BuildTopology(a.scanResults, a.myWindow)
			}
		}),
		widget.NewSeparator(),
		widget.NewButton("↺ Сброс UI", func() {
			if a.settingsMgr != nil {
				a.settingsMgr.ResetUIPanelLayoutWithFeedback(a.scanTabMainSplit, a.topologyMainSplit, a.toolsTabMainSplit, a.myWindow)
			}
		}),
	)
	a.mainToolbar.Show()

	// Обёртка: тулбар + табы
	mainContent := container.NewBorder(
		a.mainToolbar, // top
		nil,           // bottom
		nil,           // leading
		nil,           // trailing
		a.mainTabs,    // center
	)

	a.myWindow.SetContent(mainContent)
	a.setPortRangeControlsEnabled(a.scanTCPPortsCheck.Checked)
	a.refreshAutoProfileStateLabel()
	a.startResultsLayoutWatcher()
	a.startOperationsWatcher()
}
