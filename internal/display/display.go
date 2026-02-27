package display

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"

	"network-scanner/internal/network"
	"network-scanner/internal/scanner"
)

// DisplayResults выводит результаты сканирования в виде таблицы
func DisplayResults(results []scanner.Result) {
	if len(results) == 0 {
		fmt.Println("Результаты сканирования не найдены")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("РЕЗУЛЬТАТЫ СКАНИРОВАНИЯ СЕТИ")
	fmt.Println(strings.Repeat("=", 100) + "\n")

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"IP", "MAC", "Hostname", "Порты", "Протоколы", "Тип устройства", "Производитель"})

	for _, result := range results {
		// Форматируем порты
		portsStr := formatPorts(result.Ports)

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

		t.AppendRow(table.Row{
			result.IP,
			mac,
			hostname,
			portsStr,
			protocolsStr,
			deviceType,
			vendor,
		})
	}

	t.SetStyle(table.StyleColoredBright)
	t.Render()
	fmt.Println()
}

// formatPorts форматирует список портов для отображения
// Ограничивает количество портов для читаемости таблицы
func formatPorts(ports []scanner.PortInfo) string {
	if len(ports) == 0 {
		return "-"
	}

	var portStrs []string
	maxPorts := 50 // Максимальное количество портов для отображения
	openPortsCount := 0
	totalOpenPorts := 0
	
	// Сначала считаем общее количество открытых портов
	for _, p := range ports {
		if p.State == "open" {
			totalOpenPorts++
		}
	}
	
	for _, p := range ports {
		if p.State == "open" {
			if openPortsCount >= maxPorts {
				remaining := totalOpenPorts - maxPorts
				if remaining > 0 {
					portStrs = append(portStrs, fmt.Sprintf("... и еще %d", remaining))
				}
				break
			}
			portStr := fmt.Sprintf("%d/%s", p.Port, p.Protocol)
			if p.Service != "Unknown" {
				portStr += fmt.Sprintf(" (%s)", p.Service)
			}
			portStrs = append(portStrs, portStr)
			openPortsCount++
		} else if p.State == "closed" {
			// Для закрытых портов тоже ограничиваем, но отдельно
			if len(portStrs) >= maxPorts*2 { // Учитываем и открытые, и закрытые
				break
			}
			portStr := fmt.Sprintf("%d/%s (closed)", p.Port, p.Protocol)
			portStrs = append(portStrs, portStr)
		}
	}

	if len(portStrs) == 0 {
		return "-"
	}

	return strings.Join(portStrs, ", ")
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

// DisplayAnalytics выводит аналитику по сети
func DisplayAnalytics(results []scanner.Result) {
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println("АНАЛИТИКА ПРОВОДНЫХ СЕТЕЙ")
	fmt.Println(strings.Repeat("=", 100) + "\n")

	// Статистика по протоколам
	protocolStats := make(map[string]int)
	portStats := make(map[int]int)
	deviceTypes := make(map[string]int)

	for _, result := range results {
		// Подсчет протоколов
		for _, protocol := range result.Protocols {
			protocolStats[protocol]++
		}

		// Подсчет портов
		for _, port := range result.Ports {
			if port.State == "open" {
				portStats[port.Port]++
			}
		}

		// Подсчет типов устройств
		if result.DeviceType != "" {
			deviceTypes[result.DeviceType]++
		}
	}

	// Вывод статистики по протоколам
	fmt.Println("📊 ПРОТОКОЛЫ В СЕТИ:")
	fmt.Println(strings.Repeat("-", 100))
	if len(protocolStats) == 0 {
		fmt.Println("Протоколы не обнаружены")
	} else {
		protocolList := make([]struct {
			name  string
			count int
		}, 0, len(protocolStats))

		for protocol, count := range protocolStats {
			protocolList = append(protocolList, struct {
				name  string
				count int
			}{protocol, count})
		}

		sort.Slice(protocolList, func(i, j int) bool {
			return protocolList[i].count > protocolList[j].count
		})

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Протокол", "Количество устройств", "Описание"})

		for _, item := range protocolList {
			description := getProtocolDescription(item.name)
			t.AppendRow(table.Row{item.name, item.count, description})
		}

		t.SetStyle(table.StyleColoredBright)
		t.Render()
	}
	fmt.Println()

	// Вывод статистики по портам
	fmt.Println("🔌 ИСПОЛЬЗУЕМЫЕ ПОРТЫ:")
	fmt.Println(strings.Repeat("-", 100))
	if len(portStats) == 0 {
		fmt.Println("Открытые порты не обнаружены")
	} else {
		portList := make([]struct {
			port  int
			count int
		}, 0, len(portStats))

		for port, count := range portStats {
			portList = append(portList, struct {
				port  int
				count int
			}{port, count})
		}

		sort.Slice(portList, func(i, j int) bool {
			return portList[i].count > portList[j].count
		})

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Порт", "Количество устройств", "Сервис", "Назначение"})

		for _, item := range portList {
			service := getServiceNameForDisplay(item.port)
			purpose := getPortPurpose(item.port)
			t.AppendRow(table.Row{item.port, item.count, service, purpose})
		}

		t.SetStyle(table.StyleColoredBright)
		t.Render()
	}
	fmt.Println()

	// Вывод статистики по типам устройств
	fmt.Println("🖥️  ТИПЫ УСТРОЙСТВ В СЕТИ:")
	fmt.Println(strings.Repeat("-", 100))
	if len(deviceTypes) == 0 {
		fmt.Println("Типы устройств не определены")
	} else {
		deviceList := make([]struct {
			deviceType string
			count      int
		}, 0, len(deviceTypes))

		for deviceType, count := range deviceTypes {
			deviceList = append(deviceList, struct {
				deviceType string
				count      int
			}{deviceType, count})
		}

		sort.Slice(deviceList, func(i, j int) bool {
			return deviceList[i].count > deviceList[j].count
		})

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Тип устройства", "Количество"})

		for _, item := range deviceList {
			t.AppendRow(table.Row{item.deviceType, item.count})
		}

		t.SetStyle(table.StyleColoredBright)
		t.Render()
	}
	fmt.Println()

	// Общая статистика
	fmt.Println("📈 ОБЩАЯ СТАТИСТИКА:")
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("Всего обнаружено устройств: %d\n", len(results))
	fmt.Printf("Устройств с открытыми портами: %d\n", countDevicesWithOpenPorts(results))
	fmt.Printf("Всего открытых портов: %d\n", countTotalOpenPorts(results))
	fmt.Printf("Уникальных протоколов: %d\n", len(protocolStats))
	fmt.Printf("Уникальных портов: %d\n", len(portStats))
	fmt.Println()
}

// getServiceNameForDisplay возвращает название сервиса по порту (для отображения)
func getServiceNameForDisplay(port int) string {
	// Используем функцию из пакета network для единообразия
	return network.GetServiceName(port)
}

// getProtocolDescription возвращает описание протокола
func getProtocolDescription(protocol string) string {
	descriptions := map[string]string{
		"HTTP":       "Протокол передачи гипертекста - используется для веб-серверов",
		"HTTPS":      "Безопасный HTTP - зашифрованная передача данных в веб",
		"SSH":        "Secure Shell - удаленное управление системами",
		"FTP":        "File Transfer Protocol - передача файлов",
		"SMTP":       "Simple Mail Transfer Protocol - отправка электронной почты",
		"DNS":        "Domain Name System - разрешение доменных имен",
		"POP3":       "Post Office Protocol - получение электронной почты",
		"IMAP":       "Internet Message Access Protocol - доступ к почте",
		"SMB":        "Server Message Block - файловый обмен в Windows сетях",
		"MySQL":      "База данных MySQL",
		"PostgreSQL": "База данных PostgreSQL",
		"RDP":        "Remote Desktop Protocol - удаленный рабочий стол Windows",
		"VNC":        "Virtual Network Computing - удаленный доступ к рабочему столу",
		"Telnet":     "Устаревший протокол удаленного доступа (небезопасен)",
	}

	if desc, ok := descriptions[protocol]; ok {
		return desc
	}
	return "Неизвестный протокол"
}

// getPortPurpose возвращает назначение порта
func getPortPurpose(port int) string {
	purposes := map[int]string{
		20:   "FTP - передача данных",
		21:   "FTP - управление соединением",
		22:   "SSH - безопасное удаленное управление",
		23:   "Telnet - удаленное управление (небезопасно)",
		25:   "SMTP - отправка почты",
		53:   "DNS - разрешение доменных имен",
		80:   "HTTP - веб-серверы",
		110:  "POP3 - получение почты",
		143:  "IMAP - доступ к почте",
		443:  "HTTPS - защищенный веб",
		445:  "SMB - файловый обмен Windows",
		3306: "MySQL - база данных",
		3389: "RDP - удаленный рабочий стол Windows",
		5432: "PostgreSQL - база данных",
		5900: "VNC - удаленный доступ",
		8080: "HTTP - альтернативный порт для веб",
		8443: "HTTPS - альтернативный порт для защищенного веб",
	}

	if purpose, ok := purposes[port]; ok {
		return purpose
	}
	return "Неизвестное назначение"
}

func countDevicesWithOpenPorts(results []scanner.Result) int {
	count := 0
	for _, result := range results {
		if len(result.Ports) > 0 {
			count++
		}
	}
	return count
}

func countTotalOpenPorts(results []scanner.Result) int {
	count := 0
	for _, result := range results {
		for _, port := range result.Ports {
			if port.State == "open" {
				count++
			}
		}
	}
	return count
}

// FormatResultsAsText форматирует результаты сканирования в текстовый формат
func FormatResultsAsText(results []scanner.Result) string {
	if len(results) == 0 {
		return "Результаты сканирования не найдены\n"
	}

	var sb strings.Builder

	// Заголовок
	sb.WriteString(strings.Repeat("=", 100) + "\n")
	sb.WriteString("РЕЗУЛЬТАТЫ СКАНИРОВАНИЯ СЕТИ\n")
	sb.WriteString(strings.Repeat("=", 100) + "\n\n")

	// Заголовок таблицы
	sb.WriteString(fmt.Sprintf("%-18s %-18s %-25s %-400s %-25s %-25s %-20s\n",
		"IP", "MAC", "Hostname", "Порты", "Протоколы", "Тип устройства", "Производитель"))
	sb.WriteString(strings.Repeat("-", 530) + "\n")

	// Данные
	for _, result := range results {
		portsStr := formatPorts(result.Ports)
		protocolsStr := strings.Join(result.Protocols, ", ")
		if protocolsStr == "" {
			protocolsStr = "-"
		}

		mac := result.MAC
		if mac == "" {
			mac = "-"
		}

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

		// Ограничиваем длину полей для корректного отображения в таблице
		// Ширины столбцов: IP(18), MAC(18), Hostname(25), Порты(400), Протоколы(25), Тип(25), Производитель(20)
		ip := truncateString(strings.TrimSpace(result.IP), 18)
		mac = truncateString(strings.TrimSpace(mac), 18)
		hostname = truncateString(strings.TrimSpace(hostname), 25)
		portsStr = truncateString(strings.TrimSpace(portsStr), 400)
		protocolsStr = truncateString(strings.TrimSpace(protocolsStr), 25)
		deviceType = truncateString(strings.TrimSpace(deviceType), 25)
		vendor = truncateString(strings.TrimSpace(vendor), 20)

		// Убеждаемся, что пустые значения отображаются как "-"
		if ip == "" {
			ip = "-"
		}
		if mac == "" {
			mac = "-"
		}
		if hostname == "" {
			hostname = "-"
		}
		if portsStr == "" {
			portsStr = "-"
		}
		if protocolsStr == "" {
			protocolsStr = "-"
		}
		if deviceType == "" {
			deviceType = "-"
		}
		if vendor == "" {
			vendor = "-"
		}

		// Форматируем строку таблицы с фиксированной шириной столбцов
		sb.WriteString(fmt.Sprintf("%-18s %-18s %-25s %-400s %-25s %-25s %-20s\n",
			ip, mac, hostname, portsStr, protocolsStr, deviceType, vendor))
		sb.WriteString("\n")
	}

	// Аналитика
	sb.WriteString("\n" + strings.Repeat("=", 100) + "\n")
	sb.WriteString("АНАЛИТИКА ПРОВОДНЫХ СЕТЕЙ\n")
	sb.WriteString(strings.Repeat("=", 100) + "\n\n")

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

	sb.WriteString("ПРОТОКОЛЫ В СЕТИ:\n")
	sb.WriteString(strings.Repeat("-", 100) + "\n")
	if len(protocolStats) == 0 {
		sb.WriteString("Протоколы не обнаружены\n")
	} else {
		protocolList := make([]struct {
			name  string
			count int
		}, 0, len(protocolStats))
		for protocol, count := range protocolStats {
			protocolList = append(protocolList, struct {
				name  string
				count int
			}{protocol, count})
		}
		sort.Slice(protocolList, func(i, j int) bool {
			return protocolList[i].count > protocolList[j].count
		})

		for _, item := range protocolList {
			description := getProtocolDescription(item.name)
			sb.WriteString(fmt.Sprintf("%s: %d устройств - %s\n", item.name, item.count, description))
		}
	}
	sb.WriteString("\n")

	// Статистика по портам
	sb.WriteString("ИСПОЛЬЗУЕМЫЕ ПОРТЫ:\n")
	sb.WriteString(strings.Repeat("-", 100) + "\n")
	if len(portStats) == 0 {
		sb.WriteString("Открытые порты не обнаружены\n")
	} else {
		portList := make([]struct {
			port  int
			count int
		}, 0, len(portStats))
		for port, count := range portStats {
			portList = append(portList, struct {
				port  int
				count int
			}{port, count})
		}
		sort.Slice(portList, func(i, j int) bool {
			return portList[i].count > portList[j].count
		})

		for _, item := range portList {
			service := getServiceNameForDisplay(item.port)
			purpose := getPortPurpose(item.port)
			sb.WriteString(fmt.Sprintf("Порт %d: %d устройств - %s (%s)\n", item.port, item.count, service, purpose))
		}
	}
	sb.WriteString("\n")

	// Статистика по типам устройств
	sb.WriteString("ТИПЫ УСТРОЙСТВ В СЕТИ:\n")
	sb.WriteString(strings.Repeat("-", 100) + "\n")
	if len(deviceTypes) == 0 {
		sb.WriteString("Типы устройств не определены\n")
	} else {
		deviceList := make([]struct {
			deviceType string
			count      int
		}, 0, len(deviceTypes))
		for deviceType, count := range deviceTypes {
			deviceList = append(deviceList, struct {
				deviceType string
				count      int
			}{deviceType, count})
		}
		sort.Slice(deviceList, func(i, j int) bool {
			return deviceList[i].count > deviceList[j].count
		})

		for _, item := range deviceList {
			sb.WriteString(fmt.Sprintf("%s: %d\n", item.deviceType, item.count))
		}
	}
	sb.WriteString("\n")

	// Общая статистика
	sb.WriteString("ОБЩАЯ СТАТИСТИКА:\n")
	sb.WriteString(strings.Repeat("-", 100) + "\n")
	sb.WriteString(fmt.Sprintf("Всего обнаружено устройств: %d\n", len(results)))
	sb.WriteString(fmt.Sprintf("Устройств с открытыми портами: %d\n", countDevicesWithOpenPorts(results)))
	sb.WriteString(fmt.Sprintf("Всего открытых портов: %d\n", countTotalOpenPorts(results)))
	sb.WriteString(fmt.Sprintf("Уникальных протоколов: %d\n", len(protocolStats)))
	sb.WriteString(fmt.Sprintf("Уникальных портов: %d\n", len(portStats)))
	sb.WriteString("\n")

	return sb.String()
}

// SaveResultsToFile сохраняет результаты сканирования в текстовый файл
func SaveResultsToFile(results []scanner.Result, filename string) error {
	text := FormatResultsAsText(results)
	return os.WriteFile(filename, []byte(text), 0644)
}

// SaveResultsToJSON сохраняет результаты сканирования в JSON файл
func SaveResultsToJSON(results []scanner.Result, filename string) error {
	// Структуры для JSON экспорта
	type JSONPort struct {
		Port     int    `json:"port"`
		State    string `json:"state"`
		Protocol string `json:"protocol"`
		Service  string `json:"service"`
	}

	type JSONResult struct {
		IP           string     `json:"ip"`
		MAC          string     `json:"mac"`
		Hostname     string     `json:"hostname"`
		Ports        []JSONPort `json:"ports"`
		Protocols    []string   `json:"protocols"`
		DeviceType   string     `json:"device_type"`
		DeviceVendor string     `json:"device_vendor"`
		IsAlive      bool       `json:"is_alive"`
	}

	type JSONAnalytics struct {
		Protocols            map[string]int `json:"protocols"`
		Ports                map[int]int    `json:"ports"`
		DeviceTypes          map[string]int `json:"device_types"`
		TotalOpenPorts       int            `json:"total_open_ports"`
		DevicesWithOpenPorts int            `json:"devices_with_open_ports"`
	}

	type JSONExport struct {
		ScanDate     string          `json:"scan_date"`
		TotalDevices int             `json:"total_devices"`
		Devices      []JSONResult    `json:"devices"`
		Analytics    JSONAnalytics   `json:"analytics"`
	}

	// Преобразуем результаты
	jsonResults := make([]JSONResult, 0, len(results))
	protocolStats := make(map[string]int)
	portStats := make(map[int]int)
	deviceTypes := make(map[string]int)
	totalOpenPorts := 0
	devicesWithOpenPorts := 0

	for _, result := range results {
		jsonPorts := make([]JSONPort, 0, len(result.Ports))
		hasOpenPorts := false

		for _, port := range result.Ports {
			jsonPorts = append(jsonPorts, JSONPort{
				Port:     port.Port,
				State:    port.State,
				Protocol: port.Protocol,
				Service:  port.Service,
			})
			if port.State == "open" {
				portStats[port.Port]++
				totalOpenPorts++
				hasOpenPorts = true
			}
		}

		if hasOpenPorts {
			devicesWithOpenPorts++
		}

		for _, protocol := range result.Protocols {
			protocolStats[protocol]++
		}

		if result.DeviceType != "" {
			deviceTypes[result.DeviceType]++
		}

		jsonResults = append(jsonResults, JSONResult{
			IP:           result.IP,
			MAC:          result.MAC,
			Hostname:     result.Hostname,
			Ports:        jsonPorts,
			Protocols:    result.Protocols,
			DeviceType:   result.DeviceType,
			DeviceVendor: result.DeviceVendor,
			IsAlive:      result.IsAlive,
		})
	}

	export := JSONExport{
		ScanDate:     time.Now().Format("2006-01-02 15:04:05"),
		TotalDevices: len(results),
		Devices:      jsonResults,
		Analytics: JSONAnalytics{
			Protocols:            protocolStats,
			Ports:                portStats,
			DeviceTypes:          deviceTypes,
			TotalOpenPorts:       totalOpenPorts,
			DevicesWithOpenPorts: devicesWithOpenPorts,
		},
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге JSON: %v", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// SaveResultsToCSV сохраняет результаты сканирования в CSV файл
func SaveResultsToCSV(results []scanner.Result, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Записываем заголовок
	header := []string{
		"IP", "MAC", "Hostname", "Ports", "Protocols",
		"Device Type", "Device Vendor", "Is Alive",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("ошибка при записи заголовка: %v", err)
	}

	// Записываем данные
	for _, result := range results {
		// Форматируем порты
		portsStr := formatPorts(result.Ports)
		if portsStr == "" {
			portsStr = "-"
		}

		// Форматируем протоколы
		protocolsStr := strings.Join(result.Protocols, "; ")
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

		isAlive := "true"
		if !result.IsAlive {
			isAlive = "false"
		}

		row := []string{
			result.IP,
			mac,
			hostname,
			portsStr,
			protocolsStr,
			deviceType,
			vendor,
			isAlive,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("ошибка при записи строки: %v", err)
		}
	}

	return nil
}
