package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"network-scanner/internal/builder"
	"network-scanner/internal/contracts"
	"network-scanner/internal/display"
	"network-scanner/internal/gui"
	"network-scanner/internal/network"
	"network-scanner/internal/presenter"
	"network-scanner/internal/scanner"
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
				fmt.Sscanf(args[i+1], "%d", &timeout)
				i++
			}
		case "--threads":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &threads)
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
				fmt.Sscanf(args[i+1], "%d", &snmptTimeout)
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

	results, err := scannerService.Scan(nil, contracts.ScanConfig{
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

// convertToScannerResults конвертирует contracts.ScanResult в scanner.Result
func convertToScannerResults(results []contracts.ScanResult) []scanner.Result {
	out := make([]scanner.Result, 0, len(results))
	for _, r := range results {
		ports := make([]scanner.PortInfo, 0, len(r.Ports))
		for _, p := range r.Ports {
			ports = append(ports, scanner.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}
		out = append(out, scanner.Result{
			IP:           r.IP,
			Hostname:     r.Hostname,
			MAC:          r.MAC,
			Ports:        ports,
			DeviceType:   r.DeviceType,
			DeviceVendor: r.DeviceVendor,
			GuessOS:      r.GuessOS,
		})
	}
	return out
}

// ExecuteCLI выполняет CLI команду
func ExecuteCLI() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := builder.Config{
		LogLevel: "info",
		DBPath:   filepath.Join("inventory", "network_inventory.db"),
	}

	switch os.Args[1] {
	case "scan":
		if err := RunScan(cfg, os.Args[2:]...); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
			os.Exit(1)
		}
	case "security":
		fmt.Println("Security: требуется результат сканирования (используйте --security в scan)")
	case "topology":
		fmt.Println("Topology: требуется результат сканирования (используйте --topology в scan)")
	case "remote-exec":
		if err := RunRemoteExecCLI(cfg, os.Args[2:]...); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
			os.Exit(1)
		}
	case "device-control":
		if err := RunDeviceControl(cfg, os.Args[2:]...); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
			os.Exit(1)
		}
	case "inventory":
		if len(os.Args) < 3 {
			fmt.Println("Inventory subcommand required: list|diff|save")
			os.Exit(1)
		}
		switch os.Args[2] {
		case "list":
			if err := RunInventoryList(cfg, 10); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
				os.Exit(1)
			}
		case "diff":
			if len(os.Args) < 5 {
				fmt.Println("Usage: network-scanner inventory diff <idA> <idB>")
				os.Exit(1)
			}
			if err := RunInventoryDiff(cfg, os.Args[3], os.Args[4]); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
				os.Exit(1)
			}
		case "save":
			var results []contracts.ScanResult
			if err := RunInventorySave(cfg, results, "manual"); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Printf("Unknown inventory subcommand: %s\n", os.Args[2])
			os.Exit(1)
		}
	case "gui":
		fmt.Println("Запуск GUI приложения...")
		a := gui.NewApp()
		a.Run()
		return
	default:
		fmt.Printf("Неизвестная команда: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Использование: network-scanner <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  scan             Запустить сканирование")
	fmt.Println("  gui              Запустить GUI приложение")
	fmt.Println("  remote-exec      Удалённое выполнение команд")
	fmt.Println("  device-control   Управление устройствами")
	fmt.Println("  inventory        Управление инвентаризацией (list|diff|save)")
	fmt.Println()
	fmt.Println("Scan options:")
	fmt.Println("  --network        CIDR сеть (например, 192.168.1.0/24)")
	fmt.Println("  --ports          Диапазон портов (по умолчанию 1-1000)")
	fmt.Println("  --timeout        Таймаут в секундах (по умолчанию 2)")
	fmt.Println("  --threads        Количество потоков (по умолчанию 50)")
	fmt.Println("  --show-closed    Показывать закрытые порты")
	fmt.Println("  --udp            Включить UDP сканирование")
	fmt.Println("  --grab-banners   Собирать баннеры")
	fmt.Println("  --os-detect-active   Активные эвристики ОС")
	fmt.Println("  --verbose-port-logs  Детальные логи по портам")
	fmt.Println("  --security       Запустить анализ безопасности после сканирования")
	fmt.Println("  --topology       Построить топологию после сканирования")
	fmt.Println("  --inventory-save Сохранить результат в inventory")
	fmt.Println("  --inventory-id   ID снапшота для inventory (по умолчанию auto)")
	fmt.Println("  --snmp           Включить SNMP опрос устройств")
	fmt.Println("  --snmp-community SNMP community (по умолчанию public)")
	fmt.Println("  --snmp-timeout   Таймаут SNMP в секундах (по умолчанию 2)")
	fmt.Println("  --hosts-file     Файл с целями (IP, CIDR, ranges)")
	fmt.Println("  --export-html    Экспорт результатов в HTML")
	fmt.Println("  --export-xml     Экспорт результатов в XML")
	fmt.Println()
	fmt.Println("Remote exec options:")
	fmt.Println("  --transport      ssh|wmi|winrm")
	fmt.Println("  --target         Целевой хост/IP")
	fmt.Println("  --user           Пользователь")
	fmt.Println("  --pass           Пароль")
	fmt.Println("  --command        Команда для выполнения")
	fmt.Println("  --dry-run        Проверить policy без выполнения")
	fmt.Println("  --consent        Подтверждение: I_UNDERSTAND")
	fmt.Println("  --timeout        Таймаут в секундах (по умолчанию 15)")
	fmt.Println()
	fmt.Println("Device control options:")
	fmt.Println("  --action         status|reboot")
	fmt.Println("  --target         HTTP(S) endpoint устройства")
	fmt.Println("  --vendor         Провайдер (по умолчанию generic-http)")
	fmt.Println("  --user           Username")
	fmt.Println("  --pass           Password")
	fmt.Println("  --confirm        Подтверждение: I_UNDERSTAND")
	fmt.Println("  --timeout        Таймаут в секундах (по умолчанию 10)")
	fmt.Println()
	fmt.Println("Inventory options:")
	fmt.Println("  list             Показать список снапшотов")
	fmt.Println("  diff <idA> <idB> Сравнить два снапшота")
	fmt.Println("  save             Сохранить снапшот")
}
