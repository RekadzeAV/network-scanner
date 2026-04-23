package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"network-scanner/internal/audit"
	"network-scanner/internal/cve"
	"network-scanner/internal/devicecontrol"
	"network-scanner/internal/display"
	"network-scanner/internal/nettools"
	"network-scanner/internal/network"
	"network-scanner/internal/presenter"
	"network-scanner/internal/remoteexec"
	"network-scanner/internal/report"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"
	"network-scanner/internal/wol"
)

const (
	maxScanThreads                   = 512
	largeSubnetWarnHost              = 512
	securityReportUnsafeConsentToken = "I_UNDERSTAND_UNREDACTED_REPORT"
)

func main() {
	var (
		networkCIDR                 = flag.String("network", "", "CIDR сеть (например, 192.168.1.0/24)")
		portRange                   = flag.String("ports", "1-1000", "Диапазон TCP портов")
		timeout                     = flag.Int("timeout", 2, "Таймаут TCP/UDP в секундах")
		threads                     = flag.Int("threads", 50, "Количество потоков")
		showClosed                  = flag.Bool("show-closed", false, "Показывать закрытые порты")
		scanUDP                     = flag.Bool("udp", false, "Включить UDP-сканирование")
		topologyMode                = flag.Bool("topology", false, "Построить топологию сети")
		outputFormat                = flag.String("output-format", "", "Формат вывода топологии: graphml|png|svg|json")
		outputFile                  = flag.String("output-file", "", "Файл для сохранения топологии")
		snmpCommunity               = flag.String("snmp-community", "public", "SNMP community (через запятую)")
		snmpTimeout                 = flag.Int("snmp-timeout", 2, "Таймаут SNMP в секундах")
		pingTarget                  = flag.String("ping", "", "Выполнить ping для хоста/IP и завершить")
		traceTarget                 = flag.String("traceroute", "", "Выполнить traceroute для хоста/IP и завершить")
		dnsQuery                    = flag.String("dns", "", "Выполнить DNS lookup (A/AAAA или PTR) и завершить")
		dnsServer                   = flag.String("dns-server", "", "DNS сервер для --dns (IP или IP:port, например 1.1.1.1:53)")
		pingCount                   = flag.Int("ping-count", 4, "Количество ICMP пакетов для --ping")
		toolTimeout                 = flag.Int("tool-timeout", 60, "Таймаут инструментов --ping/--traceroute/--dns в секундах")
		traceMaxHops                = flag.Int("traceroute-max-hops", 30, "Максимальное число hop для --traceroute (1..64)")
		rawOutput                   = flag.Bool("raw", false, "Для режимов --ping/--traceroute/--dns печатать полный сырой вывод тоже")
		whoisQuery                  = flag.String("whois", "", "Выполнить whois для домена/IP и завершить")
		wifiInfo                    = flag.Bool("wifi", false, "Показать Wi-Fi информацию текущей ОС и завершить")
		grabBanners                 = flag.Bool("grab-banners", false, "Читать баннеры с типовых TCP-портов (медленнее)")
		showRawBanners              = flag.Bool("show-raw-banners", false, "Показывать сырой banner в выводе портов (по умолчанию скрыт)")
		osDetectActive              = flag.Bool("os-detect-active", false, "Включить расширенные (active) эвристики определения ОС")
		verbosePortLogs             = flag.Bool("verbose-port-logs", false, "Включить детальные debug-логи по каждому порту (шумно, медленнее)")
		wolMAC                      = flag.String("wol-mac", "", "Отправить Wake-on-LAN magic packet на MAC и завершить")
		wolBroadcast                = flag.String("wol-broadcast", "", "Broadcast адрес для --wol-mac (например 192.168.1.255:9)")
		wolIface                    = flag.String("wol-iface", "", "Сетевой интерфейс для --wol-mac (если --wol-broadcast не задан)")
		auditOpenPorts              = flag.Bool("audit-open-ports", false, "После сканирования выполнить базовый аудит открытых портов")
		auditMinSeverity            = flag.String("audit-min-severity", "low", "Минимальная критичность для --audit-open-ports: all|low|medium|high|critical")
		riskSignatures              = flag.Bool("risk-signatures", false, "После сканирования выполнить сигнатуры домашних рисков")
		deviceAction                = flag.String("device-action", "", "Управление оборудованием: status|reboot (требует --device-target)")
		deviceTarget                = flag.String("device-target", "", "HTTP(S) endpoint устройства (например http://192.168.1.1)")
		deviceVendor                = flag.String("device-vendor", "generic-http", "Провайдер/вендор управления устройством")
		deviceUser                  = flag.String("device-user", "", "Username для --device-action")
		devicePass                  = flag.String("device-pass", "", "Password для --device-action")
		deviceConfirm               = flag.String("device-confirm", "", "Подтверждение опасной операции: I_UNDERSTAND")
		deviceTimeout               = flag.Int("device-timeout", 10, "Таймаут device-control в секундах")
		auditLogPath                = flag.String("audit-log", filepath.Join("audit", "device-actions.log"), "Путь к audit log JSONL")
		enableCVE                   = flag.Bool("cve", false, "После сканирования выполнить базовое CVE сопоставление по баннерам")
		cveMinCVSS                  = flag.Float64("cve-min-cvss", 0, "Минимальный CVSS для вывода CVE (0..10)")
		cveMaxAgeDays               = flag.Int("cve-max-age-days", 0, "Максимальный возраст CVE в днях (0 = без фильтра)")
		securityReport              = flag.String("security-report-file", "", "Путь к HTML security report (например report-security.html) или auto")
		securityReportRedact        = flag.Bool("security-report-redact", true, "Маскировать чувствительные данные в security report")
		securityReportUnsafeConsent = flag.String("security-report-unsafe-consent", "", "Подтверждение для --security-report-redact=false")
		remoteTransport             = flag.String("remote-exec-transport", "", "Remote exec transport: ssh|wmi|winrm")
		remoteTarget                = flag.String("remote-exec-target", "", "Целевой хост/IP для remote exec")
		remoteUser                  = flag.String("remote-exec-user", "", "Пользователь для remote exec")
		remotePass                  = flag.String("remote-exec-pass", "", "Пароль для remote exec (если поддерживается транспортом)")
		remoteCommand               = flag.String("remote-exec-command", "", "Команда для удаленного выполнения")
		remoteAllowHosts            = flag.String("remote-exec-allow-hosts", "", "Разрешенные хосты через запятую (обязательно)")
		remoteAllowCommands         = flag.String("remote-exec-allow-commands", "", "Разрешенные команды через запятую (обязательно)")
		remotePolicyFile            = flag.String("remote-exec-policy-file", "", "JSON policy файл allowlist (allow_hosts, allow_commands)")
		remotePolicyStrict          = flag.Bool("remote-exec-policy-strict", false, "Строгий режим policy: только --remote-exec-policy-file, без inline allowlist")
		remoteConsent               = flag.String("remote-exec-consent", "", "Подтверждение remote exec: I_UNDERSTAND")
		remoteDryRun                = flag.Bool("remote-exec-dry-run", true, "Проверить policy без реального выполнения")
		remoteTimeout               = flag.Int("remote-exec-timeout", 15, "Таймаут remote exec в секундах")
		remoteAuditPath             = flag.String("remote-exec-audit-log", filepath.Join("audit", "remote-exec.log"), "Путь к audit log remote exec JSONL")
	)
	flag.Parse()

	if runToolsMode(*pingTarget, *traceTarget, *dnsQuery, *dnsServer, *pingCount, *toolTimeout, *traceMaxHops, *rawOutput, *whoisQuery, *wifiInfo, *wolMAC, *wolBroadcast, *wolIface, *deviceAction, *deviceTarget, *deviceVendor, *deviceUser, *devicePass, *deviceConfirm, *deviceTimeout, *auditLogPath, *remoteTransport, *remoteTarget, *remoteUser, *remotePass, *remoteCommand, *remoteAllowHosts, *remoteAllowCommands, *remotePolicyFile, *remotePolicyStrict, *remoteConsent, *remoteDryRun, *remoteTimeout, *remoteAuditPath) {
		return
	}

	cidr := strings.TrimSpace(*networkCIDR)
	if cidr == "" {
		auto, err := network.DetectLocalNetwork()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Не удалось определить локальную сеть: %v\n", err)
			os.Exit(1)
		}
		cidr = auto
	}
	if *threads < 1 {
		fmt.Fprintf(os.Stderr, "Предупреждение: --threads < 1, используется 1\n")
		*threads = 1
	}
	if *threads > maxScanThreads {
		fmt.Fprintf(os.Stderr, "Предупреждение: --threads слишком велик, используется %d\n", maxScanThreads)
		*threads = maxScanThreads
	}
	if hosts, err := network.EstimateHostCount(cidr); err == nil && hosts >= largeSubnetWarnHost {
		fmt.Fprintf(os.Stderr, "Предупреждение: крупная подсеть %s (~%d хостов). Сканирование может занять много времени.\n", cidr, hosts)
	}

	fmt.Printf("Сканирование сети: %s\n", cidr)
	display.SetShowRawBanners(*showRawBanners)
	ns := scanner.NewNetworkScanner(cidr, time.Duration(*timeout)*time.Second, *portRange, *threads, *showClosed)
	ns.SetScanUDP(*scanUDP)
	ns.SetGrabBanners(*grabBanners)
	ns.SetOSDetectActive(*osDetectActive)
	ns.SetVerbosePortLogs(*verbosePortLogs)
	ns.Scan()
	results := ns.GetResults()

	cliPresenter := presenter.CLIPresenter{}
	cliPresenter.DisplayHeader()
	display.DisplayResults(results)
	display.DisplayAnalytics(results)
	if *auditOpenPorts {
		findings := audit.EvaluateOpenPorts(results)
		if norm, ok := audit.NormalizeSeverity(*auditMinSeverity); ok {
			findings = audit.FilterByMinSeverity(findings, norm)
		} else {
			fmt.Fprintf(os.Stderr, "Предупреждение: неизвестный --audit-min-severity=%q, используется low\n", *auditMinSeverity)
			findings = audit.FilterByMinSeverity(findings, "low")
		}
		fmt.Println()
		fmt.Println(audit.FormatFindings(findings))
	}
	if *enableCVE || *riskSignatures || strings.TrimSpace(*securityReport) != "" {
		matches := cve.AnalyzeResults(results, cve.NewDefaultCatalog(), cve.Options{
			MinCVSS:    clampCVSS(*cveMinCVSS),
			MaxAgeDays: *cveMaxAgeDays,
		})
		riskFindings := make([]risksignature.Finding, 0)
		riskDBVersion := ""
		if *riskSignatures || strings.TrimSpace(*securityReport) != "" {
			if db, err := risksignature.LoadDefault(); err == nil {
				riskFindings = risksignature.Evaluate(results, db)
				riskDBVersion = strings.TrimSpace(db.Version)
			}
		}
		fmt.Println()
		fmt.Println(cve.FormatMatches(matches))
		if *riskSignatures {
			printRiskFindings(riskFindings, riskDBVersion)
		}
		if reportPath := strings.TrimSpace(*securityReport); reportPath != "" {
			if err := validateSecurityReportRedaction(*securityReportRedact, *securityReportUnsafeConsent); err != nil {
				fmt.Fprintf(os.Stderr, "Security report ошибка: %v\n", err)
				os.Exit(1)
			}
			if !*securityReportRedact {
				fmt.Fprintln(os.Stderr, "WARNING: security report redaction disabled; output may contain sensitive data.")
			}
			generationMode := "manual"
			if strings.EqualFold(reportPath, "auto") {
				generationMode = "auto"
			}
			reportID := buildSecurityReportID()
			reportPath = resolveSecurityReportPath(reportPath, *securityReportRedact, reportID)
			if err := report.SaveSecurityHTMLWithRiskOptions(reportPath, results, matches, riskFindings, time.Now(), report.Options{
				RedactSensitive: *securityReportRedact,
				PolicyVersion:   "v1",
				UnsafeConsent:   strings.TrimSpace(*securityReportUnsafeConsent) == securityReportUnsafeConsentToken,
				GenerationMode:  generationMode,
				ReportID:        reportID,
			}); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка сохранения security report: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Security report сохранен: %s (report-id=%s)\n", reportPath, reportID)
		}
	}

	if !*topologyMode {
		return
	}

	communities := splitCommunities(*snmpCommunity)
	snmpData, report, err := snmpcollector.CollectWithReport(results, communities, *snmpTimeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Сбор SNMP-данных завершился ошибкой: %v\n", err)
	}
	printSNMPReport(report)

	topo, err := topology.BuildTopologyWithOptions(results, snmpData, topology.BuildOptions{
		PartialSNMPKeys: partialSNMPKeysFromReport(report),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка построения топологии: %v\n", err)
		os.Exit(1)
	}

	if strings.TrimSpace(*outputFormat) == "" {
		fmt.Println("\nТопология (кратко):")
		for _, l := range topo.Links {
			fmt.Printf("- %s (%s) <-> %s (%s)\n", displayName(l.Source), portName(l.SourcePort), displayName(l.Target), portName(l.TargetPort))
		}
		return
	}

	format := strings.ToLower(strings.TrimSpace(*outputFormat))
	out := strings.TrimSpace(*outputFile)
	if out == "" {
		switch format {
		case "json":
			out = "topology.json"
		case "graphml":
			out = "topology.graphml"
		case "png":
			out = "topology.png"
		case "svg":
			out = "topology.svg"
		default:
			fmt.Fprintf(os.Stderr, "Неподдерживаемый формат: %s\n", format)
			os.Exit(1)
		}
	}

	switch format {
	case "json":
		err = topo.SaveJSON(out)
	case "graphml":
		err = topo.SaveGraphML(out)
	case "png", "svg":
		err = topo.RenderWithGraphviz(format, out)
		if err != nil && errors.Is(err, topology.ErrGraphvizNotInstalled) {
			fallbackFile := strings.TrimSuffix(out, "."+format) + ".json"
			if saveErr := topo.SaveJSON(fallbackFile); saveErr != nil {
				err = fmt.Errorf("%w; также не удалось сохранить fallback JSON: %v", err, saveErr)
				break
			}
			fmt.Fprintf(os.Stderr, "Graphviz недоступен, сохранен fallback JSON: %s\n", fallbackFile)
			err = nil
		}
	default:
		fmt.Fprintf(os.Stderr, "Неподдерживаемый формат: %s\n", format)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка сохранения топологии: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Топология сохранена: %s\n", out)
}

func clampCVSS(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 10 {
		return 10
	}
	return v
}

func validateSecurityReportRedaction(redactEnabled bool, unsafeConsent string) error {
	if redactEnabled {
		return nil
	}
	if strings.TrimSpace(unsafeConsent) != securityReportUnsafeConsentToken {
		return fmt.Errorf("для --security-report-redact=false требуется --security-report-unsafe-consent %s", securityReportUnsafeConsentToken)
	}
	return nil
}

func resolveSecurityReportPath(path string, redactEnabled bool, reportID string) string {
	if !strings.EqualFold(strings.TrimSpace(path), "auto") {
		return strings.TrimSpace(path)
	}
	reportID = strings.TrimSpace(reportID)
	if reportID == "" {
		reportID = "unknown"
	}
	if redactEnabled {
		return fmt.Sprintf("security-report-redacted-%s.html", reportID)
	}
	return fmt.Sprintf("security-report-unredacted-%s.html", reportID)
}

func buildSecurityReportID() string {
	return time.Now().UTC().Format("20060102T150405Z")
}

func partialSNMPKeysFromReport(report *snmpcollector.CollectReport) map[string]struct{} {
	if report == nil {
		return nil
	}
	keys := make(map[string]struct{})
	for _, s := range report.DeviceSummaries {
		if strings.TrimSpace(s.QueryErrors) == "" {
			continue
		}
		ip := strings.TrimSpace(strings.ToLower(s.IP))
		if ip == "" {
			continue
		}
		keys["ip:"+ip] = struct{}{}
	}
	if len(keys) == 0 {
		return nil
	}
	return keys
}

func runToolsMode(
	pingTarget, traceTarget, dnsQuery, dnsServer string,
	pingCount, toolTimeoutSec, traceMaxHops int, rawOutput bool,
	whoisQuery string, wifiInfo bool,
	wolMAC, wolBroadcast, wolIface string,
	deviceAction, deviceTarget, deviceVendor, deviceUser, devicePass, deviceConfirm string,
	deviceTimeoutSec int,
	auditLogPath string,
	remoteTransport, remoteTarget, remoteUser, remotePass, remoteCommand, remoteAllowHosts, remoteAllowCommands, remotePolicyFile string,
	remotePolicyStrict bool,
	remoteConsent string,
	remoteDryRun bool,
	remoteTimeoutSec int,
	remoteAuditPath string,
) bool {
	pingTarget = strings.TrimSpace(pingTarget)
	traceTarget = strings.TrimSpace(traceTarget)
	dnsQuery = strings.TrimSpace(dnsQuery)
	dnsServer = strings.TrimSpace(dnsServer)
	whoisQuery = strings.TrimSpace(whoisQuery)
	wolMAC = strings.TrimSpace(wolMAC)
	wolBroadcast = strings.TrimSpace(wolBroadcast)
	wolIface = strings.TrimSpace(wolIface)
	deviceAction = strings.TrimSpace(deviceAction)
	deviceTarget = strings.TrimSpace(deviceTarget)
	deviceVendor = strings.TrimSpace(deviceVendor)
	deviceUser = strings.TrimSpace(deviceUser)
	devicePass = strings.TrimSpace(devicePass)
	deviceConfirm = strings.TrimSpace(deviceConfirm)
	auditLogPath = strings.TrimSpace(auditLogPath)
	remoteTransport = strings.TrimSpace(remoteTransport)
	remoteTarget = strings.TrimSpace(remoteTarget)
	remoteUser = strings.TrimSpace(remoteUser)
	remotePass = strings.TrimSpace(remotePass)
	remoteCommand = strings.TrimSpace(remoteCommand)
	remoteAllowHosts = strings.TrimSpace(remoteAllowHosts)
	remoteAllowCommands = strings.TrimSpace(remoteAllowCommands)
	remotePolicyFile = strings.TrimSpace(remotePolicyFile)
	remoteConsent = strings.TrimSpace(remoteConsent)
	remoteAuditPath = strings.TrimSpace(remoteAuditPath)
	if pingTarget == "" && traceTarget == "" && dnsQuery == "" && whoisQuery == "" && !wifiInfo && wolMAC == "" && deviceAction == "" && remoteTransport == "" {
		return false
	}
	if pingCount < 1 {
		pingCount = 1
	}
	if pingCount > 50 {
		pingCount = 50
	}
	if toolTimeoutSec <= 0 {
		toolTimeoutSec = 60
	}
	if traceMaxHops <= 0 {
		traceMaxHops = 30
	}
	if traceMaxHops > 64 {
		traceMaxHops = 64
	}
	toolTimeout := time.Duration(toolTimeoutSec) * time.Second
	ctx := context.Background()
	if pingTarget != "" {
		res, err := nettools.RunPingStructured(ctx, pingTarget, pingCount, toolTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ping ошибка: %s\n", nettools.HumanizeToolError(err))
			os.Exit(1)
		}
		fmt.Printf("Ping: %s\n", pingTarget)
		fmt.Printf("- Count=%d Timeout=%ds\n", pingCount, toolTimeoutSec)
		fmt.Printf("- Sent=%d Received=%d Loss=%.1f%%\n", res.Stats.Sent, res.Stats.Received, res.Stats.PacketLoss)
		if res.Stats.RTTAvg > 0 {
			fmt.Printf("- RTT min/avg/max: %s / %s / %s\n", res.Stats.RTTMin, res.Stats.RTTAvg, res.Stats.RTTMax)
		}
		if rawOutput {
			fmt.Println("\n--- raw ping output ---")
			fmt.Println(res.RawOutput)
		}
	}
	if traceTarget != "" {
		res, err := nettools.RunTracerouteStructuredWithMaxHops(ctx, traceTarget, toolTimeout, traceMaxHops)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Traceroute ошибка: %s\n", nettools.HumanizeToolError(err))
			os.Exit(1)
		}
		fmt.Printf("Traceroute: %s\n", traceTarget)
		fmt.Printf("- Timeout=%ds MaxHops=%d\n", toolTimeoutSec, traceMaxHops)
		for _, hop := range res.Hops {
			addr := hop.Address
			if strings.TrimSpace(addr) == "" {
				addr = "*"
			}
			if hop.Measurements > 0 {
				fmt.Printf("- hop %d: %s (min/avg/max %s/%s/%s)\n", hop.Index, addr, hop.RTTMin, hop.RTTAvg, hop.RTTMax)
			} else {
				fmt.Printf("- hop %d: %s\n", hop.Index, addr)
			}
		}
		if rawOutput {
			fmt.Println("\n--- raw traceroute output ---")
			fmt.Println(res.RawOutput)
		}
	}
	if dnsQuery != "" {
		dnsCtx, cancel := context.WithTimeout(ctx, toolTimeout)
		res, err := nettools.LookupDNSWithResolver(dnsCtx, dnsQuery, dnsServer)
		cancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "DNS ошибка: %s\n", nettools.HumanizeToolError(err))
			os.Exit(1)
		}
		fmt.Printf("DNS: %s\n", dnsQuery)
		fmt.Printf("- Timeout=%ds\n", toolTimeoutSec)
		if dnsServer != "" {
			fmt.Printf("- Resolver: %s\n", dnsServer)
		}
		for _, ip := range res.ForwardIPs {
			fmt.Printf("- A/AAAA: %s\n", strings.TrimSpace(ip))
		}
		for _, name := range res.ReverseNames {
			fmt.Printf("- PTR: %s\n", strings.TrimSpace(name))
		}
		if len(res.ForwardIPs) == 0 && len(res.ReverseNames) == 0 {
			fmt.Println("- Пустой ответ")
		}
		if rawOutput {
			fmt.Println("\n--- raw dns output ---")
			fmt.Printf("query=%s\nforward=%s\nreverse=%s\n",
				strings.TrimSpace(res.Query),
				strings.Join(res.ForwardIPs, ", "),
				strings.Join(res.ReverseNames, ", "))
		}
	}
	if whoisQuery != "" {
		res, err := nettools.RunWhois(ctx, whoisQuery, toolTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Whois ошибка: %s\n", nettools.HumanizeToolError(err))
			os.Exit(1)
		}
		fmt.Printf("Whois: %s\n", whoisQuery)
		fmt.Printf("- Timeout=%ds\n", toolTimeoutSec)
		fmt.Println(res)
	}
	if wifiInfo {
		res, err := nettools.GetWiFiInfo(ctx, toolTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Wi-Fi ошибка: %s\n", nettools.HumanizeToolError(err))
			os.Exit(1)
		}
		fmt.Println("Wi-Fi:")
		fmt.Printf("- Timeout=%ds\n", toolTimeoutSec)
		fmt.Println(res)
	}
	if wolMAC != "" {
		target, err := wol.SendMagicPacketWithInterface(wolMAC, wolBroadcast, wolIface)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WOL ошибка: %v\n", err)
			os.Exit(1)
		}
		if wolIface != "" {
			fmt.Printf("WOL: отправлен magic packet на MAC %s через %s (iface=%s)\n", wolMAC, target, wolIface)
		} else {
			fmt.Printf("WOL: отправлен magic packet на MAC %s через %s\n", wolMAC, target)
		}
	}
	if deviceAction != "" {
		if deviceTarget == "" {
			fmt.Fprintln(os.Stderr, "Device control ошибка: требуется --device-target")
			os.Exit(1)
		}
		if strings.EqualFold(deviceAction, devicecontrol.ActionReboot) && deviceConfirm != "I_UNDERSTAND" {
			fmt.Fprintln(os.Stderr, "Device control: для reboot требуется --device-confirm I_UNDERSTAND")
			os.Exit(1)
		}
		req := devicecontrol.Request{
			Action:    deviceAction,
			TargetURL: deviceTarget,
			Vendor:    deviceVendor,
			Username:  deviceUser,
			Password:  devicePass,
			Timeout:   time.Duration(deviceTimeoutSec) * time.Second,
		}
		res, err := devicecontrol.Execute(ctx, req)
		entry := devicecontrol.AuditEntry{
			Action:    req.Action,
			TargetURL: req.TargetURL,
			Vendor:    req.Vendor,
			Success:   err == nil && res.Success,
			Message:   strings.TrimSpace(res.Message),
		}
		if err != nil && strings.TrimSpace(entry.Message) == "" {
			entry.Message = err.Error()
		}
		if auditErr := devicecontrol.AppendAudit(auditLogPath, entry); auditErr != nil {
			fmt.Fprintf(os.Stderr, "Предупреждение: audit log не записан: %v\n", auditErr)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Device control ошибка: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Device control: action=%s target=%s status=%d result=%s\n", res.Action, res.TargetURL, res.StatusCode, strings.TrimSpace(res.Message))
		fmt.Printf("Audit log: %s\n", auditLogPath)
	}
	if remoteTransport != "" {
		allowHosts, allowCommands, err := resolveRemoteAllowlists(remoteAllowHosts, remoteAllowCommands, remotePolicyFile, remotePolicyStrict)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Remote exec policy ошибка: %v\n", err)
			os.Exit(1)
		}
		req := remoteexec.Request{
			Transport:     remoteTransport,
			Target:        remoteTarget,
			Username:      remoteUser,
			Password:      remotePass,
			Command:       remoteCommand,
			AllowHosts:    allowHosts,
			AllowCommands: allowCommands,
			Consent:       remoteConsent,
			DryRun:        remoteDryRun,
			Timeout:       time.Duration(remoteTimeoutSec) * time.Second,
		}
		res, err := remoteexec.Execute(ctx, req)
		entry := remoteexec.AuditEntry{
			Transport: req.Transport,
			Target:    req.Target,
			Command:   remoteexec.SanitizeText(req.Command),
			DryRun:    req.DryRun,
			Success:   err == nil && res.Success,
			Message:   strings.TrimSpace(remoteexec.SanitizeText(res.Output)),
		}
		if err != nil && strings.TrimSpace(entry.Message) == "" {
			entry.Message = remoteexec.SanitizeText(err.Error())
		}
		if auditErr := remoteexec.AppendAudit(remoteAuditPath, entry); auditErr != nil {
			fmt.Fprintf(os.Stderr, "Предупреждение: remote-exec audit log не записан: %v\n", auditErr)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Remote exec ошибка: %s\n", remoteexec.SanitizeText(err.Error()))
			os.Exit(1)
		}
		fmt.Printf("Remote exec: transport=%s target=%s dry-run=%v\n", res.Transport, res.Target, req.DryRun)
		if maskedOutput := strings.TrimSpace(remoteexec.SanitizeText(res.Output)); maskedOutput != "" {
			fmt.Printf("Output:\n%s\n", maskedOutput)
		}
		fmt.Printf("Audit log: %s\n", remoteAuditPath)
	}
	return true
}

func printRiskFindings(findings []risksignature.Finding, dbVersion string) {
	fmt.Println()
	fmt.Println("Risk signatures:")
	fmt.Printf("- DB version: %s\n", strings.TrimSpace(dbVersion))
	if len(findings) == 0 {
		fmt.Println("- Findings: none")
		return
	}
	fmt.Printf("- Findings: %d\n", len(findings))
	for _, f := range findings {
		fmt.Printf("  • [%s] %s (%s) host=%s\n", strings.TrimSpace(f.Severity), strings.TrimSpace(f.Title), strings.TrimSpace(f.SignatureID), strings.TrimSpace(f.HostIP))
		if strings.TrimSpace(f.Reason) != "" {
			fmt.Printf("    reason: %s\n", strings.TrimSpace(f.Reason))
		}
		if strings.TrimSpace(f.Recommendation) != "" {
			fmt.Printf("    recommendation: %s\n", strings.TrimSpace(f.Recommendation))
		}
		if strings.TrimSpace(f.ReferenceURL) != "" {
			fmt.Printf("    reference: %s\n", strings.TrimSpace(f.ReferenceURL))
		}
	}
}

func splitCommunities(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"public"}
	}
	return out
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func appendUnique(base []string, extra ...string) []string {
	seen := make(map[string]struct{}, len(base))
	for _, v := range base {
		key := strings.TrimSpace(v)
		if key == "" {
			continue
		}
		seen[strings.ToLower(key)] = struct{}{}
	}
	out := make([]string, 0, len(base)+len(extra))
	out = append(out, base...)
	for _, v := range extra {
		key := strings.TrimSpace(v)
		if key == "" {
			continue
		}
		lower := strings.ToLower(key)
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		out = append(out, key)
	}
	return out
}

func resolveRemoteAllowlists(inlineHosts, inlineCommands, policyFile string, strict bool) ([]string, []string, error) {
	inlineHostList := splitCSV(inlineHosts)
	inlineCommandList := splitCSV(inlineCommands)

	if strict {
		if strings.TrimSpace(policyFile) == "" {
			return nil, nil, fmt.Errorf("strict mode требует --remote-exec-policy-file")
		}
		if len(inlineHostList) > 0 || len(inlineCommandList) > 0 {
			return nil, nil, fmt.Errorf("strict mode запрещает inline allowlist флаги")
		}
	}

	allowHosts := inlineHostList
	allowCommands := inlineCommandList
	if strings.TrimSpace(policyFile) != "" {
		policy, err := remoteexec.LoadPolicy(policyFile)
		if err != nil {
			return nil, nil, err
		}
		allowHosts = appendUnique(allowHosts, policy.AllowHosts...)
		allowCommands = appendUnique(allowCommands, policy.AllowCommands...)
	}
	return allowHosts, allowCommands, nil
}

func displayName(d *topology.Device) string {
	if d == nil {
		return "unknown"
	}
	if d.Hostname != "" {
		return d.Hostname
	}
	if d.IP != "" {
		return d.IP
	}
	if d.MAC != "" {
		return d.MAC
	}
	return "unknown"
}

func portName(p *topology.Port) string {
	if p == nil {
		return "-"
	}
	if p.Name != "" {
		return p.Name
	}
	if p.Index > 0 {
		return fmt.Sprintf("if%d", p.Index)
	}
	return "-"
}

func printSNMPReport(report *snmpcollector.CollectReport) {
	if report == nil {
		return
	}
	fmt.Println("\nSNMP отчет:")
	fmt.Printf("- Целей для SNMP: %d\n", report.TotalSNMPTargets)
	fmt.Printf("- Успешных подключений: %d\n", report.Connected)
	fmt.Printf("- Частичных опросов: %d\n", report.Partial)
	fmt.Printf("- Полных отказов: %d\n", report.Failed)
	if len(report.Failures) == 0 {
		return
	}
	fmt.Println("- Детали отказов:")
	for _, f := range report.Failures {
		community := ""
		if strings.TrimSpace(f.Community) != "" {
			community = fmt.Sprintf(", community=%s", strings.TrimSpace(f.Community))
		}
		fmt.Printf("  • %s [%s%s]: %s\n", f.IP, f.Kind, community, f.Message)
	}
}
