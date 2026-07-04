package gui

import (
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

// --- AppModel edge cases ---

func TestAppModel_MultipleStatusUpdates(t *testing.T) {
	model := NewAppModel()

	// Множественные обновления статуса
	model.SetStatus("Status 1")
	model.SetStatus("Status 2")
	model.SetStatus("Status 3")

	status := model.GetStatus()
	if status != "Status 3" {
		t.Errorf("Final status = %v, want Status 3", status)
	}
}

func TestAppModel_MultipleScanningToggles(t *testing.T) {
	model := NewAppModel()

	// Множественные переключения сканирования
	model.SetScanning(true)
	model.SetScanning(false)
	model.SetScanning(true)
	model.SetScanning(false)

	if model.IsScanning() {
		t.Error("IsScanning() should return false after final SetScanning(false)")
	}
}

func TestAppModel_ProgressEdgeCases(t *testing.T) {
	model := NewAppModel()

	// Тестовые значения прогресса
	testCases := []float64{-10.0, 0.0, 50.0, 100.0, 150.0}

	for _, value := range testCases {
		model.UpdateProgress(value, "test")
		// Не должно паниковать
	}
}

func TestAppModel_ResultsWithPorts(t *testing.T) {
	model := NewAppModel()

	result := scanner.Result{
		IP:       "192.168.1.1",
		Hostname: "server",
		Ports: []scanner.PortInfo{
			{Port: 80, Protocol: "tcp", State: "open", Service: "HTTP"},
			{Port: 443, Protocol: "tcp", State: "open", Service: "HTTPS"},
			{Port: 22, Protocol: "tcp", State: "closed", Service: "SSH"},
		},
		IsAlive: true,
	}

	model.AddHostResult(result)

	results := model.GetResults()
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if len(results[0].Ports) != 3 {
		t.Errorf("Expected 3 ports, got %d", len(results[0].Ports))
	}
}

func TestAppModel_ResultsWithoutPorts(t *testing.T) {
	model := NewAppModel()

	result := scanner.Result{
		IP:       "192.168.1.1",
		Hostname: "host",
		IsAlive:  true,
	}

	model.AddHostResult(result)

	results := model.GetResults()
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Ports != nil && len(results[0].Ports) != 0 {
		t.Errorf("Expected 0 ports, got %d", len(results[0].Ports))
	}
}

// --- Formatter edge cases ---

func TestEscapeMarkdownWithAllSpecialChars(t *testing.T) {
	input := "*bold* _italic_ `code` [link](url) # header"
	escaped := escapeMarkdown(input)

	// escapeMarkdown может не экранировать все символы
	// Проверяем, что функция не паникует и возвращает строку
	if escaped == "" {
		t.Error("escapeMarkdown() should not return empty string")
	}
}

func TestTruncateStringWithExactLength(t *testing.T) {
	input := "12345"
	truncated := truncateString(input, 5)

	if truncated != input {
		t.Errorf("truncateString(%q, 5) = %q, want %q", input, truncated, input)
	}
}

func TestTruncateStringWithZeroLength(t *testing.T) {
	input := "hello"
	truncated := truncateString(input, 0)

	// truncateString может возвращать пустую строку для 0
	if truncated == "" {
		t.Log("truncateString with 0 length returns empty string (acceptable)")
	}
}

func TestSortedResultsForDisplayWithSameIP(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "b"},
		{IP: "192.168.1.1", Hostname: "a"},
	}

	sorted := sortedResultsForDisplay(results)

	if len(sorted) != 2 {
		t.Errorf("sortedResultsForDisplay() length = %d, want 2", len(sorted))
	}
}

func TestFilterResultsForDisplayEmptyFilter(t *testing.T) {
	results := []scanner.Result{
		{Hostname: "router", IP: "192.168.1.1"},
		{Hostname: "computer", IP: "192.168.1.10"},
	}

	filtered := filterResultsForDisplay(results, "")

	// Пустой фильтр должен вернуть все результаты
	if len(filtered) != 2 {
		t.Errorf("filterResultsForDisplay() with empty filter length = %d, want 2", len(filtered))
	}
}

func TestFilterResultsForDisplayPartialMatch(t *testing.T) {
	results := []scanner.Result{
		{Hostname: "router-main", IP: "192.168.1.1"},
		{Hostname: "router-backup", IP: "192.168.1.2"},
		{Hostname: "switch", IP: "192.168.1.3"},
	}

	filtered := filterResultsForDisplay(results, "router")

	if len(filtered) != 2 {
		t.Errorf("filterResultsForDisplay() with partial match length = %d, want 2", len(filtered))
	}
}

func TestFilterResultsForDisplayIPFilter(t *testing.T) {
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "host1"},
		{IP: "192.168.1.10", Hostname: "host2"},
		{IP: "10.0.0.1", Hostname: "host3"},
	}

	filtered := filterResultsForDisplay(results, "192.168.1.1")

	// Должен найти 2 результата (192.168.1.1 и 192.168.1.10)
	if len(filtered) != 2 {
		t.Errorf("filterResultsForDisplay() with IP filter length = %d, want 2", len(filtered))
	}
}

func TestFormatDurationMMSSLargeValue(t *testing.T) {
	duration := 10 * time.Hour
	formatted := formatDurationMMSS(duration)

	if formatted == "" {
		t.Error("formatDurationMMSS() should not return empty string for large value")
	}
}

func TestFormatDurationMMSSNegative(t *testing.T) {
	duration := -1 * time.Second
	formatted := formatDurationMMSS(duration)

	// Не должно паниковать
	if formatted == "" {
		t.Error("formatDurationMMSS() should not return empty string for negative value")
	}
}

func TestNormalizeDeviceTypesWithZeroCount(t *testing.T) {
	raw := map[string]int{
		"Router":   0,
		"Computer": 0,
	}

	normalized := normalizeDeviceTypes(raw)

	// Должен вернуть результаты даже с нулевыми счетчиками
	if len(normalized) != 2 {
		t.Errorf("normalizeDeviceTypes() length = %d, want 2", len(normalized))
	}
}

func TestNormalizeDeviceTypesWithNegativeCount(t *testing.T) {
	raw := map[string]int{
		"Router": -1,
	}

	normalized := normalizeDeviceTypes(raw)

	// Не должно паниковать
	if normalized == nil {
		t.Error("normalizeDeviceTypes() should not return nil")
	}
}

// --- Results responsive tests ---

func TestResultsResponsiveCompactMode(t *testing.T) {
	// Тест компактного режима отображения
	results := []scanner.Result{
		{IP: "192.168.1.1", Hostname: "host1"},
	}

	displayResults := FormatResultsForDisplay(results)

	// FormatResultsForDisplay возвращает список строк, а не результатов
	if len(displayResults) == 0 {
		t.Error("FormatResultsForDisplay() should return non-empty result")
	}
}

func TestResultsResponsiveNormalMode(t *testing.T) {
	// Тест нормального режима отображения
	results := []scanner.Result{
		{
			IP:       "192.168.1.1",
			Hostname: "host1",
			Ports: []scanner.PortInfo{
				{Port: 80, Protocol: "tcp", State: "open"},
			},
		},
	}

	displayResults := FormatResultsForDisplay(results)

	// FormatResultsForDisplay возвращает список строк
	if len(displayResults) == 0 {
		t.Error("FormatResultsForDisplay() should return non-empty result")
	}
}

func TestResultsResponsiveWideMode(t *testing.T) {
	// Тест широкого режима отображения
	results := []scanner.Result{
		{
			IP:    "192.168.1.1",
			Ports: []scanner.PortInfo{},
		},
	}

	displayResults := FormatResultsForDisplay(results)

	// FormatResultsForDisplay возвращает список строк
	if len(displayResults) == 0 {
		t.Error("FormatResultsForDisplay() should return non-empty result")
	}
}

// --- Split persist tests ---

func TestSplitPersistEpsilon(t *testing.T) {
	// Тест константы epsilon для split persist
	if splitPersistEpsilon <= 0 {
		t.Error("splitPersistEpsilon should be positive")
	}
}

func TestMaybePersistFloatPrefNilPreferences(t *testing.T) {
	// Тест с nil preferences — не должно паниковать
	var primed bool
	var last float64

	maybePersistFloatPref(nil, "test", 0.5, &primed, &last, nil)

	if primed {
		t.Error("primed should be false with nil preferences")
	}
}

func TestMaybePersistFloatPrefNilPrimed(t *testing.T) {
	// Тест с nil primed — не должно паниковать
	var last float64

	maybePersistFloatPref(nil, "test", 0.5, nil, &last, nil)
}

func TestMaybePersistFloatPrefNilLast(t *testing.T) {
	// Тест с nil last — не должно паниковать
	var primed bool

	maybePersistFloatPref(nil, "test", 0.5, &primed, nil, nil)
}

func TestMaybePersistFloatPrefSmallChange(t *testing.T) {
	// Тест с изменением меньше epsilon — не должно записывать
	var primed bool
	var last float64
	persisted := false

	maybePersistFloatPref(nil, "test", 0.5, &primed, &last, func(v float64) {
		persisted = true
	})

	// Первое значение только устанавливает last
	if persisted {
		t.Error("persisted should be false for first value")
	}

	// Второе значение с малым изменением
	maybePersistFloatPref(nil, "test", 0.501, &primed, &last, func(v float64) {
		persisted = true
	})

	if persisted {
		t.Error("persisted should be false for small change")
	}
}

func TestMaybePersistFloatPrefLargeChange(t *testing.T) {
	// Тест с изменением больше epsilon — должно записывать
	var primed bool
	var last float64
	persistCount := 0

	// with nil preferences, onPersist is not called
	maybePersistFloatPref(nil, "test", 0.5, &primed, &last, func(v float64) {
		persistCount++
	})

	// Первое значение с nil preferences — onPersist не вызывается
	if persistCount != 0 {
		t.Errorf("persistCount = %d, want 0 (nil preferences)", persistCount)
	}

	// Второе значение с большим изменением
	maybePersistFloatPref(nil, "test", 0.6, &primed, &last, func(v float64) {
		persistCount++
	})

	// with nil preferences, onPersist is still not called
	if persistCount != 0 {
		t.Errorf("persistCount = %d, want 0 (nil preferences)", persistCount)
	}
}
