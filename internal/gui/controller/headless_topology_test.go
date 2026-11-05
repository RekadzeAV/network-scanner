package controller

import (
	"testing"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestTopologyController_applyTopologyRunStart_nilUI проверяет, что
// applyTopologyRunStart безопасно завершается при nil UI.
func TestTopologyController_applyTopologyRunStart_nilUI(t *testing.T) {
	ctrl := NewTopologyController(nil)
	ctrl.applyTopologyRunStart() // не должен паниковать
}

// TestTopologyController_applyTopologyRunStart проверяет обновление UI
// при запуске построения топологии.
func TestTopologyController_applyTopologyRunStart(t *testing.T) {
	ui := &TopologyUI{
		BuildTopoBtn:   widget.NewButton("", nil),
		TopoStatus:     widget.NewLabel(""),
		SNMPStageLabel: widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
	}
	ctrl := NewTopologyController(ui)
	ctrl.applyTopologyRunStart()

	if ui.BuildTopoBtn.Disabled() == false {
		t.Fatal("BuildTopoBtn should be disabled")
	}
	if ui.TopoStatus.Text != "Построение топологии..." {
		t.Fatalf("TopoStatus = %q, want %q", ui.TopoStatus.Text, "Построение топологии...")
	}
	if ui.SNMPStageLabel.Hidden == true {
		t.Fatal("SNMPStageLabel should be shown")
	}
	if ui.SNMPProgress.Hidden == true {
		t.Fatal("SNMPProgress should be shown")
	}
}

// TestTopologyController_applyTopologyProgress_nilUI проверяет, что
// applyTopologyProgress безопасно завершается при nil UI.
func TestTopologyController_applyTopologyProgress_nilUI(t *testing.T) {
	ctrl := NewTopologyController(nil)
	ctrl.applyTopologyProgress("test", 0.5) // не должен паниковать
}

// TestTopologyController_applyTopologyProgress проверяет обновление прогресса.
func TestTopologyController_applyTopologyProgress(t *testing.T) {
	ui := &TopologyUI{
		TopoStatus:     widget.NewLabel(""),
		SNMPStageLabel: widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
	}
	ctrl := NewTopologyController(ui)
	ctrl.applyTopologyProgress("SNMP: 5/10", 0.5)

	if ui.TopoStatus.Text != "SNMP: 5/10" {
		t.Fatalf("TopoStatus = %q, want %q", ui.TopoStatus.Text, "SNMP: 5/10")
	}
	if ui.SNMPProgress.Value != 0.5 {
		t.Fatalf("SNMPProgress.Value = %v, want %v", ui.SNMPProgress.Value, 0.5)
	}
}

// TestTopologyController_applyTopologyCanceled_nilUI проверяет, что
// applyTopologyCanceled безопасно завершается при nil UI.
func TestTopologyController_applyTopologyCanceled_nilUI(t *testing.T) {
	ctrl := NewTopologyController(nil)
	ctrl.applyTopologyCanceled() // не должен паниковать
}

// TestTopologyController_applyTopologyCanceled проверяет обновление UI
// при отмене построения.
func TestTopologyController_applyTopologyCanceled(t *testing.T) {
	ui := &TopologyUI{
		BuildTopoBtn:   widget.NewButton("", nil),
		StopTopoBtn:    widget.NewButton("", nil),
		TopoStatus:     widget.NewLabel(""),
		SNMPStageLabel: widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
	}
	ctrl := NewTopologyController(ui)
	// Симулируем запуск (кнопка Disable, статус-бар Show)
	ctrl.applyTopologyRunStart()
	ctrl.applyTopologyCanceled()

	if ui.TopoStatus.Text != "Построение отменено" {
		t.Fatalf("TopoStatus = %q, want %q", ui.TopoStatus.Text, "Построение отменено")
	}
	if ui.BuildTopoBtn.Disabled() {
		t.Fatal("BuildTopoBtn should be enabled")
	}
	if !ui.StopTopoBtn.Disabled() {
		t.Fatal("StopTopoBtn should be disabled")
	}
	if !ui.SNMPStageLabel.Hidden {
		t.Fatal("SNMPStageLabel should be hidden")
	}
	if !ui.SNMPProgress.Hidden {
		t.Fatal("SNMPProgress should be hidden")
	}
}

// TestTopologyController_applyTopologyFailure_nilUI проверяет, что
// applyTopologyFailure безопасно завершается при nil UI.
func TestTopologyController_applyTopologyFailure_nilUI(t *testing.T) {
	ctrl := NewTopologyController(nil)
	ctrl.applyTopologyFailure("snmp") // не должен паниковать
}

// TestTopologyController_applyTopologyFailure проверяет обновление UI
// при ошибке построения.
func TestTopologyController_applyTopologyFailure(t *testing.T) {
	ui := &TopologyUI{
		BuildTopoBtn:   widget.NewButton("", nil),
		StopTopoBtn:    widget.NewButton("", nil),
		TopoStatus:     widget.NewLabel(""),
		SNMPStageLabel: widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
	}
	ctrl := NewTopologyController(ui)
	// Симулируем запуск, чтобы кнопка была initially disabled
	ctrl.applyTopologyRunStart()
	ctrl.applyTopologyFailure("snmp")

	if ui.TopoStatus.Text != "Ошибка на этапе: snmp" {
		t.Fatalf("TopoStatus = %q, want %q", ui.TopoStatus.Text, "Ошибка на этапе: snmp")
	}
	if ui.BuildTopoBtn.Disabled() {
		t.Fatal("BuildTopoBtn should be enabled")
	}
	if !ui.StopTopoBtn.Disabled() {
		t.Fatal("StopTopoBtn should be disabled")
	}
	if !ui.SNMPStageLabel.Hidden {
		t.Fatal("SNMPStageLabel should be hidden")
	}
	if !ui.SNMPProgress.Hidden {
		t.Fatal("SNMPProgress should be hidden")
	}
}

// TestTopologyController_applyTopologySuccess_nilUI проверяет, что
// applyTopologySuccess безопасно завершается при nil UI.
func TestTopologyController_applyTopologySuccess_nilUI(t *testing.T) {
	ctrl := NewTopologyController(nil)
	ctrl.applyTopologySuccess("ok", "preview", nil, nil, topologyBuildMetrics{}) // не должен паниковать
}

// TestTopologyController_applyTopologySuccess проверяет обновление UI
// при успешном построении.
func TestTopologyController_applyTopologySuccess(t *testing.T) {
	ui := &TopologyUI{
		TopoText:       widget.NewRichText(),
		BuildTopoBtn:   widget.NewButton("", nil),
		StopTopoBtn:    widget.NewButton("", nil),
		TopoStatus:     widget.NewLabel(""),
		SNMPStageLabel: widget.NewLabel(""),
		SNMPProgress:   widget.NewProgressBar(),
	}
	ctrl := NewTopologyController(ui)
	// Симулируем запуск (BuildTopoBtn.Disable), затем success
	ctrl.applyTopologyRunStart()
	ctrl.applyTopologySuccess("ok", "preview", nil, nil, topologyBuildMetrics{})

	if ui.TopoStatus.Text != "ok" {
		t.Fatalf("TopoStatus = %q, want %q", ui.TopoStatus.Text, "ok")
	}
	if ui.BuildTopoBtn.Disabled() {
		t.Fatal("BuildTopoBtn should be enabled")
	}
	if !ui.StopTopoBtn.Disabled() {
		t.Fatal("StopTopoBtn should be disabled")
	}
	if !ui.SNMPStageLabel.Hidden {
		t.Fatal("SNMPStageLabel should be hidden")
	}
	if !ui.SNMPProgress.Hidden {
		t.Fatal("SNMPProgress should be hidden")
	}
}

// TestTopologyController_ApplyTopologyZoom_nilUI проверяет, что
// ApplyTopologyZoom безопасно завершается при nil UI.
func TestTopologyController_ApplyTopologyZoom_nilUI(t *testing.T) {
	ctrl := NewTopologyController(nil)
	ctrl.ApplyTopologyZoom("Fit", nil) // не должен паниковать
}

// TestTopologyController_ApplyTopologyZoom_nilTopoImage проверяет, что
// ApplyTopologyZoom безопасно завершается при nil TopoImage.
func TestTopologyController_ApplyTopologyZoom_nilTopoImage(t *testing.T) {
	ui := &TopologyUI{
		TopoImage: nil,
	}
	ctrl := NewTopologyController(ui)
	ctrl.ApplyTopologyZoom("Fit", nil) // не должен паниковать
}

// TestTopologyController_ApplyTopologyZoom_200 проверяет масштаб 200%.
func TestTopologyController_ApplyTopologyZoom_200(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(nil)
	defer w.Close()

	img := canvas.NewImageFromFile("")
	img.FillMode = canvas.ImageFillContain

	ui := &TopologyUI{
		TopoImage: img,
	}
	ctrl := NewTopologyController(ui)
	ctrl.ApplyTopologyZoom("200%", w)

	if img.FillMode != canvas.ImageFillOriginal {
		t.Fatalf("FillMode = %v, want %v", img.FillMode, canvas.ImageFillOriginal)
	}
}

// TestTopologyController_ApplyTopologyZoom_150 проверяет масштаб 150%.
func TestTopologyController_ApplyTopologyZoom_150(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(nil)
	defer w.Close()

	img := canvas.NewImageFromFile("")
	img.FillMode = canvas.ImageFillContain

	ui := &TopologyUI{
		TopoImage: img,
	}
	ctrl := NewTopologyController(ui)
	ctrl.ApplyTopologyZoom("150%", w)

	if img.FillMode != canvas.ImageFillOriginal {
		t.Fatalf("FillMode = %v, want %v", img.FillMode, canvas.ImageFillOriginal)
	}
}

// TestTopologyController_ApplyTopologyZoom_100 проверяет масштаб 100%.
func TestTopologyController_ApplyTopologyZoom_100(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(nil)
	defer w.Close()

	img := canvas.NewImageFromFile("")
	img.FillMode = canvas.ImageFillContain

	ui := &TopologyUI{
		TopoImage: img,
	}
	ctrl := NewTopologyController(ui)
	ctrl.ApplyTopologyZoom("100%", w)

	if img.FillMode != canvas.ImageFillOriginal {
		t.Fatalf("FillMode = %v, want %v", img.FillMode, canvas.ImageFillOriginal)
	}
}

// TestTopologyController_ApplyTopologyZoom_Fit проверяет масштаб Fit.
func TestTopologyController_ApplyTopologyZoom_Fit(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(nil)
	defer w.Close()

	img := canvas.NewImageFromFile("")
	img.FillMode = canvas.ImageFillOriginal

	ui := &TopologyUI{
		TopoImage: img,
	}
	ctrl := NewTopologyController(ui)
	ctrl.ApplyTopologyZoom("Fit", w)

	if img.FillMode != canvas.ImageFillContain {
		t.Fatalf("FillMode = %v, want %v", img.FillMode, canvas.ImageFillContain)
	}
}

// TestTopologyController_ApplyTopologyZoom_UnknownMode проверяет неизвестный
// режим масштаба.
func TestTopologyController_ApplyTopologyZoom_UnknownMode(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(nil)
	defer w.Close()

	img := canvas.NewImageFromFile("")
	img.FillMode = canvas.ImageFillOriginal

	ui := &TopologyUI{
		TopoImage: img,
	}
	ctrl := NewTopologyController(ui)
	ctrl.ApplyTopologyZoom("Unknown", w)

	if img.FillMode != canvas.ImageFillContain {
		t.Fatalf("FillMode = %v, want %v", img.FillMode, canvas.ImageFillContain)
	}
}

// TestTopologyController_ApplyTopologyZoom_nilWindow проверяет масштаб Fit
// при nil window.
func TestTopologyController_ApplyTopologyZoom_nilWindow(t *testing.T) {
	img := canvas.NewImageFromFile("")
	img.FillMode = canvas.ImageFillOriginal

	ui := &TopologyUI{
		TopoImage: img,
	}
	ctrl := NewTopologyController(ui)
	ctrl.ApplyTopologyZoom("Fit", nil)

	if img.FillMode != canvas.ImageFillContain {
		t.Fatalf("FillMode = %v, want %v", img.FillMode, canvas.ImageFillContain)
	}
}

// TestTopologyController_StopTopologyBuild_nilCancel проверяет, что при nil
// cancel StopTopologyBuild безопасно завершается.
func TestTopologyController_StopTopologyBuild_nilCancel(t *testing.T) {
	ctrl := NewTopologyController(nil)
	ctrl.StopTopologyBuild() // не должен паниковать
}
