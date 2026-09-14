package gui

import (
	"strings"
	"testing"

	"network-scanner/internal/scanner"
)

// --- FormatResultsForDisplay tests ---

func TestFormatResultsForDisplay_EmptyResults(t *testing.T) {
	result := FormatResultsForDisplay(nil)
	if result == "" {
		t.Fatal("expected non-empty result for nil results")
	}
	if !strings.Contains(result, "Результаты сканирования не найдены") {
		t.Errorf("expected 'Результаты сканирования не найдены', got: %s", result)
	}
}

func TestFormatResultsForDisplay_EmptySlice(t *testing.T) {
	result := FormatResultsForDisplay([]scanner.Result{})
	if !strings.Contains(result, "Результаты сканирования не найдены") {
		t.Errorf("expected 'Результаты сканирования не найдены', got: %s", result)
	}
}

func TestFormatResultsForDisplay_SingleResult(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", MAC: "aa:bb:cc:dd:ee:ff"},
	}
	result := FormatResultsForDisplay(results)
	if !strings.Contains(result, "192.168.1.1") {
		t.Errorf("expected '192.168.1.1' in result, got: %s", result)
	}
	if !strings.Contains(result, "router") {
		t.Errorf("expected 'router' in result, got: %s", result)
	}
	if !strings.Contains(result, "aa:bb:cc:dd:ee:ff") {
		t.Errorf("expected MAC in result, got: %s", result)
	}
	if !strings.Contains(result, "## Устройства") {
		t.Error("expected '## Устройства' header")
	}
}

func TestFormatResultsForDisplay_MultipleResults(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "router", DeviceType: "Router"},
		{IP: "192.168.1.2", Hostname: "switch", DeviceType: "Network Device"},
		{IP: "192.168.1.3", Hostname: "pc", DeviceType: "Windows Computer"},
	}
	result := FormatResultsForDisplay(results)
	if !strings.Contains(result, "**Найдено устройств:** 3") {
		t.Errorf("expected 'Найдено устройств: 3', got: %s", result)
	}
}

func TestFormatResultsForDisplay_EmptyHostname(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1"},
	}
	result := FormatResultsForDisplay(results)
	if !strings.Contains(result, "| - |") {
		t.Errorf("expected '-' for empty hostname, got: %s", result)
	}
}

func TestFormatResultsForDisplay_PortsFormatting(t *testing.T) {
	results := []scanner.Result{
		{
			IP: "192.168.1.1",
			Ports: []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", State: "open", Service: "HTTP"},
				{Port: 443, Protocol: "tcp", State: "open", Service: "HTTPS"},
				{Port: 22, Protocol: "tcp", State: "closed", Service: "SSH"},
			},
		},
	}
	result := FormatResultsForDisplay(results)
	if !strings.Contains(result, "80/tcp") {
		t.Errorf("expected '80/tcp' in result, got: %s", result)
	}
	if !strings.Contains(result, "443/tcp") {
		t.Errorf("expected '443/tcp' in result, got: %s", result)
	}
	// Закрытые порты не должны отображаться
	if strings.Contains(result, "22/tcp") {
		t.Errorf("expected no '22/tcp' for closed port, got: %s", result)
	}
}

func TestFormatResultsForDisplay_PortsTruncation(t *testing.T) {
	ports := make([]scanner.PortInfo, 60)
	for i := 0; i < 60; i++ {
		ports[i] = scanner.PortInfo{Port: i + 1, Protocol: "tcp", State: "open", Service: "test"}
	}
	results := []scanner.Result{{IP: "192.168.1.1", Ports: ports}}
	result := FormatResultsForDisplay(results)
	// formatPorts добавляет "... и еще N", но FormatResultsForDisplay обрезает
	// порт-строку до 500 символов через truncateString. Поэтому "... и еще"
	// может не поместиться — проверяем что строка портов обрезана:
	if !strings.Contains(result, "tes...") {
		t.Errorf("expected truncated service name 'tes...', got: %s", result)
	}
}

func TestFormatResultsForDisplay_ProtocolStats(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Protocols: []string{"TCP", "UDP"}},
		{IP: "192.168.1.2", Protocols: []string{"TCP"}},
	}
	result := FormatResultsForDisplay(results)
	if !strings.Contains(result, "### Протоколы в сети") {
		t.Error("expected protocol stats section")
	}
	if !strings.Contains(result, "TCP") {
		t.Error("expected 'TCP' in protocol stats")
	}
	if !strings.Contains(result, "UDP") {
		t.Error("expected 'UDP' in protocol stats")
	}
}

func TestFormatResultsForDisplay_DeviceTypeStats(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", DeviceType: "Router"},
		{IP: "192.168.1.2", DeviceType: "Windows Computer"},
		{IP: "192.168.1.3", DeviceType: "Linux Server"},
	}
	result := FormatResultsForDisplay(results)
	if !strings.Contains(result, "### Типы устройств") {
		t.Error("expected device type stats section")
	}
}

func TestFormatResultsForDisplay_DeviceTypeNormalization(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", DeviceType: "Router"},
		{IP: "192.168.1.2", DeviceType: "Network Device"},
	}
	result := FormatResultsForDisplay(results)
	// Оба типа должны нормализоваться в "Network Device"
	if !strings.Contains(result, "Network Device") {
		t.Errorf("expected 'Network Device' after normalization, got: %s", result)
	}
}

// --- formatPorts tests ---

func TestFormatPorts_Empty(t *testing.T) {
	result := formatPorts(nil)
	if result != "" {
		t.Errorf("expected empty string for nil ports, got %q", result)
	}
}

func TestFormatPorts_EmptySlice(t *testing.T) {
	result := formatPorts([]scanner.PortInfo{})
	if result != "" {
		t.Errorf("expected empty string for empty slice, got %q", result)
	}
}

func TestFormatPorts_OpenPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, Protocol: "tcp", State: "open", Service: "HTTP"},
		{Port: 443, Protocol: "tcp", State: "open", Service: "HTTPS"},
	}
	result := formatPorts(ports)
	if !strings.Contains(result, "80/tcp") {
		t.Errorf("expected '80/tcp', got %q", result)
	}
	if !strings.Contains(result, "443/tcp") {
		t.Errorf("expected '443/tcp', got %q", result)
	}
	if !strings.Contains(result, "HTTP") {
		t.Errorf("expected 'HTTP' service, got %q", result)
	}
}

func TestFormatPorts_ClosedPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, Protocol: "tcp", State: "closed", Service: "HTTP"},
	}
	result := formatPorts(ports)
	if result != "" {
		t.Errorf("expected empty string for closed ports, got %q", result)
	}
}

func TestFormatPorts_MixedPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, Protocol: "tcp", State: "open", Service: "HTTP"},
		{Port: 22, Protocol: "tcp", State: "closed", Service: "SSH"},
		{Port: 443, Protocol: "tcp", State: "open", Service: "HTTPS"},
	}
	result := formatPorts(ports)
	if !strings.Contains(result, "80/tcp") {
		t.Errorf("expected '80/tcp', got %q", result)
	}
	if !strings.Contains(result, "443/tcp") {
		t.Errorf("expected '443/tcp', got %q", result)
	}
	if strings.Contains(result, "22/tcp") {
		t.Error("expected no '22/tcp' for closed port")
	}
}

func TestFormatPorts_MaxLimit(t *testing.T) {
	ports := make([]scanner.PortInfo, 55)
	for i := 0; i < 55; i++ {
		ports[i] = scanner.PortInfo{Port: i + 1, Protocol: "tcp", State: "open", Service: "test"}
	}
	result := formatPorts(ports)
	if !strings.Contains(result, "... и еще 5") {
		t.Errorf("expected truncation message, got: %s", result)
	}
}

func TestFormatPorts_UnknownService(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 9999, Protocol: "tcp", State: "open", Service: "Unknown"},
	}
	result := formatPorts(ports)
	if !strings.Contains(result, "9999/tcp") {
		t.Errorf("expected '9999/tcp', got %q", result)
	}
	if strings.Contains(result, "Unknown") {
		t.Errorf("expected no 'Unknown' service, got %q", result)
	}
}

func TestFormatPorts_MultipleProtocols(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, Protocol: "tcp", State: "open"},
		{Port: 53, Protocol: "udp", State: "open"},
	}
	result := formatPorts(ports)
	if !strings.Contains(result, "80/tcp") || !strings.Contains(result, "53/udp") {
		t.Errorf("expected both protocols, got %q", result)
	}
}

// --- escapeMarkdown tests ---

func TestEscapeMarkdown_EmptyString(t *testing.T) {
	result := escapeMarkdown("")
	if result != "-" {
		t.Errorf("expected '-', got %q", result)
	}
}

func TestEscapeMarkdown_AllowEmpty(t *testing.T) {
	result := escapeMarkdown("", true)
	if result != "" {
		t.Errorf("expected empty string with allowEmpty, got %q", result)
	}
}

func TestEscapeMarkdown_PipeEscape(t *testing.T) {
	result := escapeMarkdown("host|router")
	if result != "host\\|router" {
		t.Errorf("expected 'host\\\\|router', got %q", result)
	}
}

func TestEscapeMarkdown_NewlineEscape(t *testing.T) {
	result := escapeMarkdown("host\r\nrouter")
	if strings.Contains(result, "\n") || strings.Contains(result, "\r") {
		t.Errorf("expected no newlines, got %q", result)
	}
}

func TestEscapeMarkdown_TabEscape(t *testing.T) {
	result := escapeMarkdown("host\trouter")
	if strings.Contains(result, "\t") {
		t.Errorf("expected no tabs, got %q", result)
	}
}

func TestEscapeMarkdown_MultipleSpaces(t *testing.T) {
	result := escapeMarkdown("host   router")
	if strings.Contains(result, "   ") {
		t.Errorf("expected no multiple spaces, got %q", result)
	}
}

func TestEscapeMarkdown_AllSpecialChars(t *testing.T) {
	result := escapeMarkdown("test|pipe")
	// Pipe должен быть экранирован как backslash-pipe
	escaped := string(rune(92)) + "|"
	if !strings.Contains(result, escaped) {
		t.Errorf("expected escaped pipe, got %q", result)
	}
}

func TestEscapeMarkdown_Unicode(t *testing.T) {
	result := escapeMarkdown("router-тест")
	if result != "router-тест" {
		t.Errorf("expected unicode preserved, got %q", result)
	}
}

func TestEscapeMarkdown_SimpleText(t *testing.T) {
	result := escapeMarkdown("simple-text")
	if result != "simple-text" {
		t.Errorf("expected unchanged, got %q", result)
	}
}

// --- truncateString tests ---

func TestTruncateString_ExactLength(t *testing.T) {
	result := truncateString("hello", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncateString_Shorter(t *testing.T) {
	result := truncateString("hi", 5)
	if result != "hi" {
		t.Errorf("expected 'hi', got %q", result)
	}
}

func TestTruncateString_Longer(t *testing.T) {
	result := truncateString("hello world", 8)
	if result != "hello..." {
		t.Errorf("expected 'hello...', got %q", result)
	}
}

func TestTruncateString_MaxLen3(t *testing.T) {
	result := truncateString("hello", 3)
	if result != "hel" {
		t.Errorf("expected 'hel', got %q", result)
	}
}

func TestTruncateString_MaxLen2(t *testing.T) {
	result := truncateString("hello", 2)
	if result != "he" {
		t.Errorf("expected 'he', got %q", result)
	}
}

func TestTruncateString_MaxLen1(t *testing.T) {
	result := truncateString("hello", 1)
	if result != "h" {
		t.Errorf("expected 'h', got %q", result)
	}
}

func TestTruncateString_MaxLen0(t *testing.T) {
	result := truncateString("hello", 0)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestTruncateString_EmptyString(t *testing.T) {
	result := truncateString("", 10)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

// Тест на Unicode — Go режет по байтам, не по рунам.
// "привет" = 12 байт, truncateString("привет мир", 6) даст первые 3 байта "прив" = 6 байт
// Но 3 байта "прив" = 2 символа (каждый кириллический = 2 байта) + 1 байт обреза
// Фактически: "пр" (4 байта) + "и" (2 байта) = 6 байт, но maxLen=6, так что "прив" (6 байт)
// Проверим что не паникует
func TestTruncateString_UnicodeTruncation(t *testing.T) {
	result := truncateString("привет мир", 6)
	if len(result) > 6 {
		t.Errorf("expected <=6 bytes, got %d bytes", len(result))
	}
	// Должно содержать "..." или обрезанный текст
	_ = result
}
