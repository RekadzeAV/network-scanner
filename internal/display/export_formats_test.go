package display

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"network-scanner/internal/scanner"
)

func sampleResults() []scanner.Result {
	return []scanner.Result{
		{
			IP: "192.168.1.10", MAC: "AA:BB:CC:DD:EE:FF", Hostname: "pc1.local",
			DeviceType: "host", DeviceVendor: "Dell", GuessOS: "Windows",
			GuessOSConfidence: "high", IsAlive: true, SNMPEnabled: false,
			Protocols: []string{"tcp"},
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
				{Port: 443, State: "open", Protocol: "tcp", Service: "https", Version: "nginx 1.24"},
				{Port: 22, State: "closed", Protocol: "tcp"},
			},
		},
		{
			IP: "192.168.1.1", MAC: "11:22:33:44:55:66", Hostname: "router",
			DeviceType: "router", IsAlive: true,
		},
	}
}

func TestFormatResultsAsCSV(t *testing.T) {
	b, err := FormatResultsAsCSV(sampleResults())
	if err != nil {
		t.Fatalf("FormatResultsAsCSV: %v", err)
	}
	if !bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatal("CSV должен начинаться с UTF-8 BOM")
	}
	s := string(b)
	for _, want := range []string{"ip;mac;hostname", "192.168.1.10", "tcp/80:http", "tcp/443:https"} {
		if !strings.Contains(s, want) {
			t.Fatalf("CSV не содержит %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "tcp/22") {
		t.Fatal("закрытый порт не должен попадать в CSV")
	}
}

func TestFormatResultsAsCSVFormulaInjection(t *testing.T) {
	results := []scanner.Result{
		{IP: "10.0.0.1", Hostname: "=cmd|' /C calc'!A0", DeviceVendor: "+SUM(A1)"},
	}
	b, err := FormatResultsAsCSV(results)
	if err != nil {
		t.Fatalf("FormatResultsAsCSV: %v", err)
	}
	s := string(b)
	// Ячейка не должна начинаться с опасного символа сразу после разделителя.
	for _, bad := range []string{";=cmd", ";+SUM", ";-cmd", ";@cmd"} {
		if strings.Contains(s, bad) {
			t.Fatalf("formula injection не нейтрализован (%q):\n%s", bad, s)
		}
	}
	if !strings.Contains(s, "'=cmd") {
		t.Fatal("опасное значение должно быть экранировано префиксом '")
	}
	if !strings.Contains(s, "'+SUM") {
		t.Fatal("префикс + должен быть экранирован")
	}
}

func TestFormatResultsAsJSON(t *testing.T) {
	b, err := FormatResultsAsJSON(sampleResults())
	if err != nil {
		t.Fatalf("FormatResultsAsJSON: %v", err)
	}
	var parsed []map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("невалидный JSON: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("ожидалось 2 записи, получено %d", len(parsed))
	}
	first := parsed[0]
	if first["ip"] != "192.168.1.10" {
		t.Fatalf("ip = %v", first["ip"])
	}
	if first["alive"] != true {
		t.Fatalf("alive = %v", first["alive"])
	}
	ports, ok := first["ports"].([]any)
	if !ok || len(ports) != 3 {
		t.Fatalf("ожидалось 3 порта (JSON экспортирует все со state), получено %v", first["ports"])
	}
	closed := map[string]any{}
	for _, p := range ports {
		pm, _ := p.(map[string]any)
		if pm["port"] == float64(22) {
			closed = pm
		}
	}
	if closed["state"] != "closed" {
		t.Fatalf("порт 22 должен иметь state=closed, получено %v", closed)
	}
	if _, has := first["mac"]; !has {
		t.Fatal("поле mac не должно быть omitted")
	}
}

func TestFormatResultsAsJSONEmpty(t *testing.T) {
	b, err := FormatResultsAsJSON(nil)
	if err != nil {
		t.Fatalf("FormatResultsAsJSON(nil): %v", err)
	}
	if string(b) != "[]" {
		t.Fatalf("ожидался [], получено %s", b)
	}
}

func TestFormatResultsAsMarkdown(t *testing.T) {
	s := FormatResultsAsMarkdown(sampleResults())
	for _, want := range []string{"| 192.168.1.10 |", "tcp/80", "| 192.168.1.1 |"} {
		if !strings.Contains(s, want) {
			t.Fatalf("markdown не содержит %q:\n%s", want, s)
		}
	}
	if FormatResultsAsMarkdown(nil) == "" {
		t.Fatal("пустой результат должен давать сообщение")
	}
}

func TestFormatResultsAsMarkdownInjection(t *testing.T) {
	results := []scanner.Result{{IP: "1.2.3.4", Hostname: "evil | boot | x"}}
	s := FormatResultsAsMarkdown(results)
	if strings.Contains(s, "| boot") && !strings.Contains(s, "\\| boot") {
		t.Fatalf("| в данных не экранирован:\n%s", s)
	}
}
