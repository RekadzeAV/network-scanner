package gui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/audit"
	"network-scanner/internal/cve"
	"network-scanner/internal/report"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
)

func (a *App) buildSecurityDashboardView(data []scanner.Result) fyne.CanvasObject {
	portFindings := audit.EvaluateOpenPorts(data)
	db, dbErr := risksignature.LoadDefault()
	signatureFindings := make([]risksignature.Finding, 0)
	if dbErr == nil {
		signatureFindings = risksignature.Evaluate(data, db)
	}

	severityCounts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}
	for _, f := range portFindings {
		key, _ := audit.NormalizeSeverity(strings.TrimSpace(f.Severity))
		if _, ok := severityCounts[key]; ok {
			severityCounts[key]++
		}
	}
	for _, f := range signatureFindings {
		key, _ := audit.NormalizeSeverity(strings.TrimSpace(f.Severity))
		if _, ok := severityCounts[key]; ok {
			severityCounts[key]++
		}
	}

	var summary strings.Builder
	summary.WriteString("### Security Dashboard\n\n")
	summary.WriteString(fmt.Sprintf("- Devices in scope: `%d`\n", len(data)))
	summary.WriteString(fmt.Sprintf("- Port audit findings: `%d`\n", len(portFindings)))
	summary.WriteString(fmt.Sprintf("- Risk signatures findings: `%d`\n", len(signatureFindings)))
	if dbErr != nil {
		summary.WriteString(fmt.Sprintf("- Risk DB status: error `%v`\n", dbErr))
	} else {
		summary.WriteString(fmt.Sprintf("- Risk DB version: `%s`\n", strings.TrimSpace(db.Version)))
	}
	summary.WriteString("\n#### Severity summary\n\n")
	summary.WriteString(fmt.Sprintf("- Critical: `%d`\n", severityCounts["critical"]))
	summary.WriteString(fmt.Sprintf("- High: `%d`\n", severityCounts["high"]))
	summary.WriteString(fmt.Sprintf("- Medium: `%d`\n", severityCounts["medium"]))
	summary.WriteString(fmt.Sprintf("- Low: `%d`\n", severityCounts["low"]))

	exportBtn := widget.NewButton("Export security report (HTML)", func() {
		a.exportSecurityDashboardReport(data, signatureFindings)
	})
	table := a.buildSecurityFindingsTable(portFindings, signatureFindings)
	return container.NewBorder(
		container.NewVBox(widget.NewRichTextFromMarkdown(summary.String()), exportBtn),
		nil, nil, nil,
		table,
	)
}

func (a *App) buildSecurityFindingsTable(portFindings []audit.Finding, signatureFindings []risksignature.Finding) fyne.CanvasObject {
	type row struct {
		source   string
		severity string
		host     string
		title    string
	}
	rows := make([]row, 0, len(portFindings)+len(signatureFindings))
	for _, f := range portFindings {
		rows = append(rows, row{
			source:   "port-audit",
			severity: strings.ToLower(strings.TrimSpace(f.Severity)),
			host:     strings.TrimSpace(f.Host),
			title:    strings.TrimSpace(f.Title),
		})
	}
	for _, f := range signatureFindings {
		rows = append(rows, row{
			source:   "risk-signature",
			severity: strings.ToLower(strings.TrimSpace(f.Severity)),
			host:     strings.TrimSpace(f.HostIP),
			title:    strings.TrimSpace(f.Title),
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].severity != rows[j].severity {
			return rows[i].severity < rows[j].severity
		}
		if rows[i].source != rows[j].source {
			return rows[i].source < rows[j].source
		}
		return rows[i].host < rows[j].host
	})
	if len(rows) == 0 {
		return container.NewCenter(widget.NewLabel("Security findings отсутствуют для текущего скоупа."))
	}

	tableRows := len(rows) + 1
	table := widget.NewTable(
		func() (int, int) { return tableRows, 4 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			l := obj.(*widget.Label)
			if id.Row == 0 {
				headers := []string{"Source", "Severity", "Host", "Title"}
				l.TextStyle = fyne.TextStyle{Bold: true}
				l.SetText(headers[id.Col])
				return
			}
			r := rows[id.Row-1]
			l.TextStyle = fyne.TextStyle{}
			switch id.Col {
			case 0:
				l.SetText(r.source)
			case 1:
				l.SetText(r.severity)
			case 2:
				l.SetText(r.host)
			case 3:
				l.SetText(r.title)
			}
		},
	)
	table.SetColumnWidth(0, 130)
	table.SetColumnWidth(1, 90)
	table.SetColumnWidth(2, 120)
	table.SetColumnWidth(3, 420)
	return table
}

func (a *App) exportSecurityDashboardReport(results []scanner.Result, risks []risksignature.Finding) {
	if a == nil || a.myWindow == nil {
		return
	}
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.myWindow)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		htmlData, renderErr := report.RenderSecurityHTMLWithRiskOptions(
			results,
			[]cve.Match{},
			risks,
			time.Now(),
			report.Options{
				RedactSensitive: true,
				GenerationMode:  "gui-security-dashboard",
				PolicyVersion:   "v1",
			},
		)
		if renderErr != nil {
			dialog.ShowError(renderErr, a.myWindow)
			return
		}
		if _, writeErr := writer.Write(htmlData); writeErr != nil {
			dialog.ShowError(writeErr, a.myWindow)
			return
		}
		dialog.ShowInformation("Готово", fmt.Sprintf("Security report сохранен: %s", filepath.Base(writer.URI().Path())), a.myWindow)
	}, a.myWindow)
}
