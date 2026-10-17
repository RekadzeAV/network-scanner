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

// updateResultsCountLabel обновляет счётчик строк в шапке области результатов
// (R-09): «Строк: показано X из Y», где Y — всего после фильтрации.
func (a *App) updateResultsCountLabel(visible, total int) {
	if a == nil || a.resultsCountLabel == nil {
		return
	}
	if visible < 0 {
		visible = 0
	}
	if total < 0 {
		total = 0
	}
	if visible == total {
		a.resultsCountLabel.SetText(fmt.Sprintf("Строк: %d", total))
		return
	}
	a.resultsCountLabel.SetText(fmt.Sprintf("Строк: показано %d из %d", visible, total))
}
