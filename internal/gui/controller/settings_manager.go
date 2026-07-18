package controller

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
)

// SplitState хранит состояние разделителей панелей.
type SplitState struct {
	ScanTabOffset      float64
	TopologyOffset     float64
	ToolsOffset        float64
	HostDetailsVOffset float64
	HostDetailsHOffset float64
}

// SettingsManager управляет настройками приложения.
type SettingsManager struct {
	app fyne.App
}

// NewSettingsManager создает менеджер настроек.
func NewSettingsManager(app fyne.App) *SettingsManager {
	return &SettingsManager{app: app}
}

// LoadSplitStates загружает сохраненные состояния разделителей.
func (m *SettingsManager) LoadSplitStates() SplitState {
	if m.app == nil {
		return SplitState{}
	}
	p := m.app.Preferences()
	return SplitState{
		ScanTabOffset:      p.FloatWithFallback("scan.ui.scan_tab_split_offset", -1),
		TopologyOffset:     p.FloatWithFallback("scan.ui.topology_main_split_offset", -1),
		ToolsOffset:        p.FloatWithFallback("scan.ui.tools_tab_split_offset", -1),
		HostDetailsVOffset: p.FloatWithFallback("scan.ui.host_details_split_offset_v", -1),
		HostDetailsHOffset: p.FloatWithFallback("scan.ui.host_details_split_offset_h", -1),
	}
}

// SaveSplitState сохраняет состояние разделителя.
func (m *SettingsManager) SaveSplitState(key string, offset float64) {
	if m.app == nil {
		return
	}
	p := m.app.Preferences()
	p.SetFloat(key, offset)
}

// ClampOffset ограничивает значение offset.
func (m *SettingsManager) ClampOffset(offset, lo, hi float64) float64 {
	if offset < lo {
		return lo
	}
	if offset > hi {
		return hi
	}
	return offset
}

// ApplySplitOffset применяет offset к контейнеру Split.
func (m *SettingsManager) ApplySplitOffset(split *container.Split, offset float64) {
	if split == nil {
		return
	}
	split.Offset = offset
}

// ResetUIPanelLayout сбрасывает положение разделителей.
func (m *SettingsManager) ResetUIPanelLayout(scanSplit, topoSplit, toolsSplit *container.Split) {
	defaults := SplitState{
		ScanTabOffset:      0.50,
		TopologyOffset:     0.55,
		ToolsOffset:        0.40,
		HostDetailsVOffset: 0.60,
		HostDetailsHOffset: 0.65,
	}
	if scanSplit != nil {
		scanSplit.Offset = defaults.ScanTabOffset
	}
	if topoSplit != nil {
		topoSplit.Offset = defaults.TopologyOffset
	}
	if toolsSplit != nil {
		toolsSplit.Offset = defaults.ToolsOffset
	}
	m.SaveSplitState("scan.ui.scan_tab_split_offset", defaults.ScanTabOffset)
	m.SaveSplitState("scan.ui.topology_main_split_offset", defaults.TopologyOffset)
	m.SaveSplitState("scan.ui.tools_tab_split_offset", defaults.ToolsOffset)
}

// ResetUIPanelLayoutWithFeedback сбрасывает положение разделителей с уведомлением.
func (m *SettingsManager) ResetUIPanelLayoutWithFeedback(scanSplit, topoSplit, toolsSplit *container.Split, window fyne.Window) {
	m.ResetUIPanelLayout(scanSplit, topoSplit, toolsSplit)
	if window != nil {
		dialog.ShowInformation("Вид", "Положение разделителей между панелями (вкладки Сканирование, Топология, Инструменты) и split «результаты / Host Details» восстановлено по умолчанию.", window)
	}
}

// ApplyDefaultSplitOffsetsForProfile применяет дефолтные offset для профиля.
func (m *SettingsManager) ApplyDefaultSplitOffsetsForProfile(profile string) {
	defaults := SplitState{
		ScanTabOffset:      0.50,
		TopologyOffset:     0.55,
		ToolsOffset:        0.40,
		HostDetailsVOffset: 0.60,
		HostDetailsHOffset: 0.65,
	}
	if strings.TrimSpace(profile) == "compact" {
		defaults.ScanTabOffset = 0.45
		defaults.TopologyOffset = 0.50
		defaults.ToolsOffset = 0.35
	}
	m.SaveSplitState("scan.ui.scan_tab_split_offset", defaults.ScanTabOffset)
	m.SaveSplitState("scan.ui.topology_main_split_offset", defaults.TopologyOffset)
	m.SaveSplitState("scan.ui.tools_tab_split_offset", defaults.ToolsOffset)
	m.SaveSplitState("scan.ui.host_details_split_offset_v", defaults.HostDetailsVOffset)
	m.SaveSplitState("scan.ui.host_details_split_offset_h", defaults.HostDetailsHOffset)
}

// ClearSplitPreferences очищает все сохраненные offset.
func (m *SettingsManager) ClearSplitPreferences() {
	if m.app == nil {
		return
	}
	p := m.app.Preferences()
	p.RemoveValue("scan.ui.scan_tab_split_offset")
	p.RemoveValue("scan.ui.topology_main_split_offset")
	p.RemoveValue("scan.ui.tools_tab_split_offset")
	p.RemoveValue("scan.ui.host_details_split_offset_v")
	p.RemoveValue("scan.ui.host_details_split_offset_h")
}
