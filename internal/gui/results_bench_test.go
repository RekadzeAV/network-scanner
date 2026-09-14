package gui

import (
	"fmt"
	"testing"

	"network-scanner/internal/scanner"
)

// BenchmarkBuildTableView бенчмарк для рендеринга таблицы.
func BenchmarkBuildTableView(b *testing.B) {
	app := &App{
		resultsMode:   "таблица",
		layoutProfile: "normal",
	}

	data := generateTestResults(500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		view := app.buildTableView(data)
		if view == nil {
			b.Fatal("buildTableView returned nil")
		}
	}
	b.ReportAllocs()
}

// BenchmarkBuildCardsView бенчмарк для рендеринга карточек.
func BenchmarkBuildCardsView(b *testing.B) {
	app := &App{
		resultsMode:       "Карточки",
		layoutProfile:     "normal",
		cardsVisibleCount: 200,
	}

	data := generateTestResults(500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		view := app.buildCardsView(data)
		if view == nil {
			b.Fatal("buildCardsView returned nil")
		}
	}
	b.ReportAllocs()
}

// generateTestResults генерирует тестовые данные.
func generateTestResults(count int) []scanner.Result {
	data := make([]scanner.Result, count)
	for i := 0; i < count; i++ {
		data[i] = scanner.Result{
			Hostname:          fmt.Sprintf("host-%d.local", i),
			IP:                fmt.Sprintf("192.168.1.%d", i%256),
			MAC:               fmt.Sprintf("AA:BB:CC:DD:EE:%02X", i%256),
			DeviceType:        deviceTypeForIndex(i),
			DeviceVendor:      "Test Vendor",
			GuessOS:           "Linux",
			GuessOSConfidence: "85%",
			GuessOSReason:     "TCP/IP stack",
			SNMPEnabled:       i%3 == 0,
			Ports:             []scanner.PortInfo{{Port: 22, Protocol: "tcp", State: "open", Service: "ssh"}},
		}
	}
	return data
}

// deviceTypeForIndex возвращает тип устройства по индексу.
func deviceTypeForIndex(i int) string {
	switch i % 8 {
	case 0:
		return "Router"
	case 1:
		return "Switch"
	case 2:
		return "Server"
	case 3:
		return "Desktop"
	case 4:
		return "Printer"
	case 5:
		return "Camera"
	case 6:
		return "NAS"
	default:
		return "IoT Device"
	}
}
