package gui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewModernTheme_Light(t *testing.T) {
	tObj := NewModernTheme(false)
	if tObj == nil {
		t.Fatal("expected non-nil theme")
	}
	if tObj.isDark {
		t.Error("expected light theme")
	}
}

func TestNewModernTheme_Dark(t *testing.T) {
	tObj := NewModernTheme(true)
	if tObj == nil {
		t.Fatal("expected non-nil theme")
	}
	if !tObj.isDark {
		t.Error("expected dark theme")
	}
}

func TestModernTheme_Variant_Light(t *testing.T) {
	tObj := NewModernTheme(false)
	v := tObj.Variant()
	if v != theme.VariantLight {
		t.Errorf("expected VariantLight, got %v", v)
	}
}

func TestModernTheme_Variant_Dark(t *testing.T) {
	tObj := NewModernTheme(true)
	v := tObj.Variant()
	if v != theme.VariantDark {
		t.Errorf("expected VariantDark, got %v", v)
	}
}

func TestModernTheme_Color_Light(t *testing.T) {
	tObj := NewModernTheme(false)
	c := tObj.Color(theme.ColorNameForeground, theme.VariantLight)
	if c == nil {
		t.Error("expected non-nil color")
	}
}

func TestModernTheme_Color_Dark(t *testing.T) {
	tObj := NewModernTheme(true)
	c := tObj.Color(theme.ColorNameForeground, theme.VariantDark)
	if c == nil {
		t.Error("expected non-nil color")
	}
}

func TestModernTheme_Color_Error_Light(t *testing.T) {
	tObj := NewModernTheme(false)
	c := tObj.Color(theme.ColorNameError, theme.VariantLight)
	if c == nil {
		t.Error("expected non-nil error color")
	}
}

func TestModernTheme_Color_Error_Dark(t *testing.T) {
	tObj := NewModernTheme(true)
	c := tObj.Color(theme.ColorNameError, theme.VariantDark)
	if c == nil {
		t.Error("expected non-nil error color")
	}
}

func TestModernTheme_Color_Success_Light(t *testing.T) {
	tObj := NewModernTheme(false)
	c := tObj.Color(theme.ColorNameSuccess, theme.VariantLight)
	if c == nil {
		t.Error("expected non-nil success color")
	}
}

func TestModernTheme_Color_Success_Dark(t *testing.T) {
	tObj := NewModernTheme(true)
	c := tObj.Color(theme.ColorNameSuccess, theme.VariantDark)
	if c == nil {
		t.Error("expected non-nil success color")
	}
}

func TestModernTheme_Color_Unknown(t *testing.T) {
	tObj := NewModernTheme(false)
	c := tObj.Color("unknown_color", theme.VariantLight)
	if c == nil {
		t.Error("expected fallback color for unknown name")
	}
}

func TestModernTheme_Color_Background_Light(t *testing.T) {
	tObj := NewModernTheme(false)
	c := tObj.Color(theme.ColorNameBackground, theme.VariantLight).(color.RGBA)
	if c.R != 0xF5 || c.G != 0xF5 || c.B != 0xF7 {
		t.Errorf("expected macOS Sierra background, got RGB(%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestModernTheme_Color_Background_Dark(t *testing.T) {
	tObj := NewModernTheme(true)
	c := tObj.Color(theme.ColorNameBackground, theme.VariantDark).(color.RGBA)
	if c.R != 0x20 || c.G != 0x20 || c.B != 0x20 {
		t.Errorf("expected Windows 11 dark background, got RGB(%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestModernTheme_Color_Foreground_Light(t *testing.T) {
	tObj := NewModernTheme(false)
	c := tObj.Color(theme.ColorNameForeground, theme.VariantLight).(color.RGBA)
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("expected black foreground, got RGB(%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestModernTheme_Color_Foreground_Dark(t *testing.T) {
	tObj := NewModernTheme(true)
	c := tObj.Color(theme.ColorNameForeground, theme.VariantDark).(color.RGBA)
	if c.R != 0xFF || c.G != 0xFF || c.B != 0xFF {
		t.Errorf("expected white foreground, got RGB(%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestModernTheme_Font(t *testing.T) {
	tObj := NewModernTheme(false)
	f := tObj.Font(fyne.TextStyle{})
	if f == nil {
		t.Error("expected non-nil font")
	}
}
