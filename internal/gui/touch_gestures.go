package gui

import (
	"fyne.io/fyne/v2"
)

// TouchGestures управляет жестами касания для мобильных устройств
type TouchGestures struct {
	app       *App
	canvas    fyne.Canvas
	isEnabled bool
}

// NewTouchGestures создает новый обработчик жестов
func NewTouchGestures(app *App) *TouchGestures {
	return &TouchGestures{
		app:       app,
		isEnabled: true,
	}
}

// Enable включает обработку жестов
func (tg *TouchGestures) Enable() {
	tg.isEnabled = true
}

// Disable отключает обработку жестов
func (tg *TouchGestures) Disable() {
	tg.isEnabled = false
}

// Setup настраивает обработчики жестов на canvas
func (tg *TouchGestures) Setup() {
	if tg.canvas == nil {
		tg.canvas = tg.app.myWindow.Canvas()
	}
}

// HandleSwipe обрабатывает свайп
func (tg *TouchGestures) HandleSwipe(event *fyne.PointEvent, dy float32) {
	if !tg.isEnabled {
		return
	}

	// Свайп вверх — прокрутка результатов
	if dy < -50 {
		tg.scrollUp()
	}
	// Свайп вниз — прокрутка вниз
	if dy > 50 {
		tg.scrollDown()
	}
}

// HandlePinch обрабатывает жест щипка (зум)
func (tg *TouchGestures) HandlePinch(event *fyne.PointEvent, scale float32) {
	if !tg.isEnabled {
		return
	}

	// Применяем масштаб к текущему canvas
	if tg.canvas != nil {
		currentScale := tg.canvas.Scale()
		newScale := currentScale * scale
		if newScale >= 0.5 && newScale <= 3.0 {
			// TODO: Установить новый масштаб
		}
	}
}

// HandleLongPress обрабатывает долгое нажатие
func (tg *TouchGestures) HandleLongPress(point fyne.Position) {
	if !tg.isEnabled {
		return
	}

	// Показываем контекстное меню
	if tg.app.myWindow != nil {
		// TODO: Показать контекстное меню
	}
}

// scrollUp прокручивает результаты вверх
func (tg *TouchGestures) scrollUp() {
	// TODO: Реализовать прокрутку
}

// scrollDown прокручивает результаты вниз
func (tg *TouchGestures) scrollDown() {
	// TODO: Реализовать прокрутку
}

// DesktopCustomShortcut кастомная обработка горячих клавиш для десктопа
type DesktopCustomShortcut struct {
	key         fyne.KeyName
	modifier    fyne.KeyModifier
	onTriggered func()
}

// NewDesktopCustomShortcut создает кастомную комбинацию клавиш
func NewDesktopCustomShortcut(key fyne.KeyName, modifier fyne.KeyModifier, triggered func()) *DesktopCustomShortcut {
	return &DesktopCustomShortcut{
		key:         key,
		modifier:    modifier,
		onTriggered: triggered,
	}
}

// Trigger вызывает обработчик
func (dcs *DesktopCustomShortcut) Trigger() {
	if dcs.onTriggered != nil {
		dcs.onTriggered()
	}
}
