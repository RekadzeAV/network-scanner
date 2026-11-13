package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestMaxWidthLayout_NarrowCanvasUsesFullWidth(t *testing.T) {
	child := widget.NewLabel("content")
	c := container.New(&maxWidthLayout{max: maxContentWidth}, child)

	c.Resize(fyne.NewSize(1280, 800))

	if child.Position().X != 0 {
		t.Errorf("узкий экран: x должен быть 0, получено %v", child.Position().X)
	}
	if child.Size().Width != 1280 {
		t.Errorf("узкий экран: ширина должна быть 1280, получено %v", child.Size().Width)
	}
	if child.Size().Height != 800 {
		t.Errorf("высота должна использоваться полностью (800), получено %v", child.Size().Height)
	}
}

func TestMaxWidthLayout_UltrawideCentersContent(t *testing.T) {
	child := widget.NewLabel("content")
	c := container.New(&maxWidthLayout{max: maxContentWidth}, child)

	// 34" ultrawide: 3440 логических пикселей.
	c.Resize(fyne.NewSize(3440, 1440))

	if child.Size().Width != maxContentWidth {
		t.Errorf("ultrawide: ширина должна быть %d, получено %v", maxContentWidth, child.Size().Width)
	}
	wantX := (float32(3440) - float32(maxContentWidth)) / 2
	if child.Position().X != wantX {
		t.Errorf("ultrawide: контент должен быть центрирован (x=%v), получено %v", wantX, child.Position().X)
	}
	// Высота на ultrawide не режется.
	if child.Size().Height != 1440 {
		t.Errorf("ultrawide: высота должна быть 1440, получено %v", child.Size().Height)
	}
}

func TestMaxWidthLayout_MinSizeDelegatesToChild(t *testing.T) {
	child := widget.NewLabel("content")
	l := &maxWidthLayout{max: maxContentWidth}

	got := l.MinSize([]fyne.CanvasObject{child})
	want := child.MinSize()

	if got.Width != want.Width || got.Height != want.Height {
		t.Errorf("MinSize должен делегировать дочернему объекту: получено %v, ожидалось %v", got, want)
	}
}

func TestMaxWidthLayout_EmptyObjectsNoPanic(t *testing.T) {
	l := &maxWidthLayout{max: maxContentWidth}
	l.Layout(nil, fyne.NewSize(1000, 600))
	if sz := l.MinSize(nil); sz.Width != 0 || sz.Height != 0 {
		t.Errorf("пустой список должен давать нулевой MinSize, получено %v", sz)
	}
}

func TestWrapContentMaxWidth_SetsWrapper(t *testing.T) {
	a := NewTestApp(t)
	content := container.NewHBox(widget.NewLabel("x"))
	a.myWindow.SetContent(content)

	a.wrapContentMaxWidth()

	wrapped := a.myWindow.Content()
	if wrapped == fyne.CanvasObject(content) {
		t.Fatalf("контент не обёрнут в maxWidthLayout")
	}
	c, ok := wrapped.(*fyne.Container)
	if !ok {
		t.Fatalf("обёртка должна быть *fyne.Container, получено %T", wrapped)
	}
	if _, ok := c.Layout.(*maxWidthLayout); !ok {
		t.Fatalf("layout обёртки должен быть *maxWidthLayout, получено %T", c.Layout)
	}
}

func TestWrapContentMaxWidth_Idempotent(t *testing.T) {
	a := NewTestApp(t)
	a.myWindow.SetContent(container.NewHBox(widget.NewLabel("x")))

	a.wrapContentMaxWidth()
	first := a.myWindow.Content()
	a.wrapContentMaxWidth()
	second := a.myWindow.Content()

	if first != second {
		t.Fatalf("повторный вызов создал вложенную обёртку: %v != %v", first, second)
	}
}

func TestWrapContentMaxWidth_NilWindowNoPanic(t *testing.T) {
	(&App{}).wrapContentMaxWidth() // не должен паниковать
}

// NewTestApp создаёт App с тестовым Fyne-приложением и окном.
func NewTestApp(t *testing.T) *App {
	t.Helper()
	fyneApp := test.NewApp()
	t.Cleanup(fyneApp.Quit)
	return &App{
		myApp:    fyneApp,
		myWindow: fyneApp.NewWindow("test"),
	}
}
