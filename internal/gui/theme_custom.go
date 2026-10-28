package gui

// theme_custom.go — кастомные цвета темы (L1.2/L1.3 из ROADMAP).
//
// Поддержка произвольного акцентного цвета: пользователь вводит #RRGGBB,
// из базового цвета выводятся все акцентные роли (primary/pressed/hover/
// selection/onPrimary), цвет корректируется под вариант темы для читаемости,
// значение сохраняется в Preferences Fyne.

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// prefCustomAccent — ключ Preferences с hex-кодом кастомного акцента.
const prefCustomAccent = "gui.theme.accent.custom_hex"

// customAccentName — внутреннее имя «пресета» для произвольного цвета.
// Отображается в меню как отдельная подпись, в Preferences хранится как есть.
const customAccentName = "Custom"

// isValidHexColor проверяет формат #RRGGBB / RRGGBB (6 hex-символов).
func isValidHexColor(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	hex := s
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return false
	}
	for _, c := range hex {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}

// parseHexColor разбирает #RRGGBB / RRGGBB в color.RGBA (A=0xFF).
func parseHexColor(s string) color.RGBA {
	c, _ := parseHexColorStrict(s)
	return c
}

// parseHexColorStrict — разбор с явной ошибкой формата.
func parseHexColorStrict(s string) (color.RGBA, error) {
	if !isValidHexColor(s) {
		return color.RGBA{}, fmt.Errorf("некорректный цвет %q: ожидается #RRGGBB", s)
	}
	hex := strings.TrimPrefix(strings.TrimSpace(s), "#")
	r, errR := strconv.ParseUint(hex[0:2], 16, 8)
	g, errG := strconv.ParseUint(hex[2:4], 16, 8)
	b, errB := strconv.ParseUint(hex[4:6], 16, 8)
	if errR != nil || errG != nil || errB != nil {
		return color.RGBA{}, fmt.Errorf("некорректный цвет %q", s)
	}
	return color.RGBA{R: byte(r), G: byte(g), B: byte(b), A: 0xFF}, nil
}

// normalizeHex — канонический вид hex-строки: верхний регистр с #.
func normalizeHex(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s != "" && s[0] != '#' {
		s = "#" + s
	}
	return s
}

// srgbLuminance — относительная яркость (ITU-R BT.709, sRGB gamma-значения).
func srgbLuminance(c color.RGBA) float64 {
	r := float64(c.R) / 0xFF
	g := float64(c.G) / 0xFF
	b := float64(c.B) / 0xFF
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// blendTo — линейное смешивание c с целевым цветом to на долю t∈[0,1].
func blendTo(c color.RGBA, to uint8, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	mix := func(v uint8) uint8 {
		return uint8(float64(v)*(1-t) + float64(to)*t + 0.5)
	}
	return color.RGBA{R: mix(c.R), G: mix(c.G), B: mix(c.B), A: c.A}
}

// adjustAccentForVariant подгоняет яркость акцента под вариант темы:
// в тёмном варианте слишком тёмные цвета осветляются, в светлом слишком
// светлые — затемняются. Гарантирует различимость primary-подложки на фоне.
func adjustAccentForVariant(c color.RGBA, dark bool) color.RGBA {
	lum := srgbLuminance(c)
	const (
		minLumDark  = 0.35 // акцент на тёмном фоне не темнее
		maxLumLight = 0.65 // акцент на светлом фоне не светлее
	)
	if dark && lum < minLumDark {
		return blendTo(c, 0xFF, (minLumDark-lum)/minLumDark)
	}
	if !dark && lum > maxLumLight {
		return blendTo(c, 0x00, (lum-maxLumLight)/(1-maxLumLight))
	}
	return c
}

// deriveAccentVariant строит полный набор акцентных ролей из базового цвета.
// Значения по структуре совпадают с пресетами accentPresets:
//   - Pressed   — на 20% темнее primary;
//   - Hover     — светлый tint (light) / тёмный tint (dark) primary;
//   - Selection — primary с альфой 0x40;
//   - OnPrimary — автоподбор контрастного текста (AutoOnColor).
func deriveAccentVariant(base color.RGBA, dark bool) accentVariant {
	primary := adjustAccentForVariant(base, dark)
	pressed := blendTo(primary, 0x00, 0.2)
	var hover color.RGBA
	if dark {
		// В тёмном варианте hover — приглушённый тёмный tint акцента.
		hover = blendTo(primary, 0x00, 0.75)
	} else {
		hover = blendTo(primary, 0xFF, 0.87)
	}
	selection := primary
	selection.A = 0x40
	return accentVariant{
		Primary:   primary,
		Pressed:   pressed,
		Hover:     hover,
		Selection: selection,
		OnPrimary: AutoOnColor(primary),
	}
}

// --- Интеграция с ModernTheme -------------------------------------------------

// SetCustomAccent задаёт произвольный акцентный цвет темы.
// Валидный hex переключает тему в режим "Custom"; пустая/некорректная строка
// очищает кастомный цвет и возвращает тему к умолчанию, если она была в Custom.
func (t *ModernTheme) SetCustomAccent(hex string) error {
	hex = strings.TrimSpace(hex)
	if hex == "" {
		t.customAccentHex = ""
		if t.accent == customAccentName {
			t.accent = defaultAccentPreset
		}
		return nil
	}
	if _, err := parseHexColorStrict(hex); err != nil {
		return err
	}
	t.customAccentHex = normalizeHex(hex)
	t.accent = customAccentName
	return nil
}

// CustomAccentHex возвращает сохранённый hex кастомного акцента ("" — нет).
func (t *ModernTheme) CustomAccentHex() string { return t.customAccentHex }

// IsCustomAccent сообщает, активен ли произвольный акцент.
func (t *ModernTheme) IsCustomAccent() bool { return t.accent == customAccentName }

// --- Preferences (L1.3: сохранение настроек темы) -----------------------------

// saveCustomAccentHex сохраняет hex кастомного акцента.
func (a *App) saveCustomAccentHex(hex string) {
	if a == nil || a.myApp == nil {
		return
	}
	a.myApp.Preferences().SetString(prefCustomAccent, normalizeHex(hex))
}

// loadCustomAccentHex читает hex кастомного акцента ("" — не задан).
func (a *App) loadCustomAccentHex() string {
	if a == nil || a.myApp == nil {
		return ""
	}
	raw := strings.TrimSpace(a.myApp.Preferences().String(prefCustomAccent))
	if raw == "" {
		return ""
	}
	if !isValidHexColor(raw) {
		return ""
	}
	return normalizeHex(raw)
}

// applyCustomAccentHex применяет кастомный акцент и сохраняет выбор (L1.2+L1.3).
// Пустой hex сбрасывает акцент к сохранённому пресету (или умолчанию).
func (a *App) applyCustomAccentHex(hex string) {
	if a == nil || a.myApp == nil {
		return
	}
	hex = strings.TrimSpace(hex)
	if hex == "" {
		a.saveCustomAccentHex("")
		a.applyAccentPreset(a.loadAccentPreset())
		return
	}
	if _, err := parseHexColorStrict(hex); err != nil {
		if a.statusLabel != nil {
			a.setStatusToastWithMarker("✗ ", "Некорректный цвет: ожидается #RRGGBB")
		}
		return
	}
	a.saveCustomAccentHex(hex)
	a.saveAccentPreset(customAccentName)
	a.applyTheme(a.loadTheme())
	a.applyThemeToUI()
	if a.statusLabel != nil {
		a.setStatus("Кастомный акцент: " + normalizeHex(hex))
	}
}
