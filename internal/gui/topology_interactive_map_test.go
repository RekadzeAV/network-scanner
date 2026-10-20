package gui

import (
	"math"
	"strings"
	"testing"

	"network-scanner/internal/topology"

	"fyne.io/fyne/v2"
)

func TestMatchTopologyNodeFilter(t *testing.T) {
	dev := &topology.Device{
		IP:       "192.168.1.10",
		MAC:      "AA:BB:CC:DD:EE:FF",
		Hostname: "core-switch",
		Type:     topology.DeviceTypeSwitch,
	}
	if !matchTopologyNodeFilter(dev, "", "all") {
		t.Fatalf("expected empty query/all type to pass")
	}
	if !matchTopologyNodeFilter(dev, "core", "switch") {
		t.Fatalf("expected hostname query + type switch to pass")
	}
	if matchTopologyNodeFilter(dev, "router", "switch") {
		t.Fatalf("expected unmatched query to fail")
	}
	if matchTopologyNodeFilter(dev, "core", "router") {
		t.Fatalf("expected mismatched type to fail")
	}
}

func TestMatchTopologyLinkConfidence(t *testing.T) {
	link := topology.Link{
		Confidence: topology.LinkConfidenceMedium,
	}
	if !matchTopologyLinkConfidence(link, "all") {
		t.Fatalf("expected all filter to pass")
	}
	if !matchTopologyLinkConfidence(link, "medium") {
		t.Fatalf("expected medium filter to pass")
	}
	if matchTopologyLinkConfidence(link, "high") {
		t.Fatalf("expected high filter to fail for medium link")
	}
}

func TestLinkSummaryIncludesEvidence(t *testing.T) {
	link := topology.Link{
		Source:     &topology.Device{IP: "192.168.1.1", Hostname: "r1"},
		Target:     &topology.Device{IP: "192.168.1.2", Hostname: "sw1"},
		SourceType: topology.LinkSourceLLDP,
		Confidence: topology.LinkConfidenceHigh,
		Evidence:   "lldp_neighbor_match",
	}
	got := linkSummary(link)
	if !strings.Contains(got, "evidence=lldp_neighbor_match") {
		t.Fatalf("expected evidence in summary, got: %s", got)
	}
}

func TestLimitTopologyKeys(t *testing.T) {
	keys := []string{"a", "b", "c", "d"}
	limited, trimmed := limitTopologyKeys(keys, 2)
	if !trimmed {
		t.Fatalf("expected trimmed=true")
	}
	if len(limited) != 2 || limited[0] != "a" || limited[1] != "b" {
		t.Fatalf("unexpected limited keys: %#v", limited)
	}
}

// H-07: clampTopologyZoom — проверка граничных значений и NaN.
func TestClampTopologyZoom(t *testing.T) {
	tests := []struct {
		in   float32
		want float32
	}{
		{0, topoZoomDefault},
		{-1, topoZoomDefault},
		{float32(math.NaN()), topoZoomDefault},
		{0.2, topoZoomMin},
		{0.35, topoZoomMin},
		{1.0, topoZoomDefault},
		{3.0, topoZoomMax},
		{5.0, topoZoomMax},
	}
	for _, tt := range tests {
		got := clampTopologyZoom(tt.in)
		if got != tt.want {
			t.Errorf("clampTopologyZoom(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// H-07: topologyZoomBy — проверка шага зума с ограничением.
func TestTopologyZoomBy(t *testing.T) {
	// Приближение от 1.0
	got := topologyZoomBy(1.0, topoZoomStep)
	if got <= 1.0 {
		t.Errorf("expected zoom > 1.0, got %v", got)
	}
	// Отдаление от 1.0
	got = topologyZoomBy(1.0, -topoZoomStep)
	if got >= 1.0 {
		t.Errorf("expected zoom < 1.0, got %v", got)
	}
	// Приближение с потолка не превышает max
	got = topologyZoomBy(topoZoomMax, topoZoomStep)
	if got != topoZoomMax {
		t.Errorf("expected topoZoomMax, got %v", got)
	}
	// Направление 0 не меняет zoom
	got = topologyZoomBy(1.5, 0)
	if got != 1.5 {
		t.Errorf("expected unchanged zoom 1.5, got %v", got)
	}
}

// H-07: clampTopologyPan — проверка ограничения панорамы.
func TestClampTopologyPan(t *testing.T) {
	world := float32(1000)
	limit := world * topoPanClampRatio
	// В пределах лимита
	got := clampTopologyPan(100, world)
	if got != 100 {
		t.Errorf("expected 100, got %v", got)
	}
	// Выше лимита
	got = clampTopologyPan(2000, world)
	if got != limit {
		t.Errorf("expected %v, got %v", limit, got)
	}
	// Ниже лимита
	got = clampTopologyPan(-2000, world)
	if got != -limit {
		t.Errorf("expected %v, got %v", -limit, got)
	}
}

// H-08: topologyNodeLayer — проверка классификации слоёв.
func TestTopologyNodeLayer(t *testing.T) {
	tests := []struct {
		devType topology.DeviceType
		want    int
	}{
		{topology.DeviceTypeRouter, topoLayerCore},
		{topology.DeviceTypeSwitch, topoLayerDist},
		{topology.DeviceTypeHost, topoLayerHost},
		{topology.DeviceTypeUnknown, topoLayerHost},
	}
	for _, tt := range tests {
		dev := &topology.Device{Type: tt.devType}
		got := topologyNodeLayer(dev)
		if got != tt.want {
			t.Errorf("topologyNodeLayer(%v) = %d, want %d", tt.devType, got, tt.want)
		}
	}
}

// H-08: topologyComputeLayout — проверка структуры раскладки.
func TestTopologyComputeLayout(t *testing.T) {
	devRouter := &topology.Device{Type: topology.DeviceTypeRouter}
	devSwitch := &topology.Device{Type: topology.DeviceTypeSwitch}
	devHost := &topology.Device{Type: topology.DeviceTypeHost}

	topo := &topology.Topology{
		Devices: map[string]*topology.Device{
			"router": devRouter,
			"switch": devSwitch,
			"host1":  devHost,
			"host2":  devHost,
		},
	}

	keys := []string{"router", "switch", "host1", "host2"}
	positions, baseW, baseH := topologyComputeLayout(keys, topo)

	// Проверка что позиции существуют для всех ключей
	if len(positions) != len(keys) {
		t.Fatalf("expected %d positions, got %d", len(keys), len(positions))
	}

	// Проверка что router (ядро) выше switch (дистрибуция)
	if positions["router"].Y >= positions["switch"].Y {
		t.Errorf("router Y (%v) should be less than switch Y (%v)", positions["router"].Y, positions["switch"].Y)
	}

	// Проверка что switch выше host
	if positions["switch"].Y >= positions["host1"].Y {
		t.Errorf("switch Y (%v) should be less than host Y (%v)", positions["switch"].Y, positions["host1"].Y)
	}

	// Проверка что размеры положительные
	if baseW <= 0 || baseH <= 0 {
		t.Errorf("expected positive baseW/baseH, got %v/%v", baseW, baseH)
	}

	// Проверка что хосты одного слоя имеют одинаковый Y
	if positions["host1"].Y != positions["host2"].Y {
		t.Errorf("host1 Y (%v) should equal host2 Y (%v)", positions["host1"].Y, positions["host2"].Y)
	}
}

// H-08: topologyApplyView — проверка применения зума/паны к позиции.
func TestTopologyApplyView(t *testing.T) {
	p := fyne.NewPos(100, 200)
	baseW, baseH := float32(1000), float32(800)
	zoom := float32(2.0)
	panX, panY := float32(50), float32(-30)

	got := topologyApplyView(p, baseW, baseH, zoom, panX, panY)

	// При zoom=2 позиция должна быть умножена на 2 и смещена на pan
	wantX := p.X*zoom + panX
	wantY := p.Y*zoom + panY

	if got.X != wantX || got.Y != wantY {
		t.Errorf("topologyApplyView(%v, %v/%v, zoom=%v, pan=%v/%v) = %v, want %v/%v",
			p, baseW, baseH, zoom, panX, panY, got, wantX, wantY)
	}
}

// H-07: topoHitNodeRadius — проверка что радиус hit-зоны не меньше минимума.
func TestTopoHitNodeRadius(t *testing.T) {
	for _, zoom := range []float32{0.1, 0.5, 1.0, 2.0, 5.0} {
		r := topoHitNodeRadius(zoom)
		if r < topoHitNodeMinRadius {
			t.Errorf("topoHitNodeRadius(%v) = %v, should be >= %v", zoom, r, topoHitNodeMinRadius)
		}
	}
}

// H-07: topoLinkHitWidth — проверка что hit-зона связи не меньше минимума.
func TestTopoLinkHitWidth(t *testing.T) {
	for _, zoom := range []float32{0.1, 0.5, 1.0, 2.0, 5.0} {
		w := topoLinkHitWidth(zoom)
		if w < topoLinkHitMinWidth {
			t.Errorf("topoLinkHitWidth(%v) = %v, should be >= %v", zoom, w, topoLinkHitMinWidth)
		}
	}
}
