package gui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

func TestBuildCardsViewLimitsVisibleItems(t *testing.T) {
	a := &App{cardsVisibleCount: 2}
	data := []scanner.Result{
		{Hostname: "h1", IP: "10.0.0.1"},
		{Hostname: "h2", IP: "10.0.0.2"},
		{Hostname: "h3", IP: "10.0.0.3"},
	}

	view := a.buildCardsView(data)
	border, ok := view.(*fyne.Container)
	if !ok {
		t.Fatalf("expected container view, got %T", view)
	}
	list, ok := border.Objects[0].(*widget.List)
	if !ok {
		t.Fatalf("expected virtualized list, got %T", border.Objects[0])
	}
	if got := list.Length(); got != 2 {
		t.Fatalf("expected 2 visible cards, got %d", got)
	}
	loadBtn, ok := border.Objects[1].(*widget.Button)
	if !ok {
		t.Fatalf("expected load-more button in bottom slot, got %T", border.Objects[1])
	}
	if txt := strings.TrimSpace(loadBtn.Text); txt == "" {
		t.Fatalf("load-more button text must not be empty")
	}
}

func TestUpdateResultsPerfLabel(t *testing.T) {
	a := &App{resultsPerfLabel: widget.NewLabel("")}
	a.updateResultsPerfLabel(resultsRenderStats{
		FilteredCount: 1200,
		VisibleCount:  200,
		Duration:      85 * time.Millisecond,
	})
	if got := strings.TrimSpace(a.resultsPerfLabel.Text); got != "Рендер: 85ms (200/1200)" {
		t.Fatalf("unexpected perf label text: %q", got)
	}
}

func TestHostDetailsMarkdownCachesByIP(t *testing.T) {
	a := &App{hostDetailsCache: make(map[string]string)}
	r := scanner.Result{Hostname: "h1", IP: "10.0.0.1", DeviceType: "Router"}
	first := a.hostDetailsMarkdown(r)
	second := a.hostDetailsMarkdown(r)
	if first != second {
		t.Fatalf("cached markdown mismatch")
	}
	if len(a.hostDetailsCache) != 1 {
		t.Fatalf("expected one cache entry, got %d", len(a.hostDetailsCache))
	}
}

func TestSelectedTypeFiltersStableOrder(t *testing.T) {
	a := &App{
		quickTypeChecks: map[string]*widget.Check{
			"Server":         widget.NewCheck("Server", nil),
			"Network Device": widget.NewCheck("Network Device", nil),
			"Computer":       widget.NewCheck("Computer", nil),
		},
	}
	a.quickTypeChecks["Server"].SetChecked(true)
	a.quickTypeChecks["Network Device"].SetChecked(true)
	a.quickTypeChecks["Computer"].SetChecked(true)
	got := a.selectedTypeFilters()
	want := []string{"Computer", "Network Device", "Server"}
	if len(got) != len(want) {
		t.Fatalf("unexpected filters len: %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("selectedTypeFilters order mismatch: got=%v want=%v", got, want)
		}
	}
}

