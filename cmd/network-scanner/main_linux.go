//go:build linux

package main

import (
	"fmt"
	"os"

	"network-scanner/cmd/network-scanner/cmd"
	"network-scanner/internal/security"
)

// Build information - set via -ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Для Linux DPI-awareness не нужна
	SetProcessDPIAwareness()

	// Check permissions before running (except for help/version)
	if len(os.Args) > 1 && !startsWithDashDash(os.Args[1]) && os.Args[1] != "help" && os.Args[1] != "--help" {
		permResult := security.CheckPermissions()
		fmt.Println(security.FormatPermissionReport(permResult))
	}

	// Use cobra for CLI
	cmd.Execute()
}

func startsWithDashDash(s string) bool {
	return len(s) >= 2 && s[0] == '-' && s[1] == '-'
}
