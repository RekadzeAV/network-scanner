package gui

import (
	"image/color"
	"testing"

	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

func TestRefreshAutoProfileStateLabel_On(t *testing.T) {
	a := &App{
		autoProfileCheck:       widget.NewCheck("", nil),
		autoProfileStateText:   canvas.NewText("", color.Black),
		autoProfileHeaderLabel: widget.NewLabel(""),
	}
	_ = fyneapp.New()
	a.autoProfileCheck.SetChecked(true)

	a.refreshAutoProfileStateLabel()

	if a.autoProfileStateText.Text != "Автопрофиль: ВКЛ" {
		t.Fatalf("unexpected state text: %q", a.autoProfileStateText.Text)
	}
	if a.autoProfileHeaderLabel.Text != "Режим сканирования: Автопрофиль ВКЛ" {
		t.Fatalf("unexpected header text: %q", a.autoProfileHeaderLabel.Text)
	}
	// Цвет — из палитры варианта темы (не хардкод), чтобы не ломался при смене палитры.
	want := themeColorSuccess()
	if a.autoProfileStateText.Color != want {
		t.Fatalf("unexpected color: %#v, want theme success color %#v", a.autoProfileStateText.Color, want)
	}
}

func TestRefreshAutoProfileStateLabel_Off(t *testing.T) {
	a := &App{
		autoProfileCheck:       widget.NewCheck("", nil),
		autoProfileStateText:   canvas.NewText("", color.Black),
		autoProfileHeaderLabel: widget.NewLabel(""),
	}
	_ = fyneapp.New()
	a.autoProfileCheck.SetChecked(false)

	a.refreshAutoProfileStateLabel()

	if a.autoProfileStateText.Text != "Автопрофиль: ВЫКЛ" {
		t.Fatalf("unexpected state text: %q", a.autoProfileStateText.Text)
	}
	if a.autoProfileHeaderLabel.Text != "Режим сканирования: Автопрофиль ВЫКЛ" {
		t.Fatalf("unexpected header text: %q", a.autoProfileHeaderLabel.Text)
	}
	// Цвет — из палитры варианта темы (disabled/серый).
	want := themeColorDisabled()
	if a.autoProfileStateText.Color != want {
		t.Fatalf("unexpected color: %#v, want theme disabled color %#v", a.autoProfileStateText.Color, want)
	}
}

func TestRefreshAutoProfileStateLabel_NilSafe(t *testing.T) {
	a := &App{}
	// Должно отработать без panic даже при неполной инициализации.
	a.refreshAutoProfileStateLabel()
}
