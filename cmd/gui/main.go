package main

import (
	"network-scanner/internal/gui"
	"network-scanner/internal/logger"
	"network-scanner/internal/platform"
)

const (
	AppName    = "network-scanner-gui"
	AppVersion = "1.0.3"
)

// main — точка входа GUI приложения.
//
// На Windows для сборки GUI без консоли используйте:
//
//	go build -ldflags "-H windowsgui" -o network-scanner-gui.exe ./cmd/gui
//
// Флаг -H windowsgui переключает подсистему PE-заголовка с IMAGE_SUBSYSTEM_WINDOWS_CUI
// (консоль) на IMAGE_SUBSYSTEM_WINDOWS_GUI (GUI), что устраняет мигание окна
// cmd.exe при запуске.
func main() {
	// Включаем Per-Monitor V2 DPI awareness ДО создания окна и приложения Fyne.
	// Иначе Windows провиртуализирует DPI процесса и сменить awareness нельзя.
	// Без этого Fyne получает системный DPI (96) вместо реального DPI монитора,
	// что ломает масштаб канвы, выход за границы ultrawide и координаты мыши.
	platform.SetProcessDPIAwareness()

	// Инициализируем логирование (работает только в debug версии)
	_ = logger.Init(AppName, AppVersion)
	defer logger.Close()

	logger.Log("Запуск GUI приложения")

	app := gui.NewApp()
	app.Run()

	logger.Log("GUI приложение завершено")
}
