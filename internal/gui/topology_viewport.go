package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// H-07: зум и панорама интерактивной карты топологии.
//
// Проектировка: трансформация применяется не к уже созданным объектам (это
// накапливает ошибку округления при повторных Move/Resize), а к раскладке —
// renderTopologyInteractiveMap пересчитывает позиции с учётом
// topologyMapState.zoom/panX/panY. Viewport поэтому только собирает ввод
// (колесо = зум, перетаскивание = панорама) и сообщает о нём колбэками.

const (
	topoZoomMin     = float32(0.35)
	topoZoomMax     = float32(3.0)
	topoZoomStep    = float32(1.2)
	topoZoomDefault = float32(1.0)

	// topoPanClampRatio — предел панорамы относительно размера мира: за его
	// пределы карту уводить нельзя, иначе граф «теряется» за краем экрана.
	topoPanClampRatio = float32(0.6)
)

// clampTopologyZoom ограничивает коэффициент масштаба допустимым диапазоном.
// Нераспознанные значения (<=0, NaN) дают масштаб по умолчанию.
func clampTopologyZoom(z float32) float32 {
	if !(z > 0) { // отсекание 0, отрицательных и NaN
		return topoZoomDefault
	}
	if z < topoZoomMin {
		return topoZoomMin
	}
	if z > topoZoomMax {
		return topoZoomMax
	}
	return z
}

// topologyZoomBy применяет шаг зума к текущему значению с ограничением.
// direction > 0 — приблизить, < 0 — отдалить, 0 — без изменения.
func topologyZoomBy(current float32, direction float32) float32 {
	z := clampTopologyZoom(current)
	switch {
	case direction > 0:
		z *= topoZoomStep
	case direction < 0:
		z /= topoZoomStep
	default:
		return z
	}
	return clampTopologyZoom(z)
}

// clampTopologyPan ограничивает смещение карты долями от её размера.
// Возвращает скорректированные panX/panY в тех же единицах, что и world.
func clampTopologyPan(pan, world float32) float32 {
	limit := world * topoPanClampRatio
	if pan > limit {
		return limit
	}
	if pan < -limit {
		return -limit
	}
	return pan
}

// topologyViewport — слой ввода поверх холста карты: колесо мыши = зум,
// перетаскивание = панорама. Собственной графики не имеет, а холст лежит
// поверх него, поэтому клики по узлам и связям обрабатываются как раньше.
type topologyViewport struct {
	widget.BaseWidget

	content fyne.CanvasObject

	// worldSize — размер «мира» (раскладки в единицах топологии).
	// Используется для корректной работы скролл-баров и определения
	// того, есть ли куда скроллить (если worldSize ≤ viewport → wheel = зум).
	worldSize fyne.Size

	// onZoom вызывается на каждый тик колеса с направлением (+/-).
	onZoom func(direction float32)
	// onPan вызывается во время перетаскивания со смещением в лог. px.
	onPan func(dx, dy float32)
	// onDragEnd вызывается по завершении перетаскивания.
	onDragEnd func()
}

// newTopologyViewport создаёт слой ввода над content.
func newTopologyViewport(content fyne.CanvasObject) *topologyViewport {
	tv := &topologyViewport{content: content}
	tv.ExtendBaseWidget(tv)
	return tv
}

// SetHandlers задаёт обработчики зума и панорамы.
func (tv *topologyViewport) SetHandlers(onZoom func(float32), onPan func(float32, float32), onDragEnd func()) {
	if tv == nil {
		return
	}
	tv.onZoom = onZoom
	tv.onPan = onPan
	tv.onDragEnd = onDragEnd
}

func (tv *topologyViewport) MinSize() fyne.Size {
	if tv == nil {
		return fyne.NewSize(200, 200)
	}
	if tv.worldSize.Width > 0 || tv.worldSize.Height > 0 {
		return tv.worldSize
	}
	if tv.content == nil {
		return fyne.NewSize(200, 200)
	}
	return tv.content.MinSize()
}

// SetWorldSize устанавливает размер «мира» (раскладки) для корректной
// работы скролл-баров и определения наличия пространства для панорамы.
func (tv *topologyViewport) SetWorldSize(w fyne.Size) {
	if tv == nil {
		return
	}
	tv.worldSize = w
}

func (tv *topologyViewport) Layout(size fyne.Size) {
	if tv != nil && tv.content != nil {
		tv.content.Resize(size)
	}
}

// Scrolled — колесо мыши: зум карты (H-07).
func (tv *topologyViewport) Scrolled(ev *fyne.ScrollEvent) {
	if tv == nil || tv.onZoom == nil || ev == nil {
		return
	}
	tv.onZoom(ev.Scrolled.DY)
}

// Dragged — перетаскивание холста: панорама (H-07).
func (tv *topologyViewport) Dragged(ev *fyne.DragEvent) {
	if tv == nil || tv.onPan == nil || ev == nil {
		return
	}
	tv.onPan(ev.Dragged.DX, ev.Dragged.DY)
}

// DragEnd — завершение панорамы.
func (tv *topologyViewport) DragEnd() {
	if tv == nil || tv.onDragEnd == nil {
		return
	}
	tv.onDragEnd()
}

// Cursor — «указатель»: холст можно перетаскивать (H-05).
func (tv *topologyViewport) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (tv *topologyViewport) CreateRenderer() fyne.WidgetRenderer {
	return &topologyViewportRenderer{view: tv}
}

// topologyViewportRenderer — рендерер слоя ввода.
type topologyViewportRenderer struct {
	view *topologyViewport
}

func (r *topologyViewportRenderer) Layout(size fyne.Size) {
	if r.view != nil && r.view.content != nil {
		r.view.content.Move(fyne.NewPos(0, 0))
		r.view.content.Resize(size)
	}
}

func (r *topologyViewportRenderer) MinSize() fyne.Size {
	if r.view == nil {
		return fyne.NewSize(200, 200)
	}
	return r.view.MinSize()
}

func (r *topologyViewportRenderer) Objects() []fyne.CanvasObject {
	if r.view == nil || r.view.content == nil {
		return nil
	}
	return []fyne.CanvasObject{r.view.content}
}

func (r *topologyViewportRenderer) Refresh() {
	if r.view != nil && r.view.content != nil {
		r.view.content.Refresh()
	}
}

func (r *topologyViewportRenderer) Destroy() {}

var _ fyne.Scrollable = (*topologyViewport)(nil)
var _ fyne.Draggable = (*topologyViewport)(nil)
