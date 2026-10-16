package cmd

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"network-scanner/internal/network"
)

// Диапазоны допустимых значений legacy-флагов scan (стр. CLI-путей).
// Заведомо «жадные» значения (timeout=0 → 0-таймаут на каждый probe,
// threads=100000 → исчерпание сокетов/файловых дескрипторов) отсекаются
// здесь, до попадания в сканер.
const (
	minScanTimeoutSec = 1
	maxScanTimeoutSec = 600
	minScanThreads    = 1
	maxScanThreads    = 1024
)

// parsePositiveInt строго разбирает положительное целое из CLI-аргумента.
// В отличие от fmt.Sscanf (который молча оставлял прежнее значение дефолта
// при неверном вводе), любая ошибка — явный результат вызывающему коду.
func parsePositiveInt(flag, value string) (int, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return 0, fmt.Errorf("%s: значение обязательно", flag)
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %q не является целым числом", flag, value)
	}
	if n < 1 {
		return 0, fmt.Errorf("%s: должно быть >= 1, получено %d", flag, n)
	}
	return n, nil
}

// validateScanTimeoutSeconds проверяет таймаут сканирования (в секундах).
func validateScanTimeoutSeconds(flag string, sec int) error {
	if sec < minScanTimeoutSec || sec > maxScanTimeoutSec {
		return fmt.Errorf("%s: %d вне диапазона %d..%d секунд",
			flag, sec, minScanTimeoutSec, maxScanTimeoutSec)
	}
	return nil
}

// validateScanThreads проверяет число потоков сканирования.
func validateScanThreads(flag string, threads int) error {
	if threads < minScanThreads || threads > maxScanThreads {
		return fmt.Errorf("%s: %d вне диапазона %d..%d",
			flag, threads, minScanThreads, maxScanThreads)
	}
	return nil
}

// validateScanCIDR проверяет, что --network является корректным CIDR.
func validateScanCIDR(cidr string) error {
	if _, _, err := net.ParseCIDR(strings.TrimSpace(cidr)); err != nil {
		return fmt.Errorf("--network: %q не является корректным CIDR: %w", cidr, err)
	}
	return nil
}

// validateScanPortRange проверяет TCP-диапазон портов через общий
// network.ParsePortRange (тот же парсер, что использует сканер).
func validateScanPortRange(portRange string) error {
	if _, err := network.ParsePortRange(portRange); err != nil {
		return fmt.Errorf("--ports: %w", err)
	}
	return nil
}
