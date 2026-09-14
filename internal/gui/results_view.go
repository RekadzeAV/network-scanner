package gui

import (
	"fmt"
	"time"
)

type resultsRenderStats struct {
	FilteredCount int
	VisibleCount  int
	Duration      time.Duration
}

// updateResultsPerfLabel обновляет лейбл производительности рендера.
func (a *App) updateResultsPerfLabel(stats resultsRenderStats) {
	if a == nil || a.resultsPerfLabel == nil {
		return
	}
	if stats.FilteredCount <= 0 {
		a.resultsPerfLabel.SetText("Рендер: n/a")
		return
	}
	a.resultsPerfLabel.SetText(fmt.Sprintf("Рендер: %dms (%d/%d)", stats.Duration.Milliseconds(), stats.VisibleCount, stats.FilteredCount))
}
