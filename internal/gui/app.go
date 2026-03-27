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
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/display"
	"network-scanner/internal/logger"
	"network-scanner/internal/network"
	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"
)

// scanUpdate содержит результаты сканирования для обновления UI
type scanUpdate struct {
	results []scanner.Result
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
	myApp          fyne.App
	myWindow       fyne.Window
	scanResults    []scanner.Result
	networkScanner *scanner.NetworkScanner
	networkEntry   *widget.Entry
	portRangeEntry *widget.Entry
	timeoutEntry   *widget.Entry
	threadsEntry   *widget.Entry
	scanUDPCheck   *widget.Check
	presetQuickBtn *widget.Button
	presetBalBtn   *widget.Button
	presetDeepBtn  *widget.Button
	statusLabel    *widget.Label
	stageLabel     *widget.Label
	progressBar    *widget.ProgressBar
	resultsText    *widget.RichText
	resultsScroll  *container.Scroll
	scanButton     *widget.Button
	stopButton     *widget.Button
	saveButton     *widget.Button
	buildTopoBtn   *widget.Button
	stopTopoBtn    *widget.Button
	saveTopoBtn    *widget.Button
	copyPerfBtn    *widget.Button
	savePerfBtn    *widget.Button
	snmpCommEntry  *widget.Entry
	snmpTimeoutEnt *widget.Entry
	lastTopology   *topology.Topology
	lastSNMPReport *snmpcollector.CollectReport
	lastTopoMetric topologyBuildMetrics
	topologyText   *widget.RichText
	topologyScroll *container.Scroll
	topologyStatus *widget.Label
	snmpStageLabel *widget.Label
	snmpProgress   *widget.ProgressBar
	mainTabs       *container.AppTabs
	topologyImage  *canvas.Image
	topologyImgBox *fyne.Container
	topologyImgScroll *container.Scroll
	previewPath    string
	refreshPreviewBtn *widget.Button
	zoomSelect        *widget.Select
	openPreviewBtn    *widget.Button
	topologyCancel    context.CancelFunc
}

const (
	prefNetwork   = "scan.network"
	prefPortRange = "scan.port_range"
	prefTimeout   = "scan.timeout_sec"
	prefThreads   = "scan.threads"
	prefScanUDP   = "scan.udp"
	prefPreset    = "scan.preset"
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
	myApp := app.NewWithID("network-scanner")

	myWindow := myApp.NewWindow("Network Scanner - Сканер локальной сети")

	// Устанавливаем иконку приложения
	if icon := createAppIcon(); icon != nil {
		myApp.SetIcon(icon)
		myWindow.SetIcon(icon)
	}

	// Устанавливаем размер окна, который гарантированно поместится на экране
	// Используем консервативные размеры, которые подходят даже для маленьких экранов
	// (например, ноутбуки с разрешением 1366x768)
	width := float32(950)
	height := float32(700)

	myWindow.Resize(fyne.NewSize(width, height))
	myWindow.CenterOnScreen()

	// Устанавливаем максимальный размер окна, чтобы оно не выходило за границы экрана
	// Fyne автоматически ограничит размер окна размером экрана
	myWindow.SetFixedSize(false) // Позволяем изменять размер, но в пределах экрана

	app := &App{
		myApp:    myApp,
		myWindow: myWindow,
	}

	app.initUI()
	app.setupEventHandlers()
	app.loadScanSettings()
	app.autoDetectNetwork()

	logger.Log("GUI приложение инициализировано")

	return app
}

// initUI инициализирует все элементы интерфейса
func (a *App) initUI() {
	// Поле ввода сети
	networkLabel := widget.NewLabel("Сеть (CIDR, например 192.168.1.0/24):")
	networkLabel.Wrapping = fyne.TextWrapWord
	a.networkEntry = widget.NewEntry()
	a.networkEntry.SetPlaceHolder("Оставьте пустым для автоматического определения")
	a.portRangeEntry = widget.NewEntry()
	a.portRangeEntry.SetText("1-1000")
	a.timeoutEntry = widget.NewEntry()
	a.timeoutEntry.SetText("2")
	a.threadsEntry = widget.NewEntry()
	a.threadsEntry.SetText("50")
	a.scanUDPCheck = widget.NewCheck("Включить UDP сканирование", nil)
	a.presetQuickBtn = widget.NewButton("Быстро", nil)
	a.presetBalBtn = widget.NewButton("Баланс", nil)
	a.presetDeepBtn = widget.NewButton("Глубоко", nil)

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

	// Метка этапа сканирования
	a.stageLabel = widget.NewLabel("")
	a.stageLabel.Wrapping = fyne.TextWrapWord
	a.stageLabel.Hide()

	// Прогресс-бар
	a.progressBar = widget.NewProgressBar()
	a.progressBar.Hide()

	// Область результатов с прокруткой
	a.resultsText = widget.NewRichText()
	a.resultsText.Wrapping = fyne.TextWrapWord

	// Создаем прокручиваемый контейнер для результатов
	// Это ключевое изменение - используем Scroll контейнер для прокрутки результатов
	a.resultsScroll = container.NewScroll(a.resultsText)
	// Устанавливаем минимальный размер, чтобы контейнер был прокручиваемым
	// и занимал доступное пространство
	a.resultsScroll.SetMinSize(fyne.NewSize(0, 300))

	// Верхняя панель сканирования
	scanControlsContainer := container.NewVBox(
		networkLabel,
		a.networkEntry,
		widget.NewLabel("Диапазон TCP портов (например 1-1000):"),
		a.portRangeEntry,
		container.NewHBox(
			widget.NewLabel("Пресет:"),
			a.presetQuickBtn,
			a.presetBalBtn,
			a.presetDeepBtn,
		),
		container.NewHBox(
			widget.NewLabel("Таймаут (сек):"),
			a.timeoutEntry,
			widget.NewLabel("Потоки:"),
			a.threadsEntry,
		),
		a.scanUDPCheck,
		container.NewHBox(a.scanButton, a.stopButton, a.saveButton),
		a.statusLabel,
		a.stageLabel,
		a.progressBar,
	)

	// Разделитель
	separator := widget.NewSeparator()

	// Заголовок результатов
	resultsLabel := widget.NewLabel("Результаты сканирования:")
	resultsLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Контейнер с результатами (заголовок + прокручиваемая область)
	resultsContainer := container.NewBorder(
		container.NewVBox(separator, resultsLabel),
		nil,
		nil,
		nil,
		a.resultsScroll, // Прокручиваемый контейнер с результатами
	)

	// Вкладка сканирования
	scanTabContent := container.NewBorder(
		scanControlsContainer,
		nil,
		nil,
		nil,
		resultsContainer,
	)

	// Вкладка топологии: отдельный экран с настройками и превью
	a.topologyText = widget.NewRichText()
	a.topologyText.Wrapping = fyne.TextWrapWord
	a.topologyText.ParseMarkdown("## Топология сети\n\nСначала выполните сканирование, затем нажмите **Построить топологию**.")
	a.topologyScroll = container.NewScroll(a.topologyText)
	a.topologyScroll.SetMinSize(fyne.NewSize(0, 350))
	a.topologyImage = canvas.NewImageFromResource(nil)
	a.topologyImage.FillMode = canvas.ImageFillContain
	a.topologyImage.SetMinSize(fyne.NewSize(0, 260))
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
		a.snmpStageLabel,
		a.snmpProgress,
		a.topologyStatus,
	)
	topologyTabContent := container.NewBorder(
		topologyControls,
		nil,
		nil,
		nil,
		container.NewVSplit(a.topologyImgScroll, a.topologyScroll),
	)

	a.mainTabs = container.NewAppTabs(
		container.NewTabItem("Сканирование", scanTabContent),
		container.NewTabItem("Топология", topologyTabContent),
	)
	a.myWindow.SetContent(a.mainTabs)
}

// setupEventHandlers настраивает обработчики событий
func (a *App) setupEventHandlers() {
	a.scanButton.OnTapped = func() {
		a.startScan()
	}
	a.stopButton.OnTapped = func() {
		a.stopScan()
	}
	a.presetQuickBtn.OnTapped = func() {
		a.applyScanPreset("quick")
	}
	a.presetBalBtn.OnTapped = func() {
		a.applyScanPreset("balanced")
	}
	a.presetDeepBtn.OnTapped = func() {
		a.applyScanPreset("deep")
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
	a.scanUDPCheck.OnChanged = func(_ bool) {
		a.saveScanSettings()
	}

	a.saveButton.OnTapped = func() {
		a.saveResults()
	}
	a.buildTopoBtn.OnTapped = func() {
		a.buildTopology()
	}
	a.stopTopoBtn.OnTapped = func() {
		a.stopTopologyBuild()
	}
	a.saveTopoBtn.OnTapped = func() {
		a.saveTopology()
	}
	a.copyPerfBtn.OnTapped = func() {
		a.copyPerformanceReport()
	}
	a.savePerfBtn.OnTapped = func() {
		a.savePerformanceReport()
	}
	a.refreshPreviewBtn.OnTapped = func() {
		a.refreshTopologyPreview()
	}
	a.openPreviewBtn.OnTapped = func() {
		a.openPreviewExternal()
	}
	a.zoomSelect.OnChanged = func(value string) {
		a.applyTopologyZoom(value)
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

func (a *App) applyScanPreset(mode string) {
	switch mode {
	case "quick":
		// Быстрый обзор сети: минимальная глубина, максимальная скорость.
		a.portRangeEntry.SetText("22,80,443,445,3389")
		a.timeoutEntry.SetText("1")
		a.threadsEntry.SetText("120")
		a.scanUDPCheck.SetChecked(false)
		a.statusLabel.SetText("Пресет: Быстро (обзор)")
	case "deep":
		// Глубокий анализ: больше портов и выше таймаут для точности.
		a.portRangeEntry.SetText("1-2000")
		a.timeoutEntry.SetText("3")
		a.threadsEntry.SetText("40")
		a.scanUDPCheck.SetChecked(true)
		a.statusLabel.SetText("Пресет: Глубоко (детальный анализ)")
	default:
		// Баланс между скоростью и полнотой.
		a.portRangeEntry.SetText("1-1000")
		a.timeoutEntry.SetText("2")
		a.threadsEntry.SetText("50")
		a.scanUDPCheck.SetChecked(false)
		a.statusLabel.SetText("Пресет: Баланс")
	}
	a.myApp.Preferences().SetString(prefPreset, mode)
	a.saveScanSettings()
	a.portRangeEntry.Refresh()
	a.timeoutEntry.Refresh()
	a.threadsEntry.Refresh()
	a.scanUDPCheck.Refresh()
	a.statusLabel.Refresh()
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

	switch strings.TrimSpace(p.String(prefPreset)) {
	case "quick":
		a.statusLabel.SetText("Пресет: Быстро (восстановлен)")
	case "deep":
		a.statusLabel.SetText("Пресет: Глубоко (восстановлен)")
	case "balanced":
		a.statusLabel.SetText("Пресет: Баланс (восстановлен)")
	}
}

// startScan запускает процесс сканирования
func (a *App) startScan() {
	scanStartTime := time.Now()
	logger.Log("Запуск сканирования из GUI")
	logger.LogDebug("Пользователь нажал кнопку 'Запустить сканирование'")

	// Определяем сеть
	networkStr := a.networkEntry.Text
	if networkStr == "" {
		logger.Log("Автоматическое определение сети...")
		logger.LogDebug("Поле сети пустое, начинаем автоматическое определение")
		detectStartTime := time.Now()
		var err error
		networkStr, err = network.DetectLocalNetwork()
		detectDuration := time.Since(detectStartTime)
		if err != nil {
			logger.LogError(err, "Определение сети в GUI")
			logger.LogDebug("Автоматическое определение сети завершилось ошибкой за %v", detectDuration)
			dialog.ShowError(fmt.Errorf("не удалось определить сеть: %v", err), a.myWindow)
			return
		}
		a.networkEntry.SetText(networkStr)
		logger.Log("Определена сеть: %s (определение заняло %v)", networkStr, detectDuration)
		logger.LogDebug("Автоматическое определение сети завершено успешно")
	} else {
		logger.Log("Использована указанная сеть: %s", networkStr)
		logger.LogDebug("Сеть указана пользователем в поле ввода")
	}

	// Отключаем кнопку сканирования
	a.scanButton.Disable()
	a.stopButton.Enable()
	a.saveButton.Disable()
	a.progressBar.Show()
	a.progressBar.SetValue(0)
	a.stageLabel.Show()
	a.statusLabel.SetText("Сканирование запущено...")
	a.topologyStatus.SetText("Ожидание завершения сканирования...")
	a.stageLabel.SetText("Инициализация...")

	// Очищаем предыдущие результаты
	a.resultsText.ParseMarkdown("## Сканирование запущено...\n\nПожалуйста, подождите.")
	a.resultsScroll.Refresh()
	a.topologyText.ParseMarkdown("## Топология сети\n\nСканирование выполняется. После завершения станет доступно построение топологии.")
	a.topologyScroll.Refresh()

	// Создаем канал для передачи результатов из горутины
	resultsChan := make(chan []scanner.Result, 1)
	progressChan := make(chan progressUpdate, 100) // Буферизованный канал для прогресса

	// Запускаем сканирование в отдельной горутине
	go func() {
		// Создаем сканер с параметрами из UI
		timeoutSec := 2
		if v, err := strconv.Atoi(strings.TrimSpace(a.timeoutEntry.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
		portRange := strings.TrimSpace(a.portRangeEntry.Text)
		if portRange == "" {
			portRange = "1-1000"
		}
		threads := 50
		if v, err := strconv.Atoi(strings.TrimSpace(a.threadsEntry.Text)); err == nil && v > 0 {
			threads = v
		}
		showClosed := false
		scanUDP := a.scanUDPCheck.Checked

		logger.LogDebug("Создание сканера в GUI: сеть=%s, порты=%s, таймаут=%v, потоков=%d, showClosed=%v",
			networkStr, portRange, time.Duration(timeoutSec)*time.Second, threads, showClosed)
		ns := scanner.NewNetworkScanner(networkStr, time.Duration(timeoutSec)*time.Second, portRange, threads, showClosed)
		ns.SetScanUDP(scanUDP)
		a.networkScanner = ns

		// Устанавливаем callback для прогресса
		ns.SetProgressCallback(func(stage string, current int, total int, message string) {
			// Вычисляем процент прогресса для текущего этапа (0-100%)
			var percent float64
			if total > 0 {
				percent = float64(current) / float64(total)
			} else {
				percent = 0
			}

			// Отправляем обновление прогресса в канал
			select {
			case progressChan <- progressUpdate{
				stage:   stage,
				current: current,
				total:   total,
				message: message,
				percent: percent,
			}:
			default:
				// Пропускаем если канал переполнен (не критично)
			}
		})

		// Запускаем сканирование
		logger.LogDebug("Запуск метода Scan() в GUI")
		scanMethodStartTime := time.Now()
		ns.Scan()
		scanMethodDuration := time.Since(scanMethodStartTime)
		logger.LogDebug("Метод Scan() завершен в GUI за %v", scanMethodDuration)

		// Получаем результаты
		results := ns.GetResults()
		logger.Log("Получено результатов сканирования в GUI: %d", len(results))
		logger.LogDebug("Результаты сканирования получены из сканера")

		// Закрываем канал прогресса
		close(progressChan)
		logger.LogDebug("Канал прогресса закрыт")

		// Отправляем результаты в канал
		resultsChan <- results
		totalDuration := time.Since(scanStartTime)
		logger.Log("Сканирование в GUI завершено за %v, найдено устройств: %d", totalDuration, len(results))
	}()

	// Обрабатываем результаты и прогресс в отдельной горутине
	go func() {
		// Создаем тикер для периодического обновления UI
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		timeout := time.NewTimer(300 * time.Second)
		defer timeout.Stop()
		stageStartedAt := map[string]time.Time{}

		for {
			select {
			case progress, ok := <-progressChan:
				if !ok {
					// Канал прогресса закрыт, продолжаем ждать результаты
					progressChan = nil
					continue
				}
				if _, exists := stageStartedAt[progress.stage]; !exists {
					stageStartedAt[progress.stage] = time.Now()
				}
				etaText := ""
				if progress.total > 0 && progress.current > 0 && progress.current < progress.total {
					elapsed := time.Since(stageStartedAt[progress.stage])
					if elapsed > 0 {
						remainingItems := progress.total - progress.current
						eta := time.Duration(float64(elapsed) * (float64(remainingItems) / float64(progress.current)))
						if eta < 0 {
							eta = 0
						}
						etaText = fmt.Sprintf(", ETA ~ %s", formatDurationMMSS(eta))
					}
				}
				// Обновляем UI через fyne.Do для выполнения в главном потоке
				fyne.Do(func() {
					// Определяем название этапа на русском
					stageName := ""
					switch progress.stage {
					case "ping":
						stageName = "Этап 1: Проверка доступности хостов"
					case "ports":
						stageName = "Этап 2: Сканирование портов"
					case "complete":
						stageName = "Завершение"
					default:
						stageName = "Сканирование"
					}

					// Обновляем прогресс-бар (0-100% для текущего этапа)
					a.progressBar.SetValue(progress.percent)

					// Формируем текст для метки этапа с процентом
					if progress.total > 0 {
						percentText := fmt.Sprintf("%.1f%%", progress.percent*100)
						a.stageLabel.SetText(fmt.Sprintf("%s: %d/%d (%s%s)", stageName, progress.current, progress.total, percentText, etaText))
					} else {
						a.stageLabel.SetText(stageName)
					}

					// Обновляем статус
					a.statusLabel.SetText(progress.message)

					// Обновляем виджеты
					a.progressBar.Refresh()
					a.stageLabel.Refresh()
					a.statusLabel.Refresh()
				})

			case results, ok := <-resultsChan:
				if !ok {
					return
				}
				// Обновляем UI через fyne.Do для выполнения в главном потоке
				fyne.Do(func() {
					a.scanResults = results
					a.progressBar.SetValue(1.0)
					a.progressBar.Hide()
					a.stageLabel.Hide()

					if len(results) == 0 {
						a.statusLabel.SetText("Сканирование завершено. Результаты не найдены.")
						a.resultsText.ParseMarkdown("## Результаты сканирования\n\nРезультаты сканирования не найдены.")
						a.topologyStatus.SetText("Нет результатов сканирования для построения топологии")
					} else {
						a.statusLabel.SetText(fmt.Sprintf("Сканирование завершено. Найдено устройств: %d", len(results)))
						formattedResults := FormatResultsForDisplay(results)
						a.resultsText.ParseMarkdown(formattedResults)
						a.saveButton.Enable()
						a.buildTopoBtn.Enable()
						a.topologyStatus.SetText("Можно строить топологию: перейдите на вкладку 'Топология'")
					}

					// Прокручиваем к началу результатов и обновляем отображение
					a.resultsScroll.ScrollToTop()
					a.resultsScroll.Refresh()
					a.resultsText.Refresh()
					a.progressBar.Refresh()
					a.stageLabel.Refresh()
					a.statusLabel.Refresh()
					a.topologyStatus.Refresh()
					a.scanButton.Enable()
					a.stopButton.Disable()
					a.networkScanner = nil
					a.myWindow.Content().Refresh()
				})
				return

			case <-ticker.C:
				// Периодически обновляем UI для обеспечения отзывчивости
				if progressChan != nil {
					fyne.Do(func() {
						a.progressBar.Refresh()
						a.stageLabel.Refresh()
						a.statusLabel.Refresh()
					})
				}

			case <-timeout.C:
				// Таймаут на случай если сканирование зависло
				fyne.Do(func() {
					a.statusLabel.SetText("Таймаут сканирования")
					a.stageLabel.Hide()
					a.scanButton.Enable()
					a.stopButton.Disable()
					a.networkScanner = nil
					a.progressBar.Hide()
					a.statusLabel.Refresh()
					a.stageLabel.Refresh()
					a.progressBar.Refresh()
				})
				return
			}
		}
	}()
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

func (a *App) stopScan() {
	if a.networkScanner == nil {
		return
	}
	a.networkScanner.Stop()
	a.statusLabel.SetText("Сканирование остановлено пользователем")
	a.stageLabel.Hide()
	a.progressBar.Hide()
	a.scanButton.Enable()
	a.stopButton.Disable()
	a.networkScanner = nil
	a.statusLabel.Refresh()
	a.stageLabel.Refresh()
	a.progressBar.Refresh()
}

// saveResults сохраняет результаты в файл
func (a *App) saveResults() {
	if len(a.scanResults) == 0 {
		dialog.ShowInformation("Информация", "Нет результатов для сохранения", a.myWindow)
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
		text := display.FormatResultsAsText(a.scanResults)

		_, err = writer.Write([]byte(text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка при сохранении файла: %v", err), a.myWindow)
			return
		}

		dialog.ShowInformation("Успех", "Результаты успешно сохранены", a.myWindow)
	}, a.myWindow)
}

func (a *App) buildTopology() {
	if len(a.scanResults) == 0 {
		dialog.ShowInformation("Информация", "Сначала выполните сканирование", a.myWindow)
		return
	}
	topologyStartedAt := time.Now()

	a.buildTopoBtn.Disable()
	a.stopTopoBtn.Enable()
	a.saveTopoBtn.Disable()
	a.copyPerfBtn.Disable()
	a.savePerfBtn.Disable()
	a.refreshPreviewBtn.Disable()
	a.openPreviewBtn.Disable()
	a.statusLabel.SetText("Сбор SNMP данных и построение топологии...")
	a.topologyStatus.SetText("Сбор SNMP данных и построение топологии...")
	a.snmpStageLabel.SetText("SNMP: подготовка...")
	a.snmpStageLabel.Show()
	a.snmpProgress.SetValue(0)
	a.snmpProgress.Show()
	a.statusLabel.Refresh()
	a.topologyStatus.Refresh()
	a.snmpStageLabel.Refresh()
	a.snmpProgress.Refresh()

	timeoutSec := 2
	if strings.TrimSpace(a.snmpTimeoutEnt.Text) != "" {
		if v, err := strconv.Atoi(strings.TrimSpace(a.snmpTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}
	communities := splitCommaValues(a.snmpCommEntry.Text)
	snmpStartedAt := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	a.topologyCancel = cancel

	go func() {
		snmpPhaseStartedAt := time.Now()
		snmpData, report, err := snmpcollector.CollectWithReportProgressContext(ctx, a.scanResults, communities, timeoutSec, func(current int, total int, ip string, message string) {
			etaText := ""
			progressValue := 0.0
			if total > 0 && current > 0 && current < total {
				elapsed := time.Since(snmpStartedAt)
				remainingItems := total - current
				eta := time.Duration(float64(elapsed) * (float64(remainingItems) / float64(current)))
				etaText = fmt.Sprintf(", ETA ~ %s", formatDurationMMSS(eta))
			}
			if total > 0 {
				progressValue = float64(current) / float64(total)
				if progressValue > 1 {
					progressValue = 1
				}
			}
			status := fmt.Sprintf("SNMP: %d/%d (%s)%s", current, total, message, etaText)
			if strings.TrimSpace(ip) != "" {
				status = fmt.Sprintf("%s, %s", status, ip)
			}
			fyne.Do(func() {
				a.statusLabel.SetText(status)
				a.topologyStatus.SetText(status)
				a.snmpStageLabel.SetText(status)
				a.snmpProgress.SetValue(progressValue)
				a.statusLabel.Refresh()
				a.topologyStatus.Refresh()
				a.snmpStageLabel.Refresh()
				a.snmpProgress.Refresh()
			})
		})
		if err != nil {
			fyne.Do(func() {
				if err == context.Canceled {
					a.statusLabel.SetText("Построение топологии остановлено пользователем")
					a.topologyStatus.SetText("Построение топологии остановлено пользователем")
					a.snmpStageLabel.SetText("SNMP: остановлено")
					a.snmpProgress.Hide()
					a.buildTopoBtn.Enable()
					a.stopTopoBtn.Disable()
					a.topologyCancel = nil
					a.statusLabel.Refresh()
					a.topologyStatus.Refresh()
					a.snmpStageLabel.Refresh()
					a.snmpProgress.Refresh()
					return
				}
				dialog.ShowError(fmt.Errorf("ошибка SNMP опроса: %v", err), a.myWindow)
				a.buildTopoBtn.Enable()
				a.stopTopoBtn.Disable()
				a.topologyCancel = nil
				a.topologyStatus.SetText("Ошибка SNMP опроса")
				a.snmpStageLabel.SetText("SNMP: ошибка")
				a.snmpProgress.Hide()
				a.topologyStatus.Refresh()
				a.snmpStageLabel.Refresh()
				a.snmpProgress.Refresh()
			})
			return
		}
		snmpDuration := time.Since(snmpPhaseStartedAt)
		buildPhaseStartedAt := time.Now()
		topo, err := topology.BuildTopology(a.scanResults, snmpData)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("ошибка построения топологии: %v", err), a.myWindow)
				a.buildTopoBtn.Enable()
				a.stopTopoBtn.Disable()
				a.topologyCancel = nil
				a.topologyStatus.SetText("Ошибка построения топологии")
				a.snmpStageLabel.SetText("Построение топологии: ошибка")
				a.snmpProgress.Hide()
				a.topologyStatus.Refresh()
				a.snmpStageLabel.Refresh()
				a.snmpProgress.Refresh()
			})
			return
		}
		buildDuration := time.Since(buildPhaseStartedAt)
		metrics := topologyBuildMetrics{
			snmpDuration:  snmpDuration,
			buildDuration: buildDuration,
			totalDuration: time.Since(topologyStartedAt),
		}
		a.renderTopologyImagePreview(topo)
		fyne.Do(func() {
			a.lastTopology = topo
			a.lastSNMPReport = report
			a.lastTopoMetric = metrics
			a.saveTopoBtn.Enable()
			a.copyPerfBtn.Enable()
			a.savePerfBtn.Enable()
			a.refreshPreviewBtn.Enable()
			a.openPreviewBtn.Enable()
			status := fmt.Sprintf("Топология построена: устройств %d, связей %d", len(topo.Devices), len(topo.Links))
			if report != nil {
				status = fmt.Sprintf("%s | SNMP: целей %d, ok %d, partial %d, failed %d",
					status, report.TotalSNMPTargets, report.Connected, report.Partial, report.Failed)
			}
			a.statusLabel.SetText(status)
			a.topologyStatus.SetText(status)
			a.snmpStageLabel.SetText("SNMP: завершено")
			a.snmpProgress.SetValue(1)
			a.snmpProgress.Hide()
			a.topologyText.ParseMarkdown(formatTopologyPreview(topo, report, metrics))
			a.topologyScroll.ScrollToTop()
			a.statusLabel.Refresh()
			a.topologyStatus.Refresh()
			a.snmpStageLabel.Refresh()
			a.snmpProgress.Refresh()
			a.topologyScroll.Refresh()
			a.topologyText.Refresh()
			a.mainTabs.SelectTabIndex(1)
			a.buildTopoBtn.Enable()
			a.stopTopoBtn.Disable()
			a.topologyCancel = nil
		})
	}()
}

func (a *App) stopTopologyBuild() {
	if a.topologyCancel == nil {
		return
	}
	a.topologyCancel()
}

func (a *App) copyPerformanceReport() {
	reportText := a.buildPerformanceReportText()
	if strings.TrimSpace(reportText) == "" {
		dialog.ShowInformation("Информация", "Отчет производительности пока недоступен", a.myWindow)
		return
	}
	a.myWindow.Clipboard().SetContent(reportText)
	dialog.ShowInformation("Готово", "Отчет производительности скопирован в буфер обмена", a.myWindow)
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

func (a *App) savePerformanceReport() {
	reportText := a.buildPerformanceReportText()
	if strings.TrimSpace(reportText) == "" {
		dialog.ShowInformation("Информация", "Отчет производительности пока недоступен", a.myWindow)
		return
	}

	defaultFileName := fmt.Sprintf("topology-performance-%s.txt", time.Now().Format("2006-01-02-150405"))
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

		// Если расширение корректное, записываем через предоставленный writer.
		// Иначе создаем файл с добавленным .txt.
		if normalizedPath == targetPath {
			defer writer.Close()
			if _, writeErr := writer.Write([]byte(reportText)); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении отчета: %v", writeErr), a.myWindow)
				return
			}
		} else {
			_ = writer.Close()
			if writeErr := os.WriteFile(normalizedPath, []byte(reportText), 0644); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении отчета: %v", writeErr), a.myWindow)
				return
			}
		}
		dialog.ShowInformation("Успех", fmt.Sprintf("Отчет производительности сохранен:\n%s", normalizedPath), a.myWindow)
	}, a.myWindow)
	saveDialog.SetFileName(defaultFileName)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
	saveDialog.Show()
}

func (a *App) saveTopology() {
	if a.lastTopology == nil {
		dialog.ShowInformation("Информация", "Сначала постройте топологию", a.myWindow)
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
		path := writer.URI().Path()
		_ = writer.Close()

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".json":
			err = a.lastTopology.SaveJSON(path)
		case ".graphml", ".xml":
			err = a.lastTopology.SaveGraphML(path)
		case ".png":
			err = a.lastTopology.RenderWithGraphviz("png", path)
		case ".svg":
			err = a.lastTopology.RenderWithGraphviz("svg", path)
		default:
			err = fmt.Errorf("поддерживаемые форматы: .json, .graphml, .png, .svg")
		}

		if err != nil {
			dialog.ShowError(err, a.myWindow)
			return
		}
		dialog.ShowInformation("Успех", "Топология сохранена", a.myWindow)
	}, a.myWindow)
}

func splitCommaValues(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"public"}
	}
	return out
}

func formatTopologyPreview(topo *topology.Topology, report *snmpcollector.CollectReport, metrics topologyBuildMetrics) string {
	if topo == nil {
		return "## Топология сети\n\nНет данных для отображения."
	}
	var sb strings.Builder
	sb.WriteString("## Топология сети\n\n")
	if metrics.totalDuration > 0 {
		sb.WriteString("### Время этапов\n\n")
		if metrics.snmpDuration > 0 {
			sb.WriteString(fmt.Sprintf("- SNMP сбор: `%s`\n", metrics.snmpDuration.Round(time.Millisecond).String()))
		}
		if metrics.buildDuration > 0 {
			sb.WriteString(fmt.Sprintf("- Построение графа: `%s`\n", metrics.buildDuration.Round(time.Millisecond).String()))
		}
		sb.WriteString(fmt.Sprintf("- Общее время: `%s`\n\n", metrics.totalDuration.Round(time.Millisecond).String()))
	}
	if report != nil {
		sb.WriteString("### SNMP отчет\n\n")
		sb.WriteString(fmt.Sprintf("- Целей для SNMP: %d\n", report.TotalSNMPTargets))
		sb.WriteString(fmt.Sprintf("- Успешных подключений: %d\n", report.Connected))
		sb.WriteString(fmt.Sprintf("- Частичных опросов: %d\n", report.Partial))
		sb.WriteString(fmt.Sprintf("- Полных отказов: %d\n\n", report.Failed))
	}
	sb.WriteString(fmt.Sprintf("**Устройств:** %d\n\n", len(topo.Devices)))
	sb.WriteString(fmt.Sprintf("**Связей:** %d\n\n", len(topo.Links)))
	sb.WriteString("### Связи\n\n")
	if len(topo.Links) == 0 {
		sb.WriteString("- Связи не найдены.\n")
		return sb.String()
	}
	for _, link := range topo.Links {
		sourceType := strings.TrimSpace(string(link.SourceType))
		confidence := strings.TrimSpace(string(link.Confidence))
		extra := ""
		if sourceType != "" || confidence != "" {
			extra = fmt.Sprintf(" [%s/%s]", sourceType, confidence)
		}
		sb.WriteString(fmt.Sprintf("- `%s (%s)` <-> `%s (%s)`%s\n",
			topoDisplayName(link.Source), topoPortName(link.SourcePort), topoDisplayName(link.Target), topoPortName(link.TargetPort), extra))
	}
	return sb.String()
}

func (a *App) renderTopologyImagePreview(topo *topology.Topology) {
	if topo == nil {
		return
	}
	tmp, err := os.CreateTemp("", "network-topology-preview-*.png")
	if err != nil {
		fyne.Do(func() {
			a.topologyStatus.SetText("Не удалось создать временный файл для превью")
			a.topologyStatus.Refresh()
		})
		return
	}
	previewPath := tmp.Name()
	_ = tmp.Close()

	if err = topo.RenderWithGraphviz("png", previewPath); err != nil {
		_ = os.Remove(previewPath)
		fyne.Do(func() {
			a.topologyStatus.SetText("Графическое превью недоступно (установите Graphviz/dot)")
			a.topologyStatus.Refresh()
		})
		return
	}

	fyne.Do(func() {
		// Удаляем предыдущее изображение превью, если оно было.
		if a.previewPath != "" && a.previewPath != previewPath {
			_ = os.Remove(a.previewPath)
		}
		a.previewPath = previewPath
		img := canvas.NewImageFromFile(previewPath)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(0, 260))
		a.topologyImage = img
		a.applyTopologyZoom(a.zoomSelect.Selected)
		a.topologyImgBox.Objects = []fyne.CanvasObject{a.topologyImage}
		a.topologyImgBox.Refresh()
	})
}

func (a *App) refreshTopologyPreview() {
	if a.lastTopology == nil {
		dialog.ShowInformation("Информация", "Сначала постройте топологию", a.myWindow)
		return
	}
	a.topologyStatus.SetText("Обновление графического превью...")
	a.topologyStatus.Refresh()
	go func() {
		a.renderTopologyImagePreview(a.lastTopology)
		fyne.Do(func() {
			a.topologyStatus.SetText(fmt.Sprintf("Топология построена: устройств %d, связей %d", len(a.lastTopology.Devices), len(a.lastTopology.Links)))
			a.topologyStatus.Refresh()
		})
	}()
}

func (a *App) applyTopologyZoom(mode string) {
	if a.topologyImage == nil {
		return
	}
	switch mode {
	case "200%":
		a.topologyImage.FillMode = canvas.ImageFillOriginal
		a.topologyImage.SetMinSize(fyne.NewSize(2400, 1400))
	case "150%":
		a.topologyImage.FillMode = canvas.ImageFillOriginal
		a.topologyImage.SetMinSize(fyne.NewSize(1800, 1050))
	case "100%":
		a.topologyImage.FillMode = canvas.ImageFillOriginal
		a.topologyImage.SetMinSize(fyne.NewSize(1200, 700))
	default:
		a.topologyImage.FillMode = canvas.ImageFillContain
		a.topologyImage.SetMinSize(fyne.NewSize(0, 260))
	}
	a.topologyImage.Refresh()
	a.topologyImgBox.Refresh()
	a.topologyImgScroll.Refresh()
}

func (a *App) openPreviewExternal() {
	if strings.TrimSpace(a.previewPath) == "" {
		dialog.ShowInformation("Информация", "Сначала постройте превью топологии", a.myWindow)
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", "", a.previewPath)
	case "darwin":
		cmd = exec.Command("open", a.previewPath)
	default:
		cmd = exec.Command("xdg-open", a.previewPath)
	}
	if err := cmd.Start(); err != nil {
		dialog.ShowError(fmt.Errorf("не удалось открыть файл: %v", err), a.myWindow)
		return
	}
}

func topoDisplayName(d *topology.Device) string {
	if d == nil {
		return "unknown"
	}
	if d.Hostname != "" {
		return d.Hostname
	}
	if d.IP != "" {
		return d.IP
	}
	if d.MAC != "" {
		return d.MAC
	}
	return "unknown"
}

func topoPortName(p *topology.Port) string {
	if p == nil {
		return "-"
	}
	if p.Name != "" {
		return p.Name
	}
	if p.Index > 0 {
		return fmt.Sprintf("if%d", p.Index)
	}
	return "-"
}

// Run запускает GUI приложение
func (a *App) Run() {
	a.myWindow.SetOnClosed(func() {
		if a.previewPath != "" {
			_ = os.Remove(a.previewPath)
		}
	})
	a.myWindow.ShowAndRun()
}
