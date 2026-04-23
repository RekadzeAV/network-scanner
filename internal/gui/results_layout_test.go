package gui

import "testing"

func TestDetectLayoutProfile(t *testing.T) {
	a := &App{}
	tests := []struct {
		width float32
		want  string
	}{
		{width: 1200, want: "compact"},
		{width: 1366, want: "compact"},
		{width: 1367, want: "normal"},
		{width: 1920, want: "normal"},
		{width: 2200, want: "wide"},
	}
	for _, tt := range tests {
		if got := a.detectLayoutProfile(tt.width); got != tt.want {
			t.Fatalf("detectLayoutProfile(%v) = %q, want %q", tt.width, got, tt.want)
		}
	}
}

func TestResultsTableColumnsAndHeadersByProfile(t *testing.T) {
	a := &App{layoutProfile: "compact"}
	compactHeaders := a.resultsTableHeaders()
	if len(compactHeaders) != 8 {
		t.Fatalf("compact headers len = %d, want 8", len(compactHeaders))
	}
	compactWidths := a.resultsTableColumnWidths()
	if len(compactWidths) != 8 {
		t.Fatalf("compact widths len = %d, want 8", len(compactWidths))
	}

	a.layoutProfile = "wide"
	wideHeaders := a.resultsTableHeaders()
	if len(wideHeaders) != 8 {
		t.Fatalf("wide headers len = %d, want 8", len(wideHeaders))
	}
	wideWidths := a.resultsTableColumnWidths()
	if len(wideWidths) != 8 {
		t.Fatalf("wide widths len = %d, want 8", len(wideWidths))
	}

	a.layoutProfile = "normal"
	normalWidths := a.resultsTableColumnWidths()
	if len(normalWidths) != 8 {
		t.Fatalf("normal widths len = %d, want 8", len(normalWidths))
	}

	// Проверяем, что compact реально уже, чем wide для критичных колонок.
	if compactWidths[7] >= wideWidths[7] {
		t.Fatalf("ports column width compact=%v should be less than wide=%v", compactWidths[7], wideWidths[7])
	}
}
