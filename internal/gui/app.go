package gui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"network-scanner/internal/builder"
	"network-scanner/internal/devicecontrol"
	"network-scanner/internal/gui/controller"
	"network-scanner/internal/inventory"
	"network-scanner/internal/logger"
	"network-scanner/internal/network"
	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

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
	networkEntry                *widget.Entry
	networkEntryBox             *fyne.Container
	networkEntryMsg             *canvas.Text
	portRangeEntry              *widget.Entry
	portRangeEntryBox           *fyne.Container
	portRangeEntryMsg           *canvas.Text
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
	scanAdvancedAccordion       *widget.Accordion
	scanAdvancedOpen            bool
	resultsFiltersAccordion     *widget.Accordion
	resultsFiltersHidden        bool
	statusToastTimer            *time.Timer
	statusToastTimerMu          sync.Mutex
	timeoutEntryBox             *fyne.Container
	timeoutEntryMsg             *canvas.Text
	threadsEntryBox             *fyne.Container
	threadsEntryMsg             *canvas.Text

	// Controllers (H2 Refactoring)
	scanCtrl    *controller.ScanController
	resultsCtrl *controller.ResultsController
	topoCtrl    *controller.TopologyController
	toolsCtrl   *controller.ToolsController
	settingsMgr *controller.SettingsManager
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
	piePalette = []color.RGBA{
		{R: 37, G: 99, B: 235, A: 255},
		{R: 59, G: 130, B: 246, A: 255},
		{R: 14, G: 165, B: 233, A: 255},
		{R: 57, G: 112, B: 168, A: 255},
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

	app.initUI()
	app.initControllers() // ← Инициализация ПОСЛЕ создания UI
	app.setupEventHandlers()
	app.loadScanSettings()
	app.autoDetectNetwork()

	logger.Log("GUI приложение инициализировано")

	return app
}

// initControllers инициализирует контроллеры после создания UI виджетов
func (a *App) initControllers() {
	a.scanCtrl = controller.NewScanController(a.myApp, &controller.ScanUI{
		NetworkEntry:         a.networkEntry,
		PortRangeEntry:       a.portRangeEntry,
		TimeoutEntry:         a.timeoutEntry,
		ThreadsEntry:         a.threadsEntry,
		ScanUDPCheck:         a.scanUDPCheck,
		ScanBannersCheck:     a.scanBannersCheck,
		ScanOSActiveCheck:    a.scanOSActiveCheck,
		ScanVerboseLogsCheck: a.scanVerboseLogsCheck,
		ScanTCPPortsCheck:    a.scanTCPPortsCheck,
		AutoProfileCheck:     a.autoProfileCheck,
		StatusLabel:          a.statusLabel,
		RecommendedBadge:     a.recommendedProfileBadge,
		PresetQuickBtn:       a.presetQuickBtn,
		PresetBalBtn:         a.presetBalBtn,
		PresetDeepBtn:        a.presetDeepBtn,
		RecommendedBtn:       a.recommendedProfileBtn,
		ScanButton:           a.scanButton,
		StopButton:           a.stopButton,
		StageLabel:           a.stageLabel,
		ProgressBar:          a.progressBar,
		ResultsStateLabel:    a.resultsStateLabel,
		CopyDiagnosticsBtn:   a.copyDiagnosticsBtn,
		SaveDiagnosticsBtn:   a.saveDiagnosticsBtn,
		MainToolbar:          a.mainToolbar,
		Window:               a.myWindow,
	}, a)
	a.resultsCtrl = controller.NewResultsController(a.myApp, &controller.ResultsUI{
		ResultsModeSel:       a.resultsModeSel,
		ResultsSubModeSel:    a.resultsSubModeSel,
		ResultsSortSel:       a.resultsSortSel,
		ResultsFilterEnt:     a.resultsFilterEnt,
		ResultsCidrFilterEnt: a.resultsCidrFilterEnt,
		ResultsPortStateSel:  a.resultsPortStateSel,
		ChipLimitSel:         a.chipLimitSel,
		ShowRawBannersCheck:  a.showRawBannersCheck,
		OpenPortsOnlyCheck:   a.openPortsOnlyCheck,
		QuickTypeChecks:      a.quickTypeChecks,
		StatusLabel:          a.statusLabel,
	})
	a.topoCtrl = controller.NewTopologyController(&controller.TopologyUI{
		SNMPCommEntry:     a.snmpCommEntry,
		SNMPTimeoutEnt:    a.snmpTimeoutEnt,
		BuildTopoBtn:      a.buildTopoBtn,
		StopTopoBtn:       a.stopTopoBtn,
		SaveTopoBtn:       a.saveTopoBtn,
		CopyPerfBtn:       a.copyPerfBtn,
		SavePerfBtn:       a.savePerfBtn,
		TopoText:          a.topologyText,
		TopoStatus:        a.topologyStatus,
		SNMPStageLabel:    a.snmpStageLabel,
		SNMPProgress:      a.snmpProgress,
		TopoSearchEntry:   a.topologySearchEntry,
		TopoTypeFilterSel: a.topologyTypeFilterSel,
		TopoConfFilterSel: a.topologyConfidenceFilterSel,
		TopoResetMapBtn:   a.topologyResetMapBtn,
		TopoGraphStatus:   a.topologyGraphStatus,
		TopoImage:         a.topologyImage,
		ZoomSelect:        a.zoomSelect,
		RefreshPreviewBtn: a.refreshPreviewBtn,
		OpenPreviewBtn:    a.openPreviewBtn,
	})
	a.toolsCtrl = controller.NewToolsController(a.myApp, &controller.ToolsUI{
		HostEntry:           a.toolsHostEntry,
		PingCountEnt:        a.toolsPingCountEnt,
		TimeoutEnt:          a.toolsTimeoutEnt,
		TraceHopsEnt:        a.toolsTraceHopsEnt,
		DNSResolverEnt:      a.toolsDNSResolverEnt,
		WOLMacEntry:         a.toolsWOLMacEntry,
		WOLBcastEntry:       a.toolsWOLBcastEntry,
		WOLIfaceEntry:       a.toolsWOLIfaceEntry,
		DeviceTargetEntry:   a.toolsDeviceTargetEntry,
		DeviceVendorEntry:   a.toolsDeviceVendorEntry,
		DeviceUserEntry:     a.toolsDeviceUserEntry,
		DevicePassEntry:     a.toolsDevicePassEntry,
		AuditMinSeveritySel: a.toolsAuditMinSeveritySel,
		PingBtn:             a.toolsPingBtn,
		TraceBtn:            a.toolsTraceBtn,
		DNSBtn:              a.toolsDNSBtn,
		WhoisBtn:            a.toolsWhoisBtn,
		WiFiBtn:             a.toolsWiFiBtn,
		AuditBtn:            a.toolsAuditBtn,
		RiskBtn:             a.toolsRiskBtn,
		WOLBtn:              a.toolsWOLBtn,
		DeviceStatusBtn:     a.toolsDeviceStatusBtn,
		DeviceRebootBtn:     a.toolsDeviceRebootBtn,
		ToolsOutput:         a.toolsOutput,
		OperationsOutput:    a.operationsOutput,
		OperationsSelect:    a.operationsSelect,
		OperationsRetryBtn:  a.operationsRetryBtn,
		OperationsCancelBtn: a.operationsCancelBtn,
		StatusLabel:         a.statusLabel,
	})
	a.settingsMgr = controller.NewSettingsManager(a.myApp)
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
				a.setStatus(fmt.Sprintf("Сеть определена автоматически: %s", networkStr))
				a.topologyStatus.SetText(fmt.Sprintf("Сеть определена автоматически: %s", networkStr))
			} else {
				// Если не удалось определить, оставляем поле пустым
				a.setStatus("Готов к сканированию (сеть будет определена автоматически при запуске)")
				a.topologyStatus.SetText("Готово к построению топологии после сканирования")
			}
			// Обновляем виджеты
			a.statusLabel.Refresh()
			a.topologyStatus.Refresh()
			a.networkEntry.Refresh()
		})
	}()
}

// Run запускает GUI приложение
func (a *App) Run() {
	a.setupMainMenu()
	a.setupAppShortcuts()
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

func (a *App) RenderScanResultsView() {
	a.renderScanResultsView()
}

func (a *App) ConfirmLargeScanBypass() bool {
	return a.confirmLargeScanBypass
}

func (a *App) SetConfirmLargeScanBypass(val bool) {
	a.confirmLargeScanBypass = val
}
