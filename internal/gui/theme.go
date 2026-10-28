package gui

import (
	"image/color"
	"math"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// defaultAccentPreset — акцент, применяемый при отсутствии сохранённого
// значения настройки (тот же список, что и в меню «Акцент»).
const defaultAccentPreset = "macOS Blue"

// ModernTheme — тема приложения на основе токенов из theme_tokens.go.
//
// Вариант темы (светлый/тёмный) задаётся явно (isDark) либо определяется
// системой (followSystem): в этом случае Fyne передаёт актуальный вариант в
// Color(), и палитра переключается без пересоздания темы (P0-3).
type ModernTheme struct {
	isDark       bool
	followSystem bool
	accent       string
	density      Density
	// customAccentHex — произвольный акцентный цвет (#RRGGBB, L1.2);
	// пустой — акцент берётся из пресета accent.
	customAccentHex string
}

// themePalette — набор семантических цветов одного варианта темы.
type themePalette struct {
	bg, surface, surfaceAlt        color.RGBA
	text, textSecond, textDisabled color.RGBA
	borderSubtle, borderStrong     color.RGBA
	selection, scrollBar           color.RGBA
	shadow                         color.RGBA
	primary, onPrimary             color.RGBA
	pressed, hover                 color.RGBA
	success, successFill           color.RGBA
	warning, warningFill           color.RGBA
	danger, dangerFill             color.RGBA

	dark bool // вариант темы — нужен для выбора подложки обычных кнопок
}

var _ fyne.Theme = (*ModernTheme)(nil)

// NewModernTheme создает новую тему со светлым/тёмным вариантом по умолчанию.
func NewModernTheme(isDark bool) *ModernTheme {
	return NewModernThemeAccent(isDark, defaultAccentPreset)
}

// NewModernThemeAccent — тема с явным вариантом и акцентным пресетом (P0-5).
// Неизвестное имя пресета приводится к defaultAccentPreset.
func NewModernThemeAccent(isDark bool, accent string) *ModernTheme {
	return &ModernTheme{isDark: isDark, accent: normalizeAccentPreset(accent)}
}

// NewModernThemeForVariant — тема для варианта темы Fyne (P0-3): используется
// настройками и меню, где значение хранится как fyne.ThemeVariant.
func NewModernThemeForVariant(variant fyne.ThemeVariant, accent string) *ModernTheme {
	return NewModernThemeAccent(variant == theme.VariantDark, accent)
}

// NewSystemTheme — тема, следующая системной настройке светлого/тёмного режима.
func NewSystemTheme() *ModernTheme {
	return &ModernTheme{followSystem: true, accent: defaultAccentPreset}
}

// normalizeAccentPreset приводит произвольное имя к известному пресету.
// Специальное имя customAccentName ("Custom") означает произвольный цвет (L1.2)
// и проходит без нормализации.
func normalizeAccentPreset(name string) string {
	name = strings.TrimSpace(name)
	if name == customAccentName {
		return customAccentName
	}
	if _, ok := accentPresets[name]; ok {
		return name
	}
	return defaultAccentPreset
}

// IsDark сообщает, используется ли тёмный вариант темы.
func (t *ModernTheme) IsDark() bool { return t.darkFor(0) }

// Accent возвращает имя активного акцентного пресета.
func (t *ModernTheme) Accent() string { return normalizeAccentPreset(t.accent) }

// settingsThemeVariant — вариант темы, предпочтительный системой/настройками.
func settingsThemeVariant() fyne.ThemeVariant {
	if app := fyne.CurrentApp(); app != nil && app.Settings() != nil {
		return app.Settings().ThemeVariant()
	}
	return theme.VariantLight
}

// darkFor сообщает, тёмный ли вариант использует тема. В режиме followSystem
// запрашиваемый у темы вариант (variant) игнорируется: значение берётся из
// системной настройки, иначе Fyne передал бы светлый вариант по умолчанию и
// тёмная палитра никогда бы не активировалась (P0-3).
func (t *ModernTheme) darkFor(_ fyne.ThemeVariant) bool {
	if t.followSystem {
		return settingsThemeVariant() == theme.VariantDark
	}
	return t.isDark
}

// Color возвращает цвет для указанного имени
func (t *ModernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	dark := t.darkFor(variant)
	p := t.palette(dark)
	fallback := theme.VariantLight
	if dark {
		fallback = theme.VariantDark
	}
	switch name {
	case theme.ColorNameBackground:
		return p.bg
	case theme.ColorNameForeground:
		return p.text
	case theme.ColorNamePrimary, theme.ColorNameFocus, theme.ColorNameHyperlink:
		return p.primary
	case ColorNameOnPrimary:
		return p.onPrimary
	case theme.ColorNameButton:
		// Обычные кнопки — нейтральная поверхность: accent остаётся для
		// акцентных (HighImportance) кнопок (P0-5: «кнопки не кислотные»).
		if p.dark {
			return p.surfaceAlt
		}
		return p.surface
	case theme.ColorNameDisabledButton:
		return p.surfaceAlt
	case theme.ColorNameDisabled:
		return p.textDisabled
	case theme.ColorNameError:
		return p.danger
	case ColorNameErrorFill:
		return p.dangerFill
	case theme.ColorNameSuccess:
		return p.success
	case ColorNameSuccessFill:
		return p.successFill
	case theme.ColorNameWarning:
		return p.warning
	case ColorNameWarningFill:
		return p.warningFill
	case theme.ColorNameHover:
		return p.hover
	case theme.ColorNameInputBackground:
		return p.surface
	case theme.ColorNameInputBorder, ColorNameBorderStrong:
		return p.borderStrong
	case theme.ColorNameSeparator:
		return p.borderSubtle
	case ColorNameSurfaceAlt:
		return p.surfaceAlt
	case theme.ColorNamePlaceHolder:
		return p.textSecond
	case theme.ColorNamePressed:
		return p.pressed
	case theme.ColorNameSelection:
		return p.selection
	case theme.ColorNameScrollBar:
		return p.scrollBar
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return p.surface
	case theme.ColorNameShadow:
		return p.shadow
	default:
		return theme.DefaultTheme().Color(name, fallback)
	}
}

// palette собирает палитру варианта темы с активным акцентным пресетом.
func (t *ModernTheme) palette(dark bool) themePalette {
	accent := t.accentVariant(dark)
	if dark {
		return themePalette{
			bg: tokDarkBg, surface: tokDarkSurface, surfaceAlt: tokDarkSurfaceAlt,
			text: tokDarkText, textSecond: tokDarkTextSecond, textDisabled: tokDarkTextDisabled,
			borderSubtle: tokDarkBorderSubtle, borderStrong: tokDarkBorderStrong,
			selection: accent.Selection, scrollBar: tokScrollBarDk,
			shadow:  tokDarkShadow,
			primary: accent.Primary, onPrimary: accent.OnPrimary, pressed: accent.Pressed, hover: accent.Hover,
			success: tokDarkSuccess, successFill: tokDarkSuccessFill,
			warning: tokDarkWarning, warningFill: tokDarkWarningFill,
			danger: tokDarkError, dangerFill: tokDarkErrorFill,
			dark: true,
		}
	}
	return themePalette{
		bg: tokLightBg, surface: tokLightSurface, surfaceAlt: tokLightSurfaceAlt,
		text: tokLightText, textSecond: tokLightTextSecond, textDisabled: tokLightTextDisabled,
		borderSubtle: tokLightBorderSubtle, borderStrong: tokLightBorderStrong,
		selection: accent.Selection, scrollBar: tokScrollBarLt,
		shadow:  tokLightShadow,
		primary: accent.Primary, onPrimary: accent.OnPrimary, pressed: accent.Pressed, hover: accent.Hover,
		success: tokLightSuccess, successFill: tokLightSuccessFill,
		warning: tokLightWarning, warningFill: tokLightWarningFill,
		danger: tokLightError, dangerFill: tokLightErrorFill,
	}
}

// accentVariant — набор акцентных цветов пресета для варианта темы.
// Для произвольного цвета (L1.2) роли выводятся из hex-значения.
func (t *ModernTheme) accentVariant(dark bool) accentVariant {
	if t.accent == customAccentName && t.customAccentHex != "" {
		return deriveAccentVariant(parseHexColor(t.customAccentHex), dark)
	}
	entry, ok := accentPresets[normalizeAccentPreset(t.accent)]
	if !ok {
		entry = accentPresets[defaultAccentPreset]
	}
	if dark {
		return entry.Dark
	}
	return entry.Light
}

// SetAccent меняет акцентный пресет темы.
func (t *ModernTheme) SetAccent(name string) {
	t.accent = normalizeAccentPreset(name)
}

// Font возвращает шрифт для указанного стиля
func (t *ModernTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon возвращает иконку для указанного имени
func (t *ModernTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Density — уровень плотности интерфейса (T-02).
type Density int

const (
	DensityComfortable Density = iota // стандартная
	DensityCompact                    // компактная
)

// String возвращает человекочитаемое имя плотности.
func (d Density) String() string {
	if d == DensityCompact {
		return "compact"
	}
	return "comfortable"
}

// Size возвращает размер для указанного имени.
// T-02: значения масштабируются по плотности (compact ~ 0.85×, comfortable 1.0×).
func (t *ModernTheme) Size(name fyne.ThemeSizeName) float32 {
	compact := t.density == DensityCompact
	switch name {
	case theme.SizeNamePadding:
		if compact {
			return 8
		}
		return 12
	case theme.SizeNameScrollBar:
		if compact {
			return 12
		}
		return 16
	case theme.SizeNameScrollBarSmall:
		return 3
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return 8 // Скругленные углы
	case theme.SizeNameSelectionRadius:
		return 8
	case theme.SizeNameInlineIcon:
		if compact {
			return 16
		}
		return 20
	case theme.SizeNameInnerPadding:
		if compact {
			return 4
		}
		return 8
	case theme.SizeNameLineSpacing:
		if compact {
			return 2
		}
		return 4
	case theme.SizeNameText:
		if compact {
			return 12
		}
		return 14
	case theme.SizeNameHeadingText:
		if compact {
			return 20
		}
		return 24
	case theme.SizeNameSubHeadingText:
		if compact {
			return 15
		}
		return 18
	case theme.SizeNameCaptionText:
		if compact {
			return 10
		}
		return 11
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// SetDensity меняет плотность интерфейса и обновляет тему.
func (t *ModernTheme) SetDensity(d Density) {
	t.density = d
}

// CurrentDensity возвращает текущую плотность (T-02).
func (t *ModernTheme) CurrentDensity() Density {
	return t.density
}

// --- Контраст (WCAG 2.1) -----------------------------------------------------

// ContrastRatio возвращает коэффициент контраста двух цветов по формуле
// WCAG 2.1 (относительная яркость), от 1.0 до 21.0.
func ContrastRatio(fg, bg color.Color) float64 {
	l1 := relativeLuminance(fg)
	l2 := relativeLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// relativeLuminance — относительная яркость sRGB-цвета (ITU-R BT.709).
func relativeLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return 0.2126*srgbChannel(r) + 0.7152*srgbChannel(g) + 0.0722*srgbChannel(b)
}

// srgbChannel переводит 16-битную компоненту RGBA в линейную яркость.
func srgbChannel(v uint32) float64 {
	s := float64(v) / 0xFFFF
	if s <= 0.03928 {
		return s / 12.92
	}
	return math.Pow((s+0.055)/1.055, 2.4)
}

// AutoOnColor подбирает цвет текста (белый/тёмно-синий) с максимальным
// контрастом на фоне bg — используется для onPrimary над акцентной подложкой.
// Кандидаты — токены OnPrimary светлой и тёмной палитры.
func AutoOnColor(bg color.Color) color.RGBA {
	if ContrastRatio(onColorLight, bg) >= ContrastRatio(onColorDark, bg) {
		return onColorLight
	}
	return onColorDark
}

var (
	// onColorLight/onColorDark — кандидаты для текста поверх акцентной подложки.
	// Совпадают с токенами OnPrimary обеих палитр.
	onColorLight = tokLightOnPrimary // белый
	onColorDark  = tokDarkOnPrimary  // тёмно-синий #10233A
)
