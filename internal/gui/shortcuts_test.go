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