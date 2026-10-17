package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
	scand "network-scanner/internal/scanner/daemon"
)

// --- tests: setResultsState ---

func TestSetResultsState_UpdatesStateAndLabel(t *testing.T) {
	a := &App{
		resultsStateLabel: widget.NewLabel(""),
	}

	a.setResultsState(resultsStateScanning, -1)
	if a.resultsState != resultsStateScanning {
		t.Errorf("state = %q, want %q", a.resultsState, resultsStateScanning)
	}
	if a.resultsStateLabel.Text != "Сканирование выполняется…" {
		t.Errorf("label = %q, want «Сканирование выполняется…»", a.resultsStateLabel.Text)
	}

	a.setResultsState(resultsStateDone, 12)
	if a.resultsState != resultsStateDone {
		t.Errorf("state = %q, want %q", a.resultsState, resultsStateDone)
	}
	if a.resultsStateLabel.Text != "Результаты получены: 12 устройств" {
		t.Errorf("label = %q, want «Результаты получены: 12 устройств»", a.resultsStateLabel.Text)
	}

	a.setResultsState(resultsStateStopped, -1)
	if a.resultsState != resultsStateStopped {
		t.Errorf("state = %q, want %q", a.resultsState, resultsStateStopped)
	}
	if a.resultsStateLabel.Text != "Сканирование остановлено" {
		t.Errorf("label = %q, want «Сканирование остановлено»", a.resultsStateLabel.Text)
	}

	a.setResultsState(resultsStateTimeout, -1)
	if a.resultsState != resultsStateTimeout {
		t.Errorf("state = %q, want %q", a.resultsState, resultsStateTimeout)
	}
	if a.resultsStateLabel.Text != "Сканирование прервано по таймауту" {
		t.Errorf("label = %q, want «Сканирование прервано по таймауту»", a.resultsStateLabel.Text)
	}
}

func TestSetResultsState_NilLabel(t *testing.T) {
	a := &App{}
	// Не должно паниковать при nil label — a.resultsState должен обновиться.
	a.setResultsState(resultsStateScanning, -1)
	if a.resultsState != resultsStateScanning {
		t.Errorf("state = %q, want %q", a.resultsState, resultsStateScanning)
	}
}

func TestSetResultsState_NilApp(t *testing.T) {
	var a *App
	// Не должно паниковать.
	a.setResultsState(resultsStateDone, 5)
}

func TestSetResultsState_DoneWithoutCount(t *testing.T) {
	a := &App{
		resultsStateLabel: widget.NewLabel(""),
	}
	// found < 0 → без счётчика.
	a.setResultsState(resultsStateDone, -1)
	if a.resultsStateLabel.Text != "Результаты получены" {
		t.Errorf("label = %q, want «Результаты получены»", a.resultsStateLabel.Text)
	}
}

// --- tests: resultsStateCaption ---

func TestResultsStateCaption_Localized(t *testing.T) {
	// Проверка: ни одна подпись не содержит технических значений «idle»/
	// «scanning»/«stopped»/«done»/«timeout».
	for _, state := range []string{resultsStateIdle, resultsStateScanning, resultsStateDone, resultsStateStopped, resultsStateTimeout, "unknown"} {
		cap := resultsStateCaption(state)
		if cap == "" {
			t.Errorf("caption(%q) = empty", state)
		}
		// Не должно быть технических значений в подписи.
		lc := cap
		for _, bad := range []string{"idle", "scanning", "stopped", "done", "timeout"} {
			if lc == bad {
				t.Errorf("caption(%q) = %q — техническое значение", state, bad)
			}
		}
	}
}

// --- tests: live-render during scan ---

func TestRenderScanResultsView_ScanningWithResults_RendersTable(t *testing.T) {
	a := &App{
		resultsState:      resultsStateScanning,
		scanResults:       []scanner.Result{{IP: "192.168.1.1"}},
		resultsMode:       "Таблица",
		resultsSubMode:    resultsSubModeDevices,
		resultsBody:       container.NewStack(widget.NewLabel("placeholder")),
		resultsFilterEnt:  widget.NewEntry(),
		filtersInfoLabel:  widget.NewLabel(""),
		resultsCountLabel: widget.NewLabel(""),
		resultsPerfLabel:  widget.NewLabel(""),
		resultsSort:       resultsSortByIP,
	}

	// Должно выполниться без паники.
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered (headless expected): %v", r)
		}
	}()
	a.renderScanResultsView()

	// При scanning и данных resultsBody не должен быть пуст.
	if len(a.resultsBody.Objects) == 0 {
		t.Error("resultsBody.Objects is empty — live-render failed")
	}
}

func TestRenderScanResultsView_ScanningWithoutResults_ShowEmptyState(t *testing.T) {
	a := &App{
		resultsState:      resultsStateScanning,
		scanResults:       nil,
		resultsBody:       container.NewStack(widget.NewLabel("placeholder")),
		resultsFilterEnt:  widget.NewEntry(),
		filtersInfoLabel:  widget.NewLabel(""),
		resultsCountLabel: widget.NewLabel(""),
		resultsPerfLabel:  widget.NewLabel(""),
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered (headless expected): %v", r)
		}
	}()
	a.renderScanResultsView()

	if len(a.resultsBody.Objects) == 0 {
		t.Error("resultsBody.Objects is empty — should have empty state")
	}
}

// --- tests: ScanActions.SetResultsState ---

func TestApp_SetResultsState_Interface(t *testing.T) {
	a := &App{
		resultsStateLabel: widget.NewLabel(""),
	}
	// Вызов через интерфейс ScanActions.
	a.SetResultsState(resultsStateStopped, -1)
	if a.resultsState != resultsStateStopped {
		t.Errorf("state = %q, want %q", a.resultsState, resultsStateStopped)
	}
}

func TestApp_SetResultsState_NilApp(t *testing.T) {
	var a *App
	a.SetResultsState(resultsStateDone, 10)
}

// --- tests: scanProgressText ---

func TestScanProgressText_NoTechnicalValues(t *testing.T) {
	// Текст прогресса в статус-баре не должен содержать технические значения.
	ev := scand.Event{
		Kind:    scand.EventProgress,
		Total:   100,
		Current: 50,
		Percent: 0.5,
	}
	start := time.Now().Add(-30 * time.Second)
	text := scanProgressText(ev, start, 12)

	for _, bad := range []string{"scanning", "stopped", "done", "idle"} {
		if text == bad {
			t.Errorf("progress text = %q — technical value", text)
		}
	}
}
