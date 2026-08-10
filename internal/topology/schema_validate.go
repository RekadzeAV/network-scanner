package topology

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONSchemaVersion определяет текущую версию схемы JSON.
const JSONSchemaVersion = "1.0.0"

// JSONSchema описывает структуру JSON-экспорта топологии.
type JSONSchema struct {
	SchemaVersion string      `json:"$schema_version"`
	Topology      *Topology   `json:"topology"`
	Validation    SchemaCheck `json:"validation"`
}

// SchemaCheck содержит результаты валидации схемы.
type SchemaCheck struct {
	Valid         bool     `json:"valid"`
	DeviceCount   int      `json:"device_count"`
	LinkCount     int      `json:"link_count"`
	DeviceTypes   []string `json:"device_types"`
	SourceTypes   []string `json:"source_types"`
	ConfidenceLvl []string `json:"confidence_levels"`
	Errors        []string `json:"errors,omitempty"`
}

// ValidateJSONSchema проверяет топологию на соответствие JSON-схеме.
// Возвращает SchemaCheck с результатами валидации.
func (t *Topology) ValidateJSONSchema() *SchemaCheck {
	check := &SchemaCheck{
		Valid: true,
	}

	if t == nil {
		check.Valid = false
		check.Errors = append(check.Errors, "topology is nil")
		return check
	}

	if t.Devices == nil || len(t.Devices) == 0 {
		check.Valid = false
		check.Errors = append(check.Errors, "devices map is empty or nil")
	}

	if len(t.Links) > 0 && t.Devices == nil {
		check.Valid = false
		check.Errors = append(check.Errors, "links exist but devices map is nil")
	}

	// Подсчёт устройств и связей
	check.DeviceCount = len(t.Devices)
	check.LinkCount = len(t.Links)

	// Сбор уникальных типов устройств
	deviceTypes := make(map[string]bool)
	for _, d := range t.Devices {
		if d == nil {
			check.Valid = false
			check.Errors = append(check.Errors, "nil device in devices map")
			continue
		}
		if strings.TrimSpace(string(d.Type)) == "" {
			check.Valid = false
			check.Errors = append(check.Errors, fmt.Sprintf("device %q has empty type", nodeID(d)))
		}
		deviceTypes[string(d.Type)] = true
	}

	// Сбор уникальных типов источников и уровней уверенности
	sourceTypes := make(map[string]bool)
	confidenceLevels := make(map[string]bool)
	validDeviceIDs := make(map[string]bool)

	for _, d := range t.Devices {
		if d != nil {
			validDeviceIDs[nodeID(d)] = true
		}
	}

	for i, l := range t.Links {
		if l.Source == nil || l.Target == nil {
			check.Valid = false
			check.Errors = append(check.Errors, fmt.Sprintf("link[%d] has nil endpoint", i))
			continue
		}

		srcID := nodeID(l.Source)
		dstID := nodeID(l.Target)

		if !validDeviceIDs[srcID] {
			check.Valid = false
			check.Errors = append(check.Errors, fmt.Sprintf("link[%d] source %q not in devices", i, srcID))
		}
		if !validDeviceIDs[dstID] {
			check.Valid = false
			check.Errors = append(check.Errors, fmt.Sprintf("link[%d] target %q not in devices", i, dstID))
		}

		sourceTypes[string(l.SourceType)] = true
		confidenceLevels[string(l.Confidence)] = true

		// Проверка валидных значений source_type
		switch l.SourceType {
		case LinkSourceLLDP, LinkSourceFDB, LinkSourceInferred:
			// OK
		default:
			check.Valid = false
			check.Errors = append(check.Errors, fmt.Sprintf("link[%d] invalid source_type: %q", i, l.SourceType))
		}

		// Проверка валидных значений confidence
		switch l.Confidence {
		case LinkConfidenceHigh, LinkConfidenceMedium, LinkConfidenceLow:
			// OK
		default:
			check.Valid = false
			check.Errors = append(check.Errors, fmt.Sprintf("link[%d] invalid confidence: %q", i, l.Confidence))
		}
	}

	for dt := range deviceTypes {
		check.DeviceTypes = append(check.DeviceTypes, dt)
	}
	for st := range sourceTypes {
		check.SourceTypes = append(check.SourceTypes, st)
	}
	for cl := range confidenceLevels {
		check.ConfidenceLvl = append(check.ConfidenceLvl, cl)
	}

	// Сортировка для детерминированного вывода
	sortStrings(check.DeviceTypes)
	sortStrings(check.SourceTypes)
	sortStrings(check.ConfidenceLvl)

	return check
}

// ToJSONSchema создаёт JSON-структуру с метаданными схемы.
func (t *Topology) ToJSONSchema() ([]byte, error) {
	check := t.ValidateJSONSchema()

	schema := JSONSchema{
		SchemaVersion: JSONSchemaVersion,
		Topology:      t,
		Validation:    *check,
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal JSON schema: %w", err)
	}

	return data, nil
}

// MarshalJSONToJSON marshals topology to JSON bytes.
func (t *Topology) MarshalJSONToJSON() ([]byte, error) {
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("topology validation failed: %w", err)
	}
	return json.MarshalIndent(t, "", "  ")
}

// GraphMLEquivalence проверяет, что json и graphml экспорты содержат
// одинаковый набор устройств и связей.
func (t *Topology) GraphMLEquivalence() *EquivalenceCheck {
	check := &EquivalenceCheck{
		Match: true,
	}

	if t == nil {
		check.Match = false
		check.Errors = append(check.Errors, "topology is nil")
		return check
	}

	// JSON: устройства и связи
	jsonDevices := make(map[string]bool)
	for _, d := range t.Devices {
		if d != nil {
			jsonDevices[nodeID(d)] = true
		}
	}

	jsonLinks := len(t.Links)

	// GraphML: парсим сгенерированный XML
	gmlData, err := t.marshalGraphML()
	if err != nil {
		check.Match = false
		check.Errors = append(check.Errors, fmt.Sprintf("graphml marshal error: %v", err))
		return check
	}

	// Парсим GraphML для проверки
	gmlDevices, gmlLinks, gmlErrors := parseGraphMLOrder(gmlData)
	if len(gmlErrors) > 0 {
		check.Match = false
		for _, e := range gmlErrors {
			check.Errors = append(check.Errors, e)
		}
		return check
	}

	// Сравнение устройств
	if len(jsonDevices) != len(gmlDevices) {
		check.Match = false
		check.Errors = append(check.Errors, fmt.Sprintf("device count mismatch: json=%d, graphml=%d", len(jsonDevices), len(gmlDevices)))
	} else {
		for id := range jsonDevices {
			if !gmlDevices[id] {
				check.Match = false
				check.Errors = append(check.Errors, fmt.Sprintf("device %q in JSON but not in GraphML", id))
			}
		}
		for id := range gmlDevices {
			if !jsonDevices[id] {
				check.Match = false
				check.Errors = append(check.Errors, fmt.Sprintf("device %q in GraphML but not in JSON", id))
			}
		}
	}

	// Сравнение связей
	if jsonLinks != gmlLinks {
		check.Match = false
		check.Errors = append(check.Errors, fmt.Sprintf("link count mismatch: json=%d, graphml=%d", jsonLinks, gmlLinks))
	}

	check.JSONDevices = len(jsonDevices)
	check.GraphMLDevices = len(gmlDevices)
	check.JSONLinks = jsonLinks
	check.GraphMLLinks = gmlLinks

	return check
}

// marshalsGraphML marshals topology to GraphML XML bytes.
func (t *Topology) marshalGraphML() ([]byte, error) {
	// Используем внутреннюю логику SaveGraphML без записи в файл
	// Это упрощённая версия для сравнения
	return t.SaveGraphMLToBytes()
}

// parseGraphMLOrder парсит GraphML XML и извлекает устройства и связи.
func parseGraphMLOrder(data []byte) (map[string]bool, int, []string) {
	devices := make(map[string]bool)
	links := 0
	errors := []string{}

	// Простой парсинг через string operations (для smoke-теста достаточно)
	content := string(data)

	// Подсчёт nodes
	nodeCount := strings.Count(content, `<node id="`)
	if nodeCount > 0 {
		// Извлекаем ID узлов
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, `<node id="`) {
				start := strings.Index(line, `id="`) + 4
				end := strings.Index(line[start:], `"`)
				if end > 0 && start > 0 {
					id := line[start : start+end]
					devices[id] = true
				}
			}
		}
	}

	// Подсчёт edges
	links = strings.Count(content, `<edge `)

	return devices, links, errors
}

// EquivalenceCheck содержит результаты проверки эквивалентности json/graphml.
type EquivalenceCheck struct {
	Match          bool     `json:"match"`
	JSONDevices    int      `json:"json_devices"`
	GraphMLDevices int      `json:"graphml_devices"`
	JSONLinks      int      `json:"json_links"`
	GraphMLLinks   int      `json:"graphml_links"`
	Errors         []string `json:"errors,omitempty"`
}

// sortStrings сортирует строковый срез.
func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}
