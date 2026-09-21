package gui

import (
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/configvalidation"
)

// --- Валидация ввода on-the-fly для полей сканирования ---

// validateIntField проверяет целое число в диапазоне [min, max].
func validateIntField(value string, minV, maxV int) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return &configvalidation.ValidationError{}
	}
	if n < minV || n > maxV {
		return &configvalidation.ValidationError{}
	}
	return nil
}

// makeValidatedEntry оборачивает entry в VBox с сообщением валидации.
// Возвращает контейнер и текст-лейбл для показа/скрытия ошибки.
func makeValidatedEntry(entry *widget.Entry) (*fyne.Container, *canvas.Text) {
	msg := canvas.NewText("", themeColorError())
	msg.TextSize = 11
	msg.Hide()
	box := container.NewVBox(entry, msg)
	return box, msg
}

// showFieldError показывает ошибку под полем или скрывает её (err == nil).
func showFieldError(msg *canvas.Text, err error) {
	if msg == nil {
		return
	}
	if err == nil {
		msg.Text = ""
		msg.Hide()
		msg.Refresh()
		return
	}
	msg.Text = "⚠ " + err.Error()
	msg.Show()
	msg.Refresh()
}

// showFieldErrorDelayed безопасно вызывает showFieldError из UI-колбэков
// (nil-безопасно по msg).
func (a *App) showFieldErrorDelayed(msg *canvas.Text, err error) {
	showFieldError(msg, err)
}

// themeColorError возвращает цвет ошибки из активной темы (с fallback).
func themeColorError() color.Color {
	if a := fyne.CurrentApp(); a != nil && a.Settings() != nil {
		if th := a.Settings().Theme(); th != nil {
			return th.Color(theme.ColorNameError, a.Settings().ThemeVariant())
		}
	}
	return color.RGBA{R: 200, G: 40, B: 40, A: 255}
}

// validateCIDRField проверяет CIDR. Пустое значение допустимо (автоопределение).

// validateCIDRField проверяет CIDR. Пустое значение допустимо (автоопределение).
func validateCIDRField(value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	if verr := configvalidation.ValidateCIDR(v); verr != nil {
		return verr
	}
	return nil
}

// validatePortRangeField проверяет диапазон портов. Пустое значение допустимо.
func validatePortRangeField(value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	if verr := configvalidation.ValidatePortRange(v); verr != nil {
		return verr
	}
	return nil
}