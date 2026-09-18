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
type tableCellCache struct {
	cache map[TableCellCacheKey]string
	mu    sync.RWMutex
}

// TableCellCacheKey уникальный ключ для кэширования ячейки.
type TableCellCacheKey struct {
	Row int
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

// buildTableView создаёт таблицу результатов сканирования.
func (a *App) buildTableView(data []scanner.Result) fyne.CanvasObject {
	rows := len(data) + 1
	cols := 8
	headers := a.resultsTableHeaders()

	// Создаём кэш для ячеек
	cellCache := newTableCellCache()

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
			l.TextStyle = fyne.TextStyle{}
			if id.Row == 0 {
				l.TextStyle = fyne.TextStyle{Bold: true}
				if id.Col < len(headers) {
					l.SetText(headers[id.Col])
				}
				return
			}
			r := data[id.Row-1]

			// Используем кэш для вычислений
			cellKey := TableCellCacheKey{Row: id.Row, Col: id.Col, IP: r.IP}
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
	widths := a.resultsTableColumnWidths()
	for col, width := range widths {
		t.SetColumnWidth(col, width)
	}
	t.OnSelected = func(id widget.TableCellID) {
		if id.Row <= 0 || id.Row-1 >= len(data) {
			return
		}
		a.selectHostForDetails(data[id.Row-1])
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
		return []string{"Host", "IP", "MAC", "Тип", "Вендор", "OS", "SNMP", "Порты"}
	}
	return []string{"Host", "IP", "MAC", "Тип", "Производитель", "ОС (оценка)", "SNMP", "Порты (открытые)"}
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
		return container.NewBorder(nil, loadMore, nil, nil, list)
	}
	return list
}

// selectHostForDetails выбирает хост для отображения деталей.
func (a *App) selectHostForDetails(r scanner.Result) {
	ip := strings.TrimSpace(r.IP)
	if ip == "" {
		return
	}
	a.primeHostDetailsCache(r)
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
