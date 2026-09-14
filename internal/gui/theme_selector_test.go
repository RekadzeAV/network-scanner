package gui

import (
	"testing"
)

// --- theme_selector.go tests ---

func TestThemeMode_Constants(t *testing.T) {
	if ThemeModeLight != "light" {
		t.Errorf("expected 'light', got %q", ThemeModeLight)
	}
	if ThemeModeDark != "dark" {
		t.Errorf("expected 'dark', got %q", ThemeModeDark)
	}
	if ThemeModeSystem != "system" {
		t.Errorf("expected 'system', got %q", ThemeModeSystem)
	}
}

func TestApp_applyTheme_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.applyTheme(ThemeModeLight)
}

func TestApp_applyTheme_NilMyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.applyTheme(ThemeModeDark)
}

func TestApp_applyAccentPreset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.applyAccentPreset("blue")
}

func TestApp_applyAccentPreset_NilMyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.applyAccentPreset("blue")
}

func TestApp_loadTheme_NilApp(t *testing.T) {
	var a *App
	mode := a.loadTheme()
	// Должен вернуть ThemeModeSystem по умолчанию
	if mode != ThemeModeSystem {
		t.Errorf("expected ThemeModeSystem for nil app, got %q", mode)
	}
}

func TestApp_loadTheme_NilMyApp(t *testing.T) {
	a := &App{}
	mode := a.loadTheme()
	// Должен вернуть ThemeModeSystem по умолчанию
	if mode != ThemeModeSystem {
		t.Errorf("expected ThemeModeSystem for nil myApp, got %q", mode)
	}
}

func TestApp_saveTheme_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.saveTheme(ThemeModeDark)
}

func TestApp_saveTheme_NilMyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.saveTheme(ThemeModeLight)
}

func TestApp_createThemeSelector_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.createThemeSelector([]string{"light", "dark"}, ThemeModeLight)
}

func TestApp_createThemeSelector_NilWindow(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.createThemeSelector([]string{"light", "dark"}, ThemeModeLight)
}

func TestApp_createThemeSelector_EmptyItems(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.createThemeSelector(nil, ThemeModeLight)
}

func TestApp_createThemeSelector_Valid(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.createThemeSelector([]string{"light", "dark", "system"}, ThemeModeDark)
}

func TestApp_createThemeSelector_UnmatchedCurrent(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.createThemeSelector([]string{"light", "dark"}, ThemeModeSystem)
}
