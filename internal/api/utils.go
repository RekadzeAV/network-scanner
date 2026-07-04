package api

import (
	"fmt"
	"time"
)

// generateScanID генерирует уникальный ID для сканирования
func generateScanID() string {
	return fmt.Sprintf("scan-%d", time.Now().UnixNano())
}
