package gui

import (
	"testing"
)

// --- accent_colors.go tests ---

func TestAccentKey_Blue(t *testing.T) {
	key := accentKey("blue")
	expected := "gui.theme.accent.blue"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestAccentKey_Red(t *testing.T) {
	key := accentKey("red")
	expected := "gui.theme.accent.red"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestAccentKey_Green(t *testing.T) {
	key := accentKey("green")
	expected := "gui.theme.accent.green"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestAccentKey_Empty(t *testing.T) {
	key := accentKey("")
	expected := "gui.theme.accent."
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestAccentKey_Spaces(t *testing.T) {
	key := accentKey("blue sky")
	expected := "gui.theme.accent.blue sky"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestAccentKey_WithNumber(t *testing.T) {
	key := accentKey("blue2")
	expected := "gui.theme.accent.blue2"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestAccentKey_Consistency(t *testing.T) {
	key1 := accentKey("blue")
	key2 := accentKey("blue")
	if key1 != key2 {
		t.Error("expected same key for same name")
	}
}

func TestAccentKey_DifferentNames(t *testing.T) {
	key1 := accentKey("blue")
	key2 := accentKey("red")
	if key1 == key2 {
		t.Error("expected different keys for different names")
	}
}
