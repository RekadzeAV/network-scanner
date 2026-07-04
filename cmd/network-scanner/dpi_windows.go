//go:build windows

package main

import (
	"syscall"
)

// SetProcessDPIAwareness включает пер-monitor DPI awareness
func SetProcessDPIAwareness() {
	// Пробуем PerMonitorV2 (Windows 10 Anniversary Update+)
	user32, err := syscall.LoadDLL("user32.dll")
	if err == nil {
		proc, err := user32.FindProc("SetProcessDpiAwarenessContext")
		if err == nil {
			ret, _, _ := proc.Call(4) // PROCESS_PER_MONITOR_V2_AWARE
			if ret != 0 {
				return // Успешно
			}
		}
	}

	// Fallback: SetProcessDPIAware (Windows Vista+)
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return
	}

	proc, err := kernel32.FindProc("SetProcessDPIAware")
	if err != nil {
		return
	}

	proc.Call()
}
