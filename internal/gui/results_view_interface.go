package gui

import (
	"network-scanner/internal/scanner"

	"fyne.io/fyne/v2"
)

// ResultsView интерфейс для отображения результатов сканирования.
// Позволяет заменять реализацию (таблица/карточки) без изменения кода App.
type ResultsView interface {
	// Render рендерит данные результатов в Fyne CanvasObject
	Render(data []scanner.Result) fyne.CanvasObject

	// UpdateMode обновляет режим отображения (таблица/карточки/Security/Inventory)
	UpdateMode(mode string)

	// GetSelectedHost возвращает выбранный хост для отображения деталей
	GetSelectedHost() (scanner.Result, bool)

	// SetOnHostSelected устанавливает обработчик выбора хоста
	SetOnHostSelected(func(scanner.Result))
}

// resultsViewImpl реализация ResultsView для стандартного режима (таблица/карточки).
type resultsViewImpl struct {
	app          *App
	mode         string
	onHostSelect func(scanner.Result)
}

// NewResultsView создаёт новую реализацию ResultsView.
func NewResultsView(app *App) ResultsView {
	return &resultsViewImpl{
		app:  app,
		mode: "таблица",
	}
}

// Render рендерит данные результатов.
func (rv *resultsViewImpl) Render(data []scanner.Result) fyne.CanvasObject {
	if rv.app == nil {
		return nil
	}
	switch rv.mode {
	case "карточки":
		return rv.app.buildCardsView(data)
	case "Security":
		return rv.app.buildSecurityDashboardView(data)
	case "Inventory":
		return rv.app.buildInventoryDashboardView()
	default:
		return rv.app.buildTableView(data)
	}
}

// UpdateMode обновляет режим отображения.
func (rv *resultsViewImpl) UpdateMode(mode string) {
	if rv == nil {
		return
	}
	rv.mode = mode
}

// GetSelectedHost возвращает выбранный хост.
func (rv *resultsViewImpl) GetSelectedHost() (scanner.Result, bool) {
	if rv == nil || rv.app == nil {
		return scanner.Result{}, false
	}
	return rv.app.selectedHostFromData(nil)
}

// SetOnHostSelected устанавливает обработчик выбора хоста.
func (rv *resultsViewImpl) SetOnHostSelected(fn func(scanner.Result)) {
	if rv == nil {
		return
	}
	rv.onHostSelect = fn
}

// inventoryViewImpl реализация ResultsView для Inventory режима.
type inventoryViewImpl struct {
	app *App
}

// NewInventoryView создаёт новую реализацию Inventory.
func NewInventoryView(app *App) ResultsView {
	return &inventoryViewImpl{app: app}
}

// Render для Inventory всегда возвращает dashboard view.
func (iv *inventoryViewImpl) Render(_ []scanner.Result) fyne.CanvasObject {
	if iv.app == nil {
		return nil
	}
	return iv.app.buildInventoryDashboardView()
}

// UpdateMode для Inventory игнорируется.
func (iv *inventoryViewImpl) UpdateMode(_ string) {}

// GetSelectedHost для Inventory.
func (iv *inventoryViewImpl) GetSelectedHost() (scanner.Result, bool) {
	if iv == nil || iv.app == nil {
		return scanner.Result{}, false
	}
	return iv.app.selectedHostFromData(nil)
}

// SetOnHostSelected для Inventory.
func (iv *inventoryViewImpl) SetOnHostSelected(_ func(scanner.Result)) {}

// securityViewImpl реализация ResultsView для Security режима.
type securityViewImpl struct {
	app *App
}

// NewSecurityView создаёт новую реализацию Security.
func NewSecurityView(app *App) ResultsView {
	return &securityViewImpl{app: app}
}

// Render для Security.
func (sv *securityViewImpl) Render(data []scanner.Result) fyne.CanvasObject {
	if sv.app == nil {
		return nil
	}
	return sv.app.buildSecurityDashboardView(data)
}

// UpdateMode для Security игнорируется.
func (sv *securityViewImpl) UpdateMode(_ string) {}

// GetSelectedHost для Security.
func (sv *securityViewImpl) GetSelectedHost() (scanner.Result, bool) {
	if sv == nil || sv.app == nil {
		return scanner.Result{}, false
	}
	return sv.app.selectedHostFromData(nil)
}

// SetOnHostSelected для Security.
func (sv *securityViewImpl) SetOnHostSelected(_ func(scanner.Result)) {}
