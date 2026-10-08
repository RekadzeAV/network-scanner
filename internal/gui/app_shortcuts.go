package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
)

// --- App — горячие клавиши (shortcuts) -------------------------------------
//
// Ядро:
//	F5 / Ctrl+Enter  -> StartScan
//	Esc              -> StopScan
//	Ctrl+S          -> сохранить (диалог с выбором формата)
//	Ctrl+F          -> фокус результатов фильтра
//	Ctrl+L          -> фокус поля сети
//	Ctrl+1..3       -> переключение вкладок: Scan / Topology / Tools
//	Ctrl+4/5        -> подрежимы вкладки Scan: Security / Inventory (вкладок под них нет)
//	Ctrl+Shift+L    -> сброс расоложения (существует отдельно, не дублировать)
//
// Остальное:
//	F1              -> диалог справки
//
// Ctrl+Shift+I (импорт) и Ctrl+Shift+R (сброс настроек) НЕ зарегистрированы —
// они только в планах (D-05: раньше комментарий заявлял их существующими,
// диалог F1 их не показывал).

// shortcutSpec описывает одну комбинацию клавиш для регистрации.
type shortcutSpec struct {
	displayComb string
	label       string
	keyName     fyne.KeyName
	modifier    fyne.KeyModifier
}

// appShortcutSpecs таблица всех горячих клавиш приложения.
var appShortcutSpecs = []shortcutSpec{
	{displayComb: "F5", label: "Старт сканирования", keyName: fyne.KeyF5, modifier: 0},
	{displayComb: "Ctrl+Enter", label: "Старт сканирования (Ctrl+Enter)", keyName: fyne.KeyReturn, modifier: fyne.KeyModifierControl},
	{displayComb: "Escape", label: "Стоп сканирования", keyName: fyne.KeyEscape, modifier: 0},
	{displayComb: "Ctrl+S", label: "Сохранить результаты (Ctrl+S)", keyName: fyne.KeyS, modifier: fyne.KeyModifierControl},
	{displayComb: "Ctrl+F", label: "Фокус фильтра результатов (Ctrl+F)", keyName: fyne.KeyF, modifier: fyne.KeyModifierControl},
	{displayComb: "Ctrl+L", label: "Фокус поля сети (Ctrl+L)", keyName: fyne.KeyL, modifier: fyne.KeyModifierControl},
	{displayComb: "Ctrl+1", label: "Вкладка Сканирование (Ctrl+1)", keyName: fyne.Key1, modifier: fyne.KeyModifierControl},
	{displayComb: "Ctrl+2", label: "Вкладка Топология (Ctrl+2)", keyName: fyne.Key2, modifier: fyne.KeyModifierControl},
	{displayComb: "Ctrl+3", label: "Вкладка Инструменты (Ctrl+3)", keyName: fyne.Key3, modifier: fyne.KeyModifierControl},
	{displayComb: "Ctrl+4", label: "Раздел «Безопасность» (Ctrl+4)", keyName: fyne.Key4, modifier: fyne.KeyModifierControl},
	{displayComb: "Ctrl+5", label: "Раздел «Инвентарь» (Ctrl+5)", keyName: fyne.Key5, modifier: fyne.KeyModifierControl},
	{displayComb: "F1", label: "Справка (F1)", keyName: fyne.KeyF1, modifier: 0},
}

// setupAppShortcuts регистрирует глобальные горячие клавиши приложения.
func (a *App) setupAppShortcuts() {
	if a == nil || a.myWindow == nil {
		return
	}
	canvas := a.myWindow.Canvas()

	for _, spec := range appShortcutSpecs {
		spec := spec
		handler := a.shortcutHandlerFor(spec.displayComb)
		if handler == nil {
			continue
		}
		canvas.AddShortcut(&desktop.CustomShortcut{
			KeyName:  spec.keyName,
			Modifier: spec.modifier,
		}, func(shortcut fyne.Shortcut) { handler() })
	}

	// Явных Ctrl+Shift+I / Ctrl+Shift+R пока нет — они появятся в следующих
	// спринтах. Не дублируем, чтобы избежать пересечений с заявленным
	// «третьим вариантом» (только icon-only tooltip и hotkeys уже собраны).
}

// shortcutHandlerFor возвращает обработчик для комбинации клавиш по строковому
// идентификатору. Возвращает nil, если обработчик не назначен.
func (a *App) shortcutHandlerFor(displayComb string) func() {
	switch displayComb {
	case "F5", "Ctrl+Enter":
		return a.startScanShortcut
	case "Escape":
		return a.stopScanShortcut
	case "Ctrl+S":
		return a.saveResults
	case "Ctrl+F":
		return a.focusResultsFilterShortcut
	case "Ctrl+L":
		return a.focusNetworkEntryShortcut
	case "Ctrl+1":
		return func() { a.selectTabShortcut(0) }
	case "Ctrl+2":
		return func() { a.selectTabShortcut(1) }
	case "Ctrl+3":
		return func() { a.selectTabShortcut(2) }
	case "Ctrl+4":
		return func() { a.showResultsSectionShortcut(resultsSubModeSecurity) }
	case "Ctrl+5":
		return func() { a.showResultsSectionShortcut(resultsSubModeInventory) }
	case "F1":
		return a.showShortcutsDialog
	default:
		return nil
	}
}

// startScanShortcut обертка для F5 / Ctrl+Enter: старт сканирования.
func (a *App) startScanShortcut() {
	if a == nil {
		return
	}
	if a.scanCtrl != nil {
		a.scanCtrl.StartScan(a.scanResults)
	}
}

// stopScanShortcut обертка для Escape: остановка сканирования.
func (a *App) stopScanShortcut() {
	if a == nil {
		return
	}
	if a.scanCtrl != nil {
		a.scanCtrl.StopScan()
	}
}

// focusResultsFilterShortcut Ctrl+F: фокус на поле фильтра результатов.
func (a *App) focusResultsFilterShortcut() {
	if a == nil || a.resultsFilterEnt == nil || a.myWindow == nil {
		return
	}
	a.myWindow.Canvas().Focus(a.resultsFilterEnt)
}

// focusNetworkEntryShortcut Ctrl+L: фокус на поле сети.
func (a *App) focusNetworkEntryShortcut() {
	if a == nil || a.networkEntry == nil || a.myWindow == nil {
		return
	}
	a.myWindow.Canvas().Focus(a.networkEntry)
}

// resultsTabScanIndex — индекс вкладки «Сканирование» в mainTabs (порядок
// задаётся в init_ui.go). Разделы «Безопасность»/«Инвентарь» — подрежимы именно
// этой вкладки, поэтому Ctrl+4/Ctrl+5 сначала открывают её.
const resultsTabScanIndex = 0

// selectTabShortcut Ctrl+1..3: выбор вкладки.
func (a *App) selectTabShortcut(index int) {
	if a == nil || a.mainTabs == nil {
		return
	}
	if index < 0 || index >= len(a.mainTabs.Items) {
		return
	}
	a.mainTabs.SelectIndex(index)
}

// showResultsSectionShortcut Ctrl+4/Ctrl+5: разделы «Безопасность» и «Инвентарь».
// Отдельных вкладок под них нет — это подрежимы вкладки «Сканирование» (N-02),
// поэтому сочетание открывает вкладку сканирования и переключает подрежим тем же
// путём, что и клик по радио-переключателю: SetSelected вызывает OnChanged,
// который сохраняет настройку, перестраивает видимость фильтров и рендер.
func (a *App) showResultsSectionShortcut(subMode string) {
	if a == nil {
		return
	}
	subMode = normalizeResultsSubMode(subMode)
	if subMode == "" {
		return
	}
	a.selectTabShortcut(resultsTabScanIndex)
	if a.resultsSubModeSel == nil {
		a.resultsSubMode = subMode
		return
	}
	// RadioGroup.SetSelected молчит, если значение уже выбрано: в этом случае
	// достаточно открытой вкладки — состояние и вывод уже корректны.
	a.resultsSubModeSel.SetSelected(subMode)
}

// showAboutDialog показывает диалог «О программе» с версией приложения.
func (a *App) showAboutDialog() {
	if a == nil || a.myWindow == nil {
		return
	}
	about := "Network Scanner — сканер сети с GUI.\n\n" +
		"Версия: " + guiVersion + "\n" +
		"GUI: Fyne v2.7\n" +
		"Сборка: " + BuildInfo() + "\n\n" +
		"Справка по горячим клавишам: F1"
	dialog.ShowInformation("О программе", about, a.myWindow)
}

// showShortcutsDialog F1 / пункт меню Вид: диалог справки по горячим клавишам.
func (a *App) showShortcutsDialog() {
	if a == nil || a.myWindow == nil {
		return
	}
	help := "Горячие клавиши:\n" +
		"\n" +
		"  F5 / Ctrl+Enter  — запустить сканирование\n" +
		"  Esc              — остановить сканирование\n" +
		"  Ctrl+S          — сохранить результаты (выбор формата)\n" +
		"  Ctrl+F          — фокус фильтра результатов\n" +
		"  Ctrl+L          — фокус поля сети\n" +
		"  Ctrl+1..3       — вкладки: Сканирование / Топология / Инструменты\n" +
		"  Ctrl+4/5        — разделы «Безопасность» / «Инвентарь» на вкладке сканирования\n" +
		"  Ctrl+Shift+L    — сброс расположения UI\n" +
		"  F1              — это диалог справки"
	dialog.ShowInformation("Справка по горячим клавишам", help, a.myWindow)
}

// appShortcuts возвращает список зарегистрированных комбинаций для тестов.
func (a *App) appShortcuts() []struct {
	displayComb string
	label       string
	sc          fyne.Shortcut
} {
	if a == nil || a.myWindow == nil {
		return nil
	}
	out := make([]struct {
		displayComb string
		label       string
		sc          fyne.Shortcut
	}, 0, len(appShortcutSpecs))
	for _, spec := range appShortcutSpecs {
		out = append(out, struct {
			displayComb string
			label       string
			sc          fyne.Shortcut
		}{
			displayComb: spec.displayComb,
			label:       spec.label,
			sc: &desktop.CustomShortcut{
				KeyName:  spec.keyName,
				Modifier: spec.modifier,
			},
		})
	}
	return out
}

// runShortcutAction тестовая обёртка пробного вызова шортката
// (no-op на nil-виджетах). Сопоставляет комбинацию с обработчиком.
func (a *App) runShortcutAction(sc fyne.Shortcut) {
	if a == nil {
		return
	}
	cs, ok := sc.(*desktop.CustomShortcut)
	if !ok {
		return
	}
	for _, spec := range appShortcutSpecs {
		if fyne.KeyName(cs.Key()) == spec.keyName && cs.Mod() == spec.modifier {
			handler := a.shortcutHandlerFor(spec.displayComb)
			if handler != nil {
				handler()
			}
			return
		}
	}
}
