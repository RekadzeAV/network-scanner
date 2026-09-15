package topology_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"network-scanner/internal/scanner"
	"network-scanner/internal/topology"
)

// ExampleBuildTopology демонстрирует построение топологии сети.
func ExampleBuildTopology() {
	// Создаем тестовые результаты сканирования
	results := []scanner.Result{
		{
			IP:          "192.168.1.1",
			MAC:         "aa:bb:cc:dd:ee:01",
			Hostname:    "router.local",
			DeviceType:  "router",
			SNMPEnabled: true,
		},
		{
			IP:          "192.168.1.10",
			MAC:         "aa:bb:cc:dd:ee:0a",
			Hostname:    "workstation.local",
			DeviceType:  "host",
			SNMPEnabled: false,
		},
	}

	// SNMP-данные (обычно собираются отдельно)
	snmpData := map[string]*topology.Device{
		"aa:bb:cc:dd:ee:01": {
			IP:          "192.168.1.1",
			MAC:         "aa:bb:cc:dd:ee:01",
			Hostname:    "router.local",
			Type:        topology.DeviceTypeRouter,
			SNMPEnabled: true,
		},
	}

	// Строим топологию
	topo, err := topology.BuildTopology(results, snmpData)
	if err != nil {
		fmt.Printf("Ошибка построения топологии: %v\n", err)
		return
	}

	// Выводим статистику
	fmt.Printf("Устройств: %d\n", len(topo.Devices))
	fmt.Printf("Связей: %d\n", len(topo.Links))

	// Выводим устройства (отсортировано для детерминированного вывода)
	devices := make([]*topology.Device, 0, len(topo.Devices))
	for _, dev := range topo.Devices {
		devices = append(devices, dev)
	}
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Hostname < devices[j].Hostname
	})
	for _, dev := range devices {
		fmt.Printf("  - %s (%s) [%s]\n", dev.Hostname, dev.IP, dev.Type)
	}

	// Output:
	// Устройств: 2
	// Связей: 0
	//   - router.local (192.168.1.1) [router]
	//   - workstation.local (192.168.1.10) [host]
}

// ExampleTopology_Validate демонстрирует валидацию топологии.
func ExampleTopology_Validate() {
	// Создаем валидную топологию
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{
			"mac_aa:bb:cc:dd:ee:01": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:01",
				Hostname: "router.local",
				Type:     topology.DeviceTypeRouter,
			},
		},
		Links: []topology.Link{},
	}

	// Валидируем
	err := topo.Validate()
	if err != nil {
		fmt.Printf("Валидация не пройдена: %v\n", err)
		return
	}

	fmt.Println("Топология валидна")

	// Output:
	// Топология валидна
}

// ExampleTopology_SaveJSON демонстрирует сохранение топологии в JSON.
func ExampleTopology_SaveJSON() {
	// Создаем тестовую топологию
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{
			"mac_aa:bb:cc:dd:ee:01": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:01",
				Hostname: "router.local",
				Type:     topology.DeviceTypeRouter,
			},
		},
		Links: []topology.Link{},
	}

	// Сохраняем в JSON (временный файл)
	tmpFile := filepath.Join(os.TempDir(), "test-topology.json")
	defer os.Remove(tmpFile)
	err := topo.SaveJSON(tmpFile)
	if err != nil {
		fmt.Printf("Ошибка сохранения: %v\n", err)
		return
	}

	// Проверяем, что файл создан
	fmt.Printf("Топология сохранена в: %s\n", tmpFile)
}

// ExampleLinkSourceType демонстрирует типы источников связей.
func ExampleLinkSourceType() {
	// Перечисляем все типы источников связей
	sources := []topology.LinkSourceType{
		topology.LinkSourceLLDP,
		topology.LinkSourceFDB,
		topology.LinkSourceInferred,
	}

	for _, source := range sources {
		fmt.Printf("Источник: %s\n", source)
	}

	// Output:
	// Источник: lldp
	// Источник: fdb
	// Источник: inferred
}

// ExampleLinkConfidence демонстрирует уровни уверенности в связях.
func ExampleLinkConfidence() {
	// Перечисляем все уровни уверенности
	confidences := []topology.LinkConfidence{
		topology.LinkConfidenceHigh,
		topology.LinkConfidenceMedium,
		topology.LinkConfidenceLow,
	}

	for _, conf := range confidences {
		fmt.Printf("Уверенность: %s\n", conf)
	}

	// Output:
	// Уверенность: high
	// Уверенность: medium
	// Уверенность: low
}

// ExampleDeviceType демонстрирует типы устройств.
func ExampleDeviceType() {
	// Перечисляем все типы устройств
	types := []topology.DeviceType{
		topology.DeviceTypeRouter,
		topology.DeviceTypeSwitch,
		topology.DeviceTypeHost,
		topology.DeviceTypeUnknown,
	}

	for _, dt := range types {
		fmt.Printf("Тип устройства: %s\n", dt)
	}

	// Output:
	// Тип устройства: router
	// Тип устройства: switch
	// Тип устройства: host
	// Тип устройства: unknown
}
