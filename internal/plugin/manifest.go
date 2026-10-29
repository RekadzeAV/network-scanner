package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ManifestFile — имя файла манифеста плагина в директории плагина.
const ManifestFile = "plugin.json"

// Manifest — описание плагина в plugin.json (L2.2).
//
// Размещается рядом с shared library в поддиректории директории плагинов:
//
//	plugins/
//	  os-filter/
//	    plugin.json     <- манифест
//	    osfilter.so     <- entry
//
// Манифест обязателен для загрузки плагина из поддиректории: метаданные нужны
// до открытия .so (plugin.Open может завершиться ошибкой на битом файле,
// а список плагинов CLI/UI показывает ещё до загрузки).
type Manifest struct {
	// Name — уникальное имя плагина.
	Name string `json:"name"`
	// Version — версия плагина (semver-подобная строка).
	Version string `json:"version"`
	// Description — краткое описание для UI/CLI.
	Description string `json:"description,omitempty"`
	// Author — автор.
	Author string `json:"author,omitempty"`
	// Type — тип плагина (filter/exporter/scanner/reporter).
	Type Type `json:"type"`
	// Entry — имя файла shared library внутри директории плагина.
	Entry string `json:"entry"`
	// Enabled — false пропускает плагин при загрузке (выключен пользователем).
	Enabled *bool `json:"enabled,omitempty"`
	// MinVersion — минимальная версия приложения (пусто — без проверки).
	MinVersion string `json:"minVersion,omitempty"`
}

// IsEnabled возвращает effective-значение enabled (nil == включён).
func (m *Manifest) IsEnabled() bool {
	return m.Enabled == nil || *m.Enabled
}

// Validate проверяет обязательные поля манифеста.
func (m *Manifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("manifest: пустое имя плагина")
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("manifest %q: пустая версия", m.Name)
	}
	if strings.TrimSpace(m.Entry) == "" {
		return fmt.Errorf("manifest %q: пустой entry", m.Name)
	}
	// Entry не должен выходить за пределы директории плагина (path traversal).
	if filepath.Base(m.Entry) != m.Entry {
		return fmt.Errorf("manifest %q: entry должен быть именем файла, а не путём: %q", m.Name, m.Entry)
	}
	switch m.Type {
	case TypeFilter, TypeExporter, TypeScanner, TypeReporter:
	default:
		return fmt.Errorf("manifest %q: неизвестный тип %q", m.Name, m.Type)
	}
	return nil
}

// LoadManifest читает и валидирует plugin.json из директории плагина.
func LoadManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, ManifestFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("manifest %s: %w", path, err)
	}
	var m Manifest
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("manifest %s: %w", path, err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// SaveManifest записывает манифест в директорию плагина (для генераторов/примеров).
func SaveManifest(dir string, m *Manifest) error {
	if err := m.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ManifestFile), data, 0o600)
}
