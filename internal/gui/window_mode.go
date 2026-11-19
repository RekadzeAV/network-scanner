package gui

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/platform"
)

// windowMode — режим отображения окна приложения.
type windowMode int

const (
	// windowModeNormal — обычный оконный режим.
	windowModeNormal windowMode = iota
	// windowModeFullscreen — полноэкранный режим (borderless, без рамок).
	windowModeFullscreen
)

// prefWindowFullscreen — ключ Preferences для сохранения полноэкранного режима.
const prefWindowFullscreen = "scan.ui.fullscreen"

// maxWindowSize — максимальный размер окна в логических пикселях канвы.
// 7680×4320 покрывает 3×4K конфигурации; больше не нужно ни одному монитору.
const (
	maxWindowWidth  = 7680
	maxWindowHeight = 4320
)

// clampToMaxSize ограничивает размер окна до maxWindowSize.
// Глобальная функция, доступна из watcher layout в results_view_layout.go.
func clampToMaxSize(sz fyne.Size) fyne.Size {
	if sz.Width > maxWindowWidth {
		sz.Width = maxWindowWidth
	}
	if sz.Height > maxWindowHeight {
		sz.Height = maxWindowHeight
	}
	return sz
}

// windowModeCtrl — контроллер режимов окна, встраивается в App.
//
// Полноэкранный режим реализован как borderless-окно (снятие WS_CAPTION +
// WS_THICKFRAME и растягивание на монитор через Win32 SetWindowPos), а НЕ
// через fyne.Window.SetFullScreen.
//
// Причина: SetFullScreen вызывает glfwSetWindowMonitor, который пересоздаёт
// OpenGL surface и меняет monitor у GLFW-окна. В Fyne v2.7.1 это приводит к
// трём регрессиям:
//   - menu bar теряется (canvas.buildMenu не вызывается повторно);
//   - canvas.scale не пересчитывается (detectScale вызывается только в
//     processMoved, а не в processResized) → координаты мыши проецируются по
//     старому масштабу, клики «не попадают» по видимому положению виджетов;
//   - layout пересобирается на промежуточный размер viewport.
//
// Borderless-подход не трогает ни surface, ни monitor, ни scale — поэтому все
// три проблемы исчезают, а пользовательский эффект (окно без рамок на весь
// экран) идентичен.
type windowModeCtrl struct {
	app       *App
	window    fyne.Window // для RunNative / watcher
	state     windowModeState
	mu        sync.Mutex    // защищает state.mode при конкурентном доступе из watcher
	watchStop chan struct{} // канал для остановки goroutine watcher
	hwnd      uintptr       // кэшированный HWND (для watcher-синхронизации)
}

// windowModeState хранит текущее режим и виджеты переключения.
type windowModeState struct {
	mode        windowMode
	normalSize  fyne.Size // размер канвы до входа в fullscreen (для восстановления)
	toggleBtn   *widget.Button
	menuItemFS  *fyne.MenuItem
	menuItemWin *fyne.MenuItem
}

// newWindowModeCtrl создаёт контроллер режимов окна для данного App.
func newWindowModeCtrl(app *App) *windowModeCtrl {
	return &windowModeCtrl{
		app: app,
		state: windowModeState{
			mode: windowModeNormal,
		},
		watchStop: make(chan struct{}),
	}
}

// initWindowMode инициализирует контроллер режима окна и восстанавливает
// сохранённый режим из Preferences.
func (a *App) initWindowMode() {
	a.winMode = newWindowModeCtrl(a)

	// Восстанавливаем сохранённый режим.
	if a.myApp != nil {
		if a.myApp.Preferences().Bool(prefWindowFullscreen) {
			a.winMode.state.mode = windowModeFullscreen
		}
	}
}

// setupWindowMode применяет сохранённый режим при старте и подключает
// обработчик изменения размера окна для отслеживания перехода в fullscreen,
// а также фоновый watcher для синхронизации с внешними изменениями окна.
func (a *App) setupWindowMode() {
	if a.winMode == nil {
		a.initWindowMode()
	}
	// Подключаем watcher для синхронизации с внешними изменениями окна
	// (Alt+Tab, Win+←, системные горячие клавиши WM и т.п.).
	a.winMode.watchForExternalFullscreenChanges(a.myWindow)
	// Примечание: stopWatcher() вызывается из SetOnClosed в Run() (app.go).
	// Сам переход в fullscreen откладывается до Run(): borderless требует
	// валидного HWND, которого нет до Show().
}

// applyPendingFullscreen вызывается из Run() ПОСЛЕ Show(): кэширует HWND и
// применяет сохранённый из Preferences полноэкранный режим.
func (a *App) applyPendingFullscreen() {
	if a.winMode == nil {
		return
	}
	a.winMode.cacheHWND()
	if a.winMode.state.mode == windowModeFullscreen {
		a.winMode.enterFullscreen()
	}
}

// ToggleFullscreen переключает полноэкранный/оконный режим.
func (a *App) ToggleFullscreen() {
	if a.winMode == nil {
		a.initWindowMode()
	}
	a.winMode.toggle()
}

// IsFullscreen возвращает true, если окно в полноэкранном режиме.
func (a *App) IsFullscreen() bool {
	if a.winMode == nil {
		return false
	}
	a.winMode.mu.Lock()
	defer a.winMode.mu.Unlock()
	return a.winMode.state.mode == windowModeFullscreen
}

// currentMode возвращает текущий режим окна.
func (c *windowModeCtrl) currentMode() windowMode {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state.mode
}

// toggle переключает между оконным и полноэкранным режимами.
func (c *windowModeCtrl) toggle() {
	if c.currentMode() == windowModeFullscreen {
		c.exitFullscreen()
	} else {
		c.enterFullscreen()
	}
}

// cacheHWND извлекает и кэширует нативный HWND окна.
// Вызывается из UI-потока после Show(), когда viewport уже создан.
func (c *windowModeCtrl) cacheHWND() {
	if c.app == nil || c.app.myWindow == nil {
		return
	}
	native, ok := c.app.myWindow.(driver.NativeWindow)
	if !ok {
		return
	}
	native.RunNative(func(ctx any) {
		if wc, ok := ctx.(driver.WindowsWindowContext); ok {
			c.hwnd = wc.HWND
		}
	})
}

// enterFullscreen переходит в полноэкранный режим.
//
// Используется нативный fyne.Window.SetFullScreen (glfwSetWindowMonitor),
// а НЕ borderless-трюк через SetWindowPos. Причина: при прямом изменении
// геометрии окна в обход GLFW драйвер не получает событие resized, его
// viewport и canvas.size остаются рассогласованы с фактическим окном, а
// DPI-масштаб (detectScale) пересчитывается только на событии moved. Итог —
// координаты мыши проецируются по устаревшей геометрии: клики «не попадают»
// по видимому положению виджетов, причём направление смещения зависело от
// кэшированного масштаба (отсюда «курсор ниже кнопки» в одном случае и
// «выше» — в другом). Нативный fullscreen синхронизирует viewport, размер
// канвы и координаты мыши средствами GLFW — рассинхрон исчезает.
func (c *windowModeCtrl) enterFullscreen() {
	if c.app == nil || c.app.myWindow == nil {
		return
	}
	c.mu.Lock()
	if c.state.mode == windowModeFullscreen {
		c.mu.Unlock()
		return
	}
	// Сохраняем размер канвы для восстановления при выходе (логические px).
	c.state.normalSize = clampToMaxSize(c.app.myWindow.Canvas().Size())
	c.state.mode = windowModeFullscreen
	c.mu.Unlock()

	geomLog.logNow("enterFullscreen BEFORE SetFullScreen(true)")
	c.app.myWindow.SetFullScreen(true)
	geomLog.logNow("enterFullscreen AFTER SetFullScreen(true)")
	c.save()
	// GLFW пересоздаёт surface при смене monitor — пересобираем menu bar
	// overlay и layout под новый размер канвы.
	c.refreshMenuBar()
	c.refreshUI()
}

// exitFullscreen выходит из полноэкранного режима.
func (c *windowModeCtrl) exitFullscreen() {
	if c.app == nil || c.app.myWindow == nil {
		return
	}
	c.mu.Lock()
	if c.state.mode == windowModeNormal {
		c.mu.Unlock()
		return
	}
	c.state.mode = windowModeNormal
	c.mu.Unlock()

	geomLog.logNow("exitFullscreen BEFORE SetFullScreen(false)")
	c.app.myWindow.SetFullScreen(false)
	geomLog.logNow("exitFullscreen AFTER SetFullScreen(false)")
	// Восстанавливаем размер и центрируем. Делаем это на следующем тике UI,
	// чтобы GLFW завершил переход monitor→window и не перезаписал размер.
	if c.state.normalSize.Width > 0 && c.state.normalSize.Height > 0 {
		size := clampToMaxSize(c.state.normalSize)
		time.AfterFunc(120*time.Millisecond, func() {
			fyne.Do(func() {
				if c.app.myWindow != nil {
					c.app.myWindow.Resize(size)
					c.app.myWindow.CenterOnScreen()
				}
			})
		})
	}
	c.save()
	c.refreshMenuBar()
	c.refreshUI()
}

// save сохраняет текущий режим в Preferences.
func (c *windowModeCtrl) save() {
	if c.app == nil || c.app.myApp == nil {
		return
	}
	c.mu.Lock()
	mode := c.state.mode
	c.mu.Unlock()
	c.app.myApp.Preferences().SetBool(prefWindowFullscreen, mode == windowModeFullscreen)
}

// syncFromWindow сверяет внутреннее состояние с фактическим режимом окна,
// установленным оконным менеджером (Alt+Tab, Win+← и т.п.).
// Возвращает true, если режим изменился (для обновления UI вызывающим кодом).
func (c *windowModeCtrl) syncFromWindow() bool {
	if c.hwnd == 0 {
		return false
	}
	activeFullscreen := platform.IsBorderlessFullscreen(c.hwnd)
	c.mu.Lock()
	changed := false
	if c.state.mode != windowModeFullscreen && activeFullscreen {
		c.state.mode = windowModeFullscreen
		changed = true
	} else if c.state.mode == windowModeFullscreen && !activeFullscreen {
		c.state.mode = windowModeNormal
		changed = true
	}
	c.mu.Unlock()
	return changed
}

// watchForExternalFullscreenChanges запускает фоновый «watcher», который
// сверяет внутреннее состояние с фактическим окном каждые 500 мс.
// При внешнем изменении режима (Alt+Tab, Win+← и т.п.) — обновляет UI-виджеты
// и пересчитывает layout. Останавливается через stopWatcher().
func (c *windowModeCtrl) watchForExternalFullscreenChanges(w fyne.Window) {
	c.window = w
	if c.watchStop == nil {
		c.watchStop = make(chan struct{})
	}
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-c.watchStop:
				return
			case <-ticker.C:
				// syncFromWindow читает HWND — делаем это из UI-потока.
				fyne.Do(func() {
					if c.syncFromWindow() {
						// Внешний переход — обновляем виджеты и layout.
						c.refreshWidgets()
						c.refreshLayout()
					}
				})
			}
		}
	}()
}

// stopWatcher останавливает фоновый watcher синхронизации.
func (c *windowModeCtrl) stopWatcher() {
	select {
	case <-c.watchStop:
		// уже остановлен
	default:
		close(c.watchStop)
	}
}

// refreshWidgets обновляет только виджеты режима окна (кнопка тулбара, пункты меню)
// без пересчёта layout. Вызывается из watcher при внешнем изменении режима.
func (c *windowModeCtrl) refreshWidgets() {
	if c.app == nil {
		return
	}
	// Страховка: тулбар должен быть виден всегда. Ранее StopScan скрывал
	// его целиком, и вместе с ним исчезала кнопка «Полный экран» —
	// пользователь оставался в fullscreen без видимого выхода.
	if c.app.mainToolbar != nil {
		c.app.mainToolbar.Show()
	}
	c.mu.Lock()
	mode := c.state.mode
	c.mu.Unlock()
	// Обновляем кнопку в тулбаре.
	if c.state.toggleBtn != nil {
		if mode == windowModeFullscreen {
			c.state.toggleBtn.SetText("Оконный режим")
			c.state.toggleBtn.SetIcon(iconRestore())
		} else {
			c.state.toggleBtn.SetText("Полный экран")
			c.state.toggleBtn.SetIcon(iconFullScreen())
		}
	}
	// Обновляем пункты меню.
	if c.state.menuItemFS != nil {
		c.state.menuItemFS.Checked = mode == windowModeFullscreen
	}
	if c.state.menuItemWin != nil {
		c.state.menuItemWin.Checked = mode == windowModeNormal
	}
}

// refreshLayout пересчитывает layout под текущий размер окна.
// Вызывается из watcher при внешнем изменении режима (уже в UI-потоке).
func (c *windowModeCtrl) refreshLayout() {
	if c.app == nil {
		return
	}
	a := c.app
	if a.myWindow == nil || a.myWindow.Canvas() == nil {
		return
	}
	size := a.myWindow.Canvas().Size()
	scale := a.myWindow.Canvas().Scale()
	if size.Width <= 0 || size.Height <= 0 {
		return
	}
	a.lastCanvasSize = size
	a.lastCanvasScale = scale
	profile := a.detectLayoutProfile(size.Width, size.Height)
	a.applyResponsiveLayout(profile)
}

// refreshUI обновляет виджеты и пересчитывает layout под текущий размер окна.
// Вызывается из enterFullscreen/exitFullscreen (UI поток).
func (c *windowModeCtrl) refreshUI() {
	c.refreshWidgets()
	c.refreshLayout()
}

// refreshMenuBar принудительно перерисовывает menu bar.
// Вызывается из enterFullscreen/exitFullscreen. В fullscreen режиме Fyne
// рисует menu bar как overlay поверх OpenGL — вызов SetMainMenu там вызывает
// glfwSetWindowMonitor(NULL), что СБРАСЫВАЕТ fullscreen. Поэтому в fullscreen
// мы просто не трогаем меню (оно уже отрисовано как overlay).
func (c *windowModeCtrl) refreshMenuBar() {
	a := c.app
	if a == nil || a.myWindow == nil {
		return
	}
	c.mu.Lock()
	mode := c.state.mode
	c.mu.Unlock()
	// В fullscreen режиме SetMainMenu вызывает glfwSetWindowMonitor(NULL),
	// что автоматически выходит из fullscreen — см. Fyne internal/driver/glfw.
	if mode == windowModeFullscreen {
		return
	}
	if menu := a.myWindow.MainMenu(); menu != nil {
		a.myWindow.SetMainMenu(menu)
	}
}

// makeFullscreenToggleBtn создаёт кнопку переключения fullscreen для тулбара.
func (a *App) makeFullscreenToggleBtn() *widget.Button {
	if a.winMode == nil {
		a.initWindowMode()
	}
	btn := widget.NewButtonWithIcon("Полный экран", iconFullScreen(), func() {
		a.ToggleFullscreen()
	})
	a.winMode.state.toggleBtn = btn
	if a.winMode.state.mode == windowModeFullscreen {
		btn.SetText("Оконный режим")
		btn.SetIcon(iconRestore())
	}
	return btn
}

// fullscreenMenuItems возвращает пункты меню «Вид» для управления режимом окна.
func (a *App) fullscreenMenuItems() []*fyne.MenuItem {
	if a.winMode == nil {
		a.initWindowMode()
	}
	isFS := a.winMode.state.mode == windowModeFullscreen

	itemFS := fyne.NewMenuItem("Полноэкранный режим (F11)", func() {
		if !a.IsFullscreen() {
			a.ToggleFullscreen()
		}
	})
	itemFS.Icon = iconFullScreen()
	itemFS.Checked = isFS

	itemWin := fyne.NewMenuItem("Оконный режим (F11)", func() {
		if a.IsFullscreen() {
			a.ToggleFullscreen()
		}
	})
	itemWin.Icon = iconRestore()
	itemWin.Checked = !isFS

	a.winMode.state.menuItemFS = itemFS
	a.winMode.state.menuItemWin = itemWin

	return []*fyne.MenuItem{itemFS, itemWin}
}
