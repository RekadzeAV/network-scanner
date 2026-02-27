//go:build debug
// +build debug

package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	AppName    = "network-scanner"
	AppVersion = "1.0.3" // Версия тестовой сборки (версия 3)
)

var (
	logFile   *os.File
	logMutex  sync.Mutex
	initialized bool
)

// Init инициализирует систему логирования
func Init(appName, version string) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if initialized {
		return nil
	}

	// Получаем рабочую директорию (каталог запуска приложения)
	workDir, err := os.Getwd()
	if err != nil {
		workDir = "неизвестно"
	}
	
	// Формируем имя файла лога: LOG-название приложения-версия релиза.txt
	logFileName := fmt.Sprintf("LOG-%s-%s.txt", appName, version)
	
	// Получаем полный путь к файлу лога
	logFilePath := filepath.Join(workDir, logFileName)
	
	// Открываем файл для записи (создаем если не существует, добавляем если существует)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл лога: %v", err)
	}

	logFile = file
	initialized = true

	// Записываем заголовок сессии с максимальной детализацией
	writeLog("==========================================")
	writeLog(fmt.Sprintf("Лог сессии: %s", time.Now().Format("2006-01-02 15:04:05.000")))
	writeLog(fmt.Sprintf("Приложение: %s v%s", appName, version))
	writeLog(fmt.Sprintf("ОС: %s %s", runtime.GOOS, runtime.GOARCH))
	writeLog(fmt.Sprintf("Рабочая директория: %s", workDir))
	writeLog(fmt.Sprintf("Файл лога: %s", logFilePath))
	writeLog(fmt.Sprintf("Количество процессоров: %d", runtime.NumCPU()))
	writeLog(fmt.Sprintf("Версия Go: %s", runtime.Version()))
	writeLog("==========================================")

	return nil
}

// Close закрывает файл лога
func Close() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		writeLog("==========================================")
		writeLog(fmt.Sprintf("Сессия завершена: %s", time.Now().Format("2006-01-02 15:04:05")))
		writeLog("==========================================")
		logFile.Close()
		logFile = nil
		initialized = false
	}
}

// writeLog записывает строку в лог
func writeLog(message string) {
	if logFile == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)
	
	// Записываем в UTF-8
	logFile.WriteString(logLine)
	logFile.Sync() // Синхронизируем для немедленной записи
}

// Log записывает сообщение в лог
func Log(format string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if !initialized {
		return
	}

	message := fmt.Sprintf(format, args...)
	writeLog(message)
}

// LogError записывает ошибку в лог
func LogError(err error, context string) {
	if err == nil {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	if !initialized {
		return
	}

	message := fmt.Sprintf("ERROR [%s]: %v", context, err)
	writeLog(message)
}

// LogDebug записывает отладочное сообщение
func LogDebug(format string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if !initialized {
		return
	}

	message := fmt.Sprintf("DEBUG: %s", fmt.Sprintf(format, args...))
	writeLog(message)
}

// GetLogFileName возвращает имя файла лога
func GetLogFileName() string {
	return fmt.Sprintf("LOG-%s-%s.txt", AppName, AppVersion)
}
