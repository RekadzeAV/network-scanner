package nettools

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"
)

// GetWiFiInfo возвращает диагностическую информацию Wi-Fi для текущей ОС.
func GetWiFiInfo(ctx context.Context, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	var args []string
	switch runtime.GOOS {
	case "windows":
		args = []string{"netsh", "wlan", "show", "interfaces"}
	case "linux":
		args = []string{"nmcli", "-t", "-f", "IN-USE,SSID,MODE,CHAN,RATE,SIGNAL,SECURITY", "dev", "wifi", "list"}
	case "darwin":
		args = []string{"/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport", "-I"}
	default:
		return "", fmt.Errorf("wifi анализ не поддерживается на %s", runtime.GOOS)
	}

	out, err := runCmd(ctx, args, timeout)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "" {
		return "Wi-Fi: данные не получены (пустой ответ).", nil
	}
	return formatWiFiSummary(runtime.GOOS, out), nil
}

func formatWiFiSummary(goos string, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "Wi-Fi: данные не получены (пустой ответ)."
	}
	var summary map[string]string
	switch goos {
	case "windows":
		summary = parseWindowsNetsh(raw)
	case "linux":
		summary = parseLinuxNmcli(raw)
	case "darwin":
		summary = parseDarwinAirport(raw)
	default:
		summary = map[string]string{}
	}
	if len(summary) == 0 {
		return "Wi-Fi summary: не удалось распознать структурированные поля.\n\nRaw output:\n" + raw
	}

	order := []string{"interface", "ssid", "bssid", "state", "signal", "channel", "rate", "security", "mode"}
	var sb strings.Builder
	sb.WriteString("Wi-Fi summary:\n")
	for _, k := range order {
		if v := strings.TrimSpace(summary[k]); v != "" {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", labelKey(k), v))
		}
	}
	// Добавим прочие ключи, если появились.
	rest := make([]string, 0)
	for k := range summary {
		found := false
		for _, ok := range order {
			if k == ok {
				found = true
				break
			}
		}
		if !found {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	for _, k := range rest {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", labelKey(k), summary[k]))
	}
	sb.WriteString("\nRaw output:\n")
	sb.WriteString(raw)
	return sb.String()
}

func parseWindowsNetsh(raw string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		k, v, ok := splitKV(line)
		if !ok {
			continue
		}
		switch normalizeWindowsWiFiKey(k) {
		case "name":
			out["interface"] = v
		case "ssid":
			if strings.TrimSpace(v) != "" && !strings.HasPrefix(strings.ToLower(strings.TrimSpace(k)), "ssid ") {
				out["ssid"] = v
			}
		case "bssid":
			out["bssid"] = v
		case "state":
			out["state"] = normalizeWindowsWiFiState(v)
		case "signal":
			out["signal"] = v
		case "channel":
			out["channel"] = v
		case "receive rate (mbps)", "transmit rate (mbps)":
			if cur := strings.TrimSpace(out["rate"]); cur == "" {
				out["rate"] = v + " Mbps"
			} else {
				out["rate"] = cur + " / " + v + " Mbps"
			}
		case "authentication":
			out["security"] = v
		case "radio type":
			out["mode"] = v
		}
	}
	if strings.TrimSpace(out["state"]) == "" {
		out["state"] = "unknown"
	}
	return out
}

func normalizeWindowsWiFiKey(k string) string {
	key := strings.ToLower(strings.TrimSpace(k))
	switch key {
	case "name", "имя":
		return "name"
	case "ssid":
		return "ssid"
	case "bssid":
		return "bssid"
	case "state", "состояние":
		return "state"
	case "signal", "сигнал":
		return "signal"
	case "channel", "канал":
		return "channel"
	case "receive rate (mbps)", "скорость приема (мбит/с)":
		return "receive rate (mbps)"
	case "transmit rate (mbps)", "скорость передачи (мбит/с)":
		return "transmit rate (mbps)"
	case "authentication", "проверка подлинности":
		return "authentication"
	case "radio type", "тип радиомодуля", "тип радио":
		return "radio type"
	default:
		return key
	}
}

func normalizeWindowsWiFiState(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "connected", "подключено":
		return "connected"
	case "disconnected", "отключено", "не подключено":
		return "disconnected"
	default:
		return strings.TrimSpace(v)
	}
}

func parseLinuxNmcli(raw string) map[string]string {
	out := map[string]string{}
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := splitNmcliFields(line)
		if len(parts) < 7 {
			continue
		}
		// IN-USE:SSID:MODE:CHAN:RATE:SIGNAL:SECURITY
		if strings.TrimSpace(parts[0]) != "*" {
			continue
		}
		out["ssid"] = strings.TrimSpace(parts[1])
		out["mode"] = strings.TrimSpace(parts[2])
		out["channel"] = strings.TrimSpace(parts[3])
		out["rate"] = strings.TrimSpace(parts[4])
		out["signal"] = strings.TrimSpace(parts[5]) + "%"
		out["security"] = strings.TrimSpace(parts[6])
		out["state"] = "connected"
		break
	}
	if len(out) == 0 && len(lines) > 0 {
		out["state"] = "not_connected_or_hidden"
	}
	return out
}

func splitNmcliFields(line string) []string {
	parts := make([]string, 0, 7)
	var cur strings.Builder
	escaped := false
	for _, r := range line {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == ':':
			parts = append(parts, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if escaped {
		// Preserve trailing '\' literally for robustness.
		cur.WriteRune('\\')
	}
	parts = append(parts, cur.String())
	return parts
}

func parseDarwinAirport(raw string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		k, v, ok := splitKV(line)
		if !ok {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(k)) {
		case "ssid":
			if strings.TrimSpace(v) != "" {
				out["ssid"] = v
			}
		case "bssid":
			if strings.TrimSpace(v) != "" {
				out["bssid"] = v
			}
		case "agrctlrssi":
			if strings.TrimSpace(v) != "" {
				out["signal"] = v + " dBm"
			}
		case "channel":
			if strings.TrimSpace(v) != "" && strings.TrimSpace(v) != "--" {
				out["channel"] = v
			}
		case "lasttxrate":
			if strings.TrimSpace(v) != "" {
				out["rate"] = v + " Mbps"
			}
		case "link auth":
			if strings.TrimSpace(v) != "" {
				out["security"] = v
			}
		case "state":
			out["state"] = normalizeDarwinWiFiState(v)
		case "airport":
			if strings.EqualFold(strings.TrimSpace(v), "off") {
				out["state"] = "disconnected"
			}
		}
	}
	if strings.TrimSpace(out["state"]) == "" && strings.TrimSpace(out["ssid"]) != "" {
		out["state"] = "connected"
	}
	if strings.TrimSpace(out["state"]) == "" && len(out) > 0 {
		out["state"] = "unknown"
	}
	return out
}

func normalizeDarwinWiFiState(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "running":
		return "connected"
	case "init", "off":
		return "disconnected"
	default:
		if s == "" {
			return ""
		}
		return strings.TrimSpace(v)
	}
}

func splitKV(line string) (k, v string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}
	idx := strings.Index(line, ":")
	if idx <= 0 {
		return "", "", false
	}
	k = strings.TrimSpace(line[:idx])
	v = strings.TrimSpace(line[idx+1:])
	if k == "" {
		return "", "", false
	}
	return k, v, true
}

func labelKey(k string) string {
	k = strings.TrimSpace(k)
	if k == "" {
		return ""
	}
	return strings.ToUpper(k[:1]) + k[1:]
}
