package gui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/topology"
)

// topologyMapState — состояние интерактивной карты.
// H-07: zoom/panX/panY участвуют в расчёте раскладки (а не в пост-трансформации
// уже созданных объектов), поэтому повторные перерисовки не накапливают ошибку.
// Пустое (нулевое) zoom трактуется как масштаб по умолчанию.
type topologyMapState struct {
	selectedNodeKey string
	selectedLinkKey string
	query           string
	typeFilter      string
	confFilter      string
	zoom            float32
	panX            float32
	panY            float32
}

// effectiveZoom возвращает масштаб с учётом «пустого» состояния.
func (s topologyMapState) effectiveZoom() float32 {
	if s.zoom == 0 {
		return topoZoomDefault
	}
	return clampTopologyZoom(s.zoom)
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
			buildResultsEmptyState(
				"Нет данных топологии",
				"Сначала выполните сканирование, затем нажмите «Построить топологию».",
				iconTopology(),
				"Построить топологию",
				func() {
					if a.topoCtrl != nil {
						a.topoCtrl.BuildTopology(a.scanResults, a.myWindow)
					}
				},
			),
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

	// H-08: слоистая раскладка (ядро → коммутаторы → хосты) вместо одного круга.
	// H-07: базовые позиции масштабируются zoom и смещаются pan перед отрисовкой.
	basePos, baseW, baseH := topologyComputeLayout(keys, topo)
	zoom := a.topologyViewState.effectiveZoom()
	width := baseW * zoom
	height := baseH * zoom

	nodePos := make(map[string]fyne.Position, len(basePos))
	for k, p := range basePos {
		nodePos[k] = topologyApplyView(p, baseW, baseH, zoom, a.topologyViewState.panX, a.topologyViewState.panY)
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
		line.StrokeWidth = topoLinkStroke(zoom)
		line.Position1 = p1
		line.Position2 = p2
		objects = append(objects, line)

		linkKey := topologyLinkKey(l)
		isSelectedLink := strings.TrimSpace(a.topologyViewState.selectedLinkKey) == strings.TrimSpace(linkKey)
		midX := (p1.X + p2.X) / 2
		midY := (p1.Y + p2.Y) / 2

		// H-02: невидимый hit-area вдоль связи. Ширина не уменьшается при
		// отдалении карты — минимум WCAG 2.5.8 (H-05).
		hitW := topoLinkHitWidth(zoom)
		hitPos, hitSize := linkHitAreaSize(p1, p2, hitW)
		hitRect := newLinkHitArea(p1, p2, hitW, func() {
			a.selectTopologyLink(topo, l, linkKey, sourceKey)
		})
		hitRect.Move(hitPos)
		hitRect.Resize(hitSize)
		objects = append(objects, hitRect)

		linkBtn := widget.NewButton(linkBadge(l), func() {
			a.selectTopologyLink(topo, l, linkKey, sourceKey)
		})
		badgeW := float32(84) * zoom
		badgeH := float32(24) * zoom
		linkBtn.Move(fyne.NewPos(midX-badgeW/2, midY-badgeH/2))
		linkBtn.Resize(fyne.NewSize(badgeW, badgeH))
		if isSelectedLink {
			highlight := canvas.NewRectangle(tokTopoLinkHighlight)
			highlight.CornerRadius = 4
			highlight.Move(fyne.NewPos(midX-badgeW/2-3*zoom, midY-badgeH/2-3*zoom))
			highlight.Resize(fyne.NewSize(badgeW+6*zoom, badgeH+6*zoom))
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
		isSelected := key == a.topologyViewState.selectedNodeKey

		nodeKey := key
		nodeDev := dev
		onTap := func() {
			a.topologyViewState.selectedNodeKey = nodeKey
			a.topologyGraphStatus.SetText(fmt.Sprintf("Выбран узел: %s", topoDisplayName(nodeDev)))
			a.topologyGraphStatus.Refresh()
			a.selectHostByTopologyDevice(nodeDev)
			a.renderTopologyInteractiveMap(topo)
		}

		tn := newTappableNode(nodeColor, colorByNodeBorder(isSelected), 2, 10*zoom, topoHitNodeRadius(zoom), onTap)
		hit := topoHitNodeRadius(zoom)
		tn.Move(fyne.NewPos(p.X-hit, p.Y-hit))
		objects = append(objects, tn)

		display := truncateString(topoDisplayName(dev), 24)
		btn := widget.NewButton(display, onTap)
		labelW := float32(190) * zoom
		labelH := float32(32) * zoom
		btn.Resize(fyne.NewSize(labelW, labelH))
		btn.Move(fyne.NewPos(p.X-labelW/2, p.Y+hit))
		objects = append(objects, btn)
	}

	legend := widget.NewLabel("Легенда: router/switch/host; связи high/medium/low. Колесо мыши — масштаб, перетаскивание — панорама")
	legend.Move(fyne.NewPos(16, 12))
	legend.Resize(fyne.NewSize(720, 20))
	objects = append(objects, legend)
	if nodesTrimmed || linksSkippedByLimit > 0 {
		warn := widget.NewLabel(fmt.Sprintf("Режим упрощения: показано узлов %d/%d, связей %d (+%d скрыто)",
			len(keys), totalFilteredNodes, renderedLinks, linksSkippedByLimit))
		warn.Move(fyne.NewPos(16, 36))
		warn.Resize(fyne.NewSize(860, 20))
		objects = append(objects, warn)
	}

	a.topologyGraphBox.Objects = objects

	// H-07: размер мира (worldSize) задаётся через viewport, а не через
	// Resize на graphBox — scroll container использует MinSize viewport'а
	// для расчёта скролл-баров.
	if a.topologyViewport != nil {
		a.topologyViewport.SetWorldSize(fyne.NewSize(width, height))
	}
	a.topologyGraphBox.Refresh()
	if a.topologyGraphScroll != nil {
		a.topologyGraphScroll.SetMinSize(fyne.NewSize(width, height))
		a.topologyGraphScroll.Refresh()
	}
	status := fmt.Sprintf("Интерактивная карта: узлов %d, связей %d, масштаб %d%%", len(keys), renderedLinks, int(zoom*100+0.5))
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

// colorByConfidence — цвет связи по уверенности. Значения берутся из токенов
// (theme_tokens.go): «Material»-оттенки #34A853/#FBBC04/#EA4335 не проходили
// WCAG 1.4.11 (жёлтый на белом фоне давал 1.7:1 при норме >=3:1), поэтому
// используются затемнённые пары токенов.
func colorByConfidence(c topology.LinkConfidence) color.Color {
	switch c {
	case topology.LinkConfidenceHigh:
		return tokTopoConfHigh
	case topology.LinkConfidenceMedium:
		return tokTopoConfMedium
	default:
		return tokTopoConfLow
	}
}

// colorByDeviceType — заливка узла графа по типу устройства (токены варианта темы).
func colorByDeviceType(t topology.DeviceType) color.Color {
	switch t {
	case topology.DeviceTypeRouter:
		return tokTopoNodeRouter
	case topology.DeviceTypeSwitch:
		return tokTopoNodeSwitch
	case topology.DeviceTypeHost:
		return tokTopoNodeHost
	default:
		return tokTopoNodeOther
	}
}

// colorByNodeBorder — обводка узла: выбранная выделяется сильнее обычной.
func colorByNodeBorder(selected bool) color.Color {
	if selected {
		return tokTopoNodeBorderSelect
	}
	return tokTopoNodeBorder
}

// selectTopologyLink выделяет связь карты и показывает её сводку (общий код
// для клика по hit-зоне и по бейджу-кнопке).
func (a *App) selectTopologyLink(topo *topology.Topology, l topology.Link, linkKey, sourceKey string) {
	if a == nil {
		return
	}
	a.topologyViewState.selectedLinkKey = linkKey
	a.topologyViewState.selectedNodeKey = sourceKey
	if a.topologyGraphStatus != nil {
		a.topologyGraphStatus.SetText(fmt.Sprintf("Связь: %s", linkSummary(l)))
		a.topologyGraphStatus.Refresh()
	}
	if l.Source != nil {
		a.selectHostByTopologyDevice(l.Source)
	}
	a.renderTopologyInteractiveMap(topo)
}

// setTopologyZoom применяет масштаб к карте и перерисовывает её (H-07).
// Возвращает фактический масштаб с учётом ограничения.
func (a *App) setTopologyZoom(zoom float32) float32 {
	if a == nil {
		return topoZoomDefault
	}
	z := clampTopologyZoom(zoom)
	a.topologyViewState.zoom = z
	if a.lastTopology != nil {
		a.renderTopologyInteractiveMap(a.lastTopology)
	}
	return z
}

// zoomTopologyBy шагает масштаб на topoZoomStep в указанном направлении (H-07).
func (a *App) zoomTopologyBy(direction float32) {
	if a == nil {
		return
	}
	a.setTopologyZoom(topologyZoomBy(a.topologyViewState.effectiveZoom(), direction))
}

// panTopologyBy смещает карту (панорама) с ограничением по размеру мира (H-07).
func (a *App) panTopologyBy(dx, dy float32) {
	if a == nil || a.lastTopology == nil {
		return
	}
	_, baseW, baseH := topologyComputeLayout(topologyVisibleKeys(a.lastTopology, a.topologyViewState), a.lastTopology)
	a.topologyViewState.panX = clampTopologyPan(a.topologyViewState.panX+dx, baseW)
	a.topologyViewState.panY = clampTopologyPan(a.topologyViewState.panY+dy, baseH)
	a.renderTopologyInteractiveMap(a.lastTopology)
}

// resetTopologyView сбрасывает зум и панораму карты (H-07).
func (a *App) resetTopologyView() {
	if a == nil {
		return
	}
	a.topologyViewState.zoom = topoZoomDefault
	a.topologyViewState.panX = 0
	a.topologyViewState.panY = 0
	if a.topologyGraphScroll != nil {
		a.topologyGraphScroll.ScrollToOffset(fyne.Position{})
	}
	if a.lastTopology != nil {
		a.renderTopologyInteractiveMap(a.lastTopology)
	}
}

// topologyVisibleKeys — ключи узлов, проходящие текущие фильтры (в том же
// порядке, что и при отрисовке). Используется для ограничения панорамы.
func topologyVisibleKeys(topo *topology.Topology, st topologyMapState) []string {
	if topo == nil {
		return nil
	}
	keys := make([]string, 0, len(topo.Devices))
	for k := range topo.Devices {
		if !matchTopologyNodeFilter(topo.Devices[k], st.query, st.typeFilter) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	keys, _ = limitTopologyKeys(keys, topologyMapMaxNodes)
	return keys
}

// selectHostByTopologyDevice выделяет хост из карты топологии.
// H-09: выбор не должен быть «молчаливым» — даём обратную связь в глобальном
// статус-баре (он виден на всех вкладках), поскольку детали хоста показываются
// на вкладке «Сканирование».
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
	a.setStatusToast("Хост " + ip + " выбран на карте — детали в разделе «Детали хоста»")
}

// topoDisplayName и topoPortName — вспомогательные функции для отображения
func topoDisplayName(d *topology.Device) string {
	if d == nil {
		return "unknown"
	}
	if d.Hostname != "" {
		return d.Hostname
	}
	if d.IP != "" {
		return d.IP
	}
	if d.MAC != "" {
		return d.MAC
	}
	return "unknown"
}

func topoPortName(p *topology.Port) string {
	if p == nil {
		return "-"
	}
	if p.Name != "" {
		return p.Name
	}
	if p.Index > 0 {
		return fmt.Sprintf("if%d", p.Index)
	}
	return "-"
}
