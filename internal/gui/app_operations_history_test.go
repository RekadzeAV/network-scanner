package gui

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2/widget"
)

// --- app.go: operations history tests ---

func TestPushOperationHistory_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.pushOperationHistory(Operation{})
}

func TestPushOperationHistory_NilOutput(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.pushOperationHistory(Operation{})
}

func TestPushOperationHistory_AddOp(t *testing.T) {
	a := &App{}
	a.operationsOutput = widget.NewRichText()
	op := Operation{ID: "op1", Title: "Ping", Status: OperationRunning}
	a.pushOperationHistory(op)
	if len(a.operationsHistory) != 1 {
		t.Errorf("expected 1 operation, got %d", len(a.operationsHistory))
	}
}

func TestPushOperationHistory_UpdateExisting(t *testing.T) {
	a := &App{}
	a.operationsOutput = widget.NewRichText()
	a.operationsHistory = []Operation{{ID: "op1", Title: "Ping", Status: OperationQueued}}
	op := Operation{ID: "op1", Title: "Ping", Status: OperationSuccess}
	a.pushOperationHistory(op)
	if len(a.operationsHistory) != 1 {
		t.Errorf("expected 1 operation (updated), got %d", len(a.operationsHistory))
	}
	if a.operationsHistory[0].Status != OperationSuccess {
		t.Errorf("expected status updated to completed")
	}
}

func TestPushOperationHistory_MaxHistory(t *testing.T) {
	a := &App{}
	a.operationsOutput = widget.NewRichText()
	for i := 0; i < 25; i++ {
		a.pushOperationHistory(Operation{ID: "op" + string(rune(i)), Title: "test"})
	}
	if len(a.operationsHistory) > 20 {
		t.Errorf("expected max 20 operations, got %d", len(a.operationsHistory))
	}
}

func TestOperationsHistoryMarkdown_Empty(t *testing.T) {
	a := &App{}
	result := a.operationsHistoryMarkdown()
	if result == "" {
		t.Error("expected non-empty markdown")
	}
}

func TestOperationsHistoryMarkdown_WithOps(t *testing.T) {
	a := &App{}
	a.operationsHistory = []Operation{
		{ID: "op1", Title: "Ping", Status: OperationSuccess, Duration: 100 * time.Millisecond},
		{ID: "op2", Title: "Trace", Status: OperationFailed, Error: "timeout"},
	}
	result := a.operationsHistoryMarkdown()
	if result == "" {
		t.Error("expected non-empty markdown")
	}
}

func TestRefreshOperationSelectOptions_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.refreshOperationSelectOptions()
}

func TestRefreshOperationSelectOptions_NilSelect(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.refreshOperationSelectOptions()
}

func TestRefreshOperationSelectOptions_WithSelect(t *testing.T) {
	a := &App{}
	a.operationsSelect = widget.NewSelect([]string{}, nil)
	a.operationsHistory = []Operation{
		{ID: "op1", Title: "Ping", Status: OperationSuccess},
	}
	a.refreshOperationSelectOptions()
	if len(a.operationsSelect.Options) != 1 {
		t.Errorf("expected 1 option, got %d", len(a.operationsSelect.Options))
	}
}

func TestRefreshOperationActionsState_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.refreshOperationActionsState()
}

func TestRefreshOperationActionsState_EmptyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.refreshOperationActionsState()
}

func TestRefreshOperationActionsState_NilButtons(t *testing.T) {
	a := &App{}
	a.operations = NewOperationsManager()
	// Не должен паниковать
	a.refreshOperationActionsState()
}

func TestRefreshOperationActionsState_WithButtons(t *testing.T) {
	a := &App{}
	a.operations = NewOperationsManager()
	a.operationsRetryBtn = widget.NewButton("Retry", nil)
	a.operationsCancelBtn = widget.NewButton("Cancel", nil)
	a.refreshOperationActionsState()
	if !a.operationsRetryBtn.Disabled() {
		t.Error("expected retry disabled for no selected operation")
	}
}

func TestRunToolOperation_EmptyApp(t *testing.T) {
	a := &App{}
	// Без toolsOutput не должен паниковать
	a.runToolOperation("test", "started", func(ctx context.Context) (string, error) {
		return "done", nil
	})
}

func TestRunToolOperation_WithOutput(t *testing.T) {
	a := &App{}
	a.toolsOutput = widget.NewRichText()
	a.runToolOperation("test", "started", func(ctx context.Context) (string, error) {
		return "done", nil
	})
}
