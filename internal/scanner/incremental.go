package scanner

import (
	"context"
	"fmt"
	"network-scanner/internal/contracts"
	"sync"
	"time"
)

// ScanEvent представляет событие сканирования, которое отправляется через канал.
type ScanEvent struct {
	// Type определяет тип события: "progress", "host", "summary"
	Type string

	// Stage указывает текущую стадию: "ping", "ports", "complete"
	Stage string

	// Current и Total — текущее и общее количество хостов на стадии
	Current int
	Total   int

	// Result — результат сканирования хоста (только для Type="host")
	Result *Result

	// Message — произвольное сообщение (для прогресса и сводки)
	Message string

	// StartTime — время начала сканирования
	StartTime time.Time

	// Duration — время выполнения (для summary)
	Duration time.Duration
}

// IncrementalScanner обёртка над NetworkScanner для инкрементального вывода.
type IncrementalScanner struct {
	inner *NetworkScanner
}

// NewIncrementalScanner создаёт новый incremental scanner.
func NewIncrementalScanner(ns *NetworkScanner) *IncrementalScanner {
	return &IncrementalScanner{
		inner: ns,
	}
}

// ScanWithEvents запускает сканирование и отправляет события в канал events.
// Возвращает канал, через который отправляются события, и канал ошибок.
// События отправляются асинхронно, поэтому caller должен читать их быстро.
func (s *IncrementalScanner) ScanWithEvents(
	ctx context.Context,
	config contracts.ScanConfig,
) (<-chan ScanEvent, <-chan error) {
	events := make(chan ScanEvent, 100) // Буферизованный канал
	errChan := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errChan)

		startTime := time.Now()

		// Отправляем начальное событие
		events <- ScanEvent{
			Type:      "start",
			Stage:     "init",
			StartTime: startTime,
			Message:   fmt.Sprintf("Начало сканирования сети: %s", s.inner.network),
		}

		// Устанавливаем progress callback, который отправляет события
		s.inner.SetProgressCallback(func(stage string, current, total int, message string) {
			events <- ScanEvent{
				Type:      "progress",
				Stage:     stage,
				Current:   current,
				Total:     total,
				Message:   message,
				StartTime: startTime,
			}
		})

		// Запускаем сканирование
		s.inner.Scan()

		// Отправляем результаты по хостам
		results := s.inner.GetResults()
		for i, result := range results {
			select {
			case events <- ScanEvent{
				Type:      "host",
				Stage:     "complete",
				Current:   i + 1,
				Total:     len(results),
				Result:    &result,
				StartTime: startTime,
				Message:   fmt.Sprintf("Обработан хост %d/%d: %s", i+1, len(results), result.IP),
			}:
			case <-ctx.Done():
				return
			}
		}

		// Отправляем итоговую сводку
		events <- ScanEvent{
			Type:      "summary",
			Stage:     "complete",
			Current:   len(results),
			Total:     len(results),
			Duration:  time.Since(startTime),
			StartTime: startTime,
			Message:   fmt.Sprintf("Сканирование завершено. Найдено устройств: %d", len(results)),
		}

		// Отправляем ошибку, если была
		// (в текущей реализации ошибки логируются, но не возвращаются)
	}()

	return events, errChan
}

// ScanWithEventsAndConfig запускает сканирование с конфигурацией и отправляет события.
func (s *IncrementalScanner) ScanWithEventsAndConfig(
	ctx context.Context,
	config contracts.ScanConfig,
) (<-chan ScanEvent, <-chan error) {
	events := make(chan ScanEvent, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errChan)

		startTime := time.Now()

		// Отправляем начальное событие
		events <- ScanEvent{
			Type:      "start",
			Stage:     "init",
			StartTime: startTime,
			Message:   fmt.Sprintf("Начало сканирования: %s, порты: %s", config.NetworkCIDR, config.PortRange),
		}

		// Устанавливаем progress callback
		s.inner.SetProgressCallback(func(stage string, current, total int, message string) {
			events <- ScanEvent{
				Type:      "progress",
				Stage:     stage,
				Current:   current,
				Total:     total,
				Message:   message,
				StartTime: startTime,
			}
		})

		// Запускаем сканирование
		s.inner.Scan()

		// Отправляем результаты
		results := s.inner.GetResults()
		for i, result := range results {
			select {
			case events <- ScanEvent{
				Type:      "host",
				Stage:     "complete",
				Current:   i + 1,
				Total:     len(results),
				Result:    &result,
				StartTime: startTime,
				Message:   fmt.Sprintf("Обработан хост %d/%d: %s", i+1, len(results), result.IP),
			}:
			case <-ctx.Done():
				return
			}
		}

		// Итоговая сводка
		events <- ScanEvent{
			Type:      "summary",
			Stage:     "complete",
			Current:   len(results),
			Total:     len(results),
			Duration:  time.Since(startTime),
			StartTime: startTime,
			Message:   fmt.Sprintf("Сканирование завершено. Найдено устройств: %d", len(results)),
		}
	}()

	return events, errChan
}

// ConsumeEvents читает события из канала и вызывает handler для каждого.
// Возвращает последнее событие и ошибку.
func ConsumeEvents(
	ctx context.Context,
	events <-chan ScanEvent,
	handler func(ScanEvent) error,
) (ScanEvent, error) {
	var lastEvent ScanEvent

	for {
		select {
		case event, ok := <-events:
			if !ok {
				// Канал закрыт
				return lastEvent, nil
			}

			lastEvent = event

			// Вызываем handler
			if err := handler(event); err != nil {
				return lastEvent, err
			}

			// Если это summary — завершаем
			if event.Type == "summary" {
				return lastEvent, nil
			}

		case <-ctx.Done():
			return lastEvent, ctx.Err()
		}
	}
}

// PrintEventHandler создаёт handler, который печатает события в консоль.
func PrintEventHandler(verbose bool) func(ScanEvent) error {
	return func(event ScanEvent) error {
		switch event.Type {
		case "start":
			fmt.Printf("[START] %s\n", event.Message)

		case "progress":
			// Обновляем строку прогресса без переноса
			if verbose {
				fmt.Printf("[PROGRESS] [%s] %s: %d/%d — %s\n",
					event.Stage, event.Message, event.Current, event.Total,
					time.Since(event.StartTime).Round(time.Millisecond))
			} else {
				fmt.Printf("\r[PROGRESS] [%s] %s: %d/%d",
					event.Stage, event.Message, event.Current, event.Total)
				if event.Current == event.Total {
					fmt.Println() // Перенос строки в конце
				}
			}

		case "host":
			if verbose {
				result := event.Result
				fmt.Printf("[HOST] %s — %s (MAC: %s, Hostname: %s)\n",
					result.IP, result.DeviceType, result.MAC, result.Hostname)
				if len(result.Ports) > 0 {
					fmt.Printf("       Открытые порты: ")
					for i, p := range result.Ports {
						if p.State == "open" {
							if i > 0 {
								fmt.Printf(", ")
							}
							fmt.Printf("%d/%s (%s)", p.Port, p.Protocol, p.Service)
						}
					}
					fmt.Println()
				}
			}

		case "summary":
			fmt.Printf("\n[SUMMARY] %s (время: %v)\n",
				event.Message, event.Duration.Round(time.Millisecond))

		default:
			fmt.Printf("[EVENT] %s\n", event.Message)
		}

		return nil
	}
}

// CollectEventHandler собирает все результаты в слайс.
type CollectEventHandler struct {
	mu       sync.Mutex
	Results  []Result
	Progress []ScanEvent
}

// NewCollectEventHandler создаёт новый handler для сбора результатов.
func NewCollectEventHandler() *CollectEventHandler {
	return &CollectEventHandler{
		Results:  make([]Result, 0),
		Progress: make([]ScanEvent, 0),
	}
}

// Handle обрабатывает событие.
func (h *CollectEventHandler) Handle(event ScanEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch event.Type {
	case "progress":
		h.Progress = append(h.Progress, event)
	case "host":
		if event.Result != nil {
			h.Results = append(h.Results, *event.Result)
		}
	}

	return nil
}

// GetResults возвращает собранные результаты.
func (h *CollectEventHandler) GetResults() []Result {
	h.mu.Lock()
	defer h.mu.Unlock()

	results := make([]Result, len(h.Results))
	copy(results, h.Results)
	return results
}

// GetProgress возвращает собранные события прогресса.
func (h *CollectEventHandler) GetProgress() []ScanEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	progress := make([]ScanEvent, len(h.Progress))
	copy(progress, h.Progress)
	return progress
}
