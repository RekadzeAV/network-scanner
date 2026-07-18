package scanner_test

import (
	"fmt"

	"network-scanner/internal/scanner"
)

// Example_result демонстрирует структуру результата сканирования.
func Example_result() {
	result := scanner.Result{
		IP:              "192.168.1.1",
		Hostname:        "gateway.local",
		MAC:             "AA:BB:CC:DD:EE:FF",
		DeviceType:      "Router",
		DeviceVendor:    "Cisco Systems",
		SNMPEnabled:     true,
		IsAlive:         true,
		GuessOS:         "Linux",
		GuessOSConfidence: "высокая",
		Ports: []scanner.PortInfo{
			{Port: 22, State: "open", Protocol: "tcp", Service: "SSH", Version: "OpenSSH 8.9"},
			{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"},
			{Port: 443, State: "open", Protocol: "tcp", Service: "HTTPS"},
		},
	}

	fmt.Printf("IP: %s\n", result.IP)
	fmt.Printf("Тип: %s (%s)\n", result.DeviceType, result.DeviceVendor)
	fmt.Printf("Открытых портов: %d\n", len(result.Ports))
	fmt.Printf("SNMP: %v, ОС: %s (%s)\n", result.SNMPEnabled, result.GuessOS, result.GuessOSConfidence)

	// Output:
	// IP: 192.168.1.1
	// Тип: Router (Cisco Systems)
	// Открытых портов: 3
	// SNMP: true, ОС: Linux (высокая)
}

// Example_portInfo демонстрирует информацию о портах.
func Example_portInfo() {
	ports := []scanner.PortInfo{
		{Port: 22, State: "open", Protocol: "tcp", Service: "SSH", Version: "OpenSSH 8.9", Banner: "SSH-2.0-OpenSSH_8.9"},
		{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"},
		{Port: 3306, State: "closed", Protocol: "tcp", Service: "MySQL"},
	}

	for _, p := range ports {
		fmt.Printf("%d/%s: %s (%s)", p.Port, p.Protocol, p.State, p.Service)
		if p.Version != "" {
			fmt.Printf(" [v%s]", p.Version)
		}
		fmt.Println()
	}

	// Output:
	// 22/tcp: open (SSH) [vOpenSSH 8.9]
	// 80/tcp: open (HTTP)
	// 3306/tcp: closed (MySQL)
}
