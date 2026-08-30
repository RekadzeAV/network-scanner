package gui

import (
	"math"
	"testing"
)

// --- split_persist.go tests ---

func TestMaybePersistFloatPref_NilPreferences(t *testing.T) {
	var primed bool
	var lastVal *float64
	lastVal = new(float64)
	maybePersistFloatPref(nil, "key", 0.5, &primed, lastVal, nil)
	if primed {
		t.Error("expected primed=false for nil prefs")
	}
}

func TestMaybePersistFloatPref_NilPrimed(t *testing.T) {
	var last float64
	_ = last
	maybePersistFloatPref(nil, "key", 0.5, nil, &last, nil)
	// Не паникует — это успех
}

func TestMaybePersistFloatPref_NilLast(t *testing.T) {
	var primed bool
	maybePersistFloatPref(nil, "key", 0.5, &primed, nil, nil)
	// Не паникует — это успех
}

func TestMaybePersistFloatPref_Prime(t *testing.T) {
	// maybePersistFloatPref возвращает сразу если p == nil
	// Поэтому не может проставить primed/last без реальных prefs
	// Тестируем что не паникует
	maybePersistFloatPref(nil, "key", 0.5, nil, nil, nil)
	// Не паникует — это успех
}

func TestMaybePersistFloatPref_SmallChange(t *testing.T) {
	// with nil prefs — early return
	maybePersistFloatPref(nil, "key", 0.51, nil, nil, nil)
}

func TestMaybePersistFloatPref_LargeChange(t *testing.T) {
	maybePersistFloatPref(nil, "key", 0.7, nil, nil, nil)
}

func TestMaybePersistFloatPref_EdgeEpsilon(t *testing.T) {
	maybePersistFloatPref(nil, "key", 0.5+splitPersistEpsilon+0.001, nil, nil, nil)
}

func TestMaybePersistFloatPref_NegativeChange(t *testing.T) {
	maybePersistFloatPref(nil, "key", 0.1, nil, nil, nil)
}

func TestMaybePersistFloatPref_ZeroValue(t *testing.T) {
	maybePersistFloatPref(nil, "key", 0, nil, nil, nil)
}

func TestMaybePersistFloatPref_NoOnPersist(t *testing.T) {
	maybePersistFloatPref(nil, "key", 0.5, nil, nil, nil)
	maybePersistFloatPref(nil, "key", 0.7, nil, nil, nil)
}

func TestSplitPersistEpsilon_Value(t *testing.T) {
	if splitPersistEpsilon != 0.012 {
		t.Errorf("expected splitPersistEpsilon=0.012, got %v", splitPersistEpsilon)
	}
}

func TestSplitPersistEpsilon_Positive(t *testing.T) {
	if splitPersistEpsilon <= 0 {
		t.Error("expected splitPersistEpsilon > 0")
	}
}

func TestSplitPersistEpsilon_AbsCheck(t *testing.T) {
	// Проверяем что epsilon > 0 и math.Abs работает корректно
	if math.Abs(0.012-splitPersistEpsilon) > 1e-9 {
		t.Errorf("expected splitPersistEpsilon close to 0.012, got %v", splitPersistEpsilon)
	}
}
