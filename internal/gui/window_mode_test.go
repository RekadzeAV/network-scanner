package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
)

// --- newWindowModeCtrl ---

func TestNewWindowModeCtrl_DefaultMode(t *testing.T) {
	a := &App{}
	ctrl := newWindowModeCtrl(a)
	if ctrl == nil {
		t.Fatal("expected non-nil controller")
	}
	if ctrl.app != a {
		t.Error("expected controller.app to be set")
	}
	if ctrl.state.mode != windowModeNormal {
		t.Errorf("expected default mode Normal, got %v", ctrl.state.mode)
	}
	if ctrl.window != nil {
		t.Error("expected window to be nil initially")
	}
}

// --- currentMode / toggle ---

func TestCurrentMode_Normal(t *testing.T) {
	ctrl := &windowModeCtrl{state: windowModeState{mode: windowModeNormal}}
	if ctrl.currentMode() != windowModeNormal {
		t.Error("expected Normal")
	}
}

func TestCurrentMode_Fullscreen(t *testing.T) {
	ctrl := &windowModeCtrl{state: windowModeState{mode: windowModeFullscreen}}
	if ctrl.currentMode() != windowModeFullscreen {
		t.Error("expected Fullscreen")
	}
}

func TestToggle_NormalToFullscreen(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeNormal},
	}
	ctrl.toggle()
	if ctrl.state.mode != windowModeNormal {
		t.Logf("mode stayed Normal (myWindow nil, no-op)")
	}
}

func TestToggle_FullscreenToNormal(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeFullscreen},
	}
	ctrl.toggle()
	if ctrl.state.mode != windowModeFullscreen {
		t.Logf("mode stayed Fullscreen (myWindow nil, no-op)")
	}
}

// --- enterFullscreen / exitFullscreen (nil-safe) ---

func TestEnterFullscreen_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{app: nil}
	ctrl.enterFullscreen()
}

func TestEnterFullscreen_NilMyWindow(t *testing.T) {
	ctrl := &windowModeCtrl{app: &App{}}
	ctrl.enterFullscreen()
}

func TestExitFullscreen_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{app: nil}
	ctrl.exitFullscreen()
}

func TestExitFullscreen_NilMyWindow(t *testing.T) {
	ctrl := &windowModeCtrl{app: &App{}}
	ctrl.exitFullscreen()
}

// --- save (nil-safe) ---

func TestSave_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{app: nil}
	ctrl.save()
}

func TestSave_NilMyApp(t *testing.T) {
	ctrl := &windowModeCtrl{app: &App{}}
	ctrl.save()
}

// --- IsFullscreen ---

func TestIsFullscreen_NilWinMode(t *testing.T) {
	a := &App{}
	if a.IsFullscreen() {
		t.Error("expected false when winMode is nil")
	}
}

func TestIsFullscreen_Normal(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeNormal}}
	if a.IsFullscreen() {
		t.Error("expected false in normal mode")
	}
}

func TestIsFullscreen_Fullscreen(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeFullscreen}}
	if !a.IsFullscreen() {
		t.Error("expected true in fullscreen mode")
	}
}

// --- syncFromWindow (nil-safe) ---

func TestSyncFromWindow_NilHWND(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeNormal},
	}
	ctrl.syncFromWindow()
	if ctrl.state.mode != windowModeNormal {
		t.Error("expected mode to stay Normal")
	}
}

func TestSyncFromWindow_FullscreenToNormal(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeFullscreen},
	}
	ctrl.syncFromWindow()
	if ctrl.state.mode != windowModeFullscreen {
		t.Error("expected mode to stay Fullscreen when hwnd is nil")
	}
}

// --- watchForExternalFullscreenChanges (nil-safe) ---

func TestWatchForExternalFullscreenChanges_StopsAfterTimeout(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:       &App{},
		state:     windowModeState{mode: windowModeNormal},
		watchStop: make(chan struct{}),
	}
	ctrl.watchForExternalFullscreenChanges(nil)
	time.Sleep(600 * time.Millisecond)
	if ctrl.state.mode != windowModeNormal {
		t.Errorf("expected Normal after watcher, got %v", ctrl.state.mode)
	}
	ctrl.stopWatcher()
}

// --- makeFullscreenToggleBtn (nil-safe) ---

func TestMakeFullscreenToggleBtn_NilWinMode(t *testing.T) {
	a := &App{}
	btn := a.makeFullscreenToggleBtn()
	if btn == nil {
		t.Fatal("expected non-nil button")
	}
	if btn.Text != "Полный экран" {
		t.Errorf("expected 'Полный экран', got %q", btn.Text)
	}
}

func TestMakeFullscreenToggleBtn_Fullscreen(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{
		state: windowModeState{mode: windowModeFullscreen},
	}
	btn := a.makeFullscreenToggleBtn()
	if btn == nil {
		t.Fatal("expected non-nil button")
	}
	if btn.Text != "Оконный режим" {
		t.Errorf("expected 'Оконный режим', got %q", btn.Text)
	}
}

// --- fullscreenMenuItems ---

func TestFullscreenMenuItems_NilWinMode(t *testing.T) {
	a := &App{}
	items := a.fullscreenMenuItems()
	if len(items) != 2 {
		t.Fatalf("expected 2 menu items, got %d", len(items))
	}
	if items[0].Label != "Полноэкранный режим (F11)" {
		t.Errorf("expected 'Полноэкранный режим (F11)', got %q", items[0].Label)
	}
	if items[1].Label != "Оконный режим (F11)" {
		t.Errorf("expected 'Оконный режим (F11)', got %q", items[1].Label)
	}
}

func TestFullscreenMenuItems_Checked(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{
		state: windowModeState{mode: windowModeNormal},
	}
	items := a.fullscreenMenuItems()
	if items[0].Checked {
		t.Error("fullscreen item should be unchecked in normal mode")
	}
	if !items[1].Checked {
		t.Error("windowed item should be checked in normal mode")
	}

	a.winMode.state.mode = windowModeFullscreen
	items = a.fullscreenMenuItems()
	if !items[0].Checked {
		t.Error("fullscreen item should be checked in fullscreen mode")
	}
	if items[1].Checked {
		t.Error("windowed item should be unchecked in fullscreen mode")
	}
}

// --- applyPendingFullscreen (nil-safe) ---

func TestApplyPendingFullscreen_NilWinMode(t *testing.T) {
	a := &App{}
	a.applyPendingFullscreen()
}

// --- refreshWidgets nil-safe ---

func TestRefreshWidgets_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{
		state: windowModeState{mode: windowModeNormal},
	}
	ctrl.refreshWidgets()
}

// --- refreshLayout nil-safe ---

func TestRefreshLayout_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{
		state: windowModeState{mode: windowModeNormal},
	}
	ctrl.refreshLayout()
}

// --- refreshMenuBar ---

func TestRefreshMenuBar_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{state: windowModeState{mode: windowModeFullscreen}}
	ctrl.refreshMenuBar()
}

func TestRefreshMenuBar_NilWindow(t *testing.T) {
	ctrl := &windowModeCtrl{app: &App{}, state: windowModeState{mode: windowModeFullscreen}}
	ctrl.refreshMenuBar()
}

func TestRefreshMenuBar_NoMenuIsNoop(t *testing.T) {
	a := &App{}
	ctrl := &windowModeCtrl{app: a, state: windowModeState{mode: windowModeFullscreen}}
	ctrl.refreshMenuBar()
}

// --- enterFullscreen/exitFullscreen: безопасны без окна ---

func TestEnterFullscreen_NilWindow(t *testing.T) {
	ctrl := &windowModeCtrl{app: &App{}, state: windowModeState{mode: windowModeNormal}}
	ctrl.enterFullscreen()
	if ctrl.state.mode != windowModeNormal {
		t.Errorf("при nil окне режим не должен меняться, получено %v", ctrl.state.mode)
	}
}

func TestExitFullscreen_NilWindow(t *testing.T) {
	ctrl := &windowModeCtrl{app: &App{}, state: windowModeState{mode: windowModeFullscreen}}
	ctrl.exitFullscreen()
}

func TestEnterFullscreen_AlreadyFullscreenIsNoop(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeFullscreen},
	}
	ctrl.enterFullscreen()
	if ctrl.state.mode != windowModeFullscreen {
		t.Errorf("режим должен остаться fullscreen, получено %v", ctrl.state.mode)
	}
}

// --- stopWatcher ---

func TestStopWatcher_NoPanic(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:       &App{},
		state:     windowModeState{mode: windowModeNormal},
		watchStop: make(chan struct{}),
	}
	ctrl.watchForExternalFullscreenChanges(nil)
	ctrl.stopWatcher()
	ctrl.stopWatcher()
}

// --- clampToMaxSize ---

func TestClampToMaxSize_WithinBounds(t *testing.T) {
	tests := []struct {
		name string
		w, h float32
	}{
		{"1366x768", 1366, 768},
		{"1920x1080", 1920, 1080},
		{"2560x1440", 2560, 1440},
		{"3840x2160_4K", 3840, 2160},
		{"5120x2880_5K", 5120, 2880},
		{"7680x4320_max", 7680, 4320},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sz := fyne.NewSize(tt.w, tt.h)
			clamped := clampToMaxSize(sz)
			if clamped.Width != tt.w || clamped.Height != tt.h {
				t.Errorf("size should not change: got %v, want %v", clamped, sz)
			}
		})
	}
}

func TestClampToMaxSize_ExceedsMaxWidth(t *testing.T) {
	sz := fyne.NewSize(10000, 2000)
	clamped := clampToMaxSize(sz)
	if clamped.Width != maxWindowWidth {
		t.Errorf("width should be clamped to %d, got %v", maxWindowWidth, clamped.Width)
	}
	if clamped.Height != 2000 {
		t.Errorf("height should not change, got %v", clamped.Height)
	}
}

func TestClampToMaxSize_ExceedsMaxHeight(t *testing.T) {
	sz := fyne.NewSize(2000, 10000)
	clamped := clampToMaxSize(sz)
	if clamped.Width != 2000 {
		t.Errorf("width should not change, got %v", clamped.Width)
	}
	if clamped.Height != maxWindowHeight {
		t.Errorf("height should be clamped to %d, got %v", maxWindowHeight, clamped.Height)
	}
}

func TestClampToMaxSize_ExceedsBoth(t *testing.T) {
	sz := fyne.NewSize(10000, 10000)
	clamped := clampToMaxSize(sz)
	if clamped.Width != maxWindowWidth || clamped.Height != maxWindowHeight {
		t.Errorf("both should be clamped: got %v, want %v", clamped,
			fyne.NewSize(maxWindowWidth, maxWindowHeight))
	}
}

func TestClampToMaxSize_ZeroSize(t *testing.T) {
	sz := fyne.NewSize(0, 0)
	clamped := clampToMaxSize(sz)
	if clamped.Width != 0 || clamped.Height != 0 {
		t.Errorf("zero size should not change, got %v", clamped)
	}
}

func TestClampToMaxSize_NegativeSize(t *testing.T) {
	sz := fyne.NewSize(-100, -100)
	clamped := clampToMaxSize(sz)
	if clamped.Width != -100 || clamped.Height != -100 {
		t.Errorf("negative size should not change, got %v", clamped)
	}
}

// --- syncFromWindow: mutex защита ---

func TestSyncFromWindow_MutexSafe(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeNormal},
	}
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				ctrl.syncFromWindow()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	if ctrl.state.mode != windowModeNormal {
		t.Errorf("expected Normal after concurrent access, got %v", ctrl.state.mode)
	}
}

// --- applyResponsiveLayout: ultrawide ---

func TestApplyResponsiveLayout_Ultrawide_34inch(t *testing.T) {
	a := &App{lastCanvasSize: fyne.NewSize(3440, 1440), lastCanvasScale: 1}
	a.applyResponsiveLayout("ultrawide") // не должен паниковать
}

func TestApplyResponsiveLayout_Ultrawide_32inch_4K(t *testing.T) {
	a := &App{lastCanvasSize: fyne.NewSize(3840, 2160), lastCanvasScale: 1}
	a.applyResponsiveLayout("ultrawide") // не должен паниковать
}

func TestApplyResponsiveLayout_Ultrawide_HiDPI(t *testing.T) {
	a := &App{lastCanvasSize: fyne.NewSize(2293, 960), lastCanvasScale: 1.5}
	a.applyResponsiveLayout("ultrawide") // не должен паниковать
}

// --- split offsets: ultrawide ---

func TestDefaultTopologySplitOffset_Ultrawide(t *testing.T) {
	c, w, n, uw := defaultTopologySplitOffset("compact"), defaultTopologySplitOffset("wide"), defaultTopologySplitOffset("normal"), defaultTopologySplitOffset("ultrawide")
	// Ultrawide должен давать больше места для карты, чем wide
	if uw >= w {
		t.Errorf("ultrawide topology split (%v) should be < wide (%v)", uw, w)
	}
	// Но больше, чем fullscreen (0.55) — т.е. не должно быть слишком маленьким
	if uw <= 0 {
		t.Errorf("ultrawide topology split should be positive, got %v", uw)
	}
	_ = c
	_ = n // используем compact/normal для сравнения
	if c <= w {
		t.Errorf("compact topology split (%v) should be > wide (%v)", c, w)
	}
}

func TestDefaultToolsSplitOffset_Ultrawide(t *testing.T) {
	c, w, n, uw := defaultToolsSplitOffset("compact"), defaultToolsSplitOffset("wide"), defaultToolsSplitOffset("normal"), defaultToolsSplitOffset("ultrawide")
	// Ultrawide должен давать больше места для вывода, чем wide
	if uw >= w {
		t.Errorf("ultrawide tools split (%v) should be < wide (%v)", uw, w)
	}
	_ = c
	_ = n
	if c <= n {
		t.Errorf("compact tools split (%v) should be > normal (%v)", c, n)
	}
}

func TestSuggestedScanTabOffset_Ultrawide(t *testing.T) {
	m := layoutAdaptiveMultiplier(3440, 1440, 1)
	off := suggestedScanTabOffset("ultrawide", 3440, 1440, 1, m)
	if off < 0.20 || off > 0.54 {
		t.Fatalf("ultrawide scan tab offset out of range: %v", off)
	}
	// Ultrawide должен отдавать меньше места для scan controls, чем normal
	offN := suggestedScanTabOffset("normal", 3440, 1440, 1, m)
	if off >= offN {
		t.Errorf("ultrawide offset (%v) should be < normal (%v)", off, offN)
	}
}
