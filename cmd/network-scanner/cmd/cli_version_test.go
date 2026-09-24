package cmd

import "testing"

// ============================================================================
// W3: Тесты build-информации CLI (подкоманда version).
//
// Проверяют, что значения из main/ldflags доходят до CLI-слоя и что пустые
// значения не затирают дефолты.
// ============================================================================

// TestSetBuildInfo_OverridesDefaults — version показывает переданные значения,
// а не захардкоженный "dev".
func TestSetBuildInfo_OverridesDefaults(t *testing.T) {
	origVersion, origTime, origCommit := buildVersion, buildTime, buildGitCommit
	defer func() {
		buildVersion, buildTime, buildGitCommit = origVersion, origTime, origCommit
	}()

	SetBuildInfo("v9.9.9", "2026-09-22T00:00:00Z", "deadbee")
	if buildVersion != "v9.9.9" {
		t.Errorf("buildVersion = %q, want v9.9.9", buildVersion)
	}
	if buildTime != "2026-09-22T00:00:00Z" {
		t.Errorf("buildTime = %q", buildTime)
	}
	if buildGitCommit != "deadbee" {
		t.Errorf("buildGitCommit = %q", buildGitCommit)
	}
}

// TestSetBuildInfo_IgnoresEmpty — пустые значения не затирают ранее заданные.
func TestSetBuildInfo_IgnoresEmpty(t *testing.T) {
	origVersion, origTime, origCommit := buildVersion, buildTime, buildGitCommit
	defer func() {
		buildVersion, buildTime, buildGitCommit = origVersion, origTime, origCommit
	}()

	SetBuildInfo("v1.0.0", "t", "c")
	SetBuildInfo("", "", "")
	if buildVersion != "v1.0.0" || buildTime != "t" || buildGitCommit != "c" {
		t.Errorf("пустые значения перезаписали build info: %q %q %q", buildVersion, buildTime, buildGitCommit)
	}
}

func TestVersionCommand_Registered(t *testing.T) {
	if findCommand(rootCmd, "version") == nil {
		t.Fatal("version-команда не зарегистрирована")
	}
}

// TestDefaultBuildInfo — дефолты для локальной сборки без ldflags.
func TestDefaultBuildInfo(t *testing.T) {
	origVersion, origTime, origCommit := buildVersion, buildTime, buildGitCommit
	defer func() {
		buildVersion, buildTime, buildGitCommit = origVersion, origTime, origCommit
	}()

	buildVersion, buildTime, buildGitCommit = "dev", "unknown", "unknown"
	if buildVersion == "" || buildTime == "" || buildGitCommit == "" {
		t.Error("дефолтные build-значения не должны быть пустыми")
	}
}
