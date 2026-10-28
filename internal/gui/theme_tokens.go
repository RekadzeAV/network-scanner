package gui

// theme_tokens.go — единый источник цветовых значений GUI
// (docs/GUI_DESIGN_RECOMMENDATIONS.md, P0-1/P0-5, таблица палитры §5.2).
//
// Правила:
//   - color.RGBA{...} вне этого файла запрещён (grep-чек из §13);
//   - семантические токены переопределены в ОБЕИХ вариантах темы;
//   - текстовые статусные цвета проходят WCAG AA 4.5:1 на bg и surface
//     своего варианта; fill-цвета (графика/подложки) — 3.0:1 (1.4.11).

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
)

// rgba — конструктор токенов: все значения палитры проходят только через него.
func rgba(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}

// --- Светлый вариант ---------------------------------------------------------

var (
	tokLightBg           = rgba(0xF5, 0xF5, 0xF7, 0xFF) // фон окна
	tokLightSurface      = rgba(0xFF, 0xFF, 0xFF, 0xFF) // поверхности/инпуты
	tokLightSurfaceAlt   = rgba(0xEF, 0xEF, 0xF1, 0xFF) // чипы/зебра (текст 14.66:1)
	tokLightText         = rgba(0x1D, 0x1D, 0x1F, 0xFF) // 16.83:1 на surface
	tokLightTextSecond   = rgba(0x55, 0x55, 0x5A, 0xFF) // 7.41:1 на surface
	tokLightTextDisabled = rgba(0x66, 0x66, 0x6B, 0xFF) // 4.95:1 на surfaceAlt, 5.24:1 на bg
	tokLightBorderSubtle = rgba(0xD5, 0xD5, 0xDA, 0xFF) // декоративные разделители
	tokLightBorderStrong = rgba(0x86, 0x86, 0x8E, 0xFF) // рамки контролов 3.61:1
	tokLightSelection    = rgba(0xCE, 0xDF, 0xF5, 0xFF) // выделение (текст 12.42:1)

	tokLightPrimary     = rgba(0x0A, 0x5F, 0xCC, 0xFF) // белый текст 5.96:1
	tokLightOnPrimary   = rgba(0xFF, 0xFF, 0xFF, 0xFF)
	tokLightSuccess     = rgba(0x14, 0x76, 0x2F, 0xFF) // текст 5.73:1 на surface
	tokLightSuccessFill = rgba(0x1E, 0x8E, 0x3E, 0xFF)
	tokLightWarning     = rgba(0x7A, 0x53, 0x00, 0xFF) // текст 6.85:1 на surface
	tokLightWarningFill = rgba(0xB2, 0x6A, 0x00, 0xFF)
	tokLightError       = rgba(0xB3, 0x26, 0x1E, 0xFF) // текст 6.54:1 на surface
	tokLightErrorFill   = rgba(0xD1, 0x34, 0x2E, 0xFF)
)

// --- Тёмный вариант ----------------------------------------------------------

var (
	tokDarkBg           = rgba(0x1B, 0x1B, 0x1F, 0xFF)
	tokDarkSurface      = rgba(0x23, 0x23, 0x28, 0xFF)
	tokDarkSurfaceAlt   = rgba(0x2A, 0x2A, 0x30, 0xFF) // текст 11.77:1
	tokDarkText         = rgba(0xE9, 0xE9, 0xEC, 0xFF) // 14.17:1 на bg
	tokDarkTextSecond   = rgba(0xB3, 0xB3, 0xB9, 0xFF) // 8.23:1
	tokDarkTextDisabled = rgba(0x96, 0x96, 0x9C, 0xFF) // 4.88:1 на surfaceAlt
	tokDarkBorderSubtle = rgba(0x3C, 0x3C, 0x43, 0xFF)
	tokDarkBorderStrong = rgba(0x6E, 0x6E, 0x78, 0xFF) // 3.40:1 на bg
	tokDarkSelection    = rgba(0x2E, 0x3D, 0x55, 0xFF) // текст 9.05:1

	tokDarkPrimary     = rgba(0x6B, 0xA8, 0xFF, 0xFF) // 7.09:1 на bg
	tokDarkOnPrimary   = rgba(0x10, 0x23, 0x3A, 0xFF) // на primary-fill 6.55:1
	tokDarkSuccess     = rgba(0x7B, 0xD3, 0x94, 0xFF) // 9.47:1
	tokDarkSuccessFill = rgba(0x2E, 0x7D, 0x46, 0xFF)
	tokDarkWarning     = rgba(0xE3, 0xB3, 0x41, 0xFF) // 8.82:1
	tokDarkWarningFill = rgba(0x8A, 0x6A, 0x17, 0xFF)
	tokDarkError       = rgba(0xF2, 0x8B, 0x82, 0xFF) // 7.19:1
	tokDarkErrorFill   = rgba(0xB3, 0x45, 0x3C, 0xFF)
)

// --- Прозрачные служебные токены ---------------------------------------------

var (
	tokLightShadow = rgba(0x00, 0x00, 0x00, 0x20)
	tokDarkShadow  = rgba(0x00, 0x00, 0x00, 0x40)
	tokScrollBarLt = rgba(0x86, 0x86, 0x8E, 0x80)
	tokScrollBarDk = rgba(0x6E, 0x6E, 0x78, 0x80)
)

// --- Иконка приложения (процедурный рендер, image-байты не темизируются) -----

// paletteBorderStrong — рамка контролов/фокус варианта темы (AA non-text).
func paletteBorderStrong(dark bool) color.RGBA {
	if dark {
		return tokDarkBorderStrong
	}
	return tokLightBorderStrong
}

// --- Состояния портов (чипы результатов, R-05) --------------------------------

// portStateStyle — стиль чипа порта: текст/подложка варианта темы и маркер
// формы. Состояние различается НЕ только цветом (WCAG 1.4.1): open — "●",
// filtered — "◌", unknown — "—", closed — без маркера. Значения — таблица
// «Семантика состояний портов» §5.2 GUI_DESIGN_RECOMMENDATIONS.md; контрасты
// текста на подложке не ниже AA (проверяются тестом).
type portStateStyle struct {
	Text   color.RGBA
	Fill   color.RGBA
	Marker string
}

var (
	portStateStylesLight = map[string]portStateStyle{
		"open":     {Text: rgba(0x14, 0x76, 0x2F, 0xFF), Fill: rgba(0xDC, 0xF0, 0xE2, 0xFF), Marker: "●"},
		"closed":   {Text: rgba(0x55, 0x55, 0x5A, 0xFF), Fill: rgba(0xEF, 0xEF, 0xF1, 0xFF), Marker: ""},
		"filtered": {Text: rgba(0x7A, 0x53, 0x00, 0xFF), Fill: rgba(0xF6, 0xE9, 0xC6, 0xFF), Marker: "◌"},
		"unknown":  {Text: rgba(0x6E, 0x6E, 0x73, 0xFF), Fill: rgba(0xEF, 0xEF, 0xF1, 0xFF), Marker: "—"},
	}

	portStateStylesDark = map[string]portStateStyle{
		"open":     {Text: rgba(0x7B, 0xD3, 0x94, 0xFF), Fill: rgba(0x24, 0x3D, 0x2C, 0xFF), Marker: "●"},
		"closed":   {Text: rgba(0xB3, 0xB3, 0xB9, 0xFF), Fill: rgba(0x2A, 0x2A, 0x30, 0xFF), Marker: ""},
		"filtered": {Text: rgba(0xE3, 0xB3, 0x41, 0xFF), Fill: rgba(0x3A, 0x32, 0x20, 0xFF), Marker: "◌"},
		"unknown":  {Text: rgba(0x8A, 0x8A, 0x90, 0xFF), Fill: rgba(0x2A, 0x2A, 0x30, 0xFF), Marker: "—"},
	}
)

// portStateKey — каноническое состояние порта (scanner.PortInfo.State допускает
// разный регистр/пробелы; всё неизвестное — "unknown", не "open").
func portStateKey(state string) string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "open":
		return "open"
	case "closed":
		return "closed"
	case "filtered":
		return "filtered"
	default:
		return "unknown"
	}
}

// portStateStyleFor — стиль чипа состояния в активном варианте темы.
func portStateStyleFor(state string) portStateStyle {
	styles := portStateStylesLight
	if currentThemeIsDark() {
		styles = portStateStylesDark
	}
	return styles[portStateKey(state)]
}

// paletteSurface — фон карточек/таблиц/подложек диаграмм варианта темы.
func paletteSurface(dark bool) color.RGBA {
	if dark {
		return tokDarkSurface
	}
	return tokLightSurface
}

// --- Топология (canvas-отрисовка графа) ---------------------------------------

// Узлы и связи графа — нетекстовые элементы: по WCAG 1.4.11 нужен контраст
// >=3:1 с фоном, поэтому «Material»-оттенки затемнены (жёлтый #FBBC04 на белом
// давал 1.7:1).
var (
	tokTopoConfHigh   = rgba(0x1B, 0x7F, 0x3B, 0xFF) // уверенность high
	tokTopoConfMedium = rgba(0x9A, 0x64, 0x00, 0xFF) // уверенность medium
	tokTopoConfLow    = rgba(0xB3, 0x26, 0x1E, 0xFF) // уверенность low

	tokTopoNodeRouter = rgba(0x1A, 0x5F, 0xC2, 0xDC) // узел «router»
	tokTopoNodeSwitch = rgba(0x1B, 0x7F, 0x3B, 0xDC) // узел «switch»
	tokTopoNodeHost   = rgba(0x9A, 0x64, 0x00, 0xDC) // узел «host»
	tokTopoNodeOther  = rgba(0x6E, 0x6E, 0x6E, 0xDC) // узел без типа

	tokTopoNodeBorder       = rgba(0x3C, 0x3C, 0x3C, 0xB4) // обводка узла
	tokTopoNodeBorderSelect = rgba(0x14, 0x14, 0x14, 0xFF) // обводка выбранного узла
	tokTopoLinkHighlight    = rgba(0xD2, 0xE6, 0xFF, 0xAA) // подсветка выбранной связи
)

var (
	tokAppIconBackground = rgba(0x0A, 0x5F, 0xCC, 0xFF)
	tokAppIconForeground = rgba(0xFF, 0xFF, 0xFF, 0xFF)
)

// --- Палитра серий диаграмм ----------------------------------------------------

// pieSeries — серии pie/графиков; все оттенки читаемы как графика (>=3:1)
// на обоих фонах (самый слабый #475569: 4.42:1 light / 3.69:1 dark).
var pieSeries = []color.RGBA{
	rgba(0x25, 0x63, 0xEB, 0xFF),
	rgba(0x3B, 0x82, 0xF6, 0xFF),
	rgba(0x0E, 0xA5, 0xE9, 0xFF),
	rgba(0x39, 0x70, 0xA8, 0xFF),
	rgba(0xA8, 0x55, 0xF7, 0xFF),
	rgba(0x47, 0x55, 0x69, 0xFF),
}

// --- Семантические имена цветов за пределами стандартного интерфейса fyne ----

const (
	// ColorNameSurfaceAlt — фон чипов/зебры (elevation 2).
	ColorNameSurfaceAlt = fyne.ThemeColorName("networkSurfaceAlt")
	// ColorNameBorderStrong — рамки контролов и фокус-обводка.
	ColorNameBorderStrong = fyne.ThemeColorName("networkBorderStrong")
	// ColorNameOnPrimary — текст поверх primary-fill (кнопки).
	ColorNameOnPrimary = fyne.ThemeColorName("networkOnPrimary")
	// ColorNameSuccessFill / WarningFill / ErrorFill — графические варианты статусов.
	ColorNameSuccessFill = fyne.ThemeColorName("networkSuccessFill")
	ColorNameWarningFill = fyne.ThemeColorName("networkWarningFill")
	ColorNameErrorFill   = fyne.ThemeColorName("networkErrorFill")
)

// accentPresetEntry — палитра пресета акцента для обоих вариантов темы.
type accentPresetEntry struct {
	Light accentVariant
	Dark  accentVariant
}

type accentVariant struct {
	Primary   color.RGBA
	Pressed   color.RGBA
	Hover     color.RGBA
	Selection color.RGBA
	OnPrimary color.RGBA // автоподбор контрастного текста (P0-5 риски)
}

// accentPresetOrder — детерминированный порядок пунктов меню «Акцент» (C-11).
var accentPresetOrder = []string{
	"macOS Blue", "Windows 11", "Green", "Purple", "Orange", "Red", "Teal",
}

// accentPresets — пересчитаны под AA: белый/тёмный текст на кнопке подбирается
// по luminance (onPrimary), hover-подложки различимы на своём фоне (>=1.15:1)
// и не «слепят» на чужом.
var accentPresets = map[string]accentPresetEntry{
	"macOS Blue": {
		Light: accentVariant{Primary: rgba(0x0A, 0x5F, 0xCC, 0xFF), Pressed: rgba(0x08, 0x4C, 0xA3, 0xFF), Hover: rgba(0xE1, 0xEC, 0xFA, 0xFF), Selection: rgba(0x0A, 0x5F, 0xCC, 0x40), OnPrimary: rgba(0xFF, 0xFF, 0xFF, 0xFF)},
		Dark:  accentVariant{Primary: rgba(0x6B, 0xA8, 0xFF, 0xFF), Pressed: rgba(0x4E, 0x88, 0xE0, 0xFF), Hover: rgba(0x2A, 0x35, 0x46, 0xFF), Selection: rgba(0x6B, 0xA8, 0xFF, 0x40), OnPrimary: rgba(0x10, 0x23, 0x3A, 0xFF)},
	},
	"Windows 11": {
		Light: accentVariant{Primary: rgba(0x00, 0x5A, 0xA8, 0xFF), Pressed: rgba(0x00, 0x47, 0x88, 0xFF), Hover: rgba(0xDE, 0xEC, 0xF6, 0xFF), Selection: rgba(0x00, 0x5A, 0xA8, 0x40), OnPrimary: rgba(0xFF, 0xFF, 0xFF, 0xFF)},
		Dark:  accentVariant{Primary: rgba(0x4C, 0xC2, 0xFF, 0xFF), Pressed: rgba(0x2E, 0xA4, 0xE4, 0xFF), Hover: rgba(0x24, 0x34, 0x40, 0xFF), Selection: rgba(0x4C, 0xC2, 0xFF, 0x40), OnPrimary: rgba(0x0B, 0x1A, 0x24, 0xFF)},
	},
	"Green": {
		Light: accentVariant{Primary: rgba(0x14, 0x76, 0x2F, 0xFF), Pressed: rgba(0x0E, 0x5C, 0x24, 0xFF), Hover: rgba(0xDD, 0xF0, 0xE2, 0xFF), Selection: rgba(0x14, 0x76, 0x2F, 0x40), OnPrimary: rgba(0xFF, 0xFF, 0xFF, 0xFF)},
		Dark:  accentVariant{Primary: rgba(0x4F, 0xC0, 0x74, 0xFF), Pressed: rgba(0x38, 0xA4, 0x5C, 0xFF), Hover: rgba(0x24, 0x3D, 0x2C, 0xFF), Selection: rgba(0x4F, 0xC0, 0x74, 0x40), OnPrimary: rgba(0x0C, 0x24, 0x14, 0xFF)},
	},
	"Purple": {
		Light: accentVariant{Primary: rgba(0x6D, 0x28, 0xD9, 0xFF), Pressed: rgba(0x55, 0x1E, 0xB0, 0xFF), Hover: rgba(0xEA, 0xE2, 0xFB, 0xFF), Selection: rgba(0x6D, 0x28, 0xD9, 0x40), OnPrimary: rgba(0xFF, 0xFF, 0xFF, 0xFF)},
		Dark:  accentVariant{Primary: rgba(0xB1, 0x8A, 0xFF, 0xFF), Pressed: rgba(0x93, 0x6B, 0xE6, 0xFF), Hover: rgba(0x32, 0x2A, 0x45, 0xFF), Selection: rgba(0xB1, 0x8A, 0xFF, 0x40), OnPrimary: rgba(0x1C, 0x10, 0x30, 0xFF)},
	},
	"Orange": {
		Light: accentVariant{Primary: rgba(0xB2, 0x6A, 0x00, 0xFF), Pressed: rgba(0x8C, 0x53, 0x00, 0xFF), Hover: rgba(0xF8, 0xEA, 0xD2, 0xFF), Selection: rgba(0xB2, 0x6A, 0x00, 0x40), OnPrimary: rgba(0xFF, 0xFF, 0xFF, 0xFF)},
		Dark:  accentVariant{Primary: rgba(0xFF, 0xA5, 0x4D, 0xFF), Pressed: rgba(0xE0, 0x8A, 0x2E, 0xFF), Hover: rgba(0x3A, 0x32, 0x20, 0xFF), Selection: rgba(0xFF, 0xA5, 0x4D, 0x40), OnPrimary: rgba(0x2A, 0x18, 0x00, 0xFF)},
	},
	"Red": {
		Light: accentVariant{Primary: rgba(0xB3, 0x26, 0x1E, 0xFF), Pressed: rgba(0x8E, 0x1D, 0x17, 0xFF), Hover: rgba(0xF9, 0xE1, 0xDF, 0xFF), Selection: rgba(0xB3, 0x26, 0x1E, 0x40), OnPrimary: rgba(0xFF, 0xFF, 0xFF, 0xFF)},
		Dark:  accentVariant{Primary: rgba(0xF2, 0x8B, 0x82, 0xFF), Pressed: rgba(0xD1, 0x6A, 0x62, 0xFF), Hover: rgba(0x40, 0x2A, 0x29, 0xFF), Selection: rgba(0xF2, 0x8B, 0x82, 0x40), OnPrimary: rgba(0x2E, 0x0C, 0x0A, 0xFF)},
	},
	"Teal": {
		Light: accentVariant{Primary: rgba(0x00, 0x69, 0x5F, 0xFF), Pressed: rgba(0x00, 0x51, 0x49, 0xFF), Hover: rgba(0xD9, 0xEE, 0xEB, 0xFF), Selection: rgba(0x00, 0x69, 0x5F, 0x40), OnPrimary: rgba(0xFF, 0xFF, 0xFF, 0xFF)},
		Dark:  accentVariant{Primary: rgba(0x4F, 0xC6, 0xBB, 0xFF), Pressed: rgba(0x33, 0xA8, 0x9E, 0xFF), Hover: rgba(0x20, 0x38, 0x37, 0xFF), Selection: rgba(0x4F, 0xC6, 0xBB, 0x40), OnPrimary: rgba(0x07, 0x24, 0x22, 0xFF)},
	},
}
