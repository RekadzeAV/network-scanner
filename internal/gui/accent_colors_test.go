package gui_test

import (
	"image/color"
	"testing"

	"network-scanner/internal/gui"
)

func TestDefaultAccentColorsLight(t *testing.T) {
	accent := gui.DefaultAccentColorsLight()
	expected := color.RGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF}
	if accent.Primary != expected {
		t.Errorf("Primary = %v, want %v", accent.Primary, expected)
	}
}

func TestDefaultAccentColorsDark(t *testing.T) {
	accent := gui.DefaultAccentColorsDark()
	expected := color.RGBA{R: 0x00, G: 0x78, B: 0xD4, A: 0xFF}
	if accent.Primary != expected {
		t.Errorf("Primary = %v, want %v", accent.Primary, expected)
	}
}

func TestPresetThemesExist(t *testing.T) {
	expectedPresets := []string{
		"macOS Blue", "Windows 11", "Green", "Purple",
		"Orange", "Red", "Teal",
	}
	for _, name := range expectedPresets {
		if _, ok := gui.PresetThemes[name]; !ok {
			t.Errorf("Preset %q not found", name)
		}
	}
}

func TestAccentColors_AllPresetsHaveValidColors(t *testing.T) {
	for name, colors := range gui.PresetThemes {
		if colors.Primary.A == 0 {
			t.Errorf("Preset %q: Primary has zero alpha", name)
		}
		if colors.Pressed.A == 0 {
			t.Errorf("Preset %q: Pressed has zero alpha", name)
		}
		if colors.Hover.A == 0 {
			t.Errorf("Preset %q: Hover has zero alpha", name)
		}
		if colors.Selection.A == 0 {
			t.Errorf("Preset %q: Selection has zero alpha", name)
		}
	}
}
