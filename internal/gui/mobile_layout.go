package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MobileLayout управляет адаптивным расположением UI для мобильных устройств
type MobileLayout struct {
	isMobile           bool
	smallScreen        bool
	currentOrientation string // "portrait" или "landscape"
	mainTabs           *container.AppTabs
	tabs               []*container.TabItem
}

// NewMobileLayout создает новый адаптивный лейаут
func NewMobileLayout(tabs *container.AppTabs) *MobileLayout {
	return &MobileLayout{
		mainTabs:           tabs,
		isMobile:           false,
		smallScreen:        false,
		currentOrientation: "portrait",
	}
}

// Update проверяет размеры экрана и обновляет layout
func (ml *MobileLayout) Update(canvas fyne.Canvas) {
	size := canvas.Size()
	ml.smallScreen = size.Width < 600 || size.Height < 800
	ml.isMobile = ml.smallScreen

	// Определяем ориентацию
	if size.Height > size.Width {
		ml.currentOrientation = "portrait"
	} else {
		ml.currentOrientation = "landscape"
	}

	// Применяем мобильный layout если нужно
	if ml.isMobile {
		ml.applyMobileLayout()
	}
}

// applyMobileLayout применяет оптимизированный layout для мобильных устройств
func (ml *MobileLayout) applyMobileLayout() {
	if ml.mainTabs == nil {
		return
	}

	// В портретном режиме показываем только 2 вкладки вместо 3
	if ml.currentOrientation == "portrait" && len(ml.tabs) > 2 {
		// Скрываем третью вкладку на маленьких экранах
	}

	// Уменьшаем размеры шрифтов для маленьких экранов
	if ml.smallScreen {
		// TODO: Уменьшить размеры шрифтов
	}
}

// GetLayoutMode возвращает текущий режим layout
func (ml *MobileLayout) GetLayoutMode() string {
	if ml.isMobile {
		if ml.smallScreen {
			return "compact"
		}
		return "mobile"
	}
	return "desktop"
}

// CreateMobileTabBar создает компактную панель вкладок для мобильных устройств
func CreateMobileTabBar(tabs []*container.TabItem) *fyne.Container {
	if len(tabs) == 0 {
		return nil
	}

	buttons := make([]fyne.CanvasObject, 0, len(tabs))
	for _, tab := range tabs {
		_ = tab
		btn := widget.NewButton("Tab", func() {
			// TODO: Переключить на эту вкладку
		})
		buttons = append(buttons, btn)
	}

	return container.NewGridWithColumns(len(buttons), buttons...)
}
