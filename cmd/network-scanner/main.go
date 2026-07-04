//go:build windows

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"network-scanner/cmd/network-scanner/cmd"
	"network-scanner/internal/api"
)

// Build information - set via -ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Включаем Per-Monitor DPI awareness для корректного отображения на Windows
	SetProcessDPIAwareness()

	// Check for --version flag first
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("network-scanner version %s\n", Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Git commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// Check for --api flag
	if len(os.Args) > 1 && os.Args[1] == "--api" {
		cfg := api.DefaultConfig()
		router := api.NewRouter(cfg)
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		fmt.Printf("Starting REST API server on %s\n", addr)
		if err := http.ListenAndServe(addr, router.GetRouter()); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Direct execution for backward compatibility
	// If first argument doesn't start with '--', it's a command
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "--") {
		cmd.ExecuteCLI()
		return
	}

	// Legacy flag interface (deprecated, but still works)
	fmt.Println("Использование: network-scanner <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  scan          Запустить сканирование")
	fmt.Println("  gui           Запустить GUI приложение")
	fmt.Println("  inventory     Управление инвентаризацией (list|diff)")
	fmt.Println("  --api         Запустить REST API сервер")
	fmt.Println()
	fmt.Println("Для подробной справки: network-scanner scan --help")
	os.Exit(0)
}
