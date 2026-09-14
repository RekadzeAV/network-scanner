package gui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// --- Методы Tools UI ---

// withToolHost возвращает хост из поля ввода инструментов.
func (a *App) withToolHost() (string, bool) {
	if a == nil || a.toolsHostEntry == nil {
		return "", false
	}
	host := strings.TrimSpace(a.toolsHostEntry.Text)
	if host == "" {
		dialog.ShowInformation("Инструменты", "Введите хост или IP", a.myWindow)
		return "", false
	}
	return host, true
}

// setToolsOutputMarkdown устанавливает markdown-вывод в инструменты.
func (a *App) setToolsOutputMarkdown(md string) {
	if a == nil || a.toolsOutput == nil {
		return
	}
	a.toolsOutput.ParseMarkdown(md)
	a.toolsOutput.Refresh()
}

// setToolsButtonsEnabled включает/отключает кнопки инструментов.
func (a *App) setToolsButtonsEnabled(enabled bool) {
	for _, b := range []*widget.Button{
		a.toolsPingBtn,
		a.toolsTraceBtn,
		a.toolsDNSBtn,
		a.toolsWhoisBtn,
		a.toolsWiFiBtn,
		a.toolsWOLBtn,
		a.toolsAuditBtn,
		a.toolsRiskBtn,
		a.toolsDeviceStatusBtn,
		a.toolsDeviceRebootBtn,
	} {
		if b == nil {
			continue
		}
		if enabled {
			b.Enable()
		} else {
			b.Disable()
		}
	}
}

// runToolOperation запускает операцию инструмента через Operations Manager.
func (a *App) runToolOperation(title string, startedMessage string, task func(context.Context) (string, error)) {
	if a == nil {
		return
	}
	if strings.TrimSpace(startedMessage) == "" {
		startedMessage = "Выполняется операция..."
	}
	a.setToolsButtonsEnabled(false)
	a.setToolsOutputMarkdown(startedMessage)

	run := func(ctx context.Context) error {
		markdown, err := task(ctx)
		fyne.Do(func() {
			a.setToolsButtonsEnabled(true)
			if err != nil {
				a.setToolsOutputMarkdown(markdown)
				return
			}
			a.setToolsOutputMarkdown(markdown)
		})
		return err
	}

	if a.operations == nil {
		go func() {
			_ = run(context.Background())
		}()
		return
	}
	a.operations.Run(OperationTypeTool, title, run)
}

// --- Методы Operations Center ---

// startOperationsWatcher запускает наблюдатель за операциями.
func (a *App) startOperationsWatcher() {
	if a == nil || a.operations == nil {
		return
	}
	events := a.operations.Subscribe(32)
	go func() {
		for ev := range events {
			op := ev.Operation
			fyne.Do(func() {
				a.pushOperationHistory(op)
			})
		}
	}()
}

// pushOperationHistory добавляет или обновляет операцию в истории.
func (a *App) pushOperationHistory(op Operation) {
	if a == nil || a.operationsOutput == nil {
		return
	}
	updated := false
	for i := range a.operationsHistory {
		if a.operationsHistory[i].ID == op.ID {
			a.operationsHistory[i] = op
			updated = true
			break
		}
	}
	if !updated {
		a.operationsHistory = append([]Operation{op}, a.operationsHistory...)
	}
	if len(a.operationsHistory) > 20 {
		a.operationsHistory = a.operationsHistory[:20]
	}
	a.refreshOperationSelectOptions()
	a.operationsOutput.ParseMarkdown(a.operationsHistoryMarkdown())
	a.operationsOutput.Refresh()
	a.refreshOperationActionsState()
}

// operationsHistoryMarkdown формирует markdown-отчёт истории операций.
func (a *App) operationsHistoryMarkdown() string {
	var sb strings.Builder
	sb.WriteString("### Operations Center\n\n")
	if len(a.operationsHistory) == 0 {
		sb.WriteString("История операций пуста.")
		return sb.String()
	}
	for _, op := range a.operationsHistory {
		dur := "-"
		if op.Duration > 0 {
			dur = op.Duration.Round(time.Millisecond).String()
		}
		sb.WriteString(fmt.Sprintf("- `%s` **%s** `%s` (%s)\n", op.ID, strings.ToUpper(string(op.Status)), op.Title, dur))
		if strings.TrimSpace(op.Error) != "" {
			sb.WriteString(fmt.Sprintf("  - error: %s\n", strings.TrimSpace(op.Error)))
		}
	}
	return sb.String()
}

// refreshOperationSelectOptions обновляет список операций в селекте.
func (a *App) refreshOperationSelectOptions() {
	if a == nil || a.operationsSelect == nil {
		return
	}
	options := make([]string, 0, len(a.operationsHistory))
	a.operationsSelectMap = make(map[string]string, len(a.operationsHistory))
	selectedLabel := ""
	for _, op := range a.operationsHistory {
		label := fmt.Sprintf("%s | %s | %s", op.ID, strings.ToUpper(string(op.Status)), strings.TrimSpace(op.Title))
		options = append(options, label)
		a.operationsSelectMap[label] = op.ID
		if strings.TrimSpace(op.ID) == strings.TrimSpace(a.selectedOperationID) {
			selectedLabel = label
		}
	}
	a.operationsSelect.Options = options
	a.operationsSelect.Refresh()
	if selectedLabel != "" {
		a.operationsSelect.SetSelected(selectedLabel)
		return
	}
	if len(options) > 0 {
		a.operationsSelect.SetSelected(options[0])
		return
	}
	a.selectedOperationID = ""
}

// refreshOperationActionsState обновляет состояние кнопок Retry/Cancel.
func (a *App) refreshOperationActionsState() {
	if a == nil {
		return
	}
	if a.operationsRetryBtn == nil || a.operationsCancelBtn == nil || a.operations == nil {
		return
	}
	a.operationsRetryBtn.Disable()
	a.operationsCancelBtn.Disable()
	id := strings.TrimSpace(a.selectedOperationID)
	if id == "" {
		return
	}
	op, ok := a.operations.Get(id)
	if !ok {
		return
	}
	if op.CanRetry && (op.Status == OperationFailed || op.Status == OperationCanceled) {
		a.operationsRetryBtn.Enable()
	}
	if op.CanCancel && (op.Status == OperationQueued || op.Status == OperationRunning) {
		a.operationsCancelBtn.Enable()
	}
}
