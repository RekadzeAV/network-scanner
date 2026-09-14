package eventbus

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================================
// C2: Тесты для Event Bus
// ============================================================================

// TestEventBus_New — ветка: создание EventBus
func TestEventBus_New(t *testing.T) {
	bus := NewEventBus()
	if bus == nil {
		t.Fatal("expected non-nil EventBus")
	}
	if bus.IsClosed() {
		t.Error("expected bus to not be closed")
	}
}

// TestEventBus_Subscribe_Publish — ветка: подписка и публикация
func TestEventBus_Subscribe_Publish(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	bus.Subscribe("test.event", func(event Event) {
		if event.Type() != "test.event" {
			t.Errorf("expected event type test.event, got %s", event.Type())
		}
		wg.Done()
	})

	bus.Publish(NewBaseEvent("test.event"))
	wg.Wait()
}

// TestEventBus_MultipleHandlers — ветка: несколько обработчиков
func TestEventBus_MultipleHandlers(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	bus.Subscribe("test.event", func(event Event) {
		wg.Done()
	})
	bus.Subscribe("test.event", func(event Event) {
		wg.Done()
	})

	bus.Publish(NewBaseEvent("test.event"))
	wg.Wait()
}

// TestEventBus_SubscribeAny — ветка: подписка на все события
func TestEventBus_SubscribeAny(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(3)
	bus.SubscribeAny(func(event Event) {
		wg.Done()
	})

	bus.Publish(NewBaseEvent("event1"))
	bus.Publish(NewBaseEvent("event2"))
	bus.Publish(NewBaseEvent("event3"))

	wg.Wait()
}

// TestEventBus_HandleOnce — ветка: обработка один раз
func TestEventBus_HandleOnce(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	bus.HandleOnce("test.event", func(event Event) {
		wg.Done()
	})

	bus.Publish(NewBaseEvent("test.event"))
	bus.Publish(NewBaseEvent("test.event"))

	wg.Wait()
}

// TestEventBus_HandleOnce_FiresExactlyOnce — регрессия: колбэк обязан
// отработать один раз, даже если следующие публикации застали обработчик
// до его удаления из списка подписчиков.
func TestEventBus_HandleOnce_FiresExactlyOnce(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var calls int32
	called := make(chan struct{})

	bus.HandleOnce("once.event", func(event Event) {
		if atomic.AddInt32(&calls, 1) == 1 {
			close(called)
		}
	})

	// Быстрые публикации подряд — без sync.Once вторая и последующие
	// публикации вызывают обработчик повторно.
	for i := 0; i < 20; i++ {
		bus.Publish(NewBaseEvent("once.event"))
	}

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("HandleOnce handler was never called")
	}

	// Обработчики выполняются асинхронно: проверяем, что счётчик
	// останавливается на единице и не ползёт вверх.
	deadline := time.Now().Add(time.Second)
	for {
		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Fatalf("HandleOnce fired %d times, want 1", got)
		}
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestEventBus_Unsubscribe — ветка: отписка
func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	handler := func(event Event) {
		wg.Done()
	}

	bus.Subscribe("test.event", handler)
	bus.Publish(NewBaseEvent("test.event"))
	wg.Wait()

	bus.Unsubscribe("test.event")

	// После отписки события не должны обрабатываться
	bus.Publish(NewBaseEvent("test.event"))
}

// TestEventBus_UnsubscribeHandler — ветка: отписка конкретного обработчика
func TestEventBus_UnsubscribeHandler(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	handler := func(event Event) {
		wg.Done()
	}

	bus.Subscribe("test.event", handler)
	bus.Subscribe("test.event", func(event Event) {})
	bus.Publish(NewBaseEvent("test.event"))
	wg.Wait()

	bus.UnsubscribeHandler("test.event", handler)
	bus.Publish(NewBaseEvent("test.event"))
}

// TestEventBus_PanicProtection — ветка: защита от panic в обработчике
func TestEventBus_PanicProtection(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	// Обработчик который паникует
	bus.Subscribe("test.event", func(event Event) {
		panic("test panic")
	})

	// Обычный обработчик
	var wg sync.WaitGroup
	wg.Add(1)
	bus.Subscribe("test.event", func(event Event) {
		wg.Done()
	})

	// Публикация не должна упасть
	bus.Publish(NewBaseEvent("test.event"))
	wg.Wait()
}

// TestEventBus_Close — ветка: закрытие EventBus
func TestEventBus_Close(t *testing.T) {
	bus := NewEventBus()

	bus.Close()

	if !bus.IsClosed() {
		t.Error("expected bus to be closed")
	}

	// Публикация после закрытия должна игнорироваться
	bus.Publish(NewBaseEvent("test.event"))
}

// TestEventBus_SubscriberCount — ветка: подсчёт подписчиков
func TestEventBus_SubscriberCount(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	bus.Subscribe("event1", func(event Event) {})
	bus.Subscribe("event1", func(event Event) {})
	bus.Subscribe("event2", func(event Event) {})

	if bus.SubscriberCount("event1") != 2 {
		t.Errorf("expected 2 subscribers for event1")
	}
	if bus.SubscriberCount("event2") != 1 {
		t.Errorf("expected 1 subscriber for event2")
	}
	if bus.SubscriberCount("event3") != 0 {
		t.Errorf("expected 0 subscribers for event3")
	}
}

// TestEventBus_TotalSubscriberCount — ветка: общий подсчёт подписчиков
func TestEventBus_TotalSubscriberCount(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	bus.Subscribe("event1", func(event Event) {})
	bus.Subscribe("event1", func(event Event) {})
	bus.Subscribe("event2", func(event Event) {})
	bus.SubscribeAny(func(event Event) {})

	total := bus.TotalSubscriberCount()
	if total != 4 {
		t.Errorf("expected 4 total subscribers, got %d", total)
	}
}

// TestEventBus_PublishBatch — ветка: batch публикация
func TestEventBus_PublishBatch(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(3)
	bus.Subscribe("batch.event", func(event Event) {
		wg.Done()
	})

	events := []Event{
		NewBaseEvent("batch.event"),
		NewBaseEvent("batch.event"),
		NewBaseEvent("batch.event"),
	}

	bus.PublishBatch(events)
	wg.Wait()
}

// TestBaseEvent — ветка: базовое событие
func TestBaseEvent(t *testing.T) {
	event := NewBaseEvent("test.event")

	if event.Type() != "test.event" {
		t.Errorf("expected type test.event, got %s", event.Type())
	}
	if event.Timestamp().IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if event.Payload() == nil {
		t.Error("expected non-nil payload")
	}
}

// TestBaseEvent_Payload — ветка: payload события
func TestBaseEvent_Payload(t *testing.T) {
	event := NewBaseEvent("test.event")
	event.SetPayload("key", "value")

	val, ok := event.GetPayload("key")
	if !ok {
		t.Error("expected to find key in payload")
	}
	if val != "value" {
		t.Errorf("expected value 'value', got %v", val)
	}

	_, ok = event.GetPayload("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent key")
	}
}

// TestScannerEvents — ветка: события сканера
func TestScannerEvents(t *testing.T) {
	// ScanStartedEvent
	scanStarted := NewScanStartedEvent("192.168.1.0/24", "2s")
	if scanStarted.Type() != "scan.started" {
		t.Error("expected scan.started event type")
	}
	if scanStarted.Network != "192.168.1.0/24" {
		t.Errorf("expected network 192.168.1.0/24, got %s", scanStarted.Network)
	}

	// ScanCompletedEvent
	scanCompleted := NewScanCompletedEvent(10, 25, "5m")
	if scanCompleted.Type() != "scan.completed" {
		t.Error("expected scan.completed event type")
	}
	if scanCompleted.HostCount != 10 {
		t.Errorf("expected 10 hosts, got %d", scanCompleted.HostCount)
	}

	// HostFoundEvent
	hostFound := NewHostFoundEvent("192.168.1.1", "aa:bb:cc:dd:ee:ff", "router")
	if hostFound.Type() != "host.found" {
		t.Error("expected host.found event type")
	}
	if hostFound.IP != "192.168.1.1" {
		t.Error("expected IP 192.168.1.1")
	}

	// PortOpenedEvent
	portOpened := NewPortOpenedEvent("192.168.1.1", 80, "tcp", "http")
	if portOpened.Type() != "port.opened" {
		t.Error("expected port.opened event type")
	}
	if portOpened.Port != 80 {
		t.Errorf("expected port 80, got %d", portOpened.Port)
	}

	// DeviceDetectedEvent
	deviceDetected := NewDeviceDetectedEvent("192.168.1.1", "router", "cisco", "ios")
	if deviceDetected.Type() != "device.detected" {
		t.Error("expected device.detected event type")
	}
	if deviceDetected.DeviceType != "router" {
		t.Error("expected device type router")
	}
}

// TestPluginEvents — ветка: события плагинов
func TestPluginEvents(t *testing.T) {
	// ProbeStartedEvent
	probeStarted := NewProbeStartedEvent("icmp-ping", "192.168.1.1", 0)
	if probeStarted.Type() != "probe.started" {
		t.Error("expected probe.started event type")
	}
	if probeStarted.ProbeName != "icmp-ping" {
		t.Error("expected probe name icmp-ping")
	}

	// ProbeCompletedEvent
	probeCompleted := NewProbeCompletedEvent("icmp-ping", "192.168.1.1", 0, "success")
	if probeCompleted.Type() != "probe.completed" {
		t.Error("expected probe.completed event type")
	}

	// ProbeFailedEvent
	probeFailed := NewProbeFailedEvent("ssh-banner", "192.168.1.1", 22, "connection refused")
	if probeFailed.Type() != "probe.failed" {
		t.Error("expected probe.failed event type")
	}
	if probeFailed.Error != "connection refused" {
		t.Error("expected error 'connection refused'")
	}
}

// TestGUIEvents — ветка: события GUI
func TestGUIEvents(t *testing.T) {
	// ProgressUpdatedEvent
	progress := NewProgressUpdatedEvent("scan", 5, 100, "Scanning hosts...")
	if progress.Type() != "progress.updated" {
		t.Error("expected progress.updated event type")
	}
	if progress.Stage != "scan" {
		t.Error("expected stage scan")
	}
	if progress.Current != 5 {
		t.Error("expected current 5")
	}

	// ThemeChangedEvent
	theme := NewThemeChangedEvent("dark")
	if theme.Type() != "theme.changed" {
		t.Error("expected theme.changed event type")
	}
	if theme.Mode != "dark" {
		t.Error("expected mode dark")
	}

	// SettingsChangedEvent
	settings := NewSettingsChangedEvent("timeout", "2s")
	if settings.Type() != "settings.changed" {
		t.Error("expected settings.changed event type")
	}
	if settings.Key != "timeout" {
		t.Error("expected key timeout")
	}
}

// TestEventBus_Concurrent — ветка: потокобезопасность
//
// Подписчики регистрируются ДО публикаций, а ожидание считается по факту
// вызова обработчиков: иначе число доставок недетерминировано (публикации
// стартуют раньше подписок, а callHandler выполняется в отдельной горутине).
func TestEventBus_Concurrent(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	const (
		publishers  = 10
		subscribers = 10
	)
	expected := publishers * subscribers

	var wg sync.WaitGroup
	callCount := 0
	var countMu sync.Mutex

	handled := &sync.WaitGroup{}
	handled.Add(expected)

	for i := 0; i < subscribers; i++ {
		bus.Subscribe("concurrent.event", func(event Event) {
			countMu.Lock()
			callCount++
			countMu.Unlock()
			handled.Done()
		})
	}

	// Конкурентная подписка на другой топик нагружает те же мьютексы,
	// но не влияет на строгий подсчёт выше.
	extraSubscriptions := &sync.WaitGroup{}
	extraSubscriptions.Add(1)
	go func() {
		defer extraSubscriptions.Done()
		for i := 0; i < subscribers; i++ {
			bus.Subscribe("concurrent.other", func(event Event) {})
		}
	}()

	for i := 0; i < publishers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(NewBaseEvent("concurrent.event"))
		}()
	}

	wg.Wait()
	extraSubscriptions.Wait()

	allHandled := make(chan struct{})
	go func() {
		handled.Wait()
		close(allHandled)
	}()

	select {
	case <-allHandled:
	case <-time.After(5 * time.Second):
		countMu.Lock()
		got := callCount
		countMu.Unlock()
		t.Fatalf("handlers did not finish: got %d calls, want %d", got, expected)
	}

	countMu.Lock()
	got := callCount
	countMu.Unlock()

	if got != expected {
		t.Errorf("expected exactly %d calls, got %d", expected, got)
	}
}

// TestEventBus_EmptyPublish — ветка: публикация без подписчиков
func TestEventBus_EmptyPublish(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	// Публикация без подписчиков не должна вызывать ошибок
	bus.Publish(NewBaseEvent("no.subscribers"))
}

// TestEventBus_EventsImplementsEvent — ветка: все события реализуют Event
func TestEventBus_EventsImplementsEvent(t *testing.T) {
	var _ Event = (*BaseEvent)(nil)
	var _ Event = (*ScanStartedEvent)(nil)
	var _ Event = (*ScanCompletedEvent)(nil)
	var _ Event = (*HostFoundEvent)(nil)
	var _ Event = (*PortOpenedEvent)(nil)
	var _ Event = (*DeviceDetectedEvent)(nil)
	var _ Event = (*ProbeStartedEvent)(nil)
	var _ Event = (*ProbeCompletedEvent)(nil)
	var _ Event = (*ProbeFailedEvent)(nil)
	var _ Event = (*ProgressUpdatedEvent)(nil)
	var _ Event = (*ThemeChangedEvent)(nil)
	var _ Event = (*SettingsChangedEvent)(nil)
}
