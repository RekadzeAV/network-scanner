package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
	rootCmd.AddCommand(versionCmd)
}

// versionCmd — отдельная команда для версии
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Показать версию приложения",
	Long:  "Выводит информацию о версии, времени сборки и коммите.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("network-scanner version dev")
		fmt.Println("Build time: unknown")
		fmt.Println("Git commit: unknown")
	},
}
