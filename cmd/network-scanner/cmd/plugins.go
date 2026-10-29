// Package cmd CLI network-scanner.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"network-scanner/internal/logger"
	"network-scanner/internal/plugin"
)

// pluginsCmd — группа команд управления плагинами.
var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Управление плагинами",
	Long:  "Команды для работы с плагинами расширения Network Scanner.",
}

// pluginsListCmd печатает каталоги поиска и состояние плагинов.
var pluginsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Список доступных плагинов",
	Long: `Показывает каталоги поиска плагинов, обнаруженные плагины,
их состояние (загружен / выключен / ошибка) и причины ошибок загрузки.

Каталоги: переменная окружения NETWORK_SCANNER_PLUGINS_DIR
(список через разделитель ОС), ./plugins и пользовательский конфигурационный каталог.`,
	// Ошибка загрузки плагина — не ошибка синтаксиса команды, usage не нужен.
	SilenceUsage: true,
	RunE:         runPluginsList,
}

func runPluginsList(_ *cobra.Command, _ []string) error {
	// Логи плагинного загрузчика идут в stderr, таблица — в stdout.
	_ = logger.Init("network-scanner", "dev")

	loader := plugin.NewPluginLoaderFromEnv()
	dirs := loader.Dirs()
	if _, err := loader.DiscoverAndLoad(); err != nil {
		// Ошибки загрузки (отсутствие библиотек, битые манифесты) — выводятся
		// в таблицу и в блок «Ошибки загрузки» ниже.
		logger.LogError(err, "plugin discover")
	}

	fmt.Println("Каталоги поиска плагинов:")
	if len(dirs) == 0 {
		fmt.Println("  (не задано ни одного каталога)")
	}
	for _, d := range dirs {
		if _, err := os.Stat(d); err != nil {
			fmt.Printf("  %s (не существует)\n", d)
			continue
		}
		fmt.Printf("  %s\n", d)
	}

	statuses := loader.Statuses()
	loadedCount := 0
	for _, s := range statuses {
		if s.Loaded {
			loadedCount++
		}
	}

	fmt.Println("\nПлагины:")
	if len(statuses) == 0 && loadedCount == 0 {
		fmt.Println("  (плагины не найдены)")
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	if err := writePluginTable(w, statuses); err != nil {
		return fmt.Errorf("не удалось вывести таблицу плагинов: %w", err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("не удалось вывести таблицу плагинов: %w", err)
	}

	if errs := loader.LoadErrors(); len(errs) > 0 {
		fmt.Println("\nОшибки загрузки:")
		for _, e := range errs {
			fmt.Printf("  %v\n", e)
		}
		return fmt.Errorf("%d плагинов не удалось загрузить", len(errs))
	}
	return nil
}

// writePluginTable печатает строки состояния обнаруженных плагинов.
func writePluginTable(w *tabwriter.Writer, statuses []plugin.PluginStatus) error {
	if _, err := fmt.Fprintln(w, "ИМЯ\tТИП\tВЕРСИЯ\tСТАТУС"); err != nil {
		return err
	}
	for _, s := range statuses {
		status := "выключен"
		switch {
		case s.Loaded:
			status = "загружен"
		case s.Err != nil:
			status = "ошибка загрузки"
		}
		name := s.Manifest.Name
		if name == "" {
			name = filepath.Base(s.Dir)
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, s.Manifest.Type, s.Manifest.Version, status); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	pluginsCmd.AddCommand(pluginsListCmd)
	rootCmd.AddCommand(pluginsCmd)
}
