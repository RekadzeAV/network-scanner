package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"network-scanner/internal/logger"
)

func (a *App) applyScanRunStart(autoProfileNote string) {
	if a == nil {
		return
	}
	a.scanButton.Disable()
	a.stopButton.Enable()
	a.saveButton.Disable()
	a.progressBar.Show()
	a.progressBar.SetValue(0)
	a.stageLabel.Show()
	a.statusLabel.SetText("Сканирование запущено...")
	a.topologyStatus.SetText("Ожидание завершения сканирования...")
	if strings.TrimSpace(autoProfileNote) != "" {
		a.stageLabel.SetText("Инициализация... " + autoProfileNote)
	} else {
		a.stageLabel.SetText("Инициализация...")
	}

	a.scanResults = nil
	if a.diagnosticsLabel != nil {
		a.diagnosticsLabel.SetText("Диагностика последнего запуска: выполняется...")
	}
	if a.copyDiagnosticsBtn != nil {
		a.copyDiagnosticsBtn.Disable()
	}
	if a.saveDiagnosticsBtn != nil {
		a.saveDiagnosticsBtn.Disable()
	}
	a.pieChartCache = make(map[string]fyne.Resource)
	a.resultsState = resultsStateScanning
	a.renderScanResultsView()
	a.resultsScroll.Refresh()
	a.topologyText.ParseMarkdown("## Топология сети\n\nСканирование выполняется. После завершения станет доступно построение топологии.")
	a.topologyScroll.Refresh()
}

func (a *App) applyScanCompletion(update scanUpdate) {
	if a == nil {
		return
	}
	results := update.results
	a.scanResults = results
	a.resultsState = resultsStateDone
	if a.diagnosticsLabel != nil {
		a.diagnosticsLabel.SetText(update.diagnostics)
	}
	if a.copyDiagnosticsBtn != nil && strings.TrimSpace(update.diagnostics) != "" {
		a.copyDiagnosticsBtn.Enable()
	}
	if a.saveDiagnosticsBtn != nil && strings.TrimSpace(update.diagnostics) != "" {
		a.saveDiagnosticsBtn.Enable()
	}
	a.progressBar.SetValue(1.0)
	a.progressBar.Hide()
	a.stageLabel.Hide()

	if len(results) == 0 {
		a.statusLabel.SetText("Сканирование завершено. Результаты не найдены.")
		a.topologyStatus.SetText("Нет результатов сканирования для построения топологии")
	} else {
		a.statusLabel.SetText(fmt.Sprintf("Сканирование завершено. Найдено устройств: %d", len(results)))
		a.saveButton.Enable()
		a.buildTopoBtn.Enable()
		a.topologyStatus.SetText("Можно строить топологию: перейдите на вкладку 'Топология'")
	}
	a.renderScanResultsView()
	a.resultsScroll.ScrollToTop()
	a.resultsScroll.Refresh()
	a.progressBar.Refresh()
	a.stageLabel.Refresh()
	a.statusLabel.Refresh()
	a.resultsStateLabel.Refresh()
	if a.diagnosticsLabel != nil {
		a.diagnosticsLabel.Refresh()
	}
	if a.copyDiagnosticsBtn != nil {
		a.copyDiagnosticsBtn.Refresh()
	}
	if a.saveDiagnosticsBtn != nil {
		a.saveDiagnosticsBtn.Refresh()
	}
	a.topologyStatus.Refresh()
	a.scanButton.Enable()
	a.stopButton.Disable()
	a.networkScanner = nil
	a.myWindow.Content().Refresh()
}

func (a *App) applyScanTimeout(scanUITimeout time.Duration) {
	if a == nil {
		return
	}
	if a.networkScanner != nil {
		logger.Log("GUI таймаут сканирования: принудительно останавливаем активный сканер")
		a.networkScanner.Stop()
	}
	a.statusLabel.SetText(fmt.Sprintf("Таймаут сканирования (%s)", formatDurationMMSS(scanUITimeout)))
	a.resultsState = resultsStateTimeout
	a.stageLabel.Hide()
	a.scanButton.Enable()
	a.stopButton.Disable()
	a.networkScanner = nil
	if a.copyDiagnosticsBtn != nil {
		a.copyDiagnosticsBtn.Disable()
	}
	if a.saveDiagnosticsBtn != nil {
		a.saveDiagnosticsBtn.Disable()
	}
	a.progressBar.Hide()
	a.renderScanResultsView()
	a.statusLabel.Refresh()
	a.stageLabel.Refresh()
	a.progressBar.Refresh()
	a.resultsStateLabel.Refresh()
}
