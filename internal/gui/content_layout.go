package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// maxContentWidth — максимальная ширина основного контента в логических
// пикселях канвы. На ultrawide-мониторах (34" 3440×1440, 49" 5120×1440)
// интерфейс, растянутый на всю ширину, нечитаем: формы и строки результатов
// «разъезжаются» к краям экрана. Контент центрируется в этом коридоре; на
// экранах уже коридора layout ничего не меняет.
const maxContentWidth = 2400

// maxWidthLayout размещает единственный дочерний объект по центру, ограничивая
// его ширину максимумом max. Высота всегда используется полная — вертикальное
// пространство сокращать нельзя.
type maxWidthLayout struct {
	max float32
}

func (l *maxWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	child := objects[0]
	w := size.Width
	if l.max > 0 && w > l.max {
		w = l.max
	}
	x := (size.Width - w) / 2
	if x < 0 {
		x = 0
	}
	child.Move(fyne.NewPos(x, 0))
	child.Resize(fyne.NewSize(w, size.Height))
}

// MinSize делегирует дочернему объекту: ограничение maxWidth — верхнее,
// оно не должно раздувать минимальный размер окна.
func (l *maxWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	return objects[0].MinSize()
}

// wrapContentMaxWidth оборачивает текущий контент окна в коридор максимальной
// ширины. Идемпотентно: повторный вызов не создаёт вложенных обёрток.
func (a *App) wrapContentMaxWidth() {
	if a == nil || a.myWindow == nil {
		return
	}
	content := a.myWindow.Content()
	if content == nil {
		return
	}
	if c, ok := content.(*fyne.Container); ok {
		if _, isWrapped := c.Layout.(*maxWidthLayout); isWrapped {
			return
		}
	}
	a.myWindow.SetContent(container.New(&maxWidthLayout{max: maxContentWidth}, content))
}
