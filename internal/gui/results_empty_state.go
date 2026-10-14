package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// buildResultsEmptyState формирует структурированное empty/error-состояние
// области результатов (S-01): иконка + заголовок + подпись + опциональное действие.
//
// title    — крупная подпись (например «Нет результатов»).
// message  — пояснение (например «Нет устройств, подходящих под фильтры»).
// icon     — тематическая иконка (nil → theme.InfoIcon()).
// actionLabel + onAction — кнопка (например «Сбросить фильтры»).
//
// Возвращает контейнер, готовый к вставке в resultsBody.Objects.
func buildResultsEmptyState(title, message string, icon fyne.Resource, actionLabel string, onAction func()) fyne.CanvasObject {
	if icon == nil {
		icon = iconInfo()
	}
	iconWidget := widget.NewIcon(icon)
	iconWidget.Resize(fyne.NewSize(48, 48))

	titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	msgLbl := widget.NewLabel(message)
	msgLbl.Wrapping = fyne.TextWrapWord

	children := []fyne.CanvasObject{
		container.NewCenter(iconWidget),
		container.NewCenter(titleLbl),
		container.NewCenter(msgLbl),
	}
	if actionLabel != "" && onAction != nil {
		btn := widget.NewButtonWithIcon(actionLabel, iconClear(), onAction)
		children = append(children, container.NewCenter(btn))
	}

	return container.NewCenter(
		container.NewGridWrap(fyne.NewSize(320, 180),
			container.NewVBox(children...),
		),
	)
}
