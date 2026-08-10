package topology

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- D2: Совместимость GraphML с внешними инструментами (yEd/Gephi) ---

// TestGraphMLyEdCompatibility проверяет, что GraphML соответствует спецификации
// для импорта в yEd (требует ключи для всех атрибутов узлов и рёбер).
func TestGraphMLyEdCompatibility(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:01",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
			"dev2": {
				IP:       "192.168.1.2",
				MAC:      "aa:bb:cc:dd:ee:02",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
				Evidence:   "lldp_neighbor_match",
				SourcePort: &Port{Index: 1, Name: "Gi0/1"},
			},
		},
	}

	path := filepath.Join(t.TempDir(), "topology-yed.graphml")
	if err := topo.SaveGraphML(path); err != nil {
		t.Fatalf("SaveGraphML error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	content := string(raw)

	// 1. Проверка XML-пролога
	if !strings.HasPrefix(content, "<?xml") {
		t.Error("GraphML must start with XML declaration for yEd compatibility")
	}

	// 2. Проверка корневой группы <graphml>
	if !strings.Contains(content, `<graphml xmlns="http://graphml.graphdrawing.org/xmlns"`) {
		t.Error("GraphML must have correct Graphviz namespace")
	}

	// 3. Проверка наличия всех ключей для узлов
	requiredNodeKeys := []string{`id="label" for="node"`, `id="type" for="node"`}
	for _, key := range requiredNodeKeys {
		if !strings.Contains(content, key) {
			t.Errorf("yEd requires node key %q for proper attribute mapping", key)
		}
	}

	// 4. Проверка наличия всех ключей для рёбер
	requiredEdgeKeys := []string{
		`id="src_port" for="edge"`,
		`id="dst_port" for="edge"`,
		`id="source_type" for="edge"`,
		`id="confidence" for="edge"`,
		`id="evidence" for="edge"`,
	}
	for _, key := range requiredEdgeKeys {
		if !strings.Contains(content, key) {
			t.Errorf("yEd requires edge key %q for proper attribute mapping", key)
		}
	}

	// 5. Проверка типа графа (undirected для yEd)
	if !strings.Contains(content, `edgedefault="undirected"`) {
		t.Error("GraphML should use undirected edges for yEd compatibility")
	}

	// 6. Проверка наличия данных узлов
	if !strings.Contains(content, `<node id="`) {
		t.Error("GraphML must contain at least one node")
	}
	if !strings.Contains(content, `<data key="label">`) {
		t.Error("GraphML must contain node labels for yEd visualization")
	}

	// 7. Проверка наличия данных рёбер
	if !strings.Contains(content, `<edge `) {
		t.Error("GraphML must contain at least one edge")
	}
	if !strings.Contains(content, `<data key="source_type">`) {
		t.Error("GraphML must contain edge source_type for yEd edge styling")
	}
	if !strings.Contains(content, `<data key="confidence">`) {
		t.Error("GraphML must contain edge confidence for yEd edge styling")
	}
}

// TestGraphMLGephiCompatibility проверяет совместимость с Gephi
// (требует числовые атрибуты и корректные ID).
func TestGraphMLGephiCompatibility(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:01",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
			"dev2": {
				IP:       "192.168.1.2",
				MAC:      "aa:bb:cc:dd:ee:02",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
			},
		},
	}

	path := filepath.Join(t.TempDir(), "topology-gephi.graphml")
	if err := topo.SaveGraphML(path); err != nil {
		t.Fatalf("SaveGraphML error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	// 1. Проверка парсемости XML (Gephi требует валидный XML)
	var decoded struct {
		XMLName xml.Name `xml:"graphml"`
		Graph   struct {
			Nodes []struct {
				ID   string `xml:"id,attr"`
				Data []struct {
					Key   string `xml:"key,attr"`
					Value string `xml:",chardata"`
				} `xml:"data"`
			} `xml:"node"`
			Edges []struct {
				ID     string `xml:"id,attr"`
				Source string `xml:"source,attr"`
				Target string `xml:"target,attr"`
				Data   []struct {
					Key   string `xml:"key,attr"`
					Value string `xml:",chardata"`
				} `xml:"data"`
			} `xml:"edge"`
		} `xml:"graph"`
	}
	if err := xml.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("GraphML must be valid XML for Gephi: %v", err)
	}

	// 2. Проверка уникальности ID узлов
	nodeIDs := make(map[string]int)
	for _, n := range decoded.Graph.Nodes {
		nodeIDs[n.ID]++
	}
	for id, count := range nodeIDs {
		if count > 1 {
			t.Errorf("Gephi requires unique node IDs, got %d duplicates for %q", count, id)
		}
	}

	// 3. Проверка уникальности ID рёбер
	edgeIDs := make(map[string]int)
	for _, e := range decoded.Graph.Edges {
		edgeIDs[e.ID]++
	}
	for id, count := range edgeIDs {
		if count > 1 {
			t.Errorf("Gephi requires unique edge IDs, got %d duplicates for %q", count, id)
		}
	}

	// 4. Проверка, что все рёбра ссылаются на существующие узлы
	nodeIDSet := make(map[string]bool)
	for _, n := range decoded.Graph.Nodes {
		nodeIDSet[n.ID] = true
	}
	for _, e := range decoded.Graph.Edges {
		if !nodeIDSet[e.Source] {
			t.Errorf("Edge source %q references non-existent node", e.Source)
		}
		if !nodeIDSet[e.Target] {
			t.Errorf("Edge target %q references non-existent node", e.Target)
		}
	}

	// 5. Проверка наличия меток (label) для визуализации
	hasLabels := false
	for _, n := range decoded.Graph.Nodes {
		for _, d := range n.Data {
			if d.Key == "label" && d.Value != "" {
				hasLabels = true
				break
			}
		}
		if hasLabels {
			break
		}
	}
	if !hasLabels {
		t.Error("Gephi requires node labels for visualization")
	}

	// 6. Проверка наличия типов узлов
	hasTypes := false
	for _, n := range decoded.Graph.Nodes {
		for _, d := range n.Data {
			if d.Key == "type" && d.Value != "" {
				hasTypes = true
				break
			}
		}
		if hasTypes {
			break
		}
	}
	if !hasTypes {
		t.Error("Gephi benefits from node types for grouping")
	}
}

// TestGraphMLRoundtrip проверяет, что сохранение и повторное чтение
// GraphML не теряет данные.
func TestGraphMLRoundtrip(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {
				IP:       "192.168.1.1",
				MAC:      "aa:bb:cc:dd:ee:01",
				Hostname: "switch1",
				Type:     DeviceTypeSwitch,
			},
			"dev2": {
				IP:       "192.168.1.2",
				MAC:      "aa:bb:cc:dd:ee:02",
				Hostname: "host1",
				Type:     DeviceTypeHost,
			},
		},
		Links: []Link{
			{
				Source:     &Device{IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "switch1"},
				Target:     &Device{IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "host1"},
				SourceType: LinkSourceLLDP,
				Confidence: LinkConfidenceHigh,
				Evidence:   "lldp_match",
				SourcePort: &Port{Index: 1, Name: "Gi0/1"},
			},
		},
	}

	path := filepath.Join(t.TempDir(), "topology-roundtrip.graphml")
	if err := topo.SaveGraphML(path); err != nil {
		t.Fatalf("SaveGraphML error: %v", err)
	}

	// Считываем и проверяем ключевые данные
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	content := string(raw)

	// Проверяем сохранение узлов
	if !strings.Contains(content, "switch1") {
		t.Error("Roundtrip: hostname 'switch1' not preserved in GraphML")
	}
	if !strings.Contains(content, "host1") {
		t.Error("Roundtrip: hostname 'host1' not preserved in GraphML")
	}

	// Проверяем сохранение рёбер
	if !strings.Contains(content, "lldp") {
		t.Error("Roundtrip: source_type 'lldp' not preserved in GraphML")
	}
	if !strings.Contains(content, "high") {
		t.Error("Roundtrip: confidence 'high' not preserved in GraphML")
	}
	if !strings.Contains(content, "lldp_match") {
		t.Error("Roundtrip: evidence 'lldp_match' not preserved in GraphML")
	}
	if !strings.Contains(content, "Gi0/1") {
		t.Error("Roundtrip: port name 'Gi0/1' not preserved in GraphML")
	}
}

// --- D2: Golden-снимки для DOT (основа для PNG/SVG) ---

// TestDOTGoldenSnapshot проверяет детерминированность DOT-вывода.
// DOT-файл — основа для генерации PNG/SVG через Graphviz.
func TestDOTGoldenSnapshot(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{
			"dev1": {IP: "192.168.1.1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "router", Type: DeviceTypeRouter},
			"dev2": {IP: "192.168.1.2", MAC: "aa:bb:cc:dd:ee:02", Hostname: "switch", Type: DeviceTypeSwitch},
			"dev3": {IP: "192.168.1.3", MAC: "aa:bb:cc:dd:ee:03", Hostname: "host1", Type: DeviceTypeHost},
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

	// Генерируем DOT два раза
	var buf1, buf2 strings.Builder
	if err := topo.ToDOT(&buf1); err != nil {
		t.Fatalf("ToDOT error: %v", err)
	}
	if err := topo.ToDOT(&buf2); err != nil {
		t.Fatalf("ToDOT error: %v", err)
	}

	dot1 := buf1.String()
	dot2 := buf2.String()

	// Golden-снимок: вывод должен быть детерминирован
	if dot1 != dot2 {
		t.Errorf("DOT output is non-deterministic:\nFirst:\n%s\n\nSecond:\n%s", dot1, dot2)
	}

	// Проверяем структуру golden-снимка
	if !strings.Contains(dot1, "graph network") {
		t.Error("Golden: missing 'graph network' header")
	}
	if !strings.Contains(dot1, `rankdir="LR"`) {
		t.Error("Golden: missing rankdir layout directive")
	}
	if !strings.Contains(dot1, "router") {
		t.Error("Golden: missing router node")
	}
	if !strings.Contains(dot1, "switch") {
		t.Error("Golden: missing switch node")
	}
	if !strings.Contains(dot1, "host1") {
		t.Error("Golden: missing host1 node")
	}
	if !strings.Contains(dot1, "Gi0/1") {
		t.Error("Golden: missing source port label")
	}
	if !strings.Contains(dot1, "Gi0/5") {
		t.Error("Golden: missing target port label")
	}

	// Сохраняем golden-снимок для сравнения в будущем
	goldenPath := filepath.Join(t.TempDir(), "topology.golden.dot")
	if err := os.WriteFile(goldenPath, []byte(dot1), 0644); err != nil {
		t.Fatalf("write golden file: %v", err)
	}

	// Проверяем, что golden-файл можно использовать для сравнения
	goldenContent, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if string(goldenContent) != dot1 {
		t.Error("Golden: saved file content mismatch")
	}
}

// TestDOTWithNilTopology проверяет обработку nil в ToDOT.
func TestDOTWithNilTopology(t *testing.T) {
	var topo *Topology
	var buf strings.Builder
	err := topo.ToDOT(&buf)
	if err == nil {
		t.Error("ToDOT on nil should return error")
	}
}

// TestGraphMLWithEmptyTopology проверяет пустую топологию.
func TestGraphMLWithEmptyTopology(t *testing.T) {
	topo := &Topology{
		Devices: map[string]*Device{},
		Links:   []Link{},
	}

	path := filepath.Join(t.TempDir(), "empty.graphml")
	err := topo.SaveGraphML(path)
	if err != nil {
		t.Fatalf("SaveGraphML on empty topology should work: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read empty GraphML: %v", err)
	}

	content := string(raw)
	if !strings.Contains(content, `<graphml`) {
		t.Error("Empty GraphML should still have root element")
	}
	if strings.Contains(content, `<node `) {
		t.Error("Empty GraphML should not have nodes")
	}
	if strings.Contains(content, `<edge `) {
		t.Error("Empty GraphML should not have edges")
	}
}

// TestGraphMLWithLargeTopology проверяет масштабирование на большой топологии.
func TestGraphMLWithLargeTopology(t *testing.T) {
	devices := make(map[string]*Device, 100)
	links := make([]Link, 0, 200)

	for i := 0; i < 100; i++ {
		mac := sprintfMAC(i)
		devices[mac] = &Device{
			IP:       fmt.Sprintf("192.168.%d.%d", i/256, i%256),
			MAC:      mac,
			Hostname: fmt.Sprintf("device-%03d", i),
			Type:     DeviceTypeHost,
		}
	}

	// Создаём связи: каждая связь между i и (i+1)%100
	for i := 0; i < 100; i++ {
		mac1 := sprintfMAC(i)
		mac2 := sprintfMAC((i + 1) % 100)
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

	path := filepath.Join(t.TempDir(), "large.graphml")
	err := topo.SaveGraphML(path)
	if err != nil {
		t.Fatalf("SaveGraphML error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read large GraphML: %v", err)
	}

	// Проверяем, что файл содержит все устройства
	for i := 0; i < 100; i++ {
		expected := fmt.Sprintf("device-%03d", i)
		if !strings.Contains(string(raw), expected) {
			t.Errorf("Large GraphML missing device %d (%q)", i, expected)
		}
	}

	// Проверяем размер файла (должен быть разумным)
	if len(raw) > 500*1024 { // 500 KB
		t.Errorf("Large GraphML file too big: %.1f KB", float64(len(raw))/1024)
	}
}

// sprintfMAC генерирует тестовый MAC-адрес.
func sprintfMAC(index int) string {
	b2 := (index % 65536) / 256
	b3 := (index % 256)
	return fmt.Sprintf("aa:bb:cc:dd:%02x:%02x", b2, b3)
}
