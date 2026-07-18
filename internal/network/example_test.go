package network_test

import (
	"fmt"

	"network-scanner/internal/network"
)

// ExampleParseNetworkRange демонстрирует парсинг CIDR диапазона.
func ExampleParseNetworkRange() {
	// Парсим диапазон сети
	ips, err := network.ParseNetworkRange("192.168.1.0/30")
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("Найдено IP: %d\n", len(ips))
	for i, ip := range ips {
		fmt.Printf("  %d: %s\n", i+1, ip.String())
	}

	// Output:
	// Найдено IP: 2
	//   1: 192.168.1.1
	//   2: 192.168.1.2
}

// ExampleEstimateHostCount демонстрирует оценку количества хостов в подсети.
func ExampleEstimateHostCount() {
	// Оцениваем количество хостов в разных подсетях
	count, err := network.EstimateHostCount("192.168.1.0/24")
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	fmt.Printf("/24: ~%d хостов\n", count)

	count, err = network.EstimateHostCount("192.168.0.0/16")
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	fmt.Printf("/16: ~%d хостов\n", count)

	// Output:
	// /24: ~254 хостов
	// /16: ~65534 хостов
}

// ExampleParsePortRange демонстрирует парсинг диапазонов портов.
func ExampleParsePortRange() {
	// Парсим разные форматы диапазонов портов
	formats := []string{
		"1-100",
		"22,80,443",
		"80,443-445,8080",
	}

	for _, format := range formats {
		ports, err := network.ParsePortRange(format)
		if err != nil {
			fmt.Printf("%s: ошибка — %v\n", format, err)
			continue
		}
		fmt.Printf("%s: %d портов — %v\n", format, len(ports), ports)
	}

	// Output:
	// 1-100: 100 портов — [1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 48 49 50 51 52 53 54 55 56 57 58 59 60 61 62 63 64 65 66 67 68 69 70 71 72 73 74 75 76 77 78 79 80 81 82 83 84 85 86 87 88 89 90 91 92 93 94 95 96 97 98 99 100]
	// 22,80,443: 3 портов — [22 80 443]
	// 80,443-445,8080: 5 портов — [80 443 444 445 8080]
}

// ExampleIsPortOpen демонстрирует проверку открытости TCP порта.
func ExampleIsPortOpen() {
	// Проверяем порт 80 на localhost
	isOpen := network.IsPortOpen("127.0.0.1", 80, 1000000000) // 1 сек
	fmt.Printf("Порт 80 открыт: %v\n", isOpen)

	// Output:
	// Порт 80 открыт: false
}

// ExampleDetectLocalNetwork демонстрирует автоматическое определение локальной сети.
func ExampleDetectLocalNetwork() {
	// Определяем локальную сеть
	networkCIDR, err := network.DetectLocalNetwork()
	if err != nil {
		fmt.Printf("Ошибка определения сети: %v\n", err)
		return
	}

	fmt.Printf("Локальная сеть: %s\n", networkCIDR)
	// Output будет зависеть от конфигурации сети пользователя
}

// ExampleGetServiceName демонстрирует получение имени сервиса по порту.
func ExampleGetServiceName() {
	// Получаем имена сервисов для типовых портов
	ports := []int{22, 80, 443, 3306, 5432, 8080}

	for _, port := range ports {
		service := network.GetServiceName(port)
		fmt.Printf("Порт %d: %s\n", port, service)
	}

	// Output:
	// Порт 22: SSH
	// Порт 80: HTTP
	// Порт 443: HTTPS
	// Порт 3306: MySQL
	// Порт 5432: PostgreSQL
	// Порт 8080: HTTP-Proxy
}
