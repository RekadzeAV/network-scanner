package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func TestApplyResponsiveLayoutCompactReflowsGrids(t *testing.T) {
	a := &App{
		resultsDiagnosticsGrid: container.New(layout.NewGridLayoutWithColumns(3), widget.NewLabel("x")),
		resultsSortGrid:        container.New(layout.NewGridLayoutWithColumns(5), widget.NewLabel("x")),
		resultsCidrGrid:        container.New(layout.NewGridLayoutWithColumns(4), widget.NewLabel("x")),
		resultsPresetGrid:      container.New(layout.NewGridLayoutWithColumns(4), widget.NewLabel("x")),
		toolButtonsGrid:        container.New(layout.NewGridLayoutWithColumns(5), widget.NewButton("b", nil)),
		operationsHeaderGrid:   container.New(layout.NewGridLayoutWithColumns(2), widget.NewLabel("x")),
	}

	a.applyResponsiveLayout("compact")

	assertLayoutChanged(t, a.resultsDiagnosticsGrid)
	assertLayoutChanged(t, a.resultsSortGrid)
	assertLayoutChanged(t, a.resultsCidrGrid)
	assertLayoutChanged(t, a.resultsPresetGrid)
	assertLayoutChanged(t, a.toolButtonsGrid)
	assertLayoutChanged(t, a.operationsHeaderGrid)
}

func TestApplyResponsiveLayoutWideRestoresGrids(t *testing.T) {
	a := &App{
		resultsDiagnosticsGrid: container.New(layout.NewGridLayoutWithColumns(1), widget.NewLabel("x")),
		resultsSortGrid:        container.New(layout.NewGridLayoutWithColumns(2), widget.NewLabel("x")),
		resultsCidrGrid:        container.New(layout.NewGridLayoutWithColumns(2), widget.NewLabel("x")),
		resultsPresetGrid:      container.New(layout.NewGridLayoutWithColumns(2), widget.NewLabel("x")),
		toolButtonsGrid:        container.New(layout.NewGridLayoutWithColumns(2), widget.NewButton("b", nil)),
		operationsHeaderGrid:   container.New(layout.NewGridLayoutWithColumns(1), widget.NewLabel("x")),
	}

	beforeDiag := a.resultsDiagnosticsGrid.Layout
	beforeSort := a.resultsSortGrid.Layout
	beforeCidr := a.resultsCidrGrid.Layout
	beforePreset := a.resultsPresetGrid.Layout
	beforeButtons := a.toolButtonsGrid.Layout
	beforeOps := a.operationsHeaderGrid.Layout

	a.applyResponsiveLayout("wide")

	assertLayoutReplaced(t, beforeDiag, a.resultsDiagnosticsGrid.Layout)
	assertLayoutReplaced(t, beforeSort, a.resultsSortGrid.Layout)
	assertLayoutReplaced(t, beforeCidr, a.resultsCidrGrid.Layout)
	assertLayoutReplaced(t, beforePreset, a.resultsPresetGrid.Layout)
	assertLayoutReplaced(t, beforeButtons, a.toolButtonsGrid.Layout)
	assertLayoutReplaced(t, beforeOps, a.operationsHeaderGrid.Layout)
}

func assertLayoutChanged(t *testing.T, c *fyne.Container) {
	t.Helper()
	if c == nil {
		t.Fatalf("container is nil")
	}
	if c.Layout == nil {
		t.Fatalf("layout is nil")
	}
}

func assertLayoutReplaced(t *testing.T, before fyne.Layout, after fyne.Layout) {
	t.Helper()
	if after == nil {
		t.Fatalf("layout is nil")
	}
	if before == after {
		t.Fatalf("layout instance was not replaced")
	}
}
