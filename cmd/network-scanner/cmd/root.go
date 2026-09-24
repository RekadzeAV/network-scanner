package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Build-информация CLI. Значения по умолчанию — для локальной сборки без
// -ldflags; main-пакет перезаписывает их через SetBuildInfo (см. main.go /
// main_unix.go), где переменные Version/BuildTime/GitCommit заполняются
// линковщиком.
var (
	buildVersion   = "dev"
	buildTime      = "unknown"
	buildGitCommit = "unknown"
)

// SetBuildInfo передаёт информацию о сборке из main-пакета в CLI-слой.
// Вызывается из main() до ExecuteCLI, чтобы подкоманда `version` показывала
// реальные значения, а не захардкоженные.
func SetBuildInfo(version, buildTimeValue, gitCommit string) {
	if version != "" {
		buildVersion = version
	}
	if buildTimeValue != "" {
		buildTime = buildTimeValue
	}
	if gitCommit != "" {
		buildGitCommit = gitCommit
	}
}

var rootCmd = &cobra.Command{
	Use:   "network-scanner",
	Short: "Network Scanner - инструмент для сканирования локальной сети",
	Long: `Network Scanner — мощный инструмент для сканирования локальной сети.
Поддерживает обнаружение хостов, сканирование портов, анализ безопасности,
построение топологии, SNMP-опрос и инвентаризацию устройств.`,
}

// Execute запускает корневую команду
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Добавляем подкоманды
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(inventoryCmd)
	rootCmd.AddCommand(remoteExecCmd)
	rootCmd.AddCommand(deviceControlCmd)
	rootCmd.AddCommand(guiCmd)
	rootCmd.AddCommand(versionCmd)
}

// guiCmd — запуск графического интерфейса.
//
// Реальная реализация подключается build-tag `gui` (scan_gui.go); без тега
// доступна заглушка (scan_gui_stub.go), сообщающая о необходимости отдельной
// сборки.
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Запустить GUI приложение",
	Long:  "Запускает графический интерфейс Network Scanner (Fyne).",
	Run: func(cmd *cobra.Command, args []string) {
		RunGUI()
	},
}

// versionCmd — отдельная команда для версии
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Показать версию приложения",
	Long:  "Выводит информацию о версии, времени сборки и коммите.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("network-scanner version %s\n", buildVersion)
		fmt.Printf("Build time: %s\n", buildTime)
		fmt.Printf("Git commit: %s\n", buildGitCommit)
	},
}
