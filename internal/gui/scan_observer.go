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
