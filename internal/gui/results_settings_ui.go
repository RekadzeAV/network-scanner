package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"network-scanner/internal/display"
	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
)

// saveResultsViewSettings сохраняет настройки отображения результатов.
func (a *App) saveResultsViewSettings() {
	if a == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	p.SetString(prefViewMode, strings.TrimSpace(a.resultsMode))
	p.SetString(prefResultsSubMode, strings.TrimSpace(a.resultsSubMode))
	p.SetString(prefSortMode, strings.TrimSpace(a.resultsSort))
	p.SetString(prefChipLimit, strconv.Itoa(a.maxPortChips))
	if a.showRawBanners {
		p.SetString(prefShowRawBanners, "true")
	} else {
		p.SetString(prefShowRawBanners, "false")
	}
	p.SetString(prefFilterQuery, strings.TrimSpace(a.resultsFilterQuery))
	if a.onlyWithOpenPorts {
		p.SetString(prefOnlyOpenPorts, "true")
	} else {
		p.SetString(prefOnlyOpenPorts, "false")
	}
	selectedTypes := make([]string, 0)
	for typeName, check := range a.quickTypeChecks {
		if check != nil && check.Checked {
			selectedTypes = append(selectedTypes, typeName)
		}
	}
	if len(selectedTypes) > 1 {
		// Keep serialized settings deterministic for easier debugging.
		sort.Strings(selectedTypes)
	}
	p.SetString(prefTypeFilters, strings.Join(selectedTypes, ","))
	if a.resultsCidrFilterEnt != nil {
		p.SetString(prefCidrFilter, strings.TrimSpace(a.resultsCidrFilterEnt.Text))
	}
	mode := strings.TrimSpace(a.resultsPortStateMode)
	if mode == "" {
		mode = "all"
	}
	p.SetString(prefPortStateMode, mode)
}

// resultsForSave возвращает результаты для сохранения и причину неудачи.
func (a *App) resultsForSave() ([]scanner.Result, string) {
	if len(a.scanResults) == 0 {
		return nil, "Нет результатов для сохранения"
	}
	resultsToSave := a.currentDisplayedResults()
	if len(resultsToSave) == 0 {
		return nil, "После применения фильтров нет данных для сохранения"
	}
	return resultsToSave, ""
}

// saveResults сохраняет результаты в файл
func (a *App) saveResults() {
	resultsToSave, reason := a.resultsForSave()
	if reason != "" {
		dialog.ShowInformation("Информация", reason, a.myWindow)
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.myWindow)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// Форматируем результаты в текстовый формат
		text := display.FormatResultsAsText(resultsToSave)

		_, err = writer.Write([]byte(text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка при сохранении файла: %v", err), a.myWindow)
			return
		}

		dialog.ShowInformation("Успех", fmt.Sprintf("Результаты успешно сохранены (устройств: %d)", len(resultsToSave)), a.myWindow)
	}, a.myWindow)
}

// formatDurationMMSS форматирует длительность в ММ:СС.
func formatDurationMMSS(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSec := int(d.Round(time.Second).Seconds())
	min := totalSec / 60
	sec := totalSec % 60
	return fmt.Sprintf("%02d:%02d", min, sec)
}

// partialSNMPKeysFromReport собирает ключи SNMP из отчёта.
func partialSNMPKeysFromReport(report *snmpcollector.CollectReport) map[string]struct{} {
	if report == nil {
		return nil
	}
	out := make(map[string]struct{})
	for _, f := range report.Failures {
		if f.Kind != snmpcollector.FailureQuery {
			continue
		}
		ip := strings.TrimSpace(strings.ToLower(f.IP))
		if ip != "" {
			out["ip:"+ip] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// copyScanDiagnostics копирует диагностику сканирования в буфер обмена.
func (a *App) copyScanDiagnostics() {
	if a == nil || a.diagnosticsLabel == nil {
		return
	}
	text := strings.TrimSpace(a.diagnosticsLabel.Text)
	if text == "" || strings.Contains(strings.ToLower(text), "n/a") || strings.Contains(strings.ToLower(text), "выполняется") {
		dialog.ShowInformation("Информация", "Диагностика сканирования пока недоступна", a.myWindow)
		return
	}
	a.myWindow.Clipboard().SetContent(text)
	dialog.ShowInformation("Готово", "Диагностика сканирования скопирована в буфер обмена", a.myWindow)
}

// saveScanDiagnostics сохраняет диагностику сканирования в файл.
func (a *App) saveScanDiagnostics() {
	if a == nil || a.diagnosticsLabel == nil {
		return
	}
	text := strings.TrimSpace(a.diagnosticsLabel.Text)
	if text == "" || strings.Contains(strings.ToLower(text), "n/a") || strings.Contains(strings.ToLower(text), "выполняется") {
		dialog.ShowInformation("Информация", "Диагностика сканирования пока недоступна", a.myWindow)
		return
	}

	defaultFileName := fmt.Sprintf("scan-diagnostics-%s.txt", time.Now().Format("2006-01-02-150405"))
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.myWindow)
			return
		}
		if writer == nil {
			return
		}
		targetPath := writer.URI().Path()
		normalizedPath := targetPath
		if strings.ToLower(filepath.Ext(normalizedPath)) != ".txt" {
			normalizedPath += ".txt"
		}

		if normalizedPath == targetPath {
			defer writer.Close()
			if _, writeErr := writer.Write([]byte(text)); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении диагностики: %v", writeErr), a.myWindow)
				return
			}
		} else {
			_ = writer.Close()
			if writeErr := os.WriteFile(normalizedPath, []byte(text), 0644); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении диагностики: %v", writeErr), a.myWindow)
				return
			}
		}
		dialog.ShowInformation("Готово", fmt.Sprintf("Диагностика сканирования сохранена: %s", normalizedPath), a.myWindow)
	}, a.myWindow)
	saveDialog.SetFileName(defaultFileName)
	saveDialog.Show()
}

// buildPerformanceReportText формирует текстовый отчёт производительности.
func (a *App) buildPerformanceReportText() string {
	if a.lastTopology == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Отчет производительности topology build\n")
	sb.WriteString(fmt.Sprintf("Устройств: %d\n", len(a.lastTopology.Devices)))
	sb.WriteString(fmt.Sprintf("Связей: %d\n", len(a.lastTopology.Links)))
	if a.lastTopoMetric.snmpDuration > 0 {
		sb.WriteString(fmt.Sprintf("SNMP сбор: %s\n", a.lastTopoMetric.snmpDuration.Round(time.Millisecond).String()))
	}
	if a.lastTopoMetric.buildDuration > 0 {
		sb.WriteString(fmt.Sprintf("Построение графа: %s\n", a.lastTopoMetric.buildDuration.Round(time.Millisecond).String()))
	}
	if a.lastTopoMetric.totalDuration > 0 {
		sb.WriteString(fmt.Sprintf("Общее время: %s\n", a.lastTopoMetric.totalDuration.Round(time.Millisecond).String()))
	}
	if a.lastSNMPReport != nil {
		sb.WriteString(fmt.Sprintf("SNMP целей: %d\n", a.lastSNMPReport.TotalSNMPTargets))
		sb.WriteString(fmt.Sprintf("SNMP ok: %d\n", a.lastSNMPReport.Connected))
		sb.WriteString(fmt.Sprintf("SNMP partial: %d\n", a.lastSNMPReport.Partial))
		sb.WriteString(fmt.Sprintf("SNMP failed: %d\n", a.lastSNMPReport.Failed))
	}
	return sb.String()
}

// resetUIPanelLayoutWithFeedback сбрасывает расположение панелей с подтверждением.
func (a *App) resetUIPanelLayoutWithFeedback() {
	if a == nil {
		return
	}
	a.settingsMgr.ResetUIPanelLayout(a.scanTabMainSplit, a.topologyMainSplit, a.toolsTabMainSplit)
	if a.myWindow != nil {
		dialog.ShowInformation("Вид", layoutResetInfoMessage, a.myWindow)
	}
}

// resetUIPanelLayout сбрасывает расположение панелей без подтверждения.
func (a *App) resetUIPanelLayout() {
	if a == nil || a.myApp == nil {
		return
	}
	a.settingsMgr.ClearSplitPreferences()

	a.rememberedHostDetailsSplitV = 0
	a.rememberedHostDetailsSplitH = 0
	a.lastHostDetailsSplitKind = ""
	a.hostDetailsSplitPrimedV = false
	a.hostDetailsSplitPrimedH = false

	prof := strings.TrimSpace(a.layoutProfile)
	if prof == "" {
		prof = "normal"
	}
	if a.myWindow != nil {
		fyne.Do(func() {
			a.settingsMgr.ApplyDefaultSplitOffsetsForProfile(prof)
			a.renderScanResultsView()
			if a.myWindow.Content() != nil {
				a.myWindow.Content().Refresh()
			}
		})
	} else {
		a.settingsMgr.ApplyDefaultSplitOffsetsForProfile(prof)
		a.renderScanResultsView()
	}
}
