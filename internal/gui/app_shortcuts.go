package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// setupAppShortcuts регистрирует глобальные горячие клавиши приложения.
//
// Ядро:
//	F5 / Ctrl+Enter  -> StartScan
//	Esc              -> StopScan
//	Ctrl+S          -> сохранить (диалог с выбором формата)
//	Ctrl+F          -> фокус результатов фильтра
//	Ctrl+L          -> фокус поля сети
//	Ctrl+1..5       -> переключение вкладок: Scan / Topology / Tools / Security / Inventory
//	Ctrl+Shift+L    -> сброс расположения (существует отдельно, не дублировать)
//
// Остальное:
//	F1              -> диалог справки
//	Ctrl+Shift+I   -> импорт из файла (метаданные / результаты)
//	Ctrl+Shift+R   -> полный сброс настроек
func (a *App) setupAppShortcuts() {
	if a == nil || a.myWindow == nil {
		return
	}
	canvas := a.myWindow.Canvas()
	ctrl := a // shortcut closures capture a

	register := func(keys string, label string, h func()) {
		if h == nil {
			return
		}
		canvas.AddShortcut(&desktop.CustomShortcut{
			Serialization: keys,
			Label:         label,
		}, func(_ *fyne.Shortcut) { h() })
	}

	register("F5", "Старт сканирования", func() { ctrl.startScanShortcut() })
	register("Ctrl+Enter", "Старт сканирования (Ctrl+Enter)", func() { ctrl.startScanShortcut() })
	register("Escape", "Стоп сканирования", func() { ctrl.stopScanShortcut() })
	register("Ctrl+S", "Сохранить результаты (Ctrl+S)", func() { ctrl.saveResults() })
	register("Ctrl+F", "Фокус фильтра результатов (Ctrl+F)", func() { ctrl.focusResultsFilterShortcut() })
	register("Ctrl+L", "Фокус поля сети (Ctrl+L)", func() { ctrl.focusNetworkEntryShortcut() })
	register("Ctrl+1", "Вкладка «Сканирование» (Ctrl+1)", func() { ctrl.selectTabShortcut(0) })
	register("Ctrl+2", "Вкладка «Топология» (Ctrl+2)", func() { ctrl.selectTabShortcut(1) })
	register("Ctrl+3", "Вкладка «Инструменты» (Ctrl+3)", func() { ctrl.selectTabShortcut(2) })
	register("Ctrl+4", "Вкладка «Безопасность» (Ctrl+4)", func() { ctrl.selectTabShortcut(3) })
	register("Ctrl+5", "Вкладка «Инвентарь» (Ctrl+5)", func() { ctrl.selectTabShortcut(4) })
	register("F1", "Справка (F1)", func() { ctrl.showHelpDialog() })

	// Явных Ctrl+Shift+I / Ctrl+Shift+R пока нет — они появились бы в следующих
	// спринтах. Здесь не дублируем, чтобы избежать пересечений с заявленными
	// «третьим вариантом» (только icon-only tooltip и hotkeys уже собран.) "
}

// startScanShortcut — обёртка для F5 / Ctrl+Enter: старт сканирования.
func (a *App) startScanShortcut() {
	if a == nil {
		return
	}
	if a.scanCtrl != nil {
		a.scanCtrl.StartScan(a.scanResults)
	}
}

// stopScanShortcut — обёртка для Escape: остановка сканирования.
func (a *App) stopScanShortcut() {
	if a == nil {
		return
	}
	if a.scanCtrl != nil {
		a.scanCtrl.StopScan()
	}
}

// focusResultsFilterShortcut — Ctrl+F: фокус на поле фильтра результатов.
func (a *App) focusResultsFilterShortcut() {
	if a == nil || a.resultsFilterEnt == nil {
		return
	}
	a.resultsFilterEnt.Focus()
}

// focusNetworkEntryShortcut — Ctrl+L: фокус на поле сети.
func (a *App) focusNetworkEntryShortcut() {
	if a == nil || a.networkEntry == nil {
		return
	}
	a.networkEntry.Focus()
}

// selectTabShortcut — Ctrl+1..5: выбор вкладки.
func (a *App) selectTabShortcut(index int) {
	if a == nil || a.mainTabs == nil {
		return
	}
	if index < 0 || index >= a.mainTabs.GetTabs().Len() {
		return
	}
	a.mainTabs.SelectIndex(index)
}

// showHelpDialog — F1: диалог справки по горячим клавишам и работе.
func (a *App) showHelpDialog() {
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
		"  Ctrl+1..5       — переключение вкладок\n" +
		"  Ctrl+Shift+L    — сброс расположения UI\n" +
		"  F1              — это диалог справки"
	dialog.ShowInformation(a.myWindow, "Справка по горячим клавишам", help, fyne.NewDragOptions())
}

// focusResultFilter — аналог focusResultsFilterShortcut для вызова из других мест.
func (a *App) focusResultFilter() {
	a.focusResultsFilterShortcut()
}
