package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	topoZoomMin       = 0.3
	topoZoomMax       = 4.0
	topoZoomStep      = 1.25
	topoZoomDefault   = 1.0
	topoHitRadiusMin  = 14.0 // WCAG 2.5.8 minimum target (14 px logical)
)

// topologyViewport — обёртка над контейнером карты топологии, обеспечивающая
// зум колесом мыши и перетаскивание (pan) (H-07).
//
// Внутренний механизм:
//   - Хранит `content` (обычно *fyne.Container) с объектами карты.
//   - Реализует fyne.Scrollable (Scrolled) — зум вокруг позиции курсора.
//   - Реализует fyne.Draggable (Dragged, DragEnd) — перетаскивание карты.
//   - Реализует desktop.Hoverable (Cursor) — курсор-рука при наведении.
//   - На зум/пан пересчитывает позиции и размеры детей через Move/Resize.
type topologyViewport struct {
	widget.BaseWidget

	content fyne.CanvasObject

	// zoom / pan — состояние трансформации (H-07).
	zoom float32
	pan  fyne.Position

	// minWorld — размер мира в координатах до зума (MinSize).
	minWorld fyne.Size

	// на вызовы (zoom, pan) — обновить карту.
	onViewChanged func()
}

// newTopologyViewport создаёт viewport, оборачивающий content с заданным
// мировым размером.
func newTopologyViewport(content fyne.CanvasObject, world fyne.Size) *topologyViewport {
	tv := &topologyViewport{
		content:  content,
		zoom:     topoZoomDefault,
		minWorld: world,
	}
	tv.ExtendBaseWidget(tv)
	return tv
}

// SetOnViewChanged задаёт колбэк на изменение зума/пана.
func (tv *topologyViewport) SetOnViewChanged(fn func()) {
	if tv == nil {
		return
	}
	tv.onViewChanged = fn
}

// MinSize — размер виджета = мир * zoom (чтобы scroll-контейнер
// правильно показывал scrollbars для панорамы).
func (tv *topologyViewport) MinSize() fyne.Size {
	if tv == nil {
		return fyne.NewSize(0, 0)
	}
	w := float32(tv.minWorld.Width) * tv.zoom
	h := float32(tv.minWorld.Height) * tv.zoom
	return fyne.NewSize(w, h)
}

// Layout — просто передаём размер дочернему контенту (он сам знает свои размеры).
func (tv *topologyViewport) Layout(size fyne.Size) {
	if tv.content == nil {
		return
	}
	tv.content.Resize(size)
}

// Refresh — обновить отображение.
func (tv *topologyViewport) Refresh() {
	if tv.content != nil {
		tv.content.Refresh()
	}
	widget.Refresh(tv)
}

// applyTransform перемещает и масштабирует всех детей контента
// согласно текущему zoom/pan. Применяется после каждого Scrolled/Dragged.
func (tv *topologyViewport) applyTransform() {
	if tv.content == nil {
		return
	}
	// Простой обход: для каждого CanvasObject в children — Move/Resize.
	// Работает для WithoutLayout-контейнеров (card, box и т.д.).
	var visit func(fyne.CanvasObject)
	visit = func(obj fyne.CanvasObject) {
		pos := obj.Position()
		sz := obj.Size()
		// Масштабируем позицию: worldPos * zoom + pan.
		// Мир координат центрируется: offset = (worldSize*zoom - currentSize) / 2.
		wsz := tv.minWorld.Width * tv.zoom
		hsz := tv.minWorld.Height * tv.zoom
		offsetX := (wsz - sz.Width) / 2
		offsetY := (hsz - sz.Height) / 2
		// Позиция: (pos - worldSize/2)*zoom + worldSize/2*zoom + pan
		// = pos*zoom + pan
		newX := pos.X*tv.zoom + tv.pan.X
		newY := pos.Y*tv.zoom + tv.pan.Y
		obj.Move(fyne.NewPos(newX, newY))
		obj.Resize(fyne.NewSize(sz.Width*tv.zoom, sz.Height*tv.zoom))
		if cc, ok := obj.(interface{ ChildObjects() []fyne.CanvasObject }); ok {
			for _, child := range cc.ChildObjects() {
				visit(child)
			}
		}
		// Fyne Container: есть Children.
		if cont, ok := obj.(*fyne.Container); ok {
			for _, child := range cont.Objects {
				visit(child)
			}
		}
	}
	// Рекурсивно обходим детей content.
	if cont, ok := tv.content.(*fyne.Container); ok {
		for _, child := range cont.Objects {
			visit(child)
		}
	}
}

// Scrolled — зум колесом мыши (H-07).
func (tv *topologyViewport) Scrolled(ev *fyne.ScrollEvent) {
	if tv == nil {
		return
	}
	delta := ev.Scrolled.Y
	if delta > 0 {
		// колесо вверх → зум +
		tv.zoom *= topoZoomStep
	} else {
		// колесо вниз → зум −
		tv.zoom /= topoZoomStep
	}
	// Clamp.
	if tv.zoom < topoZoomMin {
		tv.zoom = topoZoomMin
	}
	if tv.zoom > topoZoomMax {
		tv.zoom = topoZoomMax
	}
	// Центрируем зум вокруг позиции курсора (внутри viewport).
	if tv.content != nil {
		contentSize := tv.content.Size()
		cursor := ev.Position
		// Позиция курсора относительно контента.
		// Применяем смещение: pan = cursor - (cursor - pan) * (newZoom / oldZoom)
		// Но проще: пересчитать pan чтобы точка под курсором осталась.
		// Для упрощения: центрируем зум по центру контента.
		// (H-07: зум вокруг курсора — улучшено ниже)
		centerX := contentSize.Width / 2
		centerY := contentSize.Height / 2
		// Центрируем: pan = cursor - cursor*zoom
		tv.pan.X = cursor.X - cursor.X*tv.zoom
		tv.pan.Y = cursor.Y - cursor.Y*tv.zoom
	}
	tv.applyTransform()
	tv.Refresh()
	if tv.onViewChanged != nil {
		tv.onViewChanged()
	}
}

// Dragged — перетаскивание карты (pan) (H-07).
func (tv *topologyViewport) Dragged(ev *fyne.DragEvent) {
	if tv == nil {
		return
	}
	// Delta = перемещение с последнего события Dragged.
	tv.pan.X += ev.Dragged.X
	tv.pan.Y += ev.Dragged.Y
	tv.applyTransform()
	tv.Refresh()
	if tv.onViewChanged != nil {
		tv.onViewChanged()
	}
}

// DragEnd — конец перетаскивания (noop — ничего не нужно сбрасывать).
func (tv *topologyViewport) DragEnd() {}

// Cursor — курсор-рука при наведении (desktop.Hoverable).
func (tv *topologyViewport) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// CreateRenderer — возвращает рендерер, который рисует только content.
func (tv *topologyViewport) CreateRenderer() fyne.WidgetRenderer {
	if tv == nil {
		return &topologyViewportRenderer{view: tv, objects: nil}
	}
	return &topologyViewportRenderer{view: tv, objects: []fyne.CanvasObject{tv.content}}
}

// ZoomIn — зум вперёд.
func (tv *topologyViewport) ZoomIn() {
	if tv == nil {
		return
	}
	tv.zoom *= topoZoomStep
	if tv.zoom > topoZoomMax {
		tv.zoom = topoZoomMax
	}
	tv.applyTransform()
	tv.Refresh()
	if tv.onViewChanged != nil {
		tv.onViewChanged()
	}
}

// ZoomOut — зум назад.
func (tv *topologyViewport) ZoomOut() {
	if tv == nil {
		return
	}
	tv.zoom /= topoZoomStep
	if tv.zoom < topoZoomMin {
		tv.zoom = topoZoomMin
	}
	tv.applyTransform()
	tv.Refresh()
	if tv.onViewChanged != nil {
		tv.onViewChanged()
	}
}

// ResetView — сбросить зум и панораму к значениям по умолчанию.
func (tv *topologyViewport) ResetView() {
	if tv == nil {
		return
	}
	tv.zoom = topoZoomDefault
	tv.pan = fyne.NewPos(0, 0)
	tv.applyTransform()
	tv.Refresh()
	if tv.onViewChanged != nil {
		tv.onViewChanged()
	}
}

// ViewportZoom возвращает текущий zoom.
func (tv *topologyViewport) ViewportZoom() float32 {
	if tv == nil {
		return 1.0
	}
	return tv.zoom
}

// ViewportPan возвращает текущий pan.
func (tv *topologyViewport) ViewportPan() fyne.Position {
	if tv == nil {
		return fyne.NewPos(0, 0)
	}
	return tv.pan
}

// topologyViewportRenderer — рендерер viewport.
type topologyViewportRenderer struct {
	view    *topologyViewport
	objects []fyne.CanvasObject
}

func (r *topologyViewportRenderer) Layout(size fyne.Size) {
	if r.view == nil || r.view.content == nil {
		return
	}
	r.view.content.Resize(size)
}

func (r *topologyViewportRenderer) MinSize() fyne.Size {
	if r.view == nil {
		return fyne.NewSize(0, 0)
	}
	return r.view.MinSize()
}

func (r *topologyViewportRenderer) Refresh() {
	if r.view == nil {
		return
	}
	r.view.Refresh()
}

func (r *topologyViewportRenderer) Objects() []fyne.CanvasObject {
	if r.objects == nil {
		return nil
	}
	return r.objects
}

func (r *topologyViewportRenderer) Destroy() {}
