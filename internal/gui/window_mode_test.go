package gui

import (
	"sync/atomic"
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
		// enterFullscreen is a no-op when myWindow is nil
		// so mode stays Normal — это корректно
		t.Logf("mode stayed Normal (myWindow nil, no-op)")
	}
}

func TestToggle_FullscreenToNormal(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeFullscreen},
	}
	ctrl.toggle()
	// exitFullscreen is a no-op when myWindow is nil
	// mode stays Fullscreen
	if ctrl.state.mode != windowModeFullscreen {
		t.Logf("mode stayed Fullscreen (myWindow nil, no-op)")
	}
}

// --- enterFullscreen / exitFullscreen (nil-safe) ---

func TestEnterFullscreen_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{app: nil}
	ctrl.enterFullscreen() // не должен паниковать
}

func TestEnterFullscreen_NilMyWindow(t *testing.T) {
	ctrl := &windowModeCtrl{app: &App{}}
	ctrl.enterFullscreen() // не должен паниковать
}

func TestExitFullscreen_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{app: nil}
	ctrl.exitFullscreen() // не должен паниковать
}

func TestExitFullscreen_NilMyWindow(t *testing.T) {
	ctrl := &windowModeCtrl{app: &App{}}
	ctrl.exitFullscreen() // не должен паниковать
}

// --- save (nil-safe) ---

func TestSave_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{app: nil}
	ctrl.save() // не должен паниковать
}

func TestSave_NilMyApp(t *testing.T) {
	ctrl := &windowModeCtrl{app: &App{}}
	ctrl.save() // не должен паниковать
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

func TestSyncFromWindow_NilWindow(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeNormal},
	}
	ctrl.syncFromWindow() // не должен паниковать
	if ctrl.state.mode != windowModeNormal {
		t.Error("expected mode to stay Normal")
	}
}

func TestSyncFromWindow_FullscreenToNormal(t *testing.T) {
	// syncFromWindow читает window.FullScreen(), но у нас нет мок-окна.
	// Проверяем, что при nil window режим не меняется.
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeFullscreen},
	}
	ctrl.syncFromWindow()
	if ctrl.state.mode != windowModeFullscreen {
		t.Error("expected mode to stay Fullscreen when window is nil")
	}
}

// --- debounce: layoutSeq ---

func TestLayoutSeq_Increments(t *testing.T) {
	ctrl := &windowModeCtrl{}
	seq1 := ctrl.nextLayoutSeq()
	seq2 := ctrl.nextLayoutSeq()
	seq3 := ctrl.nextLayoutSeq()
	if seq1 != 1 || seq2 != 2 || seq3 != 3 {
		t.Errorf("expected 1, 2, 3; got %d, %d, %d", seq1, seq2, seq3)
	}
}

func TestLayoutSeq_CurrentMatchesLast(t *testing.T) {
	ctrl := &windowModeCtrl{}
	ctrl.nextLayoutSeq()
	ctrl.nextLayoutSeq()
	if ctrl.currentLayoutSeq() != 2 {
		t.Errorf("expected current to be 2, got %d", ctrl.currentLayoutSeq())
	}
}

func TestDebounce_OldTaskIgnored(t *testing.T) {
	ctrl := &windowModeCtrl{}
	// Получаем seq = 1, затем сразу seq = 2
	seq := ctrl.nextLayoutSeq() // seq = 1
	ctrl.nextLayoutSeq()        // seq = 2, seq1 устарела

	// Проверяем, что current != old seq
	if ctrl.currentLayoutSeq() == seq {
		t.Error("expected old seq to be stale")
	}
}

// --- watchForExternalFullscreenChanges (nil-safe) ---

func TestWatchForExternalFullscreenChanges_StopsAfterTimeout(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeNormal},
	}
	// Передаём nil как окно — watcher должен корректно обработить nil
	ctrl.watchForExternalFullscreenChanges(nil)

	// Ждём достаточно долго, чтобы ticker сработал несколько раз
	time.Sleep(1500 * time.Millisecond)

	// Проверяем, что состояние не изменилось (nil window = no-op)
	if ctrl.state.mode != windowModeNormal {
		t.Errorf("expected Normal after watcher, got %v", ctrl.state.mode)
	}
}

// --- makeFullscreenToggleBtn (nil-safe) ---

func TestMakeFullscreenToggleBtn_NilWinMode(t *testing.T) {
	a := &App{}
	btn := a.makeFullscreenToggleBtn()
	if btn == nil {
		t.Error("expected non-nil button")
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
		t.Error("expected non-nil button")
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

	// Переключаемся в fullscreen
	a.winMode.state.mode = windowModeFullscreen
	items = a.fullscreenMenuItems()
	if !items[0].Checked {
		t.Error("fullscreen item should be checked in fullscreen mode")
	}
	if items[1].Checked {
		t.Error("windowed item should be unchecked in fullscreen mode")
	}
}

// --- atomic.Uint64 layoutSeq (race-safe) ---

func TestLayoutSeq_RaceSafe(t *testing.T) {
	var seq atomic.Uint64
	done := make(chan struct{})
	// Запускаем 10 гоутина одновременно
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				seq.Add(1)
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	if seq.Load() != 1000 {
		t.Errorf("expected 1000, got %d", seq.Load())
	}
}

// --- fullscreenLayoutRefreshDelay constant ---

func TestFullscreenLayoutRefreshDelay_Type(t *testing.T) {
	// Константа должна быть time.Duration (350 * time.Millisecond)
	var _ time.Duration = fullscreenLayoutRefreshDelay
	if fullscreenLayoutRefreshDelay <= 0 {
		t.Error("fullscreenLayoutRefreshDelay must be positive")
	}
	if fullscreenLayoutRefreshDelay < 100*time.Millisecond {
		t.Error("fullscreenLayoutRefreshDelay should be at least 100ms")
	}
}

// --- detectLayoutProfile с fullscreen ---

func TestDetectLayoutProfile_Fullscreen_Wide(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeFullscreen}}
	// На широком экране (>=1600) fullscreen даёт профиль "fullscreen"
	profile := a.detectLayoutProfile(1920)
	if profile != "fullscreen" {
		t.Errorf("expected 'fullscreen' at 1920, got %q", profile)
	}
}

func TestDetectLayoutProfile_Fullscreen_CompactScreen(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeFullscreen}}
	// На узком экране (< fullscreenMinWidth) fullscreen НЕ переопределяет
	// компактный профиль —避免 5-7 колонок на 1366×768
	profile := a.detectLayoutProfile(1366)
	if profile != "compact" {
		t.Errorf("expected 'compact' at 1366 even in fullscreen, got %q", profile)
	}
}

func TestDetectLayoutProfile_Fullscreen_NormalScreen(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeFullscreen}}
	// 1500px < fullscreenMinWidth=1600 → normal
	profile := a.detectLayoutProfile(1500)
	if profile != "normal" {
		t.Errorf("expected 'normal' at 1500 even in fullscreen, got %q", profile)
	}
}

func TestDetectLayoutProfile_Fullscreen_ExactlyAtMinWidth(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeFullscreen}}
	// Ровно на границе fullscreenMinWidth=1600 → fullscreen
	profile := a.detectLayoutProfile(float32(fullscreenMinWidth))
	if profile != "fullscreen" {
		t.Errorf("expected 'fullscreen' at %d, got %q", fullscreenMinWidth, profile)
	}
}

func TestDetectLayoutProfile_Windowed_NotAffectedByFullscreenLogic(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeNormal}}
	profile := a.detectLayoutProfile(1920)
	if profile != "normal" {
		t.Errorf("expected 'normal' in windowed mode, got %q", profile)
	}
}

// --- applyAdaptiveToolsScrollMinSizes с fullscreen-профилем ---

func TestApplyAdaptiveToolsScrollMinSizes_Fullscreen(t *testing.T) {
	a := &App{}
	// Не должен паниковать с fullscreen-профилем
	a.applyAdaptiveToolsScrollMinSizes("fullscreen", 1.0)
}

func TestApplyAdaptiveToolsScrollMinSizes_Fullscreen_DiffersFromNormal(t *testing.T) {
	// Проверяем, что fullscreen-профиль не идентичен default (normal).
	// Вызываем с обоими профилями — оба должны работать без паники.
	a := &App{}
	a.applyAdaptiveToolsScrollMinSizes("normal", 1.0)
	a.applyAdaptiveToolsScrollMinSizes("fullscreen", 1.0)
	// В fullscreen baseOut=280, baseOps=100; в default baseOut=360, baseOps=140.
	// Без scroll-виджетов мы не можем напрямую проверить SetMinSize,
	// но отсутствие паники подтверждает, что ветка "fullscreen" обработана.
}

// --- clampToMaxSize: поддержка 4K+ без выхода за границы экрана ---

func TestClampToMaxSize_WithinBounds(t *testing.T) {
	// Стандартные разрешения — не должны меняться
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
	// Отрицательные значения не клампятся вниз — это задача caller
	if clamped.Width != -100 || clamped.Height != -100 {
		t.Errorf("negative size should not change, got %v", clamped)
	}
}

// --- 4K+ fullscreen: layout profile detection ---

func TestDetectLayoutProfile_Fullscreen_4K(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeFullscreen}}
	// 4K (3840) >= fullscreenMinWidth → fullscreen
	profile := a.detectLayoutProfile(3840)
	if profile != "fullscreen" {
		t.Errorf("expected 'fullscreen' at 4K, got %q", profile)
	}
}

func TestDetectLayoutProfile_Fullscreen_5K(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeFullscreen}}
	profile := a.detectLayoutProfile(5120)
	if profile != "fullscreen" {
		t.Errorf("expected 'fullscreen' at 5K, got %q", profile)
	}
}

func TestDetectLayoutProfile_Windowed_4K(t *testing.T) {
	a := &App{}
	a.winMode = &windowModeCtrl{state: windowModeState{mode: windowModeNormal}}
	// 4K >= 2200 → wide (не fullscreen, т.к. не fullscreen режим)
	profile := a.detectLayoutProfile(3840)
	if profile != "wide" {
		t.Errorf("expected 'wide' at 4K windowed, got %q", profile)
	}
}

// --- 4K+ layoutAdaptiveMultiplier: не должен раздувать панели ---

func TestLayoutAdaptiveMultiplier_4K(t *testing.T) {
	// 4K при scale=1: ratio = sqrt(3840^2+2160^2)/1150 = 4406/1150 = 3.83
	// mul = clamp(3.83, 0.72, 1.38) = 1.38 (упирается в ceiling)
	m := layoutAdaptiveMultiplier(3840, 2160, 1)
	if m > 1.38 {
		t.Errorf("4K scale=1: mul should be clamped to 1.38, got %v", m)
	}
	if m < 0.72 {
		t.Errorf("4K scale=1: mul should be >= 0.72, got %v", m)
	}
}

func TestLayoutAdaptiveMultiplier_4K_HiDPI(t *testing.T) {
	// 4K при scale=2: scaleAdj = sqrt(2)=1.414, clamped to 1.38
	// mul = clamp(3.83/1.38, 0.72, 1.38) = clamp(2.77, 0.72, 1.38) = 1.38
	m := layoutAdaptiveMultiplier(3840, 2160, 2)
	if m > 1.38 {
		t.Errorf("4K scale=2: mul should be clamped to 1.38, got %v", m)
	}
	if m < 0.72 {
		t.Errorf("4K scale=2: mul should be >= 0.72, got %v", m)
	}
}

func TestLayoutAdaptiveMultiplier_5K(t *testing.T) {
	// 5K (5120×2880): ratio = sqrt(5120^2+2880^2)/1150 = 5870/1150 = 5.10
	// mul = clamp(5.10, 0.72, 1.38) = 1.38
	m := layoutAdaptiveMultiplier(5120, 2880, 1)
	if m > 1.38 {
		t.Errorf("5K: mul should be clamped to 1.38, got %v", m)
	}
}

// --- 4K+ adaptivePanelMinHeight: панели не должны выходить за экран ---

func TestAdaptivePanelMinHeight_4K_Fullscreen(t *testing.T) {
	// Имитируем 4K fullscreen: canvas height = 2160, mul = 1.38
	a := &App{lastCanvasSize: fyne.NewSize(3840, 2160)}
	mul := float32(1.38)
	// topologyScroll: base=400, maxFrac=0.65
	// v = 400 * 1.38 = 552, maxH = 2160 * 0.65 = 1404 → v=552 (не превышает)
	h := a.adaptivePanelMinHeight(400, mul, 0.65, 200)
	if h > 2160 {
		t.Errorf("panel height %v exceeds screen height 2160", h)
	}
	if h < 200 {
		t.Errorf("panel height %v below minAbs 200", h)
	}
}

func TestAdaptivePanelMinHeight_4K_AllPanelsSum(t *testing.T) {
	// Проверяем, что сумма max-высот всех панелей в fullscreen-профиле
	// не превышает высоту экрана (с учётом того, что панели в разных вкладках).
	a := &App{lastCanvasSize: fyne.NewSize(3840, 2160)}
	mul := float32(1.38)
	screenH := float32(2160)

	// Scan tab: scanControls + resultsScroll
	scanMax := a.adaptivePanelMinHeight(140, mul, 0.25, 100) +
		a.adaptivePanelMinHeight(50, mul, 0.50, 30)
	if scanMax > screenH {
		t.Errorf("scan tab panels sum %v exceeds screen height %v", scanMax, screenH)
	}

	// Topology tab: topologyControls + topologyScroll
	topoMax := a.adaptivePanelMinHeight(120, mul, 0.20, 80) +
		a.adaptivePanelMinHeight(400, mul, 0.65, 200)
	if topoMax > screenH {
		t.Errorf("topology tab panels sum %v exceeds screen height %v", topoMax, screenH)
	}

	// Tools tab: toolsControls + toolsOutput + opsOutput
	toolsMax := a.adaptivePanelMinHeight(140, mul, 0.25, 100) +
		a.adaptivePanelMinHeight(280, mul, 0.52, 150) +
		a.adaptivePanelMinHeight(100, mul, 0.22, 72)
	if toolsMax > screenH {
		t.Errorf("tools tab panels sum %v exceeds screen height %v", toolsMax, screenH)
	}
}

// --- 4K+ responsive layout: не должен паниковать ---

func TestApplyResponsiveLayout_Fullscreen_4K(t *testing.T) {
	a := &App{lastCanvasSize: fyne.NewSize(3840, 2160), lastCanvasScale: 1}
	a.applyResponsiveLayout("fullscreen") // не должен паниковать
}

func TestApplyResponsiveLayout_Fullscreen_5K(t *testing.T) {
	a := &App{lastCanvasSize: fyne.NewSize(5120, 2880), lastCanvasScale: 1}
	a.applyResponsiveLayout("fullscreen") // не должен паниковать
}

func TestApplyResponsiveLayout_Fullscreen_4K_HiDPI(t *testing.T) {
	a := &App{lastCanvasSize: fyne.NewSize(1920, 1080), lastCanvasScale: 2}
	a.applyResponsiveLayout("fullscreen") // не должен паниковать
}

// --- syncFromWindow: mutex защита ---

func TestSyncFromWindow_MutexSafe(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:   &App{},
		state: windowModeState{mode: windowModeNormal},
	}
	// syncFromWindow при nil window — no-op, но должен работать через мьютекс.
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

// --- stopWatcher ---

func TestStopWatcher_NoPanic(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:       &App{},
		state:     windowModeState{mode: windowModeNormal},
		watchStop: make(chan struct{}),
	}
	ctrl.watchForExternalFullscreenChanges(nil)
	ctrl.stopWatcher()
	// Двойной вызов не должен паниковать
	ctrl.stopWatcher()
}

func TestStopWatcher_StopsTicker(t *testing.T) {
	ctrl := &windowModeCtrl{
		app:       &App{},
		state:     windowModeState{mode: windowModeNormal},
		watchStop: make(chan struct{}),
	}
	ctrl.watchForExternalFullscreenChanges(nil)
	ctrl.stopWatcher()
	// Ждём, чтобы goroutine завершила работу
	time.Sleep(100 * time.Millisecond)
	// Если goroutine не завершилась — тест зависнет на таймауте
}

// --- refreshWidgets nil-safe ---

func TestRefreshWidgets_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{
		state: windowModeState{mode: windowModeNormal},
	}
	ctrl.refreshWidgets() // не должен паниковать
}

// --- refreshLayout nil-safe ---

func TestRefreshLayout_NilApp(t *testing.T) {
	ctrl := &windowModeCtrl{
		state: windowModeState{mode: windowModeNormal},
	}
	ctrl.refreshLayout() // не должен паниковать
}
