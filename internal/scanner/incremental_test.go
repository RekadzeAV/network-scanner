package scanner

import (
	"context"
	"fmt"
	"network-scanner/internal/contracts"
	"sync"
	"testing"
	"time"
)

// --- Test ScanEvent ---

func TestScanEventTypes(t *testing.T) {
	events := []string{"start", "progress", "host", "summary"}

	for _, eventType := range events {
		event := ScanEvent{
			Type:    eventType,
			Stage:   "test",
			Message: "test message",
		}

		if event.Type != eventType {
			t.Errorf("ScanEvent.Type = %v, want %v", event.Type, eventType)
		}

		if event.Stage != "test" {
			t.Errorf("ScanEvent.Stage = %v, want test", event.Stage)
		}

		if event.Message != "test message" {
			t.Errorf("ScanEvent.Message = %v, want test message", event.Message)
		}
	}
}

func TestScanEventWithResult(t *testing.T) {
	result := &Result{
		IP:       "192.168.1.1",
		Hostname: "test-host",
		MAC:      "aa:bb:cc:dd:ee:ff",
		IsAlive:  true,
	}

	event := ScanEvent{
		Type:   "host",
		Stage:  "complete",
		Result: result,
	}

	if event.Result == nil {
		t.Fatal("ScanEvent.Result should not be nil")
	}

	if event.Result.IP != "192.168.1.1" {
		t.Errorf("ScanEvent.Result.IP = %v, want 192.168.1.1", event.Result.IP)
	}
}

func TestScanEventWithDuration(t *testing.T) {
	duration := 5 * time.Second

	event := ScanEvent{
		Type:     "summary",
		Duration: duration,
	}

	if event.Duration != duration {
		t.Errorf("ScanEvent.Duration = %v, want %v", event.Duration, duration)
	}
}

// --- Test IncrementalScanner ---

func TestNewIncrementalScanner(t *testing.T) {
	ns := NewNetworkScanner("192.168.1.0/24", time.Second, "1-100", 10, false)
	s := NewIncrementalScanner(ns)

	if s == nil {
		t.Fatal("NewIncrementalScanner() returned nil")
	}

	if s.inner == nil {
		t.Error("IncrementalScanner.inner should not be nil")
	}
}

func TestIncrementalScannerScanWithEvents(t *testing.T) {
	// Используем localhost для быстрого теста
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	s := NewIncrementalScanner(ns)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	events, errChan := s.ScanWithEvents(ctx, contracts.ScanConfig{
		NetworkCIDR: "127.0.0.1/32",
		PortRange:   "1-5",
		Timeout:     200 * time.Millisecond,
		Threads:     5,
	})

	// Читаем события
	eventCount := 0
	for {
		select {
		case event, ok := <-events:
			if !ok {
				// Канал закрыт
				goto done
			}
			eventCount++

			// Проверяем, что событие не пустое
			if event.Type == "" {
				t.Error("ScanEvent.Type should not be empty")
			}
		case err := <-errChan:
			if err != nil {
				t.Errorf("Error from errChan: %v", err)
			}
		case <-ctx.Done():
			t.Error("Timeout waiting for events")
			return
		}
	}

done:
	// Должно быть хотя бы несколько событий
	if eventCount < 2 {
		t.Errorf("Expected at least 2 events, got %d", eventCount)
	}
}

// --- Test ConsumeEvents ---

func TestConsumeEventsEmpty(t *testing.T) {
	ctx := context.Background()
	events := make(chan ScanEvent, 10)
	close(events)

	handler := func(event ScanEvent) error {
		return nil
	}

	lastEvent, err := ConsumeEvents(ctx, events, handler)

	if err != nil {
		t.Errorf("ConsumeEvents() error = %v", err)
	}

	if lastEvent.Type != "" {
		t.Errorf("lastEvent.Type = %v, want empty", lastEvent.Type)
	}
}

func TestConsumeEventsWithEvents(t *testing.T) {
	ctx := context.Background()
	events := make(chan ScanEvent, 10)

	// Отправляем события
	events <- ScanEvent{Type: "start", Message: "start"}
	events <- ScanEvent{Type: "progress", Stage: "ping", Current: 1, Total: 10}
	events <- ScanEvent{Type: "summary", Message: "complete"}
	close(events)

	var received []ScanEvent
	handler := func(event ScanEvent) error {
		received = append(received, event)
		return nil
	}

	lastEvent, err := ConsumeEvents(ctx, events, handler)

	if err != nil {
		t.Errorf("ConsumeEvents() error = %v", err)
	}

	if lastEvent.Type != "summary" {
		t.Errorf("lastEvent.Type = %v, want summary", lastEvent.Type)
	}

	if len(received) != 3 {
		t.Errorf("Received %d events, want 3", len(received))
	}
}

func TestConsumeEventsWithContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan ScanEvent, 10)

	// Отправляем одно событие
	events <- ScanEvent{Type: "start", Message: "start"}

	var received int
	handler := func(event ScanEvent) error {
		received++
		cancel() // Отменяем контекст после первого события
		return nil
	}

	lastEvent, err := ConsumeEvents(ctx, events, handler)

	if err != context.Canceled {
		t.Errorf("ConsumeEvents() error = %v, want context.Canceled", err)
	}

	if received != 1 {
		t.Errorf("Received %d events, want 1", received)
	}

	if lastEvent.Type != "start" {
		t.Errorf("lastEvent.Type = %v, want start", lastEvent.Type)
	}
}

func TestConsumeEventsWithHandlerError(t *testing.T) {
	ctx := context.Background()
	events := make(chan ScanEvent, 10)

	events <- ScanEvent{Type: "start", Message: "start"}
	events <- ScanEvent{Type: "progress", Stage: "ping"}
	close(events)

	expectedErr := fmt.Errorf("handler error")
	handler := func(event ScanEvent) error {
		if event.Type == "progress" {
			return expectedErr
		}
		return nil
	}

	_, err := ConsumeEvents(ctx, events, handler)

	if err != expectedErr {
		t.Errorf("ConsumeEvents() error = %v, want %v", err, expectedErr)
	}
}

// --- Test PrintEventHandler ---

func TestPrintEventHandlerStart(t *testing.T) {
	handler := PrintEventHandler(false)

	event := ScanEvent{
		Type:    "start",
		Message: "Начало сканирования",
	}

	err := handler(event)
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}
}

func TestPrintEventHandlerProgress(t *testing.T) {
	handler := PrintEventHandler(false)

	event := ScanEvent{
		Type:    "progress",
		Stage:   "ping",
		Current: 5,
		Total:   10,
		Message: "Проверено хостов: 5/10",
	}

	err := handler(event)
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}
}

func TestPrintEventHandlerHost(t *testing.T) {
	handler := PrintEventHandler(false)

	result := &Result{
		IP:         "192.168.1.1",
		Hostname:   "test-host",
		MAC:        "aa:bb:cc:dd:ee:ff",
		DeviceType: "Router",
		Ports: []PortInfo{
			{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"},
		},
	}

	event := ScanEvent{
		Type:   "host",
		Result: result,
	}

	err := handler(event)
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}
}

func TestPrintEventHandlerSummary(t *testing.T) {
	handler := PrintEventHandler(false)

	event := ScanEvent{
		Type:     "summary",
		Message:  "Сканирование завершено",
		Duration: 5 * time.Second,
	}

	err := handler(event)
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}
}

func TestPrintEventHandlerVerbose(t *testing.T) {
	handler := PrintEventHandler(true)

	event := ScanEvent{
		Type:    "progress",
		Stage:   "ping",
		Current: 5,
		Total:   10,
		Message: "Проверено хостов: 5/10",
	}

	err := handler(event)
	if err != nil {
		t.Errorf("PrintEventHandler() error = %v", err)
	}
}

// --- Test CollectEventHandler ---

func TestNewCollectEventHandler(t *testing.T) {
	handler := NewCollectEventHandler()

	if handler == nil {
		t.Fatal("NewCollectEventHandler() returned nil")
	}

	if handler.Results == nil {
		t.Error("CollectEventHandler.Results should not be nil")
	}

	if handler.Progress == nil {
		t.Error("CollectEventHandler.Progress should not be nil")
	}
}

func TestCollectEventHandlerHandleProgress(t *testing.T) {
	handler := NewCollectEventHandler()

	event := ScanEvent{
		Type:    "progress",
		Stage:   "ping",
		Current: 5,
		Total:   10,
	}

	err := handler.Handle(event)
	if err != nil {
		t.Errorf("CollectEventHandler.Handle() error = %v", err)
	}

	progress := handler.GetProgress()
	if len(progress) != 1 {
		t.Errorf("Progress length = %d, want 1", len(progress))
	}
}

func TestCollectEventHandlerHandleHost(t *testing.T) {
	handler := NewCollectEventHandler()

	result := &Result{
		IP:       "192.168.1.1",
		Hostname: "test-host",
		MAC:      "aa:bb:cc:dd:ee:ff",
	}

	event := ScanEvent{
		Type:   "host",
		Result: result,
	}

	err := handler.Handle(event)
	if err != nil {
		t.Errorf("CollectEventHandler.Handle() error = %v", err)
	}

	results := handler.GetResults()
	if len(results) != 1 {
		t.Errorf("Results length = %d, want 1", len(results))
	}

	if results[0].IP != "192.168.1.1" {
		t.Errorf("Results[0].IP = %v, want 192.168.1.1", results[0].IP)
	}
}

func TestCollectEventHandlerConcurrent(t *testing.T) {
	handler := NewCollectEventHandler()

	var wg sync.WaitGroup
	eventCount := 100

	// Запускаем множество горутин для параллельной обработки
	for i := 0; i < eventCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			event := ScanEvent{
				Type:    "progress",
				Stage:   "ping",
				Current: id,
				Total:   eventCount,
			}

			err := handler.Handle(event)
			if err != nil {
				t.Errorf("Handle() error = %v", err)
			}
		}(i)
	}

	wg.Wait()

	progress := handler.GetProgress()
	if len(progress) != eventCount {
		t.Errorf("Progress length = %d, want %d", len(progress), eventCount)
	}
}

func TestCollectEventHandlerGetResultsReturnsCopy(t *testing.T) {
	handler := NewCollectEventHandler()

	result := &Result{
		IP: "192.168.1.1",
	}

	event := ScanEvent{
		Type:   "host",
		Result: result,
	}

	handler.Handle(event)

	// Получаем результаты
	results := handler.GetResults()

	// Модифицируем возвращённый слайс
	if len(results) > 0 {
		results[0].IP = "modified"
	}

	// Получаем результаты снова
	results2 := handler.GetResults()

	// Должны быть оригинальные данные
	if results2[0].IP == "modified" {
		t.Error("GetResults() should return a copy, not the original slice")
	}
}

// --- Test ScanConfig ---

func TestScanConfigDefaults(t *testing.T) {
	config := contracts.ScanConfig{}

	if config.NetworkCIDR != "" {
		t.Errorf("NetworkCIDR should be empty by default, got %v", config.NetworkCIDR)
	}

	if config.PortRange != "" {
		t.Errorf("PortRange should be empty by default, got %v", config.PortRange)
	}
}

func TestScanConfigWithValues(t *testing.T) {
	config := contracts.ScanConfig{
		NetworkCIDR: "192.168.1.0/24",
		PortRange:   "1-1000",
		Timeout:     2 * time.Second,
		Threads:     50,
	}

	if config.NetworkCIDR != "192.168.1.0/24" {
		t.Errorf("NetworkCIDR = %v, want 192.168.1.0/24", config.NetworkCIDR)
	}

	if config.PortRange != "1-1000" {
		t.Errorf("PortRange = %v, want 1-1000", config.PortRange)
	}

	if config.Timeout != 2*time.Second {
		t.Errorf("Timeout = %v, want 2s", config.Timeout)
	}

	if config.Threads != 50 {
		t.Errorf("Threads = %v, want 50", config.Threads)
	}
}

// --- Test IncrementalScanner edge cases ---

func TestIncrementalScannerEmptyEvents(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	s := NewIncrementalScanner(ns)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	events, _ := s.ScanWithEvents(ctx, contracts.ScanConfig{
		NetworkCIDR: "127.0.0.1/32",
		PortRange:   "1-5",
		Timeout:     200 * time.Millisecond,
		Threads:     5,
	})

	// Должно получиться хотя бы событие start
	eventCount := 0
	for range events {
		eventCount++
	}

	if eventCount < 1 {
		t.Errorf("Expected at least 1 event, got %d", eventCount)
	}
}

func TestIncrementalScannerContextCancel(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	s := NewIncrementalScanner(ns)

	ctx, cancel := context.WithCancel(context.Background())

	events, _ := s.ScanWithEvents(ctx, contracts.ScanConfig{
		NetworkCIDR: "127.0.0.1/32",
		PortRange:   "1-5",
		Timeout:     200 * time.Millisecond,
		Threads:     5,
	})

	// Отменяем контекст сразу
	cancel()

	// Читаем события до закрытия канала
	eventCount := 0
	for range events {
		eventCount++
	}

	// Должно быть хотя бы событие start
	if eventCount < 1 {
		t.Errorf("Expected at least 1 event after cancel, got %d", eventCount)
	}
}

// --- Benchmark ---

func BenchmarkPrintEventHandlerStart(b *testing.B) {
	handler := PrintEventHandler(false)
	event := ScanEvent{
		Type:    "start",
		Message: "Начало сканирования",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler(event)
	}
}

func BenchmarkPrintEventHandlerProgress(b *testing.B) {
	handler := PrintEventHandler(false)
	event := ScanEvent{
		Type:    "progress",
		Stage:   "ping",
		Current: 50,
		Total:   100,
		Message: "Проверено хостов: 50/100",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler(event)
	}
}

func BenchmarkCollectEventHandlerHandle(b *testing.B) {
	handler := NewCollectEventHandler()
	result := &Result{
		IP: "192.168.1.1",
	}
	event := ScanEvent{
		Type:   "host",
		Result: result,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.Handle(event)
	}
}
