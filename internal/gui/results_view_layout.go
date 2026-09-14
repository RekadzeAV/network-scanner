package gui

import (
	"math"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
)

// currentLayoutProfile возвращает текущий профиль макета.
func (a *App) currentLayoutProfile() string {
	if a == nil {
		return "normal"
	}
	if a.layoutProfile == "" {
		return "normal"
	}
	return a.layoutProfile
}

// detectLayoutProfile определяет профиль макета по ширине окна.
func (a *App) detectLayoutProfile(width float32) string {
	switch {
	case width <= 1366:
		return "compact"
	case width >= 2200:
		return "wide"
	default:
		return "normal"
	}
}

func clampFloat32(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampFloat64(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func absFloat32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// layoutAdaptiveMultiplier подбирает коэффициент для минимальных высот панелей по логическому
// размеру окна, «диагонали» (гипотенуза W×H) и масштабу канвы (плотность пикселей / HiDPI).
func layoutAdaptiveMultiplier(canvasW, canvasH, canvasScale float32) float32 {
	if canvasW <= 0 || canvasH <= 0 {
		return 1
	}
	if canvasScale <= 0 {
		canvasScale = 1
	}
	d := float32(math.Hypot(float64(canvasW), float64(canvasH)))
	const refDiag = float32(1150)
	ratio := d / refDiag
	scaleAdj := float32(math.Sqrt(float64(canvasScale)))
	scaleAdj = clampFloat32(scaleAdj, 0.88, 1.38)
	return clampFloat32(ratio/scaleAdj, 0.72, 1.38)
}

func suggestedScanTabOffset(profile string, canvasW, canvasH, canvasScale, layoutMul float32) float32 {
	_ = canvasW
	_ = canvasScale
	if canvasH <= 0 {
		return 0.38
	}
	topFrac := float32(0.36)
	switch profile {
	case "compact":
		topFrac = 0.41
	case "wide":
		topFrac = 0.32
	}
	if canvasH < 680 {
		topFrac -= 0.045
	}
	if canvasH > 920 {
		topFrac += 0.035
	}
	topFrac += (layoutMul - 1) * 0.025
	return clampFloat32(topFrac, 0.26, 0.54)
}

// defaultTopologySplitOffset — стартовая доля верхней панели (превью) для вкладки «Топология».
func defaultTopologySplitOffset(profile string) float64 {
	switch profile {
	case "compact":
		return 0.72
	case "wide":
		return 0.6
	default:
		return 0.62
	}
}

// defaultToolsSplitOffset — стартовая доля верхней зоны (поля + Operations) на вкладке «Инструменты».
func defaultToolsSplitOffset(profile string) float64 {
	switch profile {
	case "compact":
		return 0.48
	case "wide":
		return 0.40
	default:
		return 0.44
	}
}

func (a *App) currentLayoutAdaptiveMultiplier() float32 {
	if a == nil {
		return 1
	}
	w, h := a.lastCanvasSize.Width, a.lastCanvasSize.Height
	s := a.lastCanvasScale
	if w <= 0 || h <= 0 {
		return 1
	}
	if s <= 0 {
		s = 1
	}
	return layoutAdaptiveMultiplier(w, h, s)
}

func (a *App) adaptivePanelMinHeight(base, layoutMul, maxFracWindow, minAbs float32) float32 {
	if a == nil {
		return base
	}
	h := a.lastCanvasSize.Height
	if h <= 0 {
		h = 720
	}
	v := base * layoutMul
	maxH := h * maxFracWindow
	if v > maxH {
		v = maxH
	}
	if v < minAbs {
		v = minAbs
	}
	return v
}

func (a *App) applyAdaptiveToolsScrollMinSizes(profile string, mul float32) {
	if a == nil {
		return
	}
	baseOut := float32(360)
	baseOps := float32(140)
	switch profile {
	case "compact":
		baseOut, baseOps = 280, 110
	case "wide":
		baseOut, baseOps = 420, 170
	}
	if a.toolsOutputScroll != nil {
		a.toolsOutputScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(baseOut, mul, 0.52, 150)))
	}
	if a.operationsOutputScroll != nil {
		a.operationsOutputScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(baseOps, mul, 0.22, 72)))
	}
}

func (a *App) applyResponsiveLayout(profile string) {
	if a == nil {
		return
	}
	a.layoutProfile = profile
	mul := a.currentLayoutAdaptiveMultiplier()
	switch profile {
	case "compact":
		if a.resultsDiagnosticsGrid != nil {
			a.resultsDiagnosticsGrid.Layout = layout.NewGridLayoutWithColumns(1)
			a.resultsDiagnosticsGrid.Refresh()
		}
		if a.resultsSortGrid != nil {
			a.resultsSortGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.resultsSortGrid.Refresh()
		}
		if a.resultsCidrGrid != nil {
			a.resultsCidrGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.resultsCidrGrid.Refresh()
		}
		if a.resultsPresetGrid != nil {
			a.resultsPresetGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.resultsPresetGrid.Refresh()
		}
		if a.scanControlsScroll != nil {
			a.scanControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(170, mul, 0.40, 110)))
		}
		if a.resultsScroll != nil {
			a.resultsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(48, mul, 0.34, 32)))
		}
		if a.topologyControlsScroll != nil {
			a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(150, mul, 0.32, 100)))
		}
		if a.topologyScroll != nil {
			a.topologyScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(200, mul, 0.42, 120)))
		}
		if a.toolsControlsScroll != nil {
			a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(180, mul, 0.38, 120)))
		}
		if a.toolButtonsGrid != nil {
			a.toolButtonsGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.toolButtonsGrid.Refresh()
		}
		if a.operationsHeaderGrid != nil {
			a.operationsHeaderGrid.Layout = layout.NewGridLayoutWithColumns(1)
			a.operationsHeaderGrid.Refresh()
		}
	case "wide":
		if a.resultsDiagnosticsGrid != nil {
			a.resultsDiagnosticsGrid.Layout = layout.NewGridLayoutWithColumns(3)
			a.resultsDiagnosticsGrid.Refresh()
		}
		if a.resultsSortGrid != nil {
			a.resultsSortGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.resultsSortGrid.Refresh()
		}
		if a.resultsCidrGrid != nil {
			a.resultsCidrGrid.Layout = layout.NewGridLayoutWithColumns(4)
			a.resultsCidrGrid.Refresh()
		}
		if a.resultsPresetGrid != nil {
			a.resultsPresetGrid.Layout = layout.NewGridLayoutWithColumns(4)
			a.resultsPresetGrid.Refresh()
		}
		if a.scanControlsScroll != nil {
			a.scanControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(250, mul, 0.42, 140)))
		}
		if a.resultsScroll != nil {
			a.resultsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(70, mul, 0.34, 44)))
		}
		if a.topologyControlsScroll != nil {
			a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(220, mul, 0.34, 140)))
		}
		if a.topologyScroll != nil {
			a.topologyScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(350, mul, 0.48, 180)))
		}
		if a.toolsControlsScroll != nil {
			a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(320, mul, 0.40, 160)))
		}
		if a.toolButtonsGrid != nil {
			a.toolButtonsGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.toolButtonsGrid.Refresh()
		}
		if a.operationsHeaderGrid != nil {
			a.operationsHeaderGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.operationsHeaderGrid.Refresh()
		}
	default:
		if a.resultsDiagnosticsGrid != nil {
			a.resultsDiagnosticsGrid.Layout = layout.NewGridLayoutWithColumns(3)
			a.resultsDiagnosticsGrid.Refresh()
		}
		if a.resultsSortGrid != nil {
			a.resultsSortGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.resultsSortGrid.Refresh()
		}
		if a.resultsCidrGrid != nil {
			a.resultsCidrGrid.Layout = layout.NewGridLayoutWithColumns(4)
			a.resultsCidrGrid.Refresh()
		}
		if a.resultsPresetGrid != nil {
			a.resultsPresetGrid.Layout = layout.NewGridLayoutWithColumns(4)
			a.resultsPresetGrid.Refresh()
		}
		if a.scanControlsScroll != nil {
			a.scanControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(220, mul, 0.38, 130)))
		}
		if a.resultsScroll != nil {
			a.resultsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(float32(75*0.775), mul, 0.34, 40)))
		}
		if a.topologyControlsScroll != nil {
			a.topologyControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(200, mul, 0.32, 120)))
		}
		if a.topologyScroll != nil {
			a.topologyScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(300, mul, 0.45, 160)))
		}
		if a.toolsControlsScroll != nil {
			a.toolsControlsScroll.SetMinSize(fyne.NewSize(0, a.adaptivePanelMinHeight(260, mul, 0.38, 140)))
		}
		if a.toolButtonsGrid != nil {
			a.toolButtonsGrid.Layout = layout.NewGridLayoutWithColumns(5)
			a.toolButtonsGrid.Refresh()
		}
		if a.operationsHeaderGrid != nil {
			a.operationsHeaderGrid.Layout = layout.NewGridLayoutWithColumns(2)
			a.operationsHeaderGrid.Refresh()
		}
	}
	a.applyAdaptiveToolsScrollMinSizes(profile, mul)
	a.clampScanTabMainSplitOffset()
	a.clampTopologyMainSplitOffset()
	a.clampToolsTabMainSplitOffset()
	if a.resultsBody != nil {
		a.renderScanResultsView()
	}
}

func (a *App) startResultsLayoutWatcher() {
	if a == nil || a.myWindow == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()
		var (
			lastProfile      string
			lastW            float32 = -1
			lastH            float32 = -1
			lastScaleApplied float32 = -1
		)
		const sizeStep = 48
		for range ticker.C {
			size := fyne.NewSize(0, 0)
			scale := float32(1)
			fyne.DoAndWait(func() {
				if a.myWindow == nil || a.myWindow.Canvas() == nil {
					return
				}
				c := a.myWindow.Canvas()
				size = c.Size()
				scale = c.Scale()
			})
			if size.Width <= 0 || size.Height <= 0 {
				continue
			}
			if scale <= 0 {
				scale = 1
			}
			profile := a.detectLayoutProfile(size.Width)
			sizeChanged := lastW < 0 ||
				absFloat32(size.Width-lastW) >= sizeStep ||
				absFloat32(size.Height-lastH) >= sizeStep ||
				absFloat32(scale-lastScaleApplied) >= 0.06
			if profile == lastProfile && !sizeChanged {
				continue
			}
			lastW = size.Width
			lastH = size.Height
			lastScaleApplied = scale
			lastProfile = profile
			fyne.Do(func() {
				a.lastCanvasSize = size
				a.lastCanvasScale = scale
				a.applyResponsiveLayout(profile)
				if a.scanTabMainSplit != nil && !a.scanTabSplitInitialized && size.Height > 200 {
					a.scanTabSplitInitialized = true
					m := layoutAdaptiveMultiplier(size.Width, size.Height, scale)
					a.scanTabMainSplit.Offset = float64(suggestedScanTabOffset(profile, size.Width, size.Height, scale, m))
					a.clampScanTabMainSplitOffset()
				}
				if a.topologyMainSplit != nil && !a.topologySplitInitialized && size.Height > 200 {
					a.topologySplitInitialized = true
					a.topologyMainSplit.Offset = defaultTopologySplitOffset(profile)
					a.clampTopologyMainSplitOffset()
				}
				if a.toolsTabMainSplit != nil && !a.toolsSplitInitialized && size.Height > 200 {
					a.toolsSplitInitialized = true
					a.toolsTabMainSplit.Offset = defaultToolsSplitOffset(profile)
					a.clampToolsTabMainSplitOffset()
				}
				a.maybePersistScanTabSplitOffset()
				a.maybePersistTopologySplitOffset()
				a.maybePersistToolsTabSplitOffset()
				a.maybePersistHostDetailsSplitOffsets()
			})
		}
	}()
}

// applyDefaultSplitOffsetsForProfile выставляет разделители вкладок по профилю и размеру окна
// и записывает их в preferences (используется из меню «Вид» → сброс).
func (a *App) applyDefaultSplitOffsetsForProfile(profile string) {
	if a == nil {
		return
	}
	prof := strings.TrimSpace(profile)
	switch prof {
	case "compact", "wide", "normal":
	default:
		prof = "normal"
	}
	w, h := a.lastCanvasSize.Width, a.lastCanvasSize.Height
	if w <= 0 || h <= 0 {
		w, h = 1280, 720
	}
	s := a.lastCanvasScale
	if s <= 0 {
		s = 1
	}
	m := layoutAdaptiveMultiplier(w, h, s)
	if a.scanTabMainSplit != nil {
		a.scanTabMainSplit.Offset = float64(suggestedScanTabOffset(prof, w, h, s, m))
		a.clampScanTabMainSplitOffset()
		a.scanTabSplitInitialized = true
		a.scanTabSplitPersistPrimed = true
		a.lastPersistedScanSplit = a.scanTabMainSplit.Offset
		if a.myApp != nil {
			a.myApp.Preferences().SetFloat(prefScanTabSplitOffset, a.scanTabMainSplit.Offset)
		}
	}
	if a.topologyMainSplit != nil {
		a.topologyMainSplit.Offset = defaultTopologySplitOffset(prof)
		a.clampTopologyMainSplitOffset()
		a.topologySplitInitialized = true
		a.topologySplitPersistPrimed = true
		a.lastPersistedTopologySplit = a.topologyMainSplit.Offset
		if a.myApp != nil {
			a.myApp.Preferences().SetFloat(prefTopologyMainSplitOffset, a.topologyMainSplit.Offset)
		}
	}
	if a.toolsTabMainSplit != nil {
		a.toolsTabMainSplit.Offset = defaultToolsSplitOffset(prof)
		a.clampToolsTabMainSplitOffset()
		a.toolsSplitInitialized = true
		a.toolsSplitPersistPrimed = true
		a.lastPersistedToolsSplit = a.toolsTabMainSplit.Offset
		if a.myApp != nil {
			a.myApp.Preferences().SetFloat(prefToolsTabSplitOffset, a.toolsTabMainSplit.Offset)
		}
	}
}
