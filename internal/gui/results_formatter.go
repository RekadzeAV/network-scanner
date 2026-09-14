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

// deviceTypeWithBadge добавляет префикс к типу устройства.
func deviceTypeWithBadge(deviceType string) string {
	dt := strings.TrimSpace(deviceType)
	if dt == "" {
		return "-"
	}
	low := strings.ToLower(dt)
	switch {
	case strings.Contains(low, "router"), strings.Contains(low, "switch"):
		return "[NET] " + dt
	case strings.Contains(low, "access point"):
		return "[AP] " + dt
	case strings.Contains(low, "printer"):
		return "[PRN] " + dt
	case strings.Contains(low, "camera"):
		return "[CAM] " + dt
	case strings.Contains(low, "nas"):
		return "[NAS] " + dt
	case strings.Contains(low, "iot"):
		return "[IOT] " + dt
	case strings.Contains(low, "desktop"), strings.Contains(low, "laptop"):
		return "[PC] " + dt
	case strings.Contains(low, "server"):
		return "[SRV] " + dt
	case strings.Contains(low, "phone"), strings.Contains(low, "tablet"):
		return "[MOB] " + dt
	default:
		return "[UNK] " + dt
	}
}

// osGuessLine формирует строку с информацией об ОС.
func osGuessLine(r scanner.Result) string {
	if strings.TrimSpace(r.GuessOS) != "" {
		label := strings.TrimSpace(r.GuessOS)
		if strings.TrimSpace(r.GuessOSConfidence) != "" {
			label = fmt.Sprintf("%s (%s)", label, strings.TrimSpace(r.GuessOSConfidence))
		}
		if strings.TrimSpace(r.GuessOSReason) != "" {
			label += " — " + strings.TrimSpace(r.GuessOSReason)
		}
		return label
	}
	return "-"
}
