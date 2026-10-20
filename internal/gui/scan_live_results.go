package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"

	"network-scanner/internal/scanner"
	scand "network-scanner/internal/scanner/daemon"
)

// liveResultsInterval — минимальный интервал публикации частичных результатов
// во время скана (P1-1). Слишком частая перекладка таблицы «съедает» кадры UI.
const liveResultsInterval = 700 * time.Millisecond

// publishLiveResults выкладывает уже найденные хосты в таблицу до окончания
// скана. Вызывается из goroutine наблюдателя, поэтому все правки UI маршалируются
// через fyne.Do. Возвращает true, если публикация состоялась (для учёта таймера).
func (a *App) publishLiveResults(runner *scand.Runner, lastLive time.Time) bool {
	if a == nil || runner == nil {
		return false
	}
	if !lastLive.IsZero() && time.Since(lastLive) < liveResultsInterval {
		return false
	}
	ns := runner.CurrentScanner()
	if ns == nil {
		return false
	}
	snapshot := snapshotScannerResults(ns.GetResults())
	if len(snapshot) == 0 {
		return false
	}
	fyne.Do(func() {
		a.setScanResults(snapshot)
		a.scheduleResultsRender(false)
	})
	return true
}

// setScanResults публикует срез результатов: данные, версия пайплайна и кэш.
// Вызывается только из UI-потока (scanResults мьютекса не имеет).
func (a *App) setScanResults(results []scanner.Result) {
	if a == nil {
		return
	}
	a.scanResults = results
	a.scanResultsVersion++
	a.invalidateResultsPipelineCache()
	// P-03: инвалидация кэша ячеек при смене данных (не при смене сортировки).
	a.resultsCellCache.invalidate()
}

// snapshotScannerResults копирует срез сканера вместе со слайсами портов:
// GetResults возвращает «живой» срез, который сканер продолжает дополнять.
func snapshotScannerResults(results []scanner.Result) []scanner.Result {
	if len(results) == 0 {
		return nil
	}
	out := make([]scanner.Result, 0, len(results))
	for _, res := range results {
		if len(res.Ports) > 0 {
			ports := make([]scanner.PortInfo, len(res.Ports))
			copy(ports, res.Ports)
			res.Ports = ports
		}
		if len(res.Protocols) > 0 {
			protocols := make([]string, len(res.Protocols))
			copy(protocols, res.Protocols)
			res.Protocols = protocols
		}
		out = append(out, res)
	}
	return out
}

// scanProgressText собирает строку прогресса со счётчиком найденных хостов,
// процентом и оценкой оставшегося времени:
// «Сканирование… 37/254 · найдено 12 · 15% · осталось 00:41».
func scanProgressText(ev scand.Event, started time.Time, found int) string {
	var parts []string
	if ev.Current > 0 && ev.Total > 0 {
		parts = append(parts, fmt.Sprintf("%d/%d", ev.Current, ev.Total))
	} else if ev.Current > 0 {
		parts = append(parts, fmt.Sprintf("%d", ev.Current))
	}
	if found >= 0 {
		parts = append(parts, fmt.Sprintf("найдено %d", found))
	}
	if pct := scanProgressPercent(ev); pct >= 0 {
		parts = append(parts, fmt.Sprintf("%d%%", pct))
	}
	if eta := scanETA(started, ev); eta {
		parts = append(parts, "осталось "+formatDurationShort(scanETADuration(started, ev)))
	}
	if len(parts) == 0 {
		return "Сканирование…"
	}
	return "Сканирование… " + strings.Join(parts, " · ")
}

// scanProgressCompact — компактная версия для центрального блока статус-бара:
// «37/254 · 15% · осталось 00:41». Пустая строка, если прогресс не показан.
func scanProgressCompact(ev scand.Event, started time.Time) string {
	var parts []string
	if ev.Current > 0 && ev.Total > 0 {
		parts = append(parts, fmt.Sprintf("%d/%d", ev.Current, ev.Total))
	}
	if pct := scanProgressPercent(ev); pct >= 0 {
		parts = append(parts, fmt.Sprintf("%d%%", pct))
	}
	if eta := scanETA(started, ev); eta {
		parts = append(parts, "осталось "+formatDurationShort(scanETADuration(started, ev)))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " · ")
}

// scanProgressPercent — процент готовности или -1, если он неизвестен.
func scanProgressPercent(ev scand.Event) int {
	if ev.Total > 0 {
		if ev.Current > ev.Total {
			return 100
		}
		return ev.Current * 100 / ev.Total
	}
	if ev.Percent > 0 {
		return int(ev.Percent)
	}
	return -1
}

// scanETA — можно ли оценивать оставшееся время.
func scanETA(started time.Time, ev scand.Event) bool {
	return !started.IsZero() && ev.Current >= 5 && ev.Total > 0 && ev.Current < ev.Total
}

// scanETADuration — линейная оценка остатка по среднему темпу скана.
func scanETADuration(started time.Time, ev scand.Event) time.Duration {
	elapsed := time.Since(started)
	if elapsed <= 0 || ev.Current <= 0 || ev.Total <= ev.Current {
		return 0
	}
	perHost := elapsed / time.Duration(ev.Current)
	return perHost * time.Duration(ev.Total-ev.Current)
}

// formatDurationShort — mm:ss (или h:mm:ss для длинных сканов).
func formatDurationShort(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
