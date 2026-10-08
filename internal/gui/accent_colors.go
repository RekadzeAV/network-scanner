package gui

import (
	"image/color"
	"strconv"
	"strings"
)

// Акцентные пресеты и их сохранение в Preferences (D-02).
//
// Единственный источник значений — accentPresets в theme_tokens.go: там лежат
// обе палитры пресета (light/dark) и onPrimary для текста поверх акцента.
// Файл — тонкая прослойка: типы и функции ниже используются настройками и
// меню «Тема → Акцент», но больше не дублируют цвета.

// prefAccentPreset — ключ Preferences с именем активного акцентного пресета.
const prefAccentPreset = "gui.theme.accent_preset"

// AccentColors — набор акцентных цветов одного варианта темы.
//
// Имя сохранено ради существовавших вызовов; полная палитра пресета описана
// типом accentVariant (theme_tokens.go) и включает OnPrimary.
type AccentColors struct {
	Primary   color.RGBA // Основной цвет (кнопки, фокус)
	Pressed   color.RGBA // Цвет нажатия
	Hover     color.RGBA // Цвет наведения
	Selection color.RGBA // Цвет выделения
}

// PresetThemes — встроенные пресеты акцентов (светлая палитра каждого пресета).
// Используется как превью и для валидации имён; применение пресета выбирает
// палитру по варианту темы.
var PresetThemes = func() map[string]AccentColors {
	out := make(map[string]AccentColors, len(accentPresets))
	for name := range accentPresets {
		out[name] = accentColorsOf(name, false)
	}
	return out
}()

// DefaultAccentColorsLight возвращает акцент по умолчанию для светлой темы.
func DefaultAccentColorsLight() AccentColors { return accentColorsOf(defaultAccentPreset, false) }

// DefaultAccentColorsDark возвращает акцент по умолчанию для тёмной темы.
func DefaultAccentColorsDark() AccentColors { return accentColorsOf(defaultAccentPreset, true) }

// accentColorsOf — палитра пресета в виде AccentColors для старых вызовов.
func accentColorsOf(name string, dark bool) AccentColors {
	variant := accentPresetOf(name)
	if dark {
		return accentColorsFrom(variant.Dark)
	}
	return accentColorsFrom(variant.Light)
}

func accentColorsFrom(v accentVariant) AccentColors {
	return AccentColors{Primary: v.Primary, Pressed: v.Pressed, Hover: v.Hover, Selection: v.Selection}
}

// accentPresetOf — пресет по имени; неизвестное имя приводит к умолчанию.
func accentPresetOf(name string) accentPresetEntry {
	if preset, ok := accentPresets[name]; ok {
		return preset
	}
	return accentPresets[defaultAccentPreset]
}

// ApplyAccentPreset применяет пресет акцента.
//
// isDark сохранён в сигнатуре ради существовавших вызовов: палитру варианта
// теперь выбирает сама тема (Color получает variant от Fyne), поэтому аргумент
// на результат не влияет. Неизвестное имя игнорируется.
func (t *ModernTheme) ApplyAccentPreset(name string, isDark bool) {
	_ = isDark
	if _, ok := accentPresets[name]; !ok {
		return
	}
	t.SetAccent(name)
}

// accentKey — ключ Preferences для произвольного цвета акцента.
func accentKey(name string) string {
	return "gui.theme.accent." + name
}

// saveAccentColor сохраняет отдельный акцентный цвет в Preferences.
func (a *App) saveAccentColor(name string, c color.RGBA) {
	if a == nil || a.myApp == nil {
		return
	}
	// Упаковка RGBA обязана идти в 32 битах: в арифметике uint8 сдвиги R и G
	// обнулялись, и в Preferences долетали только B и A.
	val := uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
	a.myApp.Preferences().SetString(accentKey(name), strconv.FormatUint(uint64(val), 16))
}

// loadAccentColor загружает отдельный акцентный цвет из Preferences.
func (a *App) loadAccentColor(name string, defaultColor color.RGBA) color.RGBA {
	if a == nil || a.myApp == nil {
		return defaultColor
	}
	raw := strings.TrimSpace(a.myApp.Preferences().String(accentKey(name)))
	if raw == "" {
		return defaultColor
	}
	val, err := strconv.ParseUint(raw, 16, 64)
	if err != nil {
		return defaultColor
	}
	return color.RGBA{
		R: byte((val >> 24) & 0xFF),
		G: byte((val >> 16) & 0xFF),
		B: byte((val >> 8) & 0xFF),
		A: byte(val & 0xFF),
	}
}

// saveAccentPreset сохраняет имя выбранного пресета в Preferences.
func (a *App) saveAccentPreset(name string) {
	if a == nil || a.myApp == nil {
		return
	}
	a.myApp.Preferences().SetString(prefAccentPreset, normalizeAccentPreset(name))
}

// loadAccentPreset загружает имя выбранного пресета (неизвестное — default).
func (a *App) loadAccentPreset() string {
	if a == nil || a.myApp == nil {
		return defaultAccentPreset
	}
	return normalizeAccentPreset(strings.TrimSpace(a.myApp.Preferences().String(prefAccentPreset)))
}
