package controller

import (
	"testing"

	"fyne.io/fyne/v2/widget"
)

// --- security_controller.go extended tests (уникальные) ---

func TestSecurityController_SetStatus_WithLabelExtended(t *testing.T) {
	ui := &SecurityUI{StatusLabel: widget.NewLabel("")}
	c := &SecurityController{ui: ui}
	c.setStatus("test message")
	if ui.StatusLabel.Text != "test message" {
		t.Errorf("expected 'test message', got %q", ui.StatusLabel.Text)
	}
}

func TestNewSecurityController_Created(t *testing.T) {
	c := NewSecurityController(nil, nil)
	if c == nil {
		t.Fatal("expected non-nil SecurityController")
	}
}

func TestNewSecurityController_WithApp(t *testing.T) {
	c := NewSecurityController(ensureApp(), &SecurityUI{})
	if c == nil {
		t.Fatal("expected non-nil SecurityController")
	}
}

// --- settings_manager.go extended tests (уникальные) ---

func TestNewSettingsManager_NilApp(t *testing.T) {
	sm := NewSettingsManager(nil)
	if sm == nil {
		t.Fatal("expected non-nil SettingsManager")
	}
}

func TestNewSettingsManager_WithApp(t *testing.T) {
	sm := NewSettingsManager(ensureApp())
	if sm == nil {
		t.Fatal("expected non-nil SettingsManager")
	}
}
