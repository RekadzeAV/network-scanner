package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/telemetry"
)

// TelemetrySettingsManager управляет настройками телеметрии
type TelemetrySettingsManager struct {
	telemetry  *telemetry.Telemetry
	enabledCheck *widget.Check
	onChanged    func(bool)
}

// NewTelemetrySettingsManager создает менеджер настроек телеметрии
func NewTelemetrySettingsManager(tel *telemetry.Telemetry) *TelemetrySettingsManager {
	return &TelemetrySettingsManager{
		telemetry: tel,
	}
}

// CreateSettingsTab создает вкладку настроек телеметрии
func (tsm *TelemetrySettingsManager) CreateSettingsTab() fyne.CanvasObject {
	if tsm.telemetry == nil {
		return widget.NewLabel("Телеметрия недоступна")
	}

	tsm.enabledCheck = widget.NewCheck("Отправлять анонимную статистику использования", func(value bool) {
		tsm.telemetry.SetEnabled(value)
		if tsm.onChanged != nil {
			tsm.onChanged(value)
		}
	})
	tsm.enabledCheck.Checked = tsm.telemetry.IsEnabled()

	infoText := widget.NewRichText()
	infoText.ParseMarkdown(`### Анонимная телеметрия

Мы собираем ТОЛЬКО анонимную статистику:

- Версия приложения
- Количество сканирований
- Время выполнения
- Типы профилей сканирования

**НЕ собирается:**

- IP-адреса
- MAC-адреса
- Имена хостов
- Результаты сканирования

Эти данные помогают улучшить приложение.
Вы можете отключить телеметрию в любой момент.`)

	return container.NewVBox(
		tsm.enabledCheck,
		widget.NewSeparator(),
		widget.NewLabel("О телеметрии:"),
		infoText,
	)
}

// OnChanged устанавливает обработчик изменения состояния
func (tsm *TelemetrySettingsManager) OnChanged(fn func(bool)) {
	tsm.onChanged = fn
}

// GetStatus возвращает текущий статус телеметрии
func (tsm *TelemetrySettingsManager) GetStatus() string {
	if tsm.telemetry == nil {
		return "Недоступна"
	}
	
	stats := tsm.telemetry.GetStats()
	queueSize := 0
	if qs, ok := stats["queue_size"].(int); ok {
		queueSize = qs
	}
	
	return "Включена" + 
		" | Очередь: " + string(rune(queueSize+'0')) + " метрик"
}
