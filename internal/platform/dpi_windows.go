//go:build windows

package platform

import (
	"syscall"
)

// setProcessDPIAwareness включает Per-Monitor V2 DPI awareness (Windows).
//
// Без него GetDpiForWindow() возвращает системный DPI (96) вместо реального
// DPI монитора. Fyne вычисляет canvas.scale из этого значения, поэтому на
// HiDPI/ultrawide-мониторах масштаб канвы расходится с физическими пикселями:
// окно отрисовывается за границами экрана, а координаты мыши проецируются по
// неверному scale — клики «не попадают» по видимому положению виджетов.
func setProcessDPIAwareness() {
	// Пробуем PerMonitorV2 (Windows 10 Anniversary Update+).
	user32, err := syscall.LoadDLL("user32.dll")
	if err == nil {
		proc, err := user32.FindProc("SetProcessDpiAwarenessContext")
		if err == nil {
			// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4
			ret, _, _ := proc.Call(^uintptr(3))
			if ret != 0 {
				return
			}
		}
	}

	// Fallback: SetProcessDPIAware (Windows Vista+).
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("SetProcessDPIAware")
	_, _, _ = proc.Call()
}
