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
	want := color.RGBA{R: 60, G: 170, B: 80, A: 255}
	if a.autoProfileStateText.Color != want {
		t.Fatalf("unexpected color: %#v", a.autoProfileStateText.Color)
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
	want := color.RGBA{R: 140, G: 140, B: 140, A: 255}
	if a.autoProfileStateText.Color != want {
		t.Fatalf("unexpected color: %#v", a.autoProfileStateText.Color)
	}
}

func TestRefreshAutoProfileStateLabel_NilSafe(t *testing.T) {
	a := &App{}
	// Должно отработать без panic даже при неполной инициализации.
	a.refreshAutoProfileStateLabel()
}

