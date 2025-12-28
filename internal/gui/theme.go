package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ModernTheme представляет современную тему в стиле macOS Sierra/Windows 11
type ModernTheme struct {
	isDark bool
}

// NewModernTheme создает новую тему
func NewModernTheme(isDark bool) *ModernTheme {
	return &ModernTheme{isDark: isDark}
}

// Color возвращает цвет для указанного имени
func (t *ModernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.isDark {
		return t.darkColor(name)
	}
	return t.lightColor(name)
}

// Variant возвращает вариант темы
func (t *ModernTheme) Variant() fyne.ThemeVariant {
	if t.isDark {
		return theme.VariantDark
	}
	return theme.VariantLight
}

// lightColor возвращает цвета для светлой темы (macOS Sierra стиль)
func (t *ModernTheme) lightColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		// Светло-серый фон macOS Sierra
		return color.RGBA{R: 0xF5, G: 0xF5, B: 0xF7, A: 0xFF}
	case theme.ColorNameButton:
		// Синий цвет кнопок macOS Sierra
		return color.RGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF}
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 0xC7, G: 0xC7, B: 0xCC, A: 0xFF}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 0xC7, G: 0xC7, B: 0xCC, A: 0xFF}
	case theme.ColorNameError:
		return color.RGBA{R: 0xFF, G: 0x3B, B: 0x30, A: 0xFF}
	case theme.ColorNameFocus:
		// Синий цвет фокуса macOS
		return color.RGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF}
	case theme.ColorNameForeground:
		// Темный текст
		return color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
	case theme.ColorNameHover:
		// Светло-синий при наведении
		return color.RGBA{R: 0xE5, G: 0xF2, B: 0xFF, A: 0xFF}
	case theme.ColorNameInputBackground:
		// Белый фон для полей ввода
		return color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	case theme.ColorNameInputBorder:
		// Серый бордер для полей ввода
		return color.RGBA{R: 0xC7, G: 0xC7, B: 0xCC, A: 0xFF}
	case theme.ColorNameMenuBackground:
		return color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x80}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 0x8E, G: 0x8E, B: 0x93, A: 0xFF}
	case theme.ColorNamePressed:
		// Темно-синий при нажатии
		return color.RGBA{R: 0x00, G: 0x5A, B: 0xCC, A: 0xFF}
	case theme.ColorNamePrimary:
		// Основной синий цвет macOS
		return color.RGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF}
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 0xC7, G: 0xC7, B: 0xCC, A: 0x80}
	case theme.ColorNameSelection:
		// Цвет выделения macOS
		return color.RGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0x40}
	case theme.ColorNameSeparator:
		return color.RGBA{R: 0xC7, G: 0xC7, B: 0xCC, A: 0xFF}
	case theme.ColorNameShadow:
		return color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x20}
	case theme.ColorNameSuccess:
		return color.RGBA{R: 0x34, G: 0xC7, B: 0x59, A: 0xFF}
	case theme.ColorNameWarning:
		return color.RGBA{R: 0xFF, G: 0x95, B: 0x00, A: 0xFF}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

// darkColor возвращает цвета для темной темы (Windows 11 стиль)
func (t *ModernTheme) darkColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		// Темный фон Windows 11
		return color.RGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xFF}
	case theme.ColorNameButton:
		// Синий цвет кнопок Windows 11
		return color.RGBA{R: 0x00, G: 0x78, B: 0xD4, A: 0xFF}
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 0x3D, G: 0x3D, B: 0x3D, A: 0xFF}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 0x6D, G: 0x6D, B: 0x6D, A: 0xFF}
	case theme.ColorNameError:
		return color.RGBA{R: 0xF1, G: 0x70, B: 0x70, A: 0xFF}
	case theme.ColorNameFocus:
		return color.RGBA{R: 0x00, G: 0x78, B: 0xD4, A: 0xFF}
	case theme.ColorNameForeground:
		// Светлый текст
		return color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	case theme.ColorNameHover:
		return color.RGBA{R: 0x2D, G: 0x2D, B: 0x30, A: 0xFF}
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 0xFF}
	case theme.ColorNameInputBorder:
		return color.RGBA{R: 0x3D, G: 0x3D, B: 0x3D, A: 0xFF}
	case theme.ColorNameMenuBackground:
		return color.RGBA{R: 0x2D, G: 0x2D, B: 0x30, A: 0xFF}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x80}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 0x6D, G: 0x6D, B: 0x6D, A: 0xFF}
	case theme.ColorNamePressed:
		return color.RGBA{R: 0x00, G: 0x5A, B: 0xA0, A: 0xFF}
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x00, G: 0x78, B: 0xD4, A: 0xFF}
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 0x3D, G: 0x3D, B: 0x3D, A: 0x80}
	case theme.ColorNameSelection:
		return color.RGBA{R: 0x00, G: 0x78, B: 0xD4, A: 0x40}
	case theme.ColorNameSeparator:
		return color.RGBA{R: 0x3D, G: 0x3D, B: 0x3D, A: 0xFF}
	case theme.ColorNameShadow:
		return color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x40}
	case theme.ColorNameSuccess:
		return color.RGBA{R: 0x6F, G: 0xCF, B: 0x97, A: 0xFF}
	case theme.ColorNameWarning:
		return color.RGBA{R: 0xFF, G: 0xB3, B: 0x66, A: 0xFF}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

// Font возвращает шрифт для указанного стиля
func (t *ModernTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon возвращает иконку для указанного имени
func (t *ModernTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size возвращает размер для указанного имени
func (t *ModernTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 12
	case theme.SizeNameScrollBar:
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
		return 20
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameLineSpacing:
		return 4
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameCaptionText:
		return 11
	default:
		return theme.DefaultTheme().Size(name)
	}
}

