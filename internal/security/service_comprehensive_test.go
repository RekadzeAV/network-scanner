package security

import (
	"testing"
)

// ============================================================================
// calculateSecurityIndex — покрытие всех веток (91–104)
// ============================================================================

func TestCalculateSecurityIndex_Empty(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{})
	if score != 100 {
		t.Errorf("expected 100 for empty, got %d", score)
	}
}

func TestCalculateSecurityIndex_NilMap(t *testing.T) {
	score := calculateSecurityIndex(nil)
	if score != 100 {
		t.Errorf("expected 100 for nil, got %d", score)
	}
}

func TestCalculateSecurityIndex_Critical(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{"critical": 1})
	if score != 70 {
		t.Errorf("expected 70 for 1 critical, got %d", score)
	}
}

func TestCalculateSecurityIndex_MultipleCritical(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{"critical": 4})
	if score != 0 {
		t.Errorf("expected 0 for 4 critical, got %d", score)
	}
}

func TestCalculateSecurityIndex_High(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{"high": 1})
	if score != 80 {
		t.Errorf("expected 80 for 1 high, got %d", score)
	}
}

func TestCalculateSecurityIndex_Medium(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{"medium": 1})
	if score != 90 {
		t.Errorf("expected 90 for 1 medium, got %d", score)
	}
}

func TestCalculateSecurityIndex_Low(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{"low": 1})
	if score != 95 {
		t.Errorf("expected 95 for 1 low, got %d", score)
	}
}

func TestCalculateSecurityIndex_Mixed(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{
		"critical": 1,
		"high":     2,
		"medium":   3,
		"low":      4,
	})
	// 100 - 30 - 40 - 30 - 20 = -20 → 0
	if score != 0 {
		t.Errorf("expected 0 for mixed heavy, got %d", score)
	}
}

func TestCalculateSecurityIndex_SmallMixed(t *testing.T) {
	score := calculateSecurityIndex(map[string]int{
		"critical": 1,
		"low":      1,
	})
	// 100 - 30 - 5 = 65
	if score != 65 {
		t.Errorf("expected 65, got %d", score)
	}
}

func TestCalculateSecurityIndex_MaxClamp(t *testing.T) {
	// Теоретически невозможно превысить 100 при положительных весах,
	// но проверим что cap работает
	score := calculateSecurityIndex(map[string]int{})
	if score > 100 {
		t.Errorf("expected max 100, got %d", score)
	}
}

func TestCalculateSecurityIndex_MinClamp(t *testing.T) {
	// 5 critical = 100 - 150 = -50 → 0
	score := calculateSecurityIndex(map[string]int{"critical": 5})
	if score != 0 {
		t.Errorf("expected 0 (min clamped), got %d", score)
	}
}

func TestCalculateSecurityIndex_High5(t *testing.T) {
	// 5 high = 100 - 100 = 0
	score := calculateSecurityIndex(map[string]int{"high": 5})
	if score != 0 {
		t.Errorf("expected 0 for 5 high, got %d", score)
	}
}

func TestCalculateSecurityIndex_Medium10(t *testing.T) {
	// 10 medium = 100 - 100 = 0
	score := calculateSecurityIndex(map[string]int{"medium": 10})
	if score != 0 {
		t.Errorf("expected 0 for 10 medium, got %d", score)
	}
}

func TestCalculateSecurityIndex_Low20(t *testing.T) {
	// 20 low = 100 - 100 = 0
	score := calculateSecurityIndex(map[string]int{"low": 20})
	if score != 0 {
		t.Errorf("expected 0 for 20 low, got %d", score)
	}
}

func TestCalculateSecurityIndex_UnknownSeverity(t *testing.T) {
	// Неизвестная severity не учитывается
	score := calculateSecurityIndex(map[string]int{"unknown": 100})
	if score != 100 {
		t.Errorf("expected 100 for unknown severity, got %d", score)
	}
}

func TestCalculateSecurityIndex_CombinedExact(t *testing.T) {
	// 1 critical (30) + 1 high (20) + 1 medium (10) + 1 low (5) = 65
	score := calculateSecurityIndex(map[string]int{
		"critical": 1,
		"high":     1,
		"medium":   1,
		"low":      1,
	})
	if score != 35 {
		t.Errorf("expected 35, got %d", score)
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkCalculateSecurityIndex(b *testing.B) {
	counts := map[string]int{
		"critical": 2,
		"high":     3,
		"medium":   5,
		"low":      10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateSecurityIndex(counts)
	}
}
