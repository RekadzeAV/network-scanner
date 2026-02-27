package display

import (
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
