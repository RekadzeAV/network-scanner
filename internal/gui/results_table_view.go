package gui

import (
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

// tableCellCache кэширует содержимое ячеек таблицы для избежания повторных вычислений.
// P-03: ключ не содержит Row — при смене сортировки порядок строк меняется,
// а содержимое ячейки для (IP, колонка) остаётся прежним, поэтому кэш попадает.
// Кэш живёт в App и инвалидируется при смене версии данных (setScanResults).
type tableCellCache struct {
	cache map[TableCellCacheKey]string
	mu    sync.RWMutex
}

// TableCellCacheKey уникальный ключ для кэширования ячейки.
// P-03: Row исключён из ключа — кэш устойчив к пересортировке.
type TableCellCacheKey struct {
	Col int
	IP  string
}

func newTableCellCache() *tableCellCache {
	return &tableCellCache{
		cache: make(map[TableCellCacheKey]string),
	}
}

func (c *tableCellCache) getOrCreate(key TableCellCacheKey, fn func() string) string {
	c.mu.RLock()
	if v, ok := c.cache[key]; ok {
		c.mu.RUnlock()
		return v
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.cache[key]; ok {
		return v
	}
	v := fn()
	c.cache[key] = v
	return v
}

// invalidate сбрасывает весь кэш (вызывается при изменении данных результатов).
func (c *tableCellCache) invalidate() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.cache = make(map[TableCellCacheKey]string)
	c.mu.Unlock()
}

// cellCache возвращает общий кэш ячеек приложения (ленивая инициализация).
func (a *App) cellCache() *tableCellCache {
	if a.resultsCellCache == nil {
		a.resultsCellCache = newTableCellCache()
	}
	return a.resultsCellCache
}

// buildTableView создаёт таблицу результатов сканирования.
//
// R-01: заголовок — встроенная строка таблицы (ShowHeaderRow + StickyRowCount),
// которая не скроллится вместе с данными.
// R-03: заголовок — кнопки; клик сортирует по колонке, индикатор ▲/▼ показывает
// текущий режим и направление.
func (a *App) buildTableView(data []scanner.Result) fyne.CanvasObject {
	rows := len(data)
	cols := 8
	headers := a.resultsTableHeaders()

	// P-03: общий кэш приложения устойчив к пересортировке.
	cellCache := a.cellCache()

	t := widget.NewTable(
		func() (int, int) { return rows, cols },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			l, _ := obj.(*widget.Label)
			if l == nil {
				return
			}
			if id.Row < 0 || id.Row >= len(data) {
				return
			}
			r := data[id.Row]

			// T-04: моноширинный шрифт для технических полей (IP, MAC).
			ts := fyne.TextStyle{}
			if id.Col == 1 || id.Col == 2 {
				ts.Monospace = true
			}

			// P-03: ключ без Row — кэш попадает при смене сортировки.
			cellKey := TableCellCacheKey{Col: id.Col, IP: r.IP}
			l.TextStyle = ts
			l.SetText(cellCache.getOrCreate(cellKey, func() string {
				switch id.Col {
				case 0:
					return nullDash(r.Hostname)
				case 1:
					return formatIPWithProtocol(r.IP)
				case 2:
					return nullDash(r.MAC)
				case 3:
					return deviceTypeWithBadge(r.DeviceType)
				case 4:
					return nullDash(r.DeviceVendor)
				case 5:
					return osGuessLine(r)
				case 6:
					if r.SNMPEnabled {
						return "да"
					}
					return "нет"
				case 7:
					return formatPorts(r.Ports)
				default:
					return ""
				}
			}))
		},
	)

	// R-01: встроенный заголовок, закреплённый над данными.
	t.ShowHeaderRow = true
	t.StickyRowCount = 0
	// R-03: заголовок — кнопка; клик сортирует, индикатор ▲/▼ показывает режим.
	t.CreateHeader = func() fyne.CanvasObject {
		btn := widget.NewButton("", nil)
		btn.Importance = widget.LowImportance
		btn.Alignment = widget.ButtonAlignLeading
		return btn
	}
	t.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {
		btn, _ := obj.(*widget.Button)
		if btn == nil {
			return
		}
		col := id.Col
		if col < 0 || col >= len(headers) {
			return
		}
		label := headers[col]
		if _, sortable := resultsTableSortColumns[col]; sortable {
			if arrow := a.sortIndicatorForColumn(col); arrow != "" {
				label += " " + arrow
			}
			col := col
			btn.OnTapped = func() { a.toggleSortByColumn(col) }
			btn.Enable()
		} else {
			// Несортируемые колонки: заголовок — обычная подпись без действия.
			btn.OnTapped = nil
		}
		btn.SetText(label)
	}

	widths := a.resultsTableColumnWidths()
	for col, width := range widths {
		t.SetColumnWidth(col, width)
	}
	t.OnSelected = func(id widget.TableCellID) {
		if id.Row < 0 || id.Row >= len(data) {
			return
		}
		a.selectHostForDetails(data[id.Row])
	}
	return t
}

// resultsTableColumnWidths возвращает ширину колонок таблицы в зависимости от профиля.
func (a *App) resultsTableColumnWidths() []float32 {
	profile := a.currentLayoutProfile()
	base := []float32{140, 120, 130, 120, 120, 140, 52, 280}
	if profile == "wide" {
		return []float32{180, 150, 170, 150, 170, 220, 80, 420}
	}
	if profile == "compact" {
		return []float32{110, 100, 96, 96, 110, 120, 52, 180}
	}
	return base
}

// resultsTableHeaders возвращает заголовки колонок таблицы.
func (a *App) resultsTableHeaders() []string {
	if a.currentLayoutProfile() == "compact" {
		return []string{"Имя", "IP", "MAC", "Тип", "Вендор", "ОС", "SNMP", "Порты"}
	}
	return []string{"Имя хоста", "IP", "MAC", "Тип", "Производитель", "ОС (оценка)", "SNMP", "Порты (открытые)"}
}

// buildCardsView создаёт карточки результатов сканирования.
// cardsTemplateCache кэширует созданные карточки для избежания аллокаций.
type cardsTemplateCache struct {
	templates map[int]*fyne.Container // id -> template container
	mu        sync.RWMutex
}

func newCardsTemplateCache() *cardsTemplateCache {
	return &cardsTemplateCache{
		templates: make(map[int]*fyne.Container),
	}
}

func (c *cardsTemplateCache) getOrCreate(id int, createFn func() *fyne.Container) *fyne.Container {
	c.mu.RLock()
	if t, ok := c.templates[id]; ok {
		c.mu.RUnlock()
		return t
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	// Double-check после получения write lock
	if t, ok := c.templates[id]; ok {
		return t
	}
	t := createFn()
	c.templates[id] = t
	return t
}

func (a *App) buildCardsView(data []scanner.Result) fyne.CanvasObject {
	visible := len(data)
	if a.cardsVisibleCount > 0 && visible > a.cardsVisibleCount {
		visible = a.cardsVisibleCount
	}
	a.lastRenderStats.VisibleCount = visible
	viewData := data[:visible]

	// R-09/R-12: счётчик строк отражает реально показанное число карточек.
	a.updateResultsCountLabel(visible, len(data))

	// Кэшируем template для карточек
	templateCache := newCardsTemplateCache()
	list := widget.NewList(
		func() int {
			return len(viewData)
		},
		func() fyne.CanvasObject {
			id := len(viewData) // уникальный id для кэша
			return templateCache.getOrCreate(id, func() *fyne.Container {
				title := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
				sub := widget.NewLabel("")
				sub.Wrapping = fyne.TextWrapWord
				vendor := widget.NewLabel("")
				os := widget.NewLabel("")
				portsLabel := widget.NewLabel("Порты:")
				chipsHolder := container.NewHBox(widget.NewLabel(""))
				openBtn := widget.NewButtonWithIcon("Открыть детали", iconInspect(), nil)
				card := container.NewVBox(title, sub, vendor, os, portsLabel, chipsHolder, openBtn, widget.NewSeparator())
				bg := canvas.NewRectangle(a.rowBackground())
				bg.CornerRadius = 4
				return container.NewStack(bg, container.NewPadded(card))
			})
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(viewData) {
				return
			}
			r := viewData[id]
			outer, ok1 := obj.(*fyne.Container)
			if !ok1 || len(outer.Objects) < 2 {
				return
			}
			mid, ok2 := outer.Objects[1].(*fyne.Container)
			if !ok2 || len(mid.Objects) < 1 {
				return
			}
			itemBox, ok3 := mid.Objects[0].(*fyne.Container)
			if !ok3 || len(itemBox.Objects) < 7 {
				return
			}
			title, _ := itemBox.Objects[0].(*widget.Label)
			sub, _ := itemBox.Objects[1].(*widget.Label)
			vendor, _ := itemBox.Objects[2].(*widget.Label)
			os, _ := itemBox.Objects[3].(*widget.Label)
			chipsHolder, _ := itemBox.Objects[5].(*fyne.Container)
			openBtn, _ := itemBox.Objects[6].(*widget.Button)
			if chipsHolder == nil || openBtn == nil || title == nil || sub == nil || vendor == nil || os == nil {
				return
			}

			rowTitle := strings.TrimSpace(r.Hostname)
			if rowTitle == "" {
				rowTitle = r.IP
			}
			title.SetText(rowTitle)
			sub.SetText(fmt.Sprintf("%s · %s · %s", formatIPWithProtocol(r.IP), nullDash(r.MAC), deviceTypeWithBadge(r.DeviceType)))
			vendor.SetText(fmt.Sprintf("Производитель: %s", nullDash(r.DeviceVendor)))
			os.SetText(fmt.Sprintf("ОС (оценка): %s", osGuessLine(r)))

			// Оптимизация: обновляем chips без пересоздания chipsHolder
			openBtn.OnTapped = func() {
				a.selectHostForDetails(r)
			}

			// Пересоздаём chips для обновлённых данных
			chipsHolder.Objects = []fyne.CanvasObject{a.buildPortChips(r)}
			chipsHolder.Refresh()
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(viewData) {
			return
		}
		a.selectHostForDetails(viewData[id])
	}
	if visible < len(data) {
		remaining := len(data) - visible
		loadMore := widget.NewButtonWithIcon(fmt.Sprintf("Показать еще (%d)", remaining), iconAdd(), func() {
			step := 200
			if a.cardsVisibleCount <= 0 {
				a.cardsVisibleCount = step
			}
			a.cardsVisibleCount += step
			a.scheduleResultsRender(true)
		})
		// R-12: явное уведомление, что показано не всё.
		counter := widget.NewLabel(fmt.Sprintf("Показано %d из %d", visible, len(data)))
		return container.NewBorder(nil, container.NewHBox(loadMore, counter), nil, nil, list)
	}
	return list
}

// selectHostForDetails выбирает хост для отображения деталей.
func (a *App) selectHostForDetails(r scanner.Result) {
	ip := strings.TrimSpace(r.IP)
	if ip == "" {
		return
	}
	a.selectedHostIP = ip
	a.renderScanResultsView()
}

// selectedHostFromData находит выбранный хост в данных.
func (a *App) selectedHostFromData(data []scanner.Result) (scanner.Result, bool) {
	if len(data) == 0 {
		return scanner.Result{}, false
	}
	selected := strings.TrimSpace(a.selectedHostIP)
	if selected != "" {
		for _, r := range data {
			if strings.TrimSpace(r.IP) == selected {
				return r, true
			}
		}
	}
	a.selectedHostIP = strings.TrimSpace(data[0].IP)
	return data[0], true
}
