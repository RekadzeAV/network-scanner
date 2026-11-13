// Package platform предоставляет кроссплатформенные системные утилиты.
package platform

// SetProcessDPIAwareness включает пер-monitor DPI awareness для Windows.
// На Linux/macOS — заглушка (система обрабатывает DPI автоматически).
//
// Вызывать ДО создания окна (в main, до app.New()). Иначе Windows уже
// провиртуализировала DPI процесса и сменить awareness нельзя.
func SetProcessDPIAwareness() {
	setProcessDPIAwareness()
}
