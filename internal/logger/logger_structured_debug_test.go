//go:build debug
// +build debug

package logger

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestLogStructured_Debug — debug-сборка пишет структурированную JSON-строку.
//
// Тест не подменяет рабочий каталог: os.Chdir в тестах небезопасен
// (процесс-глобальное состояние, мешает параллельным тестам), а Init пишет лог
// в os.Getwd() на момент вызова. Проверяем через явный путь с очисткой.
func TestLogStructured_Debug(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := Init("network-scanner", "structured-test"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(func() {
		Close()
		_ = os.Remove(logFilePathForTest(wd))
	})

	LogStructured("WARN", "policy rejected", map[string]interface{}{
		"target": "10.0.0.9",
		"reason": "not in allowlist",
	})

	data, err := os.ReadFile(logFilePathForTest(wd))
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", logFilePathForTest(wd), err)
	}

	var found map[string]interface{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		if err := json.Unmarshal([]byte(line), &found); err != nil {
			t.Fatalf("structured line is not valid JSON: %v (%s)", err, line)
		}
		break
	}
	if found == nil {
		t.Fatalf("structured log line not found in %s", string(data))
	}
	if found["level"] != "warn" {
		t.Errorf("level = %v, want warn (нормализация регистра)", found["level"])
	}
	if found["msg"] != "policy rejected" {
		t.Errorf("msg = %v", found["msg"])
	}
	if found["target"] != "10.0.0.9" {
		t.Errorf("target attribute lost: %v", found["target"])
	}
	if _, ok := found["timestamp"]; !ok {
		t.Errorf("timestamp отсутствует: %v", found)
	}
}

// logFilePathForTest вычисляет путь файла лога для рабочего каталога wd
// (Init формирует имя LOG-<app>-<version>.txt в os.Getwd()).
func logFilePathForTest(wd string) string {
	sep := string(os.PathSeparator)
	return wd + sep + "LOG-network-scanner-structured-test.txt"
}
