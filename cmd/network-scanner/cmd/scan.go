package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"network-scanner/internal/builder"
	"network-scanner/internal/contracts"
	"network-scanner/internal/display"
	"network-scanner/internal/metrics"
	"network-scanner/internal/network"
	"network-scanner/internal/presenter"
	"network-scanner/internal/snmpcollector"
)

// hostsFromFileOptions — параметры чтения целей из файла (E7/7.7).
type hostsFromFileOptions struct {
	// Format — явно заданный формат: auto|csv|txt|json|targets.
	// Пусто/"auto" — автоопределение: .csv/.json → HostListImporter,
	// остальные расширения → расширенный target-парсер (IP/CIDR/range).
	Format string
	// MaxEntries — лимит числа хостов после раскрытия CIDR (0 → по умолчанию).
	MaxEntries int
}

// loadTargetsFromFile читает список целей с поддержкой двух парсеров:
//
//   - "targets" (по умолчанию для .txt и файлов без расширения) —
//     внутренний формат: IP / CIDR / диапазон `192.168.1.1-10` / комментарии `#`;
//   - "csv"/"json" (или auto для .csv/.json) — расширенный импорт через
//     HostListImporter: колонки IP,Hostname,Comment (CSV) либо JSON-массив
//     HostEntry с раскрытием CIDR и лимитом количества хостов.
//
// Возвращает плоский список IP-адресов (совместим с прежним поведением
// ParseTargetsFromFile) и текстовое описание применённого формата.
func loadTargetsFromFile(path string, opts hostsFromFileOptions) ([]string, string, error) {
	format := strings.ToLower(strings.TrimSpace(opts.Format))
	if format == "" {
		format = "auto"
	}

	switch format {
	case "targets", "plain", "legacy":
		ips, err := network.ParseTargetsFromFile(path)
		return ips, "targets", err
	case "csv", "json", "txt":
		entries, err := importHostEntries(path, format, opts.MaxEntries)
		if err != nil {
			return nil, "", err
		}
		return hostEntriesToIPs(entries), format, nil
	case "auto":
		if ext := strings.ToLower(filepath.Ext(path)); ext == ".csv" || ext == ".json" {
			entries, err := importHostEntries(path, "auto", opts.MaxEntries)
			if err != nil {
				return nil, "", err
			}
			return hostEntriesToIPs(entries), strings.TrimPrefix(ext, "."), nil
		}
		ips, err := network.ParseTargetsFromFile(path)
		return ips, "targets", err
	default:
		return nil, "", fmt.Errorf("неподдерживаемый формат списка хостов: %s (ожидается auto|csv|txt|json|targets)", opts.Format)
	}
}

// importHostEntries читает файл через расширенный HostListImporter.
func importHostEntries(path, format string, maxEntries int) ([]network.HostEntry, error) {
	importer := network.NewHostListImporter(maxEntries)
	numeric, err := hostListFormatValue(format)
	if err != nil {
		return nil, err
	}
	return importer.ImportFromFile(path, numeric)
}

// hostListFormatValue переводит строковый формат в значение HostListFormat.
func hostListFormatValue(format string) (network.HostListFormat, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "auto":
		return network.FormatAuto, nil
	case "csv":
		return network.FormatCSV, nil
	case "txt":
		return network.FormatTXT, nil
	case "json":
		return network.FormatJSON, nil
	default:
		return network.FormatAuto, fmt.Errorf("неподдерживаемый формат импорта: %s", format)
	}
}

// hostEntriesToIPs разворачивает записи импортера в плоский список IP.
func hostEntriesToIPs(entries []network.HostEntry) []string {
	ips := make([]string, 0, len(entries))
	for _, e := range entries {
		if ip := strings.TrimSpace(e.IP); ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

// parsePortSpec разбирает список портов в формате "53,161" или "1-1024".
// Возвращает плоский список номеров портов (в порядке возрастания).
func parsePortSpec(spec string) ([]int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}
	var out []int
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("неверный порт в диапазоне %q: %w", part, err)
			}
			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("неверный порт в диапазоне %q: %w", part, err)
			}
			if start > end {
				start, end = end, start
			}
			if start < 1 || end > 65535 {
				return nil, fmt.Errorf("диапазон %q вне 1..65535", part)
			}
			for p := start; p <= end; p++ {
				out = append(out, p)
			}
			continue
		}
		port, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("неверный порт %q: %w", part, err)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("порт %d вне 1..65535", port)
		}
		out = append(out, port)
	}
	return out, nil
}

// RunScan запускает сканирование через сервис
func RunScan(cfg builder.Config, args ...string) error {
	// Метрики (E7/7.8): флаги разбираются и в legacy-пути.
	metricsOpts := parseMetricsArgs(args)

	// Парсинг флагов
	networkCIDR := ""
	portRange := "1-1000"
	timeout := 2
	threads := 50
	showClosed := false
	scanUDP := false
	udpPortsSpec := ""
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
	hostsFormat := ""
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
		case "--udp-ports":
			if i+1 < len(args) {
				udpPortsSpec = args[i+1]
				i++
			}
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
		case "--hosts-format":
			if i+1 < len(args) {
				hostsFormat = args[i+1]
				i++
			}
		case "--export-html":
			exportHTML = true
		case "--export-xml":
			exportXML = true
		}
	}

	// Цели: автоопределение сети выполняется только если нет ни --network, ни
	// файла целей — иначе указанные в файле адреса игнорировались бы (E7/7.7).
	var targetList []string
	if networkCIDR == "" && hostsFile == "" {
		auto, err := network.DetectLocalNetwork()
		if err != nil {
			return fmt.Errorf("не удалось определить сеть: %w", err)
		}
		networkCIDR = auto
	}

	if hostsFile != "" {
		// Чтение целей из файла: расширенный импорт (CSV/JSON/TXT) или
		// прежний target-формат — см. loadTargetsFromFile (E7/7.7).
		fmt.Printf("Чтение целей из файла: %s\n", hostsFile)
		ips, usedFormat, loadErr := loadTargetsFromFile(hostsFile, hostsFromFileOptions{Format: hostsFormat})
		if loadErr != nil {
			return fmt.Errorf("ошибка чтения файла целей: %w", loadErr)
		}
		targetList = ips
		fmt.Printf("Найдено %d целей в файле (формат: %s)\n", len(targetList), usedFormat)

		if len(targetList) > 0 && strings.TrimSpace(networkCIDR) == "" {
			if strings.Contains(targetList[0], "/") {
				networkCIDR = targetList[0]
			} else {
				networkCIDR = targetList[0] + "/32"
			}
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

	// Метрики (E7/7.8): opt-in, адрес по умолчанию — loopback.
	if metricsOpts.Enabled {
		cfg.MetricsAddr = metricsOpts.Addr
		reg := metrics.NewRegistry()
		subscribeScanMetrics(container, reg)
		if stop, mErr := startMetricsServer(reg, cfg); mErr == nil {
			defer stop()
			fmt.Printf("Метрики: http://%s/metrics\n", formatAddrForLog(cfg.MetricsAddr))
		} else {
			fmt.Printf("Метрики недоступны: %v\n", mErr)
		}
	}

	scannerService := container.GetScanner()

	// Запуск сканирования
	fmt.Printf("Сканирование сети: %s\n", networkCIDR)

	// UDP-порты (E7/7.2): явный список через --udp-ports, иначе дефолт.
	udpPorts, err := parsePortSpec(udpPortsSpec)
	if err != nil {
		return fmt.Errorf("--udp-ports: %w", err)
	}
	if scanUDP && len(udpPorts) > 0 {
		fmt.Printf("UDP порты: %v\n", udpPorts)
	}

	results, err := scannerService.Scan(context.TODO(), contracts.ScanConfig{
		NetworkCIDR: networkCIDR,
		PortRange:   portRange,
		Timeout:     time.Duration(timeout) * time.Second,
		Threads:     threads,
		ShowClosed:  showClosed,
		ScanUDP:     scanUDP,
		UDPPorts:    udpPorts,
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
