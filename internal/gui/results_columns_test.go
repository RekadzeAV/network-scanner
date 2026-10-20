package gui

import (
	"testing"
)

// P2-12: тесты для настройки видимых колонок таблицы результатов.

func TestVisibleColCount_AllVisible(t *testing.T) {
	a := &App{}
	a.initVisibleCols()
	count := a.visibleColCount()
	if count != 8 {
		t.Errorf("expected 8 visible columns, got %d", count)
	}
}

func TestVisibleColCount_SomeHidden(t *testing.T) {
	a := &App{}
	a.initVisibleCols()
	a.resultsVisibleCols[0] = false
	a.resultsVisibleCols[7] = false
	count := a.visibleColCount()
	if count != 6 {
		t.Errorf("expected 6 visible columns, got %d", count)
	}
}

func TestVisibleColCount_MinimumOne(t *testing.T) {
	a := &App{}
	a.initVisibleCols()
	// Скрываем все колонки — должна остаться хотя бы одна.
	for i := range a.resultsVisibleCols {
		a.resultsVisibleCols[i] = false
	}
	count := a.visibleColCount()
	if count < 1 {
		t.Errorf("expected at least 1 visible column, got %d", count)
	}
}

func TestVisibleColIndices(t *testing.T) {
	a := &App{}
	a.initVisibleCols()
	a.resultsVisibleCols[1] = false
	a.resultsVisibleCols[3] = false

	indices := a.visibleColIndices()
	expected := []int{0, 2, 4, 5, 6, 7}

	if len(indices) != len(expected) {
		t.Fatalf("expected %d indices, got %d", len(expected), len(indices))
	}
	for i, idx := range expected {
		if indices[i] != idx {
			t.Errorf("expected index[%d]=%d, got %d", i, idx, indices[i])
		}
	}
}

func TestResultsTableHeadersFiltered(t *testing.T) {
	a := &App{}
	a.initVisibleCols()

	// Все колонки видимы — 8 заголовков.
	headers := a.resultsTableHeadersFiltered()
	if len(headers) != 8 {
		t.Errorf("expected 8 headers, got %d", len(headers))
	}

	// Скрываем колонки 0 и 7.
	a.resultsVisibleCols[0] = false
	a.resultsVisibleCols[7] = false
	headers = a.resultsTableHeadersFiltered()
	if len(headers) != 6 {
		t.Errorf("expected 6 headers, got %d", len(headers))
	}
}

func TestResultsTableColumnWidthsFiltered(t *testing.T) {
	a := &App{}
	a.initVisibleCols()

	// Все колонки видимы — 8 ширин.
	widths := a.resultsTableColumnWidthsFiltered()
	if len(widths) != 8 {
		t.Errorf("expected 8 widths, got %d", len(widths))
	}

	// Скрываем колонку 0.
	a.resultsVisibleCols[0] = false
	widths = a.resultsTableColumnWidthsFiltered()
	if len(widths) != 7 {
		t.Errorf("expected 7 widths, got %d", len(widths))
	}
}

func TestColumnDefs(t *testing.T) {
	if len(resultsColumnDefs) != 8 {
		t.Errorf("expected 8 column definitions, got %d", len(resultsColumnDefs))
	}

	// Проверка ключевых колонок.
	if resultsColumnDefs[0].Title != "Имя хоста" {
		t.Errorf("expected column 0 title 'Имя хоста', got %q", resultsColumnDefs[0].Title)
	}
	if resultsColumnDefs[1].Title != "IP" {
		t.Errorf("expected column 1 title 'IP', got %q", resultsColumnDefs[1].Title)
	}
	if !resultsColumnDefs[1].Monospace {
		t.Errorf("expected column 1 (IP) to be monospace")
	}
	if !resultsColumnDefs[0].Sortable {
		t.Errorf("expected column 0 (Hostname) to be sortable")
	}
}
