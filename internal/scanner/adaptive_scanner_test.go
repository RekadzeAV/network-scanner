package scanner

import (
	"testing"
	"time"
)

// --- Test AdaptiveConfig ---

func TestDefaultAdaptiveConfig(t *testing.T) {
	config := DefaultAdaptiveConfig()

	if config.MinBudget != 64 {
		t.Errorf("MinBudget = %d, want 64", config.MinBudget)
	}

	if config.MaxBudget != 1024 {
		t.Errorf("MaxBudget = %d, want 1024", config.MaxBudget)
	}

	if config.InitialBudget != 512 {
		t.Errorf("InitialBudget = %d, want 512", config.InitialBudget)
	}

	if config.SpeedThreshold != 50*time.Millisecond {
		t.Errorf("SpeedThreshold = %v, want 50ms", config.SpeedThreshold)
	}

	if config.ErrorThreshold != 0.3 {
		t.Errorf("ErrorThreshold = %v, want 0.3", config.ErrorThreshold)
	}

	if config.AdaptInterval != 1*time.Second {
		t.Errorf("AdaptInterval = %v, want 1s", config.AdaptInterval)
	}
}

// --- Test AdaptiveScanner ---

func TestNewAdaptiveScanner(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	if scanner == nil {
		t.Fatal("NewAdaptiveScanner() returned nil")
	}

	if scanner.scanner == nil {
		t.Error("AdaptiveScanner.scanner should not be nil")
	}

	if scanner.metrics == nil {
		t.Error("AdaptiveScanner.metrics should not be nil")
	}

	if scanner.GetBudget() != config.InitialBudget {
		t.Errorf("Initial budget = %d, want %d", scanner.GetBudget(), config.InitialBudget)
	}
}

func TestNewAdaptiveScannerWithNilConfig(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)

	scanner := NewAdaptiveScanner(ns, AdaptiveConfig{})

	if scanner == nil {
		t.Fatal("NewAdaptiveScanner() with nil config returned nil")
	}
}

// --- Test GetBudget / SetBudget ---

func TestGetBudget(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	budget := scanner.GetBudget()
	if budget != config.InitialBudget {
		t.Errorf("GetBudget() = %d, want %d", budget, config.InitialBudget)
	}
}

func TestSetBudgetWithinRange(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	// Устанавливаем budget в диапазоне
	scanner.SetBudget(256)
	budget := scanner.GetBudget()
	if budget != 256 {
		t.Errorf("SetBudget(256) = %d, want 256", budget)
	}
}

func TestSetBudgetBelowMin(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	// Устанавливаем budget ниже минимума
	scanner.SetBudget(10)
	budget := scanner.GetBudget()
	if budget != config.MinBudget {
		t.Errorf("SetBudget(10) = %d, want %d (MinBudget)", budget, config.MinBudget)
	}
}

func TestSetBudgetAboveMax(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	// Устанавливаем budget выше максимума
	scanner.SetBudget(2000)
	budget := scanner.GetBudget()
	if budget != config.MaxBudget {
		t.Errorf("SetBudget(2000) = %d, want %d (MaxBudget)", budget, config.MaxBudget)
	}
}

// --- Test RecordProbe ---

func TestRecordProbeOpen(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	scanner.RecordProbe(true, false)

	metrics := scanner.GetMetrics()
	if metrics.probesTotal != 1 {
		t.Errorf("probesTotal = %d, want 1", metrics.probesTotal)
	}

	if metrics.probesOpen != 1 {
		t.Errorf("probesOpen = %d, want 1", metrics.probesOpen)
	}
}

func TestRecordProbeClosed(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	scanner.RecordProbe(false, false)

	metrics := scanner.GetMetrics()
	if metrics.probesTotal != 1 {
		t.Errorf("probesTotal = %d, want 1", metrics.probesTotal)
	}

	if metrics.probesClosed != 1 {
		t.Errorf("probesClosed = %d, want 1", metrics.probesClosed)
	}
}

func TestRecordProbeError(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	scanner.RecordProbe(false, true)

	metrics := scanner.GetMetrics()
	if metrics.probesTotal != 1 {
		t.Errorf("probesTotal = %d, want 1", metrics.probesTotal)
	}

	if metrics.probesError != 1 {
		t.Errorf("probesError = %d, want 1", metrics.probesError)
	}
}

func TestRecordProbeMultiple(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	// Записываем несколько probe
	for i := 0; i < 10; i++ {
		scanner.RecordProbe(i%2 == 0, false)
	}

	metrics := scanner.GetMetrics()
	if metrics.probesTotal != 10 {
		t.Errorf("probesTotal = %d, want 10", metrics.probesTotal)
	}

	if metrics.probesOpen != 5 {
		t.Errorf("probesOpen = %d, want 5", metrics.probesOpen)
	}

	if metrics.probesClosed != 5 {
		t.Errorf("probesClosed = %d, want 5", metrics.probesClosed)
	}
}

// --- Test GetOpenRate / GetErrorRate ---

func TestGetOpenRate(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	// Записываем probe: 5 open, 5 closed
	for i := 0; i < 10; i++ {
		scanner.RecordProbe(i < 5, false)
	}

	openRate := scanner.GetOpenRate()
	if openRate != 0.5 {
		t.Errorf("GetOpenRate() = %v, want 0.5", openRate)
	}
}

func TestGetOpenRateEmpty(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	openRate := scanner.GetOpenRate()
	if openRate != 0 {
		t.Errorf("GetOpenRate() = %v, want 0 (empty)", openRate)
	}
}

func TestGetErrorRate(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	// Записываем probe: 3 error, 7 normal
	for i := 0; i < 10; i++ {
		scanner.RecordProbe(false, i < 3)
	}

	errorRate := scanner.GetErrorRate()
	if errorRate != 0.3 {
		t.Errorf("GetErrorRate() = %v, want 0.3", errorRate)
	}
}

func TestGetErrorRateEmpty(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	errorRate := scanner.GetErrorRate()
	if errorRate != 0 {
		t.Errorf("GetErrorRate() = %v, want 0 (empty)", errorRate)
	}
}

// --- Test Adapt ---

func TestAdaptNoAdaptTooFewProbes(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := AdaptiveConfig{
		MinBudget:      64,
		MaxBudget:      1024,
		InitialBudget:  512,
		AdaptInterval:  0, // Отключаем интервал для теста
	}

	scanner := NewAdaptiveScanner(ns, config)

	// Записываем меньше 10 probe
	for i := 0; i < 5; i++ {
		scanner.RecordProbe(false, false)
	}

	initialBudget := scanner.GetBudget()
	scanner.Adapt()

	finalBudget := scanner.GetBudget()
	if initialBudget != finalBudget {
		t.Errorf("Budget changed from %d to %d, should not adapt with < 10 probes", initialBudget, finalBudget)
	}
}

func TestAdaptHighErrorRate(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := AdaptiveConfig{
		MinBudget:      64,
		MaxBudget:      1024,
		InitialBudget:  512,
		ErrorThreshold: 0.3,
		AdaptInterval:  0,
	}

	scanner := NewAdaptiveScanner(ns, config)

	// Записываем probe с высоким процентом ошибок (>30%)
	for i := 0; i < 20; i++ {
		scanner.RecordProbe(false, i < 8) // 8/20 = 40% ошибок
	}

	scanner.Adapt()

	finalBudget := scanner.GetBudget()
	if finalBudget >= 512 {
		t.Errorf("Budget should decrease with high error rate, got %d", finalBudget)
	}
}

func TestAdaptLowOpenRate(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := AdaptiveConfig{
		MinBudget:      64,
		MaxBudget:      1024,
		InitialBudget:  512,
		AdaptInterval:  0,
	}

	scanner := NewAdaptiveScanner(ns, config)

	// Записываем probe с низким процентом открытых портов (<10%) и >100 probe
	for i := 0; i < 110; i++ {
		scanner.RecordProbe(i < 5, false) // 5/110 ≈ 4.5% open
	}

	scanner.Adapt()

	finalBudget := scanner.GetBudget()
	if finalBudget <= 512 {
		t.Errorf("Budget should increase with low open rate, got %d", finalBudget)
	}
}

// --- Test GetDuration ---

func TestGetDuration(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	time.Sleep(50 * time.Millisecond)

	duration := scanner.GetDuration()
	if duration < 50*time.Millisecond {
		t.Errorf("GetDuration() = %v, want at least 50ms", duration)
	}
}

// --- Test GetSummary ---

func TestGetSummary(t *testing.T) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	// Записываем несколько probe
	for i := 0; i < 10; i++ {
		scanner.RecordProbe(i < 5, false)
	}

	summary := scanner.GetSummary()
	if summary == "" {
		t.Error("GetSummary() should not return empty string")
	}

	// Проверяем, что summary содержит ключевые слова
	if !containsString(summary, "Budget") {
		t.Error("GetSummary() should contain 'Budget'")
	}

	if !containsString(summary, "Probe всего") {
		t.Error("GetSummary() should contain 'Probe всего'")
	}
}

// --- Helper functions ---

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- Benchmark ---

func BenchmarkAdaptiveScannerGetBudget(b *testing.B) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scanner.GetBudget()
	}
}

func BenchmarkAdaptiveScannerSetBudget(b *testing.B) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.SetBudget(256)
	}
}

func BenchmarkAdaptiveScannerRecordProbe(b *testing.B) {
	ns := NewNetworkScanner("127.0.0.1/32", 200*time.Millisecond, "1-5", 5, false)
	config := DefaultAdaptiveConfig()

	scanner := NewAdaptiveScanner(ns, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.RecordProbe(i%2 == 0, false)
	}
}
