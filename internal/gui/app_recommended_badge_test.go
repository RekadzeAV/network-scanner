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
		myApp:                   myApp,
		scanUDPCheck:            widget.NewCheck("", nil),
		statusLabel:             widget.NewLabel(""),
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
		myApp:                   myApp,
		scanUDPCheck:            widget.NewCheck("", nil),
		statusLabel:             widget.NewLabel(""),
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
		myApp:                   myApp,
		scanUDPCheck:            widget.NewCheck("", nil),
		statusLabel:             widget.NewLabel(""),
		recommendedProfileBadge: canvas.NewText("Профиль: не выбран", color.Black),
	}

	a.loadScanSettings()

	if a.recommendedProfileBadge.Text != "Профиль: legacy fallback" {
		t.Fatalf("unexpected badge text for invalid class: %q", a.recommendedProfileBadge.Text)
	}
}

// TestRecommendedBadgeColorComesFromThemePalette — регрессия C-13: цвет бейджа
// «Рекомендуемый профиль» был жёстким #3782C8 (3.72:1 на surface) и не следовал
// за вариантом темы. Теперь во всех ветках восстановления берётся акцентный цвет
// варианта темы (5.96:1 light / 7.09:1 dark).
func TestRecommendedBadgeColorComesFromThemePalette(t *testing.T) {
	legacyBlue := color.RGBA{R: 55, G: 130, B: 200, A: 255} // прежний литерал

	tests := []struct {
		name       string
		badgeClass string
		legacyText string
	}{
		{name: "по классу профиля", badgeClass: "medium", legacyText: "не должен использоваться"},
		{name: "по legacy-тексту", badgeClass: "", legacyText: "Профиль: legacy fallback"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			appID := fmt.Sprintf("network-scanner-test-badge-%d", time.Now().UnixNano())
			myApp := fyneapp.NewWithID(appID)
			p := myApp.Preferences()
			p.SetString(prefPreset, "recommended")
			if tc.badgeClass != "" {
				p.SetString(prefRecommendedBadgeClass, tc.badgeClass)
			}
			p.SetString(prefRecommendedBadge, tc.legacyText)

			a := &App{
				myApp:                   myApp,
				scanUDPCheck:            widget.NewCheck("", nil),
				statusLabel:             widget.NewLabel(""),
				recommendedProfileBadge: canvas.NewText("Профиль: не выбран", legacyBlue),
			}

			a.loadScanSettings()

			got := a.recommendedProfileBadge.Color
			if got == nil {
				t.Fatal("бейдж остался без цвета")
			}
			if got == color.Color(legacyBlue) {
				t.Fatalf("цвет бейджа по-прежнему жёсткий %v", got)
			}
			if got != themeColorPrimary() {
				t.Fatalf("цвет бейджа %v, ожидали акцент варианта темы %v", got, themeColorPrimary())
			}
			switch got {
			case color.Color(tokLightPrimary), color.Color(tokDarkPrimary):
			default:
				t.Fatalf("цвет бейджа %v не входит в палитру primary", got)
			}
		})
	}
}
