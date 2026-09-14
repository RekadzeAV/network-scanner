//go:build windows

package security

import (
	"os/exec"
	"runtime"
	"strings"
)

// PermissionCheckResult содержит результат проверки прав
type PermissionCheckResult struct {
	HasAdmin        bool
	HasNetRaw       bool
	HasSysAdmin     bool
	RecommendedMode string // "admin", "user"
	Suggestions     []string
	Warnings        []string
}

// CheckPermissions проверяет права для Windows
func CheckPermissions() *PermissionCheckResult {
	result := &PermissionCheckResult{
		Suggestions: make([]string, 0, 3),
		Warnings:    make([]string, 0, 3),
	}

	// Проверяем, запущено ли от имени администратора
	result.HasAdmin = checkWindowsAdmin()

	// Для Windows нет аналога CAP_NET_RAW, но можно использовать RawSockets
	// Если запущено от администратора — можно использовать RawSockets
	result.HasNetRaw = result.HasAdmin

	// Определяем рекомендуемый режим
	if result.HasAdmin {
		result.RecommendedMode = "admin"
	} else {
		result.RecommendedMode = "user"
	}

	// Формируем рекомендации
	if !result.HasAdmin {
		result.Warnings = append(result.Warnings,
			"Для полного сканирования требуются права администратора",
			"Без прав администратора: возможен только ping-скан",
		)

		if runtime.GOOS == "windows" {
			result.Suggestions = append(result.Suggestions,
				"Запустить от имени администратора: правый клик → 'Запуск от имени администратора'",
				"Или использовать: RunAs /user:Administrator network-scanner",
			)
		}
	}

	return result
}

// checkWindowsAdmin проверяет, запущено ли приложение от имени администратора
func checkWindowsAdmin() bool {
	cmd := "whoami /groups"
	output, err := exec.Command("cmd", "/c", cmd).CombinedOutput()
	if err != nil {
		return false
	}

	// Проверяем, есть ли S-1-1-0 (Everyone) или S-1-5-32-544 (Administrators)
	return strings.Contains(string(output), "S-1-5-32-544") || strings.Contains(string(output), "S-1-1-0")
}

// FormatPermissionReport возвращает человеко-читаемый отчёт о правах
func FormatPermissionReport(result *PermissionCheckResult) string {
	report := "=== Проверка прав (Windows) ===\n"

	if result.HasAdmin {
		report += "✅ Запущен от имени администратора\n"
	} else {
		report += "⚠️  Запущен от обычного пользователя\n"
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
