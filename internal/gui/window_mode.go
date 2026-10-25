package gui

import (
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// fullscreenLayoutRefreshDelay — задержка перед пересчётом layout после смены
// режима окна. Даём оконному менеджеру принять новый размер, иначе профиль
// применится к старому размеру канвы.
const fullscreenLayoutRefreshDelay = 350 * time.Millisecond

// windowMode — режим отображения окна приложения.
type windowMode int

const (
	// windowModeNormal — обычный оконный режим.
	windowModeNormal windowMode = iota
	// windowModeFullscreen — полноэкранный режим.
	windowModeFullscreen
)

// prefWindowFullscreen — ключ Preferences для сохранения полноэкранного режима.
const prefWindowFullscreen = "scan.ui.fullscreen"

// maxWindowSize — максимальный размер окна, чтобы предотвратить выход за
// границы экранов при ручном растягивании или на мульти-мониторных конфигурациях.
// 7680×4320 — это 3×4K (UltraWide), больше не нужно ни одному монитору.
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

// windowModeState хранит текущее состояние режима окна и виджет переключения.
type windowModeState struct {
	mode        windowMode
	normalSize  fyne.Size // размер окна до входа в fullscreen
	toggleBtn   *widget.Button
	menuItemFS  *fyne.MenuItem
	menuItemWin *fyne.MenuItem
}

// windowModeCtrl — контроллер режимов окна, встраивается в App.
type windowModeCtrl struct {
	state     windowModeState
	app       *App
	window    fyne.Window   // для syncFromWindow / watcher
	layoutSeq atomic.Uint64 // debounce: отмена устаревших отложенных layout-задач
	mu        sync.Mutex    // защищает state.mode при конкурентном доступе из watcher
	watchStop chan struct{} // канал для остановки goroutine watcher
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
	if a.winMode.state.mode == windowModeFullscreen {
		a.winMode.enterFullscreen()
	}
	// Подключаем watcher для синхронизации с внешними изменениями окна
	// (Alt+Tab, Win+←, системные горячие клавиши WM и т.п.).
	a.winMode.watchForExternalFullscreenChanges(a.myWindow)
	// Примечание: stopWatcher() вызывается из SetOnClosed в Run() (app.go).
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
	return a.winMode.state.mode == windowModeFullscreen
}

// currentWindowMode возвращает текущий режим окна.
func (c *windowModeCtrl) currentMode() windowMode {
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

// enterFullscreen переходит в полноэкранный режим.
func (c *windowModeCtrl) enterFullscreen() {
	if c.app == nil || c.app.myWindow == nil {
		return
	}
	if c.state.mode == windowModeFullscreen {
		return
	}
	// Сохраняем текущий размер окна для восстановления (с ограничением max).
	c.state.normalSize = clampToMaxSize(c.app.myWindow.Canvas().Size())
	c.app.myWindow.SetFullScreen(true)
	c.state.mode = windowModeFullscreen
	c.save()
	c.refreshUI()
}

// exitFullscreen выходит из полноэкранного режима.
func (c *windowModeCtrl) exitFullscreen() {
	if c.app == nil || c.app.myWindow == nil {
		return
	}
	if c.state.mode == windowModeNormal {
		return
	}
	c.app.myWindow.SetFullScreen(false)
	// Восстанавливаем размер, если он был сохранён, с ограничением max.
	if c.state.normalSize.Width > 0 && c.state.normalSize.Height > 0 {
		c.app.myWindow.Resize(clampToMaxSize(c.state.normalSize))
		c.app.myWindow.CenterOnScreen()
	}
	c.state.mode = windowModeNormal
	c.save()
	c.refreshUI()
}

// save сохраняет текущий режим в Preferences.
func (c *windowModeCtrl) save() {
	if c.app == nil || c.app.myApp == nil {
		return
	}
	c.app.myApp.Preferences().SetBool(prefWindowFullscreen, c.state.mode == windowModeFullscreen)
}

// syncFromWindow сверяет внутреннее состояние с фактическим режимом окна,
// установленным оконным менеджером (Alt+Tab, Win+← и т.п.).
// Возвращает true, если режим изменился (для обновления UI вызывающим кодом).
func (c *windowModeCtrl) syncFromWindow() bool {
	if c.window == nil {
		return false
	}
	c.mu.Lock()
	activeFullscreen := c.window.FullScreen()
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
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-c.watchStop:
				return
			case <-ticker.C:
				if c.syncFromWindow() {
					// Внешний переход — обновляем виджеты и layout.
					c.refreshWidgets()
					c.refreshLayout()
				}
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
	// Обновляем кнопку в тулбаре.
	if c.state.toggleBtn != nil {
		if c.state.mode == windowModeFullscreen {
			c.state.toggleBtn.SetText("Оконный режим")
			c.state.toggleBtn.SetIcon(iconRestore())
		} else {
			c.state.toggleBtn.SetText("Полный экран")
			c.state.toggleBtn.SetIcon(iconFullScreen())
		}
	}
	// Обновляем пункты меню.
	if c.state.menuItemFS != nil {
		c.state.menuItemFS.Checked = c.state.mode == windowModeFullscreen
	}
	if c.state.menuItemWin != nil {
		c.state.menuItemWin.Checked = c.state.mode == windowModeNormal
	}
}

// refreshLayout пересчитывает layout под текущий размер окна без debounce.
// Вызывается из watcher при внешнем изменении режима.
func (c *windowModeCtrl) refreshLayout() {
	if c.app == nil {
		return
	}
	a := c.app
	fyne.Do(func() {
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
		profile := a.detectLayoutProfile(size.Width)
		a.applyResponsiveLayout(profile)
	})
}

// currentLayoutSeq возвращает текущее значение счётчика debounce.
func (c *windowModeCtrl) currentLayoutSeq() uint64 {
	return c.layoutSeq.Load()
}

// nextLayoutSeq инкрементирует счётчик debounce и возвращает новое значение.
func (c *windowModeCtrl) nextLayoutSeq() uint64 {
	return c.layoutSeq.Add(1)
}

// refreshUI обновляет виджеты и запускает отложенный пересчёт layout.
// Вызывается из enterFullscreen/exitFullscreen (UI поток).
func (c *windowModeCtrl) refreshUI() {
	if c.app == nil {
		return
	}
	// Обновляем виджеты (кнопка/меню) — синхронно.
	c.refreshWidgets()
	// Отложенный пересчёт layout с debounce.
	a := c.app
	seq := c.nextLayoutSeq()
	go func() {
		<-time.After(fullscreenLayoutRefreshDelay)
		fyne.Do(func() {
			// Debounce: если счётчик уже изменился — задача устарела.
			if c.currentLayoutSeq() != seq {
				return
			}
			// Синхронизируем с фактическим состоянием окна.
			if c.syncFromWindow() {
				c.refreshWidgets()
			}
			if a.myWindow == nil || a.myWindow.Canvas() == nil {
				return
			}
			size := a.myWindow.Canvas().Size()
			scale := a.myWindow.Canvas().Scale()
			if size.Width <= 0 || size.Height <= 0 {
				return
			}
			profile := a.detectLayoutProfile(size.Width)
			a.lastCanvasSize = size
			a.lastCanvasScale = scale
			a.applyResponsiveLayout(profile)
		})
	}()
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
