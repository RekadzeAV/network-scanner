package gui

import (
	"testing"

	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

// --- R-03: сортировка по клику на заголовок + направление ---

func TestSortIndicatorForColumn(t *testing.T) {
	a := &App{resultsSort: resultsSortByIP, resultsSortDesc: false}
	// IP — колонка 1: по возрастанию → ▲.
	if got := a.sortIndicatorForColumn(1); got != "▲" {
		t.Errorf("IP asc indicator = %q, want ▲", got)
	}
	a.resultsSortDesc = true
	if got := a.sortIndicatorForColumn(1); got != "▼" {
		t.Errorf("IP desc indicator = %q, want ▼", got)
	}
	// Колонка 0 (Hostname) не активна, пока сортировка по IP.
	if got := a.sortIndicatorForColumn(0); got != "" {
		t.Errorf("inactive column indicator = %q, want empty", got)
	}
	// Несортируемая колонка 2 (MAC) всегда без индикатора.
	if got := a.sortIndicatorForColumn(2); got != "" {
		t.Errorf("non-sortable column indicator = %q, want empty", got)
	}
}

func TestToggleSortByColumn(t *testing.T) {
	a := &App{resultsSort: resultsSortByIP, resultsSortDesc: false}

	// Клик по «Имя хоста» (колонка 0) переключает режим и сбрасывает направление.
	a.toggleSortByColumn(0)
	if a.resultsSort != resultsSortByHostname {
		t.Fatalf("resultsSort = %q, want %q", a.resultsSort, resultsSortByHostname)
	}
	if a.resultsSortDesc {
		t.Fatal("new column must start ascending")
	}

	// Повторный клик инвертирует направление.
	a.toggleSortByColumn(0)
	if !a.resultsSortDesc {
		t.Fatal("second click must toggle direction to descending")
	}

	// Клик по несортируемой колонке ничего не меняет.
	a.toggleSortByColumn(4)
	if a.resultsSort != resultsSortByHostname || !a.resultsSortDesc {
		t.Fatal("non-sortable column must not change state")
	}
}

func TestSortedResultsForDisplayWithModeDir_Descending(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.10"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.1"},
	}
	asc := sortedResultsForDisplayWithModeDir(results, resultsSortByIP, false)
	if asc[0].IP != "192.168.1.1" {
		t.Errorf("asc first = %q, want 192.168.1.1", asc[0].IP)
	}
	desc := sortedResultsForDisplayWithModeDir(results, resultsSortByIP, true)
	if desc[0].IP != "192.168.1.10" {
		t.Errorf("desc first = %q, want 192.168.1.10", desc[0].IP)
	}
}

func TestSortedResultsForDisplayWithModeDir_HostnameDescending(t *testing.T) {
	results := []scanner.Result{
		{Hostname: "alpha", IP: "10.0.0.1"},
		{Hostname: "charlie", IP: "10.0.0.3"},
		{Hostname: "bravo", IP: "10.0.0.2"},
	}
	desc := sortedResultsForDisplayWithModeDir(results, resultsSortByHostname, true)
	if desc[0].Hostname != "charlie" || desc[2].Hostname != "alpha" {
		t.Errorf("hostname desc order = %v", []string{desc[0].Hostname, desc[1].Hostname, desc[2].Hostname})
	}
}

// --- R-09: счётчик строк ---

func TestUpdateResultsCountLabel(t *testing.T) {
	a := &App{resultsCountLabel: widget.NewLabel("")}
	a.updateResultsCountLabel(10, 10)
	if a.resultsCountLabel.Text != "Строк: 10" {
		t.Errorf("equal counts = %q, want «Строк: 10»", a.resultsCountLabel.Text)
	}
	a.updateResultsCountLabel(5, 10)
	if a.resultsCountLabel.Text != "Строк: показано 5 из 10" {
		t.Errorf("partial counts = %q, want «Строк: показано 5 из 10»", a.resultsCountLabel.Text)
	}
}

func TestUpdateResultsCountLabel_NilSafe(t *testing.T) {
	var a *App
	a.updateResultsCountLabel(1, 1) // не должно паниковать
	(&App{}).updateResultsCountLabel(1, 1)
}

// --- R-08: визуальные категории ---

func TestDeviceCategoryGlyph(t *testing.T) {
	cases := map[string]string{
		"Router":       "⬡",
		"Switch":       "⬡",
		"Access Point": "◎",
		"Printer":      "⎙",
		"Camera":       "◉",
		"NAS":          "▤",
		"IoT Device":   "◈",
		"Desktop":      "▢",
		"Laptop":       "▢",
		"Server":       "▣",
		"Phone":        "▮",
		"Tablet":       "▮",
		"Something":    "◌",
	}
	for in, want := range cases {
		if got := deviceCategoryGlyph(in); got != want {
			t.Errorf("deviceCategoryGlyph(%q) = %q, want %q", in, got, want)
		}
	}
}
