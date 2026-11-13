//go:build windows

package platform

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Rect — физический прямоугольник в экранных координатах.
type Rect struct {
	Left, Top, Right, Bottom int32
}

func (r Rect) Width() int32  { return r.Right - r.Left }
func (r Rect) Height() int32 { return r.Bottom - r.Top }

// Win32 constants.
const (
	gwlStyle = -16

	wsCaption    = 0x00C00000 // WS_CAPTION
	wsThickFrame = 0x00040000 // WS_THICKFRAME (resizing border)

	swpNoZORDER     = 0x0004
	swpFrameChanged = 0x0020
	swpNoActivate   = 0x0010

	monitorDefaultToNearest = 2
)

type rectW struct {
	Left, Top, Right, Bottom int32
}

type monitorInfo struct {
	Size     uint32
	Monitor  rectW
	WorkArea rectW
	Flags    uint32
}

var (
	user32                = syscall.NewLazyDLL("user32.dll")
	procMonitorFromWindow = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfo    = user32.NewProc("GetMonitorInfoW")
	procSetWindowPos      = user32.NewProc("SetWindowPos")
	procGetWindowRect     = user32.NewProc("GetWindowRect")
)

// getLongPtrWrapped wraps GetWindowLongPtrW with GetWindowLongW fallback (32-bit).
func getLongPtrWrapped(hwnd uintptr, index int32) uintptr {
	proc := user32.NewProc("GetWindowLongPtrW")
	if err := proc.Find(); err != nil {
		ret, _, _ := user32.NewProc("GetWindowLongW").Call(hwnd, uintptr(index))
		return ret
	}
	ret, _, _ := proc.Call(hwnd, uintptr(index))
	return ret
}

func setLongPtrWrapped(hwnd uintptr, index int32, value uintptr) uintptr {
	proc := user32.NewProc("SetWindowLongPtrW")
	if err := proc.Find(); err != nil {
		ret, _, _ := user32.NewProc("SetWindowLongW").Call(hwnd, uintptr(index), value)
		return ret
	}
	ret, _, _ := proc.Call(hwnd, uintptr(index), value)
	return ret
}

// monitorRect возвращает rcMonitor (полный) и rcWork (рабочая зона) монитора окна.
func monitorRect(hwnd uintptr) (full, work Rect, ok bool) {
	ret, _, _ := procMonitorFromWindow.Call(hwnd, uintptr(monitorDefaultToNearest))
	if ret == 0 {
		return Rect{}, Rect{}, false
	}
	var mi monitorInfo
	mi.Size = uint32(unsafe.Sizeof(mi))
	r, _, _ := procGetMonitorInfo.Call(ret, uintptr(unsafe.Pointer(&mi)))
	if r == 0 {
		return Rect{}, Rect{}, false
	}
	return Rect{mi.Monitor.Left, mi.Monitor.Top, mi.Monitor.Right, mi.Monitor.Bottom},
		Rect{mi.WorkArea.Left, mi.WorkArea.Top, mi.WorkArea.Right, mi.WorkArea.Bottom},
		true
}

// MonitorFullRectForWindow — полный прямоугольник монитора (включая область под
// панелью задач). Используется для borderless fullscreen.
func MonitorFullRectForWindow(hwnd uintptr) (Rect, bool) {
	full, _, ok := monitorRect(hwnd)
	return full, ok
}

// MonitorWorkRectForWindow — рабочая зона монитора (без панели задач).
func MonitorWorkRectForWindow(hwnd uintptr) (Rect, bool) {
	_, work, ok := monitorRect(hwnd)
	return work, ok
}

// WindowRect возвращает текущий прямоугольник окна в экранных координатах.
func WindowRect(hwnd uintptr) (Rect, bool) {
	var r rectW
	ret, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	if ret == 0 {
		return Rect{}, false
	}
	return Rect{r.Left, r.Top, r.Right, r.Bottom}, true
}

// SetWindowPos устанавливает позицию и размер окна в физических пикселях.
func SetWindowPos(hwnd uintptr, x, y, w, h int32, noActivate bool) error {
	flags := uintptr(swpNoZORDER | swpFrameChanged)
	if noActivate {
		flags |= swpNoActivate
	}
	ret, _, err := procSetWindowPos.Call(
		hwnd, 0,
		uintptr(x), uintptr(y),
		uintptr(w), uintptr(h),
		flags,
	)
	if ret == 0 {
		return fmt.Errorf("SetWindowPos failed: %w", err)
	}
	return nil
}

// BorderlessState — снимок стиля и геометрии окна для восстановления.
type BorderlessState struct {
	Style   uintptr
	Rect    Rect
	Monitor Rect
}

// MakeBorderlessFullscreen снимает рамки окна и растягивает его на весь монитор.
//
// В отличие от glfwSetWindowMonitor (который вызывает Fyne SetFullScreen), этот
// путь НЕ пересоздаёт OpenGL surface и НЕ меняет monitor у GLFW-окна. Поэтому
// canvas.scale остаётся корректным (координаты мыши не смещаются), а menu bar
// не теряется при смене viewport.
func MakeBorderlessFullscreen(hwnd uintptr) (*BorderlessState, bool) {
	monitor, ok := MonitorFullRectForWindow(hwnd)
	if !ok {
		return nil, false
	}
	cur, ok := WindowRect(hwnd)
	if !ok {
		cur = monitor
	}
	style := getLongPtrWrapped(hwnd, gwlStyle)

	// Снимаем заголовок и рамку изменения размера.
	newStyle := style &^ (wsCaption | wsThickFrame)
	setLongPtrWrapped(hwnd, gwlStyle, newStyle)

	if err := SetWindowPos(hwnd, monitor.Left, monitor.Top, monitor.Width(), monitor.Height(), false); err != nil {
		setLongPtrWrapped(hwnd, gwlStyle, style)
		return nil, false
	}

	return &BorderlessState{Style: style, Rect: cur, Monitor: monitor}, true
}

// RestoreBorderless возвращает окну рамки и исходную геометрию.
func RestoreBorderless(hwnd uintptr, st *BorderlessState) bool {
	if st == nil {
		return false
	}
	setLongPtrWrapped(hwnd, gwlStyle, st.Style)
	return SetWindowPos(hwnd, st.Rect.Left, st.Rect.Top, st.Rect.Width(), st.Rect.Height(), false) == nil
}

// IsBorderlessFullscreen проверяет, сняты ли у окна рамки (WS_CAPTION отсутствует).
func IsBorderlessFullscreen(hwnd uintptr) bool {
	style := getLongPtrWrapped(hwnd, gwlStyle)
	return style&wsCaption == 0
}
