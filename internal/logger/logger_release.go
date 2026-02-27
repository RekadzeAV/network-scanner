//go:build !debug
// +build !debug

package logger

// Релизная версия - логирование отключено

// Init инициализирует систему логирования (заглушка для релизной версии)
func Init(appName, version string) error {
	return nil
}

// Close закрывает файл лога (заглушка для релизной версии)
func Close() {
}

// Log записывает сообщение в лог (заглушка для релизной версии)
func Log(format string, args ...interface{}) {
}

// LogError записывает ошибку в лог (заглушка для релизной версии)
func LogError(err error, context string) {
}

// LogDebug записывает отладочное сообщение (заглушка для релизной версии)
func LogDebug(format string, args ...interface{}) {
}

// GetLogFileName возвращает имя файла лога (заглушка для релизной версии)
func GetLogFileName() string {
	return ""
}
