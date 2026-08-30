package gui

import (
	"testing"

	"fyne.io/fyne/v2/widget"
)

// --- inventory_view.go tests ---

func TestBuildInventoryDashboardView_NilApp(t *testing.T) {
	var a *App
	result := a.buildInventoryDashboardView()
	if result == nil {
		t.Fatal("expected non-nil result for nil app")
	}
}

func TestBuildInventoryDashboardView_EmptyApp(t *testing.T) {
	a := &App{}
	result := a.buildInventoryDashboardView()
	if result == nil {
		t.Fatal("expected non-nil result for empty app")
	}
}

func TestBuildInventoryDashboardView_WithControls(t *testing.T) {
	a := &App{}
	a.inventoryStatusLabel = widget.NewLabel("")
	a.inventoryScanASelect = widget.NewSelect([]string{}, nil)
	a.inventoryScanBSelect = widget.NewSelect([]string{}, nil)
	result := a.buildInventoryDashboardView()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestRefreshInventorySnapshots_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.refreshInventorySnapshots()
}

func TestRefreshInventorySnapshots_NilDBEntry(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.refreshInventorySnapshots()
}

func TestInventoryDiffMarkdown_NilApp(t *testing.T) {
	var a *App
	result := a.inventoryDiffMarkdown()
	if result == "" {
		t.Fatal("expected non-empty result for nil app")
	}
}

func TestInventoryDiffMarkdown_NilControls(t *testing.T) {
	a := &App{}
	result := a.inventoryDiffMarkdown()
	if result == "" {
		t.Fatal("expected non-empty result for nil controls")
	}
}

func TestInventoryDiffMarkdown_EmptySelections(t *testing.T) {
	a := &App{}
	a.inventoryScanASelect = widget.NewSelect([]string{}, nil)
	a.inventoryScanBSelect = widget.NewSelect([]string{}, nil)
	result := a.inventoryDiffMarkdown()
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestSaveInventorySnapshotFromResults_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.saveInventorySnapshotFromResults(nil)
}

func TestSaveInventorySnapshotFromResults_NilAutoSaveCheck(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.saveInventorySnapshotFromResults(nil)
}

func TestSaveInventorySnapshotFromResults_AutoSaveUnchecked(t *testing.T) {
	a := &App{}
	a.inventoryAutoSaveCheck = widget.NewCheck("", nil)
	a.inventoryAutoSaveCheck.SetChecked(false)
	// Не должен паниковать
	a.saveInventorySnapshotFromResults(nil)
}
