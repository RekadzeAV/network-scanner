package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"

	"network-scanner/internal/platform"
)

// geomLogger — диагностика геометрии окна.
//
// Включается переменной окружения NS_GEOM_LOG.
//   - NS_GEOM_LOG=1      → пишет в %TEMP%\ns_geom.log
//   - NS_GEOM_LOG=<path> → пишет в указанный файл
//
// Каждые interval мс пишет строку:
//
//	timestamp | canvas.Size | canvas.Scale | WindowRect | ClientRect |
//	MonitorWorkRect | mode(borderless/windowed)
//
// Кроме того logNow(tag) пишет разделитель при явных переходах fullscreen
// (вызывается из enterFullscreen/exitFullscreen).
//
// Это НЕ часть продукта: при отсутствии NS_GEOM_LOG newGeomLogger() возвращает
// nil, а все методы — no-op. Нужен для съёма фактических чисел рассинхрона
// canvas.size vs реального окна, которое невозможно воспроизвести кликами.
type geomLogger struct {
	path   string
	mu     sync.Mutex
	win    fyne.Window
	closed chan struct{}
}

// newGeomLogger создаёт logger, если NS_GEOM_LOG установлен; иначе nil.
func newGeomLogger() *geomLogger {
	env := os.Getenv("NS_GEOM_LOG")
	if env == "" {
		return nil
	}

	path := env
	if env == "1" {
		path = filepath.Join(os.TempDir(), "ns_geom.log")
	}

	// Truncate существующего файла, чтобы каждый запуск начинал с чистой истории.
	_ = os.WriteFile(path, []byte(fmt.Sprintf("=== newGeomLogger @ %s path=%s ===\n",
		time.Now().Format("2006-01-02 15:04:05"), path)), 0644)

	return &geomLogger{path: path, closed: make(chan struct{})}
}

// attach привязывает окно и запускает фоновый опрос геометрии.
func (l *geomLogger) attach(win fyne.Window) {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.win = win
	l.mu.Unlock()
	go l.run()
}

// stop останавливает фоновый опрос.
func (l *geomLogger) stop() {
	if l == nil {
		return
	}
	select {
	case <-l.closed:
	default:
		close(l.closed)
	}
}

func (l *geomLogger) run() {
	time.Sleep(400 * time.Millisecond) // ждём появления viewport

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-l.closed:
			return
		case <-ticker.C:
			l.sample()
		}
	}
}

// geomSnapshot — снимок геометрии, собранный из UI-потока.
type geomSnapshot struct {
	canvasW, canvasH float32
	scale            float32
	hwnd             uintptr
}

func (l *geomLogger) sample() {
	snap := &geomSnapshot{}

	done := make(chan struct{})
	fyne.Do(func() {
		l.mu.Lock()
		win := l.win
		l.mu.Unlock()
		if win != nil {
			c := win.Canvas()
			sz := c.Size()
			snap.canvasW, snap.canvasH = sz.Width, sz.Height
			snap.scale = c.Scale()
			if native, ok := win.(driver.NativeWindow); ok {
				native.RunNative(func(ctx any) {
					if wc, ok := ctx.(driver.WindowsWindowContext); ok {
						snap.hwnd = wc.HWND
					}
				})
			}
		}
		close(done)
	})
	<-done

	l.write(snap, "")
}

// logNow пишет разделитель с меткой tag и текущий canvas size/scale.
// Вызывается из enterFullscreen/exitFullscreen. Специально НЕ трогает
// RunNative/hwnd: эти методы могут выполняться на UI-потоке (applyPendingFullscreen
// при старте), а RunNative с главного потока блокирует до GL-потока — риск
// дедлока. Точные win/client/mon/hwnd снимает фоновый sample() каждые 200мс.
func (l *geomLogger) logNow(tag string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	win := l.win
	l.mu.Unlock()

	snap := &geomSnapshot{}
	if win != nil {
		c := win.Canvas()
		sz := c.Size()
		snap.canvasW, snap.canvasH = sz.Width, sz.Height
		snap.scale = c.Scale()
	}
	l.write(snap, tag)
}

func (l *geomLogger) write(snap *geomSnapshot, tag string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.path == "" {
		return
	}

	var winRectStr, clientRectStr, monWorkStr, modeStr string
	if snap.hwnd != 0 {
		if wr, ok := platform.WindowRect(snap.hwnd); ok {
			winRectStr = fmt.Sprintf("%dx%d@%d,%d", wr.Width(), wr.Height(), wr.Left, wr.Top)
		}
		if cr, ok := platform.ClientRect(snap.hwnd); ok {
			clientRectStr = fmt.Sprintf("%dx%d", cr.Width(), cr.Height())
		}
		if mwr, ok := platform.MonitorWorkRectForWindow(snap.hwnd); ok {
			monWorkStr = fmt.Sprintf("%dx%d@%d,%d", mwr.Width(), mwr.Height(), mwr.Left, mwr.Top)
		}
		if platform.IsBorderlessFullscreen(snap.hwnd) {
			modeStr = "borderless"
		} else {
			modeStr = "windowed"
		}
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	ts := time.Now().Format("15:04:05.000")
	if tag != "" {
		fmt.Fprintf(f, "[%s] >>> %s\n", ts, tag)
	}
	fmt.Fprintf(f,
		"[%s] canvas=%.0fx%.0f scale=%.2f | win=[%s] client=[%s] mon=[%s] mode=%s\n",
		ts, snap.canvasW, snap.canvasH, snap.scale,
		winRectStr, clientRectStr, monWorkStr, modeStr)
}

// startGeomCycle — автоматический цикл входа/выхода из fullscreen для диагностики.
//
// Включается переменной окружения NS_GEOM_CYCLE=<мс> (например 3000).
// Приложение само, без мыши и клавиатуры, делает 4 переключения с интервалом
// <мс>, записывая в геометрию-лог маркеры enter/exit. Нужно потому, что в
// автоматизированной среде GUI-процесс не доживает до ручного клика по кнопке,
// а клики мышью через UI Automation нестабильны (окно перемещается watcher'ом).
//
// Без переменной — no-op, продукт не меняется.
func (a *App) startGeomCycle() {
	if geomLog == nil {
		return
	}
	raw := os.Getenv("NS_GEOM_CYCLE")
	if raw == "" {
		return
	}
	interval, err := time.ParseDuration(raw + "ms")
	if err != nil || interval <= 0 {
		return
	}

	go func() {
		// Даём окну отрисоваться и фоновому сэмплеру набрать baseline.
		time.Sleep(interval)
		for i := 0; i < 4; i++ {
			seq := i + 1
			fyne.Do(func() {
				geomLog.logNow(fmt.Sprintf("AUTO_CYCLE toggle #%d (isFullscreen=%v)", seq, a.IsFullscreen()))
				a.ToggleFullscreen()
			})
			time.Sleep(interval)
		}
		geomLog.logNow("AUTO_CYCLE done")
	}()
}
