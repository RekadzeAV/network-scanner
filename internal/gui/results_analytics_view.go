package gui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

func (a *App) buildResultsAnalyticsView(data []scanner.Result) fyne.CanvasObject {
	protocols, deviceTypes := collectAnalytics(data)
	normalizedTypes := normalizeDeviceTypes(deviceTypes)
	if strings.EqualFold(strings.TrimSpace(a.resultsMode), "Карточки") {
		return container.NewHBox(
			a.buildPieChart("Протоколы", protocols),
			a.buildPieChart("Типы устройств", normalizedTypes),
		)
	}
	return widget.NewCard("Аналитика", "", widget.NewRichTextFromMarkdown(buildAnalyticsMarkdown(protocols, normalizedTypes)))
}

func buildAnalyticsMarkdown(protocols map[string]int, deviceTypes map[string]int) string {
	var sb strings.Builder
	sb.WriteString("### Аналитика\n\n")
	sb.WriteString("#### Протоколы\n\n")
	if len(protocols) == 0 {
		sb.WriteString("- Нет данных\n\n")
	} else {
		keys := make([]string, 0, len(protocols))
		for k := range protocols {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("- `%s`: `%d`\n", strings.TrimSpace(k), protocols[k]))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("#### Типы устройств\n\n")
	if len(deviceTypes) == 0 {
		sb.WriteString("- Нет данных\n")
	} else {
		keys := make([]string, 0, len(deviceTypes))
		for k := range deviceTypes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("- `%s`: `%d`\n", strings.TrimSpace(k), deviceTypes[k]))
		}
	}
	return sb.String()
}
