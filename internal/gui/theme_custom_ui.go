package gui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// theme_custom_ui.go — диалоги кастомной темы (L1.2/L1.4).

// previewSwatch — прямоугольник фиксированного размера для превью.
type previewSwatch struct {
	*canvas.Rectangle
	size fyne.Size
}

func newPreviewSwatch(c color.Color, w, h float32) *previewSwatch {
	return &previewSwatch{Rectangle: canvas.NewRectangle(c), size: fyne.NewSize(w, h)}
}

func (s *previewSwatch) MinSize() fyne.Size { return s.size }

// themePreviewContent — содержимое диалога предпросмотра тем (L1.4):
// карточки светлого и тёмного вариантов с акцентными подложками.
func themePreviewContent(accentName string) fyne.CanvasObject {
	lb := canvas.NewText("Светлая тема", tokLightText)
	lb.TextSize = 16
	lb.TextStyle = fyne.TextStyle{Bold: true}
	db := canvas.NewText("Тёмная тема", tokDarkText)
	db.TextSize = 16
	db.TextStyle = fyne.TextStyle{Bold: true}

	cardLight := container.NewBorder(nil, nil, nil, nil,
		container.NewVBox(
			lb,
			newPreviewSwatch(tokLightPrimary, 140, 28),
			container.NewHBox(
				newPreviewSwatch(tokLightPrimary, 70, 22),
				newPreviewSwatch(tokLightSurfaceAlt, 70, 22),
				newPreviewSwatch(tokLightBorderStrong, 70, 22),
			),
		),
	)
	cardDark := container.NewBorder(nil, nil, nil, nil,
		container.NewVBox(
			db,
			newPreviewSwatch(tokDarkPrimary, 140, 28),
			container.NewHBox(
				newPreviewSwatch(tokDarkPrimary, 70, 22),
				newPreviewSwatch(tokDarkSurfaceAlt, 70, 22),
				newPreviewSwatch(tokDarkBorderStrong, 70, 22),
			),
		),
	)

	header := widget.NewLabel("Предпросмотр тем (акцент: " + accentName + ")")
	return container.NewBorder(header, nil, nil, nil,
		container.NewHBox(
			container.NewVScroll(cardLight),
			container.NewVScroll(cardDark),
		),
	)
}

// showAccentCustomDialog — диалог ввода произвольного цвета акцента (L1.2).
func (a *App) showAccentCustomDialog() {
	if a == nil || a.myWindow == nil {
		return
	}
	hexEntry := widget.NewEntry()
	hexEntry.SetPlaceHolder("#RRGGBB")
	hexEntry.SetText(a.loadCustomAccentHex())
	hexEntry.Validator = func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		if !isValidHexColor(s) {
			return fmt.Errorf("ожидается #RRGGBB")
		}
		return nil
	}

	content := container.NewVBox(
		widget.NewLabel("Введите цвет в формате #RRGGBB для кастомного акцента:"),
		hexEntry,
	)
	d := dialog.NewCustomWithoutButtons("Кастомный акцент", content, a.myWindow)
	d.Resize(fyne.NewSize(400, 120))
	applyBtn := widget.NewButton("Применить", func() {
		a.applyCustomAccentHex(hexEntry.Text)
		d.Hide()
	})
	cancelBtn := widget.NewButton("Отмена", func() {
		d.Hide()
	})
	d.SetButtons([]fyne.CanvasObject{applyBtn, cancelBtn})
	d.Show()
}

// showThemePreviewDialog — диалог предпросмотра светлой и тёмной тем (L1.4).
func (a *App) showThemePreviewDialog() {
	if a == nil || a.myWindow == nil {
		return
	}
	content := themePreviewContent(a.currentAccentCaption())
	d := dialog.NewCustomWithoutButtons("Предпросмотр тем", content, a.myWindow)
	d.Resize(fyne.NewSize(640, 400))
	d.Show()
}

// currentAccentCaption — подпись активного акцента для диалога превью.
func (a *App) currentAccentCaption() string {
	preset := a.loadAccentPreset()
	if preset == customAccentName {
		if hex := a.loadCustomAccentHex(); hex != "" {
			return preset + " " + hex
		}
	}
	return preset
}
