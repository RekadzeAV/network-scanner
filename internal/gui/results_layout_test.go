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

func TestLayoutAdaptiveMultiplierBounds(t *testing.T) {
	m := layoutAdaptiveMultiplier(1280, 720, 1)
	if m < 0.72 || m > 1.38 {
		t.Fatalf("multiplier out of bounds: %v", m)
	}
	small := layoutAdaptiveMultiplier(800, 600, 1)
	large := layoutAdaptiveMultiplier(2400, 1350, 1)
	if large < small {
		t.Fatalf("expected larger logical canvas to yield >= multiplier: small=%v large=%v", small, large)
	}
	// Более высокий scale (типичный HiDPI) не должен раздувать множитель выше верхней границы.
	hi := layoutAdaptiveMultiplier(1280, 720, 2)
	if hi > 1.38 {
		t.Fatalf("HiDPI scale should clamp: got %v", hi)
	}
}

func TestSuggestedScanTabOffsetInRange(t *testing.T) {
	m := layoutAdaptiveMultiplier(1440, 900, 1.25)
	off := suggestedScanTabOffset("normal", 1440, 900, 1.25, m)
	if off < 0.26 || off > 0.54 {
		t.Fatalf("offset out of range: %v", off)
	}
}

func TestDefaultTopologySplitOffsetByProfile(t *testing.T) {
	if defaultTopologySplitOffset("compact") <= defaultTopologySplitOffset("wide") {
		t.Fatalf("compact topology split should be > wide (more preview): compact=%v wide=%v",
			defaultTopologySplitOffset("compact"), defaultTopologySplitOffset("wide"))
	}
}

func TestClampFloat64(t *testing.T) {
	if got := clampFloat64(1.5, 0, 1); got != 1 {
		t.Fatalf("clamp high: got %v", got)
	}
	if got := clampFloat64(-0.5, 0, 1); got != 0 {
		t.Fatalf("clamp low: got %v", got)
	}
}

func TestDefaultToolsSplitOffsetByProfile(t *testing.T) {
	c, w, n := defaultToolsSplitOffset("compact"), defaultToolsSplitOffset("wide"), defaultToolsSplitOffset("normal")
	if c < 0.2 || c > 0.6 || w < 0.2 || w > 0.6 || n < 0.2 || n > 0.6 {
		t.Fatalf("unexpected defaults: compact=%v wide=%v normal=%v", c, w, n)
	}
	if w >= n || n >= c {
		t.Fatalf("expected wide <= normal <= compact (top fraction): %v %v %v", w, n, c)
	}
}
