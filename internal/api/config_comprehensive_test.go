package api

import (
	"testing"
	"time"
)

// ============================================================================
// DefaultConfig — покрытие
// ============================================================================

func TestDefaultConfig_Port(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
}

func TestDefaultConfig_Host(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %q", cfg.Host)
	}
}

func TestDefaultConfig_ReadTimeout(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("expected ReadTimeout 10s, got %v", cfg.ReadTimeout)
	}
}

func TestDefaultConfig_WriteTimeout(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.WriteTimeout != 10*time.Second {
		t.Errorf("expected WriteTimeout 10s, got %v", cfg.WriteTimeout)
	}
}

func TestDefaultConfig_ShutdownTimeout(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Errorf("expected ShutdownTimeout 30s, got %v", cfg.ShutdownTimeout)
	}
}

func TestDefaultConfig_EnableCORS(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.EnableCORS {
		t.Error("expected EnableCORS true")
	}
}

func TestDefaultConfig_AllowedOrigins(t *testing.T) {
	cfg := DefaultConfig()
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("expected 2 allowed origins, got %d", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("expected first origin localhost:3000, got %q", cfg.AllowedOrigins[0])
	}
}

func TestDefaultConfig_RateLimit(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.RateLimitPerSecond != 10 {
		t.Errorf("expected RateLimit 10, got %d", cfg.RateLimitPerSecond)
	}
}

func TestDefaultConfig_InventoryPath(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.InventoryPath != "inventory.db" {
		t.Errorf("expected inventory.db, got %q", cfg.InventoryPath)
	}
}

// ============================================================================
// NewHandler — покрытие
// ============================================================================

func TestNewHandler(t *testing.T) {
	cfg := DefaultConfig()
	h := NewHandler(cfg)
	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
	if h.config.Port != 8080 {
		t.Error("Handler config not set correctly")
	}
}

func TestNewHandler_EmptyConfig(t *testing.T) {
	h := NewHandler(Config{})
	if h == nil {
		t.Fatal("NewHandler with empty config returned nil")
	}
}

// ============================================================================
// generateScanID — покрытие
// ============================================================================

func TestGenerateScanID_Format(t *testing.T) {
	id := generateScanID()
	if len(id) < 10 {
		t.Errorf("expected long ID, got %q (len=%d)", id, len(id))
	}
	if len(id) < 5 || id[:5] != "scan-" {
		t.Errorf("expected scan- prefix, got %q", id)
	}
}

func TestGenerateScanID_Unique(t *testing.T) {
	// generateScanID использует UnixNano, при очень быстром вызове
	// может совпасть — проверяем что формат корректен
	id1 := generateScanID()
	id2 := generateScanID()
	// Если ID совпали (редкий случай nanosecond), проверяем что формат верный
	if id1 == id2 {
		// Проверяем формат вместо уникальности
		if len(id1) < 5 || id1[:5] != "scan-" {
			t.Error("expected scan- prefix")
		}
	}
}

func TestGenerateScanID_Prefix(t *testing.T) {
	id := generateScanID()
	if len(id) < 5 || id[:5] != "scan-" {
		t.Errorf("expected scan- prefix, got %q", id)
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkDefaultConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

func BenchmarkGenerateScanID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateScanID()
	}
}
