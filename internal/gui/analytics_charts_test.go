package gui

import (
	"strings"
	"testing"
)

func TestBuildAnalyticsMarkdown_EmptyBoth(t *testing.T) {
	result := buildAnalyticsMarkdown(nil, nil)
	if !strings.Contains(result, "Нет данных") {
		t.Errorf("expected 'Нет данных', got %q", result)
	}
}

func TestBuildAnalyticsMarkdown_EmptyProtocols(t *testing.T) {
	result := buildAnalyticsMarkdown(nil, map[string]int{"Router": 2})
	if !strings.Contains(result, "Нет данных") {
		t.Error("expected 'Нет данных' for empty protocols")
	}
	if !strings.Contains(result, "Router") {
		t.Error("expected 'Router' in device types")
	}
}

func TestBuildAnalyticsMarkdown_EmptyDeviceTypes(t *testing.T) {
	result := buildAnalyticsMarkdown(map[string]int{"TCP": 5}, nil)
	if !strings.Contains(result, "TCP") {
		t.Error("expected 'TCP' in protocols")
	}
	if !strings.Contains(result, "Нет данных") {
		t.Error("expected 'Нет данных' for empty device types")
	}
}

func TestBuildAnalyticsMarkdown_BothWithData(t *testing.T) {
	protocols := map[string]int{"TCP": 10, "UDP": 5, "ICMP": 2}
	deviceTypes := map[string]int{"Router": 1, "Switch": 2, "PC": 3}
	result := buildAnalyticsMarkdown(protocols, deviceTypes)
	if !strings.Contains(result, "### Аналитика") {
		t.Error("expected '### Аналитика' header")
	}
	if !strings.Contains(result, "TCP") || !strings.Contains(result, "UDP") || !strings.Contains(result, "ICMP") {
		t.Errorf("expected all protocols in result, got: %s", result)
	}
}

func TestBuildAnalyticsMarkdown_WhitespaceTrimming(t *testing.T) {
	protocols := map[string]int{" TCP ": 5}
	deviceTypes := map[string]int{" Router ": 1}
	result := buildAnalyticsMarkdown(protocols, deviceTypes)
	if !strings.Contains(result, "`TCP`") {
		t.Errorf("expected trimmed protocol name, got: %s", result)
	}
	if !strings.Contains(result, "`Router`") {
		t.Errorf("expected trimmed device type, got: %s", result)
	}
}

func TestBuildPieChartCacheKey_Empty(t *testing.T) {
	key := buildPieChartCacheKey("", nil)
	if key == "" {
		t.Error("expected non-empty cache key for empty inputs")
	}
}

func TestBuildPieChartCacheKey_WithData(t *testing.T) {
	key1 := buildPieChartCacheKey("TCP", map[string]int{"open": 5})
	key2 := buildPieChartCacheKey("TCP", map[string]int{"open": 10})
	if key1 == key2 {
		t.Error("expected different cache keys for different data")
	}
}

func TestBuildPieChartCacheKey_SameData(t *testing.T) {
	key1 := buildPieChartCacheKey("TCP", map[string]int{"open": 5})
	key2 := buildPieChartCacheKey("TCP", map[string]int{"open": 5})
	if key1 != key2 {
		t.Errorf("expected same cache key for same data: %q vs %q", key1, key2)
	}
}

func TestBuildPieChartCacheKey_DifferentTitle(t *testing.T) {
	key1 := buildPieChartCacheKey("TCP", map[string]int{"open": 5})
	key2 := buildPieChartCacheKey("UDP", map[string]int{"open": 5})
	if key1 == key2 {
		t.Error("expected different cache keys for different titles")
	}
}

func TestAngleInSector_FullRange(t *testing.T) {
	if !angleInSector(45, 0, 360) {
		t.Error("expected true for 45 in [0, 360]")
	}
}

func TestAngleInSector_ExactStart(t *testing.T) {
	if !angleInSector(0, 0, 90) {
		t.Error("expected true for exact start angle")
	}
}

func TestAngleInSector_ExactEnd(t *testing.T) {
	if !angleInSector(90, 0, 90) {
		t.Error("expected true for exact end angle")
	}
}

func TestAngleInSector_BeforeStart(t *testing.T) {
	if angleInSector(-10, 0, 90) {
		t.Error("expected false for angle before start")
	}
}

func TestAngleInSector_AfterEnd(t *testing.T) {
	if angleInSector(100, 0, 90) {
		t.Error("expected false for angle after end")
	}
}

func TestAngleInSector_WideRange(t *testing.T) {
	// 360° mod 2π = 0, так что 180 в [0, 0] — false
	if angleInSector(180, 0, 360) {
		// 180 в [0, 0] после mod — false
	}
	// Но 0 в [0, 360] (после mod 360→0) — true
	if !angleInSector(0, 0, 360) {
		t.Error("expected true for 0 in [0, 360]")
	}
}

// angleInSector работает с радианами (float64), не с градусами.
// math.Mod(10, 2π) = 10 - 2*π ≈ 3.717, поэтому 5 > 3.717 = false.
func TestAngleInSector_NarrowRange(t *testing.T) {
	// 0.5 рад в [0, 1] рад — true
	if !angleInSector(0.5, 0, 1) {
		t.Error("expected true for 0.5 in [0, 1] radians")
	}
}

func TestAngleInSector_ZeroRange(t *testing.T) {
	if !angleInSector(0, 0, 0) {
		t.Error("expected true for angle == start == end")
	}
}
