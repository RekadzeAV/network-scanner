package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ThemeMode определяет режим темы
type ThemeMode string

const (
	ThemeModeLight  ThemeMode = "light"
	ThemeModeDark   ThemeMode = "dark"
	ThemeModeSystem ThemeMode = "system"
	prefThemeMode             = "gui.theme.mode"
)

// applyTheme применяет тему к приложению
func (a *App) applyTheme(mode ThemeMode) {
	if a == nil || a.myApp == nil {
		return
	}
	var isDark bool
	switch mode {
	case ThemeModeDark:
		isDark = true
	case ThemeModeLight:
		isDark = false
	case ThemeModeSystem:
		// Fyne автоматически определяет тему системы через DefaultTheme
		a.myApp.Settings().SetTheme(nil) // nil = использовать системную
		return
	default:
		return
	}
	theme := NewModernTheme(isDark)
	// Применяем сохраненный пресет акцентных цветов
	presetName := a.loadAccentPreset()
	theme.ApplyAccentPreset(presetName, isDark)
	a.myApp.Settings().SetTheme(theme)
}

// applyAccentPreset применяет пресет акцентных цветов
func (a *App) applyAccentPreset(name string) {
	if a == nil || a.myApp == nil {
		return
	}
	a.saveAccentPreset(name)
	// Перезагружаем текущую тему с новым пресетом
	mode := a.loadTheme()
	a.applyTheme(mode)
	if a.statusLabel != nil {
		a.statusLabel.SetText("Акцентный цвет: " + name)
	}
}

// loadTheme загружает сохраненную тему из Preferences
func (a *App) loadTheme() ThemeMode {
	if a == nil || a.myApp == nil {
		return ThemeModeSystem
	}
	p := a.myApp.Preferences()
	mode := ThemeMode(p.StringWithFallback(prefThemeMode, string(ThemeModeSystem)))
	return mode
}

// saveTheme сохраняет тему в Preferences
func (a *App) saveTheme(mode ThemeMode) {
	if a == nil || a.myApp == nil {
		return
	}
	a.myApp.Preferences().SetString(prefThemeMode, string(mode))
}

// createThemeSelector создает селектор темы для меню или тулбара
func (a *App) createThemeSelector(items []string, current ThemeMode) *fyne.Container {
	if a == nil || a.myWindow == nil {
		return nil
	}
	selected := string(current)
	for _, item := range items {
		if item == string(current) {
			selected = item
			break
		}
	}
	sel := widget.NewSelect(items, func(value string) {
		if a == nil {
			return
		}
		mode := ThemeMode(value)
		a.applyTheme(mode)
		a.saveTheme(mode)
		if a.statusLabel != nil {
			a.statusLabel.SetText("Тема изменена: " + value)
		}
	})
	sel.SetSelected(selected)
	return container.NewHBox(widget.NewLabel("Тема:"), sel)
}
