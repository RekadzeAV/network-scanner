package controller

import (
	"testing"

	"fyne.io/fyne/v2/widget"
)

// --- tools_controller.go extended tests ---

func TestToolsController_SetButtonsEnabled_NilUI(t *testing.T) {
	c := &ToolsController{}
	// Не должен паниковать
	c.setButtonsEnabled(true)
	c.setButtonsEnabled(false)
}

func TestToolsController_SetButtonsEnabled_WithUI(t *testing.T) {
	ui := &ToolsUI{
		PingBtn:  widget.NewButton("Ping", nil),
		TraceBtn: widget.NewButton("Trace", nil),
		DNSBtn:   widget.NewButton("DNS", nil),
	}
	c := &ToolsController{ui: ui}
	c.setButtonsEnabled(false)
	if ui.PingBtn.Disabled() != true {
		t.Error("expected PingBtn disabled")
	}
	c.setButtonsEnabled(true)
	if ui.PingBtn.Disabled() != false {
		t.Error("expected PingBtn enabled")
	}
}

func TestToolsController_SetOutputMarkdown_NilUI(t *testing.T) {
	c := &ToolsController{}
	// Не должен паниковать
	c.setOutputMarkdown("test")
}

func TestToolsController_SetOutputMarkdown_NilOutput(t *testing.T) {
	c := &ToolsController{ui: &ToolsUI{}}
	// Не должен паниковать
	c.setOutputMarkdown("test")
}

func TestToolsController_SetOutputMarkdown_WithOutput(t *testing.T) {
	ui := &ToolsUI{
		ToolsOutput: widget.NewRichText(),
	}
	c := &ToolsController{ui: ui}
	c.setOutputMarkdown("### Test")
}

func TestParseIntOrDefault_LargeNumber(t *testing.T) {
	if parseIntOrDefault("99999", 10) != 99999 {
		t.Error("expected 99999 for '99999'")
	}
}

func TestParseIntOrDefault_WithSpaces(t *testing.T) {
	if parseIntOrDefault("  42  ", 10) != 42 {
		t.Error("expected 42 for '  42  '")
	}
}

func TestNewToolsController_Created(t *testing.T) {
	c := NewToolsController(nil, nil)
	if c == nil {
		t.Fatal("expected non-nil ToolsController")
	}
}

func TestNewToolsController_WithApp(t *testing.T) {
	c := NewToolsController(ensureApp(), &ToolsUI{})
	if c == nil {
		t.Fatal("expected non-nil ToolsController")
	}
}
