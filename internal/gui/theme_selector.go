package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeMode определяет режим темы
type ThemeMode string

const (
	ThemeModeLight  ThemeMode = "light"
	ThemeModeDark   ThemeMode = "dark"
	ThemeModeSystem ThemeMode = "system"
	prefThemeMode             = "gui.theme.mode"
	prefDensity               = "gui.theme.density"
)

// themeModeLabels — подписи пунктов меню «Тема» и соответствующие им режимы.
// Подпись больше не приводится к ThemeMode напрямую (P0-1): значения ThemeMode
// — внутренние константы, русские строки только отображаются.
var themeModeLabels = []struct {
	mode    ThemeMode
	caption string
}{
	{ThemeModeLight, "Светлая"},
	{ThemeModeDark, "Тёмная"},
	{ThemeModeSystem, "Системная"},
}

// themeCaption — подпись режима для меню и статусов.
func themeCaption(mode ThemeMode) string {
	for _, item := range themeModeLabels {
		if item.mode == mode {
			return item.caption
		}
	}
	return string(ThemeModeSystem)
}

// themeModeFromValue — режим из значения виджета/меню: принимаются и внутренние
// значения ("light"/"dark"/"system"), и подписи ("Светлая"...). Неизвестное
// значение приводит к ThemeModeSystem, а не к «пустой» теме (P0-1).
func themeModeFromValue(value string) ThemeMode {
	v := strings.ToLower(strings.TrimSpace(value))
	for _, item := range themeModeLabels {
		if v == string(item.mode) || v == strings.ToLower(item.caption) {
			return item.mode
		}
	}
	return ThemeModeSystem
}

// newThemeForMode — тема для режима (P0-3): «системная» означает нашу палитру,
// следующую варианту, определённому системой, а не отказ от темы (nil).
func newThemeForMode(mode ThemeMode, accent string) *ModernTheme {
	switch mode {
	case ThemeModeLight:
		return NewModernThemeAccent(false, accent)
	case ThemeModeDark:
		return NewModernThemeAccent(true, accent)
	default:
		t := NewSystemTheme()
		t.SetAccent(accent)
		return t
	}
}

// applyTheme применяет тему к приложению.
func (a *App) applyTheme(mode ThemeMode) {
	if a == nil || a.myApp == nil {
		return
	}
	th := newThemeForMode(mode, a.loadAccentPreset())
	// L1.2: восстановленный из Preferences произвольный акцент применяется
	// поверх пресета, если выбран режим "Custom".
	if th.Accent() == customAccentName {
		if hex := a.loadCustomAccentHex(); hex != "" {
			_ = th.SetCustomAccent(hex)
		}
	}
	a.myApp.Settings().SetTheme(th)
	a.invalidateThemeCaches()
}

// applyThemeToUI перестраивает виджеты, состояние которых зависит от темы и
// которые НЕ перекрашиваются сами при Settings().SetTheme: главное меню
// (отметки Checked у текущего пункта «Тема»/«Акцент»), canvas-тексты и
// прямоугольники, получающие цвета из токенов в момент создания.
// Вызывается из обработчиков смены темы/акцента после applyTheme.
// Рекурсии нет: setupMainMenu вызывает applyTheme, но не applyThemeToUI.
func (a *App) applyThemeToUI() {
	if a == nil {
		return
	}
	if a.myWindow != nil {
		a.setupMainMenu()
	}
	a.refreshThemeDependentWidgets()
}

// refreshThemeDependentWidgets перекрашивает canvas-примитивы (цвета которых
// задаются из токенов в момент создания, а не через fyne.Theme), чтобы смена
// темы применялась без перезапуска.
func (a *App) refreshThemeDependentWidgets() {
	if a == nil {
		return
	}
	a.refreshAutoProfileStateLabel()
}

// invalidateThemeCaches сбрасывает кэши, в которые вошли цвета палитры (P1-8).
func (a *App) invalidateThemeCaches() {
	if a == nil {
		return
	}
	a.pieChartCache = make(map[string]fyne.Resource)
	a.analyticsCacheKey = ""
	a.analyticsCacheView = nil
	a.resultsPipelineCacheMu.Lock()
	a.resultsPipelineCacheKey = ""
	a.resultsPipelineCacheData = nil
	a.resultsPipelineCacheMu.Unlock()
	a.scheduleResultsRender(true)
}

// applyAccentPreset применяет пресет акцентных цветов
func (a *App) applyAccentPreset(name string) {
	if a == nil || a.myApp == nil {
		return
	}
	a.saveAccentPreset(name)
	// Перезагружаем текущую тему с новым пресетом
	a.applyTheme(a.loadTheme())
	a.applyThemeToUI()
	if a.statusLabel != nil {
		a.setStatus("Акцентный цвет: " + normalizeAccentPreset(name))
	}
}

// loadTheme загружает сохраненную тему из Preferences
func (a *App) loadTheme() ThemeMode {
	if a == nil || a.myApp == nil {
		return ThemeModeSystem
	}
	return themeModeFromValue(a.myApp.Preferences().StringWithFallback(prefThemeMode, string(ThemeModeSystem)))
}

// saveTheme сохраняет тему в Preferences
func (a *App) saveTheme(mode ThemeMode) {
	if a == nil || a.myApp == nil {
		return
	}
	a.myApp.Preferences().SetString(prefThemeMode, string(mode))
}

// currentThemeIsDark сообщает, тёмный ли вариант темы активен прямо сейчас.
// Используется для цветов, задаваемых в обход fyne.Theme (canvas-примитивы,
// кастомные виджеты), чтобы они не «слепили» белым на тёмном фоне (P0-2).
func currentThemeIsDark() bool {
	if app := fyne.CurrentApp(); app != nil && app.Settings() != nil {
		if mt, ok := app.Settings().Theme().(*ModernTheme); ok {
			return mt.IsDark()
		}
		return app.Settings().ThemeVariant() == theme.VariantDark
	}
	return false
}

// currentModernTheme возвращает активную ModernTheme из настроек Fyne.
func (a *App) currentModernTheme() *ModernTheme {
	if a == nil || a.myApp == nil || a.myApp.Settings() == nil {
		return nil
	}
	if mt, ok := a.myApp.Settings().Theme().(*ModernTheme); ok {
		return mt
	}
	return nil
}

// applyDensity меняет плотность интерфейса (T-02).
// Обновляет тему, инвалидирует кэши и перерисовывает UI.
func (a *App) applyDensity(d Density) {
	if a == nil || a.myApp == nil || a.myApp.Settings() == nil {
		return
	}
	mt := a.currentModernTheme()
	if mt == nil {
		return
	}
	mt.SetDensity(d)
	// Сохраняем в preferences.
	a.myApp.Preferences().SetString(prefDensity, d.String())
	// Применяем тему заново, чтобы Fyne перечитал размеры.
	a.myApp.Settings().SetTheme(mt)
	// Инвалидируем кэши, зависящие от размеров.
	if a.resultsCellCache != nil {
		a.resultsCellCache.invalidate()
	}
	if a.pieChartCache != nil {
		a.pieChartCache = make(map[string]fyne.Resource)
	}
	// Перерисовываем UI.
	a.applyThemeToUI()
	a.scheduleResultsRender(true)
	if a.statusLabel != nil {
		a.setStatusToast("Плотность: " + d.String())
	}
}

// loadDensity читает плотность из preferences.
func (a *App) loadDensity() Density {
	if a == nil || a.myApp == nil {
		return DensityComfortable
	}
	v := strings.TrimSpace(a.myApp.Preferences().String(prefDensity))
	if v == DensityCompact.String() {
		return DensityCompact
	}
	return DensityComfortable
}
