package gui

import (
	"fmt"
	"strings"

	"network-scanner/internal/network"
	"network-scanner/internal/scanner"
)

// nullDash возвращает "-" для пустых строк.
func nullDash(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	return s
}

// formatIPWithProtocol форматирует IP-адрес с IPv6-индикатором.
func formatIPWithProtocol(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return "-"
	}
	if network.IsIPv6(ip) {
		return "[" + ip + "] (v6)"
	}
	return ip
}

// deviceCategoryGlyph возвращает визуальный маркер категории устройства
// (R-08: вместо ASCII-префиксов [NET]/[SRV]/[UNK] — различимый по форме символ).
func deviceCategoryGlyph(deviceType string) string {
	low := strings.ToLower(strings.TrimSpace(deviceType))
	switch {
	case strings.Contains(low, "router"), strings.Contains(low, "switch"):
		return "⬡"
	case strings.Contains(low, "access point"):
		return "◎"
	case strings.Contains(low, "printer"):
		return "⎙"
	case strings.Contains(low, "camera"):
		return "◉"
	case strings.Contains(low, "nas"):
		return "▤"
	case strings.Contains(low, "iot"):
		return "◈"
	case strings.Contains(low, "desktop"), strings.Contains(low, "laptop"):
		return "▢"
	case strings.Contains(low, "server"):
		return "▣"
	case strings.Contains(low, "phone"), strings.Contains(low, "tablet"):
		return "▮"
	default:
		return "◌"
	}
}

// deviceTypeWithBadge добавляет визуальный маркер категории к типу устройства.
// R-08: маркер различим по форме (не только по цвету) и читается как категория,
// а не как служебный ASCII-тег. Пустой тип даёт прочерк без маркера.
func deviceTypeWithBadge(deviceType string) string {
	dt := strings.TrimSpace(deviceType)
	if dt == "" {
		return "-"
	}
	return deviceCategoryGlyph(dt) + " " + dt
}

// osGuessLine формирует строку с информацией об ОС.
// R-07: усечение rune-safe до 60 символов, чтобы не переполнять колонку фикс. ширины.
func osGuessLine(r scanner.Result) string {
	if strings.TrimSpace(r.GuessOS) != "" {
		label := strings.TrimSpace(r.GuessOS)
		if strings.TrimSpace(r.GuessOSConfidence) != "" {
			label = fmt.Sprintf("%s (%s)", label, strings.TrimSpace(r.GuessOSConfidence))
		}
		if strings.TrimSpace(r.GuessOSReason) != "" {
			label += " — " + strings.TrimSpace(r.GuessOSReason)
		}
		return truncateString(label, 60)
	}
	return "-"
}
