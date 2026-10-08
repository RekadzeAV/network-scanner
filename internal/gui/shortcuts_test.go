package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// newShortcutsTestApp создаёт App с окном (headless-safe) для тестов шорткатов.
func newShortcutsTestApp(t *testing.T) *App {
	t.Helper()
	skipHeadless(t)
	_ = test.NewApp()
	a := &App{}
	a.myApp = fyneapp.New()
	a.myWindow = a.myApp.NewWindow("test-shortcuts")
	t.Cleanup(func() { a.myWindow.Close() })
	return a
}

// TestSetupAppShortcutsNoPanic проверяет, что регистрация шорткатов не паникует
// на пустом App и добавляет все комбинации на канвас.
func TestSetupAppShortcutsNoPanic(t *testing.T) {
	a := newShortcutsTestApp(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("setupAppShortcuts паниковал: %v", r)
		}
	}()
	a.setupAppShortcuts()
}

// TestSetupAppShortcutsNilSafe проверяет nil-устойчивость (headless, без окна).
func TestSetupAppShortcutsNilSafe(t *testing.T) {
	skipHeadless(t)
	var a *App
	a.setupAppShortcuts() // не должно паниковать

	bare := &App{}
	bare.setupAppShortcuts() // myWindow == nil — no-op
}

// TestRunShortcutActionsNoPanic прогоняет все зарегистрированные комбинации
// через runShortcutAction на пустом App (все виджеты nil) — ожидается no-op.
func TestRunShortcutActionsNoPanic(t *testing.T) {
	skipHeadless(t)
	a := newShortcutsTestApp(t)
	for _, sa := range a.appShortcuts() {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("runShortcutAction(%s) паниковал: %v", sa.displayComb, r)
				}
			}()
			a.runShortcutAction(sa.sc)
		}()
	}
}

// TestRunShortcutTabSelection проверяет, что Ctrl+1..3 выбирают вкладки mainTabs.
func TestRunShortcutTabSelection(t *testing.T) {
	skipHeadless(t)
	_ = test.NewApp()
	a := &App{}
	a.mainTabs = container.NewAppTabs(
		container.NewTabItem("Сканирование", widget.NewLabel("1")),
		container.NewTabItem("Топология", widget.NewLabel("2")),
		container.NewTabItem("Инструменты", widget.NewLabel("3")),
	)
	a.runShortcutAction(&desktop.CustomShortcut{KeyName: fyne.Key2, Modifier: fyne.KeyModifierControl})
	if a.mainTabs.SelectedIndex() != 1 {
		t.Fatalf("Ctrl+2: ожидалась вкладка 1, получена %d", a.mainTabs.SelectedIndex())
	}
	a.runShortcutAction(&desktop.CustomShortcut{KeyName: fyne.Key1, Modifier: fyne.KeyModifierControl})
	if a.mainTabs.SelectedIndex() != 0 {
		t.Fatalf("Ctrl+1: ожидалась вкладка 0, получена %d", a.mainTabs.SelectedIndex())
	}
	// Индекс вне диапазона — no-op
	a.runShortcutAction(&desktop.CustomShortcut{KeyName: fyne.Key9, Modifier: fyne.KeyModifierControl})
}

// TestAppShortcutsList проверяет полноту списка: у каждой записи есть label
// и отображаемая комбинация, KeyName непустой.
func TestAppShortcutsList(t *testing.T) {
	skipHeadless(t)
	a := &App{}
	list := a.appShortcuts()
	if len(list) == 0 {
		t.Fatal("appShortcuts() пуст")
	}
	seen := map[string]bool{}
	for _, sa := range list {
		if sa.label == "" || sa.displayComb == "" {
			cs, ok := sa.sc.(*desktop.CustomShortcut)
			if !ok || cs.KeyName == "" {
				t.Fatalf("некорректная запись: %+v", sa)
			}
		}
		key := sa.displayComb
		if seen[key] {
			t.Fatalf("дубликат комбинации: %s", key)
		}
		seen[key] = true
	}
}

// TestShowShortcutsDialogNoPanic проверяет, что диалог F1 не паникует в headless.
func TestShowShortcutsDialogNoPanic(t *testing.T) {
	a := newShortcutsTestApp(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("showShortcutsDialog паниковал: %v", r)
		}
	}()
	a.showShortcutsDialog()
}

// TestRunShortcutResultsSections — регрессия N-02: Ctrl+4/Ctrl+5 вели в
// несуществующие вкладки 3 и 4 и молча no-op'или. Теперь они открывают вкладку
// «Сканирование» и переключают подрежим Security/Inventory тем же путём, что и
// клик по радио-переключателю.
func TestRunShortcutResultsSections(t *testing.T) {
	a := &App{}
	a.initResultsStateAndWidgets()
	a.initResultsModeAndInventoryWidgets()
	a.mainTabs = container.NewAppTabs(
		container.NewTabItem("Сканирование", widget.NewLabel("scan")),
		container.NewTabItem("Топология", widget.NewLabel("topo")),
		container.NewTabItem("Инструменты", widget.NewLabel("tools")),
	)

	tests := []struct {
		name string
		key  fyne.KeyName
		want string
	}{
		{name: "Ctrl+4 -> Security", key: fyne.Key4, want: resultsSubModeSecurity},
		{name: "Ctrl+5 -> Inventory", key: fyne.Key5, want: resultsSubModeInventory},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a.mainTabs.SelectIndex(2) // заведомо не вкладка сканирования

			a.runShortcutAction(&desktop.CustomShortcut{KeyName: tc.key, Modifier: fyne.KeyModifierControl})

			if got := a.mainTabs.SelectedIndex(); got != resultsTabScanIndex {
				t.Fatalf("%s: вкладка = %d, ожидали %d", tc.name, got, resultsTabScanIndex)
			}
			if a.resultsSubMode != tc.want {
				t.Fatalf("%s: resultsSubMode = %q, ожидали %q", tc.name, a.resultsSubMode, tc.want)
			}
			if a.resultsSubModeSel.Selected != tc.want {
				t.Fatalf("%s: переключатель показывает %q, ожидали %q",
					tc.name, a.resultsSubModeSel.Selected, tc.want)
			}
		})
	}

	// Повторное сочетание при уже выбранном подрежиме не должно ломать состояние.
	a.runShortcutAction(&desktop.CustomShortcut{KeyName: fyne.Key5, Modifier: fyne.KeyModifierControl})
	if a.resultsSubMode != resultsSubModeInventory {
		t.Fatalf("повторный Ctrl+5 изменил подрежим на %q", a.resultsSubMode)
	}
}

// TestShowResultsSectionShortcutNilSafe — nil-App, nil-переключатель и
// неизвестное имя подрежима не должны паниковать и менять состояние.
func TestShowResultsSectionShortcutNilSafe(t *testing.T) {
	(&App{}).showResultsSectionShortcut(resultsSubModeSecurity)

	a := &App{resultsSubMode: resultsSubModeDevices}
	a.showResultsSectionShortcut("Что-то не то")
	if a.resultsSubMode != resultsSubModeDevices {
		t.Fatalf("неизвестное имя изменило подрежим на %q", a.resultsSubMode)
	}
}

// TestNormalizeResultsSubMode — значения, попадающие в prefResultsSubMode и в
// обработчик Ctrl+4/Ctrl+5, должны проходить один и тот же whitelist.
func TestNormalizeResultsSubMode(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: resultsSubModeDevices, want: resultsSubModeDevices},
		{value: resultsSubModeSecurity, want: resultsSubModeSecurity},
		{value: resultsSubModeInventory, want: resultsSubModeInventory},
		{value: "  Security\n", want: resultsSubModeSecurity},
		{value: "", want: ""},
		{value: "security", want: ""},
		{value: "Inventory ", want: resultsSubModeInventory},
	}
	for _, tc := range tests {
		if got := normalizeResultsSubMode(tc.value); got != tc.want {
			t.Fatalf("normalizeResultsSubMode(%q) = %q, ожидали %q", tc.value, got, tc.want)
		}
	}

	// Опции переключателя и whitelist обязаны совпадать.
	a := &App{}
	a.initResultsStateAndWidgets()
	a.initResultsModeAndInventoryWidgets()
	if len(a.resultsSubModeSel.Options) != len(resultsSubModeOptions) {
		t.Fatalf("опций переключателя %d, ожидали %d",
			len(a.resultsSubModeSel.Options), len(resultsSubModeOptions))
	}
	for _, opt := range a.resultsSubModeSel.Options {
		if normalizeResultsSubMode(opt) == "" {
			t.Fatalf("опция %q не проходит whitelist — не восстановится из настроек", opt)
		}
	}
}
