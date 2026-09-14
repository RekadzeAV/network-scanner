// Package commands_test — e2e-проверки сборки компонентов C1..C6:
// реестр команд + event bus + валидация конфигурации + единые ошибки.
//
// Тесты имитируют три интерфейса (CLI, GUI, API), вызывающие одни и те же
// команды, поэтому регрессия в любом слое видна сразу.
package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"network-scanner/internal/apperror"
	"network-scanner/internal/commands"
	"network-scanner/internal/configvalidation"
	"network-scanner/internal/eventbus"
)

// scannerApp — минимальное приложение из реальных компонентов.
type scannerApp struct {
	registry *commands.Registry
	bus      *eventbus.EventBus
	audit    *commands.MemoryAudit

	mu     sync.Mutex
	hosts  []string
	events []string
}

func newScannerApp(t *testing.T) *scannerApp {
	t.Helper()

	app := &scannerApp{
		bus:   eventbus.NewEventBus(),
		audit: commands.NewMemoryAudit(0),
	}

	// Порядок middleware важен: аудит снаружи, иначе паники не попадут в журнал.
	app.registry = commands.NewRegistry(
		commands.WithAudit(app.audit),
		commands.WithRecovery(),
		commands.WithTimeout(200*time.Millisecond),
	)

	app.bus.SubscribeAny(func(e eventbus.Event) {
		app.mu.Lock()
		defer app.mu.Unlock()
		app.events = append(app.events, e.Type())
	})

	err := app.registry.Register(
		commands.Command{Name: "scan", Aliases: []string{"s"}, Description: "сканирование сети", Risk: commands.RiskWrite, Handler: app.handleScan},
		commands.Command{Name: "report", Description: "отчёт по хостам", Risk: commands.RiskRead, Handler: app.handleReport},
		commands.Command{Name: "purge", Description: "удаление результатов", Risk: commands.RiskDestructive, Handler: app.handlePurge},
		commands.Command{Name: "crash", Description: "тестовый паникующий обработчик", Risk: commands.RiskRead, Handler: app.handleCrash},
		commands.Command{Name: "hang", Description: "тестовый зависающий обработчик", Risk: commands.RiskRead, Handler: app.handleHang},
	)
	if err != nil {
		t.Fatalf("register commands: %v", err)
	}

	t.Cleanup(app.bus.Close)
	return app
}

func (a *scannerApp) handleScan(ctx context.Context, req commands.Request) (commands.Response, error) {
	target, err := req.RequireArg("target")
	if err != nil {
		invalid := apperror.WrapCode(apperror.CodeInvalidInput, err, "укажите сеть для сканирования")
		return commands.NewFail(apperror.ToClient(invalid)), nil
	}
	if verr := configvalidation.ValidateCIDR(target); verr != nil {
		return commands.NewFail("неверный CIDR: " + target), nil
	}

	a.bus.Publish(eventbus.NewScanStartedEvent(target, "10s"))

	a.mu.Lock()
	a.hosts = append(a.hosts, target)
	count := len(a.hosts)
	a.mu.Unlock()

	a.bus.Publish(eventbus.NewScanCompletedEvent(count, 0, "10ms"))

	return commands.NewOK("сканирование завершено", map[string]any{"target": target, "hosts": count}), nil
}

func (a *scannerApp) handleReport(ctx context.Context, req commands.Request) (commands.Response, error) {
	a.mu.Lock()
	hosts := append([]string{}, a.hosts...)
	a.mu.Unlock()

	return commands.NewOK("хостов: "+strconv.Itoa(len(hosts)), hosts).WithMeta("format", "list"), nil
}

func (a *scannerApp) handlePurge(ctx context.Context, req commands.Request) (commands.Response, error) {
	a.mu.Lock()
	removed := len(a.hosts)
	a.hosts = nil
	a.mu.Unlock()

	return commands.NewOK("удалено записей", removed), nil
}

func (a *scannerApp) handleCrash(ctx context.Context, req commands.Request) (commands.Response, error) {
	panic("необработанный сбой обработчика")
}

func (a *scannerApp) handleHang(ctx context.Context, req commands.Request) (commands.Response, error) {
	<-ctx.Done()
	return commands.NewFail("операция прервана"), apperror.Wrap(ctx.Err(), "ожидание превысило лимит")
}

func (a *scannerApp) hostCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.hosts)
}

func (a *scannerApp) eventTypes() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string{}, a.events...)
}

func waitFor(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("условие не выполнилось за", timeout)
}

func TestE2E_CLIFlow(t *testing.T) {
	app := newScannerApp(t)

	var out bytes.Buffer
	var errOut bytes.Buffer
	presenter := commands.NewTextPresenter(&out, &errOut)

	scanReq := commands.NewRequest("s", commands.SourceCLI)
	scanReq.Args["target"] = "192.168.1.0/24"

	resp, err := app.registry.Dispatch(context.Background(), scanReq)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if err := presenter.Present(scanReq, resp); err != nil {
		t.Fatalf("present: %v", err)
	}

	reportReq := commands.NewRequest("report", commands.SourceCLI)
	resp, err = app.registry.Dispatch(context.Background(), reportReq)
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if err := presenter.Present(reportReq, resp); err != nil {
		t.Fatalf("present report: %v", err)
	}

	output := out.String()
	for _, want := range []string{"сканирование завершено", "192.168.1.0/24", "хостов: 1", "format: list"} {
		if !strings.Contains(output, want) {
			t.Errorf("вывод CLI не содержит %q:\n%s", want, output)
		}
	}
	if errOut.Len() != 0 {
		t.Errorf("неожиданные ошибки в выводе: %s", errOut.String())
	}

	entries := app.audit.Entries()
	if len(entries) != 2 {
		t.Fatalf("записей аудита = %d, want 2", len(entries))
	}
	if entries[0].Command != "s" || entries[1].Command != "report" {
		t.Errorf("аудит хранит вызванные имена: %v, %v", entries[0].Command, entries[1].Command)
	}
	for _, entry := range entries {
		if entry.Source != commands.SourceCLI || !entry.OK {
			t.Errorf("запись аудита = %+v", entry)
		}
	}
}

func TestE2E_APIFlow(t *testing.T) {
	app := newScannerApp(t)

	var out bytes.Buffer
	presenter := commands.NewJSONPresenter(&out)

	scanReq := commands.NewRequest("scan", commands.SourceAPI)
	scanReq.Args["target"] = "10.0.0.0/24"

	resp, err := app.registry.Dispatch(context.Background(), scanReq)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if err := presenter.Present(scanReq, resp); err != nil {
		t.Fatalf("present: %v", err)
	}

	var okEnvelope struct {
		Command string         `json:"command"`
		OK      bool           `json:"ok"`
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
	if err := json.Unmarshal(lastJSONLine(t, out.Bytes()), &okEnvelope); err != nil {
		t.Fatalf("decode ok envelope: %v", err)
	}
	if !okEnvelope.OK || okEnvelope.Command != "scan" || okEnvelope.Message != "сканирование завершено" {
		t.Errorf("ok envelope = %+v", okEnvelope)
	}
	if okEnvelope.Data["target"] != "10.0.0.0/24" {
		t.Errorf("data = %v", okEnvelope.Data)
	}

	// Неизвестная команда: ошибка уровня транспорта, код command_not_found.
	out.Reset()
	_, err = app.registry.Dispatch(context.Background(), commands.NewRequest("ghost", commands.SourceAPI))
	if err == nil {
		t.Fatal("unknown command should fail")
	}
	if err := presenter.PresentError(commands.NewRequest("ghost", commands.SourceAPI), err); err != nil {
		t.Fatalf("present error: %v", err)
	}

	var errEnvelope struct {
		OK        bool   `json:"ok"`
		ErrorCode string `json:"errorCode"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal(lastJSONLine(t, out.Bytes()), &errEnvelope); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	if errEnvelope.OK || errEnvelope.ErrorCode != "command_not_found" {
		t.Errorf("error envelope = %+v", errEnvelope)
	}
}

func TestE2E_GUIFlowPublishesEvents(t *testing.T) {
	app := newScannerApp(t)

	// GUI не печатает текст, а обновляет виджеты по событиям шины.
	scanReq := commands.NewRequest("scan", commands.SourceGUI)
	scanReq.Args["target"] = "172.16.0.0/24"

	resp, err := app.registry.Dispatch(context.Background(), scanReq)
	if err != nil || !resp.OK {
		t.Fatalf("scan = %+v, err = %v", resp, err)
	}

	// Обработчики шины асинхронны, поэтому проверяем набор событий, а не порядок.
	waitFor(t, time.Second, func() bool {
		types := app.eventTypes()
		seen := map[string]bool{}
		for _, eventType := range types {
			seen[eventType] = true
		}
		return seen["scan.started"] && seen["scan.completed"]
	})

	if app.hostCount() != 1 {
		t.Errorf("hosts = %d, want 1", app.hostCount())
	}
}

func TestE2E_DestructiveCommandNeedsConfirmation(t *testing.T) {
	app := newScannerApp(t)

	scanReq := commands.NewRequest("scan", commands.SourceGUI)
	scanReq.Args["target"] = "192.168.0.0/24"
	if _, err := app.registry.Dispatch(context.Background(), scanReq); err != nil {
		t.Fatalf("scan: %v", err)
	}

	// API не может удалить данные без подтверждения пользователя.
	if _, err := app.registry.Dispatch(context.Background(), commands.NewRequest("purge", commands.SourceAPI)); !strings.Contains(err.Error(), "confirmation required") {
		t.Fatalf("unconfirmed purge err = %v", err)
	}
	if app.hostCount() != 1 {
		t.Fatal("данные не должны удалиться без подтверждения")
	}

	confirmed := commands.NewRequest("purge", commands.SourceGUI)
	confirmed.Confirmed = true
	resp, err := app.registry.Dispatch(context.Background(), confirmed)
	if err != nil || !resp.OK {
		t.Fatalf("confirmed purge = %+v, %v", resp, err)
	}
	if app.hostCount() != 0 {
		t.Errorf("hosts after purge = %d, want 0", app.hostCount())
	}

	// Отказ в деструктивной команде тоже попадает в журнал — это требование аудита.
	entries := app.audit.Entries()
	if len(entries) != 3 {
		t.Fatalf("записей аудита = %d, want 3: %+v", len(entries), entries)
	}
	if !entries[0].OK || entries[0].Command != "scan" {
		t.Errorf("первая запись = %+v", entries[0])
	}
	if entries[1].OK || entries[1].Command != "purge" || !strings.Contains(entries[1].Err, "confirmation required") {
		t.Errorf("отказ должен быть записан как ошибка: %+v", entries[1])
	}
	if !entries[2].OK || entries[2].Command != "purge" {
		t.Errorf("третья запись = %+v", entries[2])
	}
}

func TestE2E_InvalidInputReturnsFailedResponse(t *testing.T) {
	app := newScannerApp(t)

	// Аргумент не указан — бизнес-отказ, а не ошибка транспорта.
	resp, err := app.registry.Dispatch(context.Background(), commands.NewRequest("scan", commands.SourceCLI))
	if err != nil {
		t.Fatalf("dispatch should not fail transport-level: %v", err)
	}
	if resp.OK {
		t.Error("response should not be OK")
	}
	if resp.Message != "укажите сеть для сканирования" {
		t.Errorf("message = %q", resp.Message)
	}

	// Неверный CIDR отсекает валидатор конфигурации.
	badReq := commands.NewRequest("scan", commands.SourceCLI)
	badReq.Args["target"] = "not-a-cidr"
	resp, err = app.registry.Dispatch(context.Background(), badReq)
	if err != nil || resp.OK {
		t.Fatalf("bad cidr = %+v, %v", resp, err)
	}
	if !strings.Contains(resp.Message, "неверный CIDR") {
		t.Errorf("message = %q", resp.Message)
	}
}

func TestE2E_PanicIsRecoveredAndAudited(t *testing.T) {
	app := newScannerApp(t)

	_, err := app.registry.Dispatch(context.Background(), commands.NewRequest("crash", commands.SourceAPI))
	if err == nil {
		t.Fatal("panic should be converted to error")
	}
	if !strings.Contains(err.Error(), "необработанный сбой обработчика") {
		t.Errorf("error = %v", err)
	}

	last, ok := app.audit.Last()
	if !ok || last.OK || last.Err == "" {
		t.Fatalf("panic should be audited as failure: %+v (ok=%v)", last, ok)
	}

	// Приложение остаётся работоспособным после сбоя команды.
	scanReq := commands.NewRequest("scan", commands.SourceAPI)
	scanReq.Args["target"] = "192.168.5.0/24"
	if resp, err := app.registry.Dispatch(context.Background(), scanReq); err != nil || !resp.OK {
		t.Errorf("app should recover: %+v, %v", resp, err)
	}
}

func TestE2E_TimeoutMappedToErrorCode(t *testing.T) {
	app := newScannerApp(t)

	_, err := app.registry.Dispatch(context.Background(), commands.NewRequest("hang", commands.SourceAPI))
	if err == nil {
		t.Fatal("hang should time out")
	}
	if code := apperror.Of(err); code != apperror.CodeTimeout {
		t.Errorf("code = %q, want %q", code, apperror.CodeTimeout)
	}

	var out bytes.Buffer
	presenter := commands.NewJSONPresenter(&out)
	if err := presenter.PresentError(commands.NewRequest("hang", commands.SourceAPI), err); err != nil {
		t.Fatalf("present error: %v", err)
	}

	var envelope struct {
		ErrorCode string `json:"errorCode"`
	}
	if err := json.Unmarshal(lastJSONLine(t, out.Bytes()), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if envelope.ErrorCode != string(apperror.CodeTimeout) {
		t.Errorf("errorCode = %q, want %q", envelope.ErrorCode, apperror.CodeTimeout)
	}
}

func TestE2E_AllInterfacesShareCommands(t *testing.T) {
	app := newScannerApp(t)

	sources := []commands.Source{commands.SourceCLI, commands.SourceGUI, commands.SourceAPI}
	var wg sync.WaitGroup

	for _, source := range sources {
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(source commands.Source) {
				defer wg.Done()

				req := commands.NewRequest("scan", source)
				req.Args["target"] = "10.20.30.0/24"
				if _, err := app.registry.Dispatch(context.Background(), req); err != nil {
					t.Errorf("dispatch from %s: %v", source, err)
				}
			}(source)
		}
	}
	wg.Wait()

	if app.hostCount() != 15 {
		t.Errorf("hosts = %d, want 15", app.hostCount())
	}
	if app.audit.Len() != 15 {
		t.Errorf("аудит = %d, want 15", app.audit.Len())
	}

	seen := map[commands.Source]int{}
	for _, entry := range app.audit.Entries() {
		seen[entry.Source]++
	}
	for _, source := range sources {
		if seen[source] != 5 {
			t.Errorf("источник %s: %d вызовов, want 5", source, seen[source])
		}
	}
}

func TestE2E_CommandDiscoveryForAllInterfaces(t *testing.T) {
	app := newScannerApp(t)

	// CLI показывает справку, GUI строит меню, API отдаёт список возможностей.
	names := app.registry.Names()
	want := []string{"crash", "hang", "purge", "report", "scan"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("names = %v, want %v", names, want)
	}

	for _, list := range [][]commands.Command{app.registry.List()} {
		for _, cmd := range list {
			if cmd.Description == "" {
				t.Errorf("команда %q без описания", cmd.Name)
			}
			if cmd.Risk == "" {
				t.Errorf("команда %q без уровня риска", cmd.Name)
			}
		}
	}

	if _, ok := app.registry.Get("s"); !ok {
		t.Error("алиас CLI-команды должен находиться из любого интерфейса")
	}
}

// lastJSONLine — последний JSON-объект из потока (президентер пишет по строке).
func lastJSONLine(t *testing.T, data []byte) []byte {
	t.Helper()

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		t.Fatal("пустой вывод презентера")
	}

	index := bytes.LastIndexByte(trimmed, '\n')
	if index < 0 {
		return trimmed
	}
	return bytes.TrimSpace(trimmed[index+1:])
}
