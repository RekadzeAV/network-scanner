package gui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/inventory"
	"network-scanner/internal/scanner"
)

func (a *App) buildInventoryDashboardView() fyne.CanvasObject {
	if a == nil {
		return widget.NewLabel("Inventory: app is not initialized")
	}
	if a.inventoryStatusLabel == nil {
		a.inventoryStatusLabel = widget.NewLabel("")
	}
	if a.inventoryScanASelect == nil || a.inventoryScanBSelect == nil {
		return widget.NewLabel("Inventory controls are not initialized")
	}
	diffText := widget.NewRichTextFromMarkdown(a.inventoryDiffMarkdown())
	diffText.Wrapping = fyne.TextWrapWord
	return container.NewVBox(
		widget.NewLabel("Инвентаризация сети (снапшоты и сравнение):"),
		a.inventoryStatusLabel,
		container.NewGridWithColumns(2, widget.NewLabel("Snapshot A:"), a.inventoryScanASelect),
		container.NewGridWithColumns(2, widget.NewLabel("Snapshot B:"), a.inventoryScanBSelect),
		widget.NewSeparator(),
		diffText,
	)
}

func (a *App) refreshInventorySnapshots() {
	if a == nil || a.inventoryDBEntry == nil {
		return
	}
	dbPath := strings.TrimSpace(a.inventoryDBEntry.Text)
	if dbPath == "" {
		dbPath = filepath.Join("inventory", "network_inventory.db")
	}
	store, err := inventory.Open(dbPath)
	if err != nil {
		if a.inventoryStatusLabel != nil {
			a.inventoryStatusLabel.SetText(fmt.Sprintf("Инвентаризация: ошибка открытия БД: %v", err))
		}
		return
	}
	defer store.Close()

	snapshots, err := store.ListSnapshots(200)
	if err != nil {
		if a.inventoryStatusLabel != nil {
			a.inventoryStatusLabel.SetText(fmt.Sprintf("Инвентаризация: ошибка чтения снапшотов: %v", err))
		}
		return
	}
	a.inventorySnapshots = snapshots
	options := make([]string, 0, len(snapshots))
	for _, s := range snapshots {
		options = append(options, snapshotOptionLabel(s))
	}
	a.inventoryScanASelect.Options = options
	a.inventoryScanBSelect.Options = options
	if len(options) >= 1 && strings.TrimSpace(a.inventoryScanASelect.Selected) == "" {
		a.inventoryScanASelect.SetSelected(options[0])
	}
	if len(options) >= 2 && strings.TrimSpace(a.inventoryScanBSelect.Selected) == "" {
		a.inventoryScanBSelect.SetSelected(options[1])
	} else if len(options) == 1 && strings.TrimSpace(a.inventoryScanBSelect.Selected) == "" {
		a.inventoryScanBSelect.SetSelected(options[0])
	}
	if a.inventoryStatusLabel != nil {
		a.inventoryStatusLabel.SetText(fmt.Sprintf("Инвентаризация: загружено снапшотов %d", len(snapshots)))
	}
}

func (a *App) inventoryDiffMarkdown() string {
	if a == nil || a.inventoryScanASelect == nil || a.inventoryScanBSelect == nil {
		return "### Inventory Diff\n\nКонтролы не инициализированы."
	}
	idA := parseSnapshotID(strings.TrimSpace(a.inventoryScanASelect.Selected))
	idB := parseSnapshotID(strings.TrimSpace(a.inventoryScanBSelect.Selected))
	if idA == "" || idB == "" {
		return "### Inventory Diff\n\nВыберите два снапшота для сравнения."
	}
	dbPath := filepath.Join("inventory", "network_inventory.db")
	if a.inventoryDBEntry != nil && strings.TrimSpace(a.inventoryDBEntry.Text) != "" {
		dbPath = strings.TrimSpace(a.inventoryDBEntry.Text)
	}
	store, err := inventory.Open(dbPath)
	if err != nil {
		return fmt.Sprintf("### Inventory Diff\n\nОшибка открытия БД: `%v`", err)
	}
	defer store.Close()
	diff, err := store.Diff(idA, idB)
	if err != nil {
		return fmt.Sprintf("### Inventory Diff\n\nОшибка сравнения: `%v`", err)
	}
	var sb strings.Builder
	sb.WriteString("### Inventory Diff\n\n")
	sb.WriteString(fmt.Sprintf("- Snapshot A: `%s`\n", diff.ScanIDA))
	sb.WriteString(fmt.Sprintf("- Snapshot B: `%s`\n", diff.ScanIDB))
	sb.WriteString(fmt.Sprintf("- New devices: `%d`\n", len(diff.New)))
	sb.WriteString(fmt.Sprintf("- Missing devices: `%d`\n", len(diff.Missing)))
	sb.WriteString(fmt.Sprintf("- Changed devices: `%d`\n", len(diff.Changed)))
	if len(diff.New) > 0 {
		sb.WriteString("\n#### New\n")
		for _, h := range diff.New {
			sb.WriteString(fmt.Sprintf("- `%s` (%s)\n", nullDash(h.IP), nullDash(h.Hostname)))
		}
	}
	if len(diff.Missing) > 0 {
		sb.WriteString("\n#### Missing\n")
		for _, h := range diff.Missing {
			sb.WriteString(fmt.Sprintf("- `%s` (%s)\n", nullDash(h.IP), nullDash(h.Hostname)))
		}
	}
	if len(diff.Changed) > 0 {
		sb.WriteString("\n#### Changed\n")
		for _, ch := range diff.Changed {
			sb.WriteString(fmt.Sprintf("- `%s`: `%s`\n", ch.Key, strings.Join(ch.ChangedField, ", ")))
		}
	}
	return sb.String()
}

func (a *App) saveInventorySnapshotFromResults(results []scanner.Result) {
	if a == nil || a.inventoryAutoSaveCheck == nil || !a.inventoryAutoSaveCheck.Checked {
		return
	}
	dbPath := filepath.Join("inventory", "network_inventory.db")
	if a.inventoryDBEntry != nil && strings.TrimSpace(a.inventoryDBEntry.Text) != "" {
		dbPath = strings.TrimSpace(a.inventoryDBEntry.Text)
	}
	store, err := inventory.Open(dbPath)
	if err != nil {
		if a.inventoryStatusLabel != nil {
			a.inventoryStatusLabel.SetText(fmt.Sprintf("Инвентаризация: не удалось открыть БД: %v", err))
		}
		return
	}
	defer store.Close()
	scanID := fmt.Sprintf("scan-%s", time.Now().UTC().Format("20060102T150405Z"))
	if err := store.SaveSnapshot(scanID, time.Now().UTC(), results); err != nil {
		if a.inventoryStatusLabel != nil {
			a.inventoryStatusLabel.SetText(fmt.Sprintf("Инвентаризация: ошибка сохранения снапшота: %v", err))
		}
		return
	}
	a.refreshInventorySnapshots()
	if a.inventoryStatusLabel != nil {
		a.inventoryStatusLabel.SetText(fmt.Sprintf("Инвентаризация: сохранен снапшот %s", scanID))
	}
}

func snapshotOptionLabel(s inventory.Snapshot) string {
	if !s.Timestamp.IsZero() {
		return fmt.Sprintf("%s (%s)", s.ID, s.Timestamp.Local().Format("2006-01-02 15:04:05"))
	}
	return s.ID
}

func parseSnapshotID(option string) string {
	if option == "" {
		return ""
	}
	idx := strings.Index(option, " (")
	if idx <= 0 {
		return option
	}
	return strings.TrimSpace(option[:idx])
}
