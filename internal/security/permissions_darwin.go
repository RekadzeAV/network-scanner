//go:build darwin

package security

import (
	"os"
	"runtime"
)

// PermissionCheckResult содержит результат проверки прав
type PermissionCheckResult struct {
	HasRoot         bool
	HasNetRaw       bool
	HasSysAdmin     bool
	RecommendedMode string // "root", "user"
	Suggestions     []string
	Warnings        []string
}

// CheckPermissions проверяет права для macOS
func CheckPermissions() *PermissionCheckResult {
	result := &PermissionCheckResult{
		Suggestions: make([]string, 0, 3),
		Warnings:    make([]string, 0, 3),
	}

	// Проверяем root
	result.HasRoot = os.Geteuid() == 0

	// Для macOS нет аналога CAP_NET_RAW, но можно использовать RawSockets
	// Если запущено от root — можно использовать RawSockets
	result.HasNetRaw = result.HasRoot

	// Определяем рекомендуемый режим
	if result.HasRoot {
		result.RecommendedMode = "root"
	} else {
		result.RecommendedMode = "user"
	}

	// Формируем рекомендации
	if !result.HasRoot {
		result.Warnings = append(result.Warnings,
			"Для полного сканирования требуются права root",
			"Без прав root: возможен только ping-скан",
		)

		if runtime.GOOS == "darwin" {
			result.Suggestions = append(result.Suggestions,
				"Запускать с sudo: sudo network-scanner scan",
			)
		}
	}

	return result
}

// FormatPermissionReport возвращает человеко-читаемый отчёт о правах
func FormatPermissionReport(result *PermissionCheckResult) string {
	report := "=== Проверка прав (macOS) ===\n"

	if result.HasRoot {
		report += "✅ Запущен от имени root\n"
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
