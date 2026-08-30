//go:build !linux && !windows && !darwin

package security

import (
	"os"
)

// PermissionCheckResult содержит результат проверки прав
type PermissionCheckResult struct {
	HasRoot         bool
	HasNetRaw       bool
	HasSysAdmin     bool
	RecommendedMode string
	Suggestions     []string
	Warnings        []string
}

// CheckPermissions проверяет права (stub для non-Linux)
func CheckPermissions() *PermissionCheckResult {
	result := &PermissionCheckResult{
		Suggestions:     []string{"Запускать с sudo для полного сканирования"},
		RecommendedMode: "sudo",
	}

	if len(os.Args) > 0 {
		// Для non-Linux просто возвращаем базовую информацию
	}

	return result
}

// FormatPermissionReport возвращает человеко-читаемый отчёт о правах
func FormatPermissionReport(result *PermissionCheckResult) string {
	report := "=== Permission Check ===\n"
	if result.RecommendedMode == "sudo" {
		report += "Recommendation: Use sudo for full scanning capabilities\n"
	}
	return report
}
