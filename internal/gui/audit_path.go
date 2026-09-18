package gui

import (
	"os"
	"path/filepath"
	"strings"
)

// auditLogDirName — подкаталог в пользовательской конфигурации для журналов
// чувствительных действий (device control).
const auditLogDirName = "network-scanner"

// deviceAuditLogPath возвращает абсолютный путь к журналу действий с
// устройствами. Журнал пишется в пользовательский конфигурационный каталог
// (os.UserConfigDir), а не в текущий рабочий каталог: это исключает попадание
// лога в репозиторий/бэкапы и делает местоположение предсказуемым.
//
// Если системный конфигурационный каталог недоступен, используется временный
// каталог как безопасный fallback (никогда не CWD).
func deviceAuditLogPath() string {
	base := ""
	if dir, err := os.UserConfigDir(); err == nil {
		base = strings.TrimSpace(dir)
	}
	if base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, auditLogDirName, "device-actions.log")
}
