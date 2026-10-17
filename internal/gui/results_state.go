package gui

import "fmt"

// --- Состояния области результатов ------------------------------------------
//
// Единая точка перехода состояний (scanning/done/stopped/timeout) и локализованной
// подписи состояния (N-09). В подписи показывается только человекочитаемый текст —
// технические значения констант в UI не выводятся.

// resultsStateCaption — локализованная подпись состояния результатов.
// Неизвестное состояние даёт нейтральную подпись ожидания.
func resultsStateCaption(state string) string {
	switch state {
	case resultsStateScanning:
		return "Сканирование выполняется…"
	case resultsStateDone:
		return "Результаты получены"
	case resultsStateStopped:
		return "Сканирование остановлено"
	case resultsStateTimeout:
		return "Сканирование прервано по таймауту"
	default:
		return "Ожидание сканирования"
	}
}

// resultsStateText — подпись состояния с числом найденных устройств.
// found < 0 — счётчик не показывать.
func resultsStateText(state string, found int) string {
	caption := resultsStateCaption(state)
	if found < 0 || state != resultsStateDone {
		return caption
	}
	return fmt.Sprintf("%s: %d устройств", caption, found)
}

// setResultsState переводит область результатов в состояние state и обновляет
// подпись a.resultsStateLabel. Метод nil-safe: вызывается из контроллера и
// наблюдателя, где виджеты могут быть ещё не созданы.
func (a *App) setResultsState(state string, found int) {
	if a == nil {
		return
	}
	a.resultsState = state
	if a.resultsStateLabel == nil {
		return
	}
	a.resultsStateLabel.SetText(resultsStateText(state, found))
}

// SetResultsState реализует controller.ScanActions: контроллер сообщает о
// переходе состояния, перевод строки и запись в UI остаются в пакете gui.
func (a *App) SetResultsState(state string, found int) {
	a.setResultsState(state, found)
}
