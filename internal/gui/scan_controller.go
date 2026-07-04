package gui

import (
	"fmt"
	"strings"
	"time"

	"network-scanner/internal/logger"
	scand "network-scanner/internal/scanner/daemon"

	"fyne.io/fyne/v2"
)

func (a *App) applyScanRunStart(autoProfileNote string) {
	if a == nil {
		return
	}
	// Показываем тулбар при сканировании
	if a.mainToolbar != nil {
		a.mainToolbar.Show()
		a.mainToolbar.Refresh()
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
	a.scanResultsVersion++
	a.invalidateResultsPipelineCache()
	a.hostDetailsCacheMu.Lock()
	a.hostDetailsCache = make(map[string]string)
	a.hostDetailsCacheMu.Unlock()
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
	// Скрываем тулбар после завершения сканирования
	if a.mainToolbar != nil {
		a.mainToolbar.Hide()
	}
	results := update.results
	a.scanResults = results
	a.saveInventorySnapshotFromResults(results)
	a.scanResultsVersion++
	a.invalidateResultsPipelineCache()
	a.hostDetailsCacheMu.Lock()
	a.hostDetailsCache = make(map[string]string)
	a.hostDetailsCacheMu.Unlock()
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
	a.scanRunner = nil
	a.myWindow.Content().Refresh()
}

func (a *App) applyScanTimeout(scanUITimeout time.Duration) {
	if a == nil {
		return
	}
	// Скрываем тулбар при таймауте
	if a.mainToolbar != nil {
		a.mainToolbar.Hide()
	}
	if a.scanRunner != nil {
		logger.Log("GUI таймаут сканирования: останавливаем runner")
		a.scanRunner.Stop()
	}
	a.statusLabel.SetText(fmt.Sprintf("Таймаут сканирования (%s)", formatDurationMMSS(scanUITimeout)))
	a.resultsState = resultsStateTimeout
	a.stageLabel.Hide()
	a.scanButton.Enable()
	a.stopButton.Disable()
	a.scanRunner = nil
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

func (a *App) observeScanRunner(runner *scand.Runner, scanStartTime time.Time, scanUITimeout time.Duration) {
	go func() {
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()

		timeout := time.NewTimer(scanUITimeout)
		defer timeout.Stop()
		stageStartedAt := map[string]time.Time{}
		var latestProgress progressUpdate
		hasPendingProgress := false

		applyProgress := func(progress progressUpdate) {
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
			fyne.Do(func() {
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
				a.progressBar.SetValue(progress.percent)
				if progress.total > 0 {
					percentText := fmt.Sprintf("%.1f%%", progress.percent*100)
					a.stageLabel.SetText(fmt.Sprintf("%s: %d/%d (%s%s)", stageName, progress.current, progress.total, percentText, etaText))
				} else {
					a.stageLabel.SetText(stageName)
				}
				a.statusLabel.SetText(progress.message)
				a.progressBar.Refresh()
				a.stageLabel.Refresh()
				a.statusLabel.Refresh()
			})
		}

		for {
			select {
			case ev := <-runner.Events():
				if ev.Kind != scand.EventProgress && ev.Kind != scand.EventDone && ev.Kind != scand.EventStopped && ev.Kind != scand.EventError {
					continue
				}
				if ev.Kind == scand.EventProgress {
					progress := progressUpdate{
						stage:   ev.Stage,
						current: ev.Current,
						total:   ev.Total,
						message: ev.Message,
						percent: ev.Percent,
					}
					if _, exists := stageStartedAt[progress.stage]; !exists {
						stageStartedAt[progress.stage] = time.Now()
					}
					latestProgress = progress
					hasPendingProgress = true
					if progress.stage == "complete" {
						applyProgress(progress)
						hasPendingProgress = false
					}
					continue
				}
				if ev.Kind == scand.EventStopped {
					fyne.Do(func() {
						a.stopScan()
					})
					return
				}
				if ev.Kind == scand.EventError {
					fyne.Do(func() {
						// Скрываем тулбар при ошибке
						if a.mainToolbar != nil {
							a.mainToolbar.Hide()
						}
						msg := strings.TrimSpace(ev.Message)
						if msg == "" && ev.Err != nil {
							msg = ev.Err.Error()
						}
						if msg == "" {
							msg = "внутренняя ошибка scan runner"
						}
						a.statusLabel.SetText("Ошибка сканирования: " + msg)
						a.resultsState = resultsStateTimeout
						a.stageLabel.Hide()
						a.progressBar.Hide()
						a.scanButton.Enable()
						a.stopButton.Disable()
						a.scanRunner = nil
						a.renderScanResultsView()
						a.statusLabel.Refresh()
						a.stageLabel.Refresh()
						a.progressBar.Refresh()
						a.resultsStateLabel.Refresh()
					})
					return
				}
				update := scanUpdate{results: ev.Results, diagnostics: ev.Diagnostics}
				if hasPendingProgress {
					applyProgress(latestProgress)
					hasPendingProgress = false
				}
				fyne.Do(func() {
					a.applyScanCompletion(update)
				})
				totalDuration := time.Since(scanStartTime)
				logger.Log("Сканирование в GUI завершено за %v, найдено устройств: %d", totalDuration, len(update.results))
				return

			case <-ticker.C:
				if hasPendingProgress {
					applyProgress(latestProgress)
					hasPendingProgress = false
				}

			case <-timeout.C:
				fyne.Do(func() {
					a.applyScanTimeout(scanUITimeout)
				})
				return
			}
		}
	}()
}
