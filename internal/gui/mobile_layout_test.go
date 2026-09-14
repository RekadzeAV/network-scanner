package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// --- mobile_layout.go tests ---

func TestNewMobileLayout_Created(t *testing.T) {
	tabs := container.NewAppTabs()
	ml := NewMobileLayout(tabs)
	if ml == nil {
		t.Fatal("expected non-nil MobileLayout")
	}
	if ml.mainTabs != tabs {
		t.Error("expected mainTabs to be set")
	}
	if ml.isMobile {
		t.Error("expected isMobile=false by default")
	}
	if ml.smallScreen {
		t.Error("expected smallScreen=false by default")
	}
	if ml.currentOrientation != "portrait" {
		t.Errorf("expected orientation='portrait', got %q", ml.currentOrientation)
	}
}

func TestNewMobileLayout_NilTabs(t *testing.T) {
	ml := NewMobileLayout(nil)
	if ml == nil {
		t.Fatal("expected non-nil MobileLayout for nil tabs")
	}
	if ml.mainTabs != nil {
		t.Error("expected mainTabs to be nil")
	}
}

func TestMobileLayout_Update_Portrait(t *testing.T) {
	ml := &MobileLayout{}
	mockCanvas := &mockCanvas{w: 360, h: 640}
	ml.Update(mockCanvas)
	if ml.currentOrientation != "portrait" {
		t.Errorf("expected 'portrait', got %q", ml.currentOrientation)
	}
	if !ml.smallScreen {
		t.Error("expected smallScreen=true for 360x640")
	}
}

func TestMobileLayout_Update_Landscape(t *testing.T) {
	ml := &MobileLayout{}
	mockCanvas := &mockCanvas{w: 800, h: 480}
	ml.Update(mockCanvas)
	if ml.currentOrientation != "landscape" {
		t.Errorf("expected 'landscape', got %q", ml.currentOrientation)
	}
	if !ml.smallScreen {
		t.Error("expected smallScreen=true for 800x480")
	}
}

func TestMobileLayout_Update_Desktop(t *testing.T) {
	ml := &MobileLayout{}
	mockCanvas := &mockCanvas{w: 1920, h: 1080}
	ml.Update(mockCanvas)
	if ml.smallScreen {
		t.Error("expected smallScreen=false for 1920x1080")
	}
	if ml.isMobile {
		t.Error("expected isMobile=false for desktop")
	}
}

func TestMobileLayout_Update_Borderline(t *testing.T) {
	ml := &MobileLayout{}
	// 600x800 — граница smallScreen
	mockCanvas := &mockCanvas{w: 600, h: 800}
	ml.Update(mockCanvas)
	if ml.smallScreen {
		t.Error("expected smallScreen=false for 600x800 (boundary)")
	}
}

func TestMobileLayout_GetLayoutMode_Desktop(t *testing.T) {
	ml := &MobileLayout{isMobile: false, smallScreen: false}
	mode := ml.GetLayoutMode()
	if mode != "desktop" {
		t.Errorf("expected 'desktop', got %q", mode)
	}
}

func TestMobileLayout_GetLayoutMode_Mobile(t *testing.T) {
	ml := &MobileLayout{isMobile: true, smallScreen: false}
	mode := ml.GetLayoutMode()
	if mode != "mobile" {
		t.Errorf("expected 'mobile', got %q", mode)
	}
}

func TestMobileLayout_GetLayoutMode_Compact(t *testing.T) {
	ml := &MobileLayout{isMobile: true, smallScreen: true}
	mode := ml.GetLayoutMode()
	if mode != "compact" {
		t.Errorf("expected 'compact', got %q", mode)
	}
}

func TestMobileLayout_ApplyMobileLayout_NilTabs(t *testing.T) {
	ml := &MobileLayout{mainTabs: nil}
	// Не должен паниковать
	ml.applyMobileLayout()
}

func TestMobileLayout_Update_PhonePortrait(t *testing.T) {
	ml := &MobileLayout{}
	mockCanvas := &mockCanvas{w: 375, h: 812}
	ml.Update(mockCanvas)
	if ml.currentOrientation != "portrait" {
		t.Errorf("expected 'portrait', got %q", ml.currentOrientation)
	}
	if !ml.smallScreen {
		t.Error("expected smallScreen=true for 375x812")
	}
}

func TestMobileLayout_Update_TabletLandscape(t *testing.T) {
	ml := &MobileLayout{}
	mockCanvas := &mockCanvas{w: 1024, h: 768}
	ml.Update(mockCanvas)
	if ml.currentOrientation != "landscape" {
		t.Errorf("expected 'landscape', got %q", ml.currentOrientation)
	}
	if !ml.smallScreen {
		t.Error("expected smallScreen=true for 1024x768")
	}
}

func TestCreateMobileTabBar_Empty(t *testing.T) {
	result := CreateMobileTabBar(nil)
	if result != nil {
		t.Error("expected nil for empty tabs")
	}
}

func TestCreateMobileTabBar_EmptySlice(t *testing.T) {
	result := CreateMobileTabBar([]*container.TabItem{})
	if result != nil {
		t.Error("expected nil for empty slice")
	}
}

func TestCreateMobileTabBar_SingleTab(t *testing.T) {
	tabs := []*container.TabItem{
		{Content: canvas.NewText("Test", nil)},
	}
	result := CreateMobileTabBar(tabs)
	if result == nil {
		t.Fatal("expected non-nil container for single tab")
	}
}

func TestCreateMobileTabBar_MultipleTabs(t *testing.T) {
	tabs := []*container.TabItem{
		{Content: canvas.NewText("Test", nil)},
		{Content: canvas.NewText("Test", nil)},
		{Content: canvas.NewText("Test", nil)},
	}
	result := CreateMobileTabBar(tabs)
	if result == nil {
		t.Fatal("expected non-nil container for multiple tabs")
	}
}

// mockCanvas — заглушка для fyne.Canvas
type mockCanvas struct {
	fyne.Canvas
	w, h float32
}

func (m *mockCanvas) Size() fyne.Size {
	return fyne.NewSize(m.w, m.h)
}

func (m *mockCanvas) Scale() float32 {
	return 1.0
}
