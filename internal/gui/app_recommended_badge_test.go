package gui

import (
	"fmt"
	"image/color"
	"testing"
	"time"

	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

func TestLoadScanSettings_RestoresRecommendedBadgeFromClass(t *testing.T) {
	appID := fmt.Sprintf("network-scanner-test-%d", time.Now().UnixNano())
	myApp := fyneapp.NewWithID(appID)
	p := myApp.Preferences()
	p.SetString(prefPreset, "recommended")
	p.SetString(prefRecommendedBadgeClass, "medium")
	p.SetString(prefRecommendedBadge, "legacy text should be ignored")

	a := &App{
		myApp:                  myApp,
		scanUDPCheck:           widget.NewCheck("", nil),
		statusLabel:            widget.NewLabel(""),
		recommendedProfileBadge: canvas.NewText("Профиль: не выбран", color.Black),
	}

	a.loadScanSettings()

	wantBadge := "Профиль: сбалансированный для средней подсети (medium)"
	if a.recommendedProfileBadge.Text != wantBadge {
		t.Fatalf("unexpected badge text: got %q, want %q", a.recommendedProfileBadge.Text, wantBadge)
	}
	if a.statusLabel.Text != "Пресет: Рекомендуемые настройки (восстановлен)" {
		t.Fatalf("unexpected status text: %q", a.statusLabel.Text)
	}
}

func TestLoadScanSettings_RestoresRecommendedBadgeFromLegacyText(t *testing.T) {
	appID := fmt.Sprintf("network-scanner-test-%d", time.Now().UnixNano())
	myApp := fyneapp.NewWithID(appID)
	p := myApp.Preferences()
	p.SetString(prefPreset, "recommended")
	p.SetString(prefRecommendedBadge, "Профиль: legacy fallback")

	a := &App{
		myApp:                  myApp,
		scanUDPCheck:           widget.NewCheck("", nil),
		statusLabel:            widget.NewLabel(""),
		recommendedProfileBadge: canvas.NewText("Профиль: не выбран", color.Black),
	}

	a.loadScanSettings()

	if a.recommendedProfileBadge.Text != "Профиль: legacy fallback" {
		t.Fatalf("unexpected legacy badge text: %q", a.recommendedProfileBadge.Text)
	}
}

func TestRecommendedBadgeClassForHosts_Boundaries(t *testing.T) {
	a := &App{}

	cases := []struct {
		hosts int
		want  string
	}{
		{hosts: 0, want: "small"},
		{hosts: autoProfileHostLarge - 1, want: "small"},
		{hosts: autoProfileHostLarge, want: "medium"},
		{hosts: autoProfileHostXLarge, want: "large"},
		{hosts: autoProfileHostXXLarge, want: "very-large"},
	}

	for _, tc := range cases {
		got := a.recommendedBadgeClassForHosts(tc.hosts)
		if got != tc.want {
			t.Fatalf("hosts=%d: got %q, want %q", tc.hosts, got, tc.want)
		}
	}
}

func TestLoadScanSettings_InvalidRecommendedBadgeClass_FallbackToLegacy(t *testing.T) {
	appID := fmt.Sprintf("network-scanner-test-%d", time.Now().UnixNano())
	myApp := fyneapp.NewWithID(appID)
	p := myApp.Preferences()
	p.SetString(prefPreset, "recommended")
	p.SetString(prefRecommendedBadgeClass, "unknown-class")
	p.SetString(prefRecommendedBadge, "Профиль: legacy fallback")

	a := &App{
		myApp:                  myApp,
		scanUDPCheck:           widget.NewCheck("", nil),
		statusLabel:            widget.NewLabel(""),
		recommendedProfileBadge: canvas.NewText("Профиль: не выбран", color.Black),
	}

	a.loadScanSettings()

	if a.recommendedProfileBadge.Text != "Профиль: legacy fallback" {
		t.Fatalf("unexpected badge text for invalid class: %q", a.recommendedProfileBadge.Text)
	}
}
