package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/display"
	"network-scanner/internal/network"
	"network-scanner/internal/scanner"
)

// App представляет GUI приложение
type App struct {
	myApp          fyne.App
	myWindow       fyne.Window
	scanResults    []scanner.Result
	networkScanner *scanner.NetworkScanner
	networkEntry   *widget.Entry
	statusLabel    *widget.Label
	progressBar    *widget.ProgressBar
	resultsText    *widget.RichText
	resultsScroll  *container.Scroll
	scanButton     *widget.Button
	saveButton     *widget.Button
}

// NewApp создает новый экземпляр GUI приложения
func NewApp() *App {
	myApp := app.NewWithID("network-scanner")
	myApp.SetMetadata(&fyne.AppMetadata{
		Title:   "Network Scanner",
		Version: "1.0.0",
	})

	myWindow := myApp.NewWindow("Network Scanner - Сканер локальной сети")
	myWindow.Resize(fyne.NewSize(1000, 700))
	myWindow.CenterOnScreen()

	app := &App{
		myApp:    myApp,
		myWindow: myWindow,
	}

	app.initUI()
	app.setupEventHandlers()

	return app
}

// initUI инициализирует все элементы интерфейса
func (a *App) initUI() {
	// Поле ввода сети
	networkLabel := widget.NewLabel("Сеть (CIDR, например 192.168.1.0/24):")
	networkLabel.Wrapping = fyne.TextWrapWord
	a.networkEntry = widget.NewEntry()
	a.networkEntry.SetPlaceHolder("Оставьте пустым для автоматического определения")

	// Кнопка сканирования
	a.scanButton = widget.NewButton("Запустить сканирование", nil)
	a.scanButton.Importance = widget.HighImportance

	// Кнопка сохранения
	a.saveButton = widget.NewButton("Сохранить результаты", nil)
	a.saveButton.Disable()

	// Статус
	a.statusLabel = widget.NewLabel("Готов к сканированию")
	a.statusLabel.Wrapping = fyne.TextWrapWord

	// Прогресс-бар
	a.progressBar = widget.NewProgressBar()
	a.progressBar.Hide()

	// Область результатов с прокруткой
	a.resultsText = widget.NewRichText()
	a.resultsText.Wrapping = fyne.TextWrapWord
	
	// Создаем прокручиваемый контейнер для результатов
	// Это ключевое изменение - используем Scroll контейнер для прокрутки результатов
	a.resultsScroll = container.NewScroll(a.resultsText)
	// Устанавливаем минимальный размер, чтобы контейнер был прокручиваемым
	// и занимал доступное пространство
	a.resultsScroll.SetMinSize(fyne.NewSize(0, 300))

	// Верхняя панель с элементами управления
	controlsContainer := container.NewVBox(
		networkLabel,
		a.networkEntry,
		container.NewHBox(a.scanButton, a.saveButton),
		a.statusLabel,
		a.progressBar,
	)

	// Разделитель
	separator := widget.NewSeparator()

	// Заголовок результатов
	resultsLabel := widget.NewLabel("Результаты сканирования:")
	resultsLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Контейнер с результатами (заголовок + прокручиваемая область)
	resultsContainer := container.NewBorder(
		container.NewVBox(separator, resultsLabel),
		nil,
		nil,
		nil,
		a.resultsScroll, // Прокручиваемый контейнер с результатами
	)

	// Основной контейнер - используем Border layout для правильного размещения
	// Верх: элементы управления
	// Центр: результаты с прокруткой (занимают все доступное пространство)
	content := container.NewBorder(
		controlsContainer,
		nil,
		nil,
		nil,
		resultsContainer,
	)

	a.myWindow.SetContent(content)
}

// setupEventHandlers настраивает обработчики событий
func (a *App) setupEventHandlers() {
	a.scanButton.OnTapped = func() {
		a.startScan()
	}

	a.saveButton.OnTapped = func() {
		a.saveResults()
	}
}

// startScan запускает процесс сканирования
func (a *App) startScan() {
	// Определяем сеть
	networkStr := a.networkEntry.Text
	if networkStr == "" {
		var err error
		networkStr, err = network.DetectLocalNetwork()
		if err != nil {
			dialog.ShowError(fmt.Errorf("не удалось определить сеть: %v", err), a.myWindow)
			return
		}
		a.networkEntry.SetText(networkStr)
	}

	// Отключаем кнопку сканирования
	a.scanButton.Disable()
	a.saveButton.Disable()
	a.progressBar.Show()
	a.progressBar.SetValue(0)
	a.statusLabel.SetText("Сканирование запущено...")
	
	// Очищаем предыдущие результаты
	a.resultsText.ParseMarkdown("## Сканирование запущено...\n\nПожалуйста, подождите.")
	a.resultsScroll.Refresh()

	// Запускаем сканирование в отдельной горутине
	go func() {
		// Создаем сканер
		timeout := 2 * time.Second
		portRange := "1-1000"
		threads := 50
		showClosed := false

		ns := scanner.NewNetworkScanner(networkStr, timeout, portRange, threads, showClosed)

		// Запускаем сканирование
		ns.Scan()

		// Получаем результаты
		results := ns.GetResults()

		// Обновляем UI в главном потоке
		a.myApp.Driver().RunInMain(func() {
			a.scanResults = results
			a.progressBar.SetValue(1.0)
			a.progressBar.Hide()
			
			if len(results) == 0 {
				a.statusLabel.SetText("Сканирование завершено. Результаты не найдены.")
				a.resultsText.ParseMarkdown("## Результаты сканирования\n\nРезультаты сканирования не найдены.")
			} else {
				a.statusLabel.SetText(fmt.Sprintf("Сканирование завершено. Найдено устройств: %d", len(results)))
				formattedResults := FormatResultsForDisplay(results)
				a.resultsText.ParseMarkdown(formattedResults)
				a.saveButton.Enable()
			}
			
			// Прокручиваем к началу результатов и обновляем отображение
			a.resultsScroll.ScrollToTop()
			a.resultsScroll.Refresh()
			a.resultsText.Refresh()
			
			a.scanButton.Enable()
		})
	}()
}

// saveResults сохраняет результаты в файл
func (a *App) saveResults() {
	if len(a.scanResults) == 0 {
		dialog.ShowInformation("Информация", "Нет результатов для сохранения", a.myWindow)
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.myWindow)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// Форматируем результаты в текстовый формат
		text := display.FormatResultsAsText(a.scanResults)
		
		_, err = writer.Write([]byte(text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка при сохранении файла: %v", err), a.myWindow)
			return
		}

		dialog.ShowInformation("Успех", "Результаты успешно сохранены", a.myWindow)
	}, a.myWindow)
}

// Run запускает GUI приложение
func (a *App) Run() {
	a.myWindow.ShowAndRun()
}

