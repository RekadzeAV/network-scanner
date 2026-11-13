//go:build !windows

package platform

// Rect — физический прямоугольник в экранных координатах.
type Rect struct {
	Left, Top, Right, Bottom int32
}

func (r Rect) Width() int32  { return r.Right - r.Left }
func (r Rect) Height() int32 { return r.Bottom - r.Top }

// BorderlessState — снимок стиля и геометрии окна для восстановления.
type BorderlessState struct {
	Style   uintptr
	Rect    Rect
	Monitor Rect
}

// MonitorFullRectForWindow — заглушка (только Windows).
func MonitorFullRectForWindow(_ uintptr) (Rect, bool) { return Rect{}, false }

// MonitorWorkRectForWindow — заглушка (только Windows).
func MonitorWorkRectForWindow(_ uintptr) (Rect, bool) { return Rect{}, false }

// WindowRect — заглушка (только Windows).
func WindowRect(_ uintptr) (Rect, bool) { return Rect{}, false }

// SetWindowPos — заглушка (только Windows).
func SetWindowPos(_ uintptr, _, _, _, _ int32, _ bool) error { return nil }

// MakeBorderlessFullscreen — заглушка (только Windows).
func MakeBorderlessFullscreen(_ uintptr) (*BorderlessState, bool) { return nil, false }

// RestoreBorderless — заглушка (только Windows).
func RestoreBorderless(_ uintptr, _ *BorderlessState) bool { return false }

// IsBorderlessFullscreen — заглушка (только Windows).
func IsBorderlessFullscreen(_ uintptr) bool { return false }
