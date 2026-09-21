package gui

import (
	"time"

	"fyne.io/fyne/v2"
)

// --- Статус-бар и неблокирующие уведомления (R7) ---

const (
	statusToastDuration = 4 * time.Second
	statusDefaultText   = "Готов к сканированию"
)

// setStatus устанавливает текст статуса в статус-баре.
// Используйте setStatusToast для временных уведомлений с авто-возвратом.
func (a *App) setStatus(text string) {
	if a == nil || a.statusLabel == nil {
		return
	}
	a.cancelStatusToast()
	a.statusLabel.SetText(text)
}

// setStatusToast показывает временное уведомление в статус-баре: сообщение
// отображается с маркером "✓ ", через 4 секунды статус возвращается к дефолту.
func (a *App) setStatusToast(text string) {
	if a == nil || a.statusLabel == nil {
		return
	}
	a.cancelStatusToast()
	a.statusLabel.SetText("✓ " + text)
	a.statusToastTimerMu.Lock()
	a.statusToastTimer = time.AfterFunc(statusToastDuration, func() {
		fyne.Do(func() {
			if a == nil || a.statusLabel == nil {
				return
			}
			a.statusLabel.SetText(statusDefaultText)
		})
	})
	a.statusToastTimerMu.Unlock()
}

// cancelStatusToast отменяет активный таймер тоста.
func (a *App) cancelStatusToast() {
	if a == nil {
		return
	}
	a.statusToastTimerMu.Lock()
	if a.statusToastTimer != nil {
		a.statusToastTimer.Stop()
		a.statusToastTimer = nil
	}
	a.statusToastTimerMu.Unlock()
}

// updateResultsFiltersVisibility автоматически сворачивает/разворачивает
// Accordion «Фильтры и режимы отображения»: в подрежиме Inventory фильтры
// устройств не нужны и панель скрывается, в остальных — разворачивается.
func (a *App) updateResultsFiltersVisibility() {
	if a == nil || a.resultsFiltersAccordion == nil || len(a.resultsFiltersAccordion.Items) == 0 {
		return
	}
	wantOpen := a.resultsSubMode != "Inventory"
	if a.resultsFiltersHidden && !wantOpen {
		return
	}
	a.resultsFiltersHidden = !wantOpen
	a.resultsFiltersAccordion.Items[0].Open = wantOpen
	a.resultsFiltersAccordion.Refresh()
}
