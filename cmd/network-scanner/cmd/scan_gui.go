//go:build gui

package cmd

import (
	"network-scanner/internal/gui"
)

// RunGUI запускает GUI приложение
func RunGUI() {
	a := gui.NewApp()
	a.Run()
}
