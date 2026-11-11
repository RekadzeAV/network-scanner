package gui

import (
	"errors"
	"image/color"
	"strings"
	"testing"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// --- S-05: showFieldErrorDelayed прокидывает ошибку в статус-бар ---

func TestShowFieldErrorDelayed_ErrSetsStatus(t *testing.T) {
	a, err := newTestAppWithStatus(t)
	if err != nil {
		t.Fatal(err)
	}
	msg := canvas.NewText("", color.RGBA{})
	a.showFieldErrorDelayed(msg, errors.New("bad cidr"))
	if got := a.statusLabel.Text; !strings.HasPrefix(got, "✗ ") {
		t.Errorf("expected status to start with marker, got %q", got)
	}
}

func TestShowFieldErrorDelayed_NilErrKeepsStatus(t *testing.T) {
	a, err := newTestAppWithStatus(t)
	if err != nil {
		t.Fatal(err)
	}
	a.setStatus("previous")
	msg := canvas.NewText("", color.RGBA{})
	a.showFieldErrorDelayed(msg, nil)
	if got := a.statusLabel.Text; got != "previous" {
		t.Errorf("status must not change on nil error, got %q", got)
	}
}

func TestShowFieldErrorDelayed_NilAppSafe(t *testing.T) {
	var a *App
	msg := canvas.NewText("", color.RGBA{})
	a.showFieldErrorDelayed(msg, errors.New("boom"))
}

func TestShowFieldError_HidesMsgOnNilError(t *testing.T) {
	msg := canvas.NewText("old", color.RGBA{})
	msg.Show()
	showFieldError(msg, nil)
	if msg.Visible() {
		t.Error("msg must be hidden on nil error")
	}
}

func TestShowFieldError_ShowsErrorWithMarker(t *testing.T) {
	msg := canvas.NewText("", color.RGBA{})
	msg.Hide()
	showFieldError(msg, errors.New("bad input"))
	if !msg.Visible() {
		t.Error("msg must be visible on error")
	}
	if msg.Text != "⚠ bad input" {
		t.Errorf("msg = %q, want ⚠ bad input", msg.Text)
	}
}

// --- makeValidatedEntry ---

func TestMakeValidatedEntry_ReturnsBoxAndMsg(t *testing.T) {
	e := widget.NewEntry()
	box, msg := makeValidatedEntry(e)
	if box == nil || msg == nil {
		t.Fatal("expected non-nil box and msg")
	}
	if msg.Visible() {
		t.Error("msg must be hidden initially")
	}
	if msg.TextSize <= 0 {
		t.Errorf("msg.TextSize must be positive, got %v", msg.TextSize)
	}
}

func TestValidateIntField(t *testing.T) {
	if err := validateIntField("", 1, 60); err != nil {
		t.Errorf("empty allowed: %v", err)
	}
	if err := validateIntField("5", 1, 60); err != nil {
		t.Errorf("5 in [1,60]: %v", err)
	}
	if err := validateIntField("61", 1, 60); err == nil {
		t.Error("61 out of range must error")
	}
	if err := validateIntField("not-a-number", 1, 60); err == nil {
		t.Error("non-numeric must error")
	}
}

func TestValidateCIDRField(t *testing.T) {
	if err := validateCIDRField(""); err != nil {
		t.Errorf("empty allowed: %v", err)
	}
	if err := validateCIDRField("192.168.1.0/24"); err != nil {
		t.Errorf("valid CIDR: %v", err)
	}
	if err := validateCIDRField("not-a-cidr"); err == nil {
		t.Error("invalid CIDR must error")
	}
}

func TestValidatePortRangeField(t *testing.T) {
	if err := validatePortRangeField(""); err != nil {
		t.Errorf("empty allowed: %v", err)
	}
	if err := validatePortRangeField("1-1024"); err != nil {
		t.Errorf("valid range: %v", err)
	}
	if err := validatePortRangeField("70000-80000"); err == nil {
		t.Error("out-of-range must error")
	}
}

// newTestAppWithStatus создаёт App с statusLabel, достаточный для
// showFieldErrorDelayed в unit-тестах.
func newTestAppWithStatus(t *testing.T) (*App, error) {
	t.Helper()
	fyneApp := test.NewApp()
	a := &App{myApp: fyneApp, statusLabel: widget.NewLabel("")}
	t.Cleanup(func() {
		a.cancelStatusToast()
		fyneApp.Quit()
	})
	return a, nil
}
