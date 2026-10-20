package gui

import (
	"fmt"
	"network-scanner/internal/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// buildScanSettingsDialog создаёт диалог настроек сканирования (P2-11).
// Содержит пресеты, профиль, производительность и опции — вместо accordion на панели.
func (a *App) buildScanSettingsDialog() fyne.CanvasObject {
	defer func() {
		if rec := recover(); rec != nil {
			logger.LogError(fmt.Errorf("PANIC в buildScanSettingsDialog: %v", rec), "buildScanSettingsDialog")
		}
	}()

	// --- Пресеты и профиль ---
	presetsRow := container.NewGridWithColumns(3,
		a.presetQuickBtn, a.presetBalBtn, a.presetDeepBtn,
	)
	recommendedRow := container.NewGridWithColumns(2,
		a.recommendedProfileBtn, a.recommendedProfileInfoBtn,
	)

	profilesSection := container.NewVBox(
		widget.NewLabel("Пресет:"),
		presetsRow,
		widget.NewLabel("Рекомендуемые настройки:"),
		recommendedRow,
		a.recommendedProfileBadge,
		container.NewGridWithColumns(2, a.autoProfileCheck, a.autoProfileInfoBtn),
		a.autoProfileStateText,
		a.autoProfileHint,
	)

	// --- Производительность и опции ---
	optionsSection := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("Таймаут (сек):"), a.timeoutEntryBox,
			widget.NewLabel("Потоки:"), a.threadsEntryBox,
		),
		a.scanUDPCheck,
		a.scanBannersCheck,
		a.scanOSActiveCheck,
		container.NewGridWithColumns(2, a.scanVerboseLogsCheck, a.scanVerboseInfoBtn),
	)

	// Группируем в табы
	tabs := container.NewAppTabs(
		container.NewTabItem("Пресеты и профиль", profilesSection),
		container.NewTabItem("Производительность", optionsSection),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}

// showScanSettingsDialog открывает диалог настроек сканирования.
func (a *App) showScanSettingsDialog() {
	if a == nil || a.myWindow == nil {
		return
	}
	content := a.buildScanSettingsDialog()
	d := dialog.NewCustom("Настройки сканирования", "OK", content, a.myWindow)
	d.Resize(fyne.NewSize(480, 400))
	d.Show()
}

// buildSettingsButton создаёт кнопку «Настройки сканирования» для панели.
// Заменяет accordion (P2-11).
func (a *App) buildSettingsButton() *widget.Button {
	btn := widget.NewButtonWithIcon("Настройки сканирования", iconSettings(), func() {
		a.showScanSettingsDialog()
	})
	btn.Importance = widget.MediumImportance
	return btn
}

// initSettingsMenu добавляет пункт меню «Настройки сканирования» (P2-11).
func (a *App) initSettingsMenu() {
	if a == nil || a.myWindow == nil {
		return
	}

	menu := fyne.NewMenu("Инструменты",
		fyne.NewMenuItem("Настройки сканирования...", func() {
			a.showScanSettingsDialog()
		}),
	)

	menuBar := fyne.NewMainMenu(menu)
	a.myWindow.SetMainMenu(menuBar)
}
