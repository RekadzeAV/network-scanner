package gui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/display"
	"network-scanner/internal/logger"
	"network-scanner/internal/network"
	"network-scanner/internal/scanner"
)

// scanUpdate содержит результаты сканирования для обновления UI
type scanUpdate struct {
	results []scanner.Result
}

// progressUpdate содержит информацию о прогрессе сканирования
type progressUpdate struct {
	stage   string
	current int
	total   int
	message string
	percent float64
}

// App представляет GUI приложение
type App struct {
	myApp          fyne.App
	myWindow       fyne.Window
	scanResults    []scanner.Result
	networkScanner *scanner.NetworkScanner
	networkEntry   *widget.Entry
	statusLabel    *widget.Label
	stageLabel     *widget.Label
	progressBar    *widget.ProgressBar
	resultsText    *widget.RichText
	resultsScroll  *container.Scroll
	scanButton     *widget.Button
	saveButton     *widget.Button
}

// createAppIcon создает иконку приложения
func createAppIcon() fyne.Resource {
	// Создаем простое изображение иконки программно
	// Иконка 64x64 пикселя с простым дизайном сетевого сканера
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	// Рисуем простую иконку: синий фон с белым символом сети
	bgColor := color.RGBA{R: 0, G: 100, B: 200, A: 255}
	iconColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Рисуем простой символ сети (круг с линиями)
	centerX, centerY := 32.0, 32.0
	radius := 20.0

	// Рисуем круг
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist >= radius-2 && dist <= radius+2 {
				img.Set(x, y, iconColor)
			}
		}
	}

	// Рисуем линии от центра
	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi * 2 / 8
		for r := radius + 3; r < 30; r++ {
			x := int(centerX + r*math.Cos(angle))
			y := int(centerY + r*math.Sin(angle))
			if x >= 0 && x < 64 && y >= 0 && y < 64 {
				img.Set(x, y, iconColor)
			}
		}
	}

	// Конвертируем изображение в PNG
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		// Если не удалось создать иконку, возвращаем nil
		return nil
	}

	return fyne.NewStaticResource("icon.png", buf.Bytes())
}

// NewApp создает новый экземпляр GUI приложения
func NewApp() *App {
	myApp := app.NewWithID("network-scanner")

	myWindow := myApp.NewWindow("Network Scanner - Сканер локальной сети")

	// Устанавливаем иконку приложения
	if icon := createAppIcon(); icon != nil {
		myApp.SetIcon(icon)
		myWindow.SetIcon(icon)
	}

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

	logger.Log("GUI приложение инициализировано")

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

	// Метка этапа сканирования
	a.stageLabel = widget.NewLabel("")
	a.stageLabel.Wrapping = fyne.TextWrapWord
	a.stageLabel.Hide()

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
		a.stageLabel,
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
		// Обновляем UI через fyne.Do для выполнения в главном потоке
		fyne.Do(func() {
			if err == nil && networkStr != "" {
				a.networkEntry.SetText(networkStr)
				a.statusLabel.SetText(fmt.Sprintf("Сеть определена автоматически: %s", networkStr))
			} else {
				// Если не удалось определить, оставляем поле пустым
				a.statusLabel.SetText("Готов к сканированию (сеть будет определена автоматически при запуске)")
			}
			// Обновляем виджеты
			a.statusLabel.Refresh()
			a.networkEntry.Refresh()
		})
	}()
}

// startScan запускает процесс сканирования
func (a *App) startScan() {
	scanStartTime := time.Now()
	logger.Log("Запуск сканирования из GUI")
	logger.LogDebug("Пользователь нажал кнопку 'Запустить сканирование'")

	// Определяем сеть
	networkStr := a.networkEntry.Text
	if networkStr == "" {
		logger.Log("Автоматическое определение сети...")
		logger.LogDebug("Поле сети пустое, начинаем автоматическое определение")
		detectStartTime := time.Now()
		var err error
		networkStr, err = network.DetectLocalNetwork()
		detectDuration := time.Since(detectStartTime)
		if err != nil {
			logger.LogError(err, "Определение сети в GUI")
			logger.LogDebug("Автоматическое определение сети завершилось ошибкой за %v", detectDuration)
			dialog.ShowError(fmt.Errorf("не удалось определить сеть: %v", err), a.myWindow)
			return
		}
		a.networkEntry.SetText(networkStr)
		logger.Log("Определена сеть: %s (определение заняло %v)", networkStr, detectDuration)
		logger.LogDebug("Автоматическое определение сети завершено успешно")
	} else {
		logger.Log("Использована указанная сеть: %s", networkStr)
		logger.LogDebug("Сеть указана пользователем в поле ввода")
	}

	// Отключаем кнопку сканирования
	a.scanButton.Disable()
	a.saveButton.Disable()
	a.progressBar.Show()
	a.progressBar.SetValue(0)
	a.stageLabel.Show()
	a.statusLabel.SetText("Сканирование запущено...")
	a.stageLabel.SetText("Инициализация...")

	// Очищаем предыдущие результаты
	a.resultsText.ParseMarkdown("## Сканирование запущено...\n\nПожалуйста, подождите.")
	a.resultsScroll.Refresh()

	// Создаем канал для передачи результатов из горутины
	resultsChan := make(chan []scanner.Result, 1)
	progressChan := make(chan progressUpdate, 100) // Буферизованный канал для прогресса

	// Запускаем сканирование в отдельной горутине
	go func() {
		// Создаем сканер
		timeout := 2 * time.Second
		portRange := "1-1000"
		threads := 50
		showClosed := false

		logger.LogDebug("Создание сканера в GUI: сеть=%s, порты=%s, таймаут=%v, потоков=%d, showClosed=%v",
			networkStr, portRange, timeout, threads, showClosed)
		ns := scanner.NewNetworkScanner(networkStr, timeout, portRange, threads, showClosed)

		// Устанавливаем callback для прогресса
		ns.SetProgressCallback(func(stage string, current int, total int, message string) {
			// Вычисляем процент прогресса для текущего этапа (0-100%)
			var percent float64
			if total > 0 {
				percent = float64(current) / float64(total)
			} else {
				percent = 0
			}

			// Отправляем обновление прогресса в канал
			select {
			case progressChan <- progressUpdate{
				stage:   stage,
				current: current,
				total:   total,
				message: message,
				percent: percent,
			}:
			default:
				// Пропускаем если канал переполнен (не критично)
			}
		})

		// Запускаем сканирование
		logger.LogDebug("Запуск метода Scan() в GUI")
		scanMethodStartTime := time.Now()
		ns.Scan()
		scanMethodDuration := time.Since(scanMethodStartTime)
		logger.LogDebug("Метод Scan() завершен в GUI за %v", scanMethodDuration)

		// Получаем результаты
		results := ns.GetResults()
		logger.Log("Получено результатов сканирования в GUI: %d", len(results))
		logger.LogDebug("Результаты сканирования получены из сканера")

		// Закрываем канал прогресса
		close(progressChan)
		logger.LogDebug("Канал прогресса закрыт")

		// Отправляем результаты в канал
		resultsChan <- results
		totalDuration := time.Since(scanStartTime)
		logger.Log("Сканирование в GUI завершено за %v, найдено устройств: %d", totalDuration, len(results))
	}()

	// Обрабатываем результаты и прогресс в отдельной горутине
	go func() {
		// Создаем тикер для периодического обновления UI
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		timeout := time.NewTimer(300 * time.Second)
		defer timeout.Stop()

		for {
			select {
			case progress, ok := <-progressChan:
				if !ok {
					// Канал прогресса закрыт, продолжаем ждать результаты
					progressChan = nil
					continue
				}
				// Обновляем UI через fyne.Do для выполнения в главном потоке
				fyne.Do(func() {
					// Определяем название этапа на русском
					stageName := ""
					switch progress.stage {
					case "ping":
						stageName = "Этап 1: Проверка доступности хостов"
					case "ports":
						stageName = "Этап 2: Сканирование портов"
					case "complete":
						stageName = "Завершение"
					default:
						stageName = "Сканирование"
					}

					// Обновляем прогресс-бар (0-100% для текущего этапа)
					a.progressBar.SetValue(progress.percent)

					// Формируем текст для метки этапа с процентом
					if progress.total > 0 {
						percentText := fmt.Sprintf("%.1f%%", progress.percent*100)
						a.stageLabel.SetText(fmt.Sprintf("%s: %d/%d (%s)", stageName, progress.current, progress.total, percentText))
					} else {
						a.stageLabel.SetText(stageName)
					}

					// Обновляем статус
					a.statusLabel.SetText(progress.message)

					// Обновляем виджеты
					a.progressBar.Refresh()
					a.stageLabel.Refresh()
					a.statusLabel.Refresh()
				})

			case results, ok := <-resultsChan:
				if !ok {
					return
				}
				// Обновляем UI через fyne.Do для выполнения в главном потоке
				fyne.Do(func() {
					a.scanResults = results
					a.progressBar.SetValue(1.0)
					a.progressBar.Hide()
					a.stageLabel.Hide()

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
					a.stageLabel.Refresh()
					a.statusLabel.Refresh()
					a.scanButton.Enable()
					a.myWindow.Content().Refresh()
				})
				return

			case <-ticker.C:
				// Периодически обновляем UI для обеспечения отзывчивости
				if progressChan != nil {
					fyne.Do(func() {
						a.progressBar.Refresh()
						a.stageLabel.Refresh()
						a.statusLabel.Refresh()
					})
				}

			case <-timeout.C:
				// Таймаут на случай если сканирование зависло
				fyne.Do(func() {
					a.statusLabel.SetText("Таймаут сканирования")
					a.stageLabel.Hide()
					a.scanButton.Enable()
					a.progressBar.Hide()
					a.statusLabel.Refresh()
					a.stageLabel.Refresh()
					a.progressBar.Refresh()
				})
				return
			}
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
