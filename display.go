package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// displayResults выводит результаты сканирования в виде таблицы
func displayResults(results []ScanResult) {
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
func formatPorts(ports []PortInfo) string {
	if len(ports) == 0 {
		return "-"
	}

	var portStrs []string
	for _, p := range ports {
		if p.State == "open" {
			portStr := fmt.Sprintf("%d/%s", p.Port, p.Protocol)
			if p.Service != "Unknown" {
				portStr += fmt.Sprintf(" (%s)", p.Service)
			}
			portStrs = append(portStrs, portStr)
		} else if p.State == "closed" {
			portStr := fmt.Sprintf("%d/%s (closed)", p.Port, p.Protocol)
			portStrs = append(portStrs, portStr)
		}
	}

	if len(portStrs) == 0 {
		return "-"
	}

	return strings.Join(portStrs, ", ")
}

// displayAnalytics выводит аналитику по сети
func displayAnalytics(results []ScanResult) {
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
			service := getServiceName(item.port)
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

// getProtocolDescription возвращает описание протокола
func getProtocolDescription(protocol string) string {
	descriptions := map[string]string{
		"HTTP":    "Протокол передачи гипертекста - используется для веб-серверов",
		"HTTPS":   "Безопасный HTTP - зашифрованная передача данных в веб",
		"SSH":     "Secure Shell - удаленное управление системами",
		"FTP":     "File Transfer Protocol - передача файлов",
		"SMTP":    "Simple Mail Transfer Protocol - отправка электронной почты",
		"DNS":     "Domain Name System - разрешение доменных имен",
		"POP3":    "Post Office Protocol - получение электронной почты",
		"IMAP":    "Internet Message Access Protocol - доступ к почте",
		"SMB":     "Server Message Block - файловый обмен в Windows сетях",
		"MySQL":   "База данных MySQL",
		"PostgreSQL": "База данных PostgreSQL",
		"RDP":     "Remote Desktop Protocol - удаленный рабочий стол Windows",
		"VNC":     "Virtual Network Computing - удаленный доступ к рабочему столу",
		"Telnet":  "Устаревший протокол удаленного доступа (небезопасен)",
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

func countDevicesWithOpenPorts(results []ScanResult) int {
	count := 0
	for _, result := range results {
		if len(result.Ports) > 0 {
			count++
		}
	}
	return count
}

func countTotalOpenPorts(results []ScanResult) int {
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

