package osdetect

import "strings"

// GuessFromHostAndPorts эвристика по имени хоста и номерам открытых TCP-портов.
// activeMode включает дополнительные (более смелые) эвристики.
func GuessFromHostAndPorts(hostname string, openPorts []int, activeMode bool) (osName, confidence, reason string) {
	h := strings.ToLower(strings.TrimSpace(hostname))
	switch {
	case strings.Contains(h, "iphone") || strings.Contains(h, "ipad"):
		return "Apple iOS/iPadOS", "средняя", "hostname содержит iphone/ipad"
	case strings.Contains(h, "android"):
		return "Android", "средняя", "hostname содержит android"
	case strings.Contains(h, "raspberry"):
		return "Linux / Raspberry Pi OS", "средняя", "hostname содержит raspberry"
	case strings.Contains(h, "win") && (strings.Contains(h, "pc") || strings.Contains(h, "desktop")):
		return "Windows", "низкая", "hostname содержит win+pc/desktop"
	}

	open := make(map[int]bool, len(openPorts))
	for _, p := range openPorts {
		open[p] = true
	}
	switch {
	case open[135] && open[445]:
		return "Windows", "средняя", "открыты порты 135 и 445"
	case open[139] && open[445]:
		return "Windows", "средняя", "открыты порты 139 и 445"
	case open[22] && open[80]:
		return "Linux/Unix или сетевое устройство", "низкая", "открыты порты 22 и 80"
	case open[22] && open[443]:
		return "Linux/Unix или сетевое устройство", "низкая", "открыты порты 22 и 443"
	case open[548] || open[631]:
		return "macOS / CUPS / печать", "низкая", "открыт порт 548 или 631"
	}

	if activeMode {
		switch {
		case open[3389] && (open[445] || open[139]):
			return "Windows", "высокая", "active-эвристика: RDP + SMB/NetBIOS"
		case open[5985] && (open[445] || open[139]):
			return "Windows Server", "средняя", "active-эвристика: WinRM + SMB/NetBIOS"
		case open[22] && (open[3306] || open[5432] || open[6379]):
			return "Linux/Unix Server", "средняя", "active-эвристика: SSH + DB-порты"
		case open[22] && (open[2375] || open[2376]):
			return "Linux/Unix Server", "средняя", "active-эвристика: SSH + Docker API"
		case open[22] && open[6443]:
			return "Linux/Unix Server", "средняя", "active-эвристика: SSH + Kubernetes API"
		case open[5353] && (open[548] || open[62078]):
			return "Apple iOS/macOS", "высокая", "active-эвристика: mDNS + Apple service ports"
		case open[62078]:
			return "Apple iOS/macOS", "средняя", "active-эвристика: порт 62078 (lockdown)"
		case open[5555] && open[8081]:
			return "Android", "средняя", "active-эвристика: Android ADB/TCP + debug web"
		}
	}
	return "", "", ""
}
