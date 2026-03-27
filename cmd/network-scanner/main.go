package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"network-scanner/internal/display"
	"network-scanner/internal/network"
	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"
)

func main() {
	var (
		networkCIDR   = flag.String("network", "", "CIDR сеть (например, 192.168.1.0/24)")
		portRange     = flag.String("ports", "1-1000", "Диапазон TCP портов")
		timeout       = flag.Int("timeout", 2, "Таймаут TCP/UDP в секундах")
		threads       = flag.Int("threads", 50, "Количество потоков")
		showClosed    = flag.Bool("show-closed", false, "Показывать закрытые порты")
		scanUDP       = flag.Bool("udp", false, "Включить UDP-сканирование")
		topologyMode  = flag.Bool("topology", false, "Построить топологию сети")
		outputFormat  = flag.String("output-format", "", "Формат вывода топологии: graphml|png|svg|json")
		outputFile    = flag.String("output-file", "", "Файл для сохранения топологии")
		snmpCommunity = flag.String("snmp-community", "public", "SNMP community (через запятую)")
		snmpTimeout   = flag.Int("snmp-timeout", 2, "Таймаут SNMP в секундах")
	)
	flag.Parse()

	cidr := strings.TrimSpace(*networkCIDR)
	if cidr == "" {
		auto, err := network.DetectLocalNetwork()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Не удалось определить локальную сеть: %v\n", err)
			os.Exit(1)
		}
		cidr = auto
	}

	fmt.Printf("Сканирование сети: %s\n", cidr)
	ns := scanner.NewNetworkScanner(cidr, time.Duration(*timeout)*time.Second, *portRange, *threads, *showClosed)
	ns.SetScanUDP(*scanUDP)
	ns.Scan()
	results := ns.GetResults()

	display.DisplayResults(results)
	display.DisplayAnalytics(results)

	if !*topologyMode {
		return
	}

	communities := splitCommunities(*snmpCommunity)
	snmpData, report, err := snmpcollector.CollectWithReport(results, communities, *snmpTimeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Сбор SNMP-данных завершился ошибкой: %v\n", err)
	}
	printSNMPReport(report)

	topo, err := topology.BuildTopology(results, snmpData)
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
