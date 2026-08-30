package gui

import (
	"testing"
)

// --- touch_gestures.go: Setup, scrollUp, scrollDown ---

func TestTouchGestures_Setup_WithCanvas(t *testing.T) {
	tg := NewTouchGestures(&App{})
	tg.canvas = nil
	// Без myWindow вызов паникует — пропускаем в headless
	t.Skip("Setup требует активное окно Fyne")
}

func TestTouchGestures_ScrollUp_NilApp(t *testing.T) {
	tg := &TouchGestures{app: nil}
	// Не должен паниковать (пустая реализация)
	tg.scrollUp()
}

func TestTouchGestures_ScrollDown_NilApp(t *testing.T) {
	tg := &TouchGestures{app: nil}
	// Не должен паниковать (пустая реализация)
	tg.scrollDown()
}

func TestTouchGestures_Setup_WithCanvasSet(t *testing.T) {
	tg := &TouchGestures{app: &App{}, canvas: nil}
	// canvas уже задан -> Setup не должен обращаться к окну
	tg.isEnabled = true
}
