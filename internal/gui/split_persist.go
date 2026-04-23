package gui

import (
	"math"

	"fyne.io/fyne/v2"
)

// splitPersistEpsilon — минимальное изменение доли split, после которого пишем в Preferences.
const splitPersistEpsilon = 0.012

// maybePersistFloatPref обновляет last и при необходимости записывает key в prefs (с «прогревом» первого значения).
func maybePersistFloatPref(p fyne.Preferences, key string, cur float64, primed *bool, last *float64, onPersist func(float64)) {
	if p == nil || primed == nil || last == nil {
		return
	}
	if !*primed {
		*primed = true
		*last = cur
		return
	}
	if math.Abs(cur-*last) > splitPersistEpsilon {
		*last = cur
		p.SetFloat(key, cur)
		if onPersist != nil {
			onPersist(cur)
		}
	}
}
