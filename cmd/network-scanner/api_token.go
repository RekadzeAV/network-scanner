//go:build !gui_only

package main

import (
	"os"
	"strings"
)

// envAPIToken — переменная окружения с Bearer-токеном для REST API (P0-1).
const envAPIToken = "NETWORK_SCANNER_API_TOKEN"

// resolveAPIToken определяет токен аутентификации REST API.
//
// Приоритет:
//  1. Флаг командной строки `--api-token=<token>` (или `--api-token <token>`).
//  2. Переменная окружения NETWORK_SCANNER_API_TOKEN.
//
// Пустой результат означает, что аутентификация отключена (dev/test).
// Токен не логируется.
func resolveAPIToken(args []string) string {
	for i, arg := range args {
		// Форма --api-token=value
		if strings.HasPrefix(arg, "--api-token=") {
			return strings.TrimPrefix(arg, "--api-token=")
		}
		// Форма --api-token value
		if arg == "--api-token" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return os.Getenv(envAPIToken)
}
