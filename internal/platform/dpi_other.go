//go:build !windows

package platform

// setProcessDPIAwareness — заглушка для Linux/macOS: эти системы сообщают
// масштаб монитора через X11/ Cocoa API, отдельная инициализация не требуется.
func setProcessDPIAwareness() {}
