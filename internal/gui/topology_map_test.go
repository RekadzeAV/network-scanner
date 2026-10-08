package gui

import (
	"strings"
	"testing"

	"image/color"

	"network-scanner/internal/topology"
)

func TestLimitTopologyKeys_NoLimit(t *testing.T) {
	keys := []string{"a", "b", "c"}
	result, truncated := limitTopologyKeys(keys, 5)
	if len(result) != 3 {
		t.Errorf("expected 3 keys, got %d", len(result))
	}
	if truncated {
		t.Error("expected truncated=false when no limit needed")
	}
	if len(result) != len(keys) {
		t.Error("expected same length")
	}
}

func TestLimitTopologyKeys_Limit(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e"}
	result, truncated := limitTopologyKeys(keys, 3)
	if len(result) != 3 {
		t.Errorf("expected 3 keys, got %d", len(result))
	}
	if !truncated {
		t.Error("expected truncated=true when limit applied")
	}
	if result[2] != "c" {
		t.Errorf("expected 'c' at index 2, got %s", result[2])
	}
}

func TestLimitTopologyKeys_ZeroMax(t *testing.T) {
	keys := []string{"a", "b"}
	// max<=0 → возвращаем оригинальный срез (не обрезаем)
	result, _ := limitTopologyKeys(keys, 0)
	if len(result) != 2 {
		t.Errorf("expected 2 keys (no truncation for max=0), got %d", len(result))
	}
}

func TestLimitTopologyKeys_NilKeys(t *testing.T) {
	result, _ := limitTopologyKeys(nil, 5)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestMatchTopologyNodeFilter_NilDev(t *testing.T) {
	if matchTopologyNodeFilter(nil, "test", "") {
		t.Error("expected false for nil device")
	}
}

func TestMatchTopologyNodeFilter_EmptyQuery(t *testing.T) {
	dev := &topology.Device{IP: "192.168.1.1", Hostname: "router"}
	if !matchTopologyNodeFilter(dev, "", "") {
		t.Error("expected true for empty query")
	}
}

func TestMatchTopologyNodeFilter_ByIP(t *testing.T) {
	dev := &topology.Device{IP: "192.168.1.1", Hostname: "router"}
	if !matchTopologyNodeFilter(dev, "192.168", "") {
		t.Error("expected true for matching IP")
	}
}

func TestMatchTopologyNodeFilter_ByHostname(t *testing.T) {
	dev := &topology.Device{IP: "192.168.1.1", Hostname: "router"}
	if !matchTopologyNodeFilter(dev, "route", "") {
		t.Error("expected true for matching hostname")
	}
}

func TestMatchTopologyNodeFilter_ByMAC(t *testing.T) {
	dev := &topology.Device{MAC: "aa:bb:cc:dd:ee:ff"}
	if !matchTopologyNodeFilter(dev, "aa:bb", "") {
		t.Error("expected true for matching MAC")
	}
}

func TestMatchTopologyNodeFilter_NoMatch(t *testing.T) {
	dev := &topology.Device{IP: "192.168.1.1", Hostname: "router"}
	if matchTopologyNodeFilter(dev, "xyz", "") {
		t.Error("expected false for no match")
	}
}

func TestMatchTopologyNodeFilter_TypeFilter(t *testing.T) {
	dev := &topology.Device{Type: topology.DeviceTypeRouter}
	if !matchTopologyNodeFilter(dev, "", "router") {
		t.Error("expected true for matching type filter")
	}
}

func TestMatchTopologyNodeFilter_TypeFilterNoMatch(t *testing.T) {
	dev := &topology.Device{Type: topology.DeviceTypeSwitch}
	if matchTopologyNodeFilter(dev, "", "router") {
		t.Error("expected false for non-matching type filter")
	}
}

func TestMatchTopologyNodeFilter_TypeFilterAll(t *testing.T) {
	dev := &topology.Device{Type: topology.DeviceTypeRouter}
	if !matchTopologyNodeFilter(dev, "", "all") {
		t.Error("expected true for 'all' type filter")
	}
}

func TestMatchTopologyNodeFilter_CaseInsensitive(t *testing.T) {
	dev := &topology.Device{Hostname: "ROUTER"}
	if !matchTopologyNodeFilter(dev, "route", "") {
		t.Error("expected case-insensitive match")
	}
}

func TestMatchTopologyLinkConfidence_EmptyFilter(t *testing.T) {
	l := topology.Link{}
	if !matchTopologyLinkConfidence(l, "") {
		t.Error("expected true for empty filter")
	}
}

func TestMatchTopologyLinkConfidence_AllFilter(t *testing.T) {
	l := topology.Link{Confidence: topology.LinkConfidenceHigh}
	if !matchTopologyLinkConfidence(l, "all") {
		t.Error("expected true for 'all' filter")
	}
}

func TestMatchTopologyLinkConfidence_Match(t *testing.T) {
	l := topology.Link{Confidence: topology.LinkConfidenceHigh}
	if !matchTopologyLinkConfidence(l, "high") {
		t.Error("expected true for matching confidence")
	}
}

func TestMatchTopologyLinkConfidence_NoMatch(t *testing.T) {
	l := topology.Link{Confidence: topology.LinkConfidenceHigh}
	if matchTopologyLinkConfidence(l, "low") {
		t.Error("expected false for non-matching confidence")
	}
}

func TestFindTopologyKeyByDevice_NilTopo(t *testing.T) {
	result := findTopologyKeyByDevice(nil, &topology.Device{IP: "1.2.3.4"})
	if result != "" {
		t.Errorf("expected empty for nil topo, got %q", result)
	}
}

func TestFindTopologyKeyByDevice_NilDev(t *testing.T) {
	topo := &topology.Topology{}
	result := findTopologyKeyByDevice(topo, nil)
	if result != "" {
		t.Errorf("expected empty for nil dev, got %q", result)
	}
}

func TestFindTopologyKeyByDevice_Found(t *testing.T) {
	dev := &topology.Device{IP: "192.168.1.1"}
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{
			"key1": dev,
			"key2": {IP: "192.168.1.2"},
		},
	}
	result := findTopologyKeyByDevice(topo, dev)
	if result != "key1" {
		t.Errorf("expected 'key1', got %q", result)
	}
}

func TestFindTopologyKeyByDevice_NotFound(t *testing.T) {
	dev := &topology.Device{IP: "10.0.0.1"}
	topo := &topology.Topology{
		Devices: map[string]*topology.Device{
			"key1": {IP: "192.168.1.1"},
		},
	}
	result := findTopologyKeyByDevice(topo, dev)
	if result != "" {
		t.Errorf("expected empty for not found device, got %q", result)
	}
}

func TestTopologyLinkKey_Basic(t *testing.T) {
	source := &topology.Device{IP: "192.168.1.1", Hostname: "router"}
	target := &topology.Device{IP: "192.168.1.2", Hostname: "switch"}
	portSrc := &topology.Port{Name: "eth0"}
	portTgt := &topology.Port{Name: "ge-0/0/0"}
	l := topology.Link{
		Source: source, Target: target,
		SourcePort: portSrc, TargetPort: portTgt,
	}
	key := topologyLinkKey(l)
	expected := "router|eth0|switch|ge-0/0/0"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestTopologyLinkKey_NilPorts(t *testing.T) {
	source := &topology.Device{IP: "192.168.1.1"}
	target := &topology.Device{IP: "192.168.1.2"}
	l := topology.Link{Source: source, Target: target}
	key := topologyLinkKey(l)
	if key == "" {
		t.Error("expected non-empty key for nil ports")
	}
}

func TestLinkBadge_Basic(t *testing.T) {
	l := topology.Link{
		SourceType: topology.LinkSourceLLDP,
		Confidence: topology.LinkConfidenceHigh,
	}
	badge := linkBadge(l)
	if badge != "LLDP/high" {
		t.Errorf("expected 'LLDP/high', got %q", badge)
	}
}

func TestLinkBadge_EmptySourceType(t *testing.T) {
	l := topology.Link{Confidence: topology.LinkConfidenceMedium}
	badge := linkBadge(l)
	if badge != "LINK/medium" {
		t.Errorf("expected 'LINK/medium', got %q", badge)
	}
}

func TestLinkBadge_EmptyConfidence(t *testing.T) {
	l := topology.Link{SourceType: topology.LinkSourceLLDP}
	badge := linkBadge(l)
	if badge != "LLDP/n/a" {
		t.Errorf("expected 'LLDP/n/a', got %q", badge)
	}
}

func TestLinkSummary_Basic(t *testing.T) {
	source := &topology.Device{Hostname: "router", IP: "192.168.1.1"}
	target := &topology.Device{Hostname: "switch", IP: "192.168.1.2"}
	l := topology.Link{
		Source: source, Target: target,
		SourceType: topology.LinkSourceLLDP,
		Confidence: topology.LinkConfidenceHigh,
		Evidence:   "arp-table",
	}
	summary := linkSummary(l)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	if !strings.Contains(summary, "router") || !strings.Contains(summary, "switch") {
		t.Errorf("expected summary to contain hostnames, got %q", summary)
	}
}

func TestLinkSummary_EmptyEvidence(t *testing.T) {
	source := &topology.Device{IP: "192.168.1.1"}
	target := &topology.Device{IP: "192.168.1.2"}
	l := topology.Link{Source: source, Target: target}
	summary := linkSummary(l)
	if !strings.Contains(summary, "n/a") {
		t.Errorf("expected 'n/a' in summary, got %q", summary)
	}
}

func TestLinkSummary_PortNames(t *testing.T) {
	source := &topology.Device{Hostname: "r"}
	target := &topology.Device{Hostname: "s"}
	l := topology.Link{
		Source: source, Target: target,
		SourcePort: &topology.Port{Name: "fa0/1"},
		TargetPort: &topology.Port{Name: "gi0/0"},
	}
	summary := linkSummary(l)
	if !strings.Contains(summary, "fa0/1") || !strings.Contains(summary, "gi0/0") {
		t.Errorf("expected port names in summary, got %q", summary)
	}
}

// TestTopologyMapColorsComeFromTokens — цвета графа обязаны приходить из
// токенов theme_tokens.go, а не из литералов: §13 запрещает color.RGBA{...}
// вне токенов, а «Material»-оттенки не проходили WCAG 1.4.11 (контраст >=3:1).
func TestTopologyMapColorsComeFromTokens(t *testing.T) {
	tests := []struct {
		name string
		got  color.Color
		want color.RGBA
	}{
		{name: "связь high", got: colorByConfidence(topology.LinkConfidenceHigh), want: tokTopoConfHigh},
		{name: "связь medium", got: colorByConfidence(topology.LinkConfidenceMedium), want: tokTopoConfMedium},
		{name: "связь low", got: colorByConfidence(topology.LinkConfidenceLow), want: tokTopoConfLow},
		{name: "узел router", got: colorByDeviceType(topology.DeviceTypeRouter), want: tokTopoNodeRouter},
		{name: "узел switch", got: colorByDeviceType(topology.DeviceTypeSwitch), want: tokTopoNodeSwitch},
		{name: "узел host", got: colorByDeviceType(topology.DeviceTypeHost), want: tokTopoNodeHost},
		{name: "узел без типа", got: colorByDeviceType(topology.DeviceType("unknown")), want: tokTopoNodeOther},
		{name: "обводка выбранного", got: colorByNodeBorder(true), want: tokTopoNodeBorderSelect},
		{name: "обводка обычного", got: colorByNodeBorder(false), want: tokTopoNodeBorder},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.got.(color.RGBA)
			if !ok {
				t.Fatalf("цвет не color.RGBA: %T", tc.got)
			}
			if got != tc.want {
				t.Fatalf("цвет = %v, ожидали токен %v", got, tc.want)
			}
		})
	}
}

// TestTopologyMapColorsDistinguishable — состояния различимы не только цветом
// между собой (WCAG 1.4.1 для набора графических состояний).
func TestTopologyMapColorsDistinguishable(t *testing.T) {
	conf := []color.RGBA{tokTopoConfHigh, tokTopoConfMedium, tokTopoConfLow}
	for i := 0; i < len(conf); i++ {
		for j := i + 1; j < len(conf); j++ {
			if conf[i] == conf[j] {
				t.Fatalf("уверенности %d и %d имеют одинаковый цвет %v", i, j, conf[i])
			}
		}
	}
	nodes := []color.RGBA{tokTopoNodeRouter, tokTopoNodeSwitch, tokTopoNodeHost, tokTopoNodeOther}
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i] == nodes[j] {
				t.Fatalf("типы узлов %d и %d имеют одинаковый цвет %v", i, j, nodes[i])
			}
		}
	}
}

func TestTopoDisplayName_Hostname(t *testing.T) {
	dev := &topology.Device{Hostname: "router"}
	name := topoDisplayName(dev)
	if name != "router" {
		t.Errorf("expected 'router', got %q", name)
	}
}

func TestTopoDisplayName_IP(t *testing.T) {
	dev := &topology.Device{IP: "192.168.1.1"}
	name := topoDisplayName(dev)
	if name != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got %q", name)
	}
}

func TestTopoDisplayName_MAC(t *testing.T) {
	dev := &topology.Device{MAC: "aa:bb:cc:dd:ee:ff"}
	name := topoDisplayName(dev)
	if name != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected MAC, got %q", name)
	}
}

func TestTopoDisplayName_Empty(t *testing.T) {
	dev := &topology.Device{}
	name := topoDisplayName(dev)
	if name != "unknown" {
		t.Errorf("expected 'unknown', got %q", name)
	}
}

func TestTopoDisplayName_Nil(t *testing.T) {
	name := topoDisplayName(nil)
	if name != "unknown" {
		t.Errorf("expected 'unknown' for nil device, got %q", name)
	}
}

func TestTopoDisplayName_IPOverMAC(t *testing.T) {
	dev := &topology.Device{IP: "10.0.0.1", MAC: "aa:bb:cc"}
	name := topoDisplayName(dev)
	if name != "10.0.0.1" {
		t.Error("expected IP priority over MAC")
	}
}

func TestTopoPortName_Name(t *testing.T) {
	port := &topology.Port{Name: "eth0"}
	name := topoPortName(port)
	if name != "eth0" {
		t.Errorf("expected 'eth0', got %q", name)
	}
}

func TestTopoPortName_Index(t *testing.T) {
	port := &topology.Port{Index: 42}
	name := topoPortName(port)
	if name != "if42" {
		t.Errorf("expected 'if42', got %q", name)
	}
}

func TestTopoPortName_Empty(t *testing.T) {
	port := &topology.Port{}
	name := topoPortName(port)
	if name != "-" {
		t.Errorf("expected '-', got %q", name)
	}
}

func TestTopoPortName_Nil(t *testing.T) {
	name := topoPortName(nil)
	if name != "-" {
		t.Errorf("expected '-' for nil port, got %q", name)
	}
}

func TestMatchTopologyNodeFilter_IPExact(t *testing.T) {
	dev := &topology.Device{IP: "192.168.1.1"}
	if !matchTopologyNodeFilter(dev, "192.168.1.1", "") {
		t.Error("expected exact IP match")
	}
}

func TestMatchTopologyNodeFilter_SubnetNoMatch(t *testing.T) {
	dev := &topology.Device{IP: "192.168.1.1"}
	if matchTopologyNodeFilter(dev, "10.0.0", "") {
		t.Error("expected no match for different subnet")
	}
}
