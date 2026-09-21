package gui

import (
	"fmt"
	"runtime"
)

// --- Информация о версии для диалога «О программе» ---

// guiVersion — версия GUI-приложения (синхронизирована с cmd/gui).
const guiVersion = "1.0.3"

// BuildInfo возвращает строку с информацией о сборке (ОС/арх/Go).
func BuildInfo() string {
	return fmt.Sprintf("%s/%s, Go %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
}
