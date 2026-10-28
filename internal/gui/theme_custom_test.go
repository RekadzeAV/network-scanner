package gui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/theme"
)

// --- isValidHexColor ---

func TestIsValidHexColor_Valid(t *testing.T) {
	for _, s := range []string{"#000000", "1a2b3c", "#1A2B3C", " ff00aa ", "#000000", "#ffffff", "#ABCDEF"} {
		if !isValidHexColor(s) {
			t.Errorf("isValidHexColor(%q) = false, want true", s)
		}
	}
}

func TestIsValidHexColor_Invalid(t *testing.T) {
	for _, s := range []string{"", "   ", "#12345", "#1234567", "gggggg", "#GGG", "12345g", "#ff00", "red", "#12345g"} {
		if isValidHexColor(s) {
			t.Errorf("isValidHexColor(%q) = true, want false", s)
		}
	}
}

// --- parseHexColorStrict / parseHexColor ---

func TestParseHexColorStrict(t *testing.T) {
	c, err := parseHexColorStrict("#1A2B3C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.R != 0x1A || c.G != 0x2B || c.B != 0x3C || c.A != 0xFF {
		t.Errorf("got %+v", c)
	}
	// Без # тоже принимается.
	c2, err := parseHexColorStrict("ff8800")
	if err != nil || c2.R != 0xFF || c2.G != 0x88 || c2.B != 0x00 {
		t.Errorf("got %+v, err %v", c2, err)
	}
	if _, err := parseHexColorStrict("bad-input"); err == nil {
		t.Error("expected error for invalid input")
	}
}

// --- normalizeHex ---

func TestNormalizeHex(t *testing.T) {
	cases := map[string]string{
		"#1a2b3c":  "#1A2B3C",
		"1a2b3c":   "#1A2B3C",
		" #FF0000": "#FF0000",
		"":         "",
	}
	for in, want := range cases {
		if got := normalizeHex(in); got != want {
			t.Errorf("normalizeHex(%q) = %q, want %q", in, got, want)
		}
	}
}

// --- blendTo ---

func TestBlendTo(t *testing.T) {
	base := color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xFF}
	// t=0 — без изменений.
	if got := blendTo(base, 0xFF, 0); got != base {
		t.Errorf("blendTo t=0 = %+v, want %+v", got, base)
	}
	// t=1 — целевой цвет.
	got := blendTo(base, 0xFF, 1)
	if got.R != 0xFF || got.G != 0xFF || got.B != 0xFF {
		t.Errorf("blendTo t=1 = %+v, want all 0xFF", got)
	}
	// Альфа сохраняется.
	if got.A != 0xFF {
		t.Errorf("blendTo changed alpha: %+v", got)
	}
	// Клиппинг t за пределами [0,1].
	if got := blendTo(base, 0x00, -1); got != base {
		t.Errorf("blendTo t=-1 = %+v, want unchanged", got)
	}
	if got := blendTo(base, 0x00, 2); got.R != 0 {
		t.Errorf("blendTo t=2 = %+v, want fully blended", got)
	}
}

// --- adjustAccentForVariant ---

func TestAdjustAccentForVariant(t *testing.T) {
	// Очень тёмный цвет в тёмном варианте должен осветляться.
	darkInput := color.RGBA{R: 0x10, G: 0x10, B: 0x10, A: 0xFF}
	lightened := adjustAccentForVariant(darkInput, true)
	if srgbLuminance(lightened) <= srgbLuminance(darkInput) {
		t.Errorf("dark variant should brighten dark accent: %+v -> %+v", darkInput, lightened)
	}
	// Очень светлый цвет в светлом варианте должен затемняться.
	lightInput := color.RGBA{R: 0xF0, G: 0xF0, B: 0xF0, A: 0xFF}
	darkened := adjustAccentForVariant(lightInput, false)
	if srgbLuminance(darkened) >= srgbLuminance(lightInput) {
		t.Errorf("light variant should darken light accent: %+v -> %+v", lightInput, darkened)
	}
	// Средний цвет остаётся без изменений в обоих вариантах.
	mid := color.RGBA{R: 0x40, G: 0x60, B: 0x80, A: 0xFF}
	if got := adjustAccentForVariant(mid, false); got != mid {
		t.Errorf("mid-tone light accent changed: %+v -> %+v", mid, got)
	}
	if got := adjustAccentForVariant(mid, true); got != mid {
		t.Errorf("mid-tone dark accent changed: %+v -> %+v", mid, got)
	}
}

// --- deriveAccentVariant ---

func TestDeriveAccentVariant_Contrast(t *testing.T) {
	// Для любого валидного акцента onPrimary остаётся читаемым (WCAG 3.0 для UI).
	bases := []string{"#FF0000", "#00FF00", "#0000FF", "#FFFF00", "#1A1A1A", "#E6E6E6", "#123456", "#ABCDEF"}
	for _, hex := range bases {
		base := parseHexColor(hex)
		for _, dark := range []bool{false, true} {
			v := deriveAccentVariant(base, dark)
			if r := ContrastRatio(v.OnPrimary, v.Primary); r < 3.0 {
				t.Errorf("%s dark=%v: onPrimary/primary contrast %.2f < 3.0", hex, dark, r)
			}
		}
	}
}

func TestDeriveAccentVariant_Structure(t *testing.T) {
	v := deriveAccentVariant(parseHexColor("#0A5FCC"), false)
	// Pressed темнее primary.
	if srgbLuminance(v.Pressed) >= srgbLuminance(v.Primary) {
		t.Error("pressed must be darker than primary")
	}
	// Hover — светлее primary в светлом варианте.
	if srgbLuminance(v.Hover) <= srgbLuminance(v.Primary) {
		t.Error("light hover must be lighter than primary")
	}
	// Selection = primary с альфой 0x40.
	if v.Selection.A != 0x40 || v.Selection.R != v.Primary.R {
		t.Errorf("selection = %+v, want primary rgb with alpha 0x40", v.Selection)
	}
	// Тёмный вариант: hover — тёмный tint.
	vd := deriveAccentVariant(parseHexColor("#0A5FCC"), true)
	if srgbLuminance(vd.Hover) >= srgbLuminance(vd.Primary) {
		t.Error("dark hover must be darker tint of primary")
	}
}

// --- SetCustomAccent / интеграция с темой ---

func TestSetCustomAccent_Valid(t *testing.T) {
	th := NewModernTheme(false)
	if err := th.SetCustomAccent("#ff8800"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !th.IsCustomAccent() {
		t.Error("theme must be in Custom mode")
	}
	if th.CustomAccentHex() != "#FF8800" {
		t.Errorf("hex = %q, want #FF8800", th.CustomAccentHex())
	}
	if th.Accent() != customAccentName {
		t.Errorf("Accent() = %q, want Custom", th.Accent())
	}
	// Primary темы берётся из кастомного цвета.
	want := deriveAccentVariant(parseHexColor("#ff8800"), false).Primary
	got := themeColor(t, th, theme.ColorNamePrimary, theme.VariantLight)
	assertColor(t, got, want, "primary custom")
}

func TestSetCustomAccent_Invalid(t *testing.T) {
	th := NewModernTheme(false)
	before := th.Accent()
	if err := th.SetCustomAccent("not-a-color"); err == nil {
		t.Fatal("expected error")
	}
	if th.Accent() != before {
		t.Error("invalid input must not change accent")
	}
}

func TestSetCustomAccent_Clear(t *testing.T) {
	th := NewModernTheme(false)
	if err := th.SetCustomAccent("#112233"); err != nil {
		t.Fatal(err)
	}
	if err := th.SetCustomAccent(""); err != nil {
		t.Fatal(err)
	}
	if th.IsCustomAccent() {
		t.Error("empty hex must clear custom mode")
	}
	if th.CustomAccentHex() != "" {
		t.Error("hex must be cleared")
	}
	if th.Accent() != defaultAccentPreset {
		t.Errorf("accent = %q, want default", th.Accent())
	}
}

func TestSetCustomAccent_PresetSelectKeepsStoredHex(t *testing.T) {
	// Выбор пресета после custom не затирает сохранённый hex (он нужен,
	// если пользователь снова выберет «Произвольный цвет»).
	th := NewModernTheme(false)
	if err := th.SetCustomAccent("#445566"); err != nil {
		t.Fatal(err)
	}
	th.SetAccent("Teal")
	if th.IsCustomAccent() {
		t.Error("SetAccent must leave custom mode")
	}
	if th.CustomAccentHex() != "#445566" {
		t.Errorf("hex must be preserved, got %q", th.CustomAccentHex())
	}
	// Палитра при этом берётся из пресета.
	got := themeColor(t, th, theme.ColorNamePrimary, theme.VariantLight)
	assertColor(t, got, accentPresets["Teal"].Light.Primary, "primary after SetAccent Teal")
}

func TestNormalizeAccentPreset_Custom(t *testing.T) {
	if got := normalizeAccentPreset("Custom"); got != customAccentName {
		t.Errorf("Custom must survive normalization, got %q", got)
	}
	if got := normalizeAccentPreset(" custom "); got != defaultAccentPreset {
		t.Errorf("lowercase/custom-with-spaces must normalize to default, got %q", got)
	}
}

// --- App-level nil-safety (headless) ---

func TestApp_CustomAccent_NilSafe(t *testing.T) {
	var a *App
	a.saveCustomAccentHex("#ff0000") // не должен паниковать
	if got := a.loadCustomAccentHex(); got != "" {
		t.Errorf("nil app must return empty hex, got %q", got)
	}
	a.applyCustomAccentHex("#ff0000")
	a.applyCustomAccentHex("")
	a.applyCustomAccentHex("garbage")
}

func TestApp_ShowCustomDialogs_NilSafe(t *testing.T) {
	var a *App
	a.showAccentCustomDialog() // nil window — ранний возврат, без паники
	a.showThemePreviewDialog()
	// loadAccentPreset для nil app возвращает defaultAccentPreset,
	// поэтому подпись акцента — имя пресета по умолчанию.
	if got := a.currentAccentCaption(); got != defaultAccentPreset {
		t.Errorf("nil app caption = %q, want %q", got, defaultAccentPreset)
	}
}

func TestThemePreviewContent(t *testing.T) {
	// Проверка, что превью строится без паники (headless).
	obj := themePreviewContent("Custom")
	if obj == nil {
		t.Fatal("expected non-nil preview")
	}
}
