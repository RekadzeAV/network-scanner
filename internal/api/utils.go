package api

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// scanIDMu защищает генерацию ID от параллельного доступа
var scanIDMu sync.Mutex

// scanIDCounter используется для уникальности ID
var scanIDCounter uint64

// generateScanID генерирует уникальный ID для сканирования
func generateScanID() string {
	scanIDMu.Lock()
	defer scanIDMu.Unlock()

	// Используем counter для уникальности
	counter := atomic.AddUint64(&scanIDCounter, 1)
	return fmt.Sprintf("scan-%d-%d", time.Now().UnixNano(), counter)
}

// resetScanStore сбрасывает состояние scanStore (только для тестов)
func resetScanStore() {
	scanStoreInstance.globalMu.Lock()
	defer scanStoreInstance.globalMu.Unlock()

	scanStoreInstance.resetMu.Lock()
	defer scanStoreInstance.resetMu.Unlock()

	scanStoreInstance.mu.Lock()
	defer scanStoreInstance.mu.Unlock()

	scanStoreInstance.scans = make(map[string]*scanState)
}
