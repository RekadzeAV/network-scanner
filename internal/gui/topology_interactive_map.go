package gui

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/topology"
)

type topologyMapState struct {
	selectedNodeKey string
	selectedLinkKey string
	query           string
	typeFilter      string
	confFilter      string
}

const (
	topologyMapMaxNodes = 80
	topologyMapMaxLinks = 160
)

func (a *App) renderTopologyInteractiveMap(topo *topology.Topology) {
	if a == nil || a.topologyGraphBox == nil || a.topologyGraphStatus == nil {
		return
	}
	if topo == nil || len(topo.Devices) == 0 {
		a.topologyGraphBox.Objects = []fyne.CanvasObject{
			widget.NewLabel("Нет данных для интерактивной карты"),
		}
		a.topologyGraphBox.Refresh()
		a.topologyGraphStatus.SetText("Интерактивная карта: нет устройств")
		a.topologyGraphStatus.Refresh()
		return
	}

	keys := make([]string, 0, len(topo.Devices))
	for k := range topo.Devices {
		dev := topo.Devices[k]
		if !matchTopologyNodeFilter(dev, a.topologyViewState.query, a.topologyViewState.typeFilter) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	totalFilteredNodes := len(keys)
	keys, nodesTrimmed := limitTopologyKeys(keys, topologyMapMaxNodes)
	if len(keys) == 0 {
		a.topologyGraphBox.Objects = []fyne.CanvasObject{
			widget.NewLabel("По текущим фильтрам узлы не найдены"),
		}
		a.topologyGraphBox.Refresh()
		a.topologyGraphStatus.SetText("Интерактивная карта: фильтр вернул 0 узлов")
		a.topologyGraphStatus.Refresh()
		return
	}

	width := float32(math.Max(1200, float64(280+len(keys)*95)))
	height := float32(math.Max(780, float64(520+len(keys)*34)))
	cx := width / 2
	cy := height / 2
	radius := float32(math.Min(float64(width), float64(height))*0.35)
	if radius < 200 {
		radius = 200
	}

	nodePos := make(map[string]fyne.Position, len(keys))
	for i, key := range keys {
		angle := (2 * math.Pi * float64(i)) / float64(len(keys))
		x := cx + radius*float32(math.Cos(angle))
		y := cy + radius*float32(math.Sin(angle))
		nodePos[key] = fyne.NewPos(x, y)
	}

	objects := make([]fyne.CanvasObject, 0, len(keys)*2+len(topo.Links)*2+2)

	renderedLinks := 0
	linksSkippedByLimit := 0
	for _, l := range topo.Links {
		if !matchTopologyLinkConfidence(l, a.topologyViewState.confFilter) {
			continue
		}
		sourceKey := findTopologyKeyByDevice(topo, l.Source)
		targetKey := findTopologyKeyByDevice(topo, l.Target)
		if sourceKey == "" || targetKey == "" || sourceKey == targetKey {
			continue
		}
		p1, ok1 := nodePos[sourceKey]
		p2, ok2 := nodePos[targetKey]
		if !ok1 || !ok2 {
			continue
		}
		if renderedLinks >= topologyMapMaxLinks {
			linksSkippedByLimit++
			continue
		}
		renderedLinks++
		line := canvas.NewLine(colorByConfidence(l.Confidence))
		line.StrokeWidth = 2
		line.Position1 = p1
		line.Position2 = p2
		objects = append(objects, line)

		linkKey := topologyLinkKey(l)
		isSelectedLink := strings.TrimSpace(a.topologyViewState.selectedLinkKey) == strings.TrimSpace(linkKey)
		midX := (p1.X + p2.X) / 2
		midY := (p1.Y + p2.Y) / 2
		linkBtn := widget.NewButton(linkBadge(l), func() {
			a.topologyViewState.selectedLinkKey = linkKey
			a.topologyViewState.selectedNodeKey = sourceKey
			a.topologyGraphStatus.SetText(fmt.Sprintf("Связь: %s", linkSummary(l)))
			a.topologyGraphStatus.Refresh()
			if l.Source != nil {
				a.selectHostByTopologyDevice(l.Source)
			}
			a.renderTopologyInteractiveMap(topo)
		})
		linkBtn.Move(fyne.NewPos(midX-42, midY-12))
		linkBtn.Resize(fyne.NewSize(84, 24))
		if isSelectedLink {
			highlight := canvas.NewRectangle(color.RGBA{R: 210, G: 230, B: 255, A: 170})
			highlight.CornerRadius = 4
			highlight.Move(fyne.NewPos(midX-45, midY-14))
			highlight.Resize(fyne.NewSize(90, 28))
			objects = append(objects, highlight)
		}
		objects = append(objects, linkBtn)
	}

	for _, key := range keys {
		dev := topo.Devices[key]
		if dev == nil {
			continue
		}
		p := nodePos[key]
		nodeColor := colorByDeviceType(dev.Type)
		circle := canvas.NewCircle(nodeColor)
		circle.StrokeColor = colorByNodeBorder(key == a.topologyViewState.selectedNodeKey)
		circle.StrokeWidth = 2
		circle.Resize(fyne.NewSize(20, 20))
		circle.Move(fyne.NewPos(p.X-10, p.Y-10))
		objects = append(objects, circle)

		display := topoDisplayName(dev)
		if len(display) > 24 {
			display = display[:24] + "..."
		}
		btn := widget.NewButton(display, func() {
			a.topologyViewState.selectedNodeKey = key
			a.topologyGraphStatus.SetText(fmt.Sprintf("Выбран узел: %s", topoDisplayName(dev)))
			a.topologyGraphStatus.Refresh()
			a.selectHostByTopologyDevice(dev)
			a.renderTopologyInteractiveMap(topo)
		})
		btn.Resize(fyne.NewSize(190, 32))
		btn.Move(fyne.NewPos(p.X-95, p.Y+12))
		objects = append(objects, btn)
	}

	legend := widget.NewLabel("Легенда: router/switch/host; связи high/medium/low")
	legend.Move(fyne.NewPos(16, 12))
	legend.Resize(fyne.NewSize(480, 20))
	objects = append(objects, legend)
	if nodesTrimmed || linksSkippedByLimit > 0 {
		warn := widget.NewLabel(fmt.Sprintf("Режим упрощения: показано узлов %d/%d, связей %d (+%d скрыто)",
			len(keys), totalFilteredNodes, renderedLinks, linksSkippedByLimit))
		warn.Move(fyne.NewPos(16, 36))
		warn.Resize(fyne.NewSize(860, 20))
		objects = append(objects, warn)
	}

	a.topologyGraphBox.Objects = objects
	a.topologyGraphBox.Resize(fyne.NewSize(width, height))
	a.topologyGraphBox.Refresh()
	if a.topologyGraphScroll != nil {
		a.topologyGraphScroll.Refresh()
	}
	status := fmt.Sprintf("Интерактивная карта: узлов %d, связей %d", len(keys), renderedLinks)
	if nodesTrimmed || linksSkippedByLimit > 0 {
		status = fmt.Sprintf("%s (упрощенный режим для больших графов)", status)
	}
	a.topologyGraphStatus.SetText(status)
	a.topologyGraphStatus.Refresh()
}

func limitTopologyKeys(keys []string, max int) ([]string, bool) {
	if max <= 0 || len(keys) <= max {
		return keys, false
	}
	return keys[:max], true
}

func matchTopologyNodeFilter(dev *topology.Device, query string, typeFilter string) bool {
	if dev == nil {
		return false
	}
	tf := strings.TrimSpace(strings.ToLower(typeFilter))
	if tf != "" && tf != "all" && strings.ToLower(string(dev.Type)) != tf {
		return false
	}
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return true
	}
	values := []string{
		strings.ToLower(strings.TrimSpace(dev.IP)),
		strings.ToLower(strings.TrimSpace(dev.MAC)),
		strings.ToLower(strings.TrimSpace(dev.Hostname)),
	}
	for _, v := range values {
		if strings.Contains(v, q) {
			return true
		}
	}
	return false
}

func matchTopologyLinkConfidence(l topology.Link, confidenceFilter string) bool {
	cf := strings.TrimSpace(strings.ToLower(confidenceFilter))
	if cf == "" || cf == "all" {
		return true
	}
	return strings.ToLower(strings.TrimSpace(string(l.Confidence))) == cf
}

func findTopologyKeyByDevice(topo *topology.Topology, d *topology.Device) string {
	if topo == nil || d == nil {
		return ""
	}
	for key, item := range topo.Devices {
		if item == d {
			return key
		}
	}
	return ""
}

func topologyLinkKey(l topology.Link) string {
	return fmt.Sprintf("%s|%s|%s|%s",
		topoDisplayName(l.Source),
		topoPortName(l.SourcePort),
		topoDisplayName(l.Target),
		topoPortName(l.TargetPort),
	)
}

func linkBadge(l topology.Link) string {
	source := strings.ToUpper(strings.TrimSpace(string(l.SourceType)))
	conf := strings.TrimSpace(string(l.Confidence))
	if source == "" {
		source = "LINK"
	}
	if conf == "" {
		conf = "n/a"
	}
	return fmt.Sprintf("%s/%s", source, conf)
}

func linkSummary(l topology.Link) string {
	evidence := strings.TrimSpace(l.Evidence)
	if evidence == "" {
		evidence = "n/a"
	}
	return fmt.Sprintf("%s (%s) <-> %s (%s), %s/%s, evidence=%s",
		topoDisplayName(l.Source),
		topoPortName(l.SourcePort),
		topoDisplayName(l.Target),
		topoPortName(l.TargetPort),
		strings.TrimSpace(string(l.SourceType)),
		strings.TrimSpace(string(l.Confidence)),
		evidence,
	)
}

func colorByConfidence(c topology.LinkConfidence) color.Color {
	switch c {
	case topology.LinkConfidenceHigh:
		return color.RGBA{R: 52, G: 168, B: 83, A: 255}
	case topology.LinkConfidenceMedium:
		return color.RGBA{R: 251, G: 188, B: 4, A: 255}
	default:
		return color.RGBA{R: 234, G: 67, B: 53, A: 255}
	}
}

func colorByDeviceType(t topology.DeviceType) color.Color {
	switch t {
	case topology.DeviceTypeRouter:
		return color.RGBA{R: 66, G: 133, B: 244, A: 220}
	case topology.DeviceTypeSwitch:
		return color.RGBA{R: 52, G: 168, B: 83, A: 220}
	case topology.DeviceTypeHost:
		return color.RGBA{R: 251, G: 188, B: 4, A: 220}
	default:
		return color.RGBA{R: 120, G: 120, B: 120, A: 220}
	}
}

func colorByNodeBorder(selected bool) color.Color {
	if selected {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}
	return color.RGBA{R: 60, G: 60, B: 60, A: 180}
}

func (a *App) selectHostByTopologyDevice(d *topology.Device) {
	if a == nil || d == nil {
		return
	}
	ip := strings.TrimSpace(d.IP)
	if ip == "" {
		return
	}
	a.selectedHostIP = ip
	a.renderScanResultsView()
}
