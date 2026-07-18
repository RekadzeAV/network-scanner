package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// PluginLoader загружает плагины из директории
type PluginLoader struct {
	pluginsDir string
}

// NewPluginLoader создает загрузчик плагинов
func NewPluginLoader(pluginsDir string) *PluginLoader {
	if pluginsDir == "" {
		pluginsDir = "plugins"
	}
	return &PluginLoader{
		pluginsDir: pluginsDir,
	}
}

// Load загружает плагин по пути
func (pl *PluginLoader) Load(path string) (Plugin, error) {
	// Проверяем расширение файла
	ext := filepath.Ext(path)
	var expectedExt string
	
	switch runtime.GOOS {
	case "windows":
		expectedExt = ".dll"
	case "darwin":
		expectedExt = ".dylib"
	default:
		expectedExt = ".so"
	}
	
	if ext != expectedExt {
		return nil, fmt.Errorf("unsupported plugin extension %q, expected %q", ext, expectedExt)
	}
	
	// Проверяем существование файла
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", path)
	}
	
	// TODO: Реализовать динамическую загрузку через plugin package
	// Для этого плагин должен быть скомпилирован с plugin.enabled
	
	return nil, fmt.Errorf("dynamic plugin loading not yet implemented: %s", path)
}

// LoadAll загружает все плагины из директории
func (pl *PluginLoader) LoadAll(dir string) ([]Plugin, error) {
	var plugins []Plugin
	
	// Проверяем существование директории
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// Директория не существует — возвращаем пустой список
		return plugins, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot access plugins directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("plugins path is not a directory: %s", dir)
	}
	
	// Читаем файлы в директории
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read plugins directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		path := filepath.Join(dir, entry.Name())
		plugin, err := pl.Load(path)
		if err != nil {
			// Пропускаем невалидные плагины
			continue
		}
		plugins = append(plugins, plugin)
	}
	
	return plugins, nil
}

// Discover ищет плагины в стандартной директории
func (pl *PluginLoader) Discover() ([]Plugin, error) {
	return pl.LoadAll(pl.pluginsDir)
}

// ValidExtensions возвращает допустимые расширения для плагинов на текущей ОС
func ValidExtensions() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{".dll"}
	case "darwin":
		return []string{".dylib"}
	default:
		return []string{".so"}
	}
}
