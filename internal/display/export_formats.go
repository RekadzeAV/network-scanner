package display

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"network-scanner/internal/scanner"
)

// export_formats.go — экспорт результатов сканирования в машиночитаемые форматы.
// Только stdlib: внешних зависимостей и сетевых загрузок нет.

// FormulaPrefixes — префиксы, с которых Excel/LibreOffice может интерпретировать
// значение ячейки CSV как формулу (CSV formula injection). Такие значения
// экранируются префиксом "'".
const FormulaPrefixes = "=+-@\t\r"

// sanitizeCSVCell нейтрализует formula injection в ячейке CSV:
// значение, начинающееся с =, +, -, @, TAB или CR, получает префикс "'",
// а сами символы CR/LF заменяются на пробелы.
func sanitizeCSVCell(v string) string {
	if v == "" {
		return v
	}
	if strings.ContainsAny(v[:1], FormulaPrefixes) {
		v = "'" + v
	}
	v = strings.ReplaceAll(v, "\r", " ")
	v = strings.ReplaceAll(v, "\n", " ")
	return v
}

// FormatResultsAsCSV сериализует результаты в CSV (RFC 4180, ';'-разделитель
// для совместимости с русскоязычным Excel). Ячейки защищены от formula
// injection. Одна строка — устройство; открытые порты перечислены в колонке
// ports в виде "tcp/80:http,tcp/443:https".
func FormatResultsAsCSV(results []scanner.Result) ([]byte, error) {
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	w.Comma = ';'

	header := []string{
		"ip", "mac", "hostname", "device_type", "vendor", "os_guess",
		"os_confidence", "alive", "snmp", "protocols", "ports",
	}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("запись заголовка CSV: %w", err)
	}

	for _, r := range results {
		ports := make([]string, 0, len(r.Ports))
		for _, p := range r.Ports {
			if p.State != "open" {
				continue
			}
			proto := p.Protocol
			if proto == "" {
				proto = "tcp"
			}
			entry := proto + "/" + strconv.Itoa(p.Port)
			if p.Service != "" {
				entry += ":" + p.Service
			}
			ports = append(ports, entry)
		}

		row := []string{
			sanitizeCSVCell(r.IP),
			sanitizeCSVCell(r.MAC),
			sanitizeCSVCell(r.Hostname),
			sanitizeCSVCell(r.DeviceType),
			sanitizeCSVCell(r.DeviceVendor),
			sanitizeCSVCell(r.GuessOS),
			sanitizeCSVCell(r.GuessOSConfidence),
			strconv.FormatBool(r.IsAlive),
			strconv.FormatBool(r.SNMPEnabled),
			sanitizeCSVCell(strings.Join(r.Protocols, ",")),
			sanitizeCSVCell(strings.Join(ports, ",")),
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("запись строки CSV (ip=%s): %w", r.IP, err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("формирование CSV: %w", err)
	}
	// UTF-8 BOM: чтобы Excel корректно открыл кириллицу без импорта.
	return append([]byte{0xEF, 0xBB, 0xBF}, sb.String()...), nil
}

// exportResultJSON — JSON-представление устройства со стабильными тегами.
type exportResultJSON struct {
	IP                string           `json:"ip"`
	MAC               string           `json:"mac,omitempty"`
	Hostname          string           `json:"hostname,omitempty"`
	DeviceType        string           `json:"device_type,omitempty"`
	DeviceVendor      string           `json:"device_vendor,omitempty"`
	GuessOS           string           `json:"os_guess,omitempty"`
	GuessOSConfidence string           `json:"os_confidence,omitempty"`
	Alive             bool             `json:"alive"`
	SNMPEnabled       bool             `json:"snmp_enabled,omitempty"`
	Protocols         []string         `json:"protocols,omitempty"`
	Ports             []exportPortJSON `json:"ports,omitempty"`
}

// exportPortJSON — JSON-представление порта.
type exportPortJSON struct {
	Port     int    `json:"port"`
	State    string `json:"state"`
	Protocol string `json:"protocol"`
	Service  string `json:"service,omitempty"`
	Version  string `json:"version,omitempty"`
	Banner   string `json:"banner,omitempty"`
}

// FormatResultsAsJSON сериализует результаты в JSON (UTF-8, отступы).
func FormatResultsAsJSON(results []scanner.Result) ([]byte, error) {
	out := make([]exportResultJSON, 0, len(results))
	for _, r := range results {
		er := exportResultJSON{
			IP:                r.IP,
			MAC:               r.MAC,
			Hostname:          r.Hostname,
			DeviceType:        r.DeviceType,
			DeviceVendor:      r.DeviceVendor,
			GuessOS:           r.GuessOS,
			GuessOSConfidence: r.GuessOSConfidence,
			Alive:             r.IsAlive,
			SNMPEnabled:       r.SNMPEnabled,
			Protocols:         r.Protocols,
			Ports:             make([]exportPortJSON, 0, len(r.Ports)),
		}
		for _, p := range r.Ports {
			er.Ports = append(er.Ports, exportPortJSON{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Version:  p.Version,
				Banner:   p.Banner,
			})
		}
		out = append(out, er)
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("сериализация JSON: %w", err)
	}
	return b, nil
}

// FormatResultsAsMarkdown формирует markdown-таблицу результатов.
func FormatResultsAsMarkdown(results []scanner.Result) string {
	if len(results) == 0 {
		return "Результаты сканирования не найдены\n"
	}
	var sb strings.Builder
	sb.WriteString("# Результаты сканирования сети\n\n")
	sb.WriteString("| IP | MAC | Hostname | Тип | ОС (оценка) | Открытые порты |\n")
	sb.WriteString("|---|---|---|---|---|---|\n")
	for _, r := range results {
		open := make([]string, 0, len(r.Ports))
		for _, p := range r.Ports {
			if p.State == "open" {
				proto := p.Protocol
				if proto == "" {
					proto = "tcp"
				}
				open = append(open, fmt.Sprintf("%s/%d", proto, p.Port))
			}
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			mdEscape(r.IP), mdEscape(r.MAC), mdEscape(r.Hostname),
			mdEscape(r.DeviceType), mdEscape(r.GuessOS), strings.Join(open, ", ")))
	}
	return sb.String()
}

// mdEscape экранирует служебные символы markdown и запрещает перевод строки.
func mdEscape(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	replacer := strings.NewReplacer("|", "\\|", "`", "\\`", "*", "\\*", "_", "\\_")
	return replacer.Replace(s)
}
