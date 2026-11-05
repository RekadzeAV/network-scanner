package controller

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

// TestSettingsManager_LoadSplitStates_nilApp проверяет, что при nil app
// возвращается пустой SplitState.
func TestSettingsManager_LoadSplitStates_nilApp(t *testing.T) {
	m := NewSettingsManager(nil)
	state := m.LoadSplitStates()
	if state.ScanTabOffset != 0 {
		t.Fatalf("ScanTabOffset = %v, want 0", state.ScanTabOffset)
	}
	if state.TopologyOffset != 0 {
		t.Fatalf("TopologyOffset = %v, want 0", state.TopologyOffset)
	}
	if state.ToolsOffset != 0 {
		t.Fatalf("ToolsOffset = %v, want 0", state.ToolsOffset)
	}
}

// TestSettingsManager_LoadSplitStates_withValues проверяет загрузку значений
// из Preferences.
func TestSettingsManager_LoadSplitStates_withValues(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()
	p.SetFloat("scan.ui.scan_tab_split_offset", 0.42)
	p.SetFloat("scan.ui.topology_main_split_offset", 0.55)
	p.SetFloat("scan.ui.tools_tab_split_offset", 0.33)
	p.SetFloat("scan.ui.host_details_split_offset_v", 0.60)
	p.SetFloat("scan.ui.host_details_split_offset_h", 0.65)

	m := NewSettingsManager(a)
	state := m.LoadSplitStates()

	if state.ScanTabOffset != 0.42 {
		t.Fatalf("ScanTabOffset = %v, want 0.42", state.ScanTabOffset)
	}
	if state.TopologyOffset != 0.55 {
		t.Fatalf("TopologyOffset = %v, want 0.55", state.TopologyOffset)
	}
	if state.ToolsOffset != 0.33 {
		t.Fatalf("ToolsOffset = %v, want 0.33", state.ToolsOffset)
	}
	if state.HostDetailsVOffset != 0.60 {
		t.Fatalf("HostDetailsVOffset = %v, want 0.60", state.HostDetailsVOffset)
	}
	if state.HostDetailsHOffset != 0.65 {
		t.Fatalf("HostDetailsHOffset = %v, want 0.65", state.HostDetailsHOffset)
	}
}

// TestSettingsManager_ClearSplitPreferences_nilApp проверяет, что при nil app
// функция безопасно завершается.
func TestSettingsManager_ClearSplitPreferences_nilApp(t *testing.T) {
	m := NewSettingsManager(nil)
	m.ClearSplitPreferences() // не должен паниковать
}

// TestSettingsManager_ClearSplitPreferences_withValues проверяет удаление всех
// сохраненных offset из Preferences.
func TestSettingsManager_ClearSplitPreferences_withValues(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := a.Preferences()
	p.SetFloat("scan.ui.scan_tab_split_offset", 0.50)
	p.SetFloat("scan.ui.topology_main_split_offset", 0.55)
	p.SetFloat("scan.ui.tools_tab_split_offset", 0.40)
	p.SetFloat("scan.ui.host_details_split_offset_v", 0.60)
	p.SetFloat("scan.ui.host_details_split_offset_h", 0.65)

	m := NewSettingsManager(a)
	m.ClearSplitPreferences()

	// После удаления значения должны вернуться к fallback (-1)
	state := m.LoadSplitStates()
	if state.ScanTabOffset != -1 {
		t.Fatalf("ScanTabOffset = %v, want -1 (removed)", state.ScanTabOffset)
	}
	if state.TopologyOffset != -1 {
		t.Fatalf("TopologyOffset = %v, want -1 (removed)", state.TopologyOffset)
	}
	if state.ToolsOffset != -1 {
		t.Fatalf("ToolsOffset = %v, want -1 (removed)", state.ToolsOffset)
	}
	if state.HostDetailsVOffset != -1 {
		t.Fatalf("HostDetailsVOffset = %v, want -1 (removed)", state.HostDetailsVOffset)
	}
	if state.HostDetailsHOffset != -1 {
		t.Fatalf("HostDetailsHOffset = %v, want -1 (removed)", state.HostDetailsHOffset)
	}
}

// TestSettingsManager_ResetUIPanelLayout_nilApp проверяет, что при nil app
// ResetUIPanelLayout безопасно завершается.
func TestSettingsManager_ResetUIPanelLayout_nilApp(t *testing.T) {
	m := NewSettingsManager(nil)
	m.ResetUIPanelLayout(nil, nil, nil) // не должен паниковать
}

// TestSettingsManager_ResetUIPanelLayout_nilSplits проверяет, что при nil
// split-контейнерах функция не паникует.
func TestSettingsManager_ResetUIPanelLayout_nilSplits(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	m := NewSettingsManager(a)
	m.ResetUIPanelLayout(nil, nil, nil) // не должен паниковать
}

// TestSettingsManager_ResetUIPanelLayoutWithFeedback_nilApp проверяет, что при
// nil app функция безопасно завершается.
func TestSettingsManager_ResetUIPanelLayoutWithFeedback_nilApp(t *testing.T) {
	m := NewSettingsManager(nil)
	m.ResetUIPanelLayoutWithFeedback(nil, nil, nil, nil) // не должен паниковать
}

// TestSettingsManager_SaveSplitState_nilApp проверяет, что при nil app
// SaveSplitState безопасно завершается.
func TestSettingsManager_SaveSplitState_nilApp(t *testing.T) {
	m := NewSettingsManager(nil)
	m.SaveSplitState("test.key", 0.5) // не должен паниковать
}

// TestSettingsManager_SaveSplitState_withApp проверяет сохранение offset.
func TestSettingsManager_SaveSplitState_withApp(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	m := NewSettingsManager(a)
	m.SaveSplitState("scan.ui.scan_tab_split_offset", 0.75)

	p := a.Preferences()
	if p.FloatWithFallback("scan.ui.scan_tab_split_offset", -1) != 0.75 {
		t.Fatalf("offset = %v, want 0.75", p.FloatWithFallback("scan.ui.scan_tab_split_offset", -1))
	}
}

// TestSettingsManager_ClampOffset проверяет граничные значения ClampOffset.
func TestSettingsManager_ClampOffset(t *testing.T) {
	m := &SettingsManager{}

	// Ниже нижнего предела
	if m.ClampOffset(0.1, 0.2, 0.8) != 0.2 {
		t.Fatal("should clamp to lower bound")
	}

	// Выше верхнего предела
	if m.ClampOffset(0.9, 0.2, 0.8) != 0.8 {
		t.Fatal("should clamp to upper bound")
	}

	// В пределах диапазона
	if m.ClampOffset(0.5, 0.2, 0.8) != 0.5 {
		t.Fatal("should return value unchanged")
	}

	// На границах
	if m.ClampOffset(0.2, 0.2, 0.8) != 0.2 {
		t.Fatal("lower boundary should pass through")
	}
	if m.ClampOffset(0.8, 0.2, 0.8) != 0.8 {
		t.Fatal("upper boundary should pass through")
	}
}

// TestSettingsManager_ApplySplitOffset_nilSplit проверяет, что ApplySplitOffset
// безопасно завершается при nil split.
func TestSettingsManager_ApplySplitOffset_nilSplit(t *testing.T) {
	m := &SettingsManager{}
	m.ApplySplitOffset(nil, 0.5) // не должен паниковать
}

// TestSettingsManager_ResetUIPanelLayout_allNil проверяет, что ResetUIPanelLayout
// с nil app и nil splits безопасно завершается.
func TestSettingsManager_ResetUIPanelLayout_allNil(t *testing.T) {
	m := NewSettingsManager(nil)
	m.ResetUIPanelLayout(nil, nil, nil) // не должен паниковать
}

// TestSettingsManager_ResetUIPanelLayoutWithFeedback_nilWindow проверяет, что
// при nil window диалог не показывается, но layout сбрасывается.
func TestSettingsManager_ResetUIPanelLayoutWithFeedback_nilWindow(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	m := NewSettingsManager(a)
	m.ResetUIPanelLayoutWithFeedback(nil, nil, nil, nil) // не должен паниковать

	// Проверим, что layout был сброшен
	p := a.Preferences()
	if p.FloatWithFallback("scan.ui.scan_tab_split_offset", -1) != 0.50 {
		t.Fatalf("ScanTabOffset = %v, want 0.50", p.FloatWithFallback("scan.ui.scan_tab_split_offset", -1))
	}
}

// TestSettingsManager_ApplyDefaultSplitOffsetsForProfile_compact проверяет
// применение compact профиля.
func TestSettingsManager_ApplyDefaultSplitOffsetsForProfile_compact(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	m := NewSettingsManager(a)
	m.ApplyDefaultSplitOffsetsForProfile("compact")

	p := a.Preferences()
	if p.FloatWithFallback("scan.ui.scan_tab_split_offset", -1) != 0.45 {
		t.Fatalf("compact scan offset = %v, want 0.45", p.FloatWithFallback("scan.ui.scan_tab_split_offset", -1))
	}
	if p.FloatWithFallback("scan.ui.topology_main_split_offset", -1) != 0.50 {
		t.Fatalf("compact topo offset = %v, want 0.50", p.FloatWithFallback("scan.ui.topology_main_split_offset", -1))
	}
	if p.FloatWithFallback("scan.ui.tools_tab_split_offset", -1) != 0.35 {
		t.Fatalf("compact tools offset = %v, want 0.35", p.FloatWithFallback("scan.ui.tools_tab_split_offset", -1))
	}
}

// TestSettingsManager_ApplyDefaultSplitOffsetsForProfile_default проверяет
// применение default профиля.
func TestSettingsManager_ApplyDefaultSplitOffsetsForProfile_default(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	m := NewSettingsManager(a)
	m.ApplyDefaultSplitOffsetsForProfile("default")

	p := a.Preferences()
	if p.FloatWithFallback("scan.ui.scan_tab_split_offset", -1) != 0.50 {
		t.Fatalf("default scan offset = %v, want 0.50", p.FloatWithFallback("scan.ui.scan_tab_split_offset", -1))
	}
}
