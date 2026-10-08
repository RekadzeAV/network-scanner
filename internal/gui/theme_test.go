package gui

import (
	"image/color"
	"strconv"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// themeColor берёт цвет у темы и приводит его к color.RGBA (значения токенов
// объявлены как color.RGBA; RGBAModel.Convert корректно обрабатывает и другие
// реализации color.Color, например color.NRGBA).
func themeColor(t *testing.T, th fyne.Theme, name fyne.ThemeColorName, variant fyne.ThemeVariant) color.RGBA {
	t.Helper()
	got := th.Color(name, variant)
	if got == nil {
		t.Fatalf("тема не вернула цвет для %q", name)
	}
	converted, ok := color.RGBAModel.Convert(got).(color.RGBA)
	if !ok {
		t.Fatalf("цвет %q не приводится к color.RGBA: %T", name, got)
	}
	return converted
}

func assertColor(t *testing.T, got, want color.RGBA, format string, args ...any) {
	t.Helper()
	if got != want {
		t.Errorf(format+": expected %+v, got %+v", append(args, want, got)...)
	}
}

func assertContrast(t *testing.T, name string, fg, bg color.RGBA, minimum float64) {
	t.Helper()
	if ratio := ContrastRatio(fg, bg); ratio < minimum {
		t.Errorf("%s: контраст %.2f < %.2f", name, ratio, minimum)
	}
}

// --- П0-1: палитра светлого варианта ----------------------------------------

func TestTheme_Color_LightPaletteMatchesTokens(t *testing.T) {
	th := NewModernTheme(false)
	v := theme.VariantLight

	assertColor(t, themeColor(t, th, theme.ColorNameBackground, v), tokLightBg, "background")
	assertColor(t, themeColor(t, th, theme.ColorNameForeground, v), tokLightText, "foreground")
	assertColor(t, themeColor(t, th, theme.ColorNameInputBackground, v), tokLightSurface, "input background")
	assertColor(t, themeColor(t, th, theme.ColorNameInputBorder, v), tokLightBorderStrong, "input border")
	assertColor(t, themeColor(t, th, theme.ColorNameSelection, v), accentPresets[defaultAccentPreset].Light.Selection, "selection")
	assertColor(t, themeColor(t, th, theme.ColorNameSeparator, v), tokLightBorderSubtle, "separator")
	assertColor(t, themeColor(t, th, theme.ColorNameScrollBar, v), tokScrollBarLt, "scroll bar")
	assertColor(t, themeColor(t, th, theme.ColorNameOverlayBackground, v), tokLightSurface, "overlay background")
	assertColor(t, themeColor(t, th, theme.ColorNameShadow, v), tokLightShadow, "shadow")
	assertColor(t, themeColor(t, th, theme.ColorNameDisabled, v), tokLightTextDisabled, "disabled")
	assertColor(t, themeColor(t, th, theme.ColorNameDisabledButton, v), tokLightSurfaceAlt, "disabled button")
	assertColor(t, themeColor(t, th, theme.ColorNamePlaceHolder, v), tokLightTextSecond, "placeholder")
	assertColor(t, themeColor(t, th, theme.ColorNamePrimary, v), tokLightPrimary, "primary")
	assertColor(t, themeColor(t, th, theme.ColorNameHyperlink, v), tokLightPrimary, "hyperlink")

	// Семантические токены за пределами стандартного интерфейса Fyne.
	assertColor(t, themeColor(t, th, ColorNameSurfaceAlt, v), tokLightSurfaceAlt, "surface alt")
	assertColor(t, themeColor(t, th, ColorNameBorderStrong, v), tokLightBorderStrong, "border strong")
	assertColor(t, themeColor(t, th, ColorNameOnPrimary, v), tokLightOnPrimary, "on primary")
	assertColor(t, themeColor(t, th, theme.ColorNameSuccess, v), tokLightSuccess, "success")
	assertColor(t, themeColor(t, th, ColorNameSuccessFill, v), tokLightSuccessFill, "success fill")
	assertColor(t, themeColor(t, th, theme.ColorNameWarning, v), tokLightWarning, "warning")
	assertColor(t, themeColor(t, th, ColorNameWarningFill, v), tokLightWarningFill, "warning fill")
	assertColor(t, themeColor(t, th, theme.ColorNameError, v), tokLightError, "error")
	assertColor(t, themeColor(t, th, ColorNameErrorFill, v), tokLightErrorFill, "error fill")
}

func TestTheme_Color_DarkPaletteMatchesTokens(t *testing.T) {
	th := NewModernTheme(true)
	v := theme.VariantLight // вариант Fyne намеренно светлый: тема явно тёмная

	assertColor(t, themeColor(t, th, theme.ColorNameBackground, v), tokDarkBg, "background")
	assertColor(t, themeColor(t, th, theme.ColorNameForeground, v), tokDarkText, "foreground")
	assertColor(t, themeColor(t, th, theme.ColorNameInputBackground, v), tokDarkSurface, "input background")
	assertColor(t, themeColor(t, th, theme.ColorNameInputBorder, v), tokDarkBorderStrong, "input border")
	assertColor(t, themeColor(t, th, theme.ColorNameSelection, v), accentPresets[defaultAccentPreset].Dark.Selection, "selection")
	assertColor(t, themeColor(t, th, theme.ColorNameSeparator, v), tokDarkBorderSubtle, "separator")
	assertColor(t, themeColor(t, th, theme.ColorNameScrollBar, v), tokScrollBarDk, "scroll bar")
	assertColor(t, themeColor(t, th, theme.ColorNameDisabled, v), tokDarkTextDisabled, "disabled")
	assertColor(t, themeColor(t, th, theme.ColorNameDisabledButton, v), tokDarkSurfaceAlt, "disabled button")
	assertColor(t, themeColor(t, th, theme.ColorNamePlaceHolder, v), tokDarkTextSecond, "placeholder")
	assertColor(t, themeColor(t, th, ColorNameSurfaceAlt, v), tokDarkSurfaceAlt, "surface alt")
	assertColor(t, themeColor(t, th, ColorNameBorderStrong, v), tokDarkBorderStrong, "border strong")
	assertColor(t, themeColor(t, th, ColorNameOnPrimary, v), tokDarkOnPrimary, "on primary")
	assertColor(t, themeColor(t, th, theme.ColorNameSuccess, v), tokDarkSuccess, "success")
	assertColor(t, themeColor(t, th, ColorNameSuccessFill, v), tokDarkSuccessFill, "success fill")
	assertColor(t, themeColor(t, th, theme.ColorNameWarning, v), tokDarkWarning, "warning")
	assertColor(t, themeColor(t, th, ColorNameWarningFill, v), tokDarkWarningFill, "warning fill")
	assertColor(t, themeColor(t, th, theme.ColorNameError, v), tokDarkError, "error")
	assertColor(t, themeColor(t, th, ColorNameErrorFill, v), tokDarkErrorFill, "error fill")
}

// --- П0-3: тема не зависит от варианта Fyne ----------------------------------

func TestTheme_ExplicitDarkIgnoresFyneVariant(t *testing.T) {
	// Без этого тёмная тема не активировалась: Fyne мог запрашивать Light.
	th := NewModernTheme(true)
	for _, v := range []fyne.ThemeVariant{theme.VariantLight, theme.VariantDark} {
		assertColor(t, themeColor(t, th, theme.ColorNameBackground, v), tokDarkBg,
			"явно тёмная тема при запросе варианта %v", v)
	}

	thLight := NewModernTheme(false)
	for _, v := range []fyne.ThemeVariant{theme.VariantLight, theme.VariantDark} {
		assertColor(t, themeColor(t, thLight, theme.ColorNameBackground, v), tokLightBg,
			"явно светлая тема при запросе варианта %v", v)
	}
}

func TestTheme_NewModernThemeForVariant(t *testing.T) {
	if got := NewModernThemeForVariant(theme.VariantDark, defaultAccentPreset); !got.IsDark() {
		t.Error("VariantDark must produce dark theme")
	}
	if got := NewModernThemeForVariant(theme.VariantLight, defaultAccentPreset); got.IsDark() {
		t.Error("VariantLight must produce light theme")
	}
}

func TestTheme_IsDark(t *testing.T) {
	if NewModernTheme(false).IsDark() {
		t.Error("light theme reports dark")
	}
	if !NewModernTheme(true).IsDark() {
		t.Error("dark theme reports light")
	}
}

// --- П0-5: акцент -------------------------------------------------------------

func TestTheme_AccentDefaultsToMacOSBlue(t *testing.T) {
	th := NewModernTheme(false)
	if got := th.Accent(); got != defaultAccentPreset {
		t.Errorf("expected %q, got %q", defaultAccentPreset, got)
	}
	assertColor(t, themeColor(t, th, theme.ColorNamePrimary, theme.VariantLight), tokLightPrimary,
		"default accent primary must equal palette primary")
}

func TestTheme_Accent_NormalizesUnknownPreset(t *testing.T) {
	for _, in := range []string{"", "   ", "Пусто", "macos blue"} {
		if got := normalizeAccentPreset(in); got != defaultAccentPreset {
			t.Errorf("normalizeAccentPreset(%q) = %q, want %q", in, got, defaultAccentPreset)
		}
	}
	for _, in := range accentPresetOrder {
		if got := normalizeAccentPreset(in); got != in {
			t.Errorf("normalizeAccentPreset(%q) = %q", in, got)
		}
	}
}

func TestTheme_Accent_ChangesPrimaryOnly(t *testing.T) {
	// Смена акцента меняет primary/hover/pressed/selection, но не семантические
	// цвета (success/warning/error) и не фоновые токены варианта.
	base := NewModernTheme(false)
	accented := NewModernThemeAccent(false, "Purple")

	if base.Accent() == accented.Accent() {
		t.Fatal("accent was not applied")
	}
	want := accentPresets["Purple"].Light.Primary
	got := themeColor(t, accented, theme.ColorNamePrimary, theme.VariantLight)
	assertColor(t, got, want, "primary must follow accent")

	for _, name := range []fyne.ThemeColorName{
		theme.ColorNameSuccess, theme.ColorNameWarning, theme.ColorNameError,
		ColorNameSuccessFill, ColorNameWarningFill, ColorNameErrorFill,
		theme.ColorNameBackground, theme.ColorNameForeground,
	} {
		assertColor(t, themeColor(t, accented, name, theme.VariantLight),
			themeColor(t, base, name, theme.VariantLight),
			"цвет %q не должен зависеть от акцента", name)
	}

	// Hover, pressed и selection следуют акценту — это оттенки акцентного цвета.
	for _, name := range []fyne.ThemeColorName{
		theme.ColorNameHover, theme.ColorNamePressed, theme.ColorNameSelection,
	} {
		if themeColor(t, accented, name, theme.VariantLight) == themeColor(t, base, name, theme.VariantLight) {
			t.Errorf("цвет %q должен следовать акценту", name)
		}
	}
}

// Fyne 2.7 не различает важность кнопок в токене: ColorNameButton обязан быть
// нейтральной поверхностью, а accent + onPrimary используются для акцентных
// кнопок (P0-5: «кнопки не кислотные»).
func TestTheme_Accent_ButtonAndOnPrimary(t *testing.T) {
	th := NewModernThemeAccent(false, "Purple")
	v := theme.VariantLight
	acc := accentPresets["Purple"].Light

	assertColor(t, themeColor(t, th, theme.ColorNameButton, v), tokLightSurface,
		"обычная кнопка — нейтральная поверхность")
	assertColor(t, themeColor(t, th, theme.ColorNameForeground, v), tokLightText,
		"текст обычной кнопки — foreground")
	assertColor(t, themeColor(t, th, theme.ColorNamePrimary, v), acc.Primary,
		"акцент — primary")
	assertColor(t, themeColor(t, th, ColorNameOnPrimary, v), acc.OnPrimary,
		"текст на акценте — onPrimary")

	thDark := NewModernThemeAccent(true, "Purple")
	accDark := accentPresets["Purple"].Dark
	assertColor(t, themeColor(t, thDark, theme.ColorNameButton, theme.VariantLight), tokDarkSurfaceAlt,
		"тёмная обычная кнопка — surfaceAlt")
	assertColor(t, themeColor(t, thDark, theme.ColorNamePrimary, theme.VariantLight), accDark.Primary,
		"тёмный акцент — primary")
	assertColor(t, themeColor(t, thDark, ColorNameOnPrimary, theme.VariantLight), accDark.OnPrimary,
		"тёмный текст на акценте — onPrimary")
}

func TestTheme_Accent_OnPrimaryIsContrasty(t *testing.T) {
	// onPrimary подбирается так, чтобы текст на кнопке был читаем (WCAG 3.0 для UI).
	for _, name := range accentPresetOrder {
		for _, dark := range []bool{false, true} {
			acc := accentPresets[name].Light
			if dark {
				acc = accentPresets[name].Dark
			}
			assertContrast(t, "onPrimary/primary "+name+" (dark="+strconv.FormatBool(dark)+")",
				acc.OnPrimary, acc.Primary, 3.0)
		}
	}
}

func TestTheme_Accent_SetAccentNormalizes(t *testing.T) {
	th := NewModernTheme(false)
	th.SetAccent("Пусто")
	if got := th.Accent(); got != defaultAccentPreset {
		t.Errorf("SetAccent must normalize: expected %q, got %q", defaultAccentPreset, got)
	}
	th.SetAccent("Teal")
	if got := th.Accent(); got != "Teal" {
		t.Errorf("expected Teal, got %q", got)
	}
	assertColor(t, themeColor(t, th, theme.ColorNamePrimary, theme.VariantLight),
		accentPresets["Teal"].Light.Primary, "primary после SetAccent")
}

func TestTheme_Accent_OrderCoversAllPresets(t *testing.T) {
	if len(accentPresetOrder) != len(accentPresets) {
		t.Fatalf("accentPresetOrder (%d) и accentPresets (%d) различаются", len(accentPresetOrder), len(accentPresets))
	}
	for _, name := range accentPresetOrder {
		if _, ok := accentPresets[name]; !ok {
			t.Errorf("пресет %q отсутствует в accentPresets", name)
		}
	}
}

// --- П0-1: контраст палитры (WCAG AA) ---------------------------------------

func TestTheme_Contrast_Light(t *testing.T) {
	assertContrast(t, "text/bg", tokLightText, tokLightBg, 4.5)
	assertContrast(t, "text/surface", tokLightText, tokLightSurface, 4.5)
	assertContrast(t, "textSecond/bg", tokLightTextSecond, tokLightBg, 4.5)
	assertContrast(t, "textSecond/surface", tokLightTextSecond, tokLightSurface, 4.5)
	assertContrast(t, "textDisabled/surfaceAlt", tokLightTextDisabled, tokLightSurfaceAlt, 4.5)
	assertContrast(t, "textOnSelection", tokLightText, tokLightSelection, 4.5)
	assertContrast(t, "onPrimary/primary", tokLightOnPrimary, tokLightPrimary, 4.5)
	assertContrast(t, "borderStrong/surface", tokLightBorderStrong, tokLightSurface, 3.0)
	assertContrast(t, "success/bg", tokLightSuccess, tokLightBg, 3.0)
	assertContrast(t, "success/surface", tokLightSuccess, tokLightSurface, 3.0)
	assertContrast(t, "warning/bg", tokLightWarning, tokLightBg, 3.0)
	assertContrast(t, "error/bg", tokLightError, tokLightBg, 3.0)
	assertContrast(t, "successFill/bg", tokLightSuccessFill, tokLightBg, 3.0)
	assertContrast(t, "errorFill/bg", tokLightErrorFill, tokLightBg, 3.0)
}

func TestTheme_Contrast_Dark(t *testing.T) {
	assertContrast(t, "text/bg", tokDarkText, tokDarkBg, 4.5)
	assertContrast(t, "text/surface", tokDarkText, tokDarkSurface, 4.5)
	assertContrast(t, "textSecond/bg", tokDarkTextSecond, tokDarkBg, 4.5)
	assertContrast(t, "textSecond/surface", tokDarkTextSecond, tokDarkSurface, 4.5)
	assertContrast(t, "textDisabled/surfaceAlt", tokDarkTextDisabled, tokDarkSurfaceAlt, 4.5)
	assertContrast(t, "textOnSelection", tokDarkText, tokDarkSelection, 4.5)
	assertContrast(t, "onPrimary/primary", tokDarkOnPrimary, tokDarkPrimary, 4.5)
	assertContrast(t, "borderStrong/surface", tokDarkBorderStrong, tokDarkSurface, 3.0)
	assertContrast(t, "success/bg", tokDarkSuccess, tokDarkBg, 3.0)
	assertContrast(t, "warning/bg", tokDarkWarning, tokDarkBg, 3.0)
	assertContrast(t, "error/bg", tokDarkError, tokDarkBg, 3.0)
}

func TestTheme_Contrast_HelpTextOnBothVariants(t *testing.T) {
	// Подсказки/подписи (SizeNameCaptionText) обязаны быть читаемы в обоих вариантах.
	assertContrast(t, "light caption text on bg", tokLightTextSecond, tokLightBg, 4.5)
	assertContrast(t, "dark caption text on bg", tokDarkTextSecond, tokDarkBg, 4.5)
}

// --- Sizes / Icon ------------------------------------------------------------

func TestTheme_Sizes(t *testing.T) {
	th := NewModernTheme(false)
	cases := map[fyne.ThemeSizeName]float32{
		theme.SizeNameText:            14,
		theme.SizeNameCaptionText:     11,
		theme.SizeNameHeadingText:     24,
		theme.SizeNameSubHeadingText:  18,
		theme.SizeNamePadding:         12,
		theme.SizeNameInlineIcon:      20,
		theme.SizeNameScrollBar:       16,
		theme.SizeNameInputRadius:     8,
		theme.SizeNameSelectionRadius: 8,
	}
	for name, want := range cases {
		if got := th.Size(name); got != want {
			t.Errorf("Size(%q): expected %v, got %v", name, want, got)
		}
	}
}

func TestTheme_Sizes_AreAccessible(t *testing.T) {
	th := NewModernTheme(false)
	if got := th.Size(theme.SizeNameText); got < 13 {
		t.Errorf("базовый шрифт %vpt слишком мелкий для desktop (нужно >= 13)", got)
	}
	if got := th.Size(theme.SizeNameCaptionText); got < 11 {
		t.Errorf("caption %vpt ниже минимума 11pt", got)
	}
	if got := th.Size(theme.SizeNameInputRadius); got > 8 {
		t.Errorf("радиус инпутов %v > 8 (перетемизация)", got)
	}
}

func TestTheme_Icon(t *testing.T) {
	th := NewModernTheme(false)
	if res := th.Icon(theme.IconNameCancel); res == nil {
		t.Fatal("тема не вернула иконку")
	}
}

func TestTheme_Color_UnknownNameFallsBack(t *testing.T) {
	th := NewModernTheme(false)
	if got := th.Color(fyne.ThemeColorName("unknown-color"), theme.VariantLight); got == nil {
		t.Fatal("unknown color name must return a fallback, not nil")
	}
}

// --- AutoOnColor -------------------------------------------------------------

func TestTheme_AutoOnColor(t *testing.T) {
	if got := AutoOnColor(color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); got != tokDarkOnPrimary {
		t.Errorf("на светлом фоне ожидается тёмный текст, got %+v", got)
	}
	if got := AutoOnColor(color.RGBA{R: 0x10, G: 0x10, B: 0x14, A: 0xFF}); got != tokLightOnPrimary {
		t.Errorf("на тёмном фоне ожидается светлый текст, got %+v", got)
	}
}
