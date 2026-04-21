package osdetect

import "strings"

// GuessFromHostAndPorts эвристика по имени хоста и номерам открытых TCP-портов.
func GuessFromHostAndPorts(hostname string, openPorts []int) (osName, confidence string) {
	h := strings.ToLower(strings.TrimSpace(hostname))
	switch {
	case strings.Contains(h, "iphone") || strings.Contains(h, "ipad"):
		return "Apple iOS/iPadOS (по имени)", "низкая"
	case strings.Contains(h, "android"):
		return "Android (по имени)", "низкая"
	case strings.Contains(h, "raspberry"):
		return "Linux / Raspberry Pi OS (по имени)", "средняя"
	case strings.Contains(h, "win") && (strings.Contains(h, "pc") || strings.Contains(h, "desktop")):
		return "Windows (по имени)", "низкая"
	}

	open := make(map[int]bool, len(openPorts))
	for _, p := range openPorts {
		open[p] = true
	}
	switch {
	case open[135] && open[445]:
		return "Windows (типичные порты 135/445)", "средняя"
	case open[22] && open[80]:
		return "Linux/Unix или сетевое устройство (22/80)", "низкая"
	case open[548] || open[631]:
		return "macOS / CUPS / печать (548/631)", "низкая"
	default:
		return "", ""
	}
}
