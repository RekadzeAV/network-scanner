package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"

	"network-scanner/internal/eventbus"
)

// SubscribeToScanEvents подписывает GUI-элементы на события сканера через eventbus.
// Вызывается после инициализации App и UI. Если шина nil — подписки не создаются
// (отказобезопасное поведение).
func (a *App) SubscribeToScanEvents(bus *eventbus.EventBus) {
	if bus == nil || a == nil {
		return
	}

	// scan.started — блокируем кнопку скана, показываем прогресс.
	bus.Subscribe("scan.started", func(event eventbus.Event) {
		payload := event.Payload()
		network := ""
		if v, ok := payload["network"]; ok {
			network = fmt.Sprintf("%v", v)
		}
		fyne.Do(func() {
			if a.statusLabel != nil {
				if network != "" {
					a.statusLabel.SetText("Сканирование сети: " + network)
				} else {
					a.statusLabel.SetText("Сканирование начато")
				}
			}
			if a.stageLabel != nil {
				a.stageLabel.Show()
				a.stageLabel.SetText("Запуск сканирования…")
			}
			if a.progressBar != nil {
				a.progressBar.Show()
				a.progressBar.SetValue(0)
			}
			if a.scanButton != nil {
				a.scanButton.Disable()
			}
			if a.stopButton != nil {
				a.stopButton.Enable()
			}
			a.setResultsState(resultsStateScanning, 0)
			a.setScanProgress("Сканирование…")
		})
	})

	// scan.completed — разблокируем кнопку, скрываем прогресс, показываем статистику.
	bus.Subscribe("scan.completed", func(event eventbus.Event) {
		payload := event.Payload()
		hostCount := parseIntFromPayload(payload, "hostCount")
		openPorts := parseIntFromPayload(payload, "openPorts")
		duration := ""
		if v, ok := payload["duration"]; ok {
			duration = fmt.Sprintf("%v", v)
		}
		fyne.Do(func() {
			if a.statusLabel != nil {
				a.statusLabel.SetText(fmt.Sprintf("Завершено за %s. Устройств: %d, открытых портов: %d", duration, hostCount, openPorts))
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
			a.setResultsState(resultsStateDone, hostCount)
			a.setScanProgress("")
		})
	})

	// scan.failed — сообщаем об ошибке, разблокируем UI.
	bus.Subscribe("scan.failed", func(event eventbus.Event) {
		payload := event.Payload()
		reason := ""
		if v, ok := payload["reason"]; ok {
			reason = fmt.Sprintf("%v", v)
		}
		fyne.Do(func() {
			if a.statusLabel != nil {
				if reason != "" {
					a.statusLabel.SetText("Ошибка сканирования: " + reason)
				} else {
					a.statusLabel.SetText("Ошибка сканирования")
				}
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
			a.setResultsState(resultsStateStopped, 0)
			a.setScanProgress("")
		})
	})

	// progress.updated — обновляем прогресс-бар и_stage_ во время сканирования.
	bus.Subscribe("progress.updated", func(event eventbus.Event) {
		payload := event.Payload()
		stage := ""
		if v, ok := payload["stage"]; ok {
			stage = fmt.Sprintf("%v", v)
		}
		current := 0
		if v, ok := payload["current"]; ok {
			switch n := v.(type) {
			case int:
				current = n
			case float64:
				current = int(n)
			}
		}
		total := 0
		if v, ok := payload["total"]; ok {
			switch n := v.(type) {
			case int:
				total = n
			case float64:
				total = int(n)
			}
		}
		_ = total
		message := ""
		if v, ok := payload["message"]; ok {
			message = fmt.Sprintf("%v", v)
		}
		fyne.Do(func() {
			if a.stageLabel != nil {
				a.stageLabel.Show()
				if stage != "" {
					a.stageLabel.SetText(stage)
				}
			}
			if a.progressBar != nil {
				a.progressBar.Show()
				a.progressBar.SetValue(float64(current))
			}
			if message != "" && a.statusLabel != nil {
				a.statusLabel.SetText(message)
			}
		})
	})

	// host.found — 찾을оги: новое устройство найдено.
	bus.Subscribe("host.found", func(event eventbus.Event) {
		payload := event.Payload()
		ip := ""
		if v, ok := payload["ip"]; ok {
			ip = fmt.Sprintf("%v", v)
		}
		_ = ip
		// Детальное обновление таблицы результатов происходит через runner.Events();
		// здесь только логируем факт в статусной строке.
		_ = time.Now()
	})
}

// parseIntFromPayload извлекает int из payload по ключу, поддерживая int,
// float64 и string представления числа. Возвращает 0 при отсутствии или ошибке.
func parseIntFromPayload(payload map[string]interface{}, key string) int {
	v, ok := payload[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case string:
		var parsed int
		if _, err := fmt.Sscanf(n, "%d", &parsed); err == nil {
			return parsed
		}
	}
	return 0
}
