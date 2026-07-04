package audit

import (
	"fmt"
	"sort"
	"strings"

	"network-scanner/internal/scanner"
)

// Finding описывает риск по открытому порту.
type Finding struct {
	Host           string
	Port           int
	Protocol       string
	Severity       string
	Title          string
	Recommendation string
}

// Summary агрегирует результаты аудита.
type Summary struct {
	TotalFindings    int
	UniqueHosts      int
	BySeverity       map[string]int
	ByHost           map[string]int
	HighestSeverity  string
	OverallRiskScore int
}

var riskyPorts = map[int]struct {
	severity string
	title    string
	rec      string
}{
	21:    {severity: "medium", title: "FTP без шифрования", rec: "Перейти на SFTP/FTPS и ограничить доступ по ACL."},
	23:    {severity: "high", title: "Telnet без шифрования", rec: "Отключить Telnet, использовать SSH с ключами."},
	139:   {severity: "high", title: "SMB/NetBIOS доступен", rec: "Ограничить SMB по сегментам и выключить наружный доступ."},
	445:   {severity: "high", title: "SMB доступен", rec: "Ограничить SMB внутренней сетью, обновить ОС и политики."},
	3389:  {severity: "high", title: "RDP доступен", rec: "Разрешать только через VPN/Zero Trust, включить MFA."},
	5900:  {severity: "high", title: "VNC доступен", rec: "Ограничить доступ, включить шифрование/туннель."},
	6379:  {severity: "high", title: "Redis доступен", rec: "Запретить внешний доступ, включить аутентификацию и TLS."},
	9200:  {severity: "high", title: "Elasticsearch доступен", rec: "Ограничить внешний доступ, включить auth и TLS."},
	27017: {severity: "high", title: "MongoDB доступен", rec: "Ограничить доступ по IP, включить auth и TLS."},
	11211: {severity: "high", title: "Memcached доступен", rec: "Запретить внешний доступ, ограничить firewall."},
}

// EvaluateOpenPorts строит список рисков по результатам сканирования.
func EvaluateOpenPorts(results []scanner.Result) []Finding {
	out := make([]Finding, 0)
	for _, host := range results {
		for _, p := range host.Ports {
			if !strings.EqualFold(p.State, "open") {
				continue
			}
			rule, ok := riskyPorts[p.Port]
			if !ok {
				continue
			}
			out = append(out, Finding{
				Host:           strings.TrimSpace(host.IP),
				Port:           p.Port,
				Protocol:       strings.TrimSpace(p.Protocol),
				Severity:       rule.severity,
				Title:          rule.title,
				Recommendation: rule.rec,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Severity != out[j].Severity {
			return severityWeight(out[i].Severity) > severityWeight(out[j].Severity)
		}
		if out[i].Host != out[j].Host {
			return out[i].Host < out[j].Host
		}
		return out[i].Port < out[j].Port
	})
	return out
}

func FormatFindings(findings []Finding) string {
	if len(findings) == 0 {
		return "Аудит портов: рисков по базовым правилам не найдено."
	}
	summary := BuildSummary(findings)
	var sb strings.Builder
	sb.WriteString("Аудит открытых портов (базовые правила):\n")
	sb.WriteString(fmt.Sprintf("- Всего находок: %d\n", summary.TotalFindings))
	sb.WriteString(fmt.Sprintf("- Хостов с рисками: %d\n", summary.UniqueHosts))
	if summary.HighestSeverity != "" {
		sb.WriteString(fmt.Sprintf("- Максимальная критичность: %s\n", strings.ToUpper(summary.HighestSeverity)))
	}
	sb.WriteString(fmt.Sprintf("- Risk score: %d\n", summary.OverallRiskScore))
	sb.WriteString("- По критичности:")
	for _, sev := range []string{"critical", "high", "medium", "low"} {
		if c := summary.BySeverity[sev]; c > 0 {
			sb.WriteString(fmt.Sprintf(" %s=%d", strings.ToUpper(sev), c))
		}
	}
	sb.WriteString("\n")
	sb.WriteString("- Хосты с рисками:\n")
	for _, hv := range sortedHosts(summary.ByHost) {
		sb.WriteString(fmt.Sprintf("  - %s: %d\n", hv.host, hv.count))
	}
	sb.WriteString("Найденные риски:\n")
	for _, f := range findings {
		sb.WriteString(fmt.Sprintf("- [%s] %s %d/%s: %s. Рекомендация: %s\n",
			strings.ToUpper(f.Severity), f.Host, f.Port, strings.ToLower(f.Protocol), f.Title, f.Recommendation))
	}
	return strings.TrimSpace(sb.String())
}

func BuildSummary(findings []Finding) Summary {
	s := Summary{
		TotalFindings:   len(findings),
		BySeverity:      map[string]int{},
		ByHost:          map[string]int{},
		HighestSeverity: "",
	}
	maxWeight := 0
	for _, f := range findings {
		sev := strings.ToLower(strings.TrimSpace(f.Severity))
		host := strings.TrimSpace(f.Host)
		s.BySeverity[sev]++
		s.ByHost[host]++
		w := severityWeight(sev)
		s.OverallRiskScore += w
		if w > maxWeight {
			maxWeight = w
			s.HighestSeverity = sev
		}
	}
	s.UniqueHosts = len(s.ByHost)
	return s
}

// HumanReadable преобразует техническую находку в сообщение для нетехнического пользователя.
func HumanReadable(f Finding) string {
	host := strings.TrimSpace(f.Host)
	if host == "" {
		host = "неизвестном устройстве"
	} else {
		host = "устройстве " + host
	}
	switch strings.TrimSpace(strings.ToLower(f.Title)) {
	case "telnet без шифрования":
		return fmt.Sprintf("На %s обнаружен Telnet: пароли могут передаваться без шифрования.", host)
	case "ftp без шифрования":
		return fmt.Sprintf("На %s открыт FTP без защиты: данные и пароли могут быть перехвачены.", host)
	case "smb доступен", "smb/netbios доступен":
		return fmt.Sprintf("На %s доступен SMB: убедитесь, что доступ ограничен только доверенной сетью.", host)
	case "rdp доступен":
		return fmt.Sprintf("На %s открыт RDP: рекомендуется доступ только через VPN и с MFA.", host)
	default:
		portPart := ""
		if f.Port > 0 {
			portPart = fmt.Sprintf(" (порт %d/%s)", f.Port, strings.ToLower(strings.TrimSpace(f.Protocol)))
		}
		title := strings.TrimSpace(f.Title)
		if title == "" {
			title = "Обнаружен потенциальный риск"
		}
		return fmt.Sprintf("На %s обнаружена проблема: %s%s.", host, title, portPart)
	}
}

// SecurityIndexFromSeverityCounts возвращает индекс безопасности 0..100.
// Формула: 100 - сумма весов рисков (critical=30, high=20, medium=10, low=5).
func SecurityIndexFromSeverityCounts(counts map[string]int) int {
	if counts == nil {
		return 100
	}
	score := 100
	score -= counts["critical"] * 30
	score -= counts["high"] * 20
	score -= counts["medium"] * 10
	score -= counts["low"] * 5
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// NormalizeSeverity нормализует строковое значение уровня критичности.
// Поддерживаются: all, critical, high, medium, low.
func NormalizeSeverity(v string) (string, bool) {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "all", "critical", "high", "medium", "low":
		return s, true
	default:
		return "", false
	}
}

// FilterByMinSeverity возвращает находки не ниже указанного порога.
// minSeverity: all|critical|high|medium|low.
func FilterByMinSeverity(findings []Finding, minSeverity string) []Finding {
	minNorm, ok := NormalizeSeverity(minSeverity)
	if !ok || minNorm == "all" {
		out := make([]Finding, len(findings))
		copy(out, findings)
		return out
	}
	minWeight := severityWeight(minNorm)
	out := make([]Finding, 0, len(findings))
	for _, f := range findings {
		if severityWeight(f.Severity) >= minWeight {
			out = append(out, f)
		}
	}
	return out
}

type hostRiskCount struct {
	host  string
	count int
}

func sortedHosts(byHost map[string]int) []hostRiskCount {
	out := make([]hostRiskCount, 0, len(byHost))
	for host, count := range byHost {
		out = append(out, hostRiskCount{host: host, count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].host < out[j].host
	})
	return out
}

func severityWeight(sev string) int {
	switch strings.ToLower(strings.TrimSpace(sev)) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
