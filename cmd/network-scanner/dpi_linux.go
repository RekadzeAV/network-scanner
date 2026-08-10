//go:build linux

package main

// SetProcessDPIAwareness — заглушка для Linux (DPI-awareness не требуется)
func SetProcessDPIAwareness() {
	// Ничего не делаем на Linux
}
