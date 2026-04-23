package display

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"network-scanner/internal/scanner"
)

func TestFormatResultsAsText(t *testing.T) {
	tests := []struct {
		name    string
		results []scanner.Result
		want    string
	}{
		{
			name:    "Empty results",
			results: []scanner.Result{},
			want:    "Результаты сканирования не найдены",
		},
		{
			name: "Single result",
			results: []scanner.Result{
				{
					IP:           "192.168.1.1",
					MAC:          "00:11:22:33:44:55",
					Hostname:     "router.local",
					Ports:        []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}},
					Protocols:    []string{"HTTP"},
					DeviceType:   "Router/Network Device",
					DeviceVendor: "Unknown",
				},
			},
			want: "РЕЗУЛЬТАТЫ СКАНИРОВАНИЯ СЕТИ",
		},
		{
			name: "Multiple results",
			results: []scanner.Result{
				{
					IP:         "192.168.1.1",
					MAC:        "00:11:22:33:44:55",
					Hostname:   "router.local",
					Ports:      []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}},
					Protocols:  []string{"HTTP"},
					DeviceType: "Router/Network Device",
				},
				{
					IP:         "192.168.1.100",
					MAC:        "aa:bb:cc:dd:ee:ff",
					Hostname:   "server.local",
					Ports:      []scanner.PortInfo{{Port: 443, State: "open", Protocol: "tcp", Service: "HTTPS"}},
					Protocols:  []string{"HTTPS"},
					DeviceType: "Web Server",
				},
			},
			want: "РЕЗУЛЬТАТЫ СКАНИРОВАНИЯ СЕТИ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatResultsAsText(tt.results)
			if !strings.Contains(got, tt.want) {
				t.Errorf("FormatResultsAsText() should contain %q", tt.want)
			}
			if len(tt.results) == 0 {
				if !strings.Contains(got, "не найдены") {
					t.Error("FormatResultsAsText() should indicate no results found")
				}
			} else {
				// Проверяем, что все IP адреса присутствуют
				for _, result := range tt.results {
					if !strings.Contains(got, result.IP) {
						t.Errorf("FormatResultsAsText() should contain IP %q", result.IP)
					}
				}
			}
		})
	}
}

func TestFormatResultsAsText_ContainsAnalytics(t *testing.T) {
	results := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Ports:      []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}},
			Protocols:  []string{"HTTP"},
			DeviceType: "Web Server",
		},
	}

	text := FormatResultsAsText(results)

	// Проверяем наличие секций аналитики
	expectedSections := []string{
		"АНАЛИТИКА ПРОВОДНЫХ СЕТЕЙ",
		"ПРОТОКОЛЫ В СЕТИ",
		"ИСПОЛЬЗУЕМЫЕ ПОРТЫ",
		"ТИПЫ УСТРОЙСТВ",
		"ОБЩАЯ СТАТИСТИКА",
	}

	for _, section := range expectedSections {
		if !strings.Contains(text, section) {
			t.Errorf("FormatResultsAsText() should contain section %q", section)
		}
	}
}

func TestFormatResultsAsText_Statistics(t *testing.T) {
	results := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Ports:      []scanner.PortInfo{{Port: 80, State: "open"}, {Port: 443, State: "open"}},
			Protocols:  []string{"HTTP", "HTTPS"},
			DeviceType: "Web Server",
		},
		{
			IP:         "192.168.1.2",
			Ports:      []scanner.PortInfo{{Port: 22, State: "open"}},
			Protocols:  []string{"SSH"},
			DeviceType: "Linux/Unix Server",
		},
	}

	text := FormatResultsAsText(results)

	// Проверяем статистику
	if !strings.Contains(text, "Всего обнаружено устройств: 2") {
		t.Error("FormatResultsAsText() should show device count")
	}
	if !strings.Contains(text, "Устройств с открытыми портами: 2") {
		t.Error("FormatResultsAsText() should show devices with open ports")
	}
	if !strings.Contains(text, "Всего открытых портов: 3") {
		t.Error("FormatResultsAsText() should show total open ports")
	}
}

func TestDisplayResults_EmptyResults(t *testing.T) {
	// Проверяем, что функция не паникует на пустых результатах
	DisplayResults([]scanner.Result{})
}

func TestDisplayAnalytics_EmptyResults(t *testing.T) {
	// Проверяем, что функция не паникует на пустых результатах
	DisplayAnalytics([]scanner.Result{})
}

func TestDisplayAnalytics_WithResults(t *testing.T) {
	results := []scanner.Result{
		{
			IP:         "192.168.1.1",
			Ports:      []scanner.PortInfo{{Port: 80, State: "open"}, {Port: 443, State: "open"}},
			Protocols:  []string{"HTTP", "HTTPS"},
			DeviceType: "Web Server",
		},
		{
			IP:         "192.168.1.2",
			Ports:      []scanner.PortInfo{{Port: 22, State: "open"}},
			Protocols:  []string{"SSH"},
			DeviceType: "Linux/Unix Server",
		},
	}

	// Проверяем, что функция не паникует
	DisplayAnalytics(results)
}

func TestFormatPorts_ShowRawBannerToggle(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP", Version: "HTTP/1.1 200 (nginx)", Banner: "HTTP/1.1 200 OK | Server=nginx"},
	}

	SetShowRawBanners(false)
	noRaw := formatPorts(ports)
	if strings.Contains(noRaw, "[banner:") {
		t.Fatalf("formatPorts should hide raw banner when disabled, got: %s", noRaw)
	}
	if !strings.Contains(noRaw, "[version:") {
		t.Fatalf("formatPorts should keep version when raw hidden, got: %s", noRaw)
	}

	SetShowRawBanners(true)
	withRaw := formatPorts(ports)
	if !strings.Contains(withRaw, "[banner:") {
		t.Fatalf("formatPorts should show raw banner when enabled, got: %s", withRaw)
	}
}

func TestFormatResultsAsText_Golden(t *testing.T) {
	t.Setenv("TZ", "UTC")
	SetShowRawBanners(false)
	results := []scanner.Result{
		{
			IP:           "192.168.1.10",
			MAC:          "aa:bb:cc:dd:ee:10",
			Hostname:     "router.local",
			Ports:        []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}, {Port: 443, State: "open", Protocol: "tcp", Service: "HTTPS"}, {Port: 22, State: "open", Protocol: "tcp", Service: "SSH"}},
			Protocols:    []string{"HTTP", "HTTPS", "SSH"},
			DeviceType:   "Router/Network Device",
			DeviceVendor: "TestVendor",
		},
		{
			IP:           "192.168.1.20",
			MAC:          "aa:bb:cc:dd:ee:20",
			Hostname:     "nas.local",
			Ports:        []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}, {Port: 443, State: "open", Protocol: "tcp", Service: "HTTPS"}},
			Protocols:    []string{"HTTP", "HTTPS"},
			DeviceType:   "Router/Network Device",
			DeviceVendor: "TestVendor",
		},
		{
			IP:           "192.168.1.30",
			MAC:          "aa:bb:cc:dd:ee:30",
			Hostname:     "printer.local",
			Ports:        []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}},
			Protocols:    []string{"HTTP"},
			DeviceType:   "Printer",
			DeviceVendor: "TestVendor",
		},
	}

	got := normalizeNewlines(FormatResultsAsText(results))
	goldenPath := filepath.Join("testdata", "format_results_as_text.golden.txt")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("failed to create golden directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}
	want := normalizeNewlines(string(wantBytes))
	if got != want {
		t.Fatalf("golden mismatch for %s; run with UPDATE_GOLDEN=1 to refresh", goldenPath)
	}
}

func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

func BenchmarkFormatResultsAsTextLarge(b *testing.B) {
	results := make([]scanner.Result, 0, 256)
	for i := 1; i <= 256; i++ {
		results = append(results, scanner.Result{
			IP:           fmt.Sprintf("192.168.1.%d", i%254+1),
			MAC:          fmt.Sprintf("aa:bb:cc:dd:ee:%02x", i%255),
			Hostname:     fmt.Sprintf("host-%03d.local", i),
			Ports:        []scanner.PortInfo{{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"}, {Port: 443, State: "open", Protocol: "tcp", Service: "HTTPS"}},
			Protocols:    []string{"HTTP", "HTTPS"},
			DeviceType:   "Router/Network Device",
			DeviceVendor: "BenchVendor",
		})
	}
	SetShowRawBanners(false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatResultsAsText(results)
	}
}
