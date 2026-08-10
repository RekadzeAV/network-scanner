//go:build linux

package security

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PermissionCheckResult содержит результат проверки прав
type PermissionCheckResult struct {
	HasRoot          bool
	HasNetRaw        bool
	HasSysAdmin      bool
	RecommendedMode  string // "root", "setcap", "user"
	Suggestions      []string
	Warnings         []string
}

// CheckPermissions проверяет права, необходимые для сетевого сканирования
func CheckPermissions() *PermissionCheckResult {
	result := &PermissionCheckResult{
		Suggestions: make([]string, 0, 3),
		Warnings:    make([]string, 0, 3),
	}

	// Проверяем root
	result.HasRoot = os.Geteuid() == 0

	// Проверяем capabilities
	result.HasNetRaw = checkCapability("CAP_NET_RAW")
	result.HasSysAdmin = checkCapability("CAP_SYS_ADMIN")

	// Определяем рекомендуемый режим
	if result.HasRoot {
		result.RecommendedMode = "root"
	} else if result.HasNetRaw {
		result.RecommendedMode = "setcap"
	} else {
		result.RecommendedMode = "user"
	}

	// Формируем рекомендации
	if !result.HasRoot && !result.HasNetRaw {
		result.Warnings = append(result.Warnings,
			"Для полного сканирования требуются расширенные права",
			"Без прав root/NET_RAW: возможен только ping-скан без ARP",
		)

		if runtime.GOOS == "linux" {
			result.Suggestions = append(result.Suggestions,
				"Запускать с sudo: sudo network-scanner scan",
				"Или установить capability: sudo setcap cap_net_raw+ep /usr/local/bin/network-scanner",
			)
		}
	}

	if !result.HasNetRaw && !result.HasRoot {
		result.Warnings = append(result.Warnings,
			"Не удалось определить MAC-адреса устройств",
			"Некоторые функции топологии могут быть недоступны",
		)
	}

	return result
}

// checkCapability проверяет наличие конкретной capability через getpcaps
func checkCapability(cap string) bool {
	cmd := exec.Command("getpcaps", "self")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// getpcaps может не работать, пробуем альтернативный метод
		return checkCapabilityProcFS(cap)
	}

	return strings.Contains(string(output), cap)
}

// checkCapabilityProcFS проверяет capabilities через /proc/self/status
func checkCapabilityProcFS(cap string) bool {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "CapEff:") {
			// CapEff: 0000000000000000
			// Проверяем, есть ли бит для CAP_NET_RAW (27)
			hexStr := strings.TrimSpace(strings.Split(line, ":")[1])
			if hexStr == "0000000000000000" {
				return false
			}
			// Парсим hex и проверяем бит
			var capValue uint64
			fmt.Sscanf(hexStr, "%x", &capValue)
			return (capValue & (1 << 27)) != 0 // CAP_NET_RAW = 27
		}
	}
	return false
}

// FormatPermissionReport возвращает человеко-читаемый отчёт о правах
func FormatPermissionReport(result *PermissionCheckResult) string {
	report := "=== Проверка прав ===\n"

	if result.HasRoot {
		report += "✅ Запущен от имени root\n"
	} else {
		report += "⚠️  Запущен от обычного пользователя\n"
	}

	if result.HasNetRaw {
		report += "✅ CAP_NET_RAW доступна\n"
	} else {
		report += "⚠️  CAP_NET_RAW недоступна\n"
	}

	if result.RecommendedMode == "user" {
		report += "\nРекомендации:\n"
		for _, s := range result.Suggestions {
			report += "  • " + s + "\n"
		}
	}

	for _, w := range result.Warnings {
		report += "  ⚠️  " + w + "\n"
	}

	return report
}
