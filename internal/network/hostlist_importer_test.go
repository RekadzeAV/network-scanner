package network

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// M6.2: Тесты для HostListImporter (Импортируемые списки хостов)
// ============================================================================

// TestHostListImporter_New — ветка: создание импортера
func TestHostListImporter_New(t *testing.T) {
	importer := NewHostListImporter(1000)
	if importer == nil {
		t.Fatal("expected non-nil importer")
	}
	if importer.maxEntries != 1000 {
		t.Errorf("expected maxEntries 1000, got %d", importer.maxEntries)
	}
}

// TestHostListImporter_Default — ветка: дефолтный импортер
func TestHostListImporter_Default(t *testing.T) {
	importer := DefaultHostListImporter()
	if importer == nil {
		t.Fatal("expected non-nil importer")
	}
	if importer.maxEntries != 65536 {
		t.Errorf("expected maxEntries 65536, got %d", importer.maxEntries)
	}
}

// TestHostListImporter_ImportTXT — ветка: импорт TXT файла
func TestHostListImporter_ImportTXT(t *testing.T) {
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "hosts.txt")

	content := `192.168.1.1 router-main
192.168.1.2 switch-core # Core switch
192.168.1.10 server-web
10.0.0.1
`
	err := os.WriteFile(txtFile, []byte(content), 0644)
	if err != nil {
		t.Skipf("cannot create temp file: %v", err)
	}

	importer := DefaultHostListImporter()
	entries, err := importer.ImportFromFile(txtFile, FormatAuto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}

	// Проверяем первую запись
	if entries[0].IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", entries[0].IP)
	}
	if entries[0].Hostname != "router-main" {
		t.Errorf("expected hostname router-main, got %s", entries[0].Hostname)
	}

	// Проверяем запись с комментарием
	if entries[1].Comment != "Core switch" {
		t.Errorf("expected comment 'Core switch', got %s", entries[1].Comment)
	}
}

// TestHostListImporter_ImportTXT_Comments — ветка: комментарии в TXT
func TestHostListImporter_ImportTXT_Comments(t *testing.T) {
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "hosts.txt")

	content := `# This is a comment
192.168.1.1
// Another comment
192.168.1.2 server

`
	err := os.WriteFile(txtFile, []byte(content), 0644)
	if err != nil {
		t.Skipf("cannot create temp file: %v", err)
	}

	importer := DefaultHostListImporter()
	entries, err := importer.ImportFromFile(txtFile, FormatTXT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries (comments ignored), got %d", len(entries))
	}
}

// TestHostListImporter_ImportCSV — ветка: импорт CSV файла
func TestHostListImporter_ImportCSV(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "hosts.csv")

	content := `192.168.1.1,router-main,Main router
192.168.1.2,switch-core,Core switch
192.168.1.10,server-web
`
	err := os.WriteFile(csvFile, []byte(content), 0644)
	if err != nil {
		t.Skipf("cannot create temp file: %v", err)
	}

	importer := DefaultHostListImporter()
	entries, err := importer.ImportFromFile(csvFile, FormatCSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	if entries[0].Hostname != "router-main" {
		t.Errorf("expected hostname router-main, got %s", entries[0].Hostname)
	}
	if entries[0].Comment != "Main router" {
		t.Errorf("expected comment 'Main router', got %s", entries[0].Comment)
	}
}

// TestHostListImporter_ImportJSON — ветка: импорт JSON файла
func TestHostListImporter_ImportJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "hosts.json")

	content := `[
		{"ip": "192.168.1.1", "hostname": "router-main", "comment": "Main router"},
		{"ip": "192.168.1.2", "hostname": "switch-core"},
		{"ip": "192.168.1.10"}
	]`
	err := os.WriteFile(jsonFile, []byte(content), 0644)
	if err != nil {
		t.Skipf("cannot create temp file: %v", err)
	}

	importer := DefaultHostListImporter()
	entries, err := importer.ImportFromFile(jsonFile, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	if entries[0].IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", entries[0].IP)
	}
	if entries[0].Hostname != "router-main" {
		t.Errorf("expected hostname router-main, got %s", entries[0].Hostname)
	}
}

// TestHostListImporter_ImportFromString_TXT — ветка: импорт TXT из строки
func TestHostListImporter_ImportFromString_TXT(t *testing.T) {
	importer := DefaultHostListImporter()
	content := `192.168.1.1 router
192.168.1.2 switch`

	entries, err := importer.ImportFromString(content, FormatTXT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

// TestHostListImporter_ImportFromString_CSV — ветка: импорт CSV из строки
func TestHostListImporter_ImportFromString_CSV(t *testing.T) {
	importer := DefaultHostListImporter()
	content := `192.168.1.1,router,Main`

	entries, err := importer.ImportFromString(content, FormatCSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Hostname != "router" {
		t.Errorf("expected hostname router, got %s", entries[0].Hostname)
	}
	if entries[0].Comment != "Main" {
		t.Errorf("expected comment 'Main', got %s", entries[0].Comment)
	}
}

// TestHostListImporter_ImportFromString_JSON — ветка: импорт JSON из строки
func TestHostListImporter_ImportFromString_JSON(t *testing.T) {
	importer := DefaultHostListImporter()
	content := `[{"ip": "192.168.1.1", "hostname": "router"}]`

	entries, err := importer.ImportFromString(content, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Hostname != "router" {
		t.Errorf("expected hostname router, got %s", entries[0].Hostname)
	}
}

// TestHostListImporter_InvalidPath — ветка: невалидный путь
func TestHostListImporter_InvalidPath(t *testing.T) {
	importer := DefaultHostListImporter()
	_, err := importer.ImportFromFile("/nonexistent/path.txt", FormatAuto)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// TestHostListImporter_EmptyContent — ветка: пустое содержимое
func TestHostListImporter_EmptyContent(t *testing.T) {
	importer := DefaultHostListImporter()
	_, err := importer.ImportFromString("", FormatTXT)
	if err == nil {
		t.Error("expected error for empty content")
	}
}

// TestHostListImporter_InvalidIP — ветка: невалидный IP
func TestHostListImporter_InvalidIP(t *testing.T) {
	importer := DefaultHostListImporter()
	content := `invalid-ip
192.168.1.1
not-an-ip`

	entries, err := importer.ImportFromString(content, FormatTXT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Должна быть только одна валидная запись
	if len(entries) != 1 {
		t.Errorf("expected 1 valid entry, got %d", len(entries))
	}
	if entries[0].IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", entries[0].IP)
	}
}

// TestHostListImporter_CIDR_Expansion — ветка: расширение CIDR
func TestHostListImporter_CIDR_Expansion(t *testing.T) {
	importer := DefaultHostListImporter()
	content := `192.168.1.0/30 test-cidr`

	entries, err := importer.ImportFromString(content, FormatTXT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 CIDR entry, got %d", len(entries))
	}
	if !entries[0].IsCIDR {
		t.Error("expected IsCIDR to be true")
	}

	// Расширяем вручную
	expanded, err := expandCIDR(entries[0])
	if err != nil {
		t.Fatalf("unexpected error expanding CIDR: %v", err)
	}

	// /30 = 4 адреса
	if len(expanded) != 4 {
		t.Errorf("expected 4 expanded IPs, got %d", len(expanded))
	}
}

// TestHostListImporter_ExportTXT — ветка: экспорт в TXT
func TestHostListImporter_ExportTXT(t *testing.T) {
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "export.txt")

	entries := []HostEntry{
		{IP: "192.168.1.1", Hostname: "router", Comment: "Main"},
		{IP: "192.168.1.2", Hostname: "switch"},
	}

	importer := DefaultHostListImporter()
	err := importer.ExportToFile(entries, txtFile, FormatTXT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем файл
	content, err := os.ReadFile(txtFile)
	if err != nil {
		t.Fatalf("cannot read exported file: %v", err)
	}

	lines := string(content)
	if len(lines) == 0 {
		t.Error("expected non-empty exported file")
	}
}

// TestHostListImporter_ExportCSV — ветка: экспорт в CSV
func TestHostListImporter_ExportCSV(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "export.csv")

	entries := []HostEntry{
		{IP: "192.168.1.1", Hostname: "router", Comment: "Main"},
	}

	importer := DefaultHostListImporter()
	err := importer.ExportToFile(entries, csvFile, FormatCSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем файл
	content, err := os.ReadFile(csvFile)
	if err != nil {
		t.Fatalf("cannot read exported file: %v", err)
	}

	if len(content) == 0 {
		t.Error("expected non-empty exported file")
	}
}

// TestHostListImporter_ExportJSON — ветка: экспорт в JSON
func TestHostListImporter_ExportJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "export.json")

	entries := []HostEntry{
		{IP: "192.168.1.1", Hostname: "router", Comment: "Main"},
	}

	importer := DefaultHostListImporter()
	err := importer.ExportToFile(entries, jsonFile, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем файл
	content, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("cannot read exported file: %v", err)
	}

	if len(content) == 0 {
		t.Error("expected non-empty exported file")
	}
}

// TestHostListImporter_GetHostCount — ветка: подсчет хостов
func TestHostListImporter_GetHostCount(t *testing.T) {
	importer := DefaultHostListImporter()

	entries := []HostEntry{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.0/30", IsCIDR: true}, // 4 адреса
	}

	count := importer.GetHostCount(entries)
	if count != 6 { // 2 + 4
		t.Errorf("expected 6 hosts, got %d", count)
	}
}

// TestHostListImporter_MaxEntries — ветка: лимит записей
func TestHostListImporter_MaxEntries(t *testing.T) {
	importer := NewHostListImporter(2)

	content := `192.168.1.1
192.168.1.2
192.168.1.3`

	_, err := importer.ImportFromString(content, FormatTXT)
	if err == nil {
		t.Error("expected error for exceeding max entries")
	}
}

// TestHostListImporter_ImportIPv6 — ветка: IPv6 адреса
func TestHostListImporter_ImportIPv6(t *testing.T) {
	importer := DefaultHostListImporter()
	content := `::1 localhost
2001:db8::1 ipv6-server`

	entries, err := importer.ImportFromString(content, FormatTXT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	if !entries[0].IsIPv6 {
		t.Error("expected first entry to be IPv6")
	}
	if !entries[1].IsIPv6 {
		t.Error("expected second entry to be IPv6")
	}
}

// TestHostListImporter_DetectFormat — ветка: автоматическое определение формата
func TestHostListImporter_DetectFormat(t *testing.T) {
	if detectFormat("hosts.csv") != FormatCSV {
		t.Error("expected FormatCSV for .csv")
	}
	if detectFormat("hosts.txt") != FormatTXT {
		t.Error("expected FormatTXT for .txt")
	}
	if detectFormat("hosts.json") != FormatJSON {
		t.Error("expected FormatJSON for .json")
	}
	if detectFormat("hosts") != FormatTXT {
		t.Error("expected FormatTXT for unknown extension")
	}
}
