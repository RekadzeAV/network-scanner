package controller

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"network-scanner/internal/scanner"
	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// TopologyUI предоставляет доступ к виджетам топологии.
type TopologyUI struct {
	SNMPCommEntry     *widget.Entry
	SNMPTimeoutEnt    *widget.Entry
	BuildTopoBtn      *widget.Button
	StopTopoBtn       *widget.Button
	SaveTopoBtn       *widget.Button
	CopyPerfBtn       *widget.Button
	SavePerfBtn       *widget.Button
	TopoText          *widget.RichText
	TopoStatus        *widget.Label
	SNMPStageLabel    *widget.Label
	SNMPProgress      *widget.ProgressBar
	TopoSearchEntry   *widget.Entry
	TopoTypeFilterSel *widget.Select
	TopoConfFilterSel *widget.Select
	TopoResetMapBtn   *widget.Button
	TopoGraphStatus   *widget.Label
	TopoImage         *canvas.Image
	ZoomSelect        *widget.Select
	RefreshPreviewBtn *widget.Button
	OpenPreviewBtn    *widget.Button
}

// TopologyController управляет построением топологии.
type TopologyController struct {
	app         fyne.App
	ui          *TopologyUI
	cancel      context.CancelFunc
	lastTopo    *topology.Topology
	lastReport  *snmpcollector.CollectReport
	lastMetrics topologyBuildMetrics
}

type topologyBuildMetrics struct {
	snmpDuration  time.Duration
	buildDuration time.Duration
	totalDuration time.Duration
}

// NewTopologyController создает контроллер.
func NewTopologyController(app fyne.App, ui *TopologyUI) *TopologyController {
	return &TopologyController{app: app, ui: ui}
}

// BuildTopology запускает построение топологии.
func (c *TopologyController) BuildTopology(results []scanner.Result, window fyne.Window) {
	if len(results) == 0 {
		dialog.ShowInformation("Информация", "Сначала выполните сканирование", window)
		return
	}
	topologyStartedAt := time.Now()
	c.applyTopologyRunStart()

	timeoutSec := 2
	if c.ui.SNMPTimeoutEnt != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(c.ui.SNMPTimeoutEnt.Text)); err == nil && v > 0 {
			timeoutSec = v
		}
	}
	communities := splitCommaValues(c.ui.SNMPCommEntry.Text)
	snmpStartedAt := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	go func() {
		snmpPhaseStartedAt := time.Now()
		snmpData, report, err := snmpcollector.CollectWithReportProgressContext(ctx, results, communities, timeoutSec, func(current int, total int, ip string, message string) {
			etaText := ""
			progressValue := 0.0
			if total > 0 && current > 0 && current < total {
				elapsed := time.Since(snmpStartedAt)
				remainingItems := total - current
				eta := time.Duration(float64(elapsed) * (float64(remainingItems) / float64(current)))
				etaText = fmt.Sprintf(", ETA ~ %s", formatDurationMMSS(eta))
			}
			if total > 0 {
				progressValue = float64(current) / float64(total)
				if progressValue > 1 {
					progressValue = 1
				}
			}
			status := fmt.Sprintf("SNMP: %d/%d (%s)%s", current, total, message, etaText)
			if strings.TrimSpace(ip) != "" {
				status = fmt.Sprintf("%s, %s", status, ip)
			}
			fyne.Do(func() {
				c.applyTopologyProgress(status, progressValue)
			})
		})
		if err != nil {
			fyne.Do(func() {
				if err == context.Canceled {
					c.applyTopologyCanceled()
					return
				}
				dialog.ShowError(fmt.Errorf("ошибка SNMP опроса: %v", err), window)
				c.applyTopologyFailure("snmp")
			})
			return
		}
		snmpDuration := time.Since(snmpPhaseStartedAt)
		buildPhaseStartedAt := time.Now()
		topo, err := topology.BuildTopologyWithOptions(results, snmpData, topology.BuildOptions{
			PartialSNMPKeys: partialSNMPKeysFromReport(report),
		})
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("ошибка построения топологии: %v", err), window)
				c.applyTopologyFailure("build")
			})
			return
		}
		buildDuration := time.Since(buildPhaseStartedAt)
		metrics := topologyBuildMetrics{
			snmpDuration:  snmpDuration,
			buildDuration: buildDuration,
			totalDuration: time.Since(topologyStartedAt),
		}
		c.lastTopo = topo
		c.lastReport = report
		c.lastMetrics = metrics
		c.renderTopologyImagePreview(topo, window)
		fyne.Do(func() {
			c.applyTopologySuccess(
				topologySuccessStatus(topo, report),
				formatTopologyPreview(topo, report, metrics),
				topo,
				report,
				metrics,
			)
		})
	}()
}

func (c *TopologyController) applyTopologyRunStart() {
	if c.ui == nil {
		return
	}
	c.ui.BuildTopoBtn.Disable()
	c.ui.TopoStatus.SetText("Построение топологии...")
	c.ui.SNMPStageLabel.Show()
	c.ui.SNMPProgress.Show()
}

func (c *TopologyController) applyTopologyProgress(status string, progress float64) {
	if c.ui == nil {
		return
	}
	c.ui.TopoStatus.SetText(status)
	c.ui.SNMPProgress.SetValue(progress)
	c.ui.SNMPStageLabel.Refresh()
}

func (c *TopologyController) applyTopologyCanceled() {
	if c.ui == nil {
		return
	}
	c.ui.TopoStatus.SetText("Построение отменено")
	c.ui.BuildTopoBtn.Enable()
	c.ui.StopTopoBtn.Disable()
	c.ui.SNMPStageLabel.Hide()
	c.ui.SNMPProgress.Hide()
}

func (c *TopologyController) applyTopologyFailure(phase string) {
	if c.ui == nil {
		return
	}
	c.ui.TopoStatus.SetText("Ошибка на этапе: " + phase)
	c.ui.BuildTopoBtn.Enable()
	c.ui.StopTopoBtn.Disable()
	c.ui.SNMPStageLabel.Hide()
	c.ui.SNMPProgress.Hide()
}

func (c *TopologyController) applyTopologySuccess(status, preview string, topo *topology.Topology, report *snmpcollector.CollectReport, metrics topologyBuildMetrics) {
	if c.ui == nil {
		return
	}
	c.ui.TopoText.ParseMarkdown("## Топология сети\n\n" + preview)
	c.ui.TopoText.Refresh()
	c.ui.TopoStatus.SetText(status)
	c.ui.BuildTopoBtn.Enable()
	c.ui.StopTopoBtn.Disable()
	c.ui.SNMPStageLabel.Hide()
	c.ui.SNMPProgress.Hide()
}

// StopTopologyBuild останавливает построение.
func (c *TopologyController) StopTopologyBuild() {
	if c.cancel == nil {
		return
	}
	c.cancel()
}

// SaveTopology сохраняет топологию в файл.
func (c *TopologyController) SaveTopology(topo *topology.Topology, window fyne.Window) {
	if topo == nil {
		dialog.ShowInformation("Информация", "Сначала постройте топологию", window)
		return
	}
	if err := topo.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("топология не прошла валидацию перед сохранением: %v", err), window)
		return
	}
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if writer == nil {
			return
		}
		path := writer.URI().Path()
		_ = writer.Close()

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".json":
			err = topo.SaveJSON(path)
		case ".graphml", ".xml":
			err = topo.SaveGraphML(path)
		case ".png":
			err = topo.RenderWithGraphviz("png", path)
		case ".svg":
			err = topo.RenderWithGraphviz("svg", path)
		default:
			err = fmt.Errorf("поддерживаемые форматы: .json, .graphml, .png, .svg")
		}

		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		dialog.ShowInformation("Успех", fmt.Sprintf("Топология сохранена (узлов: %d, связей: %d)", len(topo.Devices), len(topo.Links)), window)
	}, window)
}

// CopyPerformanceReport копирует отчет о производительности.
func (c *TopologyController) CopyPerformanceReport(window fyne.Window) {
	reportText := c.buildPerformanceReportText()
	if strings.TrimSpace(reportText) == "" {
		dialog.ShowInformation("Информация", "Отчет производительности пока недоступен", window)
		return
	}
	window.Clipboard().SetContent(reportText)
	dialog.ShowInformation("Готово", "Отчет производительности скопирован в буфер обмена", window)
}

// SavePerformanceReport сохраняет отчет о производительности.
func (c *TopologyController) SavePerformanceReport(window fyne.Window) {
	reportText := c.buildPerformanceReportText()
	if strings.TrimSpace(reportText) == "" {
		dialog.ShowInformation("Информация", "Отчет производительности пока недоступен", window)
		return
	}

	defaultFileName := fmt.Sprintf("topology-performance-%s.txt", time.Now().Format("2006-01-02-150405"))
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if writer == nil {
			return
		}
		targetPath := writer.URI().Path()
		normalizedPath := targetPath
		if strings.ToLower(filepath.Ext(normalizedPath)) != ".txt" {
			normalizedPath += ".txt"
		}

		if normalizedPath == targetPath {
			defer writer.Close()
			if _, writeErr := writer.Write([]byte(reportText)); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении отчета: %v", writeErr), window)
				return
			}
		} else {
			_ = writer.Close()
			if writeErr := os.WriteFile(normalizedPath, []byte(reportText), 0644); writeErr != nil {
				dialog.ShowError(fmt.Errorf("ошибка при сохранении отчета: %v", writeErr), window)
				return
			}
		}
		dialog.ShowInformation("Успех", fmt.Sprintf("Отчет производительности сохранен:\n%s", normalizedPath), window)
	}, window)
	saveDialog.SetFileName(defaultFileName)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
	saveDialog.Show()
}

// RefreshTopologyPreview обновляет графическое превью.
func (c *TopologyController) RefreshTopologyPreview(topo *topology.Topology, window fyne.Window) {
	if topo == nil {
		dialog.ShowInformation("Информация", "Сначала постройте топологию", window)
		return
	}
	if c.ui == nil {
		return
	}
	c.ui.TopoStatus.SetText("Обновление графического превью...")
	c.ui.TopoStatus.Refresh()
	go func() {
		c.renderTopologyImagePreview(topo, window)
		fyne.Do(func() {
			if c.ui != nil {
				c.ui.TopoStatus.SetText(fmt.Sprintf("Топология построена: устройств %d, связей %d", len(topo.Devices), len(topo.Links)))
				c.ui.TopoStatus.Refresh()
			}
		})
	}()
}

// ApplyTopologyZoom применяет масштаб к изображению топологии.
func (c *TopologyController) ApplyTopologyZoom(mode string, window fyne.Window) {
	if c.ui == nil || c.ui.TopoImage == nil {
		return
	}
	canvasSize := fyne.NewSize(1200, 700)
	if window != nil && window.Canvas() != nil {
		if s := window.Canvas().Size(); s.Width > 0 && s.Height > 0 {
			canvasSize = s
		}
	}
	baseWidth := float32(math.Max(900, float64(canvasSize.Width*0.7)))
	baseHeight := float32(math.Max(500, float64(canvasSize.Height*0.62)))
	switch mode {
	case "200%":
		c.ui.TopoImage.FillMode = canvas.ImageFillOriginal
		c.ui.TopoImage.SetMinSize(fyne.NewSize(baseWidth*2.0, baseHeight*2.0))
	case "150%":
		c.ui.TopoImage.FillMode = canvas.ImageFillOriginal
		c.ui.TopoImage.SetMinSize(fyne.NewSize(baseWidth*1.5, baseHeight*1.5))
	case "100%":
		c.ui.TopoImage.FillMode = canvas.ImageFillOriginal
		c.ui.TopoImage.SetMinSize(fyne.NewSize(baseWidth, baseHeight))
	default:
		c.ui.TopoImage.FillMode = canvas.ImageFillContain
		c.ui.TopoImage.SetMinSize(fyne.NewSize(0, 260))
	}
	c.ui.TopoImage.Refresh()
	c.ui.TopoImage.Refresh()
}

// OpenPreviewExternal открывает превью во внешнем окне.
func (c *TopologyController) OpenPreviewExternal(previewPath string, window fyne.Window) {
	if strings.TrimSpace(previewPath) == "" {
		dialog.ShowInformation("Информация", "Сначала постройте превью топологии", window)
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", "", previewPath)
	case "darwin":
		cmd = exec.Command("open", previewPath)
	default:
		cmd = exec.Command("xdg-open", previewPath)
	}
	if err := cmd.Start(); err != nil {
		dialog.ShowError(fmt.Errorf("не удалось открыть файл: %v", err), window)
		return
	}
}

func (c *TopologyController) buildPerformanceReportText() string {
	if c.lastTopo == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Отчет производительности topology build\n")
	sb.WriteString(fmt.Sprintf("Устройств: %d\n", len(c.lastTopo.Devices)))
	sb.WriteString(fmt.Sprintf("Связей: %d\n", len(c.lastTopo.Links)))
	if c.lastMetrics.snmpDuration > 0 {
		sb.WriteString(fmt.Sprintf("SNMP сбор: %s\n", c.lastMetrics.snmpDuration.Round(time.Millisecond).String()))
	}
	if c.lastMetrics.buildDuration > 0 {
		sb.WriteString(fmt.Sprintf("Построение графа: %s\n", c.lastMetrics.buildDuration.Round(time.Millisecond).String()))
	}
	if c.lastMetrics.totalDuration > 0 {
		sb.WriteString(fmt.Sprintf("Общее время: %s\n", c.lastMetrics.totalDuration.Round(time.Millisecond).String()))
	}
	if c.lastReport != nil {
		sb.WriteString(fmt.Sprintf("SNMP целей: %d\n", c.lastReport.TotalSNMPTargets))
		sb.WriteString(fmt.Sprintf("SNMP ok: %d\n", c.lastReport.Connected))
		sb.WriteString(fmt.Sprintf("SNMP partial: %d\n", c.lastReport.Partial))
		sb.WriteString(fmt.Sprintf("SNMP failed: %d\n", c.lastReport.Failed))
	}
	return sb.String()
}

func (c *TopologyController) renderTopologyImagePreview(topo *topology.Topology, window fyne.Window) {
	if topo == nil {
		return
	}
	tmp, err := os.CreateTemp("", "network-topology-preview-*.png")
	if err != nil {
		fyne.Do(func() {
			if c.ui != nil {
				c.ui.TopoStatus.SetText("Не удалось создать временный файл для превью")
				c.ui.TopoStatus.Refresh()
			}
		})
		return
	}
	previewPath := tmp.Name()
	_ = tmp.Close()

	if err = topo.RenderWithGraphviz("png", previewPath); err != nil {
		_ = os.Remove(previewPath)
		fyne.Do(func() {
			if c.ui != nil {
				c.ui.TopoStatus.SetText("Графическое превью недоступно (установите Graphviz/dot)")
				c.ui.TopoStatus.Refresh()
			}
		})
		return
	}

	fyne.Do(func() {
		if c.ui != nil && c.ui.TopoImage != nil {
			img := canvas.NewImageFromFile(previewPath)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(0, 260))
			c.ui.TopoImage = img
			c.ApplyTopologyZoom("Fit", window)
		}
	})
}

func splitCommaValues(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"public"}
	}
	return out
}

func partialSNMPKeysFromReport(report *snmpcollector.CollectReport) map[string]struct{} {
	if report == nil {
		return nil
	}
	out := make(map[string]struct{})
	for _, f := range report.Failures {
		if f.Kind != snmpcollector.FailureQuery {
			continue
		}
		ip := strings.TrimSpace(strings.ToLower(f.IP))
		if ip != "" {
			out["ip:"+ip] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func topologySuccessStatus(topo *topology.Topology, report *snmpcollector.CollectReport) string {
	if topo == nil {
		return "Топология не построена"
	}
	return fmt.Sprintf("Топология построена: устройств %d, связей %d", len(topo.Devices), len(topo.Links))
}

func formatTopologyPreview(topo *topology.Topology, report *snmpcollector.CollectReport, metrics topologyBuildMetrics) string {
	if topo == nil {
		return "## Топология сети\n\nНет данных для отображения."
	}
	var sb strings.Builder
	sb.WriteString("## Топология сети\n\n")
	if metrics.totalDuration > 0 {
		sb.WriteString("### Время этапов\n\n")
		if metrics.snmpDuration > 0 {
			sb.WriteString(fmt.Sprintf("- SNMP сбор: `%s`\n", metrics.snmpDuration.Round(time.Millisecond).String()))
		}
		if metrics.buildDuration > 0 {
			sb.WriteString(fmt.Sprintf("- Построение графа: `%s`\n", metrics.buildDuration.Round(time.Millisecond).String()))
		}
		sb.WriteString(fmt.Sprintf("- Общее время: `%s`\n\n", metrics.totalDuration.Round(time.Millisecond).String()))
	}
	if report != nil {
		sb.WriteString("### SNMP отчет\n\n")
		sb.WriteString(fmt.Sprintf("- Целей для SNMP: %d\n", report.TotalSNMPTargets))
		sb.WriteString(fmt.Sprintf("- Успешных подключений: %d\n", report.Connected))
		sb.WriteString(fmt.Sprintf("- Частичных опросов: %d\n", report.Partial))
		sb.WriteString(fmt.Sprintf("- Полных отказов: %d\n\n", report.Failed))
	}
	sb.WriteString(fmt.Sprintf("**Устройств:** %d\n\n", len(topo.Devices)))
	sb.WriteString(fmt.Sprintf("**Связей:** %d\n\n", len(topo.Links)))
	sb.WriteString("### Связи\n\n")
	if len(topo.Links) == 0 {
		sb.WriteString("- Связи не найдены.\n")
		return sb.String()
	}
	for _, link := range topo.Links {
		sourceType := strings.TrimSpace(string(link.SourceType))
		confidence := strings.TrimSpace(string(link.Confidence))
		extra := ""
		if sourceType != "" || confidence != "" {
			extra = fmt.Sprintf(" [%s/%s]", sourceType, confidence)
		}
		sb.WriteString(fmt.Sprintf("- `%s (%s)` <-> `%s (%s)`%s\n",
			topoDisplayName(link.Source), topoPortName(link.SourcePort), topoDisplayName(link.Target), topoPortName(link.TargetPort), extra))
	}
	return sb.String()
}

func topoDisplayName(d *topology.Device) string {
	if d == nil {
		return "unknown"
	}
	if d.Hostname != "" {
		return d.Hostname
	}
	if d.IP != "" {
		return d.IP
	}
	if d.MAC != "" {
		return d.MAC
	}
	return "unknown"
}

func topoPortName(p *topology.Port) string {
	if p == nil {
		return "-"
	}
	if p.Name != "" {
		return p.Name
	}
	if p.Index > 0 {
		return fmt.Sprintf("if%d", p.Index)
	}
	return "-"
}
