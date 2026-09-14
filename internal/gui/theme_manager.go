package gui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeManager управляет переключением тем и хранит состояние
type ThemeManager struct {
	mu             sync.RWMutex
	currentMode    ThemeMode
	currentVariant fyne.ThemeVariant
	themes         map[fyne.ThemeVariant]*ModernTheme
	onThemeChanged []func(mode ThemeMode)
	app            fyne.App
}

// NewThemeManager создает новый менеджер тем
func NewThemeManager(app fyne.App) *ThemeManager {
	tm := &ThemeManager{
		currentMode: ThemeModeSystem,
		themes: map[fyne.ThemeVariant]*ModernTheme{
			theme.VariantLight: NewModernTheme(false),
			theme.VariantDark:  NewModernTheme(true),
		},
		app: app,
	}

	// Автоматически определяем системную тему
	tm.autoDetectSystemTheme()

	return tm
}

// autoDetectSystemTheme определяет тему из системных настроек
func (tm *ThemeManager) autoDetectSystemTheme() {
	if tm.app == nil {
		return
	}

	settings := tm.app.Settings()
	if settings == nil {
		return
	}

	variant := settings.ThemeVariant()
	if variant == theme.VariantDark {
		tm.currentVariant = theme.VariantDark
		tm.currentMode = ThemeModeDark
	} else {
		tm.currentVariant = theme.VariantLight
		tm.currentMode = ThemeModeLight
	}
}

// GetThemeVariant возвращает текущий вариант темы
func (tm *ThemeManager) GetThemeVariant() fyne.ThemeVariant {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentVariant
}

// GetThemeMode возвращает текущий режим темы
func (tm *ThemeManager) GetThemeMode() ThemeMode {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentMode
}

// SetThemeMode устанавливает режим темы
func (tm *ThemeManager) SetThemeMode(mode ThemeMode) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.currentMode = mode

	switch mode {
	case ThemeModeLight:
		tm.currentVariant = theme.VariantLight
	case ThemeModeDark:
		tm.currentVariant = theme.VariantDark
	case ThemeModeSystem:
		// Автоматическое определение из системных настроек
		if tm.app != nil {
			variant := tm.app.Settings().ThemeVariant()
			tm.currentVariant = variant
		}
	}

	tm.notifyThemeChanged()
}

// ToggleTheme переключает тему между светлой и темной
func (tm *ThemeManager) ToggleTheme() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.currentVariant == theme.VariantLight {
		tm.currentVariant = theme.VariantDark
		tm.currentMode = ThemeModeDark
	} else {
		tm.currentVariant = theme.VariantLight
		tm.currentMode = ThemeModeLight
	}

	tm.notifyThemeChanged()
}

// GetTheme возвращает тему для текущего варианта
func (tm *ThemeManager) GetTheme() *ModernTheme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if themeObj, ok := tm.themes[tm.currentVariant]; ok {
		return themeObj
	}

	// Fallback на светлую тему
	return tm.themes[theme.VariantLight]
}

// RegisterThemeChanged регистрирует обработчик изменения темы
func (tm *ThemeManager) RegisterThemeChanged(callback func(mode ThemeMode)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onThemeChanged = append(tm.onThemeChanged, callback)
}

// notifyThemeChanged уведомляет об изменении темы
func (tm *ThemeManager) notifyThemeChanged() {
	mode := tm.currentMode
	for _, callback := range tm.onThemeChanged {
		callback(mode)
	}
}

// IsDark возвращает true если текущая тема темная
func (tm *ThemeManager) IsDark() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentVariant == theme.VariantDark
}

// ThemeSwitcherUI представляет UI компонент переключения тем
type ThemeSwitcherUI struct {
	manager  *ThemeManager
	current  ThemeMode
	onToggle func(mode ThemeMode)
}

// NewThemeSwitcherUI создает новый UI переключатель тем
func NewThemeSwitcherUI(manager *ThemeManager, onToggle func(mode ThemeMode)) *ThemeSwitcherUI {
	return &ThemeSwitcherUI{
		manager:  manager,
		current:  manager.GetThemeMode(),
		onToggle: onToggle,
	}
}

// Toggle переключает тему
func (ts *ThemeSwitcherUI) Toggle() {
	ts.manager.ToggleTheme()
	ts.current = ts.manager.GetThemeMode()
	if ts.onToggle != nil {
		ts.onToggle(ts.current)
	}
}

// SetMode устанавливает режим темы
func (ts *ThemeSwitcherUI) SetMode(mode ThemeMode) {
	ts.manager.SetThemeMode(mode)
	ts.current = mode
	if ts.onToggle != nil {
		ts.onToggle(mode)
	}
}

// GetCurrentMode возвращает текущий режим
func (ts *ThemeSwitcherUI) GetCurrentMode() ThemeMode {
	return ts.current
}
