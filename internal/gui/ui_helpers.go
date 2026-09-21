package gui

import (
	"fmt"
	"strconv"
	"strings"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/audit"
	"network-scanner/internal/devicecontrol"
)

// recommendedBadgeClassForHosts возвращает класс бейджа по количеству хостов.
func (a *App) recommendedBadgeClassForHosts(hosts int) string {
	switch {
	case hosts >= autoProfileHostXXLarge:
		return "very-large"
	case hosts >= autoProfileHostXLarge:
		return "large"
	case hosts >= autoProfileHostLarge:
		return "medium"
	default:
		return "small"
	}
}

// recommendedBadgeText формирует текст бейджа профиля.
func (a *App) recommendedBadgeText(profileName string, badgeClass string) string {
	return fmt.Sprintf("Профиль: %s (%s)", profileName, badgeClass)
}

// recommendedProfileNameForClass возвращает имя профиля по классу.
func (a *App) recommendedProfileNameForClass(badgeClass string) (string, bool) {
	switch strings.TrimSpace(badgeClass) {
	case "very-large":
		return "бережный для очень крупной подсети", true
	case "large":
		return "бережный для крупной подсети", true
	case "medium":
		return "сбалансированный для средней подсети", true
	case "small":
		return "углубленный для небольшой подсети", true
	default:
		return "", false
	}
}

// refreshAutoProfileStateLabel обновляет лейбл состояния автопрофиля.
func (a *App) refreshAutoProfileStateLabel() {
	if a == nil || a.autoProfileStateText == nil {
		return
	}
	enabled := true
	if a.autoProfileCheck != nil {
		enabled = a.autoProfileCheck.Checked
	}
	if enabled {
		a.autoProfileStateText.Text = "Автопрофиль: ВКЛ"
		a.autoProfileStateText.Color = themeColorSuccess()
		if a.autoProfileHeaderLabel != nil {
			a.autoProfileHeaderLabel.SetText("Режим сканирования: Автопрофиль ВКЛ")
			a.autoProfileHeaderLabel.Refresh()
		}
	} else {
		a.autoProfileStateText.Text = "Автопрофиль: ВЫКЛ"
		a.autoProfileStateText.Color = themeColorDisabled()
		if a.autoProfileHeaderLabel != nil {
			a.autoProfileHeaderLabel.SetText("Режим сканирования: Автопрофиль ВЫКЛ")
			a.autoProfileHeaderLabel.Refresh()
		}
	}
	a.autoProfileStateText.Refresh()
}

// buildToolCard создаёт визуальную карточку-секцию для панели инструментов:
// заголовок с иконкой, разделитель и содержимое. Упрощает восприятие панели,
// группируя параметры по категориям.
func (a *App) buildToolCard(title string, icon fyne.Resource, content fyne.CanvasObject) fyne.CanvasObject {
	header := container.NewHBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
	if icon != nil {
		header = container.NewHBox(widget.NewIcon(icon), header.Objects[0])
	}
	return container.NewVBox(
		header,
		widget.NewSeparator(),
		content,
		widget.NewSeparator(),
	)
}

// setupMainMenu создаёт главное меню приложения.
func (a *App) setupMainMenu() {
	if a == nil || a.myWindow == nil {
		return
	}
	themeMode := a.loadTheme()
	a.applyTheme(themeMode)
	// Перерисовка под активный вариант темы (светлый/тёмный) — background
	// чипов/строк/диаграмм задаются в коде рендера, а не через тему Fyne.
	a.scheduleResultsRender(true)

	resetItem := fyne.NewMenuItem("Сбросить расположение панелей (Ctrl+Shift+L)", func() {
		a.settingsMgr.ResetUIPanelLayoutWithFeedback(a.scanTabMainSplit, a.topologyMainSplit, a.toolsTabMainSplit, a.myWindow)
	})
	resetItem.Icon = iconRestore()
	shortcutsItem := fyne.NewMenuItem("Горячие клавиши (F1)", func() {
		a.showShortcutsDialog()
	})
	shortcutsItem.Icon = iconHelp()

	aboutItem := fyne.NewMenuItem("О программе", func() {
		a.showAboutDialog()
	})
	aboutItem.Icon = iconInfo()
	helpMenu := fyne.NewMenu("Справка", shortcutsItem, aboutItem)
	viewMenu := fyne.NewMenu("Вид", resetItem)

	themeItems := []string{"Светлая", "Тёмная", "Системная"}
	themeMenu := fyne.NewMenu("Тема")
	for _, item := range themeItems {
		mode := ThemeMode(item)
		menuItem := fyne.NewMenuItem(item, func() {
			a.applyTheme(mode)
			a.saveTheme(mode)
			if a.statusLabel != nil {
				a.statusLabel.SetText("Тема изменена: " + item)
			}
		})
		if string(mode) == string(themeMode) {
			menuItem.Checked = true
		}
		themeMenu.Items = append(themeMenu.Items, menuItem)
	}

	accentMenu := fyne.NewMenu("Акцент")
	currentPreset := a.loadAccentPreset()
	for name := range PresetThemes {
		menuItem := fyne.NewMenuItem(name, func(n string) func() {
			return func() {
				a.applyAccentPreset(n)
			}
		}(name))
		if name == currentPreset {
			menuItem.Checked = true
		}
		accentMenu.Items = append(accentMenu.Items, menuItem)
	}

	mainMenu := fyne.NewMainMenu(viewMenu, themeMenu, accentMenu, helpMenu)
	a.myWindow.SetMainMenu(mainMenu)
}

// setupLayoutResetShortcut создаёт хоткей сброса расположения панелей.
func (a *App) setupLayoutResetShortcut() {
	if a == nil || a.myWindow == nil {
		return
	}
	c := a.myWindow.Canvas()
	if c == nil {
		return
	}
	sc := &desktop.CustomShortcut{
		KeyName:  fyne.KeyL,
		Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift,
	}
	c.AddShortcut(sc, func(fyne.Shortcut) {
		a.settingsMgr.ResetUIPanelLayoutWithFeedback(a.scanTabMainSplit, a.topologyMainSplit, a.toolsTabMainSplit, a.myWindow)
	})
}

// loadScanSettings загружает сохранённые настройки сканирования.
func (a *App) loadScanSettings() {
	if a == nil || a.myApp == nil {
		return
	}
	p := a.myApp.Preferences()
	if v := strings.TrimSpace(p.String(prefNetwork)); v != "" {
		a.networkEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefPortRange)); v != "" {
		a.portRangeEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefTimeout)); v != "" {
		a.timeoutEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefThreads)); v != "" {
		a.threadsEntry.SetText(v)
	}
	a.scanUDPCheck.SetChecked(strings.EqualFold(strings.TrimSpace(p.String(prefScanUDP)), "true"))
	if a.scanBannersCheck != nil {
		a.scanBannersCheck.SetChecked(strings.EqualFold(strings.TrimSpace(p.String(prefScanBanners)), "true"))
	}
	if a.scanOSActiveCheck != nil {
		a.scanOSActiveCheck.SetChecked(strings.EqualFold(strings.TrimSpace(p.String(prefScanOSActive)), "true"))
	}
	if a.scanVerboseLogsCheck != nil {
		a.scanVerboseLogsCheck.SetChecked(strings.EqualFold(strings.TrimSpace(p.String(prefScanVerbosePortLogs)), "true"))
	}
	if a.scanTCPPortsCheck != nil {
		tcpPref := strings.TrimSpace(p.String(prefScanTCPPorts))
		if tcpPref == "" || strings.EqualFold(tcpPref, "true") {
			a.scanTCPPortsCheck.SetChecked(true)
		} else {
			a.scanTCPPortsCheck.SetChecked(false)
		}
		a.setPortRangeControlsEnabled(a.scanTCPPortsCheck.Checked)
	}
	if a.autoProfileCheck != nil {
		autoPref := strings.TrimSpace(p.String(prefAutoProfile))
		if autoPref == "" || strings.EqualFold(autoPref, "true") {
			a.autoProfileCheck.SetChecked(true)
		} else {
			a.autoProfileCheck.SetChecked(false)
		}
	}
	a.refreshAutoProfileStateLabel()

	switch strings.TrimSpace(p.String(prefPreset)) {
	case "quick":
		a.statusLabel.SetText("Пресет: Быстро (восстановлен)")
	case "deep":
		a.statusLabel.SetText("Пресет: Глубоко (восстановлен)")
	case "balanced":
		a.statusLabel.SetText("Пресет: Баланс (восстановлен)")
	case "recommended":
		a.statusLabel.SetText("Пресет: Рекомендуемые настройки (восстановлен)")
	}
	if a.recommendedProfileBadge != nil {
		if badgeClass := strings.TrimSpace(p.String(prefRecommendedBadgeClass)); badgeClass != "" {
			if profileName, ok := a.recommendedProfileNameForClass(badgeClass); ok {
				a.recommendedProfileBadge.Text = a.recommendedBadgeText(profileName, badgeClass)
				a.recommendedProfileBadge.Color = color.RGBA{R: 55, G: 130, B: 200, A: 255}
				a.recommendedProfileBadge.Refresh()
			} else if badgeText := strings.TrimSpace(p.String(prefRecommendedBadge)); badgeText != "" {
				a.recommendedProfileBadge.Text = badgeText
				a.recommendedProfileBadge.Color = color.RGBA{R: 55, G: 130, B: 200, A: 255}
				a.recommendedProfileBadge.Refresh()
			}
		} else if badgeText := strings.TrimSpace(p.String(prefRecommendedBadge)); badgeText != "" {
			a.recommendedProfileBadge.Text = badgeText
			a.recommendedProfileBadge.Color = color.RGBA{R: 55, G: 130, B: 200, A: 255}
			a.recommendedProfileBadge.Refresh()
		}
	}
	viewMode := strings.TrimSpace(p.String(prefViewMode))
	if viewMode == "Таблица" || viewMode == "Карточки" {
		a.resultsMode = viewMode
		a.resultsModeSel.SetSelected(viewMode)
	}
	subMode := strings.TrimSpace(p.String(prefResultsSubMode))
	if subMode == "Devices" || subMode == "Security" || subMode == "Inventory" {
		a.resultsSubMode = subMode
		if a.resultsSubModeSel != nil {
			a.resultsSubModeSel.SetSelected(subMode)
		}
	}
	if v := strings.TrimSpace(p.String(prefInventoryDBPath)); v != "" && a.inventoryDBEntry != nil {
		a.inventoryDBEntry.SetText(v)
	}
	if a.inventoryAutoSaveCheck != nil {
		autoSave := strings.TrimSpace(p.String(prefInventoryAutoSave))
		a.inventoryAutoSaveCheck.SetChecked(autoSave == "" || strings.EqualFold(autoSave, "true"))
	}
	sortMode := strings.TrimSpace(p.String(prefSortMode))
	if sortMode == "IP" || sortMode == "HostName" {
		a.resultsSort = sortMode
		a.resultsSortSel.SetSelected(sortMode)
	}
	if v, err := strconv.Atoi(strings.TrimSpace(p.String(prefChipLimit))); err == nil && v > 0 {
		a.maxPortChips = v
		if a.chipLimitSel != nil {
			a.chipLimitSel.SetSelected(strconv.Itoa(v))
		}
	}
	a.showRawBanners = strings.EqualFold(strings.TrimSpace(p.String(prefShowRawBanners)), "true")
	if a.showRawBannersCheck != nil {
		a.showRawBannersCheck.SetChecked(a.showRawBanners)
	}
	if v := strings.TrimSpace(p.String(prefFilterQuery)); v != "" {
		a.resultsFilterQuery = v
		if a.resultsFilterEnt != nil {
			a.resultsFilterEnt.SetText(v)
		}
	}
	a.onlyWithOpenPorts = strings.EqualFold(strings.TrimSpace(p.String(prefOnlyOpenPorts)), "true")
	if a.openPortsOnlyCheck != nil {
		a.openPortsOnlyCheck.SetChecked(a.onlyWithOpenPorts)
	}
	if rawTypes := strings.TrimSpace(p.String(prefTypeFilters)); rawTypes != "" {
		for _, typeName := range strings.Split(rawTypes, ",") {
			name := strings.TrimSpace(typeName)
			if check, ok := a.quickTypeChecks[name]; ok && check != nil {
				check.SetChecked(true)
			}
		}
	}
	if v := strings.TrimSpace(p.String(prefToolHost)); v != "" && a.toolsHostEntry != nil {
		a.toolsHostEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefToolPingCount)); v != "" && a.toolsPingCountEnt != nil {
		a.toolsPingCountEnt.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefToolTimeout)); v != "" && a.toolsTimeoutEnt != nil {
		a.toolsTimeoutEnt.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefToolTraceHops)); v != "" && a.toolsTraceHopsEnt != nil {
		a.toolsTraceHopsEnt.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefToolResolver)); v != "" && a.toolsDNSResolverEnt != nil {
		a.toolsDNSResolverEnt.SetText(v)
	}
	if a.toolsAuditMinSeveritySel != nil {
		sev := strings.TrimSpace(p.String(prefToolAuditMinSeverity))
		if sev == "" {
			sev = "low"
		}
		if _, ok := audit.NormalizeSeverity(sev); !ok {
			sev = "low"
		}
		a.toolsAuditMinSeveritySel.SetSelected(sev)
	}
	if v := strings.TrimSpace(p.String(prefToolDeviceTarget)); v != "" && a.toolsDeviceTargetEntry != nil {
		a.toolsDeviceTargetEntry.SetText(v)
	}
	if a.toolsDeviceVendorEntry != nil {
		vendor := strings.TrimSpace(p.String(prefToolDeviceVendor))
		if vendor == "" {
			vendor = devicecontrol.VendorGenericHTTP
		}
		a.toolsDeviceVendorEntry.SetSelected(vendor)
	}
	if v := strings.TrimSpace(p.String(prefToolDeviceUser)); v != "" && a.toolsDeviceUserEntry != nil {
		a.toolsDeviceUserEntry.SetText(v)
	}
	if v := strings.TrimSpace(p.String(prefCidrFilter)); v != "" {
		if a.resultsCidrFilterEnt != nil {
			a.resultsCidrFilterEnt.SetText(v)
		}
	}
	mode := strings.TrimSpace(p.String(prefPortStateMode))
	switch mode {
	case "has_open":
		a.resultsPortStateMode = mode
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть открытые")
		}
	case "has_closed":
		a.resultsPortStateMode = mode
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть закрытые")
		}
	case "has_filtered":
		a.resultsPortStateMode = mode
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Есть фильтруемые")
		}
	default:
		a.resultsPortStateMode = "all"
		if a.resultsPortStateSel != nil {
			a.resultsPortStateSel.SetSelected("Все")
		}
	}
	a.loadScanTabSplitFromPrefs()
	a.loadTopologySplitFromPrefs()
	a.loadToolsTabSplitFromPrefs()
	a.loadHostDetailsSplitFromPrefs()
	a.refreshInventorySnapshots()
	a.renderScanResultsView()
}
