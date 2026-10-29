package plugin

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	goplugin "plugin"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"network-scanner/internal/logger"
)

// EnvPluginsDir — переменная окружения со списком дополнительных директорий
// плагинов (разделитель — os.PathListSeparator, как в PATH). L2.2.
const EnvPluginsDir = "NETWORK_SCANNER_PLUGINS_DIR"

// PluginLoader загружает плагины из директорий (L2.2 dynamic loader).
//
// Поддерживает две структуры в директории:
//
//	plugins/
//	  os-filter/            <- поддиректория с манифестом (рекомендуется)
//	    plugin.json
//	    osfilter.so
//	  legacy.so             <- плоский shared library (обратная совместимость)
//
// Ограничение Go runtime: открытые через plugin.Open библиотеки нельзя
// выгрузить. Reload перечитывает манифесты и повторно открывает библиотеки
// (Go переиспользует уже загруженный образ), поэтому полная замена кода
// плагина без перезапуска процесса невозможна — WatchForChanges полезен
// для подхвата НОВЫХ плагинов и изменений манифестов (enabled/description).
type PluginLoader struct {
	pluginsDir string         // основная директория (обратная совместимость)
	dirs       []string       // все директории поиска (L2.2)
	mu         sync.Mutex     // защита loaded/loadErrors при Reload/Watch
	loaded     []Plugin       // успешно загруженные плагины
	loadErrors []LoadError    // ошибки последней загрузки (для CLI/UI диагностики)
	statuses   []PluginStatus // состояние обнаруженных манифестов (L2.2 CLI)
}

// PluginStatus — состояние обнаруженного плагина для диагностики (CLI/UI).
// Err=nil и Loaded=false означает «выключен манифестом».
type PluginStatus struct {
	Dir      string
	Manifest Manifest
	Loaded   bool
	Err      error
}

// LoadError — ошибка загрузки одного плагина (не прерывает LoadAll).
type LoadError struct {
	Path string
	Err  error
}

func (le LoadError) Error() string {
	return fmt.Sprintf("%s: %v", le.Path, le.Err)
}

// NewPluginLoader создает загрузчик плагинов.
// Пустая строка — директория "plugins" по умолчанию.
func NewPluginLoader(pluginsDir string) *PluginLoader {
	if pluginsDir == "" {
		pluginsDir = "plugins"
	}
	return &PluginLoader{
		pluginsDir: pluginsDir,
		dirs:       []string{pluginsDir},
	}
}

// NewPluginLoaderFromEnv — загрузчик со списком директорий:
// NETWORK_SCANNER_PLUGINS_DIR (список через os.PathListSeparator) + "./plugins"
// + пользовательская директория конфигурации. L2.2.
func NewPluginLoaderFromEnv() *PluginLoader {
	dirs := SearchPaths()
	main := "plugins"
	if len(dirs) > 0 {
		main = dirs[0]
	}
	return &PluginLoader{
		pluginsDir: main,
		dirs:       dirs,
	}
}

// SearchPaths возвращает директории поиска плагинов (L2.2):
//  1. NETWORK_SCANNER_PLUGINS_DIR (несколько путей через os.PathListSeparator);
//  2. ./plugins (рабочая директория);
//  3. <user config dir>/network-scanner/plugins.
func SearchPaths() []string {
	var dirs []string
	seen := make(map[string]bool)
	add := func(d string) {
		d = strings.TrimSpace(d)
		if d == "" {
			return
		}
		abs, err := filepath.Abs(d)
		if err != nil {
			abs = d
		}
		if seen[abs] {
			return
		}
		seen[abs] = true
		dirs = append(dirs, abs)
	}
	if env := os.Getenv(EnvPluginsDir); env != "" {
		for _, d := range filepath.SplitList(env) {
			add(d)
		}
	}
	add("plugins")
	if cfg, err := os.UserConfigDir(); err == nil && cfg != "" {
		add(filepath.Join(cfg, "network-scanner", "plugins"))
	}
	return dirs
}

// Dirs возвращает копию списка директорий поиска загрузчика.
func (pl *PluginLoader) Dirs() []string {
	out := make([]string, len(pl.dirs))
	copy(out, pl.dirs)
	return out
}

// Load загружает плагин по пути к shared library.
func (pl *PluginLoader) Load(path string) (Plugin, error) {
	// Проверяем расширение файла.
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

	// Проверяем существование файла.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", path)
	}

	// Динамическая загрузка через plugin package (CGO).
	p, err := goplugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin %s: %w", path, err)
	}

	// Ищем символ "New" — конструктор плагина.
	sym, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("plugin %s: missing 'New' symbol: %w", path, err)
	}

	// Преобразуем в функцию-конструктор.
	newFunc, ok := sym.(func() (Plugin, error))
	if !ok {
		return nil, fmt.Errorf("plugin %s: 'New' symbol is not a constructor", path)
	}

	// Создаём экземпляр плагина.
	pg, err := newFunc()
	if err != nil {
		return nil, fmt.Errorf("plugin %s: failed to create instance: %w", path, err)
	}

	return pg, nil
}

// LoadAll загружает все плагины из одной директории (L2.2):
//   - поддиректории с plugin.json: манифест решает, загружать ли entry;
//   - плоские shared library-файлы грузятся напрямую (обратная совместимость).
//
// Ошибки одного плагина не прерывают загрузку остальных — они накапливаются
// в LoadErrors() и логируются.
func (pl *PluginLoader) LoadAll(dir string) ([]Plugin, error) {
	plugins, errs, statuses, dirErr := pl.loadDir(dir)
	pl.mu.Lock()
	pl.loadErrors = errs
	pl.statuses = statuses
	pl.mu.Unlock()
	return plugins, dirErr
}

// loadDir — загрузка одной директории без записи состояния.
func (pl *PluginLoader) loadDir(dir string) ([]Plugin, []LoadError, []PluginStatus, error) {
	var plugins []Plugin
	var errs []LoadError
	var statuses []PluginStatus

	// Проверяем существование директории.
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// Директория не существует — возвращаем пустой список.
		return plugins, nil, nil, nil
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot access plugins directory: %w", err)
	}
	if !info.IsDir() {
		return nil, nil, nil, fmt.Errorf("plugins path is not a directory: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot read plugins directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			// Поддиректория с манифестом — рекомендуемая структура L2.2.
			m, err := LoadManifest(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue // без манифеста — не плагин
				}
				errs = append(errs, LoadError{Path: path, Err: err})
				statuses = append(statuses, PluginStatus{Dir: path, Manifest: Manifest{}, Err: err})
				logger.LogError(err, "plugin manifest "+path)
				continue
			}
			if !m.IsEnabled() {
				logger.Log("plugin %q выключен манифестом, пропущен", m.Name)
				statuses = append(statuses, PluginStatus{Dir: path, Manifest: *m})
				continue
			}
			libPath := filepath.Join(path, m.Entry)
			p, err := pl.Load(libPath)
			if err != nil {
				errs = append(errs, LoadError{Path: libPath, Err: err})
				statuses = append(statuses, PluginStatus{Dir: path, Manifest: *m, Err: err})
				logger.LogError(err, "plugin load "+m.Name)
				continue
			}
			plugins = append(plugins, p)
			statuses = append(statuses, PluginStatus{Dir: path, Manifest: *m, Loaded: true})
			continue
		}
		// Плоский shared library (обратная совместимость).
		if !hasPluginExt(entry.Name()) {
			continue
		}
		p, err := pl.Load(path)
		if err != nil {
			errs = append(errs, LoadError{Path: path, Err: err})
			logger.LogError(err, "plugin load "+path)
			continue
		}
		plugins = append(plugins, p)
	}

	return plugins, errs, statuses, nil
}

// DiscoverAndLoad загружает плагины из всех директорий поиска (L2.2).
// Ошибки и статусы агрегируются по всем директориям.
func (pl *PluginLoader) DiscoverAndLoad() ([]Plugin, error) {
	var all []Plugin
	var errs []LoadError
	var statuses []PluginStatus
	var firstErr error
	for _, dir := range pl.dirs {
		plugins, dirErrs, dirStatuses, err := pl.loadDir(dir)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			logger.LogError(err, "plugins discover "+dir)
			continue
		}
		all = append(all, plugins...)
		errs = append(errs, dirErrs...)
		statuses = append(statuses, dirStatuses...)
	}
	pl.mu.Lock()
	pl.loaded = all
	pl.loadErrors = errs
	pl.statuses = statuses
	pl.mu.Unlock()
	return all, firstErr
}

// Statuses возвращает состояние манифестов последней загрузки (L2.2).
func (pl *PluginLoader) Statuses() []PluginStatus {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	out := make([]PluginStatus, len(pl.statuses))
	copy(out, pl.statuses)
	return out
}

// LoadErrors возвращает ошибки последней загрузки (манифесты/открытие библиотек).
func (pl *PluginLoader) LoadErrors() []LoadError {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	out := make([]LoadError, len(pl.loadErrors))
	copy(out, pl.loadErrors)
	return out
}

// LoadedPlugins возвращает плагины последней DiscoverAndLoad/Reload.
func (pl *PluginLoader) LoadedPlugins() []Plugin {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	out := make([]Plugin, len(pl.loaded))
	copy(out, pl.loaded)
	return out
}

// Discover ищет плагины в основной директории.
func (pl *PluginLoader) Discover() ([]Plugin, error) {
	return pl.LoadAll(pl.pluginsDir)
}

// Reload перечитывает все директории поиска (L2.2 hot-reload).
//
// Ограничение Go: ранее открытые .so не выгружаются, поэтому Reload
// перечитывает манифесты и подхватывает НОВЫЕ плагины; код уже загруженного
// плагина обновляется только перезапуском процесса.
func (pl *PluginLoader) Reload() ([]Plugin, error) {
	plugins, err := pl.DiscoverAndLoad()
	if err == nil {
		logger.Log("plugin loader: reload завершён, загружено плагинов: %d", len(plugins))
	}
	return plugins, err
}

// WatchForChanges запускает поллер директорий поиска (L2.2).
// При изменении содержимого (path/size/mtime файлов и поддиректорий) вызывает
// onChange с результатом Reload. Возвращает функцию остановки; повторный вызов
// безопасен.
func (pl *PluginLoader) WatchForChanges(interval time.Duration, onChange func([]Plugin, error)) func() {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	stop := make(chan struct{})
	stopOnce := &sync.Once{}
	sig := pl.dirsSignature()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				newSig := pl.dirsSignature()
				if newSig == sig {
					continue
				}
				sig = newSig
				plugins, err := pl.Reload()
				if onChange != nil {
					onChange(plugins, err)
				}
			}
		}
	}()

	return func() { stopOnce.Do(func() { close(stop) }) }
}

// dirsSignature — рекурсивная подпись содержимого директорий поиска:
// path+size+mtime каждого файла и каталога. Изменения внутри поддиректорий
// плагинов (plugin.json, .so) тоже меняют подпись.
func (pl *PluginLoader) dirsSignature() string {
	var sb strings.Builder
	for _, dir := range pl.dirs {
		info, err := os.Stat(dir)
		if err != nil {
			sb.WriteString(dir + ":missing;")
			continue
		}
		sb.WriteString(dir + ":dir:" + sigEntry(info) + ";")
		// WalkDir обходит дерево в лексикографическом порядке — детерминированно.
		walkErr := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				sb.WriteString(p + ":error;")
				return nil // продолжаем, недоступные ветки не должны ломать подпись
			}
			if p == dir {
				return nil // корень уже учтён
			}
			fi, statErr := d.Info()
			if statErr != nil {
				sb.WriteString(p + ":error;")
				return nil
			}
			kind := "file:"
			if d.IsDir() {
				kind = "dir:"
			}
			sb.WriteString(p + ":" + kind + sigEntry(fi) + ";")
			return nil
		})
		if walkErr != nil {
			sb.WriteString(dir + ":walkerror:" + walkErr.Error() + ";")
		}
	}
	return sb.String()
}

// sigEntry — компактное представление size+mtime для подписи.
func sigEntry(info fs.FileInfo) string {
	return strconv.FormatInt(info.Size(), 10) + "@" + strconv.FormatInt(info.ModTime().UnixNano(), 10)
}

// hasPluginExt — является ли имя файла shared library для текущей ОС.
func hasPluginExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, valid := range ValidExtensions() {
		if ext == valid {
			return true
		}
	}
	return false
}

// ValidExtensions возвращает допустимые расширения для плагинов на текущей ОС.
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

// LoadBuiltin загружает встроенные плагины.
func LoadBuiltin() ([]Plugin, error) {
	plugins := []Plugin{
		NewOSFilter(),
		NewCSVExporter(),
	}
	return plugins, nil
}
