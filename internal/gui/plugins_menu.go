package gui

import (
	"fmt"
	"os"

	"network-scanner/internal/logger"
	"network-scanner/internal/plugin"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

// setupPluginsMenu создаёт пункт меню «Плагины» со списком из plugin.Loader,
// вкл/выкл манифестом и отображением ошибок загрузки (L2 integration).
func (a *App) setupPluginsMenu() *fyne.Menu {
	pluginsMenu := fyne.NewMenu("Плагины")

	// Загружаем текущее состояние.
	plugins, err := a.pluginLoader.DiscoverAndLoad()
	if err != nil {
		logger.Log("plugin loader: ошибка первичной загрузки: %v", err)
	}
	logger.Log("plugin loader: загружено плагинов: %d", len(plugins))

	statuses := a.pluginLoader.Statuses()

	// Пункты для каждого обнаруженного плагина.
	for _, s := range statuses {
		name := s.Manifest.Name
		info := fmt.Sprintf("%s (%s)", name, s.Manifest.Type)
		if s.Manifest.Version != "" {
			info = fmt.Sprintf("%s %s", info, s.Manifest.Version)
		}
		enabled := s.Loaded && s.Manifest.IsEnabled()

		// Копируем значения в замыкание.
		pluginDir := s.Dir
		wasLoaded := s.Loaded

		item := fyne.NewMenuItem(info, func() {
			a.togglePluginEnabled(pluginDir, !wasLoaded)
		})
		item.Icon = theme.BrokenImageIcon()
		item.Checked = enabled
		pluginsMenu.Items = append(pluginsMenu.Items, item)
	}

	if len(pluginsMenu.Items) > 0 {
		pluginsMenu.Items = append(pluginsMenu.Items, fyne.NewMenuItemSeparator())
	}

	reloadItem := fyne.NewMenuItem("Обновить список", func() {
		a.rebuildPluginsMenu()
	})
	reloadItem.Icon = theme.ViewRefreshIcon()
	pluginsMenu.Items = append(pluginsMenu.Items, reloadItem)

	// Ошибки загрузки.
	errs := a.pluginLoader.LoadErrors()
	if len(errs) > 0 {
		pluginsMenu.Items = append(pluginsMenu.Items, fyne.NewMenuItemSeparator())
		errorsSub := fyne.NewMenu("Ошибки загрузки")
		for _, le := range errs {
			errItem := fyne.NewMenuItem(le.Error(), func() {})
			errItem.Icon = theme.WarningIcon()
			errorsSub.Items = append(errorsSub.Items, errItem)
		}
		errsItem := fyne.NewMenuItem("Ошибки загрузки", nil)
		errsItem.ChildMenu = errorsSub
		errsItem.Icon = theme.WarningIcon()
		pluginsMenu.Items = append(pluginsMenu.Items, errsItem)
	}

	// Каталоги поиска.
	dirs := a.pluginLoader.Dirs()
	if len(dirs) > 0 {
		pluginsMenu.Items = append(pluginsMenu.Items, fyne.NewMenuItemSeparator())
		dirsSub := fyne.NewMenu("Каталоги поиска")
		for _, d := range dirs {
			dir := d // копия для замыкания
			dirItem := fyne.NewMenuItem(dir, func() {
				a.showPluginDirDialog(dir)
			})
			dirItem.Icon = theme.FolderIcon()
			dirsSub.Items = append(dirsSub.Items, dirItem)
		}
		dirsItem := fyne.NewMenuItem("Каталоги поиска", nil)
		dirsItem.ChildMenu = dirsSub
		dirsItem.Icon = theme.FolderIcon()
		pluginsMenu.Items = append(pluginsMenu.Items, dirsItem)
	}

	return pluginsMenu
}

// togglePluginEnabled переключает состояние плагина (enabled) через манифест.
func (a *App) togglePluginEnabled(dir string, enabled bool) {
	manifest, err := plugin.LoadManifest(dir)
	if err != nil {
		logger.LogError(err, "plugin loader: ошибка чтения манифеста "+dir)
		return
	}

	// Установим новое значение.
	val := enabled
	manifest.Enabled = &val

	if err := plugin.SaveManifest(dir, manifest); err != nil {
		logger.LogError(err, "plugin loader: ошибка записи манифеста "+dir)
		return
	}

	// Перезагрузим плагины.
	plugins, err := a.pluginLoader.Reload()
	if err != nil {
		logger.Log("plugin loader: reload после toggle: %v", err)
	}
	logger.Log("plugin loader: toggle %s → %v, загружено: %d",
		manifest.Name, enabled, len(plugins))

	a.rebuildPluginsMenu()
}

// rebuildPluginsMenu пересоздаёт пункт меню «Плагины».
func (a *App) rebuildPluginsMenu() {
	fyne.Do(func() {
		a.setupMainMenu()
	})
}

// showPluginDirDialog открывает диалог с информацией о каталоге.
func (a *App) showPluginDirDialog(dir string) {
	info := fmt.Sprintf("Каталог плагинов:\n\n%s\n\nФайлы:\n", dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		info += fmt.Sprintf("Ошибка чтения: %v", err)
	} else {
		for _, e := range entries {
			info += fmt.Sprintf("  %s\n", e.Name())
		}
	}
	dialog.ShowInformation("Каталог плагинов", info, a.myWindow)
}

// showPluginErrorsDialog открывает диалог с ошибками загрузки плагинов.
func (a *App) showPluginErrorsDialog() {
	errs := a.pluginLoader.LoadErrors()
	if len(errs) == 0 {
		dialog.ShowInformation("Ошибки загрузки", "Ошибок загрузки плагинов нет.", a.myWindow)
		return
	}

	var msg string
	for i, le := range errs {
		msg += fmt.Sprintf("%d. %s\n", i+1, le.Error())
	}
	dialog.ShowInformation("Ошибки загрузки плагинов", msg, a.myWindow)
}
