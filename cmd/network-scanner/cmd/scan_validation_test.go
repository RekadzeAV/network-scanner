package cmd

import (
	"network-scanner/internal/builder"
	"path/filepath"
	"testing"
)

// ============================================================================
// Тесты валидации ввода legacy-пути scan (P0, Этап 1).
//
// Раньше неверный --timeout/--threads молча оставлял дефолт (fmt.Sscanf
// игнорировал ошибку), а экстремальные значения уходили в сканер. Теперь
// parsePositiveInt и validate* возвращают явную ошибку.
// ============================================================================

func TestParsePositiveInt(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		value   string
		want    int
		wantErr bool
	}{
		{name: "valid", flag: "--timeout", value: "30", want: 30},
		{name: "valid with spaces", flag: "--threads", value: " 100 ", want: 100},
		{name: "one", flag: "--timeout", value: "1", want: 1},
		{name: "empty", flag: "--timeout", value: "", wantErr: true},
		{name: "spaces only", flag: "--timeout", value: "   ", wantErr: true},
		{name: "not a number", flag: "--timeout", value: "abc", wantErr: true},
		{name: "negative", flag: "--timeout", value: "-5", wantErr: true},
		{name: "zero", flag: "--threads", value: "0", wantErr: true},
		{name: "float", flag: "--timeout", value: "1.5", wantErr: true},
		{name: "overflow", flag: "--timeout", value: "99999999999999999999", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePositiveInt(tt.flag, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ожидалась ошибка, получено значение %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if got != tt.want {
				t.Errorf("parsePositiveInt(%q) = %d, ожидается %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestValidateScanTimeoutSeconds(t *testing.T) {
	tests := []struct {
		name    string
		sec     int
		wantErr bool
	}{
		{name: "min valid", sec: minScanTimeoutSec},
		{name: "default valid", sec: 2},
		{name: "max valid", sec: maxScanTimeoutSec},
		{name: "zero rejected", sec: 0, wantErr: true},
		{name: "negative rejected", sec: -1, wantErr: true},
		{name: "over max rejected", sec: maxScanTimeoutSec + 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScanTimeoutSeconds("--timeout", tt.sec)
			if tt.wantErr && err == nil {
				t.Errorf("validateScanTimeoutSeconds(%d) = nil, ожидалась ошибка", tt.sec)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateScanTimeoutSeconds(%d) = %v, ожидается nil", tt.sec, err)
			}
		})
	}
}

func TestValidateScanCIDR(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		wantErr bool
	}{
		{name: "valid /24", cidr: "192.168.1.0/24"},
		{name: "valid with spaces", cidr: " 10.0.0.0/8 "},
		{name: "valid IPv6", cidr: "2001:db8::/64"},
		{name: "empty rejected", cidr: "", wantErr: true},
		{name: "no mask rejected", cidr: "192.168.1.1", wantErr: true},
		{name: "garbage rejected", cidr: "not-a-cidr", wantErr: true},
		{name: "bad mask rejected", cidr: "192.168.1.0/33", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScanCIDR(tt.cidr)
			if tt.wantErr && err == nil {
				t.Errorf("validateScanCIDR(%q) = nil, ожидалась ошибка", tt.cidr)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateScanCIDR(%q) = %v, ожидается nil", tt.cidr, err)
			}
		})
	}
}

func TestValidateScanPortRange(t *testing.T) {
	tests := []struct {
		name      string
		portRange string
		wantErr   bool
	}{
		{name: "default", portRange: "1-1000"},
		{name: "list", portRange: "22,80,443"},
		{name: "mixed", portRange: "80,443-445"},
		{name: "single", portRange: "80"},
		{name: "empty rejected", portRange: "", wantErr: true},
		{name: "non numeric rejected", portRange: "abc", wantErr: true},
		// Примечание: существующий network.ParsePortRange не проверяет ни
		// инвертированные диапазоны («10-5» → пустой список, nil), ни верхнюю
		// границу 65535 (parseInt принимает любое int). Валидация порогов —
		// отдельная задача в ядре; здесь фиксируем текущий контракт.
		{name: "inverted accepted by core", portRange: "10-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScanPortRange(tt.portRange)
			if tt.wantErr && err == nil {
				t.Errorf("validateScanPortRange(%q) = nil, ожидалась ошибка", tt.portRange)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateScanPortRange(%q) = %v, ожидается nil", tt.portRange, err)
			}
		})
	}
}

// ============================================================================
// Интеграционные тесты RunScan → ExecuteLegacyScan (legacy CLI-путь).
// ============================================================================
// Переменные ниже вынесены наружу, чтобы они жили до/после вызова ParsePortRange,
// иначе тест не сможет замOCKить зависимость от сетевого стека. Они отражают
// конкретные сценарии запуска примеров из smoke- и closure-скриптов.

var (
	_legacyScanCIDR = ""
	_legacyScanPort = ""
)

func TestRunScan_RejectsInvalidCIDR_ExpectsError(t *testing.T) {
	_legacyScanCIDR = "not-a-cidr"
	if err := runScanWithLegacyFlags(t); err == nil {
		t.Fatal("ожидалась ошибка валидации CIDR, получено nil")
	}
}

func TestRunScan_RejectsInvalidPortRange_ExpectsError(t *testing.T) {
	_legacyScanPort = "80,not-a-port"
	if err := runScanWithLegacyFlags(t); err == nil {
		t.Fatal("ожидалась ошибка валидации портов, получено nil")
	}
}

func TestRunScan_AcceptsValidFlags_ExpectsNoError(t *testing.T) {
	const cidr = "192.168.1.0/24"
	const port = "80,443-8080"
	_legacyScanCIDR = cidr
	_legacyScanPort = port

	if err := runScanWithLegacyFlags(t); err != nil {
		t.Fatalf("валидные флаги не прошли валидацию: %v", err)
	}
}

// runScanWithLegacyFlags собирает флаги сканирования из глобальных переменных и
// вызывает RunScan напрямую (без os.Exit, как в ExecuteLegacyScan) — чтобы
// тест мог проверить возвращённую ошибку валидации.
func runScanWithLegacyFlags(t *testing.T) error {
	t.Helper()
	cfg := builder.Config{
		LogLevel: "info",
		DBPath:   filepath.Join(t.TempDir(), "inv.db"),
	}

	args := []string{
		"--network", _legacyScanCIDR,
		"--ports", _legacyScanPort,
		"--timeout", "2",
		"--threads", "50",
	}
	return RunScan(cfg, args...)
}
