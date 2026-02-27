package gui

import (
	"fmt"
	"sort"
	"strings"

	"network-scanner/internal/scanner"
)

// FormatResultsForDisplay форматирует результаты сканирования в Markdown формат для отображения в GUI
func FormatResultsForDisplay(results []scanner.Result) string {
	if len(results) == 0 {
		return "## Результаты сканирования\n\nРезультаты сканирования не найдены."
	}

	var sb strings.Builder

	sb.WriteString("## Устройства\n\n")
	sb.WriteString(fmt.Sprintf("**Найдено устройств:** %d\n\n", len(results)))
	sb.WriteString("---\n\n")

	// Таблица результатов - порядок колонок: HostName / IP / MAC / Порты / Протокол / Тип устройства / Производитель
	sb.WriteString("| HostName | IP | MAC | Порты | Протокол | Тип устройства | Производитель |\n")
	sb.WriteString("|----------|----|-----|-------|----------|----------------|---------------|\n")

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
			hostname, ip, mac, portsStr, protocolsStr, deviceType, vendor))
	}

	sb.WriteString("\n---\n\n")

	// Аналитика
	sb.WriteString("## Сетевая аналитика\n\n")

	// Статистика по протоколам
	protocolStats := make(map[string]int)
	deviceTypes := make(map[string]int)

	for _, result := range results {
		for _, protocol := range result.Protocols {
			protocolStats[protocol]++
		}
		if result.DeviceType != "" {
			deviceTypes[result.DeviceType]++
		}
	}

	// Таблица протоколов - каждый протокол как отдельная колонка
	if len(protocolStats) > 0 {
		sb.WriteString("### Протоколы в сети\n\n")

		// Сортируем протоколы для единообразия
		protocols := make([]string, 0, len(protocolStats))
		for protocol := range protocolStats {
			protocols = append(protocols, protocol)
		}
		sort.Strings(protocols)

		// Создаем заголовок таблицы
		header := "|"
		separator := "|"
		for _, protocol := range protocols {
			header += fmt.Sprintf(" %s |", escapeMarkdown(protocol))
			separator += "---|"
		}
		sb.WriteString(header + "\n")
		sb.WriteString(separator + "\n")

		// Создаем строку данных
		dataRow := "|"
		for _, protocol := range protocols {
			dataRow += fmt.Sprintf(" %d |", protocolStats[protocol])
		}
		sb.WriteString(dataRow + "\n\n")
	}

	// Таблица типов устройств - каждый тип как отдельная колонка
	if len(deviceTypes) > 0 {
		sb.WriteString("### Типы устройств\n\n")

		// Нормализуем названия типов устройств для колонок
		typeMapping := map[string]string{
			// Network Device категория
			"Router/Network Device": "Network Device",
			"Network Device":        "Network Device",
			"Router":                "Network Device",
			"Printer":               "Network Device",
			"IoT Device":            "Network Device",
			"IoT":                   "Network Device",

			// Computer категория
			"Windows Computer": "Computer",
			"Computer":         "Computer",
			"Windows":          "Computer",
			"PC":               "Computer",
			"Desktop":          "Computer",
			"Laptop":           "Computer",

			// Server категория
			"Web Server":        "Server",
			"Database Server":   "Server",
			"Linux/Unix Server": "Server",
			"Server":            "Server",
			"Linux Server":      "Server",
			"Unix Server":       "Server",
			"Linux":             "Server",
			"Unix":              "Server",

			// Unknown
			"Unknown Device": "Unknown",
			"Unknown":        "Unknown",
		}

		// Группируем по нормализованным типам
		normalizedTypes := make(map[string]int)
		for deviceType, count := range deviceTypes {
			normalized := typeMapping[deviceType]
			if normalized == "" {
				normalized = deviceType
			}
			normalizedTypes[normalized] += count
		}

		// Определяем порядок колонок
		columnOrder := []string{"Network Device", "Computer", "Server", "Unknown"}
		existingTypes := make([]string, 0)

		// Добавляем типы в нужном порядке, если они есть
		for _, colType := range columnOrder {
			if normalizedTypes[colType] > 0 {
				existingTypes = append(existingTypes, colType)
			}
		}

		// Добавляем остальные типы, которых нет в стандартном списке
		for deviceType := range normalizedTypes {
			found := false
			for _, colType := range columnOrder {
				if deviceType == colType {
					found = true
					break
				}
			}
			if !found {
				existingTypes = append(existingTypes, deviceType)
			}
		}

		if len(existingTypes) > 0 {
			// Создаем заголовок таблицы
			header := "|"
			separator := "|"
			for _, deviceType := range existingTypes {
				header += fmt.Sprintf(" %s |", escapeMarkdown(deviceType))
				separator += "---|"
			}
			sb.WriteString(header + "\n")
			sb.WriteString(separator + "\n")

			// Создаем строку данных
			dataRow := "|"
			for _, deviceType := range existingTypes {
				dataRow += fmt.Sprintf(" %d |", normalizedTypes[deviceType])
			}
			sb.WriteString(dataRow + "\n\n")
		}
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
