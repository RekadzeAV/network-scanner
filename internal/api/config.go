package api

import (
	"time"
)

// Config конфигурация API-сервера
type Config struct {
	Port               int
	Host               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	ShutdownTimeout    time.Duration
	EnableCORS         bool
	AllowedOrigins     []string
	RateLimitPerSecond int
	InventoryPath      string
	AuthToken          string // Bearer-токен для аутентификации API (пусто = отключить)
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		Port:               8080,
		Host:               "0.0.0.0",
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		ShutdownTimeout:    30 * time.Second,
		EnableCORS:         true,
		AllowedOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		RateLimitPerSecond: 10,
		InventoryPath:      "inventory.db",
	}
}
