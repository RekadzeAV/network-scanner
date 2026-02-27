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

	// Таблица результатов - порядок колонок: HostName / IP / MAC / Порты
	sb.WriteString("| HostName | IP | MAC | Порты |\n")
	sb.WriteString("|----------|----|-----|-------|\n")

	for _, result := range results {
		// Форматируем порты
		portsStr := formatPorts(result.Ports)
		if portsStr == "" {
			portsStr = "-"
		}

		// Форматируем MAC (если пустой, оставляем пустую строку)
		mac := result.MAC

		// Форматируем hostname
		hostname := result.Hostname
		if hostname == "" {
			hostname = "-"
		}

		// Ограничиваем длину полей для корректного отображения в таблице (до экранирования)
		hostname = strings.TrimSpace(hostname)
		ip := strings.TrimSpace(result.IP)
		mac = strings.TrimSpace(mac)
		portsStr = strings.TrimSpace(portsStr)

		// Убеждаемся, что пустые значения отображаются как "-" (кроме MAC, который может быть пустым)
		if hostname == "" {
			hostname = "-"
		}
		if ip == "" {
			ip = "-"
		}
		// MAC может быть пустым, не заменяем на "-"
		if portsStr == "" {
			portsStr = "-"
		}

		// Ограничиваем длину полей для корректного отображения в таблице
		hostname = truncateString(hostname, 30)
		ip = truncateString(ip, 18)
		mac = truncateString(mac, 18)
		// Порты не обрезаем слишком сильно, чтобы показать больше информации
		portsStr = truncateString(portsStr, 500)

		// Экранируем символы для Markdown таблицы (после обрезки)
		hostname = escapeMarkdown(hostname)
		ip = escapeMarkdown(ip)
		// MAC может быть пустым, разрешаем пустую строку
		mac = escapeMarkdown(mac, true)
		portsStr = escapeMarkdown(portsStr)

		// Формируем строку таблицы с правильным порядком колонок:
		// HostName | IP | MAC | Порты
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			hostname, ip, mac, portsStr))
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
// Ограничивает количество портов для читаемости таблицы
func formatPorts(ports []scanner.PortInfo) string {
	if len(ports) == 0 {
		return ""
	}

	var portStrs []string
	maxPorts := 50 // Максимальное количество портов для отображения
	for i, p := range ports {
		if i >= maxPorts {
			remaining := len(ports) - maxPorts
			portStrs = append(portStrs, fmt.Sprintf("... и еще %d", remaining))
			break
		}
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
// Если передать true для allowEmpty, пустая строка останется пустой (не заменится на "-")
func escapeMarkdown(s string, allowEmpty ...bool) string {
	allowEmptyStr := false
	if len(allowEmpty) > 0 {
		allowEmptyStr = allowEmpty[0]
	}

	if s == "" && !allowEmptyStr {
		return "-"
	}
	if s == "" {
		return ""
	}
	// Заменяем символы, которые могут сломать Markdown таблицу
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	// Убираем множественные пробелы
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// truncateString обрезает строку до указанной длины и добавляет "..."
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
