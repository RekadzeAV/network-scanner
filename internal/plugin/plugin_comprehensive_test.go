package plugin

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"network-scanner/internal/contracts"
)

// simplePlugin — простой мок для тестирования.
type simplePlugin struct {
	info Info
}

func (m *simplePlugin) Info() Info {
	return m.info
}

func (m *simplePlugin) Init(cfg map[string]interface{}) error {
	return nil
}

func (m *simplePlugin) Run(_ context.Context, _ []contracts.ScanResult) (interface{}, error) {
	return nil, nil
}

func (m *simplePlugin) Close() error {
	return nil
}

// failingPlugin — плагин с ошибкой при закрытии.
type failingPlugin struct {
	info     Info
	closeErr error
}

func (m *failingPlugin) Info() Info {
	return m.info
}

func (m *failingPlugin) Init(cfg map[string]interface{}) error {
	return nil
}

func (m *failingPlugin) Run(_ context.Context, _ []contracts.ScanResult) (interface{}, error) {
	return nil, nil
}

func (m *failingPlugin) Close() error {
	if m.closeErr != nil {
		return m.closeErr
	}
	return nil
}

// --- Plugin Registry Tests ---

// TestPluginRegistry_RegisterDuplicate проверяет что дубликаты игнорируются.
func TestPluginRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewPluginRegistry()

	plugin1 := &simplePlugin{info: Info{Name: "DupPlugin"}}
	plugin2 := &simplePlugin{info: Info{Name: "DupPlugin"}}

	// Первая регистрация успешна
	err := registry.Register(plugin1)
	if err != nil {
		t.Fatalf("First register should succeed, got error: %v", err)
	}

	// Вторая регистрация с тем же именем должна игнорироваться (не заменять)
	err = registry.Register(plugin2)
	if err != nil {
		t.Fatalf("Duplicate register should not error, got: %v", err)
	}

	// Получаем плагин и проверяем что это оригинальный (plugin1)
	p, ok := registry.Get("DupPlugin")
	if !ok {
		t.Fatal("Plugin should exist")
	}
	if p != plugin1 {
		t.Error("Should return original plugin, not the duplicate")
	}

	// Всего должен быть 1 плагин
	all := registry.GetAll()
	if len(all) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(all))
	}
}

// TestPluginRegistry_CloseAll проверяет корректное закрытие всех плагинов.
func TestPluginRegistry_CloseAll(t *testing.T) {
	registry := NewPluginRegistry()

	// Плагины без ошибок
	p1 := &simplePlugin{info: Info{Name: "Plugin1"}}
	p2 := &simplePlugin{info: Info{Name: "Plugin2"}}

	registry.Register(p1)
	registry.Register(p2)

	err := registry.CloseAll()
	if err != nil {
		t.Fatalf("CloseAll should succeed, got: %v", err)
	}
}

// TestPluginRegistry_CloseAllWithError проверяет обработку ошибок при закрытии.
func TestPluginRegistry_CloseAllWithError(t *testing.T) {
	registry := NewPluginRegistry()

	failingPlugin := &failingPlugin{info: Info{Name: "FailingPlugin"}, closeErr: os.ErrPermission}
	registry.Register(failingPlugin)

	err := registry.CloseAll()
	if err == nil {
		t.Error("CloseAll should return error when plugin.Close() fails")
	}
	if err != os.ErrPermission {
		t.Errorf("Expected permission error, got: %v", err)
	}
}

// TestPluginRegistry_EventBus проверяет что EventBus доступен.
func TestPluginRegistry_EventBus(t *testing.T) {
	registry := NewPluginRegistry()
	bus := registry.EventBus()

	if bus == nil {
		t.Fatal("EventBus should not be nil")
	}

	// Проверяем что можем подписаться на событие
	id := bus.Subscribe("test", func(data interface{}) {})
	if id <= 0 {
		t.Errorf("Subscribe should return positive ID, got %d", id)
	}

	// Проверяем publish
	received := false
	id2 := bus.Subscribe("publish_test", func(data interface{}) {
		received = true
	})
	bus.Publish("publish_test", "test_data")
	if !received {
		t.Error("Handler should have been called")
	}

	// Отписка
	bus.Unsubscribe("publish_test", id2)
}

// TestPluginContext проверяет создание и использование контекста.
func TestPluginContext(t *testing.T) {
	ctx := NewPluginContext("scan-123")

	if ctx.ScanID != "scan-123" {
		t.Errorf("Expected scan ID 'scan-123', got '%s'", ctx.ScanID)
	}

	if ctx.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}

	if ctx.CancelContext == nil {
		t.Fatal("CancelContext should not be nil")
	}

	if ctx.CancelFunc == nil {
		t.Fatal("CancelFunc should not be nil")
	}

	// Проверяем что cancel работает
	ctx.CancelFunc()

	// После cancel контекст должен быть done
	select {
	case <-ctx.CancelContext.Done():
		// OK
	default:
		t.Error("Context should be done after cancel")
	}
}

// TestPluginContextWithTimeout проверяет контекст с таймаутом.
func TestPluginContextWithTimeout(t *testing.T) {
	// Создаем контекст с ручным таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	pluginCtx := &PluginContext{
		ScanID:        "timeout-test",
		StartTime:     time.Now(),
		CancelContext: ctx,
		CancelFunc:    cancel,
	}

	// Ждем истечения таймаута
	time.Sleep(150 * time.Millisecond)

	select {
	case <-pluginCtx.CancelContext.Done():
		// OK — таймаут сработал
	case <-time.After(1 * time.Second):
		t.Error("Context should have timed out")
	}
}

// --- Plugin Loader Tests ---

// TestPluginLoader_ValidExtensions проверяет корректность расширений для ОС.
func TestPluginLoader_ValidExtensions(t *testing.T) {
	exts := ValidExtensions()
	if len(exts) == 0 {
		t.Fatal("ValidExtensions should return at least one extension")
	}

	// Проверяем что расширение содержит точку
	for _, ext := range exts {
		if ext[0] != '.' {
			t.Errorf("Extension %q should start with dot", ext)
		}
	}
}

// TestPluginLoader_NewPluginLoaderDefault проверяет создание loader с дефолтной директорией.
func TestPluginLoader_NewPluginLoaderDefault(t *testing.T) {
	loader := NewPluginLoader("")
	if loader == nil {
		t.Fatal("Loader should not be nil")
	}
}

// TestPluginLoader_LoadInvalidExtension проверяет ошибку при неверном расширении.
func TestPluginLoader_LoadInvalidExtension(t *testing.T) {
	loader := NewPluginLoader("")

	// Создаем временный файл с неверным расширением
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	_, err := loader.Load(tmpFile)
	if err == nil {
		t.Error("Load should fail with invalid extension")
	}
}

// TestPluginLoader_LoadNonExistentFile проверяет ошибку при отсутствии файла.
func TestPluginLoader_LoadNonExistentFile(t *testing.T) {
	loader := NewPluginLoader("")

	var ext string
	switch runtime.GOOS {
	case "windows":
		ext = ".dll"
	case "darwin":
		ext = ".dylib"
	default:
		ext = ".so"
	}

	tmpFile := "/nonexistent/plugin" + ext
	_, err := loader.Load(tmpFile)
	if err == nil {
		t.Error("Load should fail for non-existent file")
	}
}

// TestPluginLoader_LoadAllNonExistentDir проверяет что LoadAll возвращает пустой список для несуществующей директории.
func TestPluginLoader_LoadAllNonExistentDir(t *testing.T) {
	loader := NewPluginLoader("")

	plugins, err := loader.LoadAll("/nonexistent/dir/plugins")
	if err != nil {
		t.Fatalf("LoadAll should not error for non-existent dir, got: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins, got %d", len(plugins))
	}
}

// TestPluginLoader_LoadAllNotADirectory проверяет ошибку если путь не директория.
func TestPluginLoader_LoadAllNotADirectory(t *testing.T) {
	loader := NewPluginLoader("")

	// Создаем временный файл
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	_, err := loader.LoadAll(tmpFile)
	if err == nil {
		t.Error("LoadAll should fail when path is not a directory")
	}
}

// TestPluginLoader_Discover проверяет что Discover делегирует LoadAll.
func TestPluginLoader_Discover(t *testing.T) {
	loader := NewPluginLoader("plugins")

	// Discover должен вызвать LoadAll с директорией plugins
	// Если директория не существует, возвращаем пустой список
	plugins, err := loader.Discover()
	if err != nil {
		t.Fatalf("Discover should not error for non-existent dir, got: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins for non-existent dir, got %d", len(plugins))
	}
}

// --- Plugin Info Tests ---

// TestPluginInfoFields проверяет поля Info.
func TestPluginInfoFields(t *testing.T) {
	info := Info{
		Name:        "TestPlugin",
		Version:     "2.0.0",
		Description: "Test description",
		Author:      "Test Author",
		Type:        TypeFilter,
	}

	if info.Name != "TestPlugin" {
		t.Errorf("Expected name 'TestPlugin', got '%s'", info.Name)
	}
	if info.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", info.Version)
	}
	if info.Type != TypeFilter {
		t.Errorf("Expected type 'filter', got '%s'", info.Type)
	}
}

// TestPluginTypes проверяет все типы плагинов.
func TestPluginTypes(t *testing.T) {
	typeNames := []string{"filter", "exporter", "scanner", "reporter"}
	types := []Type{TypeFilter, TypeExporter, TypeScanner, TypeReporter}

	for i, pluginType := range types {
		if string(pluginType) != typeNames[i] {
			t.Errorf("Expected type %d to be %q, got %q", i, typeNames[i], pluginType)
		}
	}
}
