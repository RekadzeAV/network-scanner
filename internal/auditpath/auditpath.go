// Package auditpath централизованно определяет расположение журналов
// чувствительных действий (device-control, remote-exec и др.).
//
// Инвариант безопасности: журналы пишутся в пользовательский конфигурационный
// каталог (os.UserConfigDir), а НЕ в текущий рабочий каталог — это исключает
// попадание журнала в репозиторий, бэкапы и diff'ы, а также делает
// местоположение предсказуемым для аудита. Если системный конфигурационный
// каталог недоступен, используется временный каталог как безопасный fallback
// (тоже никогда не CWD).
package auditpath

import (
	"os"
	"path/filepath"
	"strings"
)

// DirName — подкаталог в пользовательской конфигурации для журналов проекта.
const DirName = "network-scanner"

// Имена журналов по видам операций.
const (
	DeviceActionsFile = "device-actions.log"
	RemoteExecFile    = "remote-exec.log"
)

// BaseDir возвращает каталог для журналов: <UserConfigDir>/network-scanner,
// с fallback в <TempDir>/network-scanner.
func BaseDir() string {
	base := ""
	if dir, err := os.UserConfigDir(); err == nil {
		base = strings.TrimSpace(dir)
	}
	if base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, DirName)
}

// DefaultPath возвращает абсолютный путь журнала для указанного имени файла.
// Пустое имя заменяется на общий журнал действий (actions.log).
func DefaultPath(fileName string) string {
	name := strings.TrimSpace(fileName)
	if name == "" {
		name = "actions.log"
	}
	return filepath.Join(BaseDir(), name)
}

// DeviceActionsPath — журнал действий с устройствами (device-control).
func DeviceActionsPath() string { return DefaultPath(DeviceActionsFile) }

// RemoteExecPath — журнал удалённого выполнения команд (remote-exec).
func RemoteExecPath() string { return DefaultPath(RemoteExecFile) }
