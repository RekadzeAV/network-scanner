package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	"network-scanner/internal/contracts"
)

// ============================================================================
// E7/7.3: Тесты планировщика периодических сканов (schedule) и PDF-экспорта.
// ============================================================================

// testScanResults — минимальный результат сканирования для PDF-экспорта.
func testScanResults() []contracts.ScanResult {
	return []contracts.ScanResult{
		{IP: "192.168.1.1", Hostname: "gw.local"},
		{IP: "192.168.1.10", Hostname: "srv"},
	}
}

func TestParseInterval(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    time.Duration
		wantErr bool
	}{
		{name: "часы", in: "6h", want: 6 * time.Hour},
		{name: "минуты", in: "30m", want: 30 * time.Minute},
		{name: "составной", in: "1h30m", want: 90 * time.Minute},
		{name: "секунды", in: "45s", want: 45 * time.Second},
		{name: "ноль", in: "0s", wantErr: true},
		{name: "отрицательный", in: "-5m", wantErr: true},
		{name: "пустая строка", in: "", wantErr: true},
		{name: "пробелы", in: "   ", wantErr: true},
		{name: "не duration", in: "abc", wantErr: true},
		{name: "без единиц", in: "30", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInterval(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseInterval(%q): ожидалась ошибка, got %v", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseInterval(%q): неожиданная ошибка: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("parseInterval(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestScheduleCmd_Registered — schedule доступна из rootCmd и имеет RunE.
func TestScheduleCmd_Registered(t *testing.T) {
	if scheduleCmd.RunE == nil {
		t.Fatal("scheduleCmd не имеет RunE")
	}
	if err := scheduleCmd.Args(scheduleCmd, []string{"extra"}); err == nil {
		t.Error("schedule не принимает позиционные аргументы")
	}
}

// TestScheduleCmd_HasInheritedScanFlags — планировщик наследует флаги scan
// (включая export-pdf из E7/7.3) и добавляет собственные.
func TestScheduleCmd_HasInheritedScanFlags(t *testing.T) {
	want := []string{"network", "ports", "timeout", "udp", "export-html", "export-pdf", "export-xml"}
	for _, name := range want {
		if scheduleCmd.Flags().Lookup(name) == nil {
			t.Errorf("scheduleCmd не наследует флаг --%s от scan", name)
		}
	}
	for _, name := range []string{"interval", "max-runs", "skip-first"} {
		if scheduleCmd.Flags().Lookup(name) == nil {
			t.Errorf("scheduleCmd не имеет собственного флага --%s", name)
		}
	}
	// interval по умолчанию — 6h.
	f := scheduleCmd.Flags().Lookup("interval")
	if f.DefValue != "6h" {
		t.Errorf("--interval default = %q, want %q", f.DefValue, "6h")
	}
}

// TestRunSchedule_RejectsBadInterval — ветка: runSchedule возвращает ошибку
// до запуска цикла при некорректном интервале.
func TestRunSchedule_RejectsBadInterval(t *testing.T) {
	c := scheduleCmd
	if err := c.Flags().Set("interval", "0s"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = c.Flags().Set("interval", "6h")
	})
	if err := runSchedule(c); err == nil {
		t.Fatal("ожидалась ошибка валидации интервала")
	} else if !strings.Contains(err.Error(), "положительным") {
		t.Errorf("неожиданное сообщение: %v", err)
	}
}

// TestRunSchedule_RejectsNegativeMaxRuns — ветка: max-runs < 0.
func TestRunSchedule_RejectsNegativeMaxRuns(t *testing.T) {
	c := scheduleCmd
	if err := c.Flags().Set("interval", "1h"); err != nil {
		t.Fatal(err)
	}
	if err := c.Flags().Set("max-runs", "-1"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = c.Flags().Set("interval", "6h")
		_ = c.Flags().Set("max-runs", "0")
	})
	if err := runSchedule(c); err == nil {
		t.Fatal("ожидалась ошибка для отрицательного max-runs")
	}
}

// TestRunSchedule_QuickExit — интеграция: цикл с интервалом 1ms и max-runs=1
// завершается сам, без внешнего сигнала (планировщик не зависает).
func TestRunSchedule_QuickExit(t *testing.T) {
	c := scheduleCmd
	for flag, val := range map[string]string{
		"interval": "1ms", "max-runs": "1", "skip-first": "false",
		"network": "127.0.0.1/32", "ports": "1", "timeout": "1", "threads": "2",
	} {
		if err := c.Flags().Set(flag, val); err != nil {
			t.Fatalf("set %s: %v", flag, err)
		}
	}
	t.Cleanup(func() {
		for flag, val := range map[string]string{
			"interval": "6h", "max-runs": "0", "skip-first": "false",
			"network": "", "ports": "1-1000", "timeout": "2", "threads": "50",
		} {
			_ = c.Flags().Set(flag, val)
		}
	})

	done := make(chan error, 1)
	go func() { done <- runSchedule(c) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runSchedule: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("runSchedule завис (не завершился за 30s с max-runs=1)")
	}
}

// TestExportScanPDF_CreatesFile — ветка: PDF-экспорт создаёт файл с
// валидной PDF-сигнатурой (%PDF-1.x) — проверка подключения генератора
// из internal/report к CLI (E7/7.3).
func TestExportScanPDF_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Skipf("не удалось получить рабочий каталог: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Skipf("не удалось сменить каталог: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWd)
	})

	name, err := exportScanPDF(testScanResults())
	if err != nil {
		t.Fatalf("exportScanPDF: %v", err)
	}
	if name == "" || !strings.HasSuffix(name, ".pdf") {
		t.Fatalf("неожиданное имя файла: %q", name)
	}
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("чтение %s: %v", name, err)
	}
	if len(data) < 5 || string(data[:5]) != "%PDF-" {
		prefix := data
		if len(prefix) > 5 {
			prefix = prefix[:5]
		}
		t.Errorf("сигнатура PDF не найдена, префикс %q", string(prefix))
	}
}