package controller

import (
	"math"
	"os"

	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// applyTopologyRunStart обновляет UI при запуске построения.
func (c *TopologyController) applyTopologyRunStart() {
	if c.ui == nil {
		return
	}
	c.ui.BuildTopoBtn.Disable()
	c.ui.TopoStatus.SetText("Построение топологии...")
	c.ui.SNMPStageLabel.Show()
	c.ui.SNMPProgress.Show()
}

// applyTopologyProgress обновляет прогресс SNMP-сбора.
func (c *TopologyController) applyTopologyProgress(status string, progress float64) {
	if c.ui == nil {
		return
	}
	c.ui.TopoStatus.SetText(status)
	c.ui.SNMPProgress.SetValue(progress)
	c.ui.SNMPStageLabel.Refresh()
}

// applyTopologyCanceled обновляет UI при отмене.
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

// applyTopologyFailure обновляет UI при ошибке.
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

// applyTopologySuccess обновляет UI при успешном завершении.
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
}

// renderTopologyImagePreview создает графическое превью топологии.
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
