package gui

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/audit"
	"network-scanner/internal/builder"
	"network-scanner/internal/devicecontrol"
	"network-scanner/internal/display"
	"network-scanner/internal/gui/controller"
	"network-scanner/internal/inventory"
	"network-scanner/internal/logger"
	"network-scanner/internal/nettools"
	"network-scanner/internal/network"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
	scand "network-scanner/internal/scanner/daemon"
	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"
	"network-scanner/internal/wol"
)

// scanUpdate содержит результаты сканирования для обновления UI
type scanUpdate struct {
	results     []scanner.Result
	diagnostics string
}

// progressUpdate содержит информацию о прогрессе сканирования
type progressUpdate struct {
	stage   string
	current int
	total   int
	message string
	percent float64
}

type topologyBuildMetrics struct {
	snmpDuration  time.Duration
	buildDuration time.Duration
	totalDuration time.Duration
}

// App представляет GUI приложение
type App struct {
	myApp                       fyne.App
	myWindow                    fyne.Window
	scanResults                 []scanner.Result
	scanRunner                  *scand.Runner
	networkEntry                *widget.Entry
	portRangeEntry              *widget.Entry
	timeoutEntry                *widget.Entry
	threadsEntry                *widget.Entry
	scanUDPCheck                *widget.Check
	scanBannersCheck            *widget.Check
	scanOSActiveCheck           *widget.Check
	scanVerboseLogsCheck        *widget.Check
	scanVerboseInfoBtn          *widget.Button
	autoProfileCheck            *widget.Check
	autoProfileInfoBtn          *widget.Button
	autoProfileHint             *widget.Label
	autoProfileStateText        *canvas.Text
	presetQuickBtn              *widget.Button
	presetBalBtn                *widget.Button
	presetDeepBtn               *widget.Button
	recommendedProfileBtn       *widget.Button
	recommendedProfileInfoBtn   *widget.Button
	recommendedProfileBadge     *canvas.Text
	scanTCPPortsCheck           *widget.Check
	portWellKnownBtn            *widget.Button
	portRegisteredBtn           *widget.Button
	portDynamicBtn              *widget.Button
	statusLabel                 *widget.Label
	resultsStateLabel           *widget.Label
	autoProfileHeaderLabel      *widget.Label
	diagnosticsLabel            *widget.Label
	copyDiagnosticsBtn          *widget.Button
	saveDiagnosticsBtn          *widget.Button
	stageLabel                  *widget.Label
	progressBar                 *widget.ProgressBar
	scanControlsScroll          *container.Scroll
	scanTabMainSplit            *container.Split
	scanTabSplitInitialized     bool
	scanTabSplitPersistPrimed   bool
	lastPersistedScanSplit      float64
	resultsScroll               *container.Scroll
	resultsBody                 *fyne.Container
	resultsMode                 string
	resultsModeSel              *widget.RadioGroup
	resultsSubMode              string
	resultsSubModeSel           *widget.RadioGroup
	inventoryDBEntry            *widget.Entry
	inventoryAutoSaveCheck      *widget.Check
	inventoryScanASelect        *widget.Select
	inventoryScanBSelect        *widget.Select
	inventoryRefreshBtn         *widget.Button
	inventoryStatusLabel        *widget.Label
	inventorySnapshots          []inventory.Snapshot
	resultsSort                 string
	resultsSortSel              *widget.Select
	resultsFilterEnt            *widget.Entry
	resultsFilterQuery          string
	resultsCidrFilterEnt        *widget.Entry
	resultsPortStateSel         *widget.Select
	resultsPortStateMode        string
	filtersInfoLabel            *widget.Label
	resultsPerfLabel            *widget.Label
	resultsDiagnosticsGrid      *fyne.Container
	resultsSortGrid             *fyne.Container
	resultsCidrGrid             *fyne.Container
	resultsPresetGrid           *fyne.Container
	clearFilterBtn              *widget.Button
	filterPresetSel             *widget.Select
	saveFilterPresetBtn         *widget.Button
	applyFilterPresetBtn        *widget.Button
	chipLimitSel                *widget.Select
	showRawBannersCheck         *widget.Check
	maxPortChips                int
	showRawBanners              bool
	onlyWithOpenPorts           bool
	openPortsOnlyCheck          *widget.Check
	quickTypeChecks             map[string]*widget.Check
	resetFiltersBtn             *widget.Button
	scanButton                  *widget.Button
	stopButton                  *widget.Button
	saveButton                  *widget.Button
	buildTopoBtn                *widget.Button
	stopTopoBtn                 *widget.Button
	saveTopoBtn                 *widget.Button
	copyPerfBtn                 *widget.Button
	savePerfBtn                 *widget.Button
	snmpCommEntry               *widget.Entry
	snmpTimeoutEnt              *widget.Entry
	lastTopology                *topology.Topology
	lastSNMPReport              *snmpcollector.CollectReport
	lastTopoMetric              topologyBuildMetrics
	topologyViewState           topologyMapState
	topologyText                *widget.RichText
	topologyControlsScroll      *container.Scroll
	topologyScroll              *container.Scroll
	topologySearchEntry         *widget.Entry
	topologyTypeFilterSel       *widget.Select
	topologyConfidenceFilterSel *widget.Select
	topologyResetMapBtn         *widget.Button
	topologyGraphStatus         *widget.Label
	topologyStatus              *widget.Label
	snmpStageLabel              *widget.Label
	snmpProgress                *widget.ProgressBar
	mainTabs                    *container.AppTabs
	topologyImage               *canvas.Image
	topologyImgBox              *fyne.Container
	topologyImgScroll           *container.Scroll
	topologyGraphBox            *fyne.Container
	topologyGraphScroll         *container.Scroll
	topologyMainSplit           *container.Split
	topologySplitInitialized    bool
	topologySplitPersistPrimed  bool
	lastPersistedTopologySplit  float64
	previewPath                 string
	refreshPreviewBtn           *widget.Button
	zoomSelect                  *widget.Select
	openPreviewBtn              *widget.Button
	topologyCancel              context.CancelFunc
	toolsHostEntry              *widget.Entry
	toolsPingCountEnt           *widget.Entry
	toolsTimeoutEnt             *widget.Entry
	toolsTraceHopsEnt           *widget.Entry
	toolsDNSResolverEnt         *widget.Entry
	toolsControlsScroll         *container.Scroll
	toolsTabMainSplit           *container.Split
	toolsSplitInitialized       bool
	toolsSplitPersistPrimed     bool
	lastPersistedToolsSplit     float64
	toolsOutputScroll           *container.Scroll
	operationsOutputScroll      *container.Scroll
	toolButtonsGrid             *fyne.Container
	operationsHeaderGrid        *fyne.Container
	toolsOutput                 *widget.RichText
	toolsPingBtn                *widget.Button
	toolsTraceBtn               *widget.Button
	toolsDNSBtn                 *widget.Button
	toolsWhoisBtn               *widget.Button
	toolsWiFiBtn                *widget.Button
	toolsAuditBtn               *widget.Button
	toolsAuditMinSeveritySel    *widget.Select
	toolsRiskBtn                *widget.Button
	toolsWOLMacEntry            *widget.Entry
	toolsWOLBcastEntry          *widget.Entry
	toolsWOLIfaceEntry          *widget.Entry
	toolsWOLBtn                 *widget.Button
	toolsDeviceTargetEntry      *widget.Entry
	toolsDeviceVendorEntry      *widget.Select
	toolsDeviceUserEntry        *widget.Entry
	toolsDevicePassEntry        *widget.Entry
	toolsDeviceStatusBtn        *widget.Button
	toolsDeviceRebootBtn        *widget.Button
	operationsOutput            *widget.RichText
	operationsSelect            *widget.Select
	operationsSelectMap         map[string]string
	selectedOperationID         string
	operationsRetryBtn          *widget.Button
	operationsCancelBtn         *widget.Button
	operationsHistory           []Operation
	confirmLargeScanBypass      bool
	selectedHostIP              string
	resultsState                string
	resultsMainSplit            *container.Split
	lastHostDetailsSplitKind    string // "V" (compact) или "H" — ориентация split с Host Details
	rememberedHostDetailsSplitV float64
	rememberedHostDetailsSplitH float64
	hostDetailsSplitPrimedV     bool
	hostDetailsSplitPrimedH     bool
	lastPersistedHostDetailsV   float64
	lastPersistedHostDetailsH   float64
	lastCanvasSize              fyne.Size
	lastCanvasScale             float32
	layoutProfile               string
	pieChartCache               map[string]fyne.Resource
	resultsRenderDebounce       time.Duration
	resultsRenderTimerMu        sync.Mutex
	resultsRenderTimer          *time.Timer
	cardsVisibleCount           int
	lastRenderStats             resultsRenderStats
	analyticsCacheKey           string
	analyticsCacheView          fyne.CanvasObject
	hostDetailsCacheMu          sync.RWMutex
	hostDetailsCache            map[string]string
	scanResultsVersion          uint64
	resultsPipelineCacheMu      sync.RWMutex
	resultsPipelineCacheKey     string
	resultsPipelineCacheData    []scanner.Result
	operations                  *OperationsManager
	services                    *AppServices
	mainToolbar                 *fyne.Container

	// Controllers (H2 Refactoring)
	scanCtrl    *controller.ScanController
	resultsCtrl *controller.ResultsController
	topoCtrl    *controller.TopologyController
	toolsCtrl   *controller.ToolsController
	settingsMgr *controller.SettingsManager

	// Mobile Support (L3)
	mobileLayout   *MobileLayout
	touchGestures  *TouchGestures
	isMobileDevice bool
}

const (
	prefNetwork                 = "scan.network"
	prefPortRange               = "scan.port_range"
	prefTimeout                 = "scan.timeout_sec"
	prefThreads                 = "scan.threads"
	prefScanUDP                 = "scan.udp"
	prefScanBanners             = "scan.grab_banners"
	prefScanOSActive            = "scan.os_detect_active"
	prefScanVerbosePortLogs     = "scan.verbose_port_logs"
	prefScanTCPPorts            = "scan.scan_tcp_ports"
	prefAutoProfile             = "scan.auto_profile"
	prefPreset                  = "scan.preset"
	prefRecommendedBadge        = "scan.recommended_badge"
	prefRecommendedBadgeClass   = "scan.recommended_badge_class"
	prefInventoryDBPath         = "scan.inventory.db_path"
	prefInventoryAutoSave       = "scan.inventory.auto_save"
	prefViewMode                = "scan.results_view_mode"
	prefResultsSubMode          = "scan.results_submode"
	prefSortMode                = "scan.results_sort_mode"
	prefChipLimit               = "scan.results_chip_limit"
	prefShowRawBanners          = "scan.results_show_raw_banners"
	prefFilterQuery             = "scan.results_filter_query"
	prefOnlyOpenPorts           = "scan.results_only_open_ports"
	prefTypeFilters             = "scan.results_type_filters"
	prefCidrFilter              = "scan.results_cidr_filter"
	prefPortStateMode           = "scan.results_port_state_mode"
	prefFilterPreset1           = "scan.results_filter_preset_1"
	prefFilterPreset2           = "scan.results_filter_preset_2"
	prefFilterPreset3           = "scan.results_filter_preset_3"
	prefToolHost                = "scan.tools.host"
	prefToolPingCount           = "scan.tools.ping_count"
	prefToolTimeout             = "scan.tools.timeout_sec"
	prefToolTraceHops           = "scan.tools.trace_hops"
	prefToolResolver            = "scan.tools.dns_resolver"
	prefToolAuditMinSeverity    = "scan.tools.audit_min_severity"
	prefToolDeviceTarget        = "scan.tools.device_target"
	prefToolDeviceVendor        = "scan.tools.device_vendor"
	prefToolDeviceUser          = "scan.tools.device_user"
	prefScanTabSplitOffset      = "scan.ui.scan_tab_split_offset"
	prefTopologyMainSplitOffset = "scan.ui.topology_main_split_offset"
	prefToolsTabSplitOffset     = "scan.ui.tools_tab_split_offset"
	prefHostDetailsSplitOffsetV = "scan.ui.host_details_split_offset_v"
	prefHostDetailsSplitOffsetH = "scan.ui.host_details_split_offset_h"
	maxScanThreadsGUI           = 512
	largeSubnetWarnHostGUI      = 512
	autoProfileHostWarn         = 256
	autoProfileHostLarge        = 512
	autoProfileHostXLarge       = 1024
	autoProfileHostXXLarge      = 2048
	minWindowWidth              = 1024

	layoutResetInfoMessage = "Положение разделителей между панелями (вкладки Сканирование, Топология, Инструменты) и split «результаты / Host Details» восстановлено по умолчанию."
)

var deviceControlVendors = []string{
	devicecontrol.VendorGenericHTTP,
	devicecontrol.VendorTPLINKHTTP,
}

const (
	resultsStateIdle     = "idle"
	resultsStateScanning = "scanning"
	resultsStateDone     = "done"
	resultsStateStopped  = "stopped"
	resultsStateTimeout  = "timeout"
)

var (
	chipBgColor        = color.RGBA{R: 222, G: 234, B: 255, A: 255}
	tableRowBgColor    = color.RGBA{R: 250, G: 251, B: 253, A: 255}
	tableHeaderBgColor = color.RGBA{R: 229, G: 236, B: 247, A: 255}
	piePalette         = []color.RGBA{
		{R: 37, G: 99, B: 235, A: 255},
		{R: 59, G: 130, B: 246, A: 255},
		{R: 14, G: 165, B: 233, A: 255},
		{R: 99, G: 102, B: 241, A: 255},
		{R: 168, G: 85, B: 247, A: 255},
		{R: 71, G: 85, B: 105, A: 255},
	}
)

// createAppIcon создает иконку приложения
func createAppIcon() fyne.Resource {
	// Создаем простое изображение иконки программно
	// Иконка 64x64 пикселя с простым дизайном сетевого сканера
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	// Рисуем простую иконку: синий фон с белым символом сети
	bgColor := color.RGBA{R: 0, G: 100, B: 200, A: 255}
	iconColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Рисуем простой символ сети (круг с линиями)
	centerX, centerY := 32.0, 32.0
	radius := 20.0

	// Рисуем круг
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist >= radius-2 && dist <= radius+2 {
				img.Set(x, y, iconColor)
			}
		}
	}

	// Рисуем линии от центра
	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi * 2 / 8
		for r := radius + 3; r < 30; r++ {
			x := int(centerX + r*math.Cos(angle))
			y := int(centerY + r*math.Sin(angle))
			if x >= 0 && x < 64 && y >= 0 && y < 64 {
				img.Set(x, y, iconColor)
			}
		}
	}

	// Конвертируем изображение в PNG
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		// Если не удалось создать иконку, возвращаем nil
		return nil
	}

	return fyne.NewStaticResource("icon.png", buf.Bytes())
}

// NewApp создает новый экземпляр GUI приложения
func NewApp() *App {
	// Явно устанавливаем scale для корректного отображения на Windows
	// Это исправляет проблему смещения курсора и наложения элементов
	// ДОЛЖНО быть вызвано ДО создания app.NewWithID
	os.Setenv("FYNE_SCALE", "1")

	myApp := app.NewWithID("network-scanner")

	myWindow := myApp.NewWindow("Network Scanner - Сканер локальной сети")

	// Устанавливаем иконку приложения
	if icon := createAppIcon(); icon != nil {
		myApp.SetIcon(icon)
		myWindow.SetIcon(icon)
	}

	// Адаптивный стартовый размер: 85% от размера экрана
	screenSize := myWindow.Canvas().Size()
	width := float32(int(screenSize.Width * 0.85))
	height := float32(int(screenSize.Height * 0.85))
	if width < 1024 {
		width = 1024
	}
	if height < 600 {
		height = 600
	}

	myWindow.Resize(fyne.NewSize(width, height))
	myWindow.CenterOnScreen()

	// Устанавливаем максимальный размер окна, чтобы оно не выходило за границы экрана
	// Fyne автоматически ограничит размер окна размером экрана
	myWindow.SetFixedSize(false) // Позволяем изменять размер, но в пределах экрана

	app := &App{
		myApp:            myApp,
		myWindow:         myWindow,
		layoutProfile:    "normal",
		pieChartCache:    make(map[string]fyne.Resource),
		hostDetailsCache: make(map[string]string),
		operations:       NewOperationsManager(),
	}

	// Инициализация сервисов через DI контейнер
	container := builder.NewContainer(builder.Config{
		LogLevel: "info",
	})
	app.services = NewAppServices(container)

	// Инициализация контроллеров (H2 Refactoring)
	app.scanCtrl = controller.NewScanController(myApp, &controller.ScanUI{
		NetworkEntry:         app.networkEntry,
		PortRangeEntry:       app.portRangeEntry,
		TimeoutEntry:         app.timeoutEntry,
		ThreadsEntry:         app.threadsEntry,
		ScanUDPCheck:         app.scanUDPCheck,
		ScanBannersCheck:     app.scanBannersCheck,
		ScanOSActiveCheck:    app.scanOSActiveCheck,
		ScanVerboseLogsCheck: app.scanVerboseLogsCheck,
		ScanTCPPortsCheck:    app.scanTCPPortsCheck,
		AutoProfileCheck:     app.autoProfileCheck,
		StatusLabel:          app.statusLabel,
		RecommendedBadge:     app.recommendedProfileBadge,
		PresetQuickBtn:       app.presetQuickBtn,
		PresetBalBtn:         app.presetBalBtn,
		PresetDeepBtn:        app.presetDeepBtn,
		RecommendedBtn:       app.recommendedProfileBtn,
		ScanButton:           app.scanButton,
		StopButton:           app.stopButton,
		StageLabel:           app.stageLabel,
		ProgressBar:          app.progressBar,
		ResultsStateLabel:    app.resultsStateLabel,
		CopyDiagnosticsBtn:   app.copyDiagnosticsBtn,
		SaveDiagnosticsBtn:   app.saveDiagnosticsBtn,
		MainToolbar:          app.mainToolbar,
		Window:               app.myWindow,
	}, app)
	app.resultsCtrl = controller.NewResultsController(myApp, &controller.ResultsUI{
		ResultsModeSel:       app.resultsModeSel,
		ResultsSubModeSel:    app.resultsSubModeSel,
		ResultsSortSel:       app.resultsSortSel,
		ResultsFilterEnt:     app.resultsFilterEnt,
		ResultsCidrFilterEnt: app.resultsCidrFilterEnt,
		ResultsPortStateSel:  app.resultsPortStateSel,
		ChipLimitSel:         app.chipLimitSel,
		ShowRawBannersCheck:  app.showRawBannersCheck,
		OpenPortsOnlyCheck:   app.openPortsOnlyCheck,
		QuickTypeChecks:      app.quickTypeChecks,
		StatusLabel:          app.statusLabel,
	})
	app.topoCtrl = controller.NewTopologyController(myApp, &controller.TopologyUI{
		SNMPCommEntry:     app.snmpCommEntry,
		SNMPTimeoutEnt:    app.snmpTimeoutEnt,
		BuildTopoBtn:      app.buildTopoBtn,
		StopTopoBtn:       app.stopTopoBtn,
		SaveTopoBtn:       app.saveTopoBtn,
		CopyPerfBtn:       app.copyPerfBtn,
		SavePerfBtn:       app.savePerfBtn,
		TopoText:          app.topologyText,
		TopoStatus:        app.topologyStatus,
		SNMPStageLabel:    app.snmpStageLabel,
		SNMPProgress:      app.snmpProgress,
		TopoSearchEntry:   app.topologySearchEntry,
		TopoTypeFilterSel: app.topologyTypeFilterSel,
		TopoConfFilterSel: app.topologyConfidenceFilterSel,
		TopoResetMapBtn:   app.topologyResetMapBtn,
		TopoGraphStatus:   app.topologyGraphStatus,
		TopoImage:         app.topologyImage,
		ZoomSelect:        app.zoomSelect,
		RefreshPreviewBtn: app.refreshPreviewBtn,
		OpenPreviewBtn:    app.openPreviewBtn,
	})
	app.toolsCtrl = controller.NewToolsController(myApp, &controller.ToolsUI{
		HostEntry:           app.toolsHostEntry,
		PingCountEnt:        app.toolsPingCountEnt,
		TimeoutEnt:          app.toolsTimeoutEnt,
		TraceHopsEnt:        app.toolsTraceHopsEnt,
		DNSResolverEnt:      app.toolsDNSResolverEnt,
		WOLMacEntry:         app.toolsWOLMacEntry,
		WOLBcastEntry:       app.toolsWOLBcastEntry,
		WOLIfaceEntry:       app.toolsWOLIfaceEntry,
		DeviceTargetEntry:   app.toolsDeviceTargetEntry,
		DeviceVendorEntry:   app.toolsDeviceVendorEntry,
		DeviceUserEntry:     app.toolsDeviceUserEntry,
		DevicePassEntry:     app.toolsDevicePassEntry,
		AuditMinSeveritySel: app.toolsAuditMinSeveritySel,
		PingBtn:             app.toolsPingBtn,
		TraceBtn:            app.toolsTraceBtn,
		DNSBtn:              app.toolsDNSBtn,
		WhoisBtn:            app.toolsWhoisBtn,
		WiFiBtn:             app.toolsWiFiBtn,
		AuditBtn:            app.toolsAuditBtn,
		RiskBtn:             app.toolsRiskBtn,
		WOLBtn:              app.toolsWOLBtn,
		DeviceStatusBtn:     app.toolsDeviceStatusBtn,
		DeviceRebootBtn:     app.toolsDeviceRebootBtn,
		ToolsOutput:         app.toolsOutput,
		OperationsOutput:    app.operationsOutput,
		OperationsSelect:    app.operationsSelect,
		OperationsRetryBtn:  app.operationsRetryBtn,
		OperationsCancelBtn: app.operationsCancelBtn,
		StatusLabel:         app.statusLabel,
	})
	app.settingsMgr = controller.NewSettingsManager(myApp)

	app.initUI()
	app.setupEventHandlers()
	app.loadScanSettings()
	app.autoDetectNetwork()

	logger.Log("GUI приложение инициализировано")

	return app
}

// initUI инициализирует все элементы интерфейса
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
	a.topologyImgBox = container.NewMax(a.topologyImage)
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

	// Тулбар с основными действиями — виден всегда, включая fullscreen режим.
	// На Windows Fyne рендерит MainMenu в системной заголовочной панели,
	// которая скрывается при переходе в fullscreen. Используем горизонтальный контейнер.
	a.mainToolbar = container.NewHBox(
		widget.NewSeparator(),
		widget.NewButton("▶ Сканирование", func() {
			a.scanCtrl.StartScan(a.scanResults)
		}),
		widget.NewButton("⏹ Стоп", func() {
			a.scanCtrl.StopScan()
		}),
		widget.NewSeparator(),
		widget.NewButton("💾 Сохранить", func() {
			a.saveResults()
		}),
		widget.NewButton("🗺 Топология", func() {
			a.topoCtrl.BuildTopology(a.scanResults, a.myWindow)
		}),
		widget.NewSeparator(),
		widget.NewButton("↺ Сброс UI", func() {
			a.settingsMgr.ResetUIPanelLayoutWithFeedback(a.scanTabMainSplit, a.topologyMainSplit, a.toolsTabMainSplit, a.myWindow)
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

// setupEventHandlers настраивает обработчики событий
func (a *App) setupEventHandlers() {
	a.scanButton.OnTapped = func() {
		a.scanCtrl.StartScan(a.scanResults)
	}
	a.stopButton.OnTapped = func() {
		a.scanCtrl.StopScan()
	}
	a.presetQuickBtn.OnTapped = func() {
		a.scanCtrl.ApplyPreset("quick")
	}
	a.presetBalBtn.OnTapped = func() {
		a.scanCtrl.ApplyPreset("balanced")
	}
	a.presetDeepBtn.OnTapped = func() {
		a.scanCtrl.ApplyPreset("deep")
	}
	if a.recommendedProfileBtn != nil {
		a.recommendedProfileBtn.OnTapped = func() {
			a.scanCtrl.ApplyRecommendedProfile(a.networkEntry.Text)
		}
	}
	if a.recommendedProfileInfoBtn != nil {
		a.recommendedProfileInfoBtn.OnTapped = func() {
			dialog.ShowInformation(
				"Логика рекомендованного профиля",
				"Кнопка подбирает безопасные параметры под размер подсети:\n\n"+
					"- небольшие сети: чуть глубже диапазон портов;\n"+
					"- средние/крупные: диапазон и параллелизм умеренные;\n"+
					"- очень крупные: минимально нагружающий профиль.\n\n"+
					"Во всех случаях для стабильности отключаются: UDP, баннеры,\n"+
					"active-эвристики ОС и детальные портовые логи.",
				a.myWindow,
			)
		}
	}
	a.networkEntry.OnChanged = func(_ string) {
		a.saveScanSettings()
	}
	a.portRangeEntry.OnChanged = func(_ string) {
		a.saveScanSettings()
	}
	a.timeoutEntry.OnChanged = func(_ string) {
		a.saveScanSettings()
	}
	a.threadsEntry.OnChanged = func(_ string) {
		a.saveScanSettings()
	}
	if a.inventoryDBEntry != nil {
		a.inventoryDBEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.inventoryAutoSaveCheck != nil {
		a.inventoryAutoSaveCheck.OnChanged = func(_ bool) {
			a.saveScanSettings()
		}
	}
	a.scanUDPCheck.OnChanged = func(_ bool) {
		a.saveScanSettings()
	}
	a.scanBannersCheck.OnChanged = func(_ bool) {
		a.saveScanSettings()
	}
	a.scanOSActiveCheck.OnChanged = func(_ bool) {
		a.saveScanSettings()
	}
	if a.scanVerboseLogsCheck != nil {
		a.scanVerboseLogsCheck.OnChanged = func(_ bool) {
			a.saveScanSettings()
		}
	}
	if a.scanVerboseInfoBtn != nil {
		a.scanVerboseInfoBtn.OnTapped = func() {
			dialog.ShowInformation(
				"Детальные логи по портам",
				"Режим включает debug-логи по отдельным TCP/UDP probe.\n\n"+
					"Используйте только для диагностики:\n"+
					"- заметно увеличивает размер лога;\n"+
					"- может замедлить сканирование на больших диапазонах.\n\n"+
					"Для обычной работы оставляйте опцию выключенной.",
				a.myWindow,
			)
		}
	}
	if a.autoProfileCheck != nil {
		a.autoProfileCheck.OnChanged = func(_ bool) {
			a.refreshAutoProfileStateLabel()
			a.saveScanSettings()
		}
	}
	if a.autoProfileInfoBtn != nil {
		a.autoProfileInfoBtn.OnTapped = func() {
			dialog.ShowInformation(
				"Автопрофиль сканирования",
				fmt.Sprintf(
					"Автопрофиль снижает риск перегрузки сети и UI на больших диапазонах.\n\n"+
						"Логика:\n"+
						"- от ~%d хостов: мягкое ограничение при слишком тяжелых настройках\n"+
						"- от ~%d хостов: до ports=1-2000 и threads<=64\n"+
						"- от ~%d хостов: до ports=1-1024 и threads<=40\n"+
						"- от ~%d хостов: до ports=1-512 и threads<=24\n\n"+
						"Опцию можно отключить, если нужен полный ручной контроль.",
					autoProfileHostWarn,
					autoProfileHostLarge,
					autoProfileHostXLarge,
					autoProfileHostXXLarge,
				),
				a.myWindow,
			)
		}
	}
	a.portWellKnownBtn.OnTapped = func() {
		a.portRangeEntry.SetText("0-1023")
		a.saveScanSettings()
	}
	a.portRegisteredBtn.OnTapped = func() {
		a.portRangeEntry.SetText("1024-49151")
		a.saveScanSettings()
	}
	a.portDynamicBtn.OnTapped = func() {
		a.portRangeEntry.SetText("49152-65535")
		a.saveScanSettings()
	}

	a.saveButton.OnTapped = func() {
		a.saveResults()
	}
	a.buildTopoBtn.OnTapped = func() {
		a.topoCtrl.BuildTopology(a.scanResults, a.myWindow)
	}
	a.stopTopoBtn.OnTapped = func() {
		a.topoCtrl.StopTopologyBuild()
	}
	a.saveTopoBtn.OnTapped = func() {
		a.topoCtrl.SaveTopology(a.lastTopology, a.myWindow)
	}
	a.copyPerfBtn.OnTapped = func() {
		a.topoCtrl.CopyPerformanceReport(a.myWindow)
	}
	if a.copyDiagnosticsBtn != nil {
		a.copyDiagnosticsBtn.OnTapped = func() {
			if a.diagnosticsLabel != nil {
				a.scanCtrl.CopyScanDiagnostics(a.diagnosticsLabel.Text)
			}
		}
	}
	if a.saveDiagnosticsBtn != nil {
		a.saveDiagnosticsBtn.OnTapped = func() {
			if a.diagnosticsLabel != nil {
				a.scanCtrl.SaveScanDiagnostics(a.diagnosticsLabel.Text)
			}
		}
	}
	a.savePerfBtn.OnTapped = func() {
		a.topoCtrl.SavePerformanceReport(a.myWindow)
	}
	a.refreshPreviewBtn.OnTapped = func() {
		a.topoCtrl.RefreshTopologyPreview(a.lastTopology, a.myWindow)
	}
	a.openPreviewBtn.OnTapped = func() {
		a.topoCtrl.OpenPreviewExternal(a.previewPath, a.myWindow)
	}
	a.zoomSelect.OnChanged = func(value string) {
		a.topoCtrl.ApplyTopologyZoom(value, a.myWindow)
	}
	if a.topologySearchEntry != nil {
		a.topologySearchEntry.OnChanged = func(v string) {
			a.topologyViewState.query = strings.TrimSpace(v)
			a.renderTopologyInteractiveMap(a.lastTopology)
		}
	}
	if a.topologyTypeFilterSel != nil {
		a.topologyTypeFilterSel.OnChanged = func(v string) {
			a.topologyViewState.typeFilter = strings.TrimSpace(strings.ToLower(v))
			a.renderTopologyInteractiveMap(a.lastTopology)
		}
	}
	if a.topologyConfidenceFilterSel != nil {
		a.topologyConfidenceFilterSel.OnChanged = func(v string) {
			a.topologyViewState.confFilter = strings.TrimSpace(strings.ToLower(v))
			a.renderTopologyInteractiveMap(a.lastTopology)
		}
	}
	if a.topologyResetMapBtn != nil {
		a.topologyResetMapBtn.OnTapped = func() {
			a.topologyViewState = topologyMapState{}
			if a.topologySearchEntry != nil {
				a.topologySearchEntry.SetText("")
			}
			if a.topologyTypeFilterSel != nil {
				a.topologyTypeFilterSel.SetSelected("all")
			}
			if a.topologyConfidenceFilterSel != nil {
				a.topologyConfidenceFilterSel.SetSelected("all")
			}
			a.renderTopologyInteractiveMap(a.lastTopology)
		}
	}
	if a.toolsHostEntry != nil {
		a.toolsHostEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsPingCountEnt != nil {
		a.toolsPingCountEnt.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsTimeoutEnt != nil {
		a.toolsTimeoutEnt.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsTraceHopsEnt != nil {
		a.toolsTraceHopsEnt.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsDNSResolverEnt != nil {
		a.toolsDNSResolverEnt.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsDeviceTargetEntry != nil {
		a.toolsDeviceTargetEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsDeviceUserEntry != nil {
		a.toolsDeviceUserEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsDevicePassEntry != nil {
		a.toolsDevicePassEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	a.toolsPingBtn.OnTapped = func() {
		a.toolsCtrl.RunPingTool()
	}
	a.toolsTraceBtn.OnTapped = func() {
		a.toolsCtrl.RunTracerouteTool()
	}
	a.toolsDNSBtn.OnTapped = func() {
		a.toolsCtrl.RunDNSTool()
	}
	a.toolsWhoisBtn.OnTapped = func() {
		a.toolsCtrl.RunWhoisTool()
	}
	a.toolsWiFiBtn.OnTapped = func() {
		a.toolsCtrl.RunWiFiTool()
	}
	a.toolsWOLBtn.OnTapped = func() {
		a.toolsCtrl.RunWOLTool()
	}
	a.toolsAuditBtn.OnTapped = func() {
		a.toolsCtrl.RunPortAuditTool(a.scanResults)
	}
	a.toolsRiskBtn.OnTapped = func() {
		a.toolsCtrl.RunRiskSignaturesTool(a.scanResults)
	}
	a.toolsDeviceStatusBtn.OnTapped = func() {
		a.toolsCtrl.RunDeviceControlTool(devicecontrol.ActionStatus)
	}
	a.toolsDeviceRebootBtn.OnTapped = func() {
		dialog.NewConfirm(
			"Подтверждение опасного действия",
			"Подтвердите reboot устройства. Действие может привести к кратковременной недоступности сети.",
			func(ok bool) {
				if !ok {
					return
				}
				a.runDeviceControlTool(devicecontrol.ActionReboot)
			},
			a.myWindow,
		).Show()
	}
}

// autoDetectNetwork автоматически определяет сеть и заполняет поле ввода
func (a *App) autoDetectNetwork() {
	// Запускаем определение сети в отдельной горутине, чтобы не блокировать UI
	go func() {
		networkStr, err := network.DetectLocalNetwork()
		// Обновляем UI через fyne.Do для выполнения в главном потоке
		fyne.Do(func() {
			if err == nil && networkStr != "" {
				a.networkEntry.SetText(networkStr)
				a.saveScanSettings()
				a.statusLabel.SetText(fmt.Sprintf("Сеть определена автоматически: %s", networkStr))
				a.topologyStatus.SetText(fmt.Sprintf("Сеть определена автоматически: %s", networkStr))
			} else {
				// Если не удалось определить, оставляем поле пустым
				a.statusLabel.SetText("Готов к сканированию (сеть будет определена автоматически при запуске)")
				a.topologyStatus.SetText("Готово к построению топологии после сканирования")
			}
			// Обновляем виджеты
			a.statusLabel.Refresh()
			a.topologyStatus.Refresh()
			a.networkEntry.Refresh()
		})
	}()
}

func (a *App) recommendedBadgeClassForHosts(hosts int) string {
	switch {
	case hosts >= autoProfileHostXXLarge:
		return "very-large"
	case hosts >= autoProfileHostXLarge:
		return "large"
	case hosts >= autoProfileHostLarge:
		return "medium"
	default:
		return "small"
	}
}

func (a *App) recommendedBadgeText(profileName string, badgeClass string) string {
	return fmt.Sprintf("Профиль: %s (%s)", profileName, badgeClass)
}

func (a *App) recommendedProfileNameForClass(badgeClass string) (string, bool) {
	switch strings.TrimSpace(badgeClass) {
	case "very-large":
		return "бережный для очень крупной подсети", true
	case "large":
		return "бережный для крупной подсети", true
	case "medium":
		return "сбалансированный для средней подсети", true
	case "small":
		return "углубленный для небольшой подсети", true
	default:
		return "", false
	}
}

func (a *App) saveScanSettings() {
	if a == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	p.SetString(prefNetwork, strings.TrimSpace(a.networkEntry.Text))
	p.SetString(prefPortRange, strings.TrimSpace(a.portRangeEntry.Text))
	p.SetString(prefTimeout, strings.TrimSpace(a.timeoutEntry.Text))
	p.SetString(prefThreads, strings.TrimSpace(a.threadsEntry.Text))
	if a.scanUDPCheck.Checked {
		p.SetString(prefScanUDP, "true")
	} else {
		p.SetString(prefScanUDP, "false")
	}
	if a.scanBannersCheck != nil && a.scanBannersCheck.Checked {
		p.SetString(prefScanBanners, "true")
	} else {
		p.SetString(prefScanBanners, "false")
	}
	if a.scanOSActiveCheck != nil && a.scanOSActiveCheck.Checked {
		p.SetString(prefScanOSActive, "true")
	} else {
		p.SetString(prefScanOSActive, "false")
	}
	if a.scanVerboseLogsCheck != nil && a.scanVerboseLogsCheck.Checked {
		p.SetString(prefScanVerbosePortLogs, "true")
	} else {
		p.SetString(prefScanVerbosePortLogs, "false")
	}
	if a.scanTCPPortsCheck != nil {
		if a.scanTCPPortsCheck.Checked {
			p.SetString(prefScanTCPPorts, "true")
		} else {
			p.SetString(prefScanTCPPorts, "false")
		}
	}
	if a.autoProfileCheck != nil {
		if a.autoProfileCheck.Checked {
			p.SetString(prefAutoProfile, "true")
		} else {
			p.SetString(prefAutoProfile, "false")
		}
	}
	if a.toolsHostEntry != nil {
		p.SetString(prefToolHost, strings.TrimSpace(a.toolsHostEntry.Text))
	}
	if a.toolsPingCountEnt != nil {
		p.SetString(prefToolPingCount, strings.TrimSpace(a.toolsPingCountEnt.Text))
	}
	if a.toolsTimeoutEnt != nil {
		p.SetString(prefToolTimeout, strings.TrimSpace(a.toolsTimeoutEnt.Text))
	}
	if a.toolsTraceHopsEnt != nil {
		p.SetString(prefToolTraceHops, strings.TrimSpace(a.toolsTraceHopsEnt.Text))
	}
	if a.toolsDNSResolverEnt != nil {
		p.SetString(prefToolResolver, strings.TrimSpace(a.toolsDNSResolverEnt.Text))
	}
	if a.toolsAuditMinSeveritySel != nil {
		p.SetString(prefToolAuditMinSeverity, strings.TrimSpace(a.toolsAuditMinSeveritySel.Selected))
	}
	if a.toolsDeviceTargetEntry != nil {
		p.SetString(prefToolDeviceTarget, strings.TrimSpace(a.toolsDeviceTargetEntry.Text))
	}
	if a.toolsDeviceVendorEntry != nil {
		p.SetString(prefToolDeviceVendor, strings.TrimSpace(a.toolsDeviceVendorEntry.Selected))
	}
	if a.toolsDeviceUserEntry != nil {
		p.SetString(prefToolDeviceUser, strings.TrimSpace(a.toolsDeviceUserEntry.Text))
	}
	if a.recommendedProfileBadge != nil {
		p.SetString(prefRecommendedBadge, strings.TrimSpace(a.recommendedProfileBadge.Text))
	}
	if a.inventoryDBEntry != nil {
		p.SetString(prefInventoryDBPath, strings.TrimSpace(a.inventoryDBEntry.Text))
	}
	if a.inventoryAutoSaveCheck != nil {
		if a.inventoryAutoSaveCheck.Checked {
			p.SetString(prefInventoryAutoSave, "true")
		} else {
			p.SetString(prefInventoryAutoSave, "false")
		}
	}
}

func (a *App) setPortRangeControlsEnabled(enabled bool) {
	if a == nil {
		return
	}
	if a.portRangeEntry != nil {
		if enabled {
			a.portRangeEntry.Enable()
		} else {
			a.portRangeEntry.Disable()
		}
	}
	for _, b := range []*widget.Button{
		a.presetQuickBtn, a.presetBalBtn, a.presetDeepBtn,
		a.portWellKnownBtn, a.portRegisteredBtn, a.portDynamicBtn,
	} {
		if b != nil {
			if enabled {
				b.Enable()
			} else {
				b.Disable()
			}
		}
	}
}

func (a *App) loadScanTabSplitFromPrefs() {
	if a == nil || a.scanTabMainSplit == nil || a.myApp == nil {
		return
	}
	v := a.myApp.Preferences().FloatWithFallback(prefScanTabSplitOffset, -1)
	if v >= 0.16 && v <= 0.82 {
		a.scanTabMainSplit.Offset = v
		a.scanTabSplitInitialized = true
		a.scanTabSplitPersistPrimed = true
		a.lastPersistedScanSplit = v
	}
}

func (a *App) clampScanTabMainSplitOffset() {
	if a == nil || a.scanTabMainSplit == nil {
		return
	}
	const lo, hi = 0.15, 0.78
	o := a.scanTabMainSplit.Offset
	if o < lo {
		a.scanTabMainSplit.Offset = lo
	} else if o > hi {
		a.scanTabMainSplit.Offset = hi
	}
}

func (a *App) maybePersistScanTabSplitOffset() {
	if a == nil || a.scanTabMainSplit == nil || a.myApp == nil {
		return
	}
	maybePersistFloatPref(a.myApp.Preferences(), prefScanTabSplitOffset, a.scanTabMainSplit.Offset,
		&a.scanTabSplitPersistPrimed, &a.lastPersistedScanSplit, nil)
}

func (a *App) loadTopologySplitFromPrefs() {
	if a == nil || a.topologyMainSplit == nil || a.myApp == nil {
		return
	}
	v := a.myApp.Preferences().FloatWithFallback(prefTopologyMainSplitOffset, -1)
	if v >= 0.18 && v <= 0.88 {
		a.topologyMainSplit.Offset = v
		a.topologySplitInitialized = true
		a.topologySplitPersistPrimed = true
		a.lastPersistedTopologySplit = v
	}
}

func (a *App) clampTopologyMainSplitOffset() {
	if a == nil || a.topologyMainSplit == nil {
		return
	}
	const lo, hi = 0.18, 0.85
	o := a.topologyMainSplit.Offset
	if o < lo {
		a.topologyMainSplit.Offset = lo
	} else if o > hi {
		a.topologyMainSplit.Offset = hi
	}
}

func (a *App) maybePersistTopologySplitOffset() {
	if a == nil || a.topologyMainSplit == nil || a.myApp == nil {
		return
	}
	maybePersistFloatPref(a.myApp.Preferences(), prefTopologyMainSplitOffset, a.topologyMainSplit.Offset,
		&a.topologySplitPersistPrimed, &a.lastPersistedTopologySplit, nil)
}

func (a *App) loadToolsTabSplitFromPrefs() {
	if a == nil || a.toolsTabMainSplit == nil || a.myApp == nil {
		return
	}
	v := a.myApp.Preferences().FloatWithFallback(prefToolsTabSplitOffset, -1)
	if v >= 0.22 && v <= 0.82 {
		a.toolsTabMainSplit.Offset = v
		a.toolsSplitInitialized = true
		a.toolsSplitPersistPrimed = true
		a.lastPersistedToolsSplit = v
	}
}

func (a *App) clampToolsTabMainSplitOffset() {
	if a == nil || a.toolsTabMainSplit == nil {
		return
	}
	const lo, hi = 0.22, 0.78
	o := a.toolsTabMainSplit.Offset
	if o < lo {
		a.toolsTabMainSplit.Offset = lo
	} else if o > hi {
		a.toolsTabMainSplit.Offset = hi
	}
}

func (a *App) maybePersistToolsTabSplitOffset() {
	if a == nil || a.toolsTabMainSplit == nil || a.myApp == nil {
		return
	}
	maybePersistFloatPref(a.myApp.Preferences(), prefToolsTabSplitOffset, a.toolsTabMainSplit.Offset,
		&a.toolsSplitPersistPrimed, &a.lastPersistedToolsSplit, nil)
}

func (a *App) loadHostDetailsSplitFromPrefs() {
	if a == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	if v := p.FloatWithFallback(prefHostDetailsSplitOffsetV, -1); v >= 0.28 && v <= 0.92 {
		a.rememberedHostDetailsSplitV = v
		a.hostDetailsSplitPrimedV = true
		a.lastPersistedHostDetailsV = v
	}
	if h := p.FloatWithFallback(prefHostDetailsSplitOffsetH, -1); h >= 0.35 && h <= 0.90 {
		a.rememberedHostDetailsSplitH = h
		a.hostDetailsSplitPrimedH = true
		a.lastPersistedHostDetailsH = h
	}
}

func (a *App) maybePersistHostDetailsSplitOffsets() {
	if a == nil || a.resultsMainSplit == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	cur := a.resultsMainSplit.Offset
	switch a.lastHostDetailsSplitKind {
	case "V":
		maybePersistFloatPref(p, prefHostDetailsSplitOffsetV, cur, &a.hostDetailsSplitPrimedV, &a.lastPersistedHostDetailsV, func(v float64) {
			a.rememberedHostDetailsSplitV = v
		})
	case "H":
		maybePersistFloatPref(p, prefHostDetailsSplitOffsetH, cur, &a.hostDetailsSplitPrimedH, &a.lastPersistedHostDetailsH, func(v float64) {
			a.rememberedHostDetailsSplitH = v
		})
	}
}

func (a *App) loadScanSettings() {
	if a == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	if v := strings.TrimSpace(p.String(prefNetwork)); v != "" {
		a.networkEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefPortRange)); v != "" {
		a.portRangeEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefTimeout)); v != "" {
		a.timeoutEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefThreads)); v != "" {
		a.threadsEntry.SetText(v)
	}
	a.scanUDPCheck.SetChecked(strings.EqualFold(strings.TrimSpace(p.String(prefScanUDP)), "true"))
	if a.scanBannersCheck != nil {
		a.scanBannersCheck.SetChecked(strings.EqualFold(strings.TrimSpace(p.String(prefScanBanners)), "true"))
	}
	if a.scanOSActiveCheck != nil {
		a.scanOSActiveCheck.SetChecked(strings.EqualFold(strings.TrimSpace(p.String(prefScanOSActive)), "true"))
	}
	if a.scanVerboseLogsCheck != nil {
		a.scanVerboseLogsCheck.SetChecked(strings.EqualFold(strings.TrimSpace(p.String(prefScanVerbosePortLogs)), "true"))
	}
	if a.scanTCPPortsCheck != nil {
		tcpPref := strings.TrimSpace(p.String(prefScanTCPPorts))
		if tcpPref == "" || strings.EqualFold(tcpPref, "true") {
			a.scanTCPPortsCheck.SetChecked(true)
		} else {
			a.scanTCPPortsCheck.SetChecked(false)
		}
		a.setPortRangeControlsEnabled(a.scanTCPPortsCheck.Checked)
	}
	if a.autoProfileCheck != nil {
		autoPref := strings.TrimSpace(p.String(prefAutoProfile))
		if autoPref == "" || strings.EqualFold(autoPref, "true") {
			a.autoProfileCheck.SetChecked(true)
		} else {
			a.autoProfileCheck.SetChecked(false)
		}
	}
	a.refreshAutoProfileStateLabel()

	switch strings.TrimSpace(p.String(prefPreset)) {
	case "quick":
		a.statusLabel.SetText("Пресет: Быстро (восстановлен)")
	case "deep":
		a.statusLabel.SetText("Пресет: Глубоко (восстановлен)")
	case "balanced":
		a.statusLabel.SetText("Пресет: Баланс (восстановлен)")
	case "recommended":
		a.statusLabel.SetText("Пресет: Рекомендуемые настройки (восстановлен)")
	}
	if a.recommendedProfileBadge != nil {
		if badgeClass := strings.TrimSpace(p.String(prefRecommendedBadgeClass)); badgeClass != "" {
			if profileName, ok := a.recommendedProfileNameForClass(badgeClass); ok {
				a.recommendedProfileBadge.Text = a.recommendedBadgeText(profileName, badgeClass)
				a.recommendedProfileBadge.Color = color.RGBA{R: 55, G: 130, B: 200, A: 255}
				a.recommendedProfileBadge.Refresh()
			} else if badgeText := strings.TrimSpace(p.String(prefRecommendedBadge)); badgeText != "" {
				a.recommendedProfileBadge.Text = badgeText
				a.recommendedProfileBadge.Color = color.RGBA{R: 55, G: 130, B: 200, A: 255}
				a.recommendedProfileBadge.Refresh()
			}
		} else if badgeText := strings.TrimSpace(p.String(prefRecommendedBadge)); badgeText != "" {
			a.recommendedProfileBadge.Text = badgeText
			a.recommendedProfileBadge.Color = color.RGBA{R: 55, G: 130, B: 200, A: 255}
			a.recommendedProfileBadge.Refresh()
		}
	}
	viewMode := strings.TrimSpace(p.String(prefViewMode))
	if viewMode == "Таблица" || viewMode == "Карточки" {
		a.resultsMode = viewMode
		a.resultsModeSel.SetSelected(viewMode)
	}
	subMode := strings.TrimSpace(p.String(prefResultsSubMode))
	if subMode == "Devices" || subMode == "Security" || subMode == "Inventory" {
		a.resultsSubMode = subMode
		if a.resultsSubModeSel != nil {
			a.resultsSubModeSel.SetSelected(subMode)
		}
	}
	if v := strings.TrimSpace(p.String(prefInventoryDBPath)); v != "" && a.inventoryDBEntry != nil {
		a.inventoryDBEntry.SetText(v)
	}
	if a.inventoryAutoSaveCheck != nil {
		autoSave := strings.TrimSpace(p.String(prefInventoryAutoSave))
		a.inventoryAutoSaveCheck.SetChecked(autoSave == "" || strings.EqualFold(autoSave, "true"))
	}
	sortMode := strings.TrimSpace(p.String(prefSortMode))
	if sortMode == "IP" || sortMode == "HostName" {
		a.resultsSort = sortMode
		a.resultsSortSel.SetSelected(sortMode)
	}
	if v, err := strconv.Atoi(strings.TrimSpace(p.String(prefChipLimit))); err == nil && v > 0 {
		a.maxPortChips = v
		if a.chipLimitSel != nil {
			a.chipLimitSel.SetSelected(strconv.Itoa(v))
		}
	}
	a.showRawBanners = strings.EqualFold(strings.TrimSpace(p.String(prefShowRawBanners)), "true")
	if a.showRawBannersCheck != nil {
		a.showRawBannersCheck.SetChecked(a.showRawBanners)
	}
	if v := strings.TrimSpace(p.String(prefFilterQuery)); v != "" {
		a.resultsFilterQuery = v
		if a.resultsFilterEnt != nil {
			a.resultsFilterEnt.SetText(v)
		}
	}
	a.onlyWithOpenPorts = strings.EqualFold(strings.TrimSpace(p.String(prefOnlyOpenPorts)), "true")
	if a.openPortsOnlyCheck != nil {
		a.openPortsOnlyCheck.SetChecked(a.onlyWithOpenPorts)
	}
	if rawTypes := strings.TrimSpace(p.String(prefTypeFilters)); rawTypes != "" {
		for _, typeName := range strings.Split(rawTypes, ",") {
			name := strings.TrimSpace(typeName)
			if check, ok := a.quickTypeChecks[name]; ok && check != nil {
				check.SetChecked(true)
			}
		}
	}
	if v := strings.TrimSpace(p.String(prefToolHost)); v != "" && a.toolsHostEntry != nil {
		a.toolsHostEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefToolPingCount)); v != "" && a.toolsPingCountEnt != nil {
		a.toolsPingCountEnt.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefToolTimeout)); v != "" && a.toolsTimeoutEnt != nil {
		a.toolsTimeoutEnt.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefToolTraceHops)); v != "" && a.toolsTraceHopsEnt != nil {
		a.toolsTraceHopsEnt.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefToolResolver)); v != "" && a.toolsDNSResolverEnt != nil {
		a.toolsDNSResolverEnt.SetText(v)
	}
	if a.toolsAuditMinSeveritySel != nil {
		sev := strings.TrimSpace(p.String(prefToolAuditMinSeverity))
		if sev == "" {
			sev = "low"
		}
		if _, ok := audit.NormalizeSeverity(sev); !ok {
			sev = "low"
		}
		a.toolsAuditMinSeveritySel.SetSelected(sev)
	}
	if v := strings.TrimSpace(p.String(prefToolDeviceTarget)); v != "" && a.toolsDeviceTargetEntry != nil {
		a.toolsDeviceTargetEntry.SetText(v)
	}
	if a.toolsDeviceVendorEntry != nil {
		vendor := strings.TrimSpace(p.String(prefToolDeviceVendor))
		if vendor == "" {
			vendor = devicecontrol.VendorGenericHTTP
		}
		a.toolsDeviceVendorEntry.SetSelected(vendor)
	}
	if v := strings.TrimSpace(p.String(prefToolDeviceUser)); v != "" && a.toolsDeviceUserEntry != nil {
		a.toolsDeviceUserEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefCidrFilter)); v != "" {
		if a.resultsCidrFilterEnt != nil {
			a.resultsCidrFilterEnt.SetText(v)
		}
	}
	mode := strings.TrimSpace(p.String(prefPortStateMode))
	switch mode {
	case "has_open":
		a.resultsPortStateMode = mode
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть открытые")
		}
	case "has_closed":
		a.resultsPortStateMode = mode
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть закрытые")
		}
	case "has_filtered":
		a.resultsPortStateMode = mode
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть фильтруемые")
		}
	default:
		a.resultsPortStateMode = "all"
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Все")
		}
	}
	a.loadScanTabSplitFromPrefs()
	a.loadTopologySplitFromPrefs()
	a.loadToolsTabSplitFromPrefs()
	a.loadHostDetailsSplitFromPrefs()
	a.refreshInventorySnapshots()
	a.renderScanResultsView()
}

func (a *App) refreshAutoProfileStateLabel() {
	if a == nil || a.autoProfileStateText == nil {
		return
	}
	enabled := true
	if a.autoProfileCheck != nil {
		enabled = a.autoProfileCheck.Checked
	}
	if enabled {
		a.autoProfileStateText.Text = "Автопрофиль: ВКЛ"
		a.autoProfileStateText.Color = color.RGBA{R: 60, G: 170, B: 80, A: 255}
		if a.autoProfileHeaderLabel != nil {
			a.autoProfileHeaderLabel.SetText("Режим сканирования: Автопрофиль ВКЛ")
			a.autoProfileHeaderLabel.Refresh()
		}
	} else {
		a.autoProfileStateText.Text = "Автопрофиль: ВЫКЛ"
		a.autoProfileStateText.Color = color.RGBA{R: 140, G: 140, B: 140, A: 255}
		if a.autoProfileHeaderLabel != nil {
			a.autoProfileHeaderLabel.SetText("Режим сканирования: Автопрофиль ВЫКЛ")
			a.autoProfileHeaderLabel.Refresh()
		}
	}
	a.autoProfileStateText.Refresh()
}

func (a *App) saveResultsViewSettings() {
	if a == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	p.SetString(prefViewMode, strings.TrimSpace(a.resultsMode))
	p.SetString(prefResultsSubMode, strings.TrimSpace(a.resultsSubMode))
	p.SetString(prefSortMode, strings.TrimSpace(a.resultsSort))
	p.SetString(prefChipLimit, strconv.Itoa(a.maxPortChips))
	if a.showRawBanners {
		p.SetString(prefShowRawBanners, "true")
	} else {
		p.SetString(prefShowRawBanners, "false")
	}
	p.SetString(prefFilterQuery, strings.TrimSpace(a.resultsFilterQuery))
	if a.onlyWithOpenPorts {
		p.SetString(prefOnlyOpenPorts, "true")
	} else {
		p.SetString(prefOnlyOpenPorts, "false")
	}
	selectedTypes := make([]string, 0)
	for typeName, check := range a.quickTypeChecks {
		if check != nil && check.Checked {
			selectedTypes = append(selectedTypes, typeName)
		}
	}
	if len(selectedTypes) > 1 {
		// Keep serialized settings deterministic for easier debugging.
		sort.Strings(selectedTypes)
	}
	p.SetString(prefTypeFilters, strings.Join(selectedTypes, ","))
	if a.resultsCidrFilterEnt != nil {
		p.SetString(prefCidrFilter, strings.TrimSpace(a.resultsCidrFilterEnt.Text))
	}
	mode := strings.TrimSpace(a.resultsPortStateMode)
	if mode == "" {
		mode = "all"
	}
	p.SetString(prefPortStateMode, mode)
}

func (a *App) filterPresetKey(slot string) string {
	switch strings.TrimSpace(slot) {
	case "1":
		return prefFilterPreset1
	case "2":
		return prefFilterPreset2
	case "3":
		return prefFilterPreset3
	default:
		return prefFilterPreset1
	}
}

func (a *App) serializeCurrentFilters() string {
	selectedTypes := make([]string, 0)
	for typeName, check := range a.quickTypeChecks {
		if check != nil && check.Checked {
			selectedTypes = append(selectedTypes, typeName)
		}
	}
	if len(selectedTypes) > 1 {
		sort.Strings(selectedTypes)
	}
	cidr := ""
	if a.resultsCidrFilterEnt != nil {
		cidr = strings.TrimSpace(a.resultsCidrFilterEnt.Text)
	}
	mode := strings.TrimSpace(a.resultsPortStateMode)
	if mode == "" {
		mode = "all"
	}
	onlyOpen := "false"
	if a.onlyWithOpenPorts {
		onlyOpen = "true"
	}
	return strings.Join([]string{
		strings.TrimSpace(a.resultsFilterQuery),
		cidr,
		mode,
		onlyOpen,
		strings.Join(selectedTypes, ","),
	}, "|")
}

func (a *App) saveFilterPreset(slot string) {
	if a == nil || a.myApp == nil {
		return
	}
	key := a.filterPresetKey(slot)
	a.myApp.Preferences().SetString(key, a.serializeCurrentFilters())
	if a.statusLabel != nil {
		a.statusLabel.SetText(fmt.Sprintf("Пресет фильтров %s сохранен", slot))
		a.statusLabel.Refresh()
	}
}

func (a *App) applyFilterPreset(slot string) {
	if a == nil || a.myApp == nil {
		return
	}
	key := a.filterPresetKey(slot)
	raw := strings.TrimSpace(a.myApp.Preferences().String(key))
	if raw == "" {
		if a.statusLabel != nil {
			a.statusLabel.SetText(fmt.Sprintf("Пресет фильтров %s пуст", slot))
			a.statusLabel.Refresh()
		}
		return
	}
	parts := strings.SplitN(raw, "|", 5)
	if len(parts) < 5 {
		if a.statusLabel != nil {
			a.statusLabel.SetText(fmt.Sprintf("Пресет фильтров %s поврежден", slot))
			a.statusLabel.Refresh()
		}
		return
	}
	a.resultsFilterQuery = strings.TrimSpace(parts[0])
	if a.resultsFilterEnt != nil {
		a.resultsFilterEnt.SetText(a.resultsFilterQuery)
	}
	if a.resultsCidrFilterEnt != nil {
		a.resultsCidrFilterEnt.SetText(strings.TrimSpace(parts[1]))
	}
	a.resultsPortStateMode = strings.TrimSpace(parts[2])
	switch a.resultsPortStateMode {
	case "has_open":
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть открытые")
		}
	case "has_closed":
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть закрытые")
		}
	case "has_filtered":
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть фильтруемые")
		}
	default:
		a.resultsPortStateMode = "all"
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Все")
		}
	}
	a.onlyWithOpenPorts = strings.EqualFold(strings.TrimSpace(parts[3]), "true")
	if a.openPortsOnlyCheck != nil {
		a.openPortsOnlyCheck.SetChecked(a.onlyWithOpenPorts)
	}
	for _, ch := range a.quickTypeChecks {
		if ch != nil {
			ch.SetChecked(false)
		}
	}
	if typeCSV := strings.TrimSpace(parts[4]); typeCSV != "" {
		for _, typeName := range strings.Split(typeCSV, ",") {
			name := strings.TrimSpace(typeName)
			if check, ok := a.quickTypeChecks[name]; ok && check != nil {
				check.SetChecked(true)
			}
		}
	}
	a.saveResultsViewSettings()
	a.renderScanResultsView()
	if a.statusLabel != nil {
		a.statusLabel.SetText(fmt.Sprintf("Пресет фильтров %s применен", slot))
		a.statusLabel.Refresh()
	}
}

func (a *App) withToolHost() (string, bool) {
	if a == nil || a.toolsHostEntry == nil {
		return "", false
	}
	host := strings.TrimSpace(a.toolsHostEntry.Text)
	if host == "" {
		dialog.ShowInformation("Инструменты", "Введите хост или IP", a.myWindow)
		return "", false
	}
	return host, true
}

func (a *App) setToolsOutputMarkdown(md string) {
	if a == nil || a.toolsOutput == nil {
		return
	}
	a.toolsOutput.ParseMarkdown(md)
	a.toolsOutput.Refresh()
}

func (a *App) setToolsButtonsEnabled(enabled bool) {
	for _, b := range []*widget.Button{
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
	} {
		if b == nil {
			continue
		}
		if enabled {
			b.Enable()
		} else {
			b.Disable()
		}
	}
}

func (a *App) runToolOperation(title string, startedMessage string, task func(context.Context) (string, error)) {
	if a == nil {
		return
	}
	if strings.TrimSpace(startedMessage) == "" {
		startedMessage = "Выполняется операция..."
	}
	a.setToolsButtonsEnabled(false)
	a.setToolsOutputMarkdown(startedMessage)

	run := func(ctx context.Context) error {
		markdown, err := task(ctx)
		fyne.Do(func() {
			a.setToolsButtonsEnabled(true)
			if err != nil {
				a.setToolsOutputMarkdown(markdown)
				return
			}
			a.setToolsOutputMarkdown(markdown)
		})
		return err
	}

	if a.operations == nil {
		go func() {
			_ = run(context.Background())
		}()
		return
	}
	a.operations.Run(OperationTypeTool, title, run)
}

func (a *App) startOperationsWatcher() {
	if a == nil || a.operations == nil {
		return
	}
	events := a.operations.Subscribe(32)
	go func() {
		for ev := range events {
			op := ev.Operation
			fyne.Do(func() {
				a.pushOperationHistory(op)
			})
		}
	}()
}

func (a *App) pushOperationHistory(op Operation) {
	if a == nil || a.operationsOutput == nil {
		return
	}
	updated := false
	for i := range a.operationsHistory {
		if a.operationsHistory[i].ID == op.ID {
			a.operationsHistory[i] = op
			updated = true
			break
		}
	}
	if !updated {
		a.operationsHistory = append([]Operation{op}, a.operationsHistory...)
	}
	if len(a.operationsHistory) > 20 {
		a.operationsHistory = a.operationsHistory[:20]
	}
	a.refreshOperationSelectOptions()
	a.operationsOutput.ParseMarkdown(a.operationsHistoryMarkdown())
	a.operationsOutput.Refresh()
	a.refreshOperationActionsState()
}

func (a *App) operationsHistoryMarkdown() string {
	var sb strings.Builder
	sb.WriteString("### Operations Center\n\n")
	if len(a.operationsHistory) == 0 {
		sb.WriteString("История операций пуста.")
		return sb.String()
	}
	for _, op := range a.operationsHistory {
		dur := "-"
		if op.Duration > 0 {
			dur = op.Duration.Round(time.Millisecond).String()
		}
		sb.WriteString(fmt.Sprintf("- `%s` **%s** `%s` (%s)\n", op.ID, strings.ToUpper(string(op.Status)), op.Title, dur))
		if strings.TrimSpace(op.Error) != "" {
			sb.WriteString(fmt.Sprintf("  - error: %s\n", strings.TrimSpace(op.Error)))
		}
	}
	return sb.String()
}

func (a *App) refreshOperationSelectOptions() {
	if a == nil || a.operationsSelect == nil {
		return
	}
	options := make([]string, 0, len(a.operationsHistory))
	a.operationsSelectMap = make(map[string]string, len(a.operationsHistory))
	selectedLabel := ""
	for _, op := range a.operationsHistory {
		label := fmt.Sprintf("%s | %s | %s", op.ID, strings.ToUpper(string(op.Status)), strings.TrimSpace(op.Title))
		options = append(options, label)
		a.operationsSelectMap[label] = op.ID
		if strings.TrimSpace(op.ID) == strings.TrimSpace(a.selectedOperationID) {
			selectedLabel = label
		}
	}
	a.operationsSelect.Options = options
	a.operationsSelect.Refresh()
	if selectedLabel != "" {
		a.operationsSelect.SetSelected(selectedLabel)
		return
	}
	if len(options) > 0 {
		a.operationsSelect.SetSelected(options[0])
		return
	}
	a.selectedOperationID = ""
}

func (a *App) refreshOperationActionsState() {
	if a == nil {
		return
	}
	if a.operationsRetryBtn == nil || a.operationsCancelBtn == nil || a.operations == nil {
		return
	}
	a.operationsRetryBtn.Disable()
	a.operationsCancelBtn.Disable()
	id := strings.TrimSpace(a.selectedOperationID)
	if id == "" {
		return
	}
	op, ok := a.operations.Get(id)
	if !ok {
		return
	}
	if op.CanRetry && (op.Status == OperationFailed || op.Status == OperationCanceled) {
		a.operationsRetryBtn.Enable()
	}
	if op.CanCancel && (op.Status == OperationQueued || op.Status == OperationRunning) {
		a.operationsCancelBtn.Enable()
	}
}

func (a *App) runPortAuditTool() {
	a.runToolOperation("Port Audit", "Выполняется аудит портов...", func(ctx context.Context) (string, error) {
		findings := audit.EvaluateOpenPorts(a.scanResults)
		minSeverity := "all"
		if a.toolsAuditMinSeveritySel != nil {
			if norm, ok := audit.NormalizeSeverity(strings.TrimSpace(a.toolsAuditMinSeveritySel.Selected)); ok {
				minSeverity = norm
			}
		}
		findings = audit.FilterByMinSeverity(findings, minSeverity)
		var sb strings.Builder
		sb.WriteString("### Аудит открытых портов\n\n")
		sb.WriteString(fmt.Sprintf("- Min severity: `%s`\n\n", minSeverity))
		if len(findings) == 0 {
			sb.WriteString("- Рисков по базовым правилам не найдено.")
			return sb.String(), nil
		}
		sb.WriteString("```text\n")
		sb.WriteString(audit.FormatFindings(findings))
		sb.WriteString("\n```")
		return sb.String(), nil
	})
}

func (a *App) runRiskSignaturesTool() {
	a.runToolOperation("Risk Signatures", "Запуск Risk Signatures...", func(ctx context.Context) (string, error) {
		var sb strings.Builder
		sb.WriteString("### Risk Signatures\n\n")
		if len(a.scanResults) == 0 {
			sb.WriteString("- Сначала выполните сканирование сети.")
			return sb.String(), nil
		}
		db, err := risksignature.LoadDefault()
		if err != nil {
			return fmt.Sprintf("### Risk Signatures\n\nОшибка загрузки сигнатур: `%v`", err), err
		}
		findings := risksignature.Evaluate(a.scanResults, db)
		sb.WriteString(fmt.Sprintf("- DB version: `%s`\n", strings.TrimSpace(db.Version)))
		if len(findings) == 0 {
			sb.WriteString("- Findings: нет\n")
			return sb.String(), nil
		}
		sb.WriteString(fmt.Sprintf("- Findings: `%d`\n\n", len(findings)))
		for _, f := range findings {
			sb.WriteString(fmt.Sprintf("- [%s] `%s` `%s` host `%s`\n",
				strings.ToUpper(strings.TrimSpace(f.Severity)),
				strings.TrimSpace(f.Title),
				strings.TrimSpace(f.SignatureID),
				strings.TrimSpace(f.HostIP)))
			if strings.TrimSpace(f.Reason) != "" {
				sb.WriteString(fmt.Sprintf("  - reason: %s\n", strings.TrimSpace(f.Reason)))
			}
			if strings.TrimSpace(f.Recommendation) != "" {
				sb.WriteString(fmt.Sprintf("  - recommendation: %s\n", strings.TrimSpace(f.Recommendation)))
			}
			if strings.TrimSpace(f.ReferenceURL) != "" {
				sb.WriteString(fmt.Sprintf("  - reference: %s\n", strings.TrimSpace(f.ReferenceURL)))
			}
		}
		return sb.String(), nil
	})
}

func (a *App) runDeviceControlTool(action string) {
	if a == nil || a.toolsDeviceTargetEntry == nil {
		return
	}
	target := strings.TrimSpace(a.toolsDeviceTargetEntry.Text)
	if target == "" {
		dialog.ShowInformation("Device Control", "Введите Device target URL", a.myWindow)
		return
	}
	vendor := devicecontrol.VendorGenericHTTP
	if a.toolsDeviceVendorEntry != nil && strings.TrimSpace(a.toolsDeviceVendorEntry.Selected) != "" {
		vendor = strings.TrimSpace(a.toolsDeviceVendorEntry.Selected)
	}
	username := ""
	if a.toolsDeviceUserEntry != nil {
		username = strings.TrimSpace(a.toolsDeviceUserEntry.Text)
	}
	password := ""
	if a.toolsDevicePassEntry != nil {
		password = strings.TrimSpace(a.toolsDevicePassEntry.Text)
	}
	timeoutSec := 10
	if a.toolsTimeoutEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}
	a.runToolOperation("Device Control", fmt.Sprintf("Выполняется device action `%s`...", action), func(ctx context.Context) (string, error) {
		req := devicecontrol.Request{
			Action:    action,
			TargetURL: target,
			Vendor:    vendor,
			Username:  username,
			Password:  password,
			Timeout:   time.Duration(timeoutSec) * time.Second,
		}
		res, err := devicecontrol.Execute(ctx, req)
		entry := devicecontrol.AuditEntry{
			Action:    req.Action,
			TargetURL: req.TargetURL,
			Vendor:    req.Vendor,
			Success:   err == nil && res.Success,
			Message:   strings.TrimSpace(res.Message),
		}
		if err != nil && strings.TrimSpace(entry.Message) == "" {
			entry.Message = err.Error()
		}
		auditPath := filepath.Join("audit", "device-actions.log")
		_ = devicecontrol.AppendAudit(auditPath, entry)
		if err != nil {
			return fmt.Sprintf("### Device Control\n\nОшибка: `%v`\n\n- audit: `%s`", err, auditPath), err
		}
		var sb strings.Builder
		sb.WriteString("### Device Control\n\n")
		sb.WriteString(fmt.Sprintf("- action: `%s`\n", strings.TrimSpace(res.Action)))
		sb.WriteString(fmt.Sprintf("- target: `%s`\n", strings.TrimSpace(res.TargetURL)))
		sb.WriteString(fmt.Sprintf("- status_code: `%d`\n", res.StatusCode))
		sb.WriteString(fmt.Sprintf("- result: `%s`\n", strings.TrimSpace(res.Message)))
		sb.WriteString(fmt.Sprintf("- audit: `%s`\n", auditPath))
		return sb.String(), nil
	})
}

func (a *App) runPingTool() {
	host, ok := a.withToolHost()
	if !ok {
		return
	}
	count := 4
	if a.toolsPingCountEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsPingCountEnt.Text)); err == nil && v > 0 {
			count = v
		}
	}
	if count < 1 {
		count = 1
	}
	if count > 50 {
		count = 50
	}
	timeoutSec := 60
	if a.toolsTimeoutEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}
	if timeoutSec <= 0 {
		timeoutSec = 60
	}
	timeout := time.Duration(timeoutSec) * time.Second
	a.runToolOperation("Ping", "Выполняется `ping`...", func(ctx context.Context) (string, error) {
		res, err := nettools.RunPingStructured(ctx, host, count, timeout)
		if err != nil {
			return fmt.Sprintf("### Ping\n\nОшибка: `%s`", nettools.HumanizeToolError(err)), err
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("### Ping: `%s`\n\n", host))
		sb.WriteString(fmt.Sprintf("- Count: `%d`\n", count))
		sb.WriteString(fmt.Sprintf("- Timeout: `%ds`\n", timeoutSec))
		sb.WriteString(fmt.Sprintf("- Sent: `%d`\n", res.Stats.Sent))
		sb.WriteString(fmt.Sprintf("- Received: `%d`\n", res.Stats.Received))
		sb.WriteString(fmt.Sprintf("- Loss: `%.1f%%`\n", res.Stats.PacketLoss))
		if res.Stats.RTTAvg > 0 {
			sb.WriteString(fmt.Sprintf("- RTT min/avg/max: `%s / %s / %s`\n", res.Stats.RTTMin, res.Stats.RTTAvg, res.Stats.RTTMax))
		}
		sb.WriteString("\n#### Raw output\n\n```\n")
		sb.WriteString(res.RawOutput)
		sb.WriteString("\n```")
		return sb.String(), nil
	})
}

func (a *App) runTracerouteTool() {
	host, ok := a.withToolHost()
	if !ok {
		return
	}
	timeoutSec := 60
	if a.toolsTimeoutEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}
	if timeoutSec <= 0 {
		timeoutSec = 60
	}
	maxHops := 30
	if a.toolsTraceHopsEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTraceHopsEnt.Text)); err == nil && v > 0 {
			maxHops = v
		}
	}
	if maxHops <= 0 {
		maxHops = 30
	}
	if maxHops > 64 {
		maxHops = 64
	}
	timeout := time.Duration(timeoutSec) * time.Second
	a.runToolOperation("Traceroute", "Выполняется `traceroute`...", func(ctx context.Context) (string, error) {
		res, err := nettools.RunTracerouteStructuredWithMaxHops(ctx, host, timeout, maxHops)
		if err != nil {
			return fmt.Sprintf("### Traceroute\n\nОшибка: `%s`", nettools.HumanizeToolError(err)), err
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("### Traceroute: `%s`\n\n", host))
		sb.WriteString(fmt.Sprintf("- Timeout: `%ds`\n", timeoutSec))
		sb.WriteString(fmt.Sprintf("- Max hops: `%d`\n", maxHops))
		if len(res.Hops) == 0 {
			sb.WriteString("- Hop-данные не распознаны.\n")
		}
		for _, hop := range res.Hops {
			addr := strings.TrimSpace(hop.Address)
			if addr == "" {
				addr = "*"
			}
			if hop.Measurements > 0 {
				sb.WriteString(fmt.Sprintf("- hop `%d`: `%s` (min/avg/max `%s/%s/%s`)\n", hop.Index, addr, hop.RTTMin, hop.RTTAvg, hop.RTTMax))
			} else {
				sb.WriteString(fmt.Sprintf("- hop `%d`: `%s`\n", hop.Index, addr))
			}
		}
		sb.WriteString("\n#### Raw output\n\n```\n")
		sb.WriteString(res.RawOutput)
		sb.WriteString("\n```")
		return sb.String(), nil
	})
}

func (a *App) runDNSTool() {
	host, ok := a.withToolHost()
	if !ok {
		return
	}
	resolver := ""
	if a.toolsDNSResolverEnt != nil {
		resolver = strings.TrimSpace(a.toolsDNSResolverEnt.Text)
	}
	timeoutSec := 60
	if a.toolsTimeoutEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(a.toolsTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}
	if timeoutSec <= 0 {
		timeoutSec = 60
	}
	timeout := time.Duration(timeoutSec) * time.Second
	a.runToolOperation("DNS", "Выполняется DNS lookup...", func(ctx context.Context) (string, error) {
		lookupCtx, cancel := context.WithTimeout(ctx, timeout)
		res, err := nettools.LookupDNSWithResolver(lookupCtx, host, resolver)
		cancel()
		if err != nil {
			return fmt.Sprintf("### DNS\n\nОшибка: `%s`", nettools.HumanizeToolError(err)), err
		}
		var sb strings.Builder
		sb.WriteString("### DNS lookup\n\n")
		sb.WriteString(fmt.Sprintf("- Запрос: `%s`\n", host))
		sb.WriteString(fmt.Sprintf("- Timeout: `%ds`\n", timeoutSec))
		if resolver != "" {
			sb.WriteString(fmt.Sprintf("- Resolver: `%s`\n", resolver))
		}
		if len(res.ForwardIPs) > 0 {
			sb.WriteString("- A/AAAA:\n")
			for _, ip := range res.ForwardIPs {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", strings.TrimSpace(ip)))
			}
		}
		if len(res.ReverseNames) > 0 {
			sb.WriteString("- PTR:\n")
			for _, name := range res.ReverseNames {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", strings.TrimSpace(name)))
			}
		}
		if len(res.ForwardIPs) == 0 && len(res.ReverseNames) == 0 {
			sb.WriteString("- Ответ пустой.\n")
		}
		return sb.String(), nil
	})
}

func (a *App) runWOLTool() {
	if a == nil || a.toolsWOLMacEntry == nil {
		return
	}
	mac := strings.TrimSpace(a.toolsWOLMacEntry.Text)
	if mac == "" {
		dialog.ShowInformation("Wake-on-LAN", "Введите MAC адрес", a.myWindow)
		return
	}
	broadcast := ""
	if a.toolsWOLBcastEntry != nil {
		broadcast = strings.TrimSpace(a.toolsWOLBcastEntry.Text)
	}
	iface := ""
	if a.toolsWOLIfaceEntry != nil {
		iface = strings.TrimSpace(a.toolsWOLIfaceEntry.Text)
	}

	a.runToolOperation("Wake-on-LAN", "Отправка Wake-on-LAN magic packet...", func(ctx context.Context) (string, error) {
		target, err := wol.SendMagicPacketWithInterface(mac, broadcast, iface)
		if err != nil {
			return fmt.Sprintf("### Wake-on-LAN\n\nОшибка: `%v`", err), err
		}
		var sb strings.Builder
		sb.WriteString("### Wake-on-LAN\n\n")
		sb.WriteString(fmt.Sprintf("- MAC: `%s`\n", mac))
		sb.WriteString(fmt.Sprintf("- Broadcast: `%s`\n", target))
		if strings.TrimSpace(iface) != "" {
			sb.WriteString(fmt.Sprintf("- Interface: `%s`\n", iface))
		}
		sb.WriteString("- Статус: magic packet отправлен")
		return sb.String(), nil
	})
}

func partialSNMPKeysFromReport(report *snmpcollector.CollectReport) map[string]struct{} {
	if report == nil {
		return nil
	}
	out := make(map[string]struct{})
	for _, f := range report.Failures {
		if f.Kind != snmpcollector.FailureQuery {
			continue
		}
		ip := strings.TrimSpace(strings.ToLower(f.IP))
		if ip != "" {
			out["ip:"+ip] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func formatDurationMMSS(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSec := int(d.Round(time.Second).Seconds())
	min := totalSec / 60
	sec := totalSec % 60
	return fmt.Sprintf("%02d:%02d", min, sec)
}

// stopScan — migrated to scanCtrl.StopScan
// startScan — migrated to scanCtrl.StartScan
func (a *App) resultsForSave() ([]scanner.Result, string) {
	if len(a.scanResults) == 0 {
		return nil, "Нет результатов для сохранения"
	}
	resultsToSave := a.currentDisplayedResults()
	if len(resultsToSave) == 0 {
		return nil, "После применения фильтров нет данных для сохранения"
	}
	return resultsToSave, ""
}

// saveResults сохраняет результаты в файл
func (a *App) saveResults() {
	resultsToSave, reason := a.resultsForSave()
	if reason != "" {
		dialog.ShowInformation("Информация", reason, a.myWindow)
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.myWindow)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// Форматируем результаты в текстовый формат
		text := display.FormatResultsAsText(resultsToSave)

		_, err = writer.Write([]byte(text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка при сохранении файла: %v", err), a.myWindow)
			return
		}

		dialog.ShowInformation("Успех", fmt.Sprintf("Результаты успешно сохранены (устройств: %d)", len(resultsToSave)), a.myWindow)
	}, a.myWindow)
}

// buildTopology — migrated to topoCtrl.BuildTopology
// stopTopologyBuild — migrated to topoCtrl.StopTopologyBuild
// copyPerformanceReport — migrated to topoCtrl.CopyPerformanceReport
func (a *App) copyScanDiagnostics() {
	if a == nil || a.diagnosticsLabel == nil {
		return
	}
	text := strings.TrimSpace(a.diagnosticsLabel.Text)
	if text == "" || strings.Contains(strings.ToLower(text), "n/a") || strings.Contains(strings.ToLower(text), "выполняется") {
		dialog.ShowInformation("Информация", "Диагностика сканирования пока недоступна", a.myWindow)
		return
	}
	a.myWindow.Clipboard().SetContent(text)
	dialog.ShowInformation("Готово", "Диагностика сканирования скопирована в буфер обмена", a.myWindow)
}

func (a *App) saveScanDiagnostics() {
	if a == nil || a.diagnosticsLabel == nil {
		return
	}
	text := strings.TrimSpace(a.diagnosticsLabel.Text)
	if text == "" || strings.Contains(strings.ToLower(text), "n/a") || strings.Contains(strings.ToLower(text), "выполняется") {
		dialog.ShowInformation("Информация", "Диагностика сканирования пока недоступна", a.myWindow)
		return
	}

	defaultFileName := fmt.Sprintf("scan-diagnostics-%s.txt", time.Now().Format("2006-01-02-150405"))
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.myWindow)
			return
		}
		if writer == nil {
			return
		}
		targetPath := writer.URI().Path()
		normalizedPath := targetPath
		if strings.ToLower(filepath.Ext(normalizedPath)) != ".txt" {
			normalizedPath += ".txt"
		}

		if normalizedPath == targetPath {
			defer writer.Close()
			if _, writeErr := writer.Write([]byte(text)); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении диагностики: %v", writeErr), a.myWindow)
				return
			}
		} else {
			_ = writer.Close()
			if writeErr := os.WriteFile(normalizedPath, []byte(text), 0644); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении диагностики: %v", writeErr), a.myWindow)
				return
			}
		}
		dialog.ShowInformation("Готово", fmt.Sprintf("Диагностика сканирования сохранена: %s", normalizedPath), a.myWindow)
	}, a.myWindow)
	saveDialog.SetFileName(defaultFileName)
	saveDialog.Show()
}

func (a *App) buildPerformanceReportText() string {
	if a.lastTopology == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Отчет производительности topology build\n")
	sb.WriteString(fmt.Sprintf("Устройств: %d\n", len(a.lastTopology.Devices)))
	sb.WriteString(fmt.Sprintf("Связей: %d\n", len(a.lastTopology.Links)))
	if a.lastTopoMetric.snmpDuration > 0 {
		sb.WriteString(fmt.Sprintf("SNMP сбор: %s\n", a.lastTopoMetric.snmpDuration.Round(time.Millisecond).String()))
	}
	if a.lastTopoMetric.buildDuration > 0 {
		sb.WriteString(fmt.Sprintf("Построение графа: %s\n", a.lastTopoMetric.buildDuration.Round(time.Millisecond).String()))
	}
	if a.lastTopoMetric.totalDuration > 0 {
		sb.WriteString(fmt.Sprintf("Общее время: %s\n", a.lastTopoMetric.totalDuration.Round(time.Millisecond).String()))
	}
	if a.lastSNMPReport != nil {
		sb.WriteString(fmt.Sprintf("SNMP целей: %d\n", a.lastSNMPReport.TotalSNMPTargets))
		sb.WriteString(fmt.Sprintf("SNMP ok: %d\n", a.lastSNMPReport.Connected))
		sb.WriteString(fmt.Sprintf("SNMP partial: %d\n", a.lastSNMPReport.Partial))
		sb.WriteString(fmt.Sprintf("SNMP failed: %d\n", a.lastSNMPReport.Failed))
	}
	return sb.String()
}

// savePerformanceReport — migrated to topoCtrl.SavePerformanceReport
// saveTopology — migrated to topoCtrl.SaveTopology
func (a *App) resetUIPanelLayoutWithFeedback() {
	if a == nil {
		return
	}
	a.settingsMgr.ResetUIPanelLayout(a.scanTabMainSplit, a.topologyMainSplit, a.toolsTabMainSplit)
	if a.myWindow != nil {
		dialog.ShowInformation("Вид", layoutResetInfoMessage, a.myWindow)
	}
}

func (a *App) resetUIPanelLayout() {
	if a == nil || a.myApp == nil {
		return
	}
	a.settingsMgr.ClearSplitPreferences()

	a.rememberedHostDetailsSplitV = 0
	a.rememberedHostDetailsSplitH = 0
	a.lastHostDetailsSplitKind = ""
	a.hostDetailsSplitPrimedV = false
	a.hostDetailsSplitPrimedH = false

	prof := strings.TrimSpace(a.layoutProfile)
	if prof == "" {
		prof = "normal"
	}
	if a.myWindow != nil {
		fyne.Do(func() {
			a.settingsMgr.ApplyDefaultSplitOffsetsForProfile(prof)
			a.renderScanResultsView()
			if a.myWindow.Content() != nil {
				a.myWindow.Content().Refresh()
			}
		})
	} else {
		a.settingsMgr.ApplyDefaultSplitOffsetsForProfile(prof)
		a.renderScanResultsView()
	}
}

func (a *App) setupMainMenu() {
	if a == nil || a.myWindow == nil {
		return
	}
	// Загружаем сохраненную тему
	themeMode := a.loadTheme()
	a.applyTheme(themeMode)

	// Элементы меню "Вид"
	resetItem := fyne.NewMenuItem("Сбросить расположение панелей (Ctrl+Shift+L)", func() {
		a.settingsMgr.ResetUIPanelLayoutWithFeedback(a.scanTabMainSplit, a.topologyMainSplit, a.toolsTabMainSplit, a.myWindow)
	})
	viewMenu := fyne.NewMenu("Вид", resetItem)

	// Элементы меню "Тема"
	themeItems := []string{"Светлая", "Тёмная", "Системная"}
	themeMenu := fyne.NewMenu("Тема")
	for _, item := range themeItems {
		mode := ThemeMode(item)
		menuItem := fyne.NewMenuItem(item, func() {
			a.applyTheme(mode)
			a.saveTheme(mode)
			if a.statusLabel != nil {
				a.statusLabel.SetText("Тема изменена: " + item)
			}
		})
		// Отмечаем текущую тему
		if string(mode) == string(themeMode) {
			menuItem.Checked = true
		}
		themeMenu.Items = append(themeMenu.Items, menuItem)
	}

	// Элементы меню "Акцентный цвет"
	accentMenu := fyne.NewMenu("Акцент")
	currentPreset := a.loadAccentPreset()
	for name := range PresetThemes {
		menuItem := fyne.NewMenuItem(name, func(n string) func() {
			return func() {
				a.applyAccentPreset(n)
			}
		}(name))
		if name == currentPreset {
			menuItem.Checked = true
		}
		accentMenu.Items = append(accentMenu.Items, menuItem)
	}

	// Создаем главное меню
	mainMenu := fyne.NewMainMenu(viewMenu, themeMenu, accentMenu)
	a.myWindow.SetMainMenu(mainMenu)
}

func (a *App) setupLayoutResetShortcut() {
	if a == nil || a.myWindow == nil {
		return
	}
	c := a.myWindow.Canvas()
	if c == nil {
		return
	}
	sc := &desktop.CustomShortcut{
		KeyName:  fyne.KeyL,
		Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift,
	}
	c.AddShortcut(sc, func(fyne.Shortcut) {
		a.settingsMgr.ResetUIPanelLayoutWithFeedback(a.scanTabMainSplit, a.topologyMainSplit, a.toolsTabMainSplit, a.myWindow)
	})
}

// Run запускает GUI приложение
func (a *App) Run() {
	a.setupMainMenu()
	a.setupLayoutResetShortcut()
	a.myWindow.SetOnClosed(func() {
		if a.previewPath != "" {
			_ = os.Remove(a.previewPath)
		}
	})
	a.myWindow.ShowAndRun()
}

// ScanActions реализует интерфейс для контроллера сканирования
func (a *App) ApplyScanRunStart(autoProfileNote string) {
	// Реализация остается в App для доступа к internal state
}

func (a *App) ObserveScanRunner(runner *scand.Runner, startTime time.Time, timeout time.Duration) {
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				fyne.Do(func() {
					if a.statusLabel != nil {
						a.statusLabel.SetText(fmt.Sprintf("Критическая ошибка сканирования: %v", rec))
					}
				})
			}
		}()

		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-time.After(timeout):
				fyne.Do(func() {
					if a.statusLabel != nil {
						a.statusLabel.SetText("Таймаут сканирования")
					}
					if a.resultsStateLabel != nil {
						a.resultsStateLabel.SetText(resultsStateTimeout)
					}
					if a.stageLabel != nil {
						a.stageLabel.Hide()
					}
					if a.progressBar != nil {
						a.progressBar.Hide()
					}
					if a.scanButton != nil {
						a.scanButton.Enable()
					}
					if a.stopButton != nil {
						a.stopButton.Disable()
					}
					a.renderScanResultsView()
				})
				return
			case ev, ok := <-runner.Events():
				if !ok {
					return
				}
				switch ev.Kind {
				case scand.EventProgress:
					fyne.Do(func() {
						if a.stageLabel != nil {
							a.stageLabel.Show()
							a.stageLabel.SetText(ev.Stage + ": " + ev.Message)
						}
						if a.progressBar != nil {
							a.progressBar.Show()
							a.progressBar.SetValue(ev.Percent * 100)
						}
						if a.statusLabel != nil {
							a.statusLabel.SetText(ev.Message)
						}
					})
				case scand.EventDone:
					fyne.Do(func() {
						if a.statusLabel != nil {
							a.statusLabel.SetText(fmt.Sprintf("Сканирование завершено за %v. Найдено устройств: %d", time.Since(startTime), len(ev.Results)))
						}
						if a.resultsStateLabel != nil {
							a.resultsStateLabel.SetText(resultsStateDone)
						}
						if a.stageLabel != nil {
							a.stageLabel.Hide()
						}
						if a.progressBar != nil {
							a.progressBar.Hide()
						}
						if a.scanButton != nil {
							a.scanButton.Enable()
						}
						if a.stopButton != nil {
							a.stopButton.Disable()
						}
						// Обновляем результаты
						a.scanResults = ev.Results
						a.renderScanResultsView()
					})
					return
				case scand.EventError:
					fyne.Do(func() {
						if a.statusLabel != nil {
							a.statusLabel.SetText("Ошибка сканирования: " + ev.Message)
						}
						if a.resultsStateLabel != nil {
							a.resultsStateLabel.SetText(resultsStateStopped)
						}
						if a.stageLabel != nil {
							a.stageLabel.Hide()
						}
						if a.progressBar != nil {
							a.progressBar.Hide()
						}
						if a.scanButton != nil {
							a.scanButton.Enable()
						}
						if a.stopButton != nil {
							a.stopButton.Disable()
						}
					})
					return
				case scand.EventStopped:
					fyne.Do(func() {
						if a.statusLabel != nil {
							a.statusLabel.SetText(ev.Message)
						}
						if a.resultsStateLabel != nil {
							a.resultsStateLabel.SetText(resultsStateStopped)
						}
						if a.stageLabel != nil {
							a.stageLabel.Hide()
						}
						if a.progressBar != nil {
							a.progressBar.Hide()
						}
						if a.scanButton != nil {
							a.scanButton.Enable()
						}
						if a.stopButton != nil {
							a.stopButton.Disable()
						}
						a.renderScanResultsView()
					})
					return
				}
			case <-ticker.C:
				// Проверяем, не остановлен ли runner
				if !runner.IsRunning() {
					return
				}
			}
		}
	}()
}

func (a *App) RenderScanResultsView() {
	a.renderScanResultsView()
}

func (a *App) ConfirmLargeScanBypass() bool {
	return a.confirmLargeScanBypass
}

func (a *App) SetConfirmLargeScanBypass(val bool) {
	a.confirmLargeScanBypass = val
}
