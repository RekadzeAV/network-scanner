package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"network-scanner/internal/builder"
	"network-scanner/internal/contracts"
	"network-scanner/internal/display"
	"network-scanner/internal/network"
	"network-scanner/internal/presenter"
	"network-scanner/internal/snmpcollector"
)

// RunScan запускает сканирование через сервис
func RunScan(cfg builder.Config, args ...string) error {
	// Парсинг флагов
	networkCIDR := ""
	portRange := "1-1000"
	timeout := 2
	threads := 50
	showClosed := false
	scanUDP := false
	grabBanners := false
	osDetectActive := false
	verboseLogs := false
	runSecurity := false
	runTopology := false
	runInventorySave := false
	inventoryID := ""
	runSNMP := false
	snmptCommunity := "public"
	snmptTimeout := 2
	hostsFile := ""
	exportHTML := false
	exportXML := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--network", "-n":
			if i+1 < len(args) {
				networkCIDR = args[i+1]
				i++
			}
		case "--ports", "-p":
			if i+1 < len(args) {
				portRange = args[i+1]
				i++
			}
		case "--timeout", "-t":
			if i+1 < len(args) {
				_, _ = fmt.Sscanf(args[i+1], "%d", &timeout)
				i++
			}
		case "--threads":
			if i+1 < len(args) {
				_, _ = fmt.Sscanf(args[i+1], "%d", &threads)
				i++
			}
		case "--show-closed":
			showClosed = true
		case "--udp":
			scanUDP = true
		case "--grab-banners":
			grabBanners = true
		case "--os-detect-active":
			osDetectActive = true
		case "--verbose-port-logs":
			verboseLogs = true
		case "--security":
			runSecurity = true
		case "--topology":
			runTopology = true
		case "--inventory-save":
			runInventorySave = true
		case "--inventory-id":
			if i+1 < len(args) {
				inventoryID = args[i+1]
				i++
			}
		case "--snmp":
			runSNMP = true
		case "--snmp-community":
			if i+1 < len(args) {
				snmptCommunity = args[i+1]
				i++
			}
		case "--snmp-timeout":
			if i+1 < len(args) {
				_, _ = fmt.Sscanf(args[i+1], "%d", &snmptTimeout)
				i++
			}
		case "--hosts-file":
			if i+1 < len(args) {
				hostsFile = args[i+1]
				i++
			}
		case "--export-html":
			exportHTML = true
		case "--export-xml":
			exportXML = true
		}
	}

	// Автоопределение сети или чтение из файла
	var targets []string
	var err error

	if hostsFile != "" {
		// Чтение целей из файла
		fmt.Printf("Чтение целей из файла: %s\n", hostsFile)
		targets, err = network.ParseTargetsFromFile(hostsFile)
		if err != nil {
			return fmt.Errorf("ошибка чтения файла целей: %w", err)
		}
		fmt.Printf("Найдено %d целей в файле\n", len(targets))

		// Если networkCIDR не указан, используем первый CIDR из файла
		if networkCIDR == "" && len(targets) > 0 {
			// Проверяем, есть ли CIDR в файле
			networkCIDR = targets[0] // Используем первый IP как точку отсчёта
		}
	}

	if networkCIDR == "" {
		auto, err := network.DetectLocalNetwork()
		if err != nil {
			return fmt.Errorf("не удалось определить сеть: %w", err)
		}
		networkCIDR = auto
	}

	// Создание контейнера и сервиса
	container := builder.NewContainer(cfg)
	scannerService := container.GetScanner()

	// Запуск сканирования
	fmt.Printf("Сканирование сети: %s\n", networkCIDR)

	results, err := scannerService.Scan(context.TODO(), contracts.ScanConfig{
		NetworkCIDR: networkCIDR,
		PortRange:   portRange,
		Timeout:     time.Duration(timeout) * time.Second,
		Threads:     threads,
		ShowClosed:  showClosed,
		ScanUDP:     scanUDP,
		GrabBanners: grabBanners,
		OSActive:    osDetectActive,
		VerboseLogs: verboseLogs,
	}, func(stage string, current, total int, message string) {
		fmt.Printf("[%s] %s: %d/%d\n", stage, message, current, total)
	})
	if err != nil {
		return fmt.Errorf("сканирование завершено ошибкой: %w", err)
	}

	// Вывод результатов
	internalResults := ConvertToInternalResults(results)
	display.SetShowRawBanners(false)
	display.DisplayResults(internalResults)
	display.DisplayAnalytics(internalResults)

	// Экспорт в HTML/XML
	if exportHTML {
		fmt.Println("\nЭкспорт в HTML...")
		err := presenter.HTMLPresenter{}.Export(internalResults, "html")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка экспорта HTML: %v\n", err)
		}
	}

	if exportXML {
		fmt.Println("\nЭкспорт в XML...")
		err := presenter.XMLPresenter{}.Export(internalResults, "xml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка экспорта XML: %v\n", err)
		}
	}

	// Дополнительные операции
	if runSecurity {
		if err := RunSecurity(cfg, results); err != nil {
			fmt.Fprintf(os.Stderr, "Security error: %v\n", err)
		}
	}

	if runTopology {
		if err := RunTopology(cfg, results, "public", 2); err != nil {
			fmt.Fprintf(os.Stderr, "Topology error: %v\n", err)
		}
	}

	if runInventorySave {
		if inventoryID == "" {
			inventoryID = fmt.Sprintf("scan-%d", time.Now().Unix())
		}
		if err := RunInventorySave(cfg, results, inventoryID); err != nil {
			fmt.Fprintf(os.Stderr, "Inventory save error: %v\n", err)
		}
	}

	if runSNMP {
		fmt.Println("SNMP опрос устройств...")
		communities := []string{snmptCommunity}
		devices := convertToScannerResults(results)
		snmpDevices, report, err := snmpcollector.CollectWithReport(devices, communities, snmptTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SNMP error: %v\n", err)
		} else {
			fmt.Printf("SNMP опрос завершён: подключено %d/%d устройств\n", report.Connected, report.TotalSNMPTargets)
			if len(snmpDevices) > 0 {
				fmt.Printf("Получено SNMP данных: %d устройств\n", len(snmpDevices))
			}
			if len(report.Failures) > 0 {
				fmt.Printf("Ошибки SNMP: %d\n", len(report.Failures))
			}
		}
	}

	return nil
}

// ExecuteCLI выполняет CLI команду.
//
// Делегирует в cobra rootCmd (root.go) — единый слой диспатча для scan,
// inventory, remote-exec, device-control, gui, version. Легаси-ручной switch
// удалён как дублирующий слой (дедупликация E6): он не поддерживал флаги/справку
// и вызывал inventory save с пустым набором результатов. Подробная справка
// доступна через cobra: `network-scanner --help`, `network-scanner <cmd> --help`.
func ExecuteCLI() {
	Execute()
}

// ExecuteLegacyScan поддерживает вызов без подкоманды с флагами сканирования:
//
//	network-scanner --network 192.168.1.0/24 --ports 1-1000
//
// Форма сохранена для обратной совместимости: она используется smoke- и
// closure-скриптами проекта (smoke-cli-no-topology, smoke-cli-topology,
// p2-closure-check) и примерами в документации.
func ExecuteLegacyScan(args []string) {
	cfg := builder.Config{
		LogLevel: "info",
		DBPath:   filepath.Join("inventory", "network_inventory.db"),
	}
	if err := RunScan(cfg, args...); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}
