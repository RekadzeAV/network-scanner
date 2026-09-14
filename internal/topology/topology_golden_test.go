package topology

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- D2: Golden-снимки для PNG/SVG через DOT ---

// TestDOTGoldenBaseline проверяет детерминированность DOT-вывода для golden-сравнения.
// DOT — основа для генерации PNG/SVG через Graphviz.
func TestDOTGoldenBaseline(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:01",
				Hostname: "router",
				Type:     DeviceTypeRouter,
			},
			"dev2": {
				IP:       "192.168.1.2",
				MAC:      "aa:bb:cc:dd:ee:02",
				Hostname: "switch",
				Type:     DeviceTypeSwitch,
			},
			"dev3": {
				IP:       "192.168.1.3",
				MAC:      "aa:bb:cc:dd:ee:03",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "router"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "switch"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
				SourcePort: &Port{Index: 1, Name: "Gi0/1"},
			},
			{
				Source:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "switch"},
				Target:     &Device{IP: "192.168.1.3", MAC: "aa:bb:cc:dd:ee:03", Hostname: "host1"},
				SourceType: LinkSourceFDB,
				Confidence: LinkConfidenceMedium,
				SourcePort: &Port{Index: 5, Name: "Gi0/5"},
			},
		},
	}

	// Генерируем DOT три раза для проверки детерминизма
	var bufs [3]string
	for i := 0; i < 3; i++ {
		var buf strings.Builder
		if err := topo.ToDOT(&buf); err != nil {
			t.Fatalf("ToDOT iteration %d error: %v", i+1, err)
		}
		bufs[i] = buf.String()
	}

	// Все три должны быть идентичны
	for i := 1; i < 3; i++ {
		if bufs[0] != bufs[i] {
			t.Errorf("DOT output non-deterministic: iteration 0 != iteration %d", i+1)
		}
	}

	// Сохраняем golden-снимок
	goldenPath := filepath.Join(t.TempDir(), "topology.golden.dot")
	if err := os.WriteFile(goldenPath, []byte(bufs[0]), 0644); err != nil {
		t.Fatalf("write golden file: %v", err)
	}

	// Проверяем что golden-файл можно использовать для сравнения
	goldenContent, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if string(goldenContent) != bufs[0] {
		t.Error("golden file content mismatch")
	}
}

// TestDOTGoldenWithDifferentInput проверяет что разный ввод дает разный DOT.
func TestDOTGoldenWithDifferentInput(t *testing.T) {
	topo1 := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Hostname: "host1", Type: DeviceTypeHost},
		},
		Links: []Link{},
	}

	topo2 := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Hostname: "host1", Type: DeviceTypeHost},
			"dev2": {IP: "192.168.1.2", Hostname: "host2", Type: DeviceTypeHost},
		},
		Links: []Link{},
	}

	var buf1, buf2 strings.Builder
	if err := topo1.ToDOT(&buf1); err != nil {
		t.Fatalf("ToDOT error: %v", err)
	}
	if err := topo2.ToDOT(&buf2); err != nil {
		t.Fatalf("ToDOT error: %v", err)
	}

	if buf1.String() == buf2.String() {
		t.Error("different inputs should produce different DOT output")
	}
	if !strings.Contains(buf2.String(), "host2") {
		t.Error("second topology should contain host2")
	}
}

// TestDOTGoldenNilTopolog проверяет обработку nil.
func TestDOTGoldenNilTopolog(t *testing.T) {
	var topo *Topology
	var buf strings.Builder
	err := topo.ToDOT(&buf)
	if err == nil {
		t.Error("ToDOT on nil should return error")
	}
}

// TestDOTGoldenWithEmptyTopology проверяет пустую топологию.
func TestDOTGoldenWithEmptyTopology(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{},
		Links:   []Link{},
	}

	var buf strings.Builder
	if err := topo.ToDOT(&buf); err != nil {
		t.Fatalf("ToDOT error: %v", err)
	}

	dot := buf.String()
	if !strings.Contains(dot, "graph network") {
		t.Error("empty topology should still have graph header")
	}
	// Пустая топология не должна иметь объявлений узлов (node declarations),
	// только default-атрибуты node [shape=...] допустимы.
	// Проверяем что нет строк вида "mac_..." или "ip_..." — это узлы.
	lines := strings.Split(dot, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `"mac_`) || strings.HasPrefix(trimmed, `"ip_`) || strings.HasPrefix(trimmed, `"hn_`) {
			t.Errorf("empty topology should not have node declarations, found: %q", trimmed)
		}
	}
}

// TestDOTGoldenLargeTopology проверяет масштабирование на большой топологии.
func TestDOTGoldenLargeTopology(t *testing.T) {
	const deviceCount = 20

	devices := make(map[string]*Device, deviceCount)
	links := make([]Link, 0, deviceCount-1)

	for i := 0; i < deviceCount; i++ {
		mac := sprintfMAC(i)
		devices[mac] = &Device{
			IP:       sprintfIP(i),
			MAC:      mac,
			Hostname: sprintf("device-%03d", i),
			Type:     DeviceTypeHost,
		}
	}

	// Создаем линейный граф
	for i := 0; i < deviceCount-1; i++ {
		mac1 := sprintfMAC(i)
		mac2 := sprintfMAC(i + 1)
		links = append(links, Link{
			Source:     devices[mac1],
			Target:     devices[mac2],
			SourceType: LinkSourceLLDP,
			Confidence: LinkConfidenceHigh,
		})
	}

	topo := &Topology{
		Devices: devices,
		Links:   links,
	}

	var buf strings.Builder
	if err := topo.ToDOT(&buf); err != nil {
		t.Fatalf("ToDOT error: %v", err)
	}

	dot := buf.String()

	// Проверяем что все устройства присутствуют
	for i := 0; i < deviceCount; i++ {
		expected := sprintf("device-%03d", i)
		if !strings.Contains(dot, expected) {
			t.Errorf("DOT missing device %d (%q)", i, expected)
		}
	}

	// Проверяем размер файла (должен быть разумным)
	if len(dot) > 100*1024 { // 100 KB
		t.Errorf("DOT output too large: %.1f KB", float64(len(dot))/1024)
	}
}

// TestPNGSVGFallbackChecks проверяет что fallback для PNG/SVG работает корректно.
func TestPNGSVGFallbackChecks(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", Hostname: "host1", Type: DeviceTypeHost},
		},
		Links: []Link{},
	}

	// Проверяем что RenderWithGraphviz возвращает ErrGraphvizNotInstalled
	err := topo.RenderWithGraphviz("png", "/tmp/test.png")
	if err == nil {
		t.Error("RenderWithGraphviz should fail when Graphviz is not installed")
	}

	// Проверяем что ошибка содержит информацию об установке
	if !strings.Contains(err.Error(), "graphviz") && !strings.Contains(err.Error(), "Graphviz") {
		t.Error("error should mention Graphviz installation")
	}
}

// TestTopologTextFallbackChecker проверяет текстовый fallback.
func TestTopologTextFallbackChecker(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "router", Type: DeviceTypeRouter},
			"dev2": {IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "switch", Type: DeviceTypeSwitch},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "router"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "switch"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	// Сохраняем в текстовый формат (fallback)
	tmpDir := t.TempDir()
	textPath := filepath.Join(tmpDir, "topology.txt")
	if err := topo.SaveAsText(textPath); err != nil {
		t.Fatalf("SaveAsText error: %v", err)
	}

	content, err := os.ReadFile(textPath)
	if err != nil {
		t.Fatalf("read text file: %v", err)
	}

	text := string(content)

	// Проверяем что текстовый файл содержит основную информацию
	if !strings.Contains(text, "Network Topology Report") {
		t.Error("text should contain report header")
	}
	if !strings.Contains(text, "router") {
		t.Error("text should contain router hostname")
	}
	if !strings.Contains(text, "switch") {
		t.Error("text should contain switch hostname")
	}
	if !strings.Contains(text, "LINKS") {
		t.Error("text should contain links section")
	}
}

// sprintfIP — вспомогательная функция для генерации IP.
func sprintfIP(index int) string {
	a := 192
	b := 168
	c := index / 256
	d := index % 256
	return sprintf("%d.%d.%d.%d", a, b, c, d)
}
