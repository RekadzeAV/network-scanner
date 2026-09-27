package gui

import (
	"path/filepath"

	"network-scanner/internal/auditpath"
)

// auditLogDirName — подкаталог в пользовательской конфигурации для журналов
// чувствительных действий (device control). Оставлен для тестов и совместимости.
const auditLogDirName = auditpath.DirName

// deviceAuditLogPath возвращает абсолютный путь к журналу действий с
// устройствами. Журнал пишется в пользовательский конфигурационный каталог
// (os.UserConfigDir), а не в текущий рабочий каталог: это исключает попадание
// лога в репозиторий/бэкапы и делает местоположение предсказуемым.
//
// Если системный конфигурационный каталог недоступен, используется временный
// каталог как безопасный fallback (никогда не CWD).
//
// Реализация делегирована в internal/auditpath — единый источник путей журналов
// для GUI и CLI (E7/7.9).
func deviceAuditLogPath() string {
	return filepath.Clean(auditpath.DeviceActionsPath())
}
