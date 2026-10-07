package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"network-scanner/internal/builder"
)

// ============================================================================
// E7/7.3: Планировщик периодических сканов (schedule).
//
// Запускает сканирование с заданным интервалом, пока процесс не получен
// SIGINT/SIGTERM. Реализация встроенного daemon-планировщика из плана
// (docs/UNIFIED_OPTIMIZED_PLAN_2026-09-15.md, 7.3) — периодические сканы
// без внешнего cron/systemd-timer, для Windows и простых сценариев.
//
// Сканирование переиспользует существующий cobra-dispatch (`RunScanCobra`):
// schedule наследует все флаги scan (напрямую от scanCmd) и вызывает
// RunScanCobra с собственной командой, чтобы не дублировать конфигурацию.
// Единственное отличие — экспорты по умолчанию выключены, а ошибки
// отдельного тика не останавливают цикл (лог + продолжение).
// ============================================================================

// scheduleCmd — команда `network-scanner schedule`.
var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Периодическое сканирование по интервалу (планировщик)",
	Long: `Запускает сканирование циклически с указанным интервалом.

Примеры:
  network-scanner schedule --interval 6h --network 192.168.1.0/24 --ports 1-1000
  network-scanner schedule --interval 30m --hosts-file targets.txt --export-html

Интервал задаётся duration-строкой (Go-синтаксис): 30m, 1h30m, 6h.
Процесс останавливается по Ctrl+C (SIGINT) или SIGTERM; текущий тик
дожидается завершения, чтобы не прерывать сканирование.`,
	Args: cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		return runSchedule(c)
	},
}

func init() {
	// Наследуем все флаги scan: планировщик должен принимать те же параметры
	// сканирования (E7/7.3: без дублирования объявлений флагов).
	scanCmd.Flags().VisitAll(func(f *pflag.Flag) {
		scheduleCmd.Flags().AddFlag(f)
	})

	scheduleCmd.Flags().String("interval", "6h", "Интервал между сканированиями (Go duration, напр. 30m, 1h30m, 6h)")
	scheduleCmd.Flags().Int("max-runs", 0, "Максимальное число запусков (0 — без ограничения)")
	scheduleCmd.Flags().Bool("skip-first", false, "Не выполнять первый скан сразу, а ждать интервал")
}

// parseInterval парсит duration-строку интервала. Возвращает ошибку для
// пустой строки, нуля и отрицательного значения — планировщик без валидного
// интервала завис бы или крутился вхолостую.
func parseInterval(s string) (time.Duration, error) {
	if strings.TrimSpace(s) == "" {
		return 0, fmt.Errorf("интервал не задан")
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("некорректный интервал %q: %w", s, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("интервал должен быть положительным, получено %q", s)
	}
	return d, nil
}

// runSchedule выполняет цикл планировщика: первый (опционально) запуск
// сканирования, затем тик каждые interval до max-runs или сигнала остановки.
func runSchedule(c *cobra.Command) error {
	intervalStr, _ := c.Flags().GetString("interval")
	interval, err := parseInterval(intervalStr)
	if err != nil {
		return err
	}
	maxRuns, _ := c.Flags().GetInt("max-runs")
	if maxRuns < 0 {
		return fmt.Errorf("max-runs не может быть отрицательным: %d", maxRuns)
	}
	skipFirst, _ := c.Flags().GetBool("skip-first")

	// Обработчик сигналов: SIGINT/SIGTERM → graceful stop после текущего тика.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	runOne := func(runIdx int) error {
		fmt.Printf("\n[schedule] тик #%d: %s\n", runIdx, time.Now().Format(time.RFC3339))
		// RunScanCobra читает флаги из c: у schedule они те же, что у scan.
		// cfg — как в scanCommandRun (DBPath по умолчанию, LogLevel info).
		return RunScanCobra(c, builder.Config{
			LogLevel: "info",
			DBPath:   filepath.Join("inventory", "network_inventory.db"),
		})
	}

	runIdx := 0
	if !skipFirst {
		runIdx++
		if err := runOne(runIdx); err != nil {
			// Ошибка тика не останавливает планировщик — логируем и продолжаем.
			fmt.Fprintf(os.Stderr, "[schedule] ошибка тика #%d: %v\n", runIdx, err)
		}
		if maxRuns > 0 && runIdx >= maxRuns {
			fmt.Printf("[schedule] достигнут max-runs=%d, выход\n", maxRuns)
			return nil
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			fmt.Println("\n[schedule] получен сигнал остановки, выход")
			return nil
		case <-ticker.C:
			runIdx++
			if err := runOne(runIdx); err != nil {
				fmt.Fprintf(os.Stderr, "[schedule] ошибка тика #%d: %v\n", runIdx, err)
			}
			if maxRuns > 0 && runIdx >= maxRuns {
				fmt.Printf("[schedule] достигнут max-runs=%d, выход\n", maxRuns)
				return nil
			}
		}
	}
}