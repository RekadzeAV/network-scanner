package gui

import (
	"fmt"
	"sort"
	"strings"
)

// filterPresetKey возвращает ключ preference для слота пресета.
func (a *App) filterPresetKey(slot string) string {
	switch strings.TrimSpace(slot) {
	case "1":
		return prefFilterPreset1
	case "2":
		return prefFilterPreset2
	case "3":
		return prefFilterPreset3
	default:
		return prefFilterPreset1
	}
}

// serializeCurrentFilters сериализует текущие фильтры в строку.
func (a *App) serializeCurrentFilters() string {
	selectedTypes := make([]string, 0)
	for typeName, check := range a.quickTypeChecks {
		if check != nil && check.Checked {
			selectedTypes = append(selectedTypes, typeName)
		}
	}
	if len(selectedTypes) > 1 {
		sort.Strings(selectedTypes)
	}
	cidr := ""
	if a.resultsCidrFilterEnt != nil {
		cidr = strings.TrimSpace(a.resultsCidrFilterEnt.Text)
	}
	mode := strings.TrimSpace(a.resultsPortStateMode)
	if mode == "" {
		mode = "all"
	}
	onlyOpen := "false"
	if a.onlyWithOpenPorts {
		onlyOpen = "true"
	}
	return strings.Join([]string{
		strings.TrimSpace(a.resultsFilterQuery),
		cidr,
		mode,
		onlyOpen,
		strings.Join(selectedTypes, ","),
	}, "|")
}

// saveFilterPreset сохраняет текущие фильтры в выбранный слот пресета.
func (a *App) saveFilterPreset(slot string) {
	if a == nil || a.myApp == nil {
		return
	}
	key := a.filterPresetKey(slot)
	a.myApp.Preferences().SetString(key, a.serializeCurrentFilters())
	if a.statusLabel != nil {
		a.statusLabel.SetText(fmt.Sprintf("Пресет фильтров %s сохранен", slot))
		a.statusLabel.Refresh()
	}
}

// applyFilterPreset применяет фильтры из выбранного слота пресета.
func (a *App) applyFilterPreset(slot string) {
	if a == nil || a.myApp == nil {
		return
	}
	key := a.filterPresetKey(slot)
	raw := strings.TrimSpace(a.myApp.Preferences().String(key))
	if raw == "" {
		if a.statusLabel != nil {
			a.statusLabel.SetText(fmt.Sprintf("Пресет фильтров %s пуст", slot))
			a.statusLabel.Refresh()
		}
		return
	}
	parts := strings.SplitN(raw, "|", 5)
	if len(parts) < 5 {
		if a.statusLabel != nil {
			a.statusLabel.SetText(fmt.Sprintf("Пресет фильтров %s поврежден", slot))
			a.statusLabel.Refresh()
		}
		return
	}
	a.resultsFilterQuery = strings.TrimSpace(parts[0])
	if a.resultsFilterEnt != nil {
		a.resultsFilterEnt.SetText(a.resultsFilterQuery)
	}
	if a.resultsCidrFilterEnt != nil {
		a.resultsCidrFilterEnt.SetText(strings.TrimSpace(parts[1]))
	}
	a.resultsPortStateMode = strings.TrimSpace(parts[2])
	switch a.resultsPortStateMode {
	case "has_open":
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть открытые")
		}
	case "has_closed":
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть закрытые")
		}
	case "has_filtered":
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть фильтруемые")
		}
	default:
		a.resultsPortStateMode = "all"
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Все")
		}
	}
	a.onlyWithOpenPorts = strings.EqualFold(strings.TrimSpace(parts[3]), "true")
	if a.openPortsOnlyCheck != nil {
		a.openPortsOnlyCheck.SetChecked(a.onlyWithOpenPorts)
	}
	for _, ch := range a.quickTypeChecks {
		if ch != nil {
			ch.SetChecked(false)
		}
	}
	if typeCSV := strings.TrimSpace(parts[4]); typeCSV != "" {
		for _, typeName := range strings.Split(typeCSV, ",") {
			name := strings.TrimSpace(typeName)
			if check, ok := a.quickTypeChecks[name]; ok && check != nil {
				check.SetChecked(true)
			}
		}
	}
	a.scheduleResultsRender(true)
}
