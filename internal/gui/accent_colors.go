package gui

import (
	"image/color"
	"strconv"
	"strings"
)

// AccentColors хранит пользовательские акцентные цвета
type AccentColors struct {
	Primary   color.RGBA // Основной цвет (кнопки, фокус)
	Pressed   color.RGBA // Цвет нажатия
	Hover     color.RGBA // Цвет наведения
	Selection color.RGBA // Цвет выделения
}

// DefaultAccentColorsLight возвращает стандартные светлые акцентные цвета (macOS)
func DefaultAccentColorsLight() AccentColors {
	return AccentColors{
		Primary:   color.RGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF},
		Pressed:   color.RGBA{R: 0x00, G: 0x5A, B: 0xCC, A: 0xFF},
		Hover:     color.RGBA{R: 0xE5, G: 0xF2, B: 0xFF, A: 0xFF},
		Selection: color.RGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0x40},
	}
}

// DefaultAccentColorsDark возвращает стандартные темные акцентные цвета (Windows 11)
func DefaultAccentColorsDark() AccentColors {
	return AccentColors{
		Primary:   color.RGBA{R: 0x00, G: 0x78, B: 0xD4, A: 0xFF},
		Pressed:   color.RGBA{R: 0x00, G: 0x5A, B: 0xA0, A: 0xFF},
		Hover:     color.RGBA{R: 0x2D, G: 0x2D, B: 0x30, A: 0xFF},
		Selection: color.RGBA{R: 0x00, G: 0x78, B: 0xD4, A: 0x40},
	}
}

// PresetTheme — встроенные пресеты цветов
var PresetThemes = map[string]AccentColors{
	"macOS Blue": {
		Primary:   color.RGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF},
		Pressed:   color.RGBA{R: 0x00, G: 0x5A, B: 0xCC, A: 0xFF},
		Hover:     color.RGBA{R: 0xE5, G: 0xF2, B: 0xFF, A: 0xFF},
		Selection: color.RGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0x40},
	},
	"Windows 11": {
		Primary:   color.RGBA{R: 0x00, G: 0x78, B: 0xD4, A: 0xFF},
		Pressed:   color.RGBA{R: 0x00, G: 0x5A, B: 0xA0, A: 0xFF},
		Hover:     color.RGBA{R: 0x2D, G: 0x2D, B: 0x30, A: 0xFF},
		Selection: color.RGBA{R: 0x00, G: 0x78, B: 0xD4, A: 0x40},
	},
	"Green": {
		Primary:   color.RGBA{R: 0x34, G: 0xC7, B: 0x59, A: 0xFF},
		Pressed:   color.RGBA{R: 0x28, G: 0xA7, B: 0x45, A: 0xFF},
		Hover:     color.RGBA{R: 0xD4, G: 0xF7, B: 0xDB, A: 0xFF},
		Selection: color.RGBA{R: 0x34, G: 0xC7, B: 0x59, A: 0x40},
	},
	"Purple": {
		Primary:   color.RGBA{R: 0x8E, G: 0x44, B: 0xAD, A: 0xFF},
		Pressed:   color.RGBA{R: 0x6C, G: 0x34, B: 0x83, A: 0xFF},
		Hover:     color.RGBA{R: 0xEB, G: 0xE5, B: 0xF9, A: 0xFF},
		Selection: color.RGBA{R: 0x8E, G: 0x44, B: 0xAD, A: 0x40},
	},
	"Orange": {
		Primary:   color.RGBA{R: 0xFF, G: 0x95, B: 0x00, A: 0xFF},
		Pressed:   color.RGBA{R: 0xE6, G: 0x7E, B: 0x00, A: 0xFF},
		Hover:     color.RGBA{R: 0xFF, G: 0xF3, B: 0xE0, A: 0xFF},
		Selection: color.RGBA{R: 0xFF, G: 0x95, B: 0x00, A: 0x40},
	},
	"Red": {
		Primary:   color.RGBA{R: 0xFF, G: 0x3B, B: 0x30, A: 0xFF},
		Pressed:   color.RGBA{R: 0xD7, G: 0x30, B: 0x25, A: 0xFF},
		Hover:     color.RGBA{R: 0xFF, G: 0xE5, B: 0xE5, A: 0xFF},
		Selection: color.RGBA{R: 0xFF, G: 0x3B, B: 0x30, A: 0x40},
	},
	"Teal": {
		Primary:   color.RGBA{R: 0x00, G: 0x96, B: 0x88, A: 0xFF},
		Pressed:   color.RGBA{R: 0x00, G: 0x79, B: 0x70, A: 0xFF},
		Hover:     color.RGBA{R: 0xE0, G: 0xF2, B: 0xF1, A: 0xFF},
		Selection: color.RGBA{R: 0x00, G: 0x96, B: 0x88, A: 0x40},
	},
}

// accentKey генерирует ключ для сохранения цвета в Preferences
func accentKey(name string) string {
	return "gui.theme.accent." + name
}

// saveAccentColor сохраняет акцентный цвет в Preferences
func (a *App) saveAccentColor(name string, c color.RGBA) {
	if a == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	val := uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
	p.SetString(accentKey(name), strconv.FormatUint(uint64(val), 16))
}

// loadAccentColor загружает акцентный цвет из Preferences
func (a *App) loadAccentColor(name string, defaultColor color.RGBA) color.RGBA {
	if a == nil || a.myApp == nil {
		return defaultColor
	}
	p := a.myApp.Preferences()
	raw := strings.TrimSpace(p.String(accentKey(name)))
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

// saveAccentPreset сохраняет выбранный пресет в Preferences
func (a *App) saveAccentPreset(name string) {
	if a == nil || a.myApp == nil {
		return
	}
	a.myApp.Preferences().SetString("gui.theme.accent_preset", name)
}

// loadAccentPreset загружает выбранный пресет из Preferences
func (a *App) loadAccentPreset() string {
	if a == nil || a.myApp == nil {
		return "macOS Blue"
	}
	p := a.myApp.Preferences()
	name := strings.TrimSpace(p.StringWithFallback("gui.theme.accent_preset", "macOS Blue"))
	// Проверяем, существует ли пресет
	if _, ok := PresetThemes[name]; !ok {
		return "macOS Blue"
	}
	return name
}

// applyAccentPreset применяет акцентный пресет к ModernTheme
func (t *ModernTheme) ApplyAccentPreset(name string, isDark bool) {
	preset, ok := PresetThemes[name]
	if !ok {
		return
	}
	if isDark {
		t.darkAccent = preset
	} else {
		t.lightAccent = preset
	}
}

// lightAccent акцентные цвета для светлой темы
var lightAccent = DefaultAccentColorsLight()

// darkAccent акцентные цвета для темной темы
var darkAccent = DefaultAccentColorsDark()
