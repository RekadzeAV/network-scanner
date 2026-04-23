package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestMaybePersistFloatPrefPrimesThenWrites(t *testing.T) {
	my := test.NewApp()
	p := my.Preferences()
	key := "gui.test.split_persist_e2e"
	var primed bool
	var last float64
	var persistCalls int
	onPersist := func(float64) { persistCalls++ }

	maybePersistFloatPref(p, key, 0.4, &primed, &last, onPersist)
	if !primed || last != 0.4 {
		t.Fatalf("after prime: primed=%v last=%v", primed, last)
	}
	maybePersistFloatPref(p, key, 0.4, &primed, &last, onPersist)
	if persistCalls != 0 {
		t.Fatalf("no persist on unchanged offset")
	}
	maybePersistFloatPref(p, key, 0.55, &primed, &last, onPersist)
	got := p.Float(key)
	if got < 0.54 || got > 0.56 {
		t.Fatalf("prefs value: got %v want ~0.55", got)
	}
	if persistCalls != 1 {
		t.Fatalf("onPersist calls: %d", persistCalls)
	}
}
