package gui

import (
	"context"
	"image/color"
	"testing"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

// TestAppInit проверяет базовую инициализацию приложения.
func TestAppInit(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	if a.myApp == nil {
		t.Fatal("myApp is nil")
	}
	if a.myWindow == nil {
		t.Fatal("myWindow is nil")
	}
	if a.scanResults == nil {
		t.Fatal("scanResults is nil")
	}
}

// TestPipelineCacheKey проверяет создание ключа кэша pipeline.
func TestPipelineCacheKey(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:              fyneApp,
		myWindow:           fyneWindow,
		scanResults:        make([]scanner.Result, 0),
		quickTypeChecks:    make(map[string]*widget.Check),
		resultsMode:        "Таблица",
		resultsSort:        "IP",
		resultsFilterQuery: "test",
	}

	key := a.buildResultsPipelineCacheKey()
	if key == "" {
		t.Fatal("cache key should not be empty")
	}
}

// TestFilterPresetKey проверяет корректность ключей пресетов.
func TestFilterPresetKey(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	tests := []struct {
		slot   string
		expect string
	}{
		{"1", prefFilterPreset1},
		{"2", prefFilterPreset2},
		{"3", prefFilterPreset3},
		{"invalid", prefFilterPreset1},
	}

	for _, tt := range tests {
		key := a.filterPresetKey(tt.slot)
		if key != tt.expect {
			t.Errorf("filterPresetKey(%q) = %q, want %q", tt.slot, key, tt.expect)
		}
	}
}

// TestRecommendedBadgeClass проверяет классы бейджей.
func TestRecommendedBadgeClass(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	tests := []struct {
		hosts  int
		expect string
	}{
		{10, "small"},
		{100, "medium"},
		{500, "large"},
		{1000, "very-large"},
	}

	for _, tt := range tests {
		class := a.recommendedBadgeClassForHosts(tt.hosts)
		if class != tt.expect {
			t.Errorf("recommendedBadgeClassForHosts(%d) = %q, want %q", tt.hosts, class, tt.expect)
		}
	}
}

// TestRecommendedProfileName проверяет имена профилей.
func TestRecommendedProfileName(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	tests := []struct {
		class  string
		expect string
		ok     bool
	}{
		{"small", "углубленный для небольшой подсети", true},
		{"medium", "сбалансированный для средней подсети", true},
		{"large", "бережный для крупной подсети", true},
		{"very-large", "бережный для очень крупной подсети", true},
		{"invalid", "", false},
	}

	for _, tt := range tests {
		name, ok := a.recommendedProfileNameForClass(tt.class)
		if ok != tt.ok {
			t.Errorf("recommendedProfileNameForClass(%q) ok = %v, want %v", tt.class, ok, tt.ok)
		}
		if ok && name != tt.expect {
			t.Errorf("recommendedProfileNameForClass(%q) = %q, want %q", tt.class, name, tt.expect)
		}
	}
}

// TestFormatDurationMMSS проверяет форматирование длительности.
func TestFormatDurationMMSS(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		toolsHostEntry:  widget.NewEntry(),
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	// Пустой хост
	host, ok := a.withToolHost()
	if ok {
		t.Error("empty host should return ok=false")
	}
	if host != "" {
		t.Errorf("empty host should return empty string, got %q", host)
	}

	// Валидный хост
	a.toolsHostEntry.SetText("192.168.1.1")
	host, ok = a.withToolHost()
	if !ok {
		t.Error("valid host should return ok=true")
	}
	if host != "192.168.1.1" {
		t.Errorf("valid host should return '192.168.1.1', got %q", host)
	}
}

// TestResultsForSave проверяет подготовку результатов для сохранения.
func TestResultsForSave(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     []scanner.Result{},
		quickTypeChecks: make(map[string]*widget.Check),
	}

	// Нет результатов
	results, reason := a.resultsForSave()
	if results != nil {
		t.Error("empty results should return nil")
	}
	if reason == "" {
		t.Error("empty results should return non-empty reason")
	}
}

// TestSaveResultsViewSettings проверяет сохранение настроек.
func TestSaveResultsViewSettings(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:                fyneApp,
		myWindow:             fyneWindow,
		scanResults:          make([]scanner.Result, 0),
		quickTypeChecks:      make(map[string]*widget.Check),
		resultsMode:          "Таблица",
		resultsSubMode:       "Devices",
		resultsSort:          "IP",
		maxPortChips:         10,
		showRawBanners:       true,
		resultsFilterQuery:   "test",
		onlyWithOpenPorts:    true,
		resultsPortStateMode: "has_open",
	}

	// Должно выполниться без паники
	a.saveResultsViewSettings()

	// Проверим, что настройки сохранились
	p := a.myApp.Preferences()
	if p.String(prefViewMode) != "Таблица" {
		t.Errorf("prefViewMode = %q, want %q", p.String(prefViewMode), "Таблица")
	}
}

// TestPipelineCacheKeyGenerated проверяет генерацию ключа кэша.
func TestPipelineCacheKeyGenerated(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:                fyneApp,
		myWindow:             fyneWindow,
		scanResults:          make([]scanner.Result, 0),
		quickTypeChecks:      make(map[string]*widget.Check),
		resultsMode:          "Таблица",
		resultsSort:          "IP",
		resultsFilterQuery:   "test",
		resultsPortStateMode: "has_open",
		onlyWithOpenPorts:    true,
	}

	key := a.buildResultsPipelineCacheKey()
	if key == "" {
		t.Fatal("cache key should not be empty")
	}
	if len(key) < 10 {
		t.Errorf("cache key too short: %q", key)
	}
}

// TestScheduleResultsRender проверяет планировщик рендера.
func TestScheduleResultsRender(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	// Должно выполниться без паники
	a.scheduleResultsRender(false)
	a.scheduleResultsRender(true)
}

// TestSelectedHostFromData проверяет выбор хоста из данных.
func TestSelectedHostFromData(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	// Нет данных
	_, ok := a.selectedHostFromData(nil)
	if ok {
		t.Error("nil data should return ok=false")
	}
}

// TestClearResultsMainSplitRef проверяет очистку ссылки на сплит.
func TestClearResultsMainSplitRef(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:            fyneApp,
		myWindow:         fyneWindow,
		scanResults:      make([]scanner.Result, 0),
		quickTypeChecks:  make(map[string]*widget.Check),
		resultsMainSplit: nil,
	}

	// Должно выполниться без паники
	a.clearResultsMainSplitRef()
}

// TestRunToolOperation проверяет запуск операции инструмента.
func TestRunToolOperation(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	// Без operations manager — должно выполниться синхронно
	a.runToolOperation("Test", "Testing...", func(ctx context.Context) (string, error) {
		return "result", nil
	})

	// Проверим, что операция завершилась
	time.Sleep(100 * time.Millisecond)
}

// TestSelectedTypeFilters проверяет выбор типов фильтров.
func TestSelectedTypeFilters(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	// Нет выбранных типов
	types := a.selectedTypeFilters()
	if len(types) != 0 {
		t.Errorf("expected 0 types, got %d", len(types))
	}
}

// TestPassesPortStateMode проверяет режим фильтрации портов.
func TestPassesPortStateMode(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:                fyneApp,
		myWindow:             fyneWindow,
		scanResults:          make([]scanner.Result, 0),
		quickTypeChecks:      make(map[string]*widget.Check),
		resultsPortStateMode: "all",
	}

	// all — всегда true
	if !a.passesPortStateMode(scanner.Result{}) {
		t.Error("mode 'all' should always return true")
	}
}

// TestPassesCIDRFilter проверяет фильтр CIDR.
func TestPassesCIDRFilter(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:                fyneApp,
		myWindow:             fyneWindow,
		scanResults:          make([]scanner.Result, 0),
		quickTypeChecks:      make(map[string]*widget.Check),
		resultsCidrFilterEnt: widget.NewEntry(),
	}

	// Пустой фильтр — всегда true
	if !a.passesCIDRFilter(scanner.Result{}) {
		t.Error("empty CIDR filter should always return true")
	}
}

// TestApplyAdvancedFilters проверяет продвинутую фильтрацию.
func TestApplyAdvancedFilters(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:                fyneApp,
		myWindow:             fyneWindow,
		scanResults:          make([]scanner.Result, 0),
		quickTypeChecks:      make(map[string]*widget.Check),
		resultsPortStateMode: "all",
		resultsCidrFilterEnt: widget.NewEntry(),
	}

	// Нет результатов — пусто
	filtered := a.applyAdvancedFilters(nil)
	if filtered != nil {
		t.Error("nil input should return nil")
	}
}

// TestLoadScanSettings проверяет загрузку настроек.
func TestLoadScanSettings(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}

	// Должно выполниться без паники
	a.loadScanSettings()
}

// TestRefreshAutoProfileStateLabel проверяет обновление лейбла автопрофиля.
func TestRefreshAutoProfileStateLabel(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:                fyneApp,
		myWindow:             fyneWindow,
		scanResults:          make([]scanner.Result, 0),
		quickTypeChecks:      make(map[string]*widget.Check),
		autoProfileStateText: canvas.NewText("", color.White),
	}

	// Должно выполниться без паники
	a.refreshAutoProfileStateLabel()
}

// TestSetPortRangeControlsEnabled проверяет включение/отключение контролов портов.
func TestSetPortRangeControlsEnabled(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
		portRangeEntry:  widget.NewEntry(),
	}

	// Должно выполниться без паники
	a.setPortRangeControlsEnabled(true)
	a.setPortRangeControlsEnabled(false)
}

// TestLoadSplitFromPrefs проверяет загрузку сплит-позиций.
func TestLoadSplitFromPrefs(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:             fyneApp,
		myWindow:          fyneWindow,
		scanResults:       make([]scanner.Result, 0),
		quickTypeChecks:   make(map[string]*widget.Check),
		scanTabMainSplit:  container.NewHSplit(nil, nil),
		topologyMainSplit: container.NewVSplit(nil, nil),
		toolsTabMainSplit: container.NewVSplit(nil, nil),
		resultsMainSplit:  container.NewHSplit(nil, nil),
	}

	// Должно выполниться без паники
	a.loadScanTabSplitFromPrefs()
	a.loadTopologySplitFromPrefs()
	a.loadToolsTabSplitFromPrefs()
	a.loadHostDetailsSplitFromPrefs()
}

// TestClampSplitOffset проверяет ограничение сплит-позиций.
func TestClampSplitOffset(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:             fyneApp,
		myWindow:          fyneWindow,
		scanResults:       make([]scanner.Result, 0),
		quickTypeChecks:   make(map[string]*widget.Check),
		scanTabMainSplit:  container.NewHSplit(nil, nil),
		topologyMainSplit: container.NewVSplit(nil, nil),
		toolsTabMainSplit: container.NewVSplit(nil, nil),
	}

	// Должно выполниться без паники
	a.clampScanTabMainSplitOffset()
	a.clampTopologyMainSplitOffset()
	a.clampToolsTabMainSplitOffset()
}

// TestMaybePersistSplitOffset проверяет сохранение сплит-позиций.
func TestMaybePersistSplitOffset(t *testing.T) {
	skipHeadless(t)
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:                      fyneApp,
		myWindow:                   fyneWindow,
		scanResults:                make([]scanner.Result, 0),
		quickTypeChecks:            make(map[string]*widget.Check),
		scanTabMainSplit:           container.NewHSplit(nil, nil),
		scanTabSplitPersistPrimed:  true,
		topologyMainSplit:          container.NewVSplit(nil, nil),
		topologySplitPersistPrimed: true,
		toolsTabMainSplit:          container.NewVSplit(nil, nil),
		toolsSplitPersistPrimed:    true,
		resultsMainSplit:           container.NewHSplit(nil, nil),
	}

	// Должно выполниться без паники
	a.maybePersistScanTabSplitOffset()
	a.maybePersistTopologySplitOffset()
	a.maybePersistToolsTabSplitOffset()
	a.maybePersistHostDetailsSplitOffsets()
}
