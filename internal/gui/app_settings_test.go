package gui

import (
	"testing"
)

// --- app.go tests: nil-safe и чистые функции ---

func TestApp_saveScanSettings_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.saveScanSettings()
}

func TestApp_setPortRangeControlsEnabled_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.setPortRangeControlsEnabled(true)
	a.setPortRangeControlsEnabled(false)
}

func TestApp_clampScanTabMainSplitOffset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.clampScanTabMainSplitOffset()
}

func TestApp_clampScanTabMainSplitOffset_NilSplit(t *testing.T) {
	a := &App{}
	// Не должен паниковать при nil split
	a.clampScanTabMainSplitOffset()
}

func TestApp_clampTopologyMainSplitOffset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.clampTopologyMainSplitOffset()
}

func TestApp_clampTopologyMainSplitOffset_NilSplit(t *testing.T) {
	a := &App{}
	// Не должен паниковать при nil split
	a.clampTopologyMainSplitOffset()
}

func TestApp_clampToolsTabMainSplitOffset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.clampToolsTabMainSplitOffset()
}

func TestApp_clampToolsTabMainSplitOffset_NilSplit(t *testing.T) {
	a := &App{}
	// Не должен паниковать при nil split
	a.clampToolsTabMainSplitOffset()
}

func TestApp_loadScanTabSplitFromPrefs_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.loadScanTabSplitFromPrefs()
}

func TestApp_loadTopologySplitFromPrefs_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.loadTopologySplitFromPrefs()
}

func TestApp_loadToolsTabSplitFromPrefs_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.loadToolsTabSplitFromPrefs()
}

func TestApp_loadHostDetailsSplitFromPrefs_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.loadHostDetailsSplitFromPrefs()
}

func TestApp_maybePersistScanTabSplitOffset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.maybePersistScanTabSplitOffset()
}

func TestApp_maybePersistTopologySplitOffset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.maybePersistTopologySplitOffset()
}

func TestApp_maybePersistToolsTabSplitOffset_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.maybePersistToolsTabSplitOffset()
}

func TestApp_maybePersistHostDetailsSplitOffsets_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.maybePersistHostDetailsSplitOffsets()
}

func TestApp_resultsForSave_EmptyResults(t *testing.T) {
	a := &App{}
	results, reason := a.resultsForSave()
	if results != nil {
		t.Error("expected nil results for empty scanResults")
	}
	if reason == "" {
		t.Error("expected non-empty reason for empty results")
	}
	if reason != "Нет результатов для сохранения" {
		t.Errorf("expected 'Нет результатов для сохранения', got %q", reason)
	}
}

func TestApp_recommendedBadgeClassForHosts_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	_ = a.recommendedBadgeClassForHosts(100)
}

func TestApp_recommendedBadgeClassForHosts_Small(t *testing.T) {
	a := &App{}
	if a.recommendedBadgeClassForHosts(0) != "small" {
		t.Error("expected 'small' for 0 hosts")
	}
	if a.recommendedBadgeClassForHosts(10) != "small" {
		t.Error("expected 'small' for 10 hosts")
	}
	if a.recommendedBadgeClassForHosts(autoProfileHostLarge-1) != "small" {
		t.Error("expected 'small' for hosts below large threshold")
	}
}

func TestApp_recommendedBadgeClassForHosts_Medium(t *testing.T) {
	a := &App{}
	if a.recommendedBadgeClassForHosts(autoProfileHostLarge) != "medium" {
		t.Error("expected 'medium' for hosts at large threshold")
	}
	if a.recommendedBadgeClassForHosts(autoProfileHostXLarge-1) != "medium" {
		t.Error("expected 'medium' for hosts below xlarge threshold")
	}
}

func TestApp_recommendedBadgeClassForHosts_Large(t *testing.T) {
	a := &App{}
	if a.recommendedBadgeClassForHosts(autoProfileHostXLarge) != "large" {
		t.Error("expected 'large' for hosts at xlarge threshold")
	}
	if a.recommendedBadgeClassForHosts(autoProfileHostXXLarge-1) != "large" {
		t.Error("expected 'large' for hosts below xxlarge threshold")
	}
}

func TestApp_recommendedBadgeClassForHosts_VeryLarge(t *testing.T) {
	a := &App{}
	if a.recommendedBadgeClassForHosts(autoProfileHostXXLarge) != "very-large" {
		t.Error("expected 'very-large' for hosts at xxlarge threshold")
	}
	if a.recommendedBadgeClassForHosts(autoProfileHostXXLarge+1) != "very-large" {
		t.Error("expected 'very-large' for hosts above xxlarge threshold")
	}
}

func TestApp_recommendedBadgeClassForHosts_Extreme(t *testing.T) {
	a := &App{}
	if a.recommendedBadgeClassForHosts(100000) != "very-large" {
		t.Error("expected 'very-large' for 100000 hosts")
	}
}

func TestApp_recommendedBadgeText(t *testing.T) {
	a := &App{}
	text := a.recommendedBadgeText("balanced", "medium")
	expected := "Профиль: balanced (medium)"
	if text != expected {
		t.Errorf("expected %q, got %q", expected, text)
	}
}

func TestApp_recommendedBadgeText_EmptyName(t *testing.T) {
	a := &App{}
	text := a.recommendedBadgeText("", "small")
	if text == "" {
		t.Error("expected non-empty text")
	}
}

func TestApp_recommendedProfileNameForClass_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.recommendedProfileNameForClass("small")
}

func TestApp_recommendedProfileNameForClass_Small(t *testing.T) {
	a := &App{}
	name, ok := a.recommendedProfileNameForClass("small")
	if !ok {
		t.Error("expected ok=true for 'small'")
	}
	if name != "углубленный для небольшой подсети" {
		t.Errorf("expected 'углубленный для небольшой подсети', got %q", name)
	}
}

func TestApp_recommendedProfileNameForClass_Medium(t *testing.T) {
	a := &App{}
	name, ok := a.recommendedProfileNameForClass("medium")
	if !ok {
		t.Error("expected ok=true for 'medium'")
	}
	if name != "сбалансированный для средней подсети" {
		t.Errorf("expected 'сбалансированный для средней подсети', got %q", name)
	}
}

func TestApp_recommendedProfileNameForClass_Large(t *testing.T) {
	a := &App{}
	name, ok := a.recommendedProfileNameForClass("large")
	if !ok {
		t.Error("expected ok=true for 'large'")
	}
	if name != "бережный для крупной подсети" {
		t.Errorf("expected 'бережный для крупной подсети', got %q", name)
	}
}

func TestApp_recommendedProfileNameForClass_VeryLarge(t *testing.T) {
	a := &App{}
	name, ok := a.recommendedProfileNameForClass("very-large")
	if !ok {
		t.Error("expected ok=true for 'very-large'")
	}
	if name != "бережный для очень крупной подсети" {
		t.Errorf("expected 'бережный для очень крупной подсети', got %q", name)
	}
}

func TestApp_recommendedProfileNameForClass_Unknown(t *testing.T) {
	a := &App{}
	_, ok := a.recommendedProfileNameForClass("unknown")
	if ok {
		t.Error("expected ok=false for 'unknown'")
	}
}

func TestApp_recommendedProfileNameForClass_Empty(t *testing.T) {
	a := &App{}
	_, ok := a.recommendedProfileNameForClass("")
	if ok {
		t.Error("expected ok=false for empty string")
	}
}
