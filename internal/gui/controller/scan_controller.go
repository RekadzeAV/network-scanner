package controller

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"network-scanner/internal/gui/errors"
	"network-scanner/internal/logger"
	"network-scanner/internal/network"
	"network-scanner/internal/scanner"
	scand "network-scanner/internal/scanner/daemon"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	largeSubnetWarnHostGUI = 512
	maxScanThreadsGUI      = 512
	resultsStateStopped    = "stopped"
)

// ScanUI предоставляет доступ к виджетам сканирования.
type ScanUI struct {
	NetworkEntry         *widget.Entry
	PortRangeEntry       *widget.Entry
	TimeoutEntry         *widget.Entry
	ThreadsEntry         *widget.Entry
	ScanUDPCheck         *widget.Check
	ScanBannersCheck     *widget.Check
	ScanOSActiveCheck    *widget.Check
	ScanVerboseLogsCheck *widget.Check
	ScanTCPPortsCheck    *widget.Check
	AutoProfileCheck     *widget.Check
	StatusLabel          *widget.Label
	RecommendedBadge     *canvas.Text
	PresetQuickBtn       *widget.Button
	PresetBalBtn         *widget.Button
	PresetDeepBtn        *widget.Button
	RecommendedBtn       *widget.Button

	// Дополнительные виджеты для полного контроля
	ScanButton         *widget.Button
	StopButton         *widget.Button
	StageLabel         *widget.Label
	ProgressBar        *widget.ProgressBar
	ResultsStateLabel  *widget.Label
	CopyDiagnosticsBtn *widget.Button
	SaveDiagnosticsBtn *widget.Button
	MainToolbar        *fyne.Container
	Window             fyne.Window
}

// ScanController управляет логикой сканирования и настройками.
type ScanController struct {
	app        fyne.App
	ui         *ScanUI
	errHandler *errors.ErrorHandler
	runner     *scand.Runner
	results    []interface{}
	actions    ScanActions
}

// ScanActions интерфейс для сложных операций, которые остаются в App
type ScanActions interface {
	ApplyScanRunStart(autoProfileNote string)
	ObserveScanRunner(runner *scand.Runner, startTime time.Time, timeout time.Duration)
	RenderScanResultsView()
	ConfirmLargeScanBypass() bool
	SetConfirmLargeScanBypass(val bool)
}

// NewScanController создает контроллер.
func NewScanController(app fyne.App, ui *ScanUI, actions ScanActions) *ScanController {
	return &ScanController{
		app:        app,
		ui:         ui,
		errHandler: errors.NewErrorHandler("scan"),
		actions:    actions,
	}
}

// LoadSettings загружает настройки из Preferences в UI.
func (c *ScanController) LoadSettings() {
	if c.app == nil || c.ui == nil {
		return
	}
	defer c.errHandler.HandlePanic(nil)

	p := c.app.Preferences()
	if v := p.String("scan.network"); v != "" {
		c.ui.NetworkEntry.SetText(v)
	}
	if v := p.String("scan.port_range"); v != "" {
		c.ui.PortRangeEntry.SetText(v)
	}
	if v := p.String("scan.timeout_sec"); v != "" {
		c.ui.TimeoutEntry.SetText(v)
	}
	if v := p.String("scan.threads"); v != "" {
		c.ui.ThreadsEntry.SetText(v)
	}
	c.ui.ScanUDPCheck.SetChecked(p.String("scan.udp") == "true")
	if c.ui.ScanBannersCheck != nil {
		c.ui.ScanBannersCheck.SetChecked(p.String("scan.grab_banners") == "true")
	}
	if c.ui.ScanOSActiveCheck != nil {
		c.ui.ScanOSActiveCheck.SetChecked(p.String("scan.os_detect_active") == "true")
	}
	if c.ui.ScanVerboseLogsCheck != nil {
		c.ui.ScanVerboseLogsCheck.SetChecked(p.String("scan.verbose_port_logs") == "true")
	}
	if c.ui.ScanTCPPortsCheck != nil {
		tcpPref := p.String("scan.scan_tcp_ports")
		c.ui.ScanTCPPortsCheck.SetChecked(tcpPref == "" || tcpPref == "true")
	}
	if c.ui.AutoProfileCheck != nil {
		autoPref := p.String("scan.auto_profile")
		c.ui.AutoProfileCheck.SetChecked(autoPref == "" || autoPref == "true")
	}
}

// SaveSettings сохраняет текущие значения UI в Preferences.
func (c *ScanController) SaveSettings() {
	if c.app == nil || c.ui == nil {
		return
	}
	p := c.app.Preferences()
	p.SetString("scan.network", c.ui.NetworkEntry.Text)
	p.SetString("scan.port_range", c.ui.PortRangeEntry.Text)
	p.SetString("scan.timeout_sec", c.ui.TimeoutEntry.Text)
	p.SetString("scan.threads", c.ui.ThreadsEntry.Text)
	if c.ui.ScanUDPCheck.Checked {
		p.SetString("scan.udp", "true")
	} else {
		p.SetString("scan.udp", "false")
	}
	if c.ui.ScanBannersCheck != nil {
		if c.ui.ScanBannersCheck.Checked {
			p.SetString("scan.grab_banners", "true")
		} else {
			p.SetString("scan.grab_banners", "false")
		}
	}
	if c.ui.ScanOSActiveCheck != nil {
		if c.ui.ScanOSActiveCheck.Checked {
			p.SetString("scan.os_detect_active", "true")
		} else {
			p.SetString("scan.os_detect_active", "false")
		}
	}
	if c.ui.ScanVerboseLogsCheck != nil {
		if c.ui.ScanVerboseLogsCheck.Checked {
			p.SetString("scan.verbose_port_logs", "true")
		} else {
			p.SetString("scan.verbose_port_logs", "false")
		}
	}
	if c.ui.ScanTCPPortsCheck != nil {
		if c.ui.ScanTCPPortsCheck.Checked {
			p.SetString("scan.scan_tcp_ports", "true")
		} else {
			p.SetString("scan.scan_tcp_ports", "false")
		}
	}
	if c.ui.AutoProfileCheck != nil {
		if c.ui.AutoProfileCheck.Checked {
			p.SetString("scan.auto_profile", "true")
		} else {
			p.SetString("scan.auto_profile", "false")
		}
	}
}

// ApplyPreset применяет пресет сканирования.
func (c *ScanController) ApplyPreset(mode string) {
	switch mode {
	case "quick":
		c.ui.PortRangeEntry.SetText("22,80,443,445,3389")
		c.ui.TimeoutEntry.SetText("1")
		c.ui.ThreadsEntry.SetText("120")
		c.ui.ScanUDPCheck.SetChecked(false)
		c.ui.ScanBannersCheck.SetChecked(false)
		c.ui.ScanOSActiveCheck.SetChecked(false)
		c.ui.StatusLabel.SetText("Пресет: Быстро (обзор)")
	case "deep":
		c.ui.PortRangeEntry.SetText("1-2000")
		c.ui.TimeoutEntry.SetText("3")
		c.ui.ThreadsEntry.SetText("40")
		c.ui.ScanUDPCheck.SetChecked(true)
		c.ui.ScanBannersCheck.SetChecked(true)
		c.ui.ScanOSActiveCheck.SetChecked(true)
		c.ui.StatusLabel.SetText("Пресет: Глубоко (детальный анализ)")
	default:
		c.ui.PortRangeEntry.SetText("1-1000")
		c.ui.TimeoutEntry.SetText("2")
		c.ui.ThreadsEntry.SetText("50")
		c.ui.ScanUDPCheck.SetChecked(false)
		c.ui.ScanBannersCheck.SetChecked(false)
		c.ui.ScanOSActiveCheck.SetChecked(false)
		c.ui.StatusLabel.SetText("Пресет: Баланс")
	}
	if c.app != nil {
		c.app.Preferences().SetString("scan.preset", mode)
	}
	c.RefreshPresetUI()
}

// ApplyRecommendedProfile применяет рекомендованный профиль.
func (c *ScanController) ApplyRecommendedProfile(networkStr string) {
	hosts := 0
	if networkStr != "" {
		if h, err := network.EstimateHostCount(networkStr); err == nil && h > 0 {
			hosts = h
		}
	}
	profileName := "стандарт"
	switch {
	case hosts >= 2048:
		c.ui.PortRangeEntry.SetText("22,80,443,445,3389")
		c.ui.TimeoutEntry.SetText("1")
		c.ui.ThreadsEntry.SetText("40")
		profileName = "бережный для очень крупной подсети"
	case hosts >= 1024:
		c.ui.PortRangeEntry.SetText("1-1024")
		c.ui.TimeoutEntry.SetText("2")
		c.ui.ThreadsEntry.SetText("60")
		profileName = "бережный для крупной подсети"
	case hosts >= 512:
		c.ui.PortRangeEntry.SetText("1-1024,3389")
		c.ui.TimeoutEntry.SetText("2")
		c.ui.ThreadsEntry.SetText("80")
		profileName = "сбалансированный для средней подсети"
	default:
		c.ui.PortRangeEntry.SetText("1-2048,3389")
		c.ui.TimeoutEntry.SetText("2")
		c.ui.ThreadsEntry.SetText("100")
		profileName = "углубленный для небольшой подсети"
	}
	c.ui.ScanUDPCheck.SetChecked(false)
	c.ui.ScanBannersCheck.SetChecked(false)
	c.ui.ScanOSActiveCheck.SetChecked(false)
	if c.ui.ScanVerboseLogsCheck != nil {
		c.ui.ScanVerboseLogsCheck.SetChecked(false)
	}
	if c.ui.AutoProfileCheck != nil {
		c.ui.AutoProfileCheck.SetChecked(true)
	}
	if c.app != nil {
		c.app.Preferences().SetString("scan.preset", "recommended")
	}
	if hosts > 0 {
		c.ui.StatusLabel.SetText(fmt.Sprintf("Применен рекомендованный профиль (%s), оценка подсети: ~%d хостов", profileName, hosts))
	} else {
		c.ui.StatusLabel.SetText(fmt.Sprintf("Применен рекомендованный профиль (%s)", profileName))
	}
	badgeClass := c.recommendedBadgeClassForHosts(hosts)
	if c.ui.RecommendedBadge != nil {
		c.ui.RecommendedBadge.Text = c.recommendedBadgeText(profileName, badgeClass)
		c.ui.RecommendedBadge.Color = &color.RGBA{R: 55, G: 130, B: 200, A: 255}
		c.ui.RecommendedBadge.Refresh()
	}
	c.RefreshPresetUI()
}

// recommendedBadgeClassForHosts возвращает класс бейджа.
func (c *ScanController) recommendedBadgeClassForHosts(hosts int) string {
	switch {
	case hosts >= 2048:
		return "red"
	case hosts >= 1024:
		return "orange"
	case hosts >= 512:
		return "yellow"
	default:
		return "green"
	}
}

// recommendedBadgeText возвращает текст бейджа.
func (c *ScanController) recommendedBadgeText(profileName, badgeClass string) string {
	return fmt.Sprintf("%s (%s)", profileName, badgeClass)
}

// RefreshPresetUI обновляет состояние виджетов после применения пресета.
func (c *ScanController) RefreshPresetUI() {
	if c.ui == nil {
		return
	}
	c.ui.PortRangeEntry.Refresh()
	c.ui.TimeoutEntry.Refresh()
	c.ui.ThreadsEntry.Refresh()
	c.ui.ScanUDPCheck.Refresh()
	c.ui.StatusLabel.Refresh()
}

// StartScan запускает сканирование сети
func (c *ScanController) StartScan(results []scanner.Result) {
	c.results = make([]interface{}, len(results))
	for i, r := range results {
		c.results[i] = r
	}
	scanStartTime := time.Now()
	logger.Log("Запуск сканирования из GUI")
	logger.LogDebug("Пользователь нажал кнопку 'Запустить сканирование'")

	// Определяем сеть
	networkStr := c.ui.NetworkEntry.Text
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
			dialog.ShowError(fmt.Errorf("не удалось определить сеть: %v", err), c.ui.Window)
			return
		}
		c.ui.NetworkEntry.SetText(networkStr)
		logger.Log("Определена сеть: %s (определение заняло %v)", networkStr, detectDuration)
		logger.LogDebug("Автоматическое определение сети завершено успешно")
	} else {
		logger.Log("Использована указанная сеть: %s", networkStr)
		logger.LogDebug("Сеть указана пользователем в поле ввода")
	}
	if hosts, err := network.EstimateHostCount(networkStr); err == nil && hosts >= largeSubnetWarnHostGUI && !c.actions.ConfirmLargeScanBypass() {
		dialog.NewConfirm(
			"Предупреждение о крупной подсети",
			fmt.Sprintf("Подсеть %s содержит примерно %d хостов.\nСканирование может занять продолжительное время и повлиять на отзывчивость интерфейса.\n\nПродолжить?", networkStr, hosts),
			func(ok bool) {
				if !ok {
					c.ui.StatusLabel.SetText("Сканирование отменено пользователем")
					return
				}
				c.actions.SetConfirmLargeScanBypass(true)
				c.StartScan(results)
			},
			c.ui.Window,
		).Show()
		return
	}
	c.actions.SetConfirmLargeScanBypass(false)
	if c.ui.ThreadsEntry != nil {
		threads := 50
		if v, err := strconv.Atoi(strings.TrimSpace(c.ui.ThreadsEntry.Text)); err == nil && v > 0 {
			threads = v
		}
		if threads < 1 {
			threads = 1
			c.ui.ThreadsEntry.SetText("1")
			c.ui.StatusLabel.SetText("Параметр threads скорректирован до 1")
		}
		if threads > maxScanThreadsGUI {
			threads = maxScanThreadsGUI
			c.ui.ThreadsEntry.SetText(strconv.Itoa(maxScanThreadsGUI))
			c.ui.StatusLabel.SetText(fmt.Sprintf("Параметр threads скорректирован до %d", maxScanThreadsGUI))
		}
	}
	autoProfileEnabled := true
	autoProfileNote := ""
	if c.ui.AutoProfileCheck != nil {
		autoProfileEnabled = c.ui.AutoProfileCheck.Checked
	}
	if autoProfileEnabled {
		portRange := ""
		if c.ui.PortRangeEntry != nil {
			portRange = strings.TrimSpace(c.ui.PortRangeEntry.Text)
		}
		threadsForProfile := 50
		if c.ui.ThreadsEntry != nil {
			if v, err := strconv.Atoi(strings.TrimSpace(c.ui.ThreadsEntry.Text)); err == nil && v > 0 {
				threadsForProfile = v
			}
		}
		profilePortRange, profileThreads, profileNote := autoScanProfile(networkStr, portRange, threadsForProfile)
		if profilePortRange != "" && c.ui.PortRangeEntry != nil && profilePortRange != strings.TrimSpace(c.ui.PortRangeEntry.Text) {
			c.ui.PortRangeEntry.SetText(profilePortRange)
		}
		if c.ui.ThreadsEntry != nil && profileThreads > 0 && profileThreads != threadsForProfile {
			c.ui.ThreadsEntry.SetText(strconv.Itoa(profileThreads))
		}
		if strings.TrimSpace(profileNote) != "" {
			c.ui.StatusLabel.SetText(profileNote)
			autoProfileNote = profileNote
		}
	}

	c.actions.ApplyScanRunStart(autoProfileNote)

	scanUITimeout := estimateScanUITimeout(networkStr, strings.TrimSpace(c.ui.PortRangeEntry.Text), strings.TrimSpace(c.ui.TimeoutEntry.Text), strings.TrimSpace(c.ui.ThreadsEntry.Text), c.ui.ScanTCPPortsCheck.Checked, c.ui.ScanUDPCheck.Checked)
	logger.LogDebug("GUI таймаут сканирования: %v", scanUITimeout)
	timeoutSec := 2
	if v, err := strconv.Atoi(strings.TrimSpace(c.ui.TimeoutEntry.Text)); err == nil && v > 0 {
		timeoutSec = v
	}
	portRange := strings.TrimSpace(c.ui.PortRangeEntry.Text)
	if portRange == "" {
		portRange = "1-65535"
	}
	threads := 50
	if v, err := strconv.Atoi(strings.TrimSpace(c.ui.ThreadsEntry.Text)); err == nil && v > 0 {
		threads = v
	}
	workerCfg := scand.Config{
		NetworkCIDR:    networkStr,
		Timeout:        time.Duration(timeoutSec) * time.Second,
		PortRange:      portRange,
		Threads:        threads,
		ShowClosed:     false,
		ScanTCPPorts:   c.ui.ScanTCPPortsCheck.Checked,
		ScanUDP:        c.ui.ScanUDPCheck.Checked,
		GrabBanners:    c.ui.ScanBannersCheck != nil && c.ui.ScanBannersCheck.Checked,
		OSDetectActive: c.ui.ScanOSActiveCheck != nil && c.ui.ScanOSActiveCheck.Checked,
		VerbosePortLog: c.ui.ScanVerboseLogsCheck != nil && c.ui.ScanVerboseLogsCheck.Checked,
	}
	runner := scand.NewRunner()
	c.runner = runner
	if err := runner.Start(workerCfg); err != nil {
		c.ui.StatusLabel.SetText("Не удалось запустить сканирование: " + err.Error())
		c.ui.ResultsStateLabel.SetText(resultsStateStopped)
		c.ui.StageLabel.Hide()
		c.ui.ProgressBar.Hide()
		c.ui.ScanButton.Enable()
		c.ui.StopButton.Disable()
		c.runner = nil
		c.actions.RenderScanResultsView()
		c.ui.StatusLabel.Refresh()
		c.ui.StageLabel.Refresh()
		c.ui.ProgressBar.Refresh()
		c.ui.ResultsStateLabel.Refresh()
		return
	}
	c.actions.ObserveScanRunner(runner, scanStartTime, scanUITimeout)
}

// StopScan останавливает сканирование
func (c *ScanController) StopScan() {
	if c.runner == nil {
		return
	}
	// Скрываем тулбар при остановке
	if c.ui.MainToolbar != nil {
		c.ui.MainToolbar.Hide()
	}
	logger.Log("Пользователь инициировал остановку сканирования из GUI")
	if c.runner != nil {
		c.runner.Stop()
	}
	c.ui.StatusLabel.SetText("Сканирование остановлено пользователем")
	c.ui.ResultsStateLabel.SetText(resultsStateStopped)
	c.ui.StageLabel.Hide()
	c.ui.ProgressBar.Hide()
	c.ui.ScanButton.Enable()
	c.ui.StopButton.Disable()
	if c.ui.CopyDiagnosticsBtn != nil {
		c.ui.CopyDiagnosticsBtn.Disable()
	}
	if c.ui.SaveDiagnosticsBtn != nil {
		c.ui.SaveDiagnosticsBtn.Disable()
	}
	c.runner = nil
	c.actions.RenderScanResultsView()
	c.ui.StatusLabel.Refresh()
	c.ui.StageLabel.Refresh()
	c.ui.ProgressBar.Refresh()
	c.ui.ResultsStateLabel.Refresh()
}

// autoScanProfile применяет автоматический профиль сканирования
func autoScanProfile(networkStr string, portRange string, threads int) (string, int, string) {
	portRange = strings.TrimSpace(portRange)
	if threads < 1 {
		threads = 1
	}
	hosts, err := network.EstimateHostCount(strings.TrimSpace(networkStr))
	if err != nil || hosts < 256 {
		return portRange, threads, ""
	}

	portCount := 0
	if portRange != "" {
		if ports, perr := network.ParsePortRange(portRange); perr == nil {
			portCount = len(ports)
		}
	}

	newPortRange := portRange
	newThreads := threads
	msg := ""

	switch {
	case hosts >= 2048:
		if portCount > 512 {
			newPortRange = "1-512"
		}
		if newThreads > 24 {
			newThreads = 24
		}
	case hosts >= 1024:
		if portCount > 1024 {
			newPortRange = "1-1024"
		}
		if newThreads > 40 {
			newThreads = 40
		}
	case hosts >= 512:
		if portCount > 2000 {
			newPortRange = "1-2000"
		}
		if newThreads > 64 {
			newThreads = 64
		}
	default:
		if portCount > 10000 {
			newPortRange = "1-4000"
		}
		if newThreads > 96 {
			newThreads = 96
		}
	}

	if newPortRange != portRange || newThreads != threads {
		parts := make([]string, 0, 2)
		if newPortRange != portRange {
			parts = append(parts, fmt.Sprintf("ports: %s -> %s", portRange, newPortRange))
		}
		if newThreads != threads {
			parts = append(parts, fmt.Sprintf("threads: %d -> %d", threads, newThreads))
		}
		msg = fmt.Sprintf("Автопрофиль: подсеть ~%d хостов, %s", hosts, strings.Join(parts, ", "))
	}
	return newPortRange, newThreads, msg
}

// estimateScanUITimeout оценивает таймаут сканирования
func estimateScanUITimeout(networkStr, portRange, timeoutText, threadsText string, scanTCP, scanUDP bool) time.Duration {
	base := 300 * time.Second

	timeoutSec := 2
	if v, err := strconv.Atoi(strings.TrimSpace(timeoutText)); err == nil && v > 0 {
		timeoutSec = v
	}
	threads := 50
	if v, err := strconv.Atoi(strings.TrimSpace(threadsText)); err == nil && v > 0 {
		threads = v
	}
	if threads < 1 {
		threads = 1
	}

	hosts := 256
	if h, err := network.EstimateHostCount(strings.TrimSpace(networkStr)); err == nil && h > 0 {
		hosts = h
	}

	ports := 0
	if scanTCP {
		effectiveRange := strings.TrimSpace(portRange)
		if effectiveRange == "" {
			effectiveRange = "1-65535"
		}
		if parsed, err := network.ParsePortRange(effectiveRange); err == nil {
			ports = len(parsed)
		}
	}
	if scanUDP {
		ports += 9
	}
	if ports == 0 {
		ports = 1
	}

	workUnits := hosts * ports
	estimatedSec := (workUnits * timeoutSec) / threads
	estimated := time.Duration(estimatedSec) * time.Second

	estimated = estimated / 4
	estimated += 90 * time.Second

	if estimated < base {
		return base
	}
	maxTimeout := 45 * time.Minute
	if estimated > maxTimeout {
		return maxTimeout
	}
	return estimated
}

// formatDurationMMSS форматирует длительность в MM:SS
func formatDurationMMSS(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSec := int(d.Round(time.Second).Seconds())
	min := totalSec / 60
	sec := totalSec % 60
	return fmt.Sprintf("%02d:%02d", min, sec)
}

// CopyScanDiagnostics копирует диагностику сканирования в буфер обмена.
func (c *ScanController) CopyScanDiagnostics(diagnosticsText string) {
	if diagnosticsText == "" || strings.Contains(strings.ToLower(diagnosticsText), "n/a") || strings.Contains(strings.ToLower(diagnosticsText), "выполняется") {
		dialog.ShowInformation("Информация", "Диагностика сканирования пока недоступна", c.ui.Window)
		return
	}
	c.ui.Window.Clipboard().SetContent(diagnosticsText)
	dialog.ShowInformation("Готово", "Диагностика сканирования скопирована в буфер обмена", c.ui.Window)
}

// SaveScanDiagnostics сохраняет диагностику сканирования в файл.
func (c *ScanController) SaveScanDiagnostics(diagnosticsText string) {
	if diagnosticsText == "" || strings.Contains(strings.ToLower(diagnosticsText), "n/a") || strings.Contains(strings.ToLower(diagnosticsText), "выполняется") {
		dialog.ShowInformation("Информация", "Диагностика сканирования пока недоступна", c.ui.Window)
		return
	}

	defaultFileName := fmt.Sprintf("scan-diagnostics-%s.txt", time.Now().Format("2006-01-02-150405"))
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, c.ui.Window)
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
			if _, writeErr := writer.Write([]byte(diagnosticsText)); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении диагностики: %v", writeErr), c.ui.Window)
				return
			}
		} else {
			_ = writer.Close()
			if writeErr := os.WriteFile(normalizedPath, []byte(diagnosticsText), 0644); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении диагностики: %v", writeErr), c.ui.Window)
				return
			}
		}
		dialog.ShowInformation("Готово", fmt.Sprintf("Диагностика сканирования сохранена: %s", normalizedPath), c.ui.Window)
	}, c.ui.Window)
	saveDialog.SetFileName(defaultFileName)
	saveDialog.Show()
}
