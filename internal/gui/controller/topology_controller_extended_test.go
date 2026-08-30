package controller

import (
	"testing"
	"time"

	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// --- topology_controller.go extended tests ---

func TestNewTopologyController_Created(t *testing.T) {
	c := NewTopologyController(nil, nil)
	if c == nil {
		t.Fatal("expected non-nil TopologyController")
	}
}

func TestNewTopologyController_WithApp(t *testing.T) {
	c := NewTopologyController(ensureApp(), &TopologyUI{})
	if c == nil {
		t.Fatal("expected non-nil TopologyController")
	}
}

func TestTopologyController_ApplyTopologyRunStart_NilUI(t *testing.T) {
	c := &TopologyController{}
	// Не должен паниковать
	c.applyTopologyRunStart()
}

func TestTopologyController_ApplyTopologyRunStart_WithUI(t *testing.T) {
	ui := &TopologyUI{
		BuildTopoBtn:   widget.NewButton("Build", nil),
		TopoStatus:     widget.NewLabel(""),
		SNMPStageLabel: widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
	}
	c := &TopologyController{ui: ui}
	c.applyTopologyRunStart()
	if ui.TopoStatus.Text != "Построение топологии..." {
		t.Errorf("expected status text, got %q", ui.TopoStatus.Text)
	}
}

func TestTopologyController_ApplyTopologyProgress_NilUI(t *testing.T) {
	c := &TopologyController{}
	// Не должен паниковать
	c.applyTopologyProgress("test", 0.5)
}

func TestTopologyController_ApplyTopologyProgress_WithUI(t *testing.T) {
	ui := &TopologyUI{
		TopoStatus:     widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
		SNMPStageLabel: widget.NewLabel(""),
	}
	c := &TopologyController{ui: ui}
	c.applyTopologyProgress("status", 0.5)
	if ui.TopoStatus.Text != "status" {
		t.Errorf("expected 'status', got %q", ui.TopoStatus.Text)
	}
	if ui.SNMPProgress.Value != 0.5 {
		t.Errorf("expected 0.5, got %f", ui.SNMPProgress.Value)
	}
}

func TestTopologyController_ApplyTopologyCanceled_NilUI(t *testing.T) {
	c := &TopologyController{}
	// Не должен паниковать
	c.applyTopologyCanceled()
}

func TestTopologyController_ApplyTopologyCanceled_WithUI(t *testing.T) {
	ui := &TopologyUI{
		TopoStatus:     widget.NewLabel(""),
		BuildTopoBtn:   widget.NewButton("Build", nil),
		StopTopoBtn:    widget.NewButton("Stop", nil),
		SNMPStageLabel: widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
	}
	c := &TopologyController{ui: ui}
	c.applyTopologyCanceled()
	if ui.TopoStatus.Text != "Построение отменено" {
		t.Errorf("expected 'Построение отменено', got %q", ui.TopoStatus.Text)
	}
}

func TestTopologyController_ApplyTopologyFailure_NilUI(t *testing.T) {
	c := &TopologyController{}
	// Не должен паниковать
	c.applyTopologyFailure("snmp")
}

func TestTopologyController_ApplyTopologyFailure_WithUI(t *testing.T) {
	ui := &TopologyUI{
		TopoStatus:     widget.NewLabel(""),
		BuildTopoBtn:   widget.NewButton("Build", nil),
		StopTopoBtn:    widget.NewButton("Stop", nil),
		SNMPStageLabel: widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
	}
	c := &TopologyController{ui: ui}
	c.applyTopologyFailure("build")
	if ui.TopoStatus.Text != "Ошибка на этапе: build" {
		t.Errorf("expected 'Ошибка на этапе: build', got %q", ui.TopoStatus.Text)
	}
}

func TestTopologyController_ApplyTopologySuccess_NilUI(t *testing.T) {
	c := &TopologyController{}
	// Не должен паниковать
	c.applyTopologySuccess("ok", "preview", nil, nil, topologyBuildMetrics{})
}

func TestTopologyController_ApplyTopologySuccess_WithUI(t *testing.T) {
	ui := &TopologyUI{
		TopoText:       widget.NewRichText(),
		TopoStatus:     widget.NewLabel(""),
		BuildTopoBtn:   widget.NewButton("Build", nil),
		StopTopoBtn:    widget.NewButton("Stop", nil),
		SNMPStageLabel: widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
	}
	c := &TopologyController{ui: ui}
	c.applyTopologySuccess("ok", "preview", nil, nil, topologyBuildMetrics{})
	if ui.TopoStatus.Text != "ok" {
		t.Errorf("expected 'ok', got %q", ui.TopoStatus.Text)
	}
}

func TestTopologyController_StopTopologyBuild_NilCancel(t *testing.T) {
	c := &TopologyController{}
	// Не должен паниковать
	c.StopTopologyBuild()
}

func TestTopologyController_BuildPerformanceReportText_NilLastTopo(t *testing.T) {
	c := &TopologyController{}
	text := c.buildPerformanceReportText()
	if text != "" {
		t.Errorf("expected empty, got %q", text)
	}
}

func TestTopologyController_BuildPerformanceReportText_WithTopo(t *testing.T) {
	c := &TopologyController{}
	c.lastTopo = &topology.Topology{
		Devices: map[string]*topology.Device{
			"192.168.1.1": {IP: "192.168.1.1", Hostname: "router"},
		},
		Links: []topology.Link{
			{},
		},
	}
	c.lastMetrics = topologyBuildMetrics{
		snmpDuration:  1 * time.Second,
		buildDuration: 2 * time.Second,
		totalDuration: 3 * time.Second,
	}
	c.lastReport = &snmpcollector.CollectReport{
		TotalSNMPTargets: 10,
		Connected:        8,
		Partial:          1,
		Failed:           1,
	}
	text := c.buildPerformanceReportText()
	if text == "" {
		t.Fatal("expected non-empty report")
	}
	if len(c.lastTopo.Devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(c.lastTopo.Devices))
	}
}

func TestTopologyController_ApplyTopologyZoom_NilUI(t *testing.T) {
	c := &TopologyController{}
	// Не должен паниковать
	c.ApplyTopologyZoom("Fit", nil)
}

func TestTopologyController_ApplyTopologyZoom_NilImage(t *testing.T) {
	c := &TopologyController{ui: &TopologyUI{}}
	// Не должен паниковать
	c.ApplyTopologyZoom("Fit", nil)
}

func TestTopologyController_ApplyTopologyZoom_WithImage(t *testing.T) {
	ui := &TopologyUI{
		TopoImage: canvas.NewImageFromResource(nil),
	}
	c := &TopologyController{ui: ui}
	// Не должен паниковать
	c.ApplyTopologyZoom("100%", nil)
	c.ApplyTopologyZoom("150%", nil)
	c.ApplyTopologyZoom("200%", nil)
	c.ApplyTopologyZoom("Fit", nil)
}

func TestTopologyController_OpenPreviewExternal_EmptyPathRequiresWindow(t *testing.T) {
	// Без валидного window вызов паникует, поэтому проверяем только через инициализированное приложение.
	// В headless-окружении dialog.ShowInformation требует активного приложения, тест пропускаем.
	t.Skip("требует активное окно Fyne — не применимо в headless")
}

func TestSplitCommaValues_WithValues(t *testing.T) {
	result := splitCommaValues("public, private , ,admin")
	if len(result) != 3 {
		t.Errorf("expected 3 values, got %d", len(result))
	}
	if result[0] != "public" {
		t.Errorf("expected 'public', got %q", result[0])
	}
}

func TestTopologyController_RenderTopologyImagePreview_NilTopo(t *testing.T) {
	c := &TopologyController{}
	// Не должен паниковать
	c.renderTopologyImagePreview(nil, nil)
}
