package api

import (
	"fmt"
	"network-scanner/internal/contracts"
	"network-scanner/internal/scanner"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// scanIDMu защищает генерацию ID от параллельного доступа
var scanIDMu sync.Mutex

// scanIDCounter используется для уникальности ID
var scanIDCounter uint64

// generateScanID генерирует уникальный ID для сканирования
func generateScanID() string {
	scanIDMu.Lock()
	defer scanIDMu.Unlock()

	// Используем counter для уникальности
	counter := atomic.AddUint64(&scanIDCounter, 1)
	return fmt.Sprintf("scan-%d-%d", time.Now().UnixNano(), counter)
}

// resetScanStore сбрасывает состояние scanStore (только для тестов)
func resetScanStore() {
	scanStoreInstance.globalMu.Lock()
	defer scanStoreInstance.globalMu.Unlock()

	scanStoreInstance.resetMu.Lock()
	defer scanStoreInstance.resetMu.Unlock()

	scanStoreInstance.mu.Lock()
	defer scanStoreInstance.mu.Unlock()

	scanStoreInstance.scans = make(map[string]*scanState)
}

// mapString читает поле map как строку и обрезает пробелы.
//
// Значения приходят из JSON, где тип заранее неизвестен: не-строка даёт "".
func mapString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

// mapBool читает поле map как bool. Отсутствующее или не-bool значение — false.
func mapBool(m map[string]interface{}, key string) bool {
	v, ok := m[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

// parsePortFromMap извлекает номер порта из JSON-объекта.
//
// JSON отдаёт числа как float64, но клиенты могут прислать порт строкой
// ("8080"). Другие типы (в том числе int из собранных вручную map) не
// поддерживаются и дают 0 — вызывающий код трактует 0 как «порт не указан».
func parsePortFromMap(m map[string]interface{}) int {
	v, ok := m["port"]
	if !ok {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case string:
		port, err := strconv.Atoi(strings.TrimSpace(t))
		if err != nil {
			return 0
		}
		return port
	default:
		return 0
	}
}

// mapToPortInfo преобразует JSON-объект порта в scanner.PortInfo.
func mapToPortInfo(m map[string]interface{}) scanner.PortInfo {
	return scanner.PortInfo{
		Port:     parsePortFromMap(m),
		State:    mapString(m, "state"),
		Protocol: mapString(m, "protocol"),
		Service:  mapString(m, "service"),
		Banner:   mapString(m, "banner"),
		Version:  mapString(m, "version"),
	}
}

// mapToScannerResult преобразует JSON-объект хоста в scanner.Result.
//
// Нужен HTTP-слою, где хосты приходят как []map[string]interface{} (теле
// POST /api/v1/alerts/check). Элементы "ports", не являющиеся объектами, и
// элементы "protocols", не являющиеся строками, пропускаются: один мусорный
// элемент не должен обрывать обработку всего запроса.
func mapToScannerResult(m map[string]interface{}) scanner.Result {
	result := scanner.Result{
		IP:                mapString(m, "ip"),
		MAC:               mapString(m, "mac"),
		Hostname:          mapString(m, "hostname"),
		DeviceType:        mapString(m, "device_type"),
		DeviceVendor:      mapString(m, "device_vendor"),
		SNMPEnabled:       mapBool(m, "snmp_enabled"),
		IsAlive:           mapBool(m, "is_alive"),
		GuessOS:           mapString(m, "guess_os"),
		GuessOSConfidence: mapString(m, "guess_os_confidence"),
		GuessOSReason:     mapString(m, "guess_os_reason"),
	}

	if raw, ok := m["ports"].([]interface{}); ok {
		for _, item := range raw {
			portMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			result.Ports = append(result.Ports, mapToPortInfo(portMap))
		}
	}

	if raw, ok := m["protocols"].([]interface{}); ok {
		for _, item := range raw {
			proto, ok := item.(string)
			if !ok {
				continue
			}
			result.Protocols = append(result.Protocols, proto)
		}
	}

	return result
}

// mapToScannerResults — пакетное преобразование списка хостов из JSON в results.
func mapToScannerResults(items []map[string]interface{}) []scanner.Result {
	results := make([]scanner.Result, 0, len(items))
	for _, item := range items {
		results = append(results, mapToScannerResult(item))
	}
	return results
}

// contractPortToScanner конвертирует contracts.PortInfo во внутренний
// scanner.PortInfo.
//
// Состояние и протокол нормализуются к нижнему регистру: контракты приходят из
// внешнего JSON ("OPEN", "TCP"), а внутренняя модель сравнивает их как есть.
func contractPortToScanner(p contracts.PortInfo) scanner.PortInfo {
	return scanner.PortInfo{
		Port:     p.Port,
		State:    strings.ToLower(strings.TrimSpace(p.State)),
		Protocol: strings.ToLower(strings.TrimSpace(p.Protocol)),
		Service:  strings.TrimSpace(p.Service),
		Banner:   strings.TrimSpace(p.Banner),
		Version:  strings.TrimSpace(p.Version),
	}
}
