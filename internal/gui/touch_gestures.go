package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
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
			// В Fyne v2.7.1 scale canvas устанавливается при запуске
			// Для динамического масштабирования используем env variable
			// TODO: Реализовать через FyneApp.toml или theme override
			_ = newScale
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
		tg.showContextMenu(point)
	}
}

// showContextMenu показывает контекстное меню в указанной позиции
func (tg *TouchGestures) showContextMenu(point fyne.Position) {
	// Создаём пункты меню
	menuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("Обновить", func() {
			// TODO: Реализовать обновление данных
		}),
		fyne.NewMenuItem("Настройки", func() {
			// TODO: Открыть настройки
		}),
		fyne.NewMenuItem("Экспорт", func() {
			// TODO: Реализовать экспорт данных
		}),
	}

	// Создаём меню
	menu := fyne.NewMenu("Контекстное меню", menuItems...)
	popUp := widget.NewPopUpMenu(menu, tg.app.myWindow.Canvas())

	// Показываем меню в указанной позиции
	popUp.ShowAtPosition(point)
}

// scrollUp прокручивает результаты вверх
func (tg *TouchGestures) scrollUp() {
	if tg.app == nil || tg.app.resultsScroll == nil {
		return
	}
	tg.app.resultsScroll.ScrollToTop()
}

// scrollDown прокручивает результаты вниз
func (tg *TouchGestures) scrollDown() {
	if tg.app == nil || tg.app.resultsScroll == nil {
		return
	}
	tg.app.resultsScroll.ScrollToBottom()
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
