package gui

import (
	"image/color"
	"testing"
)

// --- accent_colors.go tests ---

func TestSaveAccentColor_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.saveAccentColor("test", color.RGBA{R: 255, G: 0, B: 0, A: 255})
}

func TestSaveAccentColor_NilMyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.saveAccentColor("test", color.RGBA{R: 255, G: 0, B: 0, A: 255})
}

func TestLoadAccentColor_NilApp(t *testing.T) {
	var a *App
	def := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	result := a.loadAccentColor("test", def)
	if result != def {
		t.Error("expected default color for nil app")
	}
}

func TestLoadAccentColor_NilMyApp(t *testing.T) {
	a := &App{}
	def := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	result := a.loadAccentColor("test", def)
	if result != def {
		t.Error("expected default color for nil myApp")
	}
}

func TestSaveAccentPreset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.saveAccentPreset("macOS Blue")
}

func TestSaveAccentPreset_NilMyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.saveAccentPreset("macOS Blue")
}

func TestLoadAccentPreset_NilApp(t *testing.T) {
	var a *App
	result := a.loadAccentPreset()
	if result != "macOS Blue" {
		t.Errorf("expected 'macOS Blue', got %q", result)
	}
}

func TestLoadAccentPreset_NilMyApp(t *testing.T) {
	a := &App{}
	result := a.loadAccentPreset()
	if result != "macOS Blue" {
		t.Errorf("expected 'macOS Blue', got %q", result)
	}
}

func TestApplyAccentPreset_UnknownPreset(t *testing.T) {
	tm := &ModernTheme{}
	tm.ApplyAccentPreset("nonexistent", false)
	// Не должен паниковать
}

func TestApplyAccentPreset_ValidPreset_Light(t *testing.T) {
	tm := &ModernTheme{}
	tm.ApplyAccentPreset("macOS Blue", false)
	// Не должен паниковать
}

func TestApplyAccentPreset_ValidPreset_Dark(t *testing.T) {
	tm := &ModernTheme{}
	tm.ApplyAccentPreset("macOS Blue", true)
	// Не должен паниковать
}

func TestDefaultAccentColorsLight_NotEmpty(t *testing.T) {
	colors := DefaultAccentColorsLight()
	if colors.Primary.A == 0 {
		t.Error("expected non-zero alpha for Primary")
	}
}

func TestDefaultAccentColorsDark_NotEmpty(t *testing.T) {
	colors := DefaultAccentColorsDark()
	if colors.Primary.A == 0 {
		t.Error("expected non-zero alpha for Primary")
	}
}
