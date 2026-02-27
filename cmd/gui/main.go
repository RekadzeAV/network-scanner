package main

import (
	"network-scanner/internal/gui"
	"network-scanner/internal/logger"
)

const (
	AppName    = "network-scanner-gui"
	AppVersion = "1.0.3"
)

func main() {
	// Инициализируем логирование (работает только в debug версии)
	logger.Init(AppName, AppVersion)
	defer logger.Close()

	logger.Log("Запуск GUI приложения")
	
	app := gui.NewApp()
	app.Run()
	
	logger.Log("GUI приложение завершено")
}
