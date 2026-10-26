package gui

import (
	"fyne.io/fyne/v2"

	"network-scanner/internal/topology"
)

// H-08: слоистая раскладка карты топологии.
//
// Прежняя раскладка ставила все узлы на одну окружность независимо от структуры
// сети: связи шли хордами через весь круг, иерархия не читалась. Здесь узлы
// раскладываются ярусами по роли:
//
//	y=0  ядро      — маршрутизаторы (L3)
//	y=1  доступ    — коммутаторы (L2)
//	y=2  конечные   — хосты и неизвестные
//
// Внутри яруса узлы распределяются по горизонтали, при нехватке места ярус
// разбивается на несколько строк. Холст при этом остаётся широким и прокручивается,
// а зум (H-07) позволяет уместить крупный граф целиком.

const (
	topoLayerCore  = 0 // маршрутизаторы — ядро
	topoLayerDist  = 1 // коммутаторы — уровень доступа
	topoLayerHost  = 2 // хосты и неизвестные
	topoLayerCount = 3
)

const (
	topoLayerMinWidth  = float32(1100) // минимальная ширина мира
	topoLayerMaxWidth  = float32(2600) // максимальная ширина мира
	topoLayerPitch     = float32(170)  // горизонтальный шаг узлов
	topoLayerRowHeight = float32(150)  // вертикальный шаг строк внутри яруса
	topoLayerBandGap   = float32(110)  // зазор между ярусами
	topoLayerMargin    = float32(100)  // боковые поля
	topoLayerTopPad    = float32(80)   // отступ сверху (под легенду)
	topoLayerBottomPad = float32(60)   // отступ снизу
	topoLayerMaxPerRow = 12            // максимум узлов в строке яруса
)

// topologyNodeLayer возвращает ярус раскладки по типу устройства.
func topologyNodeLayer(dev *topology.Device) int {
	if dev == nil {
		return topoLayerHost
	}
	switch dev.Type {
	case topology.DeviceTypeRouter:
		return topoLayerCore
	case topology.DeviceTypeSwitch:
		return topoLayerDist
	default:
		return topoLayerHost
	}
}

// topologyLayerRows — число строк, которое займут count узлов яруса.
func topologyLayerRows(count int) int {
	if count <= 0 {
		return 0
	}
	rows := count / topoLayerMaxPerRow
	if count%topoLayerMaxPerRow != 0 {
		rows++
	}
	return rows
}

// topologyComputeLayout раскладывает узлы по ярусам и возвращает базовые
// позиции вместе с размером «мира» до применения зума/панорамы.
func topologyComputeLayout(keys []string, topo *topology.Topology) (map[string]fyne.Position, float32, float32) {
	pos := make(map[string]fyne.Position, len(keys))
	if len(keys) == 0 {
		return pos, topoLayerMinWidth, float32(780)
	}

	// Группировка по ярусам (порядок внутри яруса = порядок keys, он отсортирован).
	layers := make([][]string, topoLayerCount)
	for _, k := range keys {
		var dev *topology.Device
		if topo != nil {
			dev = topo.Devices[k]
		}
		layers[topologyNodeLayer(dev)] = append(layers[topologyNodeLayer(dev)], k)
	}

	// Ширина мира определяется самым широким ярусом.
	widest := 0
	for _, l := range layers {
		if len(l) > widest {
			widest = len(l)
		}
	}
	perRow := widest
	if perRow > topoLayerMaxPerRow {
		perRow = topoLayerMaxPerRow
	}
	width := topoLayerMargin*2 + float32(perRow)*topoLayerPitch
	if width < topoLayerMinWidth {
		width = topoLayerMinWidth
	}
	if width > topoLayerMaxWidth {
		width = topoLayerMaxWidth
	}

	// Высота — сумма ярусов с их строками.
	height := topoLayerTopPad + topoLayerBottomPad
	bands := 0
	for _, l := range layers {
		rows := topologyLayerRows(len(l))
		if rows == 0 {
			continue
		}
		bands++
		height += float32(rows)*topoLayerRowHeight + topoLayerBandGap
	}
	if bands > 1 {
		height -= topoLayerBandGap // лишний зазор после последнего яруса
	}
	if height < 620 {
		height = 620
	}

	// Размещение: по ярусам сверху вниз, внутри яруса — строками по центру.
	usedPitch := topoLayerPitch
	if float32(perRow)*topoLayerPitch > width-topoLayerMargin*2 {
		usedPitch = (width - topoLayerMargin*2) / float32(perRow)
	}

	yCursor := topoLayerTopPad
	for _, l := range layers {
		rows := topologyLayerRows(len(l))
		if rows == 0 {
			continue
		}
		for row := 0; row < rows; row++ {
			start := row * topoLayerMaxPerRow
			end := start + topoLayerMaxPerRow
			if end > len(l) {
				end = len(l)
			}
			inRow := end - start
			// Строка центрируется по ширине мира.
			rowWidth := float32(inRow-1) * usedPitch
			xStart := (width - rowWidth) / 2
			y := yCursor + float32(row)*topoLayerRowHeight
			for i, k := range l[start:end] {
				pos[k] = fyne.NewPos(xStart+float32(i)*usedPitch, y)
			}
		}
		yCursor += float32(rows)*topoLayerRowHeight + topoLayerBandGap
	}

	return pos, width, height
}

// topologyApplyView пересчитывает базовую позицию мира в экранные координаты:
// масштабирование вокруг центра мира + смещение панорамы (H-07).
func topologyApplyView(p fyne.Position, baseW, baseH, zoom, panX, panY float32) fyne.Position {
	z := clampTopologyZoom(zoom)
	return fyne.NewPos(
		(p.X-baseW/2)*z+baseW*z/2+panX,
		(p.Y-baseH/2)*z+baseH*z/2+panY,
	)
}

// topoHitNodeMinRadius — нижняя граница hit-зоны узла (WCAG 2.5.8, 14 лог. px).
const topoHitNodeMinRadius = float32(14)

// topoHitNodeRadius возвращает радиус hit-зоны узла при масштабе zoom.
// При уменьшении карты зона не должна становиться меньше порога доступности,
// иначе по узлам невозможно попасть (H-05).
func topoHitNodeRadius(zoom float32) float32 {
	r := topoHitNodeMinRadius * clampTopologyZoom(zoom)
	if r < topoHitNodeMinRadius {
		return topoHitNodeMinRadius
	}
	return r
}

// topoLinkStroke — толщина линии связи при масштабе zoom (не тоньше 1 px).
func topoLinkStroke(zoom float32) float32 {
	s := float32(2) * clampTopologyZoom(zoom)
	if s < 1 {
		return 1
	}
	return s
}

// topoLinkHitMinWidth — минимальная ширина hit-зоны связи (WCAG 2.5.8).
const topoLinkHitMinWidth = float32(12)

// topoLinkHitWidth — ширина hit-зоны связи при масштабе zoom с нижним порогом.
func topoLinkHitWidth(zoom float32) float32 {
	w := topoLinkHitMinWidth * clampTopologyZoom(zoom)
	if w < topoLinkHitMinWidth {
		return topoLinkHitMinWidth
	}
	return w
}
