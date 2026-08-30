package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"network-scanner/internal/builder"
	"network-scanner/internal/contracts"
	"network-scanner/internal/display"
	"network-scanner/internal/network"
	"network-scanner/internal/presenter"
	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
)

var scanCmd = &cobra.Command{
	Use:   "scan [flags]",
	Short: "Запустить сканирование сети",
	Long: `Сканирует указанную сеть или автоматически определяет локальную сеть.
Поддерживает TCP/UDP сканирование, обнаружение устройств, SNMP-опрос,
анализ безопасности и построение топологии.

Примеры:
  network-scanner scan
  network-scanner scan --network 192.168.1.0/24 --ports 1-1000
  network-scanner scan --udp --grab-banners --topology
  network-scanner scan --hosts-file targets.txt --export-html`,
	RunE: scanCommandRun,
}

func init() {
	// Определяем флаги cobra
	scanCmd.Flags().StringP("network", "n", "", "CIDR сеть (например, 192.168.1.0/24)")
	scanCmd.Flags().StringP("ports", "p", "1-1000", "Диапазон портов (по умолчанию 1-1000)")
	scanCmd.Flags().IntP("timeout", "t", 2, "Таймаут в секундах (по умолчанию 2)")
	scanCmd.Flags().Int("threads", 50, "Количество потоков (по умолчанию 50)")
	scanCmd.Flags().Bool("show-closed", false, "Показывать закрытые порты")
	scanCmd.Flags().BoolP("udp", "u", false, "Включить UDP сканирование")
	scanCmd.Flags().Bool("grab-banners", false, "Собирать баннеры")
	scanCmd.Flags().Bool("os-detect-active", false, "Активные эвристики ОС")
	scanCmd.Flags().Bool("verbose-port-logs", false, "Детальные логи по портам")
	scanCmd.Flags().Bool("security", false, "Запустить анализ безопасности после сканирования")
	scanCmd.Flags().Bool("topology", false, "Построить топологию после сканирования")
	scanCmd.Flags().Bool("inventory-save", false, "Сохранить результат в inventory")
	scanCmd.Flags().String("inventory-id", "", "ID снапшота для inventory (по умолчанию auto)")
	scanCmd.Flags().Bool("snmp", false, "Включить SNMP опрос устройств")
	scanCmd.Flags().String("snmp-community", "public", "SNMP community (по умолчанию public)")
	scanCmd.Flags().Int("snmp-timeout", 2, "Таймаут SNMP в секундах (по умолчанию 2)")
	scanCmd.Flags().String("hosts-file", "", "Файл с целями (IP, CIDR, ranges)")
	scanCmd.Flags().Bool("export-html", false, "Экспорт результатов в HTML")
	scanCmd.Flags().Bool("export-xml", false, "Экспорт результатов в XML")
	scanCmd.Flags().Bool("json", false, "Вывод результатов в JSON формате")

	// Группировка флагов
	scanCmd.Flags().SetAnnotation("network", "category", []string{"network"})
	scanCmd.Flags().SetAnnotation("ports", "category", []string{"network"})
	scanCmd.Flags().SetAnnotation("hosts-file", "category", []string{"network"})
	scanCmd.Flags().SetAnnotation("udp", "category", []string{"scan"})
	scanCmd.Flags().SetAnnotation("grab-banners", "category", []string{"scan"})
	scanCmd.Flags().SetAnnotation("os-detect-active", "category", []string{"scan"})
	scanCmd.Flags().SetAnnotation("timeout", "category", []string{"scan"})
	scanCmd.Flags().SetAnnotation("threads", "category", []string{"scan"})
	scanCmd.Flags().SetAnnotation("show-closed", "category", []string{"scan"})
	scanCmd.Flags().SetAnnotation("security", "category", []string{"post-scan"})
	scanCmd.Flags().SetAnnotation("topology", "category", []string{"post-scan"})
	scanCmd.Flags().SetAnnotation("snmp", "category", []string{"post-scan"})
	scanCmd.Flags().SetAnnotation("inventory-save", "category", []string{"post-scan"})
	scanCmd.Flags().SetAnnotation("export-html", "category", []string{"export"})
	scanCmd.Flags().SetAnnotation("export-xml", "category", []string{"export"})
}

// scanCommandRun — обработчик команды scan
func scanCommandRun(c *cobra.Command, _ []string) error {
	cfg := builder.Config{
		LogLevel: "info",
		DBPath:   filepath.Join("inventory", "network_inventory.db"),
	}
	return RunScanCobra(c, cfg)
}

// RunScanCobra запускает сканирование через cobra флаги
func RunScanCobra(c *cobra.Command, cfg builder.Config) error {
	// Парсинг флагов
	networkCIDR, _ := c.Flags().GetString("network")
	portRange, _ := c.Flags().GetString("ports")
	timeout, _ := c.Flags().GetInt("timeout")
	threads, _ := c.Flags().GetInt("threads")
	showClosed, _ := c.Flags().GetBool("show-closed")
	scanUDP, _ := c.Flags().GetBool("udp")
	grabBanners, _ := c.Flags().GetBool("grab-banners")
	osDetectActive, _ := c.Flags().GetBool("os-detect-active")
	verboseLogs, _ := c.Flags().GetBool("verbose-port-logs")
	runSecurity, _ := c.Flags().GetBool("security")
	runTopology, _ := c.Flags().GetBool("topology")
	runInventorySave, _ := c.Flags().GetBool("inventory-save")
	inventoryID, _ := c.Flags().GetString("inventory-id")
	runSNMP, _ := c.Flags().GetBool("snmp")
	snmptCommunity, _ := c.Flags().GetString("snmp-community")
	snmptTimeout, _ := c.Flags().GetInt("snmp-timeout")
	hostsFile, _ := c.Flags().GetString("hosts-file")
	exportHTML, _ := c.Flags().GetBool("export-html")
	exportXML, _ := c.Flags().GetBool("export-xml")
	jsonOutput, _ := c.Flags().GetBool("json")

	// Автоопределение сети или чтение из файла
	var targets []string
	var err error

	if hostsFile != "" {
		fmt.Printf("Чтение целей из файла: %s\n", hostsFile)
		targets, err = network.ParseTargetsFromFile(hostsFile)
		if err != nil {
			return fmt.Errorf("ошибка чтения файла целей: %w", err)
		}
		fmt.Printf("Найдено %d целей в файле\n", len(targets))

		if networkCIDR == "" && len(targets) > 0 {
			networkCIDR = targets[0]
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

	// JSON output
	if jsonOutput {
		jsonData, err := json.MarshalIndent(map[string]interface{}{
			"network":   networkCIDR,
			"hosts":     len(internalResults),
			"scan_time": time.Now().UTC().Format(time.RFC3339),
			"results":   internalResults,
		}, "", "  ")
		if err != nil {
			return fmt.Errorf("ошибка формирования JSON: %w", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}

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

// inventoryCmd — команда управления инвентаризацией
var inventoryCmd = &cobra.Command{
	Use:   "inventory [command]",
	Short: "Управление инвентаризацией устройств",
	Long:  "Управление снапшотами инвентаризации: list, diff, save.",
}

func init() {
	inventoryCmd.AddCommand(inventoryListCmd)
	inventoryCmd.AddCommand(inventoryDiffCmd)
	inventoryCmd.AddCommand(inventorySaveCmd)
}

var inventoryListCmd = &cobra.Command{
	Use:   "list [limit]",
	Short: "Показать список снапшотов",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Inventory list: реализация требует integration с builder")
	},
}

var inventoryDiffCmd = &cobra.Command{
	Use:   "diff <idA> <idB>",
	Short: "Сравнить два снапшота",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Inventory diff: реализация требует integration с builder")
	},
}

var inventorySaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Сохранить снапшот",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Inventory save: реализация требует integration с builder")
	},
}

// remoteExecCmd — команда удалённого выполнения
var remoteExecCmd = &cobra.Command{
	Use:   "remote-exec [flags]",
	Short: "Удалённое выполнение команд",
	Long:  "Выполнение команд на удалённых хостах через SSH/WMI/WinRM.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Remote exec: требуется указать параметры через флаги")
	},
}

// deviceControlCmd — команда управления устройствами
var deviceControlCmd = &cobra.Command{
	Use:   "device-control [flags]",
	Short: "Управление устройствами",
	Long:  "Управление сетевыми устройствами через HTTP API.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Device control: требуется указать параметры через флаги")
	},
}

// guiCmd — запуск GUI (только при сборке с тегом gui)
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Запустить GUI приложение",
	Long:  "Запуск графического интерфейса Network Scanner.",
	Run: func(cmd *cobra.Command, args []string) {
		RunGUI()
	},
}

// GetScanFlags возвращает pflag.FlagSet для scanCmd (для совместимости)
func GetScanFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("scan", pflag.ContinueOnError)
	flags.StringP("network", "n", "", "CIDR сеть")
	flags.StringP("ports", "p", "1-1000", "Диапазон портов")
	flags.IntP("timeout", "t", 2, "Таймаут в секундах")
	flags.Int("threads", 50, "Количество потоков")
	flags.Bool("show-closed", false, "Показывать закрытые порты")
	flags.BoolP("udp", "u", false, "UDP сканирование")
	flags.Bool("grab-banners", false, "Собирать баннеры")
	flags.Bool("os-detect-active", false, "Активные эвристики ОС")
	flags.Bool("verbose-port-logs", false, "Детальные логи")
	flags.Bool("security", false, "Анализ безопасности")
	flags.Bool("topology", false, "Построить топологию")
	flags.Bool("inventory-save", false, "Сохранить в inventory")
	flags.String("inventory-id", "", "ID снапшота")
	flags.Bool("snmp", false, "SNMP опрос")
	flags.String("snmp-community", "public", "SNMP community")
	flags.Int("snmp-timeout", 2, "Таймаут SNMP")
	flags.String("hosts-file", "", "Файл с целями")
	flags.Bool("export-html", false, "Экспорт в HTML")
	flags.Bool("export-xml", false, "Экспорт в XML")
	return flags
}
