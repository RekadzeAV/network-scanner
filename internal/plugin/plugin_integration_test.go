package plugin

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"network-scanner/internal/contracts"
)

// === Integration: Plugin Types ===

func TestIntegrationPluginTypes(t *testing.T) {
	if TypeFilter != "filter" {
		t.Errorf("expected TypeFilter='filter', got %q", TypeFilter)
	}
	if TypeExporter != "exporter" {
		t.Errorf("expected TypeExporter='exporter', got %q", TypeExporter)
	}
	if TypeScanner != "scanner" {
		t.Errorf("expected TypeScanner='scanner', got %q", TypeScanner)
	}
	if TypeReporter != "reporter" {
		t.Errorf("expected TypeReporter='reporter', got %q", TypeReporter)
	}
}

// === Integration: Plugin Info ===

func TestIntegrationPluginInfo(t *testing.T) {
	info := Info{
		Name:        "TestPlugin",
		Version:     "1.0.0",
		Description: "Test plugin for integration testing",
		Author:      "Tester",
		Type:        TypeFilter,
	}

	if info.Name != "TestPlugin" {
		t.Errorf("expected Name 'TestPlugin', got %q", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %q", info.Version)
	}
	if info.Description != "Test plugin for integration testing" {
		t.Errorf("expected Description to match, got %q", info.Description)
	}
	if info.Author != "Tester" {
		t.Errorf("expected Author 'Tester', got %q", info.Author)
	}
	if info.Type != TypeFilter {
		t.Errorf("expected Type TypeFilter, got %q", info.Type)
	}
}

// === Integration: DefaultEventBus ===

func TestIntegrationEventBus_SubscribeAndPublish(t *testing.T) {
	bus := NewDefaultEventBus()
	if bus == nil {
		t.Fatal("expected non-nil event bus")
	}

	var receivedData interface{}
	handler := func(data interface{}) {
		receivedData = data
	}

	id := bus.Subscribe("test_event", handler)
	if id <= 0 {
		t.Errorf("expected positive handler ID, got %d", id)
	}

	bus.Publish("test_event", "test_data")
	if receivedData != "test_data" {
		t.Errorf("expected received data 'test_data', got %v", receivedData)
	}
}

func TestIntegrationEventBus_MultipleHandlers(t *testing.T) {
	bus := NewDefaultEventBus()

	var count1, count2 int
	handler1 := func(data interface{}) {
		count1++
	}
	handler2 := func(data interface{}) {
		count2++
	}

	bus.Subscribe("event", handler1)
	bus.Subscribe("event", handler2)

	bus.Publish("event", "data")

	if count1 != 1 {
		t.Errorf("expected handler1 called once, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("expected handler2 called once, got %d", count2)
	}
}

func TestIntegrationEventBus_Unsubscribe(t *testing.T) {
	bus := NewDefaultEventBus()

	var called bool
	handler := func(data interface{}) {
		called = true
	}

	id := bus.Subscribe("event", handler)
	bus.Unsubscribe("event", id)

	bus.Publish("event", "data")

	if called {
		t.Error("expected handler not to be called after unsubscribe")
	}
}

func TestIntegrationEventBus_UnsubscribeNonExistent(t *testing.T) {
	bus := NewDefaultEventBus()

	// Unsubscribing non-existent handler should not panic
	bus.Unsubscribe("non_existent_event", 999)
}

func TestIntegrationEventBus_PublishNonExistentEvent(t *testing.T) {
	bus := NewDefaultEventBus()

	// Publishing non-existent event should not panic
	bus.Publish("non_existent_event", "data")
}

func TestIntegrationEventBus_MultipleEvents(t *testing.T) {
	bus := NewDefaultEventBus()

	var event1Count, event2Count int
	handler1 := func(data interface{}) {
		event1Count++
	}
	handler2 := func(data interface{}) {
		event2Count++
	}

	bus.Subscribe("event1", handler1)
	bus.Subscribe("event2", handler2)

	bus.Publish("event1", "data1")
	bus.Publish("event2", "data2")
	bus.Publish("event1", "data3")

	if event1Count != 2 {
		t.Errorf("expected event1 called twice, got %d", event1Count)
	}
	if event2Count != 1 {
		t.Errorf("expected event2 called once, got %d", event2Count)
	}
}

// === Integration: PluginRegistry ===

func TestIntegrationPluginRegistry_Create(t *testing.T) {
	registry := NewPluginRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestIntegrationPluginRegistry_GetEmpty(t *testing.T) {
	registry := NewPluginRegistry()

	plugin, exists := registry.Get("non_existent")
	if plugin != nil {
		t.Error("expected nil plugin for non-existent name")
	}
	if exists {
		t.Error("expected exists=false for non-existent name")
	}
}

func TestIntegrationPluginRegistry_GetAllEmpty(t *testing.T) {
	registry := NewPluginRegistry()

	allPlugins := registry.GetAll()
	if allPlugins == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(allPlugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(allPlugins))
	}
}

func TestIntegrationPluginRegistry_EventBus(t *testing.T) {
	registry := NewPluginRegistry()

	eventBus := registry.EventBus()
	if eventBus == nil {
		t.Fatal("expected non-nil event bus")
	}
}

// === Integration: PluginContext ===

func TestIntegrationPluginContext_Create(t *testing.T) {
	ctx := NewPluginContext("scan-001")
	if ctx == nil {
		t.Fatal("expected non-nil plugin context")
	}
	if ctx.ScanID != "scan-001" {
		t.Errorf("expected ScanID 'scan-001', got %q", ctx.ScanID)
	}
	if ctx.StartTime.IsZero() {
		t.Error("expected non-zero StartTime")
	}
	if ctx.CancelContext == nil {
		t.Error("expected non-nil CancelContext")
	}
	if ctx.CancelFunc == nil {
		t.Error("expected non-nil CancelFunc")
	}
}

func TestIntegrationPluginContext_Cancel(t *testing.T) {
	ctx := NewPluginContext("scan-002")

	ctx.CancelFunc()

	select {
	case <-ctx.CancelContext.Done():
		// Expected
	default:
		t.Error("expected context to be cancelled")
	}
}

func TestIntegrationPluginContext_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = ctx.Done()
}

// === Integration: PluginLoader ===

func TestIntegrationPluginLoader_Create(t *testing.T) {
	loader := NewPluginLoader("")
	if loader == nil {
		t.Fatal("expected non-nil loader")
	}
}

func TestIntegrationPluginLoader_CreateWithDir(t *testing.T) {
	loader := NewPluginLoader("/tmp/plugins")
	if loader == nil {
		t.Fatal("expected non-nil loader")
	}
}

func TestIntegrationPluginLoader_Load_InvalidExtension(t *testing.T) {
	loader := NewPluginLoader("")

	_, err := loader.Load("plugin.txt")
	if err == nil {
		t.Error("expected error for invalid extension")
	}
}

func TestIntegrationPluginLoader_Load_FileNotFound(t *testing.T) {
	loader := NewPluginLoader("")

	_, err := loader.Load("non_existent_plugin.dll")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestIntegrationPluginLoader_LoadAll_NonExistentDir(t *testing.T) {
	loader := NewPluginLoader("")

	plugins, err := loader.LoadAll("/non/existent/dir")
	if err != nil {
		t.Fatalf("expected no error for non-existent dir, got %v", err)
	}
	// LoadAll returns nil slice for non-existent directory (existing behavior)
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestIntegrationPluginLoader_LoadAll_NotADirectory(t *testing.T) {
	loader := NewPluginLoader("")

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = loader.LoadAll(tmpFile.Name())
	if err == nil {
		t.Error("expected error for non-directory path")
	}
}

func TestIntegrationPluginLoader_Detect(t *testing.T) {
	loader := NewPluginLoader("")
	_ = loader
	// Discover calls LoadAll internally
}

// === Integration: ValidExtensions ===

func TestIntegrationValidExtensions(t *testing.T) {
	extensions := ValidExtensions()
	if extensions == nil {
		t.Fatal("expected non-nil extensions slice")
	}
	if len(extensions) == 0 {
		t.Error("expected at least one extension")
	}

	// On Windows, should be .dll
	// On Linux, should be .so
	// On macOS, should be .dylib
	switch runtime.GOOS {
	case "windows":
		if extensions[0] != ".dll" {
			t.Errorf("expected .dll on Windows, got %q", extensions[0])
		}
	case "linux":
		if extensions[0] != ".so" {
			t.Errorf("expected .so on Linux, got %q", extensions[0])
		}
	case "darwin":
		if extensions[0] != ".dylib" {
			t.Errorf("expected .dylib on macOS, got %q", extensions[0])
		}
	}
}

// === Integration: Plugin Interface ===

func TestIntegrationPluginInterface_Compliance(t *testing.T) {
	// Verify that the Plugin interface is properly defined
	// This test ensures the interface has the required methods
	var _ Plugin = (*mockPlugin)(nil)
}

// mockPlugin implements Plugin interface for testing
type mockPlugin struct {
	info Info
}

func (m *mockPlugin) Info() Info {
	return m.info
}

func (m *mockPlugin) Init(cfg map[string]interface{}) error {
	return nil
}

func (m *mockPlugin) Run(ctx context.Context, results []contracts.ScanResult) (interface{}, error) {
	return results, nil
}

func (m *mockPlugin) Close() error {
	return nil
}

// === Integration: FilterPlugin Interface ===

func TestIntegrationFilterPluginInterface(t *testing.T) {
	var _ FilterPlugin = (*mockFilterPlugin)(nil)
}

type mockFilterPlugin struct {
	mockPlugin
}

func (m *mockFilterPlugin) Filter(results []contracts.ScanResult) ([]contracts.ScanResult, error) {
	return results, nil
}

// === Integration: ExporterPlugin Interface ===

func TestIntegrationExporterPluginInterface(t *testing.T) {
	var _ ExporterPlugin = (*mockExporterPlugin)(nil)
}

type mockExporterPlugin struct {
	mockPlugin
}

func (m *mockExporterPlugin) Export(results []contracts.ScanResult, path string) error {
	return nil
}

func (m *mockExporterPlugin) Format() string {
	return "json"
}

// === Integration: ScannerPlugin Interface ===

func TestIntegrationScannerPluginInterface(t *testing.T) {
	var _ ScannerPlugin = (*mockScannerPlugin)(nil)
}

type mockScannerPlugin struct {
	mockPlugin
}

func (m *mockScannerPlugin) Scan(ctx context.Context, target string) ([]contracts.ScanResult, error) {
	return nil, nil
}

// === Integration: ReporterPlugin Interface ===

func TestIntegrationReporterPluginInterface(t *testing.T) {
	var _ ReporterPlugin = (*mockReporterPlugin)(nil)
}

type mockReporterPlugin struct {
	mockPlugin
}

func (m *mockReporterPlugin) Generate(results []contracts.ScanResult, format string) ([]byte, error) {
	return []byte{}, nil
}

// === Integration: PluginRegistry with Mock Plugin ===

func TestIntegrationPluginRegistry_RegisterAndGet(t *testing.T) {
	registry := NewPluginRegistry()
	mockP := &mockPlugin{
		info: Info{
			Name:        "MockPlugin",
			Version:     "1.0.0",
			Description: "Mock plugin",
			Author:      "Tester",
			Type:        TypeFilter,
		},
	}

	err := registry.Register(mockP)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	plugin, exists := registry.Get("MockPlugin")
	if !exists {
		t.Error("expected plugin to exist")
	}
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}
}

func TestIntegrationPluginRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewPluginRegistry()
	mockP := &mockPlugin{
		info: Info{Name: "DuplicatePlugin"},
	}

	err := registry.Register(mockP)
	if err != nil {
		t.Fatalf("expected no error for first register, got %v", err)
	}

	// Registering duplicate should not error
	err = registry.Register(mockP)
	if err != nil {
		t.Errorf("expected no error for duplicate register, got %v", err)
	}
}

func TestIntegrationPluginRegistry_GetAll(t *testing.T) {
	registry := NewPluginRegistry()

	mockP1 := &mockPlugin{info: Info{Name: "Plugin1"}}
	mockP2 := &mockPlugin{info: Info{Name: "Plugin2"}}

	registry.Register(mockP1)
	registry.Register(mockP2)

	allPlugins := registry.GetAll()
	if len(allPlugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(allPlugins))
	}
}

func TestIntegrationPluginRegistry_CloseAll(t *testing.T) {
	registry := NewPluginRegistry()

	mockP := &mockPlugin{info: Info{Name: "TestPlugin"}}
	registry.Register(mockP)

	err := registry.CloseAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// === Integration: PluginContext Lifecycle ===

func TestIntegrationPluginContext_Lifecycle(t *testing.T) {
	ctx := NewPluginContext("scan-003")

	// Verify initial state
	if ctx.ScanID != "scan-003" {
		t.Errorf("expected ScanID 'scan-003', got %q", ctx.ScanID)
	}
	if ctx.StartTime.IsZero() {
		t.Error("expected non-zero StartTime")
	}
	if ctx.CancelContext == nil {
		t.Error("expected non-nil CancelContext")
	}
	if ctx.CancelFunc == nil {
		t.Error("expected non-nil CancelFunc")
	}

	// Test cancellation
	ctx.CancelFunc()

	select {
	case <-ctx.CancelContext.Done():
		// Expected
	default:
		t.Error("expected context to be cancelled")
	}
}

// === Integration: EventBus Edge Cases ===

func TestIntegrationEventBus_SubscribeMultipleTimes(t *testing.T) {
	bus := NewDefaultEventBus()

	var count int
	handler := func(data interface{}) {
		count++
	}

	// Subscribe same handler multiple times
	id1 := bus.Subscribe("event", handler)
	id2 := bus.Subscribe("event", handler)

	if id1 == id2 {
		t.Error("expected different handler IDs")
	}

	bus.Publish("event", "data")

	// Both handlers should be called
	if count != 2 {
		t.Errorf("expected handler called twice, got %d", count)
	}
}

func TestIntegrationEventBus_PublishWithNilData(t *testing.T) {
	bus := NewDefaultEventBus()

	var received interface{}
	handler := func(data interface{}) {
		received = data
	}

	bus.Subscribe("event", handler)
	bus.Publish("event", nil)

	if received != nil {
		t.Errorf("expected nil data, got %v", received)
	}
}

// === Integration: PluginLoader with Real File ===

func TestIntegrationPluginLoader_LoadAll_EmptyDir(t *testing.T) {
	// Create a temporary empty directory
	tmpDir, err := os.MkdirTemp("", "plugins_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	loader := NewPluginLoader("")

	plugins, err := loader.LoadAll(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// LoadAll returns nil slice for empty directory (existing behavior)
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins in empty dir, got %d", len(plugins))
	}
}

// === Integration: PluginRegistry with Multiple Plugins ===

func TestIntegrationPluginRegistry_MultiplePlugins(t *testing.T) {
	registry := NewPluginRegistry()

	p1 := &mockPlugin{info: Info{Name: "PluginAlpha", Type: TypeFilter}}
	p2 := &mockPlugin{info: Info{Name: "PluginBeta", Type: TypeExporter}}
	p3 := &mockPlugin{info: Info{Name: "PluginGamma", Type: TypeScanner}}

	registry.Register(p1)
	registry.Register(p2)
	registry.Register(p3)

	all := registry.GetAll()
	if len(all) != 3 {
		t.Errorf("expected 3 plugins, got %d", len(all))
	}
}

func TestIntegrationPluginRegistry_DuplicateNoError(t *testing.T) {
	registry := NewPluginRegistry()
	p := &mockPlugin{info: Info{Name: "UniquePlugin"}}

	err := registry.Register(p)
	if err != nil {
		t.Fatalf("expected no error on first register, got %v", err)
	}

	// Register same plugin again — should not error, but not add duplicate
	err = registry.Register(p)
	if err != nil {
		t.Errorf("expected no error on duplicate register, got %v", err)
	}

	all := registry.GetAll()
	if len(all) != 1 {
		t.Errorf("expected 1 plugin (no duplicate), got %d", len(all))
	}
}

func TestIntegrationPluginRegistry_GetNonExistent(t *testing.T) {
	registry := NewPluginRegistry()

	p, exists := registry.Get("DoesNotExist")
	if p != nil {
		t.Error("expected nil plugin for non-existent name")
	}
	if exists {
		t.Error("expected exists=false")
	}
}

func TestIntegrationPluginRegistry_CloseAllMultiple(t *testing.T) {
	registry := NewPluginRegistry()

	p1 := &mockPlugin{info: Info{Name: "Plugin1"}}
	p2 := &mockPlugin{info: Info{Name: "Plugin2"}}

	registry.Register(p1)
	registry.Register(p2)

	err := registry.CloseAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// === Integration: PluginContext Advanced ===

func TestIntegrationPluginContext_ScanID(t *testing.T) {
	ctx := NewPluginContext("scan-abc-123")
	if ctx.ScanID != "scan-abc-123" {
		t.Errorf("expected 'scan-abc-123', got %q", ctx.ScanID)
	}
}

func TestIntegrationPluginContext_StartTimeNotZero(t *testing.T) {
	ctx := NewPluginContext("test-scan")
	if ctx.StartTime.IsZero() {
		t.Error("expected non-zero StartTime")
	}
	if ctx.StartTime.After(time.Now()) {
		t.Error("StartTime should not be in the future")
	}
}

func TestIntegrationPluginContext_CancelTwice(t *testing.T) {
	ctx := NewPluginContext("test-cancel")
	ctx.CancelFunc()
	ctx.CancelFunc() // Second cancel should not panic

	select {
	case <-ctx.CancelContext.Done():
		// Expected
	default:
		t.Error("expected context to be cancelled")
	}
}

func TestIntegrationPluginContext_ContextNotCancelledInitially(t *testing.T) {
	ctx := NewPluginContext("test-not-cancelled")

	select {
	case <-ctx.CancelContext.Done():
		t.Error("expected context NOT to be cancelled initially")
	default:
		// Expected — not cancelled yet
	}
}

// === Integration: EventBus Advanced ===

func TestIntegrationEventBus_SubscribeDifferentEvents(t *testing.T) {
	bus := NewDefaultEventBus()

	var count1, count2, count3 int
	handler1 := func(data interface{}) { count1++ }
	handler2 := func(data interface{}) { count2++ }
	handler3 := func(data interface{}) { count3++ }

	bus.Subscribe("event_a", handler1)
	bus.Subscribe("event_b", handler2)
	bus.Subscribe("event_c", handler3)

	bus.Publish("event_a", nil)
	bus.Publish("event_b", nil)
	bus.Publish("event_c", nil)

	if count1 != 1 {
		t.Errorf("expected count1=1, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("expected count2=1, got %d", count2)
	}
	if count3 != 1 {
		t.Errorf("expected count3=1, got %d", count3)
	}
}

func TestIntegrationEventBus_UnsubscribeSpecificHandler(t *testing.T) {
	bus := NewDefaultEventBus()

	var count1, count2 int
	handler1 := func(data interface{}) { count1++ }
	handler2 := func(data interface{}) { count2++ }

	id1 := bus.Subscribe("event", handler1)
	bus.Subscribe("event", handler2)

	bus.Unsubscribe("event", id1)
	bus.Publish("event", nil)

	if count1 != 0 {
		t.Errorf("expected count1=0 after unsubscribe, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("expected count2=1, got %d", count2)
	}
}

func TestIntegrationEventBus_PublishComplexData(t *testing.T) {
	bus := NewDefaultEventBus()

	type eventData struct {
		Key   string
		Value int
	}

	var received eventData
	handler := func(data interface{}) {
		if d, ok := data.(eventData); ok {
			received = d
		}
	}

	bus.Subscribe("complex_event", handler)
	bus.Publish("complex_event", eventData{Key: "test", Value: 42})

	if received.Key != "test" || received.Value != 42 {
		t.Errorf("expected {test, 42}, got {%q, %d}", received.Key, received.Value)
	}
}

// === Integration: Plugin Types Edge Cases ===

func TestIntegrationPluginType_EmptyString(t *testing.T) {
	var tp Type = ""
	if tp != "" {
		t.Errorf("expected empty Type, got %q", tp)
	}
}

func TestIntegrationPluginType_AllTypes(t *testing.T) {
	types := []Type{TypeFilter, TypeExporter, TypeScanner, TypeReporter}
	if len(types) != 4 {
		t.Errorf("expected 4 types, got %d", len(types))
	}
}

// === Integration: Info Structure ===

func TestIntegrationInfo_AllFields(t *testing.T) {
	info := Info{
		Name:        "FullInfoPlugin",
		Version:     "2.1.0",
		Description: "Plugin with all fields set",
		Author:      "FullAuthor",
		Type:        TypeReporter,
	}

	if info.Name != "FullInfoPlugin" {
		t.Errorf("expected 'FullInfoPlugin', got %q", info.Name)
	}
	if info.Version != "2.1.0" {
		t.Errorf("expected '2.1.0', got %q", info.Version)
	}
	if info.Description != "Plugin with all fields set" {
		t.Errorf("expected 'Plugin with all fields set', got %q", info.Description)
	}
	if info.Author != "FullAuthor" {
		t.Errorf("expected 'FullAuthor', got %q", info.Author)
	}
	if info.Type != TypeReporter {
		t.Errorf("expected TypeReporter, got %q", info.Type)
	}
}

func TestIntegrationInfo_EmptyFields(t *testing.T) {
	info := Info{}

	if info.Name != "" {
		t.Errorf("expected empty Name, got %q", info.Name)
	}
	if info.Version != "" {
		t.Errorf("expected empty Version, got %q", info.Version)
	}
	if info.Description != "" {
		t.Errorf("expected empty Description, got %q", info.Description)
	}
	if info.Author != "" {
		t.Errorf("expected empty Author, got %q", info.Author)
	}
}
