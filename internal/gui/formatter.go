package gui

import (
	"fmt"
	"strings"

	"network-scanner/internal/scanner"
)

// FormatResultsForDisplay форматирует результаты сканирования в Markdown формат для отображения в GUI
func FormatResultsForDisplay(results []scanner.Result) string {
	if len(results) == 0 {
		return "## Результаты сканирования\n\nРезультаты сканирования не найдены."
	}

	var sb strings.Builder
	
	sb.WriteString("## Результаты сканирования сети\n\n")
	sb.WriteString(fmt.Sprintf("**Найдено устройств:** %d\n\n", len(results)))
	sb.WriteString("---\n\n")

	// Таблица результатов
	sb.WriteString("| IP | MAC | Hostname | Порты | Протоколы | Тип устройства | Производитель |\n")
	sb.WriteString("|----|-----|----------|-------|-----------|----------------|---------------|\n")

	for _, result := range results {
		// Форматируем порты
		portsStr := formatPorts(result.Ports)
		if portsStr == "" {
			portsStr = "-"
		}
		
		// Форматируем протоколы
		protocolsStr := strings.Join(result.Protocols, ", ")
		if protocolsStr == "" {
			protocolsStr = "-"
		}

		// Форматируем MAC
		mac := result.MAC
		if mac == "" {
			mac = "-"
		}

		// Форматируем hostname
		hostname := result.Hostname
		if hostname == "" {
			hostname = "-"
		}

		deviceType := result.DeviceType
		if deviceType == "" {
			deviceType = "Unknown"
		}

		vendor := result.DeviceVendor
		if vendor == "" {
			vendor = "-"
		}

		// Экранируем символы для Markdown таблицы
		ip := escapeMarkdown(result.IP)
		mac = escapeMarkdown(mac)
		hostname = escapeMarkdown(hostname)
		portsStr = escapeMarkdown(portsStr)
		protocolsStr = escapeMarkdown(protocolsStr)
		deviceType = escapeMarkdown(deviceType)
		vendor = escapeMarkdown(vendor)

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |\n",
			ip, mac, hostname, portsStr, protocolsStr, deviceType, vendor))
	}

	sb.WriteString("\n---\n\n")

	// Аналитика
	sb.WriteString("## Аналитика\n\n")

	// Статистика по протоколам
	protocolStats := make(map[string]int)
	portStats := make(map[int]int)
	deviceTypes := make(map[string]int)

	for _, result := range results {
		for _, protocol := range result.Protocols {
			protocolStats[protocol]++
		}
		for _, port := range result.Ports {
			if port.State == "open" {
				portStats[port.Port]++
			}
		}
		if result.DeviceType != "" {
			deviceTypes[result.DeviceType]++
		}
	}

	if len(protocolStats) > 0 {
		sb.WriteString("### Протоколы в сети\n\n")
		for protocol, count := range protocolStats {
			sb.WriteString(fmt.Sprintf("- **%s**: %d устройств\n", protocol, count))
		}
		sb.WriteString("\n")
	}

	if len(portStats) > 0 {
		sb.WriteString("### Используемые порты\n\n")
		for port, count := range portStats {
			sb.WriteString(fmt.Sprintf("- **Порт %d**: %d устройств\n", port, count))
		}
		sb.WriteString("\n")
	}

	if len(deviceTypes) > 0 {
		sb.WriteString("### Типы устройств\n\n")
		for deviceType, count := range deviceTypes {
			sb.WriteString(fmt.Sprintf("- **%s**: %d\n", deviceType, count))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatPorts форматирует список портов для отображения
func formatPorts(ports []scanner.PortInfo) string {
	if len(ports) == 0 {
		return ""
	}

	var portStrs []string
	for _, p := range ports {
		if p.State == "open" {
			portStr := fmt.Sprintf("%d/%s", p.Port, p.Protocol)
			if p.Service != "Unknown" {
				portStr += fmt.Sprintf(" (%s)", p.Service)
			}
			portStrs = append(portStrs, portStr)
		}
	}

	return strings.Join(portStrs, ", ")
}

// escapeMarkdown экранирует специальные символы Markdown
func escapeMarkdown(s string) string {
	// Заменяем символы, которые могут сломать Markdown таблицу
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}




