package gui

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"

	"network-scanner/internal/scanner"
)

// filteredSortedResults применяет фильтры и сортировку к a.scanResults.
func (a *App) filteredSortedResults() []scanner.Result {
	if len(a.scanResults) == 0 {
		return nil
	}
	cacheKey := a.buildResultsPipelineCacheKey()
	a.resultsPipelineCacheMu.RLock()
	if cacheKey != "" && cacheKey == a.resultsPipelineCacheKey && len(a.resultsPipelineCacheData) > 0 {
		cached := a.resultsPipelineCacheData
		a.resultsPipelineCacheMu.RUnlock()
		return cached
	}
	a.resultsPipelineCacheMu.RUnlock()
	base := filterResultsForDisplayAdvanced(
		a.scanResults,
		a.resultsFilterQuery,
		a.selectedTypeFilters(),
		a.onlyWithOpenPorts,
	)
	out := a.applyAdvancedFilters(base)
	sorted := sortedResultsForDisplayWithMode(out, a.resultsSort)
	a.resultsPipelineCacheMu.Lock()
	a.resultsPipelineCacheKey = cacheKey
	a.resultsPipelineCacheData = sorted
	a.resultsPipelineCacheMu.Unlock()
	return sorted
}

func (a *App) currentDisplayedResults() []scanner.Result {
	return a.filteredSortedResults()
}

func (a *App) selectedTypeFilters() []string {
	if a == nil || len(a.quickTypeChecks) == 0 {
		return nil
	}
	out := make([]string, 0, len(a.quickTypeChecks))
	for name, ch := range a.quickTypeChecks {
		if ch != nil && ch.Checked {
			out = append(out, strings.TrimSpace(name))
		}
	}
	sort.Strings(out)
	return out
}

func (a *App) buildResultsPipelineCacheKey() string {
	if a == nil {
		return ""
	}
	parts := []string{
		strconv.FormatUint(a.scanResultsVersion, 10),
		strings.TrimSpace(a.resultsFilterQuery),
		strings.TrimSpace(a.resultsSort),
		strings.TrimSpace(a.resultsPortStateMode),
		strconv.FormatBool(a.onlyWithOpenPorts),
	}
	if a.resultsCidrFilterEnt != nil {
		parts = append(parts, strings.TrimSpace(a.resultsCidrFilterEnt.Text))
	}
	parts = append(parts, strings.Join(a.selectedTypeFilters(), ","))
	return strings.Join(parts, "|")
}

func (a *App) invalidateResultsPipelineCache() {
	if a == nil {
		return
	}
	a.resultsPipelineCacheMu.Lock()
	a.resultsPipelineCacheKey = ""
	a.resultsPipelineCacheData = nil
	a.resultsPipelineCacheMu.Unlock()
}

func (a *App) applyAdvancedFilters(base []scanner.Result) []scanner.Result {
	out := make([]scanner.Result, 0, len(base))
	for _, r := range base {
		if !a.passesCIDRFilter(r) {
			continue
		}
		if !a.passesPortStateMode(r) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func (a *App) passesCIDRFilter(r scanner.Result) bool {
	if a.resultsCidrFilterEnt == nil {
		return true
	}
	cidr := strings.TrimSpace(a.resultsCidrFilterEnt.Text)
	if cidr == "" {
		return true
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return true
	}
	ip := net.ParseIP(strings.TrimSpace(r.IP))
	if ip == nil {
		return false
	}
	return ipnet.Contains(ip)
}

func (a *App) passesPortStateMode(r scanner.Result) bool {
	mode := strings.TrimSpace(a.resultsPortStateMode)
	if mode == "" || mode == "all" {
		return true
	}
	switch mode {
	case "has_open":
		for _, p := range r.Ports {
			if p.State == "open" {
				return true
			}
		}
		return false
	case "has_closed":
		for _, p := range r.Ports {
			if p.State == "closed" {
				return true
			}
		}
		return false
	case "has_filtered":
		for _, p := range r.Ports {
			if p.State == "filtered" {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func (a *App) activeFilterCount() int {
	n := 0
	if strings.TrimSpace(a.resultsFilterQuery) != "" {
		n++
	}
	for _, ch := range a.quickTypeChecks {
		if ch != nil && ch.Checked {
			n++
		}
	}
	if a.onlyWithOpenPorts {
		n++
	}
	if a.resultsCidrFilterEnt != nil && strings.TrimSpace(a.resultsCidrFilterEnt.Text) != "" {
		n++
	}
	if a.resultsPortStateMode != "" && a.resultsPortStateMode != "all" {
		n++
	}
	return n
}

func (a *App) updateFiltersInfoLabel() {
	if a.filtersInfoLabel == nil {
		return
	}
	a.filtersInfoLabel.SetText(fmt.Sprintf("Активных фильтров: %d", a.activeFilterCount()))
}

// adaptiveDebounce вычисляет оптимальную задержку debounce на основе производительности.
func (a *App) adaptiveDebounce() time.Duration {
	// Используем lastRenderStats для оценки производительности
	if a.lastRenderStats.Duration > 100*time.Millisecond {
		return 300 * time.Millisecond // Медленный рендер — больше debounce
	}
	if a.lastRenderStats.Duration > 50*time.Millisecond {
		return 180 * time.Millisecond // Средний рендер
	}
	return 100 * time.Millisecond // Быстрый рендер — минимальная задержка
}

func (a *App) scheduleResultsRender(immediate bool) {
	if a == nil {
		return
	}
	if immediate || a.resultsRenderDebounce <= 0 {
		a.cancelPendingResultsRender()
		a.renderScanResultsView()
		return
	}
	a.resultsRenderTimerMu.Lock()
	defer a.resultsRenderTimerMu.Unlock()
	if a.resultsRenderTimer != nil {
		a.resultsRenderTimer.Stop()
	}
	// Используем адаптивный debounce
	delay := a.adaptiveDebounce()
	a.resultsRenderTimer = time.AfterFunc(delay, func() {
		fyne.Do(func() {
			a.renderScanResultsView()
		})
	})
}

func (a *App) cancelPendingResultsRender() {
	if a == nil {
		return
	}
	a.resultsRenderTimerMu.Lock()
	defer a.resultsRenderTimerMu.Unlock()
	if a.resultsRenderTimer != nil {
		a.resultsRenderTimer.Stop()
		a.resultsRenderTimer = nil
	}
}

func (a *App) captureHostDetailsSplitOffsetBeforeRebuild() {
	if a == nil || a.resultsMainSplit == nil || a.lastHostDetailsSplitKind == "" {
		return
	}
	o := a.resultsMainSplit.Offset
	switch a.lastHostDetailsSplitKind {
	case "V":
		if o >= 0.28 && o <= 0.92 {
			a.rememberedHostDetailsSplitV = o
		}
	case "H":
		if o >= 0.35 && o <= 0.90 {
			a.rememberedHostDetailsSplitH = o
		}
	}
}

func (a *App) clearResultsMainSplitRef() {
	if a == nil {
		return
	}
	a.resultsMainSplit = nil
	a.lastHostDetailsSplitKind = ""
}
