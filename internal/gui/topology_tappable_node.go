package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// tappableNode — интерактивный узел карты топологии (H-01).
//
// canvas.Circle — примитив отрисовки без обработки мыши. tappableNode оборачивает
// круг в BaseWidget, реализует desktop.Tappable (клик) и desktop.Hoverable
// (курсор-указатель + подсветка при наведении). Hit-зона — квадрат side × side
// вокруг центра, что заметно больше видимого кружка (WCAG 2.5.8 minimum target).
type tappableNode struct {
	widget.BaseWidget
	circle *canvas.Circle
	side   float32
	tapped func()
	hover  bool
}

// newTappableNode создаёт кликабельный узел радиуса radius (диаметр = 2*radius)
// с увеличенной hit-зоной hitRadius (лог. px от центра).
func newTappableNode(fill color.Color, strokeColor color.Color, strokeWidth float32, radius, hitRadius float32, onTap func()) *tappableNode {
	side := hitRadius * 2
	if side < radius*2 {
		side = radius * 2
	}
	tn := &tappableNode{
		side:   side,
		tapped: onTap,
	}
	tn.circle = canvas.NewCircle(fill)
	tn.circle.StrokeColor = strokeColor
	tn.circle.StrokeWidth = strokeWidth
	tn.circle.Resize(fyne.NewSize(radius*2, radius*2))
	tn.ExtendBaseWidget(tn)
	return tn
}

// SetSelected меняет обводку: выбранная — контрастнее.
func (tn *tappableNode) SetSelected(selected bool) {
	if tn == nil || tn.circle == nil {
		return
	}
	tn.circle.StrokeColor = colorByNodeBorder(selected)
	tn.Refresh()
}

// SetFillColor меняет заливку (например, при смене темы).
func (tn *tappableNode) SetFillColor(c color.Color) {
	if tn == nil || tn.circle == nil {
		return
	}
	tn.circle.FillColor = c
	tn.Refresh()
}

// MinSize — размер hit-зоны.
func (tn *tappableNode) MinSize() fyne.Size {
	return fyne.NewSize(tn.side, tn.side)
}

// Layout — центрирует круг внутри hit-зоны.
func (tn *tappableNode) Layout(size fyne.Size) {
	if tn.circle == nil {
		return
	}
	diam := size.Width
	if size.Height < diam {
		diam = size.Height
	}
	if diam < 4 {
		diam = 4
	}
	tn.circle.Resize(fyne.NewSize(diam, diam))
	tn.circle.Move(fyne.NewPos((size.Width-diam)/2, (size.Height-diam)/2))
}

// Tapped — вызов колбэка (desktop.Tappable).
func (tn *tappableNode) Tapped(_ *fyne.PointEvent) {
	if tn != nil && tn.tapped != nil {
		tn.tapped()
	}
}

// TappedSecondary — нет действия.
func (tn *tappableNode) TappedSecondary(_ *fyne.PointEvent) {}

// Cursor — указатель-рука при наведении (desktop.Hoverable).
func (tn *tappableNode) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// MouseIn — подсветка при наведении.
func (tn *tappableNode) MouseIn(_ *desktop.MouseEvent) {
	if tn == nil {
		return
	}
	tn.hover = true
	tn.Refresh()
}

// MouseMoved — no-op.
func (tn *tappableNode) MouseMoved(_ *desktop.MouseEvent) {}

// MouseOut — сброс подсветки.
func (tn *tappableNode) MouseOut() {
	if tn == nil {
		return
	}
	tn.hover = false
	tn.Refresh()
}

// Compile-time interface assertions.
var _ fyne.Tappable = (*tappableNode)(nil)
var _ desktop.Hoverable = (*tappableNode)(nil)

// linkHitArea — невидимый кликабельный прямоугольник вдоль связи (H-02).
// Увеличивает hit-area линии StrokeWidth=2 до hitWidth лог. px.
type linkHitArea struct {
	widget.BaseWidget
	hitWidth float32
	tapped   func()
}

// newLinkHitArea создаёт невидимый кликабельный прямоугольник вдоль линии p1→p2.
func newLinkHitArea(p1, p2 fyne.Position, hitWidth float32, onTap func()) *linkHitArea {
	la := &linkHitArea{hitWidth: hitWidth, tapped: onTap}
	la.ExtendBaseWidget(la)
	return la
}

// minFloat32 / maxFloat32 — вспомогательные функции для bounding box.
func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// linkHitAreaSize вычисляет bounding box для hit-area линии p1→p2 с шириной hitWidth.
func linkHitAreaSize(p1, p2 fyne.Position, hitWidth float32) (pos fyne.Position, size fyne.Size) {
	minX := minFloat32(p1.X, p2.X)
	minY := minFloat32(p1.Y, p2.Y)
	maxX := maxFloat32(p1.X, p2.X)
	maxY := maxFloat32(p1.Y, p2.Y)
	pad := hitWidth / 2
	return fyne.NewPos(minX-pad, minY-pad), fyne.NewSize(maxX-minX+hitWidth, maxY-minY+hitWidth)
}

func (la *linkHitArea) MinSize() fyne.Size {
	return fyne.NewSize(20, la.hitWidth) // минимум 20px
}

func (la *linkHitArea) Layout(size fyne.Size) {
	// Позиция и размер уже заданы через Move/Resize снаружи.
}

func (la *linkHitArea) Tapped(_ *fyne.PointEvent) {
	if la != nil && la.tapped != nil {
		la.tapped()
	}
}

func (la *linkHitArea) TappedSecondary(_ *fyne.PointEvent) {}

func (la *linkHitArea) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (la *linkHitArea) CreateRenderer() fyne.WidgetRenderer {
	return &linkHitAreaRenderer{area: la}
}

type linkHitAreaRenderer struct {
	area *linkHitArea
}

func (r *linkHitAreaRenderer) Layout(size fyne.Size) {}
func (r *linkHitAreaRenderer) MinSize() fyne.Size    { return r.area.MinSize() }
func (r *linkHitAreaRenderer) Refresh()              {}
func (r *linkHitAreaRenderer) Objects() []fyne.CanvasObject {
	return nil // полностью невидим
}
func (r *linkHitAreaRenderer) Destroy() {}

// tappableNodeRenderer — рендерер, отрисовывающий круг и hover-подсветку.
type tappableNodeRenderer struct {
	node *tappableNode
}

func (r *tappableNodeRenderer) Layout(size fyne.Size) {
	r.node.Layout(size)
}

func (r *tappableNodeRenderer) MinSize() fyne.Size {
	return r.node.MinSize()
}

func (r *tappableNodeRenderer) Refresh() {
	if r.node == nil || r.node.circle == nil {
		return
	}
	if r.node.hover {
		r.node.circle.StrokeWidth = 3
	} else {
		r.node.circle.StrokeWidth = 2
	}
	canvas.Refresh(r.node.circle)
	canvas.Refresh(r.node)
}

func (r *tappableNodeRenderer) Objects() []fyne.CanvasObject {
	if r.node == nil || r.node.circle == nil {
		return nil
	}
	return []fyne.CanvasObject{r.node.circle}
}

func (r *tappableNodeRenderer) Destroy() {}

func (tn *tappableNode) CreateRenderer() fyne.WidgetRenderer {
	return &tappableNodeRenderer{node: tn}
}
