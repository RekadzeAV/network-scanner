package controller

import (
	"context"
	"fmt"
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

type topologyBuildMetrics struct {
	snmpDuration  time.Duration
	buildDuration time.Duration
	totalDuration time.Duration
}

// TopologyController управляет построением топологии.
type TopologyController struct {
	ui      *TopologyUI
	cancel  context.CancelFunc
	lastTopo    *topology.Topology
	lastReport  *snmpcollector.CollectReport
	lastMetrics topologyBuildMetrics
}

// NewTopologyController создает контроллер.
func NewTopologyController(ui *TopologyUI) *TopologyController {
	return &TopologyController{ui: ui}
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
