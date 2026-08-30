package gui

import (
	"testing"

	"fyne.io/fyne/v2"
)

// --- touch_gestures.go tests ---

func TestNewTouchGestures_Created(t *testing.T) {
	app := &App{}
	tg := NewTouchGestures(app)
	if tg == nil {
		t.Fatal("expected non-nil TouchGestures")
	}
	if tg.app != app {
		t.Error("expected app to be set")
	}
	if !tg.isEnabled {
		t.Error("expected gestures to be enabled by default")
	}
}

func TestNewTouchGestures_NilApp(t *testing.T) {
	tg := NewTouchGestures(nil)
	if tg == nil {
		t.Fatal("expected non-nil TouchGestures for nil app")
	}
	// isEnabled всегда true по умолчанию
	if !tg.isEnabled {
		t.Error("expected gestures to be enabled by default")
	}
}

func TestTouchGestures_Enable(t *testing.T) {
	tg := &TouchGestures{isEnabled: false}
	tg.Enable()
	if !tg.isEnabled {
		t.Error("expected isEnabled=true after Enable()")
	}
}

func TestTouchGestures_Disable(t *testing.T) {
	tg := &TouchGestures{isEnabled: true}
	tg.Disable()
	if tg.isEnabled {
		t.Error("expected isEnabled=false after Disable()")
	}
}

func TestTouchGestures_HandleSwipe_Disabled(t *testing.T) {
	tg := &TouchGestures{isEnabled: false}
	// Не паникует при отключенных жестах
	tg.HandleSwipe(&fyne.PointEvent{}, -100)
	if tg.isEnabled {
		t.Error("expected isEnabled to remain false")
	}
}

func TestTouchGestures_HandleSwipe_Up(t *testing.T) {
	tg := &TouchGestures{isEnabled: true}
	// Свайп вверх: dy < -50
	tg.HandleSwipe(&fyne.PointEvent{}, -100)
	if !tg.isEnabled {
		t.Error("expected isEnabled to remain true")
	}
	// scrollUp — заглушка, но не паникует
}

func TestTouchGestures_HandleSwipe_Down(t *testing.T) {
	tg := &TouchGestures{isEnabled: true}
	// Свайп вниз: dy > 50
	tg.HandleSwipe(&fyne.PointEvent{}, 100)
	if !tg.isEnabled {
		t.Error("expected isEnabled to remain true")
	}
}

func TestTouchGestures_HandleSwipe_Small(t *testing.T) {
	tg := &TouchGestures{isEnabled: true}
	// Свайп < 50 — не должен вызывать прокрутку
	tg.HandleSwipe(&fyne.PointEvent{}, 30)
	if !tg.isEnabled {
		t.Error("expected isEnabled to remain true")
	}
}

func TestTouchGestures_HandlePinch_Disabled(t *testing.T) {
	tg := &TouchGestures{isEnabled: false}
	tg.HandlePinch(&fyne.PointEvent{}, 1.5)
	// Не паникует при отключенных жестах
}

func TestTouchGestures_HandlePinch_Enabled(t *testing.T) {
	tg := &TouchGestures{isEnabled: true}
	// scale=1.0 — не должен изменять масштаб
	tg.HandlePinch(&fyne.PointEvent{}, 1.0)
	if !tg.isEnabled {
		t.Error("expected isEnabled to remain true")
	}
}

func TestTouchGestures_HandleLongPress_Disabled(t *testing.T) {
	tg := &TouchGestures{isEnabled: false}
	tg.HandleLongPress(fyne.NewPos(100, 100))
	// Не паникует при отключенных жестах
}

func TestTouchGestures_HandleLongPress_Enabled(t *testing.T) {
	app := &App{}
	tg := &TouchGestures{isEnabled: true, app: app}
	// Не паникует при включенных жестах
	tg.HandleLongPress(fyne.NewPos(100, 100))
	if !tg.isEnabled {
		t.Error("expected isEnabled to remain true")
	}
}

func TestDesktopCustomShortcut_Create(t *testing.T) {
	dcs := NewDesktopCustomShortcut(fyne.KeyF5, fyne.KeyModifierShift, func() {})
	if dcs == nil {
		t.Fatal("expected non-nil DesktopCustomShortcut")
	}
	if dcs.key != fyne.KeyF5 {
		t.Errorf("expected key=F5, got %v", dcs.key)
	}
	if dcs.modifier != fyne.KeyModifierShift {
		t.Errorf("expected modifier=Shift, got %v", dcs.modifier)
	}
}

func TestDesktopCustomShortcut_Trigger(t *testing.T) {
	triggered := false
	dcs := NewDesktopCustomShortcut(fyne.KeyF5, fyne.KeyModifierShift, func() {
		triggered = true
	})
	dcs.Trigger()
	if !triggered {
		t.Error("expected onTriggered to be called")
	}
}

func TestDesktopCustomShortcut_TriggerNil(t *testing.T) {
	dcs := &DesktopCustomShortcut{}
	// onTriggered=nil — не должен паниковать
	dcs.Trigger()
}
