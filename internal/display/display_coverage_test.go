package display

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"network-scanner/internal/scanner"
)

// ============================================================================
// truncateString — ветки: <=maxLen, >maxLen, maxLen<=3, normal case
// ============================================================================

func TestTruncateString_UnderLimit(t *testing.T) {
	got := truncateString("Hello", 10)
	if got != "Hello" {
		t.Fatalf("truncateString(\"Hello\", 10) = %q, want %q", got, "Hello")
	}
}

func TestTruncateString_EqualLimit(t *testing.T) {
	got := truncateString("Hello", 5)
	if got != "Hello" {
		t.Fatalf("truncateString(\"Hello\", 5) = %q, want %q", got, "Hello")
	}
}

func TestTruncateString_OverLimit(t *testing.T) {
	got := truncateString("Hello World", 8)
	want := "Hello..."
	if got != want {
		t.Fatalf("truncateString(\"Hello World\", 8) = %q, want %q", got, want)
	}
}

func TestTruncateString_MaxLenLessThan3(t *testing.T) {
	got := truncateString("Hello", 1)
	if got != "H" {
		t.Fatalf("truncateString(\"Hello\", 1) = %q, want %q", got, "H")
	}
}

func TestTruncateString_MaxLenZero(t *testing.T) {
	got := truncateString("Hello", 0)
	if got != "" {
		t.Fatalf("truncateString(\"Hello\", 0) = %q, want %q", got, "")
	}
}

func TestTruncateString_MaxLenTwo(t *testing.T) {
	got := truncateString("Hello", 2)
	if got != "He" {
		t.Fatalf("truncateString(\"Hello\", 2) = %q, want %q", got, "He")
	}
}

func TestTruncateString_MaxLenThree(t *testing.T) {
	got := truncateString("Hello World", 3)
	want := "Hel"
	if got != want {
		t.Fatalf("truncateString(\"Hello World\", 3) = %q, want %q", got, want)
	}
}

func TestTruncateString_LongString(t *testing.T) {
	long := ""
	for i := 0; i < 200; i++ {
		long += "A"
	}
	got := truncateString(long, 100)
	want := long[:97] + "..."
	if got != want {
		t.Fatalf("long string truncated incorrectly")
	}
}

func TestTruncateString_Empty(t *testing.T) {
	got := truncateString("", 10)
	if got != "" {
		t.Fatalf("truncateString(empty) = %q, want empty", got)
	}
}

// ============================================================================
// formatOSGuess — ветки: empty, with confidence, with reason, all fields
// ============================================================================

func TestFormatOSGuess_Empty(t *testing.T) {
	r := scanner.Result{GuessOS: "", GuessOSConfidence: "", GuessOSReason: ""}
	got := formatOSGuess(r)
	if got != "-" {
		t.Fatalf("formatOSGuess(empty) = %q, want %q", got, "-")
	}
}

func TestFormatOSGuess_WhitespaceOnly(t *testing.T) {
	r := scanner.Result{GuessOS: "   ", GuessOSConfidence: "", GuessOSReason: ""}
	got := formatOSGuess(r)
	if got != "-" {
		t.Fatalf("formatOSGuess(whitespace) = %q, want %q", got, "-")
	}
}

func TestFormatOSGuess_OnlyOS(t *testing.T) {
	r := scanner.Result{GuessOS: "Linux", GuessOSConfidence: "", GuessOSReason: ""}
	got := formatOSGuess(r)
	if got != "Linux" {
		t.Fatalf("formatOSGuess(only OS) = %q, want %q", got, "Linux")
	}
}

func TestFormatOSGuess_WithConfidence(t *testing.T) {
	r := scanner.Result{GuessOS: "Linux", GuessOSConfidence: "90%", GuessOSReason: ""}
	got := formatOSGuess(r)
	want := "Linux (90%)"
	if got != want {
		t.Fatalf("formatOSGuess(confidence) = %q, want %q", got, want)
	}
}

func TestFormatOSGuess_WithReason(t *testing.T) {
	r := scanner.Result{GuessOS: "Linux", GuessOSConfidence: "", GuessOSReason: "TTL match"}
	got := formatOSGuess(r)
	want := "Linux — TTL match"
	if got != want {
		t.Fatalf("formatOSGuess(reason) = %q, want %q", got, want)
	}
}

func TestFormatOSGuess_AllFields(t *testing.T) {
	r := scanner.Result{GuessOS: "Windows", GuessOSConfidence: "95%", GuessOSReason: "TCP Window size 65535"}
	got := formatOSGuess(r)
	want := "Windows (95%) — TCP Window size 65535"
	if got != want {
		t.Fatalf("formatOSGuess(all) = %q, want %q", got, want)
	}
}

func TestFormatOSGuess_TruncatedReason(t *testing.T) {
	longReason := ""
	for i := 0; i < 100; i++ {
		longReason += "X"
	}
	r := scanner.Result{GuessOS: "Linux", GuessOSConfidence: "90%", GuessOSReason: longReason}
	got := formatOSGuess(r)
	if len(got) > len("Linux (90%) — "+strings.Repeat("X", 60)) {
		t.Fatalf("reason should be truncated: got len=%d", len(got))
	}
}

// ============================================================================
// getPortPurpose — ветки: known ports, unknown port
// ============================================================================

func TestGetPortPurpose_KnownPorts(t *testing.T) {
	// Для большинства портов getPortPurpose возвращает описание из IANA
	// (portdb.Description), а не из локальной карты. Проверяем что возвращает непустое.
	knownPorts := []int{20, 21, 22, 23, 25, 53, 80, 110, 143, 443, 445, 3306, 3389, 5432, 5900, 8080, 8443}
	for _, port := range knownPorts {
		got := getPortPurpose(port)
		if got == "" {
			t.Errorf("getPortPurpose(%d) should return non-empty string", port)
		}
	}
}

func TestGetPortPurpose_Unknown(t *testing.T) {
	got := getPortPurpose(9999)
	// Port 9999 может быть в IANA реестре, проверяем что не пустое
	if got == "" {
		t.Fatal("getPortPurpose(9999) should return non-empty string")
	}
}

func TestGetPortPurpose_Zero(t *testing.T) {
	got := getPortPurpose(0)
	if got == "" {
		t.Fatal("getPortPurpose(0) should return non-empty string")
	}
}

func TestGetPortPurpose_Negative(t *testing.T) {
	got := getPortPurpose(-1)
	want := "Неизвестное назначение"
	if got != want {
		t.Fatalf("getPortPurpose(-1) = %q, want %q", got, want)
	}
}

// ============================================================================
// SaveResultsToFile — ветки: success, error (invalid path)
// ============================================================================

func TestSaveResultsToFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "output.txt")

	results := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Ports:      []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}},
			Protocols:  []string{"HTTP"},
			DeviceType: "Web Server",
		},
	}

	err := SaveResultsToFile(results, filename)
	if err != nil {
		t.Fatalf("SaveResultsToFile() error = %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if !strings.Contains(string(data), "192.168.1.1") {
		t.Fatal("Saved file should contain IP address")
	}
	if !strings.Contains(string(data), "АНАЛИТИКА") {
		t.Fatal("Saved file should contain analytics section")
	}
}

func TestSaveResultsToFile_EmptyResults(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty.txt")

	err := SaveResultsToFile([]scanner.Result{}, filename)
	if err != nil {
		t.Fatalf("SaveResultsToFile() error = %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if !strings.Contains(string(data), "не найдены") {
		t.Fatal("Saved file should indicate no results")
	}
}

func TestSaveResultsToFile_InvalidPath(t *testing.T) {
	err := SaveResultsToFile([]scanner.Result{}, "C:/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("SaveResultsToFile() should return error for invalid path")
	}
}

// ============================================================================
// SaveResultsToJSON — ветки: success, error (invalid path)
// ============================================================================

func TestSaveResultsToJSON_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "output.json")

	results := []scanner.Result{
		{
			IP:                "192.168.1.1",
			MAC:               "aa:bb:cc:dd:ee:ff",
			Hostname:          "test.local",
			Ports:             []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP", Version: "nginx/1.25"}},
			Protocols:         []string{"HTTP"},
			DeviceType:        "Web Server",
			DeviceVendor:      "TestVendor",
			IsAlive:           true,
			GuessOS:           "Linux",
			GuessOSConfidence: "90%",
			GuessOSReason:     "TTL match",
		},
	}

	err := SaveResultsToJSON(results, filename)
	if err != nil {
		t.Fatalf("SaveResultsToJSON() error = %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if !strings.Contains(string(data), "192.168.1.1") {
		t.Fatal("JSON should contain IP address")
	}
	if !strings.Contains(string(data), "test.local") {
		t.Fatal("JSON should contain hostname")
	}
	if !strings.Contains(string(data), "analytics") {
		t.Fatal("JSON should contain analytics section")
	}
	if !strings.Contains(string(data), "scan_date") {
		t.Fatal("JSON should contain scan_date")
	}
}

func TestSaveResultsToJSON_EmptyResults(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty.json")

	err := SaveResultsToJSON([]scanner.Result{}, filename)
	if err != nil {
		t.Fatalf("SaveResultsToJSON() error = %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if !strings.Contains(string(data), "total_devices") {
		t.Fatal("JSON should contain total_devices")
	}
}

func TestSaveResultsToJSON_InvalidPath(t *testing.T) {
	err := SaveResultsToJSON([]scanner.Result{}, "C:/nonexistent/path/file.json")
	if err == nil {
		t.Fatal("SaveResultsToJSON() should return error for invalid path")
	}
}

// ============================================================================
// SaveResultsToCSV — ветки: success, error (invalid path)
// ============================================================================

func TestSaveResultsToCSV_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "output.csv")

	results := []scanner.Result{
		{
			IP:           "192.168.1.1",
			MAC:          "aa:bb:cc:dd:ee:ff",
			Hostname:     "test.local",
			Ports:        []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}},
			Protocols:    []string{"HTTP"},
			DeviceType:   "Web Server",
			DeviceVendor: "TestVendor",
			IsAlive:      true,
		},
		{
			IP:           "192.168.1.2",
			MAC:          "11:22:33:44:55:66",
			Hostname:     "router.local",
			Ports:        []scanner.PortInfo{{Port: 22, State: "open", Protocol: "tcp", Service: "SSH"}},
			Protocols:    []string{"SSH"},
			DeviceType:   "Router",
			DeviceVendor: "Cisco",
			IsAlive:      true,
		},
	}

	err := SaveResultsToCSV(results, filename)
	if err != nil {
		t.Fatalf("SaveResultsToCSV() error = %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 3 {
		t.Fatalf("CSV should have header + 2 data rows, got %d lines", len(lines))
	}

	// Проверка заголовка
	if !strings.Contains(lines[0], "IP") || !strings.Contains(lines[0], "MAC") {
		t.Fatal("CSV header should contain IP and MAC columns")
	}

	// Проверка данных
	if !strings.Contains(lines[1], "192.168.1.1") {
		t.Fatal("First row should contain first IP")
	}
	if !strings.Contains(lines[2], "192.168.1.2") {
		t.Fatal("Second row should contain second IP")
	}
}

func TestSaveResultsToCSV_EmptyResults(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty.csv")

	err := SaveResultsToCSV([]scanner.Result{}, filename)
	if err != nil {
		t.Fatalf("SaveResultsToCSV() error = %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	// Должен быть только заголовок
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("CSV with empty results should have only header, got %d lines", len(lines))
	}
}

func TestSaveResultsToCSV_InvalidPath(t *testing.T) {
	err := SaveResultsToCSV([]scanner.Result{}, "C:/nonexistent/path/file.csv")
	if err == nil {
		t.Fatal("SaveResultsToCSV() should return error for invalid path")
	}
}

func TestSaveResultsToCSV_ClosedPort(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "closed.csv")

	results := []scanner.Result{
		{
			IP:      "192.168.1.1",
			Ports:   []scanner.PortInfo{{Port: 80, State: "closed", Protocol: "tcp"}},
			IsAlive: false,
		},
	}

	err := SaveResultsToCSV(results, filename)
	if err != nil {
		t.Fatalf("SaveResultsToCSV() error = %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if !strings.Contains(string(data), "false") {
		t.Fatal("CSV should contain 'false' for IsAlive")
	}
}

// ============================================================================
// formatPorts — ветки: много открытых портов (>50), закрытые порты
// ============================================================================

func TestFormatPorts_ManyOpenPorts(t *testing.T) {
	ports := make([]scanner.PortInfo, 0, 60)
	for i := 1; i <= 60; i++ {
		ports = append(ports, scanner.PortInfo{
			Port:     i,
			State:    "open",
			Protocol: "tcp",
			Service:  "HTTP",
		})
	}

	result := formatPorts(ports)
	if !strings.Contains(result, "... и еще") {
		t.Fatal("formatPorts should show truncated message for >50 open ports")
	}
}

func TestFormatPorts_ClosedPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, State: "closed", Protocol: "tcp"},
		{Port: 443, State: "closed", Protocol: "tcp"},
	}

	result := formatPorts(ports)
	if !strings.Contains(result, "closed") {
		t.Fatal("formatPorts should show closed ports")
	}
}

func TestFormatPorts_EmptyPorts(t *testing.T) {
	result := formatPorts([]scanner.PortInfo{})
	if result != "-" {
		t.Fatalf("formatPorts(empty) = %q, want %q", result, "-")
	}
}

func TestFormatPorts_MixedPorts(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"},
		{Port: 443, State: "open", Protocol: "tcp", Service: "HTTPS"},
		{Port: 22, State: "closed", Protocol: "tcp"},
	}

	result := formatPorts(ports)
	if !strings.Contains(result, "80/tcp") {
		t.Fatal("formatPorts should contain open port 80")
	}
	if !strings.Contains(result, "443/tcp") {
		t.Fatal("formatPorts should contain open port 443")
	}
	if !strings.Contains(result, "22/tcp (closed)") {
		t.Fatal("formatPorts should contain closed port 22")
	}
}

func TestFormatPorts_VersionAndBanner(t *testing.T) {
	ports := []scanner.PortInfo{
		{
			Port:     80,
			State:    "open",
			Protocol: "tcp",
			Service:  "HTTP",
			Version:  "nginx/1.25.0",
			Banner:   "HTTP/1.1 200 OK | Server=nginx",
		},
	}

	// Без showRawBanners
	SetShowRawBanners(false)
	result := formatPorts(ports)
	if !strings.Contains(result, "[version: nginx/1.25.0]") {
		t.Fatal("formatPorts should contain version when raw banners disabled")
	}
	if strings.Contains(result, "[banner:") {
		t.Fatal("formatPorts should NOT contain banner when raw banners disabled")
	}

	// С showRawBanners
	SetShowRawBanners(true)
	result = formatPorts(ports)
	if !strings.Contains(result, "[banner: HTTP/1.1 200 OK | Server=nginx]") {
		t.Fatal("formatPorts should contain banner when raw banners enabled")
	}
}

func TestFormatPorts_VerboseTruncation(t *testing.T) {
	longVersion := ""
	for i := 0; i < 200; i++ {
		longVersion += "V"
	}
	ports := []scanner.PortInfo{
		{
			Port:     80,
			State:    "open",
			Protocol: "tcp",
			Service:  "HTTP",
			Version:  longVersion,
		},
	}

	result := formatPorts(ports)
	if len(result) > 200 {
		t.Fatal("formatPorts should truncate long version strings")
	}
}

// ============================================================================
// DisplayResults — ветки: с результатами, с пустыми полями
// ============================================================================

func TestDisplayResults_WithResults(t *testing.T) {
	// Проверяем что не паникует
	results := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Ports:      []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}},
			Protocols:  []string{"HTTP"},
			DeviceType: "Web Server",
		},
	}
	DisplayResults(results)
}

func TestDisplayResults_EmptyFields(t *testing.T) {
	// Проверяем что не паникует при пустых полях
	results := []scanner.Result{
		{
			IP:        "192.168.1.1",
			MAC:       "",
			Hostname:  "",
			Ports:     []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}},
			Protocols: []string{"HTTP"},
		},
	}
	DisplayResults(results)
}

// ============================================================================
// getProtocolDescription — ветки: known, unknown
// ============================================================================

func TestGetProtocolDescription_KnownProtocols(t *testing.T) {
	tests := []struct {
		protocol string
		check    string
	}{
		{"HTTP", "веб-серверов"},
		{"HTTPS", "зашифрованная"},
		{"SSH", "удаленное управление"},
		{"FTP", "передача файлов"},
		{"SMTP", "электронной почты"},
		{"DNS", "доменных имен"},
		{"RDP", "удаленный рабочий стол"},
		{"MySQL", "База данных MySQL"},
		{"PostgreSQL", "База данных PostgreSQL"},
	}

	for _, tt := range tests {
		got := getProtocolDescription(tt.protocol)
		if !strings.Contains(got, tt.check) {
			t.Errorf("getProtocolDescription(%q) = %q, should contain %q", tt.protocol, got, tt.check)
		}
	}
}

func TestGetProtocolDescription_Unknown(t *testing.T) {
	got := getProtocolDescription("UNKNOWN")
	want := "Неизвестный протокол"
	if got != want {
		t.Fatalf("getProtocolDescription(\"UNKNOWN\") = %q, want %q", got, want)
	}
}

func TestGetProtocolDescription_Empty(t *testing.T) {
	got := getProtocolDescription("")
	want := "Неизвестный протокол"
	if got != want {
		t.Fatalf("getProtocolDescription(\"\") = %q, want %q", got, want)
	}
}
