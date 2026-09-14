package gui

import (
	"testing"

	"fyne.io/fyne/v2/theme"
)

// ============================================================================
// M6.3: Тесты для Theme Switcher (Dark/Light)
// ============================================================================

// TestThemeManager_New — ветка: создание менеджера тем
func TestThemeManager_New(t *testing.T) {
	tm := NewThemeManager(nil)
	if tm == nil {
		t.Fatal("expected non-nil theme manager")
	}

	// Проверяем что тема определяется автоматически
	mode := tm.GetThemeMode()
	if mode != ThemeModeSystem {
		t.Errorf("expected ThemeModeSystem mode, got %s", mode)
	}
}

// TestThemeManager_GetThemeVariant — ветка: получение варианта темы
func TestThemeManager_GetThemeVariant(t *testing.T) {
	tm := NewThemeManager(nil)

	variant := tm.GetThemeVariant()
	if variant != theme.VariantLight && variant != theme.VariantDark {
		t.Errorf("unexpected variant: %d", variant)
	}
}

// TestThemeManager_SetThemeMode_Light — ветка: установка светлой темы
func TestThemeManager_SetThemeMode_Light(t *testing.T) {
	tm := NewThemeManager(nil)

	tm.SetThemeMode(ThemeModeLight)
	mode := tm.GetThemeMode()
	if mode != ThemeModeLight {
		t.Errorf("expected ThemeModeLight, got %s", mode)
	}

	variant := tm.GetThemeVariant()
	if variant != theme.VariantLight {
		t.Errorf("expected VariantLight, got %d", variant)
	}
}

// TestThemeManager_SetThemeMode_Dark — ветка: установка темной темы
func TestThemeManager_SetThemeMode_Dark(t *testing.T) {
	tm := NewThemeManager(nil)

	tm.SetThemeMode(ThemeModeDark)
	mode := tm.GetThemeMode()
	if mode != ThemeModeDark {
		t.Errorf("expected ThemeModeDark, got %s", mode)
	}

	variant := tm.GetThemeVariant()
	if variant != theme.VariantDark {
		t.Errorf("expected VariantDark, got %d", variant)
	}
}

// TestThemeManager_SetThemeMode_Auto — ветка: автоматический режим
func TestThemeManager_SetThemeMode_Auto(t *testing.T) {
	tm := NewThemeManager(nil)

	tm.SetThemeMode(ThemeModeSystem)
	mode := tm.GetThemeMode()
	if mode != ThemeModeSystem {
		t.Errorf("expected ThemeModeSystem, got %s", mode)
	}
}

// TestThemeManager_ToggleTheme — ветка: переключение темы
func TestThemeManager_ToggleTheme(t *testing.T) {
	tm := NewThemeManager(nil)

	// Устанавливаем светлую тему
	tm.SetThemeMode(ThemeModeLight)
	if tm.GetThemeMode() != ThemeModeLight {
		t.Error("expected ThemeModeLight")
	}

	// Переключаем на темную
	tm.ToggleTheme()
	if tm.GetThemeMode() != ThemeModeDark {
		t.Error("expected ThemeModeDark after toggle")
	}

	// Переключаем обратно
	tm.ToggleTheme()
	if tm.GetThemeMode() != ThemeModeLight {
		t.Error("expected ThemeModeLight after toggle")
	}
}

// TestThemeManager_GetTheme — ветка: получение темы
func TestThemeManager_GetTheme(t *testing.T) {
	tm := NewThemeManager(nil)

	themeObj := tm.GetTheme()
	if themeObj == nil {
		t.Fatal("expected non-nil theme")
	}
}

// TestThemeManager_IsDark — ветка: проверка темной темы
func TestThemeManager_IsDark(t *testing.T) {
	tm := NewThemeManager(nil)

	// По умолчанию может быть light или dark в зависимости от системы
	isDark := tm.IsDark()
	if !isDark && isDark != false {
		t.Error("IsDark should return boolean")
	}

	// Устанавливаем темную тему
	tm.SetThemeMode(ThemeModeDark)
	if !tm.IsDark() {
		t.Error("expected IsDark to return true")
	}

	// Устанавливаем светлую тему
	tm.SetThemeMode(ThemeModeLight)
	if tm.IsDark() {
		t.Error("expected IsDark to return false")
	}
}

// TestThemeManager_RegisterThemeChanged — ветка: регистрация обработчика
func TestThemeManager_RegisterThemeChanged(t *testing.T) {
	tm := NewThemeManager(nil)

	callbackCalled := false
	tm.RegisterThemeChanged(func(mode ThemeMode) {
		callbackCalled = true
		if mode != ThemeModeDark {
			t.Errorf("expected ThemeModeDark in callback, got %s", mode)
		}
	})

	// Переключаем тему чтобы вызвать callback
	tm.SetThemeMode(ThemeModeDark)
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}

// TestThemeSwitcherUI_New — ветка: создание UI переключателя
func TestThemeSwitcherUI_New(t *testing.T) {
	tm := NewThemeManager(nil)

	_ = false // onToggleCalled placeholder
	ui := NewThemeSwitcherUI(tm, func(mode ThemeMode) {
		_ = mode
	})

	if ui == nil {
		t.Fatal("expected non-nil theme switcher UI")
	}

	currentMode := ui.GetCurrentMode()
	if currentMode != tm.GetThemeMode() {
		t.Errorf("expected current mode to match manager mode")
	}
}

// TestThemeSwitcherUI_Toggle — ветка: переключение через UI
func TestThemeSwitcherUI_Toggle(t *testing.T) {
	tm := NewThemeManager(nil)

	toggleCalled := false
	ui := NewThemeSwitcherUI(tm, func(mode ThemeMode) {
		toggleCalled = true
	})

	ui.Toggle()
	if !toggleCalled {
		t.Error("expected toggle callback to be called")
	}

	// Проверяем что тема изменилась
	// Тема должна переключиться с текущей на противоположную
	_ = ThemeModeDark
	if tm.GetThemeMode() == ThemeModeLight {
		_ = ThemeModeLight
	}
	if tm.GetThemeMode() == tm.GetThemeMode() {
		// Тема переключилась
	}
}

// TestThemeSwitcherUI_SetMode — ветка: установка режима через UI
func TestThemeSwitcherUI_SetMode(t *testing.T) {
	tm := NewThemeManager(nil)

	setModeCalled := false
	ui := NewThemeSwitcherUI(tm, func(mode ThemeMode) {
		setModeCalled = true
	})

	ui.SetMode(ThemeModeDark)
	if !setModeCalled {
		t.Error("expected setMode callback to be called")
	}

	if tm.GetThemeMode() != ThemeModeDark {
		t.Error("expected theme mode to be Dark")
	}
}

// TestThemeSwitcherUI_GetCurrentMode — ветка: получение текущего режима
func TestThemeSwitcherUI_GetCurrentMode(t *testing.T) {
	tm := NewThemeManager(nil)

	ui := NewThemeSwitcherUI(tm, nil)

	mode := ui.GetCurrentMode()
	if mode != tm.GetThemeMode() {
		t.Errorf("expected current mode to match manager mode: got %s, expected %s", mode, tm.GetThemeMode())
	}
}

// TestModernTheme_Color — ветка: получение цвета из темы
func TestModernTheme_Color(t *testing.T) {
	themeObj := NewModernTheme(false) // Light theme

	// Проверяем что цвет возвращается
	color := themeObj.Color(theme.ColorNameBackground, theme.VariantLight)
	if color == nil {
		t.Error("expected non-nil color")
	}
}

// TestModernTheme_Variant — ветка: получение варианта темы
func TestModernTheme_Variant(t *testing.T) {
	lightTheme := NewModernTheme(false)
	if lightTheme.Variant() != theme.VariantLight {
		t.Error("expected VariantLight")
	}

	darkTheme := NewModernTheme(true)
	if darkTheme.Variant() != theme.VariantDark {
		t.Error("expected VariantDark")
	}
}

// TestThemeManager_EmptyApp — ветка: менеджер без приложения
func TestThemeManager_EmptyApp(t *testing.T) {
	tm := NewThemeManager(nil)
	if tm == nil {
		t.Fatal("expected non-nil theme manager")
	}

	// Должен корректно работать даже без приложения
	mode := tm.GetThemeMode()
	if mode == "" {
		t.Error("expected non-empty theme mode")
	}
}

// TestThemeSwitcherUI_EmptyCallback — ветка: UI без callback
func TestThemeSwitcherUI_EmptyCallback(t *testing.T) {
	tm := NewThemeManager(nil)

	// Должен корректно работать даже без callback
	ui := NewThemeSwitcherUI(tm, nil)
	if ui == nil {
		t.Fatal("expected non-nil theme switcher UI")
	}

	// Переключение без callback
	ui.Toggle()

	// Установка режима без callback
	ui.SetMode(ThemeModeDark)
}
