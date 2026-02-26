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

// scanUpdate содержит результаты сканирования для обновления UI
type scanUpdate struct {
	results []scanner.Result
}

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

	myWindow := myApp.NewWindow("Network Scanner - Сканер локальной сети")
	
	// Устанавливаем размер окна, который гарантированно поместится на экране
	// Используем консервативные размеры, которые подходят даже для маленьких экранов
	// (например, ноутбуки с разрешением 1366x768)
	width := float32(950)
	height := float32(700)
	
	myWindow.Resize(fyne.NewSize(width, height))
	myWindow.CenterOnScreen()
	
	// Устанавливаем максимальный размер окна, чтобы оно не выходило за границы экрана
	// Fyne автоматически ограничит размер окна размером экрана
	myWindow.SetFixedSize(false) // Позволяем изменять размер, но в пределах экрана

	app := &App{
		myApp:    myApp,
		myWindow: myWindow,
	}

	app.initUI()
	app.setupEventHandlers()
	app.autoDetectNetwork()

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

// autoDetectNetwork автоматически определяет сеть и заполняет поле ввода
func (a *App) autoDetectNetwork() {
	// Запускаем определение сети в отдельной горутине, чтобы не блокировать UI
	go func() {
		networkStr, err := network.DetectLocalNetwork()
		// Обновляем UI в главном потоке через RunOnMainThread
		a.myApp.Driver().RunOnMainThread(func() {
			if err == nil && networkStr != "" {
				// Обновляем UI в главном потоке
				a.networkEntry.SetText(networkStr)
				a.statusLabel.SetText(fmt.Sprintf("Сеть определена автоматически: %s", networkStr))
				a.statusLabel.Refresh()
				a.networkEntry.Refresh()
			} else {
				// Если не удалось определить, оставляем поле пустым
				a.statusLabel.SetText("Готов к сканированию (сеть будет определена автоматически при запуске)")
				a.statusLabel.Refresh()
			}
		})
	}()
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

	// Создаем канал для передачи результатов из горутины
	resultsChan := make(chan []scanner.Result, 1)

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

		// Отправляем результаты в канал
		resultsChan <- results
	}()

	// Обрабатываем результаты в отдельной горутине с периодической проверкой
	go func() {
		select {
		case results := <-resultsChan:
			// Обновляем UI в главном потоке через Canvas Refresh
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
			a.progressBar.Refresh()
			a.statusLabel.Refresh()
			a.scanButton.Enable()
			a.myWindow.Content().Refresh()
		case <-time.After(300 * time.Second):
			// Таймаут на случай если сканирование зависло
			a.statusLabel.SetText("Таймаут сканирования")
			a.scanButton.Enable()
		}
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
