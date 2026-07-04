package gui

import (
	"network-scanner/internal/scanner"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
)

// AppModel представляет модель данных для GUI приложения.
// Использует Fyne data.Binding для безопасного обновления UI из горутин.
type AppModel struct {
	// scanResults — результаты сканирования
	scanResults []scanner.Result

	// progressValue — binding для прогресс-бара
	progressValue binding.Float

	// statusText — binding для текстового поля статуса
	statusText binding.String

	// scanButtonDisabled — binding для состояния кнопки сканирования
	scanButtonDisabled binding.Bool

	// resultsList — binding для списка результатов
	resultsList binding.UntypedList
}

// NewAppModel создаёт новую модель с инициализированными binding.
func NewAppModel() *AppModel {
	return &AppModel{
		scanResults:        make([]scanner.Result, 0),
		progressValue:      binding.NewFloat(),
		statusText:         binding.NewString(),
		scanButtonDisabled: binding.NewBool(),
		resultsList:        binding.NewUntypedList(),
	}
}

// AddHostResult добавляет результат сканирования в модель.
// Безопасно для вызова из горутин — использует fyne.Do для обновления UI.
func (m *AppModel) AddHostResult(result scanner.Result) {
	m.scanResults = append(m.scanResults, result)
	if m.resultsList != nil {
		fyne.Do(func() {
			// Обновление списка через binding
			_ = m.resultsList
		})
	}
}

// UpdateProgress обновляет прогресс сканирования.
func (m *AppModel) UpdateProgress(percent float64, stage string) {
	if m.progressValue != nil {
		fyne.Do(func() {
			_ = m.progressValue.Set(percent)
		})
	}
}

// SetStatus устанавливает статус приложения.
func (m *AppModel) SetStatus(status string) {
	if m.statusText != nil {
		fyne.Do(func() {
			_ = m.statusText.Set(status)
		})
	}
}

// SetScanning устанавливает флаг активного сканирования.
func (m *AppModel) SetScanning(scanning bool) {
	if m.scanButtonDisabled != nil {
		fyne.Do(func() {
			_ = m.scanButtonDisabled.Set(scanning)
		})
	}
}

// GetResults возвращает текущие результаты сканирования.
func (m *AppModel) GetResults() []scanner.Result {
	return m.scanResults
}

// GetStatus возвращает текущий статус.
func (m *AppModel) GetStatus() string {
	if m.statusText != nil {
		val, _ := m.statusText.Get()
		return val
	}
	return ""
}

// IsScanning возвращает флаг активного сканирования.
func (m *AppModel) IsScanning() bool {
	if m.scanButtonDisabled != nil {
		val, _ := m.scanButtonDisabled.Get()
		return val
	}
	return false
}

// GetProgressValue возвращает binding для прогресс-бара.
func (m *AppModel) GetProgressValue() binding.Float {
	return m.progressValue
}

// GetStatusText возвращает binding для текстового поля статуса.
func (m *AppModel) GetStatusText() binding.String {
	return m.statusText
}

// GetScanButtonDisabled возвращает binding для состояния кнопки сканирования.
func (m *AppModel) GetScanButtonDisabled() binding.Bool {
	return m.scanButtonDisabled
}

// GetResultsList возвращает binding для списка результатов.
func (m *AppModel) GetResultsList() binding.UntypedList {
	return m.resultsList
}
