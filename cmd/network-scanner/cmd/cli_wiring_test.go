package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ============================================================================
// W1: Тесты единого cobra-слоя диспатча CLI.
//
// Покрывают дедупликацию диспатча (E6): rootCmd регистрирует все команды, а
// inventory-подкоманды имеют рабочие флаги и валидацию аргументов.
// ============================================================================

// findCommand возвращает зарегистрированную подкоманду по имени.
func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, c := range root.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestRootCommand_RegistersAllSubcommands(t *testing.T) {
	want := []string{"scan", "inventory", "remote-exec", "device-control", "gui", "version"}
	for _, name := range want {
		if findCommand(rootCmd, name) == nil {
			t.Errorf("rootCmd не содержит подкоманду %q", name)
		}
	}
}

func TestInventoryCommand_RegistersSubcommands(t *testing.T) {
	for _, name := range []string{"list", "diff", "save"} {
		if findCommand(inventoryCmd, name) == nil {
			t.Errorf("inventoryCmd не содержит подкоманду %q", name)
		}
	}
}

func TestInventoryCommand_HasPersistentDBFlag(t *testing.T) {
	flag := inventoryCmd.PersistentFlags().Lookup("db")
	if flag == nil {
		t.Fatal("inventoryCmd не имеет persistent-флага --db")
	}
	if flag.DefValue != defaultInventoryDBPath() {
		t.Errorf("--db default = %q, want %q", flag.DefValue, defaultInventoryDBPath())
	}
}

func TestInventoryList_Flags(t *testing.T) {
	if inventoryListCmd.Flags().Lookup("limit") == nil {
		t.Error("inventory list не имеет флага --limit")
	}
}

func TestInventoryDiff_HasHistoryFlag(t *testing.T) {
	if inventoryDiffCmd.Flags().Lookup("history") == nil {
		t.Error("inventory diff не имеет флага --history")
	}
}

func TestInventorySave_Flags(t *testing.T) {
	want := []string{"id", "hosts-file", "network", "ports", "timeout", "threads"}
	for _, name := range want {
		if inventorySaveCmd.Flags().Lookup(name) == nil {
			t.Errorf("inventory save не имеет флага --%s", name)
		}
	}
}

// TestInventoryList_RejectsInvalidLimit — ветка: позиционный limit не число.
func TestInventoryList_RejectsInvalidLimit(t *testing.T) {
	err := inventoryListCmd.Args(inventoryListCmd, []string{"not-a-number"})
	if err != nil {
		t.Fatalf("Args() вернул ошибку для одного позиционного аргумента: %v", err)
	}

	// RunE должен вернуть ошибку парсинга.
	runErr := inventoryListCmd.RunE(inventoryListCmd, []string{"not-a-number"})
	if runErr == nil {
		t.Fatal("ожидалась ошибка парсинга limit")
	}
	if !strings.Contains(runErr.Error(), "некорректный limit") {
		t.Errorf("неожиданное сообщение об ошибке: %v", runErr)
	}
}

// TestInventoryList_RejectsExtraArgs — ветка: слишком много аргументов.
func TestInventoryList_RejectsExtraArgs(t *testing.T) {
	if err := inventoryListCmd.Args(inventoryListCmd, []string{"1", "2"}); err == nil {
		t.Error("ожидалась ошибка для >1 позиционного аргумента")
	}
}

// TestInventoryDiff_RequiresTwoArgs — ветка: неверное число аргументов.
func TestInventoryDiff_RequiresTwoArgs(t *testing.T) {
	if err := inventoryDiffCmd.Args(inventoryDiffCmd, []string{"only-one"}); err == nil {
		t.Error("ожидалась ошибка для одного аргумента")
	}
	if err := inventoryDiffCmd.Args(inventoryDiffCmd, []string{"a", "b"}); err != nil {
		t.Errorf("два аргумента должны приниматься: %v", err)
	}
}

// TestInventorySave_RequiresSource — ветка: нет ни --hosts-file, ни --network.
func TestInventorySave_RequiresSource(t *testing.T) {
	err := inventorySaveCmd.RunE(inventorySaveCmd, nil)
	if err == nil {
		t.Fatal("ожидалась ошибка отсутствия источника данных")
	}
	if !strings.Contains(err.Error(), "требуется источник данных") {
		t.Errorf("неожиданное сообщение об ошибке: %v", err)
	}
}

// TestInventorySave_RejectsPositionalArgs — ветка: save не принимает позиции.
func TestInventorySave_RejectsPositionalArgs(t *testing.T) {
	if err := inventorySaveCmd.Args(inventorySaveCmd, []string{"extra"}); err == nil {
		t.Error("ожидалась ошибка для позиционного аргумента")
	}
}

// TestInventoryConfig_UsesDefaultDBPath — ветка: --db пуст → значение по умолчанию.
func TestInventoryConfig_UsesDefaultDBPath(t *testing.T) {
	cfg := inventoryConfig(inventoryListCmd)
	if cfg.DBPath != defaultInventoryDBPath() {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, defaultInventoryDBPath())
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want \"info\"", cfg.LogLevel)
	}
}

// TestGUICommand_Registered — ветка: gui-команда доступна из rootCmd.
func TestGUICommand_Registered(t *testing.T) {
	gui := findCommand(rootCmd, "gui")
	if gui == nil {
		t.Fatal("gui-команда не зарегистрирована")
	}
	if gui.Run == nil {
		t.Error("gui-команда не имеет Run")
	}
}
