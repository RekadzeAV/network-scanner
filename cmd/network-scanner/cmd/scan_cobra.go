package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	_ = scanCmd.Flags().SetAnnotation("network", "category", []string{"network"})
	_ = scanCmd.Flags().SetAnnotation("ports", "category", []string{"network"})
	_ = scanCmd.Flags().SetAnnotation("hosts-file", "category", []string{"network"})
	_ = scanCmd.Flags().SetAnnotation("udp", "category", []string{"scan"})
	_ = scanCmd.Flags().SetAnnotation("grab-banners", "category", []string{"scan"})
	_ = scanCmd.Flags().SetAnnotation("os-detect-active", "category", []string{"scan"})
	_ = scanCmd.Flags().SetAnnotation("timeout", "category", []string{"scan"})
	_ = scanCmd.Flags().SetAnnotation("threads", "category", []string{"scan"})
	_ = scanCmd.Flags().SetAnnotation("show-closed", "category", []string{"scan"})
	_ = scanCmd.Flags().SetAnnotation("security", "category", []string{"post-scan"})
	_ = scanCmd.Flags().SetAnnotation("topology", "category", []string{"post-scan"})
	_ = scanCmd.Flags().SetAnnotation("snmp", "category", []string{"post-scan"})
	_ = scanCmd.Flags().SetAnnotation("inventory-save", "category", []string{"post-scan"})
	_ = scanCmd.Flags().SetAnnotation("export-html", "category", []string{"export"})
	_ = scanCmd.Flags().SetAnnotation("export-xml", "category", []string{"export"})
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

	// Общие флаги inventory: путь к SQLite-базе (--db) и история (--history).
	inventoryCmd.PersistentFlags().String("db", defaultInventoryDBPath(), "Путь к inventory SQLite базе")

	inventoryListCmd.Flags().Int("limit", 10, "Максимум снапшотов (0 — без ограничения)")
	inventoryDiffCmd.Flags().Bool("history", false, "Использовать формат вывода истории (store.CompareSnapshotsByName)")

	inventorySaveCmd.Flags().String("id", "", "ID снапшота (по умолчанию auto: scan-<unix>)")
	inventorySaveCmd.Flags().String("hosts-file", "", "Файл с целями вместо живого сканирования")
	inventorySaveCmd.Flags().String("network", "", "CIDR для сканирования перед сохранением")
	inventorySaveCmd.Flags().String("ports", "1-1000", "Диапазон портов сканирования")
	inventorySaveCmd.Flags().Int("timeout", 2, "Таймаут сканирования, сек")
	inventorySaveCmd.Flags().Int("threads", 50, "Количество потоков сканирования")
}

// defaultInventoryDBPath — путь к inventory базе по умолчанию (согласован с
// scanCommandRun и services.NewInventoryService).
func defaultInventoryDBPath() string {
	return filepath.Join("inventory", "network_inventory.db")
}

// inventoryConfig собирает builder.Config с путём к базе из флага --db.
func inventoryConfig(c *cobra.Command) builder.Config {
	db, _ := c.Flags().GetString("db")
	if db == "" {
		db = defaultInventoryDBPath()
	}
	return builder.Config{
		LogLevel: "info",
		DBPath:   db,
	}
}

var inventoryListCmd = &cobra.Command{
	Use:   "list [limit]",
	Short: "Показать список снапшотов",
	Long: `Показывает снапшоты инвентаризации в порядке убывания даты.

Примеры:
  network-scanner inventory list
  network-scanner inventory list 25 --db inventory/network_inventory.db`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		limit, _ := c.Flags().GetInt("limit")
		// Позиционный аргумент имеет приоритет над флагом.
		if len(args) == 1 {
			parsed, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("некорректный limit %q: %w", args[0], err)
			}
			limit = parsed
		}
		return RunInventoryList(inventoryConfig(c), limit)
	},
}

var inventoryDiffCmd = &cobra.Command{
	Use:   "diff <idA> <idB>",
	Short: "Сравнить два снапшота",
	Long: `Сравнивает два снапшота инвентаризации и печатает новые, пропавшие и
изменившиеся хосты.

Примеры:
  network-scanner inventory diff scan-1 scan-2
  network-scanner inventory diff scan-1 scan-2 --history`,
	Args: cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		useHistory, _ := c.Flags().GetBool("history")
		cfg := inventoryConfig(c)
		if !useHistory {
			return RunInventoryDiff(cfg, args[0], args[1])
		}
		// Развёрнутый формат истории (comparator): новые/удалённые/изменённые
		// хосты и port-changes.
		return ExecuteHistory(cfg.DBPath, 0, args[0], args[1])
	},
}

var inventorySaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Сохранить снапшот",
	Long: `Сохраняет снапшот инвентаризации.

Источник данных: --hosts-file (список целей) или живое сканирование сети
(--network / автоопределение). Если указан только --id без источника, из
существующего снапшота ничего не копируется — требуется источник.

Примеры:
  network-scanner inventory save --hosts-file targets.txt --id baseline
  network-scanner inventory save --network 192.168.1.0/24 --ports 1-1000 --id home`,
	Args: cobra.NoArgs,
	RunE: func(c *cobra.Command, _ []string) error {
		cfg := inventoryConfig(c)

		hostsFile, _ := c.Flags().GetString("hosts-file")
		networkCIDR, _ := c.Flags().GetString("network")
		portRange, _ := c.Flags().GetString("ports")
		timeout, _ := c.Flags().GetInt("timeout")
		threads, _ := c.Flags().GetInt("threads")
		id, _ := c.Flags().GetString("id")

		if hostsFile == "" && networkCIDR == "" {
			return fmt.Errorf("требуется источник данных: --hosts-file или --network")
		}
		if id == "" {
			id = fmt.Sprintf("scan-%d", time.Now().Unix())
		}

		// Сканирование целей в паузе — тот же сервисный путь, что и scanCmd.
		results, err := runScanForInventory(cfg, hostsFile, networkCIDR, portRange, timeout, threads)
		if err != nil {
			return err
		}

		return RunInventorySave(cfg, results, id)
	},
}

// runScanForInventory выполняет сканирование для inventory save, переиспользуя
// сервисный слой builder (без прямого вызова display/presenter).
func runScanForInventory(cfg builder.Config, hostsFile, networkCIDR, portRange string, timeout, threads int) ([]contracts.ScanResult, error) {
	if hostsFile != "" {
		targets, err := network.ParseTargetsFromFile(hostsFile)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения файла целей: %w", err)
		}
		if networkCIDR == "" && len(targets) > 0 {
			networkCIDR = targets[0]
		}
	}
	if networkCIDR == "" {
		auto, err := network.DetectLocalNetwork()
		if err != nil {
			return nil, fmt.Errorf("не удалось определить сеть: %w", err)
		}
		networkCIDR = auto
	}

	container := builder.NewContainer(cfg)
	scannerService := container.GetScanner()

	fmt.Printf("Сканирование сети: %s\n", networkCIDR)
	results, err := scannerService.Scan(context.TODO(), contracts.ScanConfig{
		NetworkCIDR: networkCIDR,
		PortRange:   portRange,
		Timeout:     time.Duration(timeout) * time.Second,
		Threads:     threads,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("сканирование завершено ошибкой: %w", err)
	}
	return results, nil
}

// remoteExecCmd — команда удалённого выполнения
var remoteExecCmd = &cobra.Command{
	Use:   "remote-exec [flags]",
	Short: "Удалённое выполнение команд",
	Long: `Выполнение команд на удалённых хостах через SSH/WMI/WinRM.

По умолчанию включён DRY-RUN: проверяются транспорт, цель и команда против
политики (allowlist хостов/команд), но подключение к хосту не выполняется.

Для реального выполнения нужны ОБА флага:
  --execute            переключает команду из dry-run в режим выполнения
  --consent I_UNDERSTAND  явное согласие на удалённый запуск команды

Строгий TLS-канал (--require-tls): ssh использует StrictHostKeyChecking=yes,
winrm — winrs -usessl по https; транспорт wmi в строгом режиме отклоняется.

Примеры:
  network-scanner remote-exec --transport ssh --target 10.0.0.5 --command "uptime"
  network-scanner remote-exec --transport ssh --target 10.0.0.5 --command "uptime" \
      --allow-hosts 10.0.0.5 --allow-commands "uptime" --execute --consent I_UNDERSTAND
  network-scanner remote-exec --transport winrm --target host1 --command "hostname" \
      --require-tls --execute --consent I_UNDERSTAND`,
	RunE: func(c *cobra.Command, _ []string) error {
		cfg := builder.Config{LogLevel: "info", DBPath: defaultInventoryDBPath()}
		// Флаги передаются в существующий ручной парсер RunRemoteExecCLI для
		// обратной совместимости поведения.
		return RunRemoteExecCLI(cfg, flagsToArgs(c)...)
	},
}

// deviceControlCmd — команда управления устройствами
var deviceControlCmd = &cobra.Command{
	Use:   "device-control [flags]",
	Short: "Управление устройствами",
	Long: `Управление сетевыми устройствами через HTTP API (status|reboot).

Для reboot требуется --confirm I_UNDERSTAND (защита необратимого действия на
уровне CLI и сервиса devicecontrol). Рекомендуется указывать --audit-log для
журнала действий.

Примеры:
  network-scanner device-control --action status --target http://192.168.1.1
  network-scanner device-control -a reboot -T http://192.168.1.1 --confirm I_UNDERSTAND --audit-log audit.jsonl`,
	RunE: func(c *cobra.Command, _ []string) error {
		cfg := builder.Config{LogLevel: "info", DBPath: defaultInventoryDBPath()}
		return RunDeviceControl(cfg, flagsToArgs(c)...)
	},
}

func init() {
	// remote-exec: флаги зеркалят ручной парсер RunRemoteExecCLI.
	remoteExecCmd.Flags().StringP("transport", "t", "", "Транспорт: ssh|wmi|winrm")
	remoteExecCmd.Flags().StringP("target", "T", "", "Целевой хост/IP")
	remoteExecCmd.Flags().StringP("user", "u", "", "Пользователь")
	remoteExecCmd.Flags().StringP("pass", "p", "", "Пароль")
	remoteExecCmd.Flags().StringP("command", "c", "", "Команда для выполнения")
	remoteExecCmd.Flags().String("allow-hosts", "", "Список разрешённых хостов (CSV)")
	remoteExecCmd.Flags().String("allow-commands", "", "Список разрешённых команд (CSV)")
	remoteExecCmd.Flags().String("policy-file", "", "Файл политики")
	remoteExecCmd.Flags().Bool("policy-strict", false, "Строгая политика")
	remoteExecCmd.Flags().String("consent", "", "Явное согласие на выполнение: I_UNDERSTAND")
	remoteExecCmd.Flags().Bool("dry-run", true, "Только проверка политики без выполнения (по умолчанию)")
	remoteExecCmd.Flags().Bool("execute", false, "Реально выполнить команду (требует --consent I_UNDERSTAND)")
	remoteExecCmd.Flags().Int("timeout", defaultRemoteExecTimeout, "Таймаут в секундах")
	remoteExecCmd.Flags().String("audit-log", "", "Путь к audit-логу")
	// E7/7.10: строгий TLS-канал (ssh StrictHostKeyChecking=yes / winrm -usessl https).
	remoteExecCmd.Flags().Bool("require-tls", false, "Строгий TLS-канал: ssh host key, winrm https; wmi не поддерживает")
	remoteExecCmd.Flags().Bool("strict-tls", false, "Синоним --require-tls")
	remoteExecCmd.MarkFlagsMutuallyExclusive("dry-run", "execute")

	// device-control: флаги зеркалят ручной парсер RunDeviceControl.
	deviceControlCmd.Flags().StringP("action", "a", "", "Действие: status|reboot")
	deviceControlCmd.Flags().StringP("target", "T", "", "HTTP(S) endpoint устройства")
	deviceControlCmd.Flags().String("vendor", "generic-http", "Провайдер: generic-http|tp-link-http")
	deviceControlCmd.Flags().StringP("user", "u", "", "Username")
	deviceControlCmd.Flags().StringP("pass", "p", "", "Password")
	deviceControlCmd.Flags().String("confirm", "", "Подтверждение reboot: I_UNDERSTAND")
	deviceControlCmd.Flags().Int("timeout", 10, "Таймаут в секундах")
	deviceControlCmd.Flags().String("audit-log", "", "Путь к audit-логу (JSONL)")
}

// flagsToArgs конвертирует установленные флаги cobra в argv-совместимый срез
// для ручных парсеров (RunRemoteExecCLI / RunDeviceControl). Передаются только
// флаги, реально указанные пользователем.
//
// Булевы флаги сериализуются в форме --name=value, чтобы парсер различал
// "--dry-run" и "--dry-run=false" (см. P0-2: dry-run по умолчанию).
func flagsToArgs(c *cobra.Command) []string {
	args := make([]string, 0)
	c.Flags().Visit(func(f *pflag.Flag) {
		if f.Value.Type() == "bool" {
			args = append(args, "--"+f.Name+"="+f.Value.String())
			return
		}
		args = append(args, "--"+f.Name, f.Value.String())
	})
	return args
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
