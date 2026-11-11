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
// T-03/C-12: размер — основной текст темы, а не Caption (11 px): сообщение
// об ошибке — критичное, оно не должно мельчать на фоне поля ввода.
// (canvas.Text в Fyne v2.7.1 не поддерживает перенос/усечение — текст
// сообщения остаётся коротким по построению в configvalidation.)
func makeValidatedEntry(entry *widget.Entry) (*fyne.Container, *canvas.Text) {
	msg := canvas.NewText("", themeColorError())
	msg.TextSize = theme.TextSize()
	msg.Hide()
	box := container.NewVBox(entry, msg)
	return box, msg
}

// showFieldError показывает ошибку под полем или скрывает её (err == nil).
// S-05: устанавливает tooltip с текстом ошибки для скринридеров.
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

// showFieldErrorDelayed — UI-колбэк валидации поля: показывает/скрывает
// сообщение об ошибке под полем и дублирует ошибку в статус-баре (S-05).
//
// Fyne v2.7.1 не предоставляет API Tooltip/AriaLabel (см. ROADMAP: S-05
// aria-describedby — нет API Fyne), поэтому вместо tooltip текст ошибки
// прокидывается в постоянный статус-бар — единственную «приборную панель»
// приложения (N-07), которую считывают скринридеры и которая видна всегда.
// Это соответствует WCAG 3.3.1 (Error Identification): ошибка описана
// текстом и доступна вне визуального канала.
//
// При успешной валидации статус-бар не трогаем: колбэк вызывается на каждый
// keystroke и сброс к дефолту затирал бы полезные тосты («Сохранено ✓» и т.п.).
func (a *App) showFieldErrorDelayed(msg *canvas.Text, err error) {
	showFieldError(msg, err)
	if a == nil || err == nil {
		return
	}
	a.setStatusError("Ошибка ввода: " + err.Error())
}

// themeColorError возвращает цвет ошибки из токенов темы (C-12).
// Гарантирует WCAG AA ≥4.5:1: tokLightError = 6.54:1 на surface,
// tokDarkError = 7.19:1 на bg. Прямое чтение токенов вместо
// th.Color(theme.ColorNameError) — защищает от дефолтного цвета
// Fyne (#C82828, 3.26:1), который не проходит WCAG AA.
func themeColorError() color.Color {
	if a := fyne.CurrentApp(); a != nil && a.Settings() != nil {
		v := a.Settings().ThemeVariant()
		if v == theme.VariantDark {
			return tokDarkError
		}
		return tokLightError
	}
	return tokLightError
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
