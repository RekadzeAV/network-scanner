package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"

	scand "network-scanner/internal/scanner/daemon"
)

// ObserveScanRunner наблюдает за раннером сканирования и обновляет UI.
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

		// Момент последней публикации частичных результатов (P1-1): троттлим,
		// чтобы не перерисовывать таблицу на каждое событие прогресса.
		var lastLive time.Time

		for {
			select {
			case <-time.After(timeout):
				fyne.Do(func() {
					if a.statusLabel != nil {
						a.statusLabel.SetText("Таймаут сканирования")
					}
					a.setResultsState(resultsStateTimeout, -1)
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
					published := a.publishLiveResults(runner, lastLive)
					if published {
						lastLive = time.Now()
					}
					fyne.Do(func() {
						if a.stageLabel != nil {
							a.stageLabel.Show()
							a.stageLabel.SetText(fmt.Sprintf("%s: %s", ev.Stage, ev.Message))
						}
						if a.progressBar != nil {
							a.progressBar.Show()
							a.progressBar.SetValue(ev.Percent)
						}
						if a.statusLabel != nil {
							a.statusLabel.SetText(scanProgressText(ev, startTime, len(a.scanResults)))
						}
					})
				case scand.EventDone:
					fyne.Do(func() {
						if a.statusLabel != nil {
							a.statusLabel.SetText(fmt.Sprintf("Сканирование завершено за %v. Найдено устройств: %d", time.Since(startTime), len(ev.Results)))
						}
						a.setResultsState(resultsStateDone, len(ev.Results))
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
						// Обновляем результаты: через setScanResults, чтобы версия
						// пайплайна и кэш фильтрации были обновлены (P1-1).
						a.setScanResults(ev.Results)
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
