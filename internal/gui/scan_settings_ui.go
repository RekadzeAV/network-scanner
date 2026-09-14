package gui

import (
	"strings"

	"fyne.io/fyne/v2/widget"
)

// --- Методы сохранения настроек сканирования ---

// saveScanSettings сохраняет все настройки сканирования и инструментов в Preferences.
func (a *App) saveScanSettings() {
	if a == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	p.SetString(prefNetwork, strings.TrimSpace(a.networkEntry.Text))
	p.SetString(prefPortRange, strings.TrimSpace(a.portRangeEntry.Text))
	p.SetString(prefTimeout, strings.TrimSpace(a.timeoutEntry.Text))
	p.SetString(prefThreads, strings.TrimSpace(a.threadsEntry.Text))
	if a.scanUDPCheck.Checked {
		p.SetString(prefScanUDP, "true")
	} else {
		p.SetString(prefScanUDP, "false")
	}
	if a.scanBannersCheck != nil && a.scanBannersCheck.Checked {
		p.SetString(prefScanBanners, "true")
	} else {
		p.SetString(prefScanBanners, "false")
	}
	if a.scanOSActiveCheck != nil && a.scanOSActiveCheck.Checked {
		p.SetString(prefScanOSActive, "true")
	} else {
		p.SetString(prefScanOSActive, "false")
	}
	if a.scanVerboseLogsCheck != nil && a.scanVerboseLogsCheck.Checked {
		p.SetString(prefScanVerbosePortLogs, "true")
	} else {
		p.SetString(prefScanVerbosePortLogs, "false")
	}
	if a.scanTCPPortsCheck != nil {
		if a.scanTCPPortsCheck.Checked {
			p.SetString(prefScanTCPPorts, "true")
		} else {
			p.SetString(prefScanTCPPorts, "false")
		}
	}
	if a.autoProfileCheck != nil {
		if a.autoProfileCheck.Checked {
			p.SetString(prefAutoProfile, "true")
		} else {
			p.SetString(prefAutoProfile, "false")
		}
	}
	if a.toolsHostEntry != nil {
		p.SetString(prefToolHost, strings.TrimSpace(a.toolsHostEntry.Text))
	}
	if a.toolsPingCountEnt != nil {
		p.SetString(prefToolPingCount, strings.TrimSpace(a.toolsPingCountEnt.Text))
	}
	if a.toolsTimeoutEnt != nil {
		p.SetString(prefToolTimeout, strings.TrimSpace(a.toolsTimeoutEnt.Text))
	}
	if a.toolsTraceHopsEnt != nil {
		p.SetString(prefToolTraceHops, strings.TrimSpace(a.toolsTraceHopsEnt.Text))
	}
	if a.toolsDNSResolverEnt != nil {
		p.SetString(prefToolResolver, strings.TrimSpace(a.toolsDNSResolverEnt.Text))
	}
	if a.toolsAuditMinSeveritySel != nil {
		p.SetString(prefToolAuditMinSeverity, strings.TrimSpace(a.toolsAuditMinSeveritySel.Selected))
	}
	if a.toolsDeviceTargetEntry != nil {
		p.SetString(prefToolDeviceTarget, strings.TrimSpace(a.toolsDeviceTargetEntry.Text))
	}
	if a.toolsDeviceVendorEntry != nil {
		p.SetString(prefToolDeviceVendor, strings.TrimSpace(a.toolsDeviceVendorEntry.Selected))
	}
	if a.toolsDeviceUserEntry != nil {
		p.SetString(prefToolDeviceUser, strings.TrimSpace(a.toolsDeviceUserEntry.Text))
	}
	if a.recommendedProfileBadge != nil {
		p.SetString(prefRecommendedBadge, strings.TrimSpace(a.recommendedProfileBadge.Text))
	}
	if a.inventoryDBEntry != nil {
		p.SetString(prefInventoryDBPath, strings.TrimSpace(a.inventoryDBEntry.Text))
	}
	if a.inventoryAutoSaveCheck != nil {
		if a.inventoryAutoSaveCheck.Checked {
			p.SetString(prefInventoryAutoSave, "true")
		} else {
			p.SetString(prefInventoryAutoSave, "false")
		}
	}
}

// setPortRangeControlsEnabled включает/отключает элементы управления портами.
func (a *App) setPortRangeControlsEnabled(enabled bool) {
	if a == nil {
		return
	}
	if a.portRangeEntry != nil {
		if enabled {
			a.portRangeEntry.Enable()
		} else {
			a.portRangeEntry.Disable()
		}
	}
	for _, b := range []*widget.Button{
		a.presetQuickBtn, a.presetBalBtn, a.presetDeepBtn,
		a.portWellKnownBtn, a.portRegisteredBtn, a.portDynamicBtn,
	} {
		if b != nil {
			if enabled {
				b.Enable()
			} else {
				b.Disable()
			}
		}
	}
}

// --- Методы Split-настроек вкладок ---

// loadScanTabSplitFromPrefs загружает позицию сплита вкладки сканирования.
func (a *App) loadScanTabSplitFromPrefs() {
	if a == nil || a.scanTabMainSplit == nil || a.myApp == nil {
		return
	}
	v := a.myApp.Preferences().FloatWithFallback(prefScanTabSplitOffset, -1)
	if v >= 0.16 && v <= 0.82 {
		a.scanTabMainSplit.Offset = v
		a.scanTabSplitInitialized = true
		a.scanTabSplitPersistPrimed = true
		a.lastPersistedScanSplit = v
	}
}

// clampScanTabMainSplitOffset ограничивает позицию сплита вкладки сканирования.
func (a *App) clampScanTabMainSplitOffset() {
	if a == nil || a.scanTabMainSplit == nil {
		return
	}
	const lo, hi = 0.15, 0.78
	o := a.scanTabMainSplit.Offset
	if o < lo {
		a.scanTabMainSplit.Offset = lo
	} else if o > hi {
		a.scanTabMainSplit.Offset = hi
	}
}

// maybePersistScanTabSplitOffset сохраняет позицию сплита вкладки сканирования.
func (a *App) maybePersistScanTabSplitOffset() {
	if a == nil || a.scanTabMainSplit == nil || a.myApp == nil {
		return
	}
	maybePersistFloatPref(a.myApp.Preferences(), prefScanTabSplitOffset, a.scanTabMainSplit.Offset,
		&a.scanTabSplitPersistPrimed, &a.lastPersistedScanSplit, nil)
}

// loadTopologySplitFromPrefs загружает позицию сплита вкладки топологии.
func (a *App) loadTopologySplitFromPrefs() {
	if a == nil || a.topologyMainSplit == nil || a.myApp == nil {
		return
	}
	v := a.myApp.Preferences().FloatWithFallback(prefTopologyMainSplitOffset, -1)
	if v >= 0.18 && v <= 0.88 {
		a.topologyMainSplit.Offset = v
		a.topologySplitInitialized = true
		a.topologySplitPersistPrimed = true
		a.lastPersistedTopologySplit = v
	}
}

// clampTopologyMainSplitOffset ограничивает позицию сплита вкладки топологии.
func (a *App) clampTopologyMainSplitOffset() {
	if a == nil || a.topologyMainSplit == nil {
		return
	}
	const lo, hi = 0.18, 0.85
	o := a.topologyMainSplit.Offset
	if o < lo {
		a.topologyMainSplit.Offset = lo
	} else if o > hi {
		a.topologyMainSplit.Offset = hi
	}
}

// maybePersistTopologySplitOffset сохраняет позицию сплита вкладки топологии.
func (a *App) maybePersistTopologySplitOffset() {
	if a == nil || a.topologyMainSplit == nil || a.myApp == nil {
		return
	}
	maybePersistFloatPref(a.myApp.Preferences(), prefTopologyMainSplitOffset, a.topologyMainSplit.Offset,
		&a.topologySplitPersistPrimed, &a.lastPersistedTopologySplit, nil)
}

// loadToolsTabSplitFromPrefs загружает позицию сплита вкладки инструментов.
func (a *App) loadToolsTabSplitFromPrefs() {
	if a == nil || a.toolsTabMainSplit == nil || a.myApp == nil {
		return
	}
	v := a.myApp.Preferences().FloatWithFallback(prefToolsTabSplitOffset, -1)
	if v >= 0.22 && v <= 0.82 {
		a.toolsTabMainSplit.Offset = v
		a.toolsSplitInitialized = true
		a.toolsSplitPersistPrimed = true
		a.lastPersistedToolsSplit = v
	}
}

// clampToolsTabMainSplitOffset ограничивает позицию сплита вкладки инструментов.
func (a *App) clampToolsTabMainSplitOffset() {
	if a == nil || a.toolsTabMainSplit == nil {
		return
	}
	const lo, hi = 0.22, 0.78
	o := a.toolsTabMainSplit.Offset
	if o < lo {
		a.toolsTabMainSplit.Offset = lo
	} else if o > hi {
		a.toolsTabMainSplit.Offset = hi
	}
}

// maybePersistToolsTabSplitOffset сохраняет позицию сплита вкладки инструментов.
func (a *App) maybePersistToolsTabSplitOffset() {
	if a == nil || a.toolsTabMainSplit == nil || a.myApp == nil {
		return
	}
	maybePersistFloatPref(a.myApp.Preferences(), prefToolsTabSplitOffset, a.toolsTabMainSplit.Offset,
		&a.toolsSplitPersistPrimed, &a.lastPersistedToolsSplit, nil)
}

// --- Методы Split-настроек Host Details ---

// loadHostDetailsSplitFromPrefs загружает позиции сплитов Host Details.
func (a *App) loadHostDetailsSplitFromPrefs() {
	if a == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	if v := p.FloatWithFallback(prefHostDetailsSplitOffsetV, -1); v >= 0.28 && v <= 0.92 {
		a.rememberedHostDetailsSplitV = v
		a.hostDetailsSplitPrimedV = true
		a.lastPersistedHostDetailsV = v
	}
	if h := p.FloatWithFallback(prefHostDetailsSplitOffsetH, -1); h >= 0.35 && h <= 0.90 {
		a.rememberedHostDetailsSplitH = h
		a.hostDetailsSplitPrimedH = true
		a.lastPersistedHostDetailsH = h
	}
}

// maybePersistHostDetailsSplitOffsets сохраняет позиции сплитов Host Details.
func (a *App) maybePersistHostDetailsSplitOffsets() {
	if a == nil || a.resultsMainSplit == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	cur := a.resultsMainSplit.Offset
	switch a.lastHostDetailsSplitKind {
	case "V":
		maybePersistFloatPref(p, prefHostDetailsSplitOffsetV, cur, &a.hostDetailsSplitPrimedV, &a.lastPersistedHostDetailsV, func(v float64) {
			a.rememberedHostDetailsSplitV = v
		})
	case "H":
		maybePersistFloatPref(p, prefHostDetailsSplitOffsetH, cur, &a.hostDetailsSplitPrimedH, &a.lastPersistedHostDetailsH, func(v float64) {
			a.rememberedHostDetailsSplitH = v
		})
	}
}

// --- Методы загрузки настроек ---

// --- Методы Filter Presets ---
