package gui

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// P2-12: настройка видимых колонок таблицы результатов с сохранением.

// resultsColumnDefs — определения колонок (заголовок, ширина compact/normal/wide,
// сортировка, моноширинный шрифт).
var resultsColumnDefs = []struct {
	Title      string
	TitleShort string
	Widths     [3]float32 // compact, normal, wide
	Sortable   bool
	Monospace  bool
}{
	{"Имя хоста", "Имя", [3]float32{110, 140, 180}, true, false},
	{"IP", "IP", [3]float32{100, 120, 150}, true, true},
	{"MAC", "MAC", [3]float32{96, 130, 170}, false, true},
	{"Тип", "Тип", [3]float32{96, 120, 150}, true, false},
	{"Производитель", "Вендор", [3]float32{110, 120, 170}, false, false},
	{"ОС (оценка)", "ОС", [3]float32{120, 140, 220}, false, false},
	{"SNMP", "SNMP", [3]float32{52, 52, 80}, false, false},
	{"Порты (открытые)", "Порты", [3]float32{180, 280, 420}, false, false},
}

// initVisibleCols инициализирует все колонки видимыми.
func (a *App) initVisibleCols() {
	for i := range a.resultsVisibleCols {
		a.resultsVisibleCols[i] = true
	}
}

// visibleColIndices возвращает индексы видимых колонок.
func (a *App) visibleColIndices() []int {
	indices := make([]int, 0, 8)
	for i, visible := range a.resultsVisibleCols {
		if visible {
			indices = append(indices, i)
		}
	}
	return indices
}

// visibleColCount возвращает количество видимых колонок (минимум 1).
func (a *App) visibleColCount() int {
	count := 0
	for _, v := range a.resultsVisibleCols {
		if v {
			count++
		}
	}
	if count == 0 {
		return 1 // IP всегда виден
	}
	return count
}

// saveVisibleCols сохраняет видимость колонок в preferences (CSV: "0,1,3,7").
func (a *App) saveVisibleCols() {
	if a == nil || a.myApp == nil {
		return
	}
	var parts []string
	for i, v := range a.resultsVisibleCols {
		if v {
			parts = append(parts, strconv.Itoa(i))
		}
	}
	a.myApp.Preferences().SetString(prefResultsVisibleCols, strings.Join(parts, ","))
}

// loadVisibleCols загружает видимость колонок из preferences.
func (a *App) loadVisibleCols() {
	a.initVisibleCols()
	if a == nil || a.myApp == nil {
		return
	}
	raw := strings.TrimSpace(a.myApp.Preferences().String(prefResultsVisibleCols))
	if raw == "" {
		return // все видимы по умолчанию
	}
	// Сбрасываем и устанавливаем только перечисленные индексы.
	for i := range a.resultsVisibleCols {
		a.resultsVisibleCols[i] = false
	}
	for _, s := range strings.Split(raw, ",") {
		idx, err := strconv.Atoi(strings.TrimSpace(s))
		if err == nil && idx >= 0 && idx < len(a.resultsVisibleCols) {
			a.resultsVisibleCols[idx] = true
		}
	}
	// Гарантируем минимум одну видимую колонку.
	if a.visibleColCount() == 0 {
		a.resultsVisibleCols[1] = true // IP
	}
}

// buildColumnsDialog показывает диалог выбора видимых колонок (P2-12).
func (a *App) buildColumnsDialog() {
	if a == nil || a.myWindow == nil {
		return
	}

	profile := a.currentLayoutProfile()
	checks := make([]*widget.Check, len(resultsColumnDefs))
	for i, col := range resultsColumnDefs {
		title := col.Title
		if profile == "compact" {
			title = col.TitleShort
		}
		checks[i] = widget.NewCheck(title, nil)
		checks[i].SetChecked(a.resultsVisibleCols[i])
	}

	content := widget.NewForm()
	for _, c := range checks {
		content.Append("", c)
	}

	applyBtn := widget.NewButton("Применить", func() {
		changed := false
		for i, c := range checks {
			if c.Checked != a.resultsVisibleCols[i] {
				a.resultsVisibleCols[i] = c.Checked
				changed = true
			}
		}
		if changed {
			a.saveVisibleCols()
			a.renderScanResultsView()
		}
	})

	cancelBtn := widget.NewButton("Отмена", func() {
		// Отмена — ничего не делаем.
	})

	grid := container.NewGridWithColumns(2, applyBtn, cancelBtn)
	wrapper := container.NewVBox(content, grid)

	d := dialog.NewCustom("Видимые колонки", "Отмена", wrapper, a.myWindow)
	d.Resize(fyne.NewSize(320, float32(80+len(checks)*36)))
	d.Show()
}

// buildColumnsBtn создаёт кнопку "Колонки" для открытия диалога настройки.
func (a *App) buildColumnsBtn() *widget.Button {
	btn := widget.NewButtonWithIcon("Колонки", iconSettings(), nil)
	btn.Importance = widget.LowImportance
	btn.OnTapped = func() {
		a.buildColumnsDialog()
	}
	return btn
}

// resultsTableHeadersFiltered возвращает заголовки только видимых колонок.
func (a *App) resultsTableHeadersFiltered() []string {
	profile := a.currentLayoutProfile()
	headers := make([]string, 0, len(resultsColumnDefs))
	for i, col := range resultsColumnDefs {
		if !a.resultsVisibleCols[i] {
			continue
		}
		if profile == "compact" {
			headers = append(headers, col.TitleShort)
		} else {
			headers = append(headers, col.Title)
		}
	}
	return headers
}

// resultsTableColumnWidthsFiltered возвращает ширины только видимых колонок.
func (a *App) resultsTableColumnWidthsFiltered() []float32 {
	profile := a.currentLayoutProfile()
	widths := make([]float32, 0, len(resultsColumnDefs))
	for i, col := range resultsColumnDefs {
		if !a.resultsVisibleCols[i] {
			continue
		}
		switch profile {
		case "compact":
			widths = append(widths, col.Widths[0])
		case "wide":
			widths = append(widths, col.Widths[2])
		default:
			widths = append(widths, col.Widths[1])
		}
	}
	return widths
}
