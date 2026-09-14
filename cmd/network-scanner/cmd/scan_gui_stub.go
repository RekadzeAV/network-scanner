//go:build !gui

package cmd

// RunGUI — stub для GUI при сборке без тега gui
func RunGUI() {
	// При сборке без тега gui GUI недоступен
	// Используйте отдельную сборку GUI: go build -tags gui ./cmd/network-scanner
	// или используйте ./cmd/gui напрямую
}
